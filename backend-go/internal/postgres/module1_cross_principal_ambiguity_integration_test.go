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

func TestModule1CrossPrincipalAmbiguousCredentialsPreservePrincipalIntegrity(t *testing.T) {
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
		AccessTokenSecret:   []byte("module1-audit-access-token-secret-32bytes"),
		RefreshCookieSecure: true,
	})
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}
	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	app, err := httpapi.New(service, authService, []byte("module1-audit-import-review-secret-32bytes"))
	if err != nil {
		t.Fatalf("new http app: %v", err)
	}

	register := func(label string) auth.AuthResult {
		t.Helper()
		result, err := authService.Register(context.Background(), auth.RegistrationRequest{
			Email:    "module1-ambiguous-" + label + "-" + uuid.NewString() + "@example.test",
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

	createPortfolio := func(result auth.AuthResult, name string) string {
		t.Helper()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/portfolios", strings.NewReader(`{"name":"`+name+`","baseCurrency":"RUB"}`))
		req.Header.Set("Authorization", "Bearer "+result.Session.AccessToken)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", "module1-ambiguous-"+strings.ReplaceAll(uuid.NewString(), "-", ""))
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("create portfolio %s: %v", name, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("create portfolio %s status=%d", name, resp.StatusCode)
		}
		var payload struct{ Data struct{ ID string `json:"id"` } `json:"data"` }
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			t.Fatalf("decode create %s: %v", name, err)
		}
		return payload.Data.ID
	}
	aPortfolio := createPortfolio(a, "Principal A")
	bPortfolio := createPortfolio(b, "Principal B")

	assertListPrincipal := func(name string, first, second auth.AuthResult, wantID, forbidID string) {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/portfolios", nil)
		req.Header.Add("Authorization", "Bearer "+first.Session.AccessToken)
		req.Header.Add("Authorization", "Bearer "+second.Session.AccessToken)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("%s request: %v", name, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("%s status=%d", name, resp.StatusCode)
		}
		var payload struct {
			Data struct {
				Items []struct{ ID string `json:"id"` } `json:"items"`
			} `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			t.Fatalf("%s decode: %v", name, err)
		}
		seenWant := false
		for _, item := range payload.Data.Items {
			if item.ID == wantID {
				seenWant = true
			}
			if item.ID == forbidID {
				t.Fatalf("%s leaked second principal resource id=%s", name, forbidID)
			}
		}
		if !seenWant {
			t.Fatalf("%s did not resolve to first principal resource id=%s", name, wantID)
		}
		t.Logf("AUTH_CROSS_PRINCIPAL_COLLISION=%s STATUS=200 FIRST_PRINCIPAL_PRESERVED=YES", name)
	}

	assertListPrincipal("A_then_B", a, b, aPortfolio, bPortfolio)
	assertListPrincipal("B_then_A", b, a, bPortfolio, aPortfolio)

	refreshWithDuplicateCookies := func(name string, first, second auth.AuthResult) {
		t.Helper()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
		req.Header.Set("X-CSRF-Token", first.Session.CSRFToken)
		req.Header.Add("Cookie", auth.RefreshCookieName+"="+first.RefreshToken)
		req.Header.Add("Cookie", auth.RefreshCookieName+"="+second.RefreshToken)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("%s refresh: %v", name, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("%s status=%d want=200 first principal cookie must remain authoritative", name, resp.StatusCode)
		}
		t.Logf("COOKIE_CROSS_PRINCIPAL_COLLISION=%s STATUS=200 FIRST_PRINCIPAL_PRESERVED=YES", name)
	}

	// Use fresh principals for each order because a successful refresh rotates the selected credential.
	c := register("c")
	d := register("d")
	refreshWithDuplicateCookies("C_then_D", c, d)
	e := register("e")
	f := register("f")
	refreshWithDuplicateCookies("F_then_E", f, e)

	t.Log("CROSS_PRINCIPAL_AMBIGUOUS_AUTHORIZATION_CONFUSION=NO_FOR_TESTED_FIRST_VALUE_SEMANTICS")
	t.Log("CROSS_PRINCIPAL_AMBIGUOUS_REFRESH_COOKIE_CONFUSION=NO_FOR_TESTED_FIRST_VALUE_SEMANTICS")
}
