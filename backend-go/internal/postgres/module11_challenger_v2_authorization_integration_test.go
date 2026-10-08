package postgres_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/openinvest/openinvest/backend-go/internal/auth"
	"github.com/openinvest/openinvest/backend-go/internal/httpapi"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

type m11V2HTTPSession struct {
	AccessToken  string
	CSRFToken    string
	RefreshToken string
}

func m11V2NewHTTPApp(t *testing.T) (*fiber.App, *auth.Service, *sql.DB) {
	t.Helper()
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" { t.Skip("OPENINVEST_DATABASE_TEST_URL is not set") }
	store, err := postgres.Open(databaseURL)
	if err != nil { t.Fatal(err) }
	t.Cleanup(func() { _ = store.Close() })
	authService, err := auth.NewService(store, verticalslice.SystemClock{}, auth.Config{
		AccessTokenSecret: []byte("m11-v2-access-token-secret-32-bytes"),
		RefreshCookieSecure: false,
	})
	if err != nil { t.Fatal(err) }
	app, err := httpapi.NewReplay(
		verticalslice.NewService(store, verticalslice.SystemClock{}),
		authService,
		[]byte("m11-v2-import-review-token-secret-32"),
	)
	if err != nil { t.Fatal(err) }
	db, err := sql.Open("pgx", databaseURL)
	if err != nil { t.Fatal(err) }
	t.Cleanup(func() { _ = db.Close() })
	return app, authService, db
}

func m11V2HTTPRequest(t *testing.T, app *fiber.App, method, path, body, bearer, idempotency string) (*http.Response, []byte) {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if body != "" { req.Header.Set("Content-Type", "application/json") }
	if bearer != "" { req.Header.Set("Authorization", "Bearer "+bearer) }
	if idempotency != "" { req.Header.Set("Idempotency-Key", idempotency) }
	resp, err := app.Test(req)
	if err != nil { t.Fatalf("%s %s: %v", method, path, err) }
	raw, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil { t.Fatal(err) }
	return resp, raw
}

