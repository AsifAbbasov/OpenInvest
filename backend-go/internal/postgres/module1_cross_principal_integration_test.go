package postgres_test

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/openinvest/openinvest/backend-go/internal/auth"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
)

type module1Principal struct {
	UserID        string
	AccessToken   string
	CSRFToken     string
	RefreshCookie string
}

func module1RegisterPrincipal(t *testing.T, app *fiber.App, email string) module1Principal {
	t.Helper()
	body := `{"email":"` + email + `","password":"correct horse battery staple","language":"en","theme":"system","timezone":"UTC"}`
	req := module1JSONRequest(http.MethodPost, "/api/v1/auth/register", body)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		data, _ := io.ReadAll(resp.Body)
		t.Fatalf("register %s status=%d body=%s", email, resp.StatusCode, data)
	}
	var payload struct {
		Data struct {
			User struct {
				ID string `json:"id"`
			} `json:"user"`
			Session struct {
				AccessToken string `json:"accessToken"`
				CSRFToken   string `json:"csrfToken"`
			} `json:"session"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	var refresh string
	for _, cookie := range resp.Cookies() {
		if cookie.Name == auth.RefreshCookieName {
			refresh = cookie.Value
		}
	}
	if payload.Data.User.ID == "" || payload.Data.Session.AccessToken == "" || payload.Data.Session.CSRFToken == "" || refresh == "" {
		t.Fatal("registration fixture missing auth material")
	}
	return module1Principal{
		UserID:        payload.Data.User.ID,
		AccessToken:   payload.Data.Session.AccessToken,
		CSRFToken:     payload.Data.Session.CSRFToken,
		RefreshCookie: refresh,
	}
}

func module1AuthedRequest(method, path, body, accessToken, idempotencyKey string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}
	return req
}

func module1CreatePortfolio(t *testing.T, app *fiber.App, principal module1Principal, key, name string) string {
	t.Helper()
	body := `{"name":"` + name + `","baseCurrency":"RUB"}`
	req := module1AuthedRequest(http.MethodPost, "/api/v1/portfolios", body, principal.AccessToken, key)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		data, _ := io.ReadAll(resp.Body)
		t.Fatalf("create portfolio status=%d body=%s", resp.StatusCode, data)
	}
	var payload struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if payload.Data.ID == "" {
		t.Fatal("missing portfolio id")
	}
	if !strings.Contains(strings.ToLower(resp.Header.Get("Cache-Control")), "no-store") {
		t.Fatalf("authenticated mutation response lacks no-store: %q", resp.Header.Get("Cache-Control"))
	}
	return payload.Data.ID
}

func TestModule1CrossPrincipalAuthorizationIdempotencyAndCache(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}
	store, err := postgres.Open(databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	app := module1AuthHTTPApp(t, store)

	a := module1RegisterPrincipal(t, app, "m1-a-"+uuid.NewString()+"@example.test")
	b := module1RegisterPrincipal(t, app, "m1-b-"+uuid.NewString()+"@example.test")

	sharedKey := "module1-cross-principal-shared-key-" + uuid.NewString()
	portfolioA := module1CreatePortfolio(t, app, a, sharedKey, "A private portfolio")
	portfolioB := module1CreatePortfolio(t, app, b, sharedKey, "B private portfolio")
	if portfolioA == portfolioB {
		t.Fatal("IDEMPOTENCY_RESULT_CROSSES_PRINCIPAL: same key returned same portfolio across principals")
	}

	bReadA := module1AuthedRequest(http.MethodGet, "/api/v1/portfolios/"+portfolioA, "", b.AccessToken, "")
	bReadAResp, err := app.Test(bReadA)
	if err != nil {
		t.Fatal(err)
	}
	bReadABody, _ := io.ReadAll(bReadAResp.Body)
	bReadAResp.Body.Close()
	if bReadAResp.StatusCode >= 200 && bReadAResp.StatusCode < 300 {
		t.Fatalf("CROSS_PRINCIPAL_READ status=%d body=%s", bReadAResp.StatusCode, bReadABody)
	}
	if strings.Contains(string(bReadABody), "A private portfolio") {
		t.Fatalf("A_PRIVATE_BODY_RETURNED_TO_B body=%s", bReadABody)
	}

	aRead := module1AuthedRequest(http.MethodGet, "/api/v1/portfolios/"+portfolioA, "", a.AccessToken, "")
	aReadResp, err := app.Test(aRead)
	if err != nil {
		t.Fatal(err)
	}
	aReadBody, _ := io.ReadAll(aReadResp.Body)
	aReadResp.Body.Close()
	if aReadResp.StatusCode != http.StatusOK {
		t.Fatalf("authorized A read expected=200 got=%d body=%s", aReadResp.StatusCode, aReadBody)
	}
	if !strings.Contains(strings.ToLower(aReadResp.Header.Get("Cache-Control")), "no-store") {
		t.Fatalf("AUTHENTICATED_RESPONSE_PUBLIC_CACHEABLE cache-control=%q", aReadResp.Header.Get("Cache-Control"))
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var before int
	if err := db.QueryRow(`SELECT count(*) FROM investment.transaction_entries WHERE portfolio_id=$1`, portfolioA).Scan(&before); err != nil {
		t.Fatal(err)
	}

	transactionBody := `{"transactionType":"DEPOSIT","ticker":null,"quantity":null,"unitPrice":null,"grossAmount":{"amount":"100.00000000","currency":"RUB"},"commission":{"amount":"0.00000000","currency":"RUB"},"tax":{"amount":"0.00000000","currency":"RUB"},"tradeDate":"2026-10-13","settlementDate":null}`
	bMutateA := module1AuthedRequest(http.MethodPost, "/api/v1/portfolios/"+portfolioA+"/transactions", transactionBody, b.AccessToken, "m1-cross-mutation-"+uuid.NewString())
	bMutateAResp, err := app.Test(bMutateA)
	if err != nil {
		t.Fatal(err)
	}
	bMutateAResp.Body.Close()
	if bMutateAResp.StatusCode >= 200 && bMutateAResp.StatusCode < 300 {
		t.Fatalf("CROSS_PRINCIPAL_MUTATION status=%d", bMutateAResp.StatusCode)
	}
	var after int
	if err := db.QueryRow(`SELECT count(*) FROM investment.transaction_entries WHERE portfolio_id=$1`, portfolioA).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if after != before {
		t.Fatalf("CROSS_PRINCIPAL_MUTATION persisted state before=%d after=%d", before, after)
	}

	t.Logf("CROSS_PRINCIPAL_READ=NO CROSS_PRINCIPAL_MUTATION=NO IDEMPOTENCY_RESULT_CROSSES_PRINCIPAL=NO AUTHENTICATED_RESPONSE_PUBLIC_CACHEABLE=NO A_PORTFOLIO=%s B_PORTFOLIO=%s", portfolioA, portfolioB)
}

func TestModule1CrossPrincipalCSRFCannotRotateAnotherUsersRefreshCookie(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}
	store, err := postgres.Open(databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	app := module1AuthHTTPApp(t, store)

	a := module1RegisterPrincipal(t, app, "m1-csrf-a-"+uuid.NewString()+"@example.test")
	b := module1RegisterPrincipal(t, app, "m1-csrf-b-"+uuid.NewString()+"@example.test")

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var before int
	if err := db.QueryRow(`SELECT count(*) FROM identity.sessions WHERE user_id=$1 AND session_state='active'`, a.UserID).Scan(&before); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: auth.RefreshCookieName, Value: a.RefreshCookie})
	req.Header.Set("X-CSRF-Token", b.CSRFToken)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("CSRF_SESSION_BINDING_BYPASS expected=401 got=%d", resp.StatusCode)
	}

	var after int
	if err := db.QueryRow(`SELECT count(*) FROM identity.sessions WHERE user_id=$1 AND session_state='active'`, a.UserID).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if after != before {
		t.Fatalf("CSRF_SESSION_BINDING_BYPASS persistent-state changed before=%d after=%d", before, after)
	}
	t.Logf("CSRF_SESSION_BINDING_BYPASS=NO ACTIVE_SESSIONS_BEFORE=%d ACTIVE_SESSIONS_AFTER=%d", before, after)
}
