package postgres_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/openinvest/openinvest/backend-go/internal/auth"
	"github.com/openinvest/openinvest/backend-go/internal/httpapi"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestEmailOwnershipJWTHTTPFinancialIsolation(t *testing.T) {
	url := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if url == "" {
		t.Skip("isolated PostgreSQL test service required")
	}
	store, err := postgres.Open(url)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	service, err := auth.NewService(store, nil, auth.Config{AccessTokenSecret: []byte("jwt-http-contract-local-signing-key-32-bytes"), RefreshCookieSecure: true})
	if err != nil {
		t.Fatal(err)
	}
	app, err := httpapi.New(verticalslice.NewService(store, jwtHTTPClock{}), service, []byte("jwt-http-review-local-signing-key-32-bytes"))
	if err != nil {
		t.Fatal(err)
	}
	call := func(method, path, body, token, cookie, csrf string) (int, map[string]json.RawMessage, *http.Cookie) {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		if cookie != "" {
			req.AddCookie(&http.Cookie{Name: auth.RefreshCookieName, Value: cookie})
		}
		if csrf != "" {
			req.Header.Set("X-CSRF-Token", csrf)
		}
		if method == "POST" && strings.HasPrefix(path, "/api/v1/portfolios") {
			req.Header.Set("Idempotency-Key", uuid.NewString())
		}
		resp, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second, FailOnTimeout: true})
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		var envelope map[string]json.RawMessage
		if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
			t.Fatal(err)
		}
		var refresh *http.Cookie
		for _, c := range resp.Cookies() {
			if c.Name == auth.RefreshCookieName {
				refresh = c
			}
		}
		return resp.StatusCode, envelope, refresh
	}
	type session struct {
		AccessToken string `json:"accessToken"`
		CSRF        string `json:"csrfToken"`
	}
	sessions := make([]session, 2)
	cookies := make([]string, 2)
	subjects := make([]string, 2)
	portfolios := make([]string, 2)
	for i := 0; i < 2; i++ {
		email := "jwt-http-" + uuid.NewString() + "@example.test"
		password := "synthetic-owner-password-2026"
		// Seed a legacy fixture directly; exercise actual HTTP login and refresh.
		seeded, err := service.Register(context.Background(), auth.RegistrationRequest{Email: email, Password: password, Language: "en", Theme: "system", Timezone: "UTC"})
		if err != nil {
			t.Fatal(err)
		}
		subjects[i] = seeded.User.InvestmentSubjectID
		body, _ := json.Marshal(map[string]string{"email": email, "password": password})
		status, envelope, cookie := call("POST", "/api/v1/auth/login", string(body), "", "", "")
		if status != 200 || cookie == nil {
			t.Fatalf("login status=%d", status)
		}
		cookies[i] = cookie.Value
		var data struct {
			Session session `json:"session"`
		}
		if err := json.Unmarshal(envelope["data"], &data); err != nil {
			t.Fatal(err)
		}
		sessions[i] = data.Session
		subject, err := service.AuthenticateAccessToken(sessions[i].AccessToken)
		if err != nil || subject != subjects[i] {
			t.Fatalf("financial propagation changed: %v", err)
		}
		status, envelope, _ = call("POST", "/api/v1/portfolios", `{"name":"JWT contract fixture","baseCurrency":"RUB"}`, sessions[i].AccessToken, "", "")
		if status != 201 {
			t.Fatalf("create portfolio status=%d body=%s", status, envelope["error"])
		}
		var portfolio struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(envelope["data"], &portfolio); err != nil {
			t.Fatal(err)
		}
		portfolios[i] = portfolio.ID
	}
	for i := 0; i < 2; i++ {
		status, _, _ := call("GET", "/api/v1/portfolios/"+portfolios[i], "", sessions[i].AccessToken, "", "")
		if status != 200 {
			t.Fatalf("own access status=%d", status)
		}
		status, _, _ = call("GET", "/api/v1/portfolios/"+portfolios[1-i], "", sessions[i].AccessToken, "", "")
		if status != 404 {
			t.Fatalf("other principal resource status=%d", status)
		}
		status, envelope, cookie := call("POST", "/api/v1/auth/refresh", "", "", cookies[i], sessions[i].CSRF)
		if status != 200 || cookie == nil || cookie.Value == cookies[i] {
			t.Fatalf("rotation status=%d", status)
		}
		var rotated session
		if err := json.Unmarshal(envelope["data"], &rotated); err != nil {
			t.Fatal(err)
		}
		subject, err := service.AuthenticateAccessToken(rotated.AccessToken)
		if err != nil || subject != subjects[i] {
			t.Fatalf("refresh changed financial principal: %v", err)
		}
		status, _, _ = call("GET", "/api/v1/portfolios/"+portfolios[1-i], "", rotated.AccessToken, "", "")
		if status != 404 {
			t.Fatalf("refreshed cross-principal status=%d", status)
		}
	}
	t.Log("JWT01_REAL_HTTP_LOGIN_REFRESH=PASS FINANCIAL_SUBJECT_PROPAGATION=PASS BIDIRECTIONAL_FINANCIAL_ISOLATION=PASS")
}

type jwtHTTPClock struct{}

func (jwtHTTPClock) Now() time.Time { return time.Now().UTC() }