func m11V2RegisterHTTP(t *testing.T, app *fiber.App, email string) m11V2HTTPSession {
	t.Helper()
	body := `{"email":"`+email+`","password":"correct horse battery staple","language":"en","theme":"system","timezone":"UTC"}`
	resp, raw := m11V2HTTPRequest(t, app, http.MethodPost, "/api/v1/auth/register", body, "", "")
	if resp.StatusCode != http.StatusCreated { t.Fatalf("register status=%d body=%s", resp.StatusCode, string(raw)) }
	var payload struct {
		Data struct {
			Session struct {
				AccessToken string `json:"accessToken"`
				CSRFToken string `json:"csrfToken"`
			} `json:"session"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil { t.Fatal(err) }
	var refresh string
	for _, c := range resp.Cookies() {
		if c.Name == auth.RefreshCookieName { refresh = c.Value }
	}
	if payload.Data.Session.AccessToken == "" || payload.Data.Session.CSRFToken == "" || refresh == "" {
		t.Fatal("incomplete registered session")
	}
	return m11V2HTTPSession{AccessToken:payload.Data.Session.AccessToken, CSRFToken:payload.Data.Session.CSRFToken, RefreshToken:refresh}
}

func m11V2CreatePortfolioHTTP(t *testing.T, app *fiber.App, token string) string {
	t.Helper()
	body := `{"name":"M11 V2 `+uuid.NewString()+`","baseCurrency":"RUB"}`
	resp, raw := m11V2HTTPRequest(t, app, http.MethodPost, "/api/v1/portfolios", body, token, uuid.NewString())
	if resp.StatusCode != http.StatusCreated { t.Fatalf("create portfolio status=%d body=%s", resp.StatusCode, string(raw)) }
	var payload struct { Data struct { ID string `json:"id"` } `json:"data"` }
	if err := json.Unmarshal(raw, &payload); err != nil { t.Fatal(err) }
	if payload.Data.ID == "" { t.Fatal("missing portfolio id") }
	return payload.Data.ID
}

func m11V2AppendHTTP(t *testing.T, app *fiber.App, token, portfolioID, txType, ticker, quantity, price, gross, date string) (string, int) {
	t.Helper()
	var tickerJSON, quantityJSON, priceJSON, grossJSON string
	if ticker == "" { tickerJSON = "null" } else { tickerJSON = `"`+ticker+`"` }
	if quantity == "" { quantityJSON = "null" } else { quantityJSON = `"`+quantity+`"` }
	if price == "" { priceJSON = "null" } else { priceJSON = `{"amount":"`+price+`","currency":"RUB"}` }
	if gross == "" { grossJSON = "null" } else { grossJSON = `{"amount":"`+gross+`","currency":"RUB"}` }
	body := `{"transactionType":"`+txType+`","ticker":`+tickerJSON+`,"quantity":`+quantityJSON+`,"unitPrice":`+priceJSON+`,"grossAmount":`+grossJSON+`,"commission":{"amount":"0.00000000","currency":"RUB"},"tax":{"amount":"0.00000000","currency":"RUB"},"tradeDate":"`+date+`","settlementDate":null,"note":null}`
	resp, raw := m11V2HTTPRequest(t, app, http.MethodPost, "/api/v1/portfolios/"+portfolioID+"/transactions", body, token, uuid.NewString())
	if resp.StatusCode != http.StatusCreated { t.Fatalf("append status=%d body=%s", resp.StatusCode, string(raw)) }
	var payload struct { Data struct { ID string `json:"id"`; Revision int `json:"revision"` } `json:"data"` }
	if err := json.Unmarshal(raw, &payload); err != nil { t.Fatal(err) }
	return payload.Data.ID, payload.Data.Revision
}

func m11V2CountRows(t *testing.T, db *sql.DB, query string, args ...any) int {
	t.Helper()
	var count int
	if err := db.QueryRow(query, args...).Scan(&count); err != nil { t.Fatal(err) }
	return count
}

func TestM11V2ProtectedMutationAuthorizationHTTP(t *testing.T) {
	app, authService, db := m11V2NewHTTPApp(t)
	owner := m11V2RegisterHTTP(t, app, "m11-v2-owner-"+uuid.NewString()+"@example.com")
	foreign := m11V2RegisterHTTP(t, app, "m11-v2-foreign-"+uuid.NewString()+"@example.com")
	foreignSubject, err := authService.AuthenticateAccessToken(foreign.AccessToken)
	if err != nil { t.Fatal(err) }
	portfolioID := m11V2CreatePortfolioHTTP(t, app, owner.AccessToken)
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM investment.portfolio_manual_valuations WHERE portfolio_id=$1`, portfolioID)
		_, _ = db.Exec(`DELETE FROM investment.command_deduplication WHERE canonical_path LIKE $1`, "%"+portfolioID+"%")
		cleanupPortfolioRows(t, context.Background(), db, portfolioID)
	})

	correctionID, correctionRevision := m11V2AppendHTTP(t, app, owner.AccessToken, portfolioID, "DEPOSIT", "", "", "", "1000.00000000", "2026-10-01")
	reversalID, reversalRevision := m11V2AppendHTTP(t, app, owner.AccessToken, portfolioID, "DEPOSIT", "", "", "", "200.00000000", "2026-10-02")
	_, _ = m11V2AppendHTTP(t, app, owner.AccessToken, portfolioID, "BUY", "SBER", "10.00000000", "100.00000000", "", "2026-10-03")

	total, passed := 8, 0
	t.Run("portfolio create anonymous rejected", func(t *testing.T) {
		before := m11V2CountRows(t, db, `SELECT count(*) FROM investment.portfolios`)
		resp, _ := m11V2HTTPRequest(t, app, http.MethodPost, "/api/v1/portfolios", `{"name":"anonymous","baseCurrency":"RUB"}`, "", uuid.NewString())
		after := m11V2CountRows(t, db, `SELECT count(*) FROM investment.portfolios`)
		if resp.StatusCode != http.StatusUnauthorized || after != before { t.Fatalf("status=%d before=%d after=%d", resp.StatusCode, before, after) }
		passed++
	})

	t.Run("transaction append foreign rejected", func(t *testing.T) {
		before := m11V2CountRows(t, db, `SELECT count(*) FROM investment.transaction_entries WHERE portfolio_id=$1`, portfolioID)
		body := `{"transactionType":"DEPOSIT","ticker":null,"quantity":null,"unitPrice":null,"grossAmount":{"amount":"99.00000000","currency":"RUB"},"commission":{"amount":"0.00000000","currency":"RUB"},"tax":{"amount":"0.00000000","currency":"RUB"},"tradeDate":"2026-10-04","settlementDate":null,"note":null}`
		key := uuid.NewString()
		resp, _ := m11V2HTTPRequest(t, app, http.MethodPost, "/api/v1/portfolios/"+portfolioID+"/transactions", body, foreign.AccessToken, key)
		after := m11V2CountRows(t, db, `SELECT count(*) FROM investment.transaction_entries WHERE portfolio_id=$1`, portfolioID)
		residue := m11V2CountRows(t, db, `SELECT count(*) FROM investment.command_deduplication WHERE principal_id=$1 AND idempotency_key=$2`, foreignSubject, key)
		if resp.StatusCode != http.StatusNotFound || after != before || residue != 0 { t.Fatalf("status=%d before=%d after=%d residue=%d", resp.StatusCode, before, after, residue) }
		passed++
	})

	t.Run("correction foreign rejected owner succeeds", func(t *testing.T) {
		body := `{"expectedRevision":`+strconv.Itoa(correctionRevision)+`,"reason":"m11 v2 correction","corrected":{"transactionType":"DEPOSIT","ticker":null,"quantity":null,"unitPrice":null,"grossAmount":{"amount":"1100.00000000","currency":"RUB"},"commission":{"amount":"0.00000000","currency":"RUB"},"tax":{"amount":"0.00000000","currency":"RUB"},"tradeDate":"2026-10-01","settlementDate":null,"note":null}}`
		path := "/api/v1/portfolios/"+portfolioID+"/transactions/"+correctionID
		before := m11V2CountRows(t, db, `SELECT count(*) FROM investment.transaction_entries WHERE transaction_id=$1`, correctionID)
		resp, _ := m11V2HTTPRequest(t, app, http.MethodPatch, path, body, foreign.AccessToken, uuid.NewString())
		mid := m11V2CountRows(t, db, `SELECT count(*) FROM investment.transaction_entries WHERE transaction_id=$1`, correctionID)
		ownerResp, _ := m11V2HTTPRequest(t, app, http.MethodPatch, path, body, owner.AccessToken, uuid.NewString())
		after := m11V2CountRows(t, db, `SELECT count(*) FROM investment.transaction_entries WHERE transaction_id=$1`, correctionID)
		if resp.StatusCode != http.StatusNotFound || mid != before || ownerResp.StatusCode != http.StatusOK || after != before+1 { t.Fatalf("foreign=%d owner=%d counts=%d/%d/%d", resp.StatusCode, ownerResp.StatusCode, before, mid, after) }
		passed++
	})

	t.Run("reversal foreign rejected owner succeeds", func(t *testing.T) {
		body := `{"expectedRevision":`+strconv.Itoa(reversalRevision)+`,"reason":"m11 v2 reversal","effectiveDate":"2026-10-05"}`
		path := "/api/v1/portfolios/"+portfolioID+"/transactions/"+reversalID
		before := m11V2CountRows(t, db, `SELECT count(*) FROM investment.transaction_entries WHERE transaction_id=$1 OR reverses_transaction_id=$1`, reversalID)
		resp, _ := m11V2HTTPRequest(t, app, http.MethodDelete, path, body, foreign.AccessToken, uuid.NewString())
		mid := m11V2CountRows(t, db, `SELECT count(*) FROM investment.transaction_entries WHERE transaction_id=$1 OR reverses_transaction_id=$1`, reversalID)
		ownerResp, _ := m11V2HTTPRequest(t, app, http.MethodDelete, path, body, owner.AccessToken, uuid.NewString())
		after := m11V2CountRows(t, db, `SELECT count(*) FROM investment.transaction_entries WHERE transaction_id=$1 OR reverses_transaction_id=$1`, reversalID)
		if resp.StatusCode != http.StatusNotFound || mid != before || ownerResp.StatusCode != http.StatusOK || after != before+1 { t.Fatalf("foreign=%d owner=%d counts=%d/%d/%d", resp.StatusCode, ownerResp.StatusCode, before, mid, after) }
		passed++
	})

	t.Run("valuation update delete ownership", func(t *testing.T) {
		path := "/api/v1/portfolios/"+portfolioID+"/valuations/SBER"
		ownerBody := `{"marketPrice":{"amount":"110.00000000","currency":"RUB"},"asOfDate":"2026-10-06"}`
		ownerResp, _ := m11V2HTTPRequest(t, app, http.MethodPut, path, ownerBody, owner.AccessToken, "")
		if ownerResp.StatusCode != http.StatusOK { t.Fatalf("owner put=%d", ownerResp.StatusCode) }
		foreignResp, _ := m11V2HTTPRequest(t, app, http.MethodPut, path, `{"marketPrice":{"amount":"999.00000000","currency":"RUB"},"asOfDate":"2026-10-06"}`, foreign.AccessToken, "")
		if foreignResp.StatusCode != http.StatusNotFound { t.Fatalf("foreign put=%d", foreignResp.StatusCode) }
		var price string
		if err := db.QueryRow(`SELECT valuation.price_amount::text FROM investment.portfolio_manual_valuations valuation JOIN investment.assets asset ON asset.id=valuation.asset_id WHERE valuation.portfolio_id=$1 AND asset.ticker='SBER'`, portfolioID).Scan(&price); err != nil { t.Fatal(err) }
		if price != "110.00000000" { t.Fatalf("foreign changed price=%s", price) }
		deleteResp, _ := m11V2HTTPRequest(t, app, http.MethodDelete, path, "", foreign.AccessToken, "")
		if deleteResp.StatusCode != http.StatusNotFound { t.Fatalf("foreign delete=%d", deleteResp.StatusCode) }
		passed++
	})

	csvPayload := "transaction_type,ticker,quantity,unit_price,gross_amount,commission,tax,trade_date,settlement_date,currency,broker_operation_id,note\nDEPOSIT,,,,50.00000000,0.00000000,0.00000000,2026-10-07,,RUB,m11-v2-import-1,note\n"
	reviewBodyBytes, _ := json.Marshal(map[string]any{"sourceAccountLabel":"M11 V2","csvPayload":csvPayload})
	var reviewData struct {
		SourceFileHash string `json:"sourceFileHash"`
		ReviewToken string `json:"reviewToken"`
		Rows []struct { RowNumber int `json:"rowNumber"`; RowHash string `json:"rowHash"` } `json:"rows"`
	}
	t.Run("import review ownership", func(t *testing.T) {
		path := "/api/v1/portfolios/"+portfolioID+"/imports/review"
		foreignResp, _ := m11V2HTTPRequest(t, app, http.MethodPost, path, string(reviewBodyBytes), foreign.AccessToken, "")
		if foreignResp.StatusCode != http.StatusNotFound { t.Fatalf("foreign review=%d", foreignResp.StatusCode) }
		ownerResp, raw := m11V2HTTPRequest(t, app, http.MethodPost, path, string(reviewBodyBytes), owner.AccessToken, "")
		if ownerResp.StatusCode != http.StatusOK { t.Fatalf("owner review=%d body=%s", ownerResp.StatusCode, string(raw)) }
		var payload struct { Data json.RawMessage `json:"data"` }
		if err := json.Unmarshal(raw, &payload); err != nil { t.Fatal(err) }
		if err := json.Unmarshal(payload.Data, &reviewData); err != nil { t.Fatal(err) }
		if reviewData.ReviewToken == "" || reviewData.SourceFileHash == "" || len(reviewData.Rows) != 1 { t.Fatalf("review data=%+v", reviewData) }
		passed++
	})

	t.Run("import append signed proof cannot cross principal", func(t *testing.T) {
		before := m11V2CountRows(t, db, `SELECT count(*) FROM investment.transaction_entries WHERE portfolio_id=$1`, portfolioID)
		appendBody, _ := json.Marshal(map[string]any{
			"sourceAccountLabel":"M11 V2","sourceFileHash":reviewData.SourceFileHash,"reviewToken":reviewData.ReviewToken,
			"csvPayload":csvPayload,
			"decisions":[]map[string]any{{"rowNumber":reviewData.Rows[0].RowNumber,"rowHash":reviewData.Rows[0].RowHash,"action":"APPROVE"}},
		})
		foreignResp, _ := m11V2HTTPRequest(t, app, http.MethodPost, "/api/v1/portfolios/"+portfolioID+"/imports/append", string(appendBody), foreign.AccessToken, uuid.NewString())
		mid := m11V2CountRows(t, db, `SELECT count(*) FROM investment.transaction_entries WHERE portfolio_id=$1`, portfolioID)
		ownerResp, _ := m11V2HTTPRequest(t, app, http.MethodPost, "/api/v1/portfolios/"+portfolioID+"/imports/append", string(appendBody), owner.AccessToken, uuid.NewString())
		after := m11V2CountRows(t, db, `SELECT count(*) FROM investment.transaction_entries WHERE portfolio_id=$1`, portfolioID)
		if foreignResp.StatusCode == http.StatusOK || foreignResp.StatusCode == http.StatusCreated || mid != before || ownerResp.StatusCode != http.StatusCreated || after != before+1 {
			t.Fatalf("foreign=%d owner=%d counts=%d/%d/%d", foreignResp.StatusCode, ownerResp.StatusCode, before, mid, after)
		}
		passed++
	})

	t.Run("anonymous protected mutation rejected", func(t *testing.T) {
		before := m11V2CountRows(t, db, `SELECT count(*) FROM investment.transaction_entries WHERE portfolio_id=$1`, portfolioID)
		body := `{"transactionType":"DEPOSIT","ticker":null,"quantity":null,"unitPrice":null,"grossAmount":{"amount":"1.00000000","currency":"RUB"},"commission":{"amount":"0.00000000","currency":"RUB"},"tax":{"amount":"0.00000000","currency":"RUB"},"tradeDate":"2026-10-08","settlementDate":null,"note":null}`
		resp, _ := m11V2HTTPRequest(t, app, http.MethodPost, "/api/v1/portfolios/"+portfolioID+"/transactions", body, "", uuid.NewString())
		after := m11V2CountRows(t, db, `SELECT count(*) FROM investment.transaction_entries WHERE portfolio_id=$1`, portfolioID)
		if resp.StatusCode != http.StatusUnauthorized || after != before { t.Fatalf("status=%d before=%d after=%d", resp.StatusCode, before, after) }
		passed++
	})

	t.Logf("M11_ROUTE_SECURITY_MATRIX_TOTAL=%d", total)
	t.Logf("M11_ROUTE_SECURITY_MATRIX_PASS=%d", passed)
	t.Logf("M11_ROUTE_SECURITY_MATRIX_FAIL=%d", total-passed)
	t.Log("M11_ROUTE_DISCOVERY_PORTFOLIO_UPDATE_DELETE=NOT_PRESENT")
}
