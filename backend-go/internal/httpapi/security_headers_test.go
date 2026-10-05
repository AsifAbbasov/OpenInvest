package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/openinvest/openinvest/backend-go/internal/auth"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestSensitiveResponseCachePolicyCoversAuthLifecycleAndErrors(t *testing.T) {
	store := &httpAuthTestStore{}
	app := newHTTPAuthApp(t, newHTTPAuthService(t, store))
	registration := `{
		"email":"cache-policy@example.com",
		"password":"correct horse battery staple",
		"language":"en",
		"theme":"system",
		"timezone":"UTC"
	}`

	register := securityAuthRequest(t, app, http.MethodPost, "/api/v1/auth/register", registration, "", "")
	assertNoStoreResponse(t, register, http.StatusCreated)
	refreshCookie := requireCookie(t, register, auth.RefreshCookieName)
	csrfToken := readCSRFToken(t, register)
	register.Body.Close()

	loginFailure := securityAuthRequest(t, app, http.MethodPost, "/api/v1/auth/login", `{
		"email":"cache-policy@example.com",
		"password":"wrong password value"
	}`, "", "")
	assertNoStoreResponse(t, loginFailure, http.StatusUnauthorized)
	loginFailure.Body.Close()

	refreshFailure := securityAuthRequest(t, app, http.MethodPost, "/api/v1/auth/refresh", "", refreshCookie.Value, "")
	assertNoStoreResponse(t, refreshFailure, http.StatusUnauthorized)
	refreshFailure.Body.Close()

	refreshSuccess := securityAuthRequest(t, app, http.MethodPost, "/api/v1/auth/refresh", "", refreshCookie.Value, csrfToken)
	assertNoStoreResponse(t, refreshSuccess, http.StatusOK)
	rotatedCookie := requireCookie(t, refreshSuccess, auth.RefreshCookieName)
	rotatedCSRF := readRefreshedCSRFToken(t, refreshSuccess)
	refreshSuccess.Body.Close()

	logout := securityAuthRequest(t, app, http.MethodPost, "/api/v1/auth/logout", `{"allSessions":false}`, rotatedCookie.Value, rotatedCSRF)
	assertNoStoreResponse(t, logout, http.StatusOK)
	logout.Body.Close()
}

func TestSensitiveResponseCachePolicyCoversAuthenticatedPortfolioSuccessAndError(t *testing.T) {
	portfolioStore := &importAPITestStore{portfolios: []verticalslice.Portfolio{{
		ID:           "portfolio-cache-policy",
		Name:         "Cache policy portfolio",
		BaseCurrency: "RUB",
		Version:      1,
		CreatedAt:    time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC),
	}}}
	authStore := &httpAuthTestStore{}
	app, err := New(
		verticalslice.NewService(portfolioStore, fixedHTTPClock{}),
		newHTTPAuthService(t, authStore),
		[]byte("test-import-review-token-secret-32-bytes"),
	)
	if err != nil {
		t.Fatalf("new HTTP app: %v", err)
	}

	register := securityAuthRequest(t, app, http.MethodPost, "/api/v1/auth/register", `{
		"email":"portfolio-cache-policy@example.com",
		"password":"correct horse battery staple",
		"language":"en",
		"theme":"system",
		"timezone":"UTC"
	}`, "", "")
	accessToken := readAccessToken(t, register)
	register.Body.Close()

	success := requestWithBearer(t, app, http.MethodGet, "/api/v1/portfolios", accessToken)
	assertNoStoreResponse(t, success, http.StatusOK)
	success.Body.Close()

	unauthorized := requestWithBearer(t, app, http.MethodGet, "/api/v1/portfolios", "not-a-valid-token")
	assertNoStoreResponse(t, unauthorized, http.StatusUnauthorized)
	unauthorized.Body.Close()
}

func TestSensitiveResponseCachePolicyLeavesPublicAssetRouteUnclassified(t *testing.T) {
	app := newApp(&API{
		service:                verticalslice.NewService(&importAPITestStore{}, fixedHTTPClock{}),
		importReviewSecret:     []byte("test-import-review-token-secret-32-bytes"),
		paginationCursorSecret: []byte("test-pagination-cursor-secret-32bytes"),
	})

	request := httptest.NewRequest(http.MethodGet, "/api/v1/assets/search?query=SBER", nil)
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request public asset route: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("public asset status = %d, want %d", response.StatusCode, http.StatusOK)
	}
	if got := response.Header.Get("Cache-Control"); got != "" {
		t.Fatalf("public asset route must retain handler-owned cache policy, got %q", got)
	}
}

func TestSensitiveResponsePathClassification(t *testing.T) {
	for _, path := range []string{
		"/api/v1/auth/login",
		"/api/v1/auth/refresh",
		"/api/v1/portfolios",
		"/api/v1/portfolios/portfolio-id/summary",
	} {
		if !isSensitiveResponsePath(path) {
			t.Fatalf("expected %q to be sensitive", path)
		}
	}
	for _, path := range []string{
		"/api/v1/assets/search",
		"/api/v1/corporate-actions/projection",
		"/api/v1/portfolios-public",
	} {
		if isSensitiveResponsePath(path) {
			t.Fatalf("expected %q to remain public-route classified", path)
		}
	}
}

func assertNoStoreResponse(t *testing.T, response *http.Response, wantStatus int) {
	t.Helper()
	if response.StatusCode != wantStatus {
		t.Fatalf("status = %d, want %d", response.StatusCode, wantStatus)
	}
	if got := response.Header.Get("Cache-Control"); got != sensitiveResponseCacheControl {
		t.Fatalf("Cache-Control = %q, want %q", got, sensitiveResponseCacheControl)
	}
	if got := response.Header.Get("ETag"); got != "" {
		t.Fatalf("sensitive response must not emit ETag, got %q", got)
	}
	if got := response.Header.Get("Last-Modified"); got != "" {
		t.Fatalf("sensitive response must not emit Last-Modified, got %q", got)
	}
}

func requestWithBearer(t *testing.T, app *fiber.App, method string, path string, token string) *http.Response {
	t.Helper()
	request := httptest.NewRequest(method, path, bytes.NewReader(nil))
	request.Header.Set("Authorization", "Bearer "+token)
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request %s %s: %v", method, path, err)
	}
	return response
}

func securityAuthRequest(t *testing.T, app *fiber.App, method string, path string, body string, refreshToken string, csrfToken string) *http.Response {
	t.Helper()
	request := httptest.NewRequest(method, path, bytes.NewReader([]byte(body)))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	if refreshToken != "" {
		request.AddCookie(&http.Cookie{Name: auth.RefreshCookieName, Value: refreshToken})
	}
	if csrfToken != "" {
		request.Header.Set("X-CSRF-Token", csrfToken)
	}
	response, err := app.Test(request, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("request %s %s: %v", method, path, err)
	}
	return response
}

func readAccessToken(t *testing.T, response *http.Response) string {
	t.Helper()
	var payload struct {
		Data struct {
			Session struct {
				AccessToken string `json:"accessToken"`
			} `json:"session"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode access token: %v", err)
	}
	if payload.Data.Session.AccessToken == "" {
		t.Fatal("missing access token")
	}
	return payload.Data.Session.AccessToken
}

func readRefreshedCSRFToken(t *testing.T, response *http.Response) string {
	t.Helper()
	var payload struct {
		Data struct {
			CSRFToken string `json:"csrfToken"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode refreshed csrf token: %v", err)
	}
	if payload.Data.CSRFToken == "" {
		t.Fatal("missing refreshed csrf token")
	}
	return payload.Data.CSRFToken
}
