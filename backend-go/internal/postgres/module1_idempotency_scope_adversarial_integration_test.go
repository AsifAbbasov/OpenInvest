package postgres_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/openinvest/openinvest/backend-go/internal/auth"
	"github.com/openinvest/openinvest/backend-go/internal/httpapi"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestModule1IdempotencySecurityContextScopeMatrix(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}
	store, err := postgres.Open(databaseURL)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	defer store.Close()

	authService, err := auth.NewService(store, verticalslice.SystemClock{}, auth.Config{
		AccessTokenSecret:   []byte("module1-idempotency-scope-secret-32b"),
		RefreshCookieSecure: true,
	})
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}
	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	app, err := httpapi.New(service, authService, []byte("module1-idempotency-import-secret-32bytes"))
	if err != nil {
		t.Fatalf("new app: %v", err)
	}

	register := func(label string) auth.AuthResult {
		t.Helper()
		result, err := authService.Register(context.Background(), auth.RegistrationRequest{
			Email:    "module1-idem-" + label + "-" + uuid.NewString() + "@example.test",
			Password: "correct horse battery staple",
			Language: auth.LanguageEN,
			Theme:    auth.ThemeSystem,
			Timezone: "UTC",
		})
		if err != nil {
			t.Fatalf("register %s: %v", label, err)
		}
		return result
	}
	a := register("a")
	b := register("b")

	do := func(method, path, token, body, key string) *http.Response {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		if body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		if key != "" {
			req.Header.Set("Idempotency-Key", key)
		}
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("%s %s: %v", method, path, err)
		}
		return resp
	}

	decodeID := func(resp *http.Response) string {
		t.Helper()
		defer resp.Body.Close()
		var payload struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		return payload.Data.ID
	}

	createPortfolio := func(principal auth.AuthResult, key, name string) (int, string) {
		t.Helper()
		resp := do(http.MethodPost, "/api/v1/portfolios", principal.Session.AccessToken,
			`{"name":"`+name+`","baseCurrency":"RUB"}`, key)
		status := resp.StatusCode
		if status == http.StatusCreated {
			return status, decodeID(resp)
		}
		resp.Body.Close()
		return status, ""
	}

	// Same principal + same operation + same canonical payload: replay must be stable.
	keyRetry := "module1-idem-retry-" + strings.ReplaceAll(uuid.NewString(), "-", "")
	statusA1, portfolioA1 := createPortfolio(a, keyRetry, "A replay")
	if statusA1 != http.StatusCreated || portfolioA1 == "" {
		t.Fatalf("A first create status=%d id=%q", statusA1, portfolioA1)
	}
	statusA2, portfolioA2 := createPortfolio(a, keyRetry, "A replay")
	if statusA2 != http.StatusCreated || portfolioA2 != portfolioA1 {
		t.Fatalf("same-principal exact retry drift: status=%d first=%q replay=%q", statusA2, portfolioA1, portfolioA2)
	}
	t.Log("IDEMPOTENCY_EXACT_RETRY_STABLE=YES")

	// Same principal + same operation + same key + different payload: must not silently rebind.
	respConflict := do(http.MethodPost, "/api/v1/portfolios", a.Session.AccessToken,
		`{"name":"A different payload","baseCurrency":"RUB"}`, keyRetry)
	defer respConflict.Body.Close()
	if respConflict.StatusCode != http.StatusConflict {
		t.Fatalf("same-key different payload status=%d want=409", respConflict.StatusCode)
	}
	t.Log("IDEMPOTENCY_KEY_REBINDS_TO_DIFFERENT_PAYLOAD=NO")

	// Same key under another principal must not replay A's result.
	statusB, portfolioB := createPortfolio(b, keyRetry, "B independent")
	if statusB != http.StatusCreated || portfolioB == "" || portfolioB == portfolioA1 {
		t.Fatalf("cross-principal idempotency isolation failed: status=%d A=%q B=%q", statusB, portfolioA1, portfolioB)
	}
	t.Log("IDEMPOTENCY_RESULT_CROSSES_PRINCIPAL=NO")

	depositBody := `{"transactionType":"DEPOSIT","ticker":null,"quantity":null,"unitPrice":null,"grossAmount":{"amount":"100.00000000","currency":"RUB"},"commission":{"amount":"0.00000000","currency":"RUB"},"tax":{"amount":"0.00000000","currency":"RUB"},"tradeDate":"2026-10-13","settlementDate":null,"note":"module1 idempotency scope"}`

	// Same key string used by a different operation must not alias the create-portfolio replay record.
	operationKey := "module1-idem-operation-" + strings.ReplaceAll(uuid.NewString(), "-", "")
	statusOpPortfolio, opPortfolio := createPortfolio(a, operationKey, "A operation scope")
	if statusOpPortfolio != http.StatusCreated {
		t.Fatalf("operation-scope portfolio create status=%d", statusOpPortfolio)
	}
	txResp := do(http.MethodPost, "/api/v1/portfolios/"+opPortfolio+"/transactions", a.Session.AccessToken, depositBody, operationKey)
	if txResp.StatusCode != http.StatusCreated {
		status := txResp.StatusCode
		txResp.Body.Close()
		t.Fatalf("same key across different operation aliased unexpectedly: status=%d", status)
	}
	txResp.Body.Close()
	t.Log("IDEMPOTENCY_KEY_REBINDS_TO_DIFFERENT_OPERATION=NO")

	// Same principal and operation family but different resource path must remain resource/path scoped.
	resourceKey := "module1-idem-resource-" + strings.ReplaceAll(uuid.NewString(), "-", "")
	_, resourceP1 := createPortfolio(a, "module1-resource-p1-"+strings.ReplaceAll(uuid.NewString(), "-", ""), "A resource one")
	_, resourceP2 := createPortfolio(a, "module1-resource-p2-"+strings.ReplaceAll(uuid.NewString(), "-", ""), "A resource two")
	if resourceP1 == "" || resourceP2 == "" || resourceP1 == resourceP2 {
		t.Fatalf("resource fixture invalid: p1=%q p2=%q", resourceP1, resourceP2)
	}

	tx1 := do(http.MethodPost, "/api/v1/portfolios/"+resourceP1+"/transactions", a.Session.AccessToken, depositBody, resourceKey)
	if tx1.StatusCode != http.StatusCreated {
		status := tx1.StatusCode
		tx1.Body.Close()
		t.Fatalf("resource1 transaction status=%d", status)
	}
	tx1.Body.Close()

	tx2 := do(http.MethodPost, "/api/v1/portfolios/"+resourceP2+"/transactions", a.Session.AccessToken, depositBody, resourceKey)
	if tx2.StatusCode != http.StatusCreated {
		status := tx2.StatusCode
		tx2.Body.Close()
		t.Fatalf("same key across resource paths aliased unexpectedly: status=%d", status)
	}
	tx2.Body.Close()
	t.Log("IDEMPOTENCY_KEY_REBINDS_TO_DIFFERENT_RESOURCE=NO")
}
