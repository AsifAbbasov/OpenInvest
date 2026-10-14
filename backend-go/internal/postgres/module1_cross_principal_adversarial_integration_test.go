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

func TestModule1CrossPrincipalPortfolioAndIdempotencyIsolation(t *testing.T) {
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
			Email: "module1-" + label + "-" + uuid.NewString() + "@example.test",
			Password: "correct horse battery staple",
			Language: auth.LanguageEN,
			Theme: auth.ThemeSystem,
			Timezone: "UTC",
		})
		if err != nil {
			t.Fatalf("register %s: %v", label, err)
		}
		return result
	}
	a := register("a")
	b := register("b")

	do := func(method, path, bearer, body, idem string) *http.Response {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		if bearer != "" {
			req.Header.Set("Authorization", "Bearer "+bearer)
		}
		if body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		if idem != "" {
			req.Header.Set("Idempotency-Key", idem)
		}
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("%s %s: %v", method, path, err)
		}
		return resp
	}

	key := "module1-cross-principal-idem-" + strings.ReplaceAll(uuid.NewString(), "-", "")
	createBody := `{"name":"Isolation test","baseCurrency":"RUB"}`

	createB := do(http.MethodPost, "/api/v1/portfolios", b.Session.AccessToken, createBody, key)
	defer createB.Body.Close()
	if createB.StatusCode != http.StatusCreated {
		t.Fatalf("B create status=%d", createB.StatusCode)
	}
	var bPayload struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(createB.Body).Decode(&bPayload); err != nil {
		t.Fatalf("decode B create: %v", err)
	}
	if bPayload.Data.ID == "" {
		t.Fatal("missing B portfolio id")
	}

	createA := do(http.MethodPost, "/api/v1/portfolios", a.Session.AccessToken, createBody, key)
	defer createA.Body.Close()
	if createA.StatusCode != http.StatusCreated {
		t.Fatalf("same idempotency key under A must remain principal-scoped, status=%d", createA.StatusCode)
	}
	var aPayload struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(createA.Body).Decode(&aPayload); err != nil {
		t.Fatalf("decode A create: %v", err)
	}
	if aPayload.Data.ID == "" || aPayload.Data.ID == bPayload.Data.ID {
		t.Fatalf("idempotency response crossed principals: A=%q B=%q", aPayload.Data.ID, bPayload.Data.ID)
	}

	crossRead := do(http.MethodGet, "/api/v1/portfolios/"+bPayload.Data.ID, a.Session.AccessToken, "", "")
	defer crossRead.Body.Close()
	if crossRead.StatusCode == http.StatusOK {
		t.Fatalf("principal A read principal B portfolio")
	}
	if crossRead.StatusCode != http.StatusNotFound && crossRead.StatusCode != http.StatusForbidden {
		t.Fatalf("cross-principal read must fail closed as 404/403, got %d", crossRead.StatusCode)
	}

	ownerRead := do(http.MethodGet, "/api/v1/portfolios/"+bPayload.Data.ID, b.Session.AccessToken, "", "")
	defer ownerRead.Body.Close()
	if ownerRead.StatusCode != http.StatusOK {
		t.Fatalf("owner B cannot read own portfolio, status=%d", ownerRead.StatusCode)
	}

	listA := do(http.MethodGet, "/api/v1/portfolios", a.Session.AccessToken, "", "")
	defer listA.Body.Close()
	if listA.StatusCode != http.StatusOK {
		t.Fatalf("A list status=%d", listA.StatusCode)
	}
	var listPayload struct {
		Data struct {
			Items []struct {
				ID string `json:"id"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.NewDecoder(listA.Body).Decode(&listPayload); err != nil {
		t.Fatalf("decode A list: %v", err)
	}
	for _, item := range listPayload.Data.Items {
		if item.ID == bPayload.Data.ID {
			t.Fatalf("B portfolio leaked into A portfolio list")
		}
	}

	t.Log("CROSS_PRINCIPAL_READ=NO")
	t.Log("IDEMPOTENCY_RESULT_CROSSES_PRINCIPAL=NO")
	t.Log("SESSION_ISOLATION=PASS_FOR_TESTED_PORTFOLIO_BOUNDARY")
}
