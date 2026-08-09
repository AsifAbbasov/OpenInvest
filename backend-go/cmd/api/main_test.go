package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	setExplicitDevelopmentEnvironment(t)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	response, err := newApp().Test(request)
	if err != nil {
		t.Fatalf("request health endpoint: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.StatusCode)
	}
}

func TestReadyWithoutDatabase(t *testing.T) {
	setExplicitDevelopmentEnvironment(t)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/ready", nil)
	response, err := newApp().Test(request)
	if err != nil {
		t.Fatalf("request readiness endpoint: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, response.StatusCode)
	}
}

func TestHealthPropagatesRequestAndTraceHeaders(t *testing.T) {
	setExplicitDevelopmentEnvironment(t)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	request.Header.Set("X-Request-ID", "11111111-1111-4111-8111-111111111111")
	request.Header.Set("traceparent", "00-11111111111111111111111111111111-2222222222222222-01")

	response, err := newApp().Test(request)
	if err != nil {
		t.Fatalf("request health endpoint: %v", err)
	}
	defer response.Body.Close()

	if got := response.Header.Get("X-Request-ID"); got != "11111111-1111-4111-8111-111111111111" {
		t.Fatalf("expected propagated request id, got %s", got)
	}
	if got := response.Header.Get("X-Trace-ID"); got != "11111111111111111111111111111111" {
		t.Fatalf("expected propagated trace id, got %s", got)
	}
}

func TestLocalDevelopmentCORSPreflight(t *testing.T) {
	setExplicitDevelopmentEnvironment(t)
	for _, origin := range []string{"http://localhost:3000", "http://127.0.0.1:3000"} {
		t.Run(origin, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodOptions, "/api/v1/portfolios", nil)
			request.Header.Set("Origin", origin)
			request.Header.Set("Access-Control-Request-Method", "POST")
			request.Header.Set("Access-Control-Request-Headers", "Content-Type, Idempotency-Key")

			response, err := newApp().Test(request)
			if err != nil {
				t.Fatalf("request preflight: %v", err)
			}
			defer response.Body.Close()

			if response.StatusCode != http.StatusNoContent {
				t.Fatalf("expected status %d, got %d", http.StatusNoContent, response.StatusCode)
			}
			if got := response.Header.Get("Access-Control-Allow-Origin"); got != origin {
				t.Fatalf("expected allowed local origin, got %q", got)
			}
			if got := response.Header.Get("Access-Control-Allow-Credentials"); got != "true" {
				t.Fatalf("expected credentialed local auth CORS, got %q", got)
			}
			if got := response.Header.Get("Access-Control-Allow-Headers"); got != "Accept, Authorization, Content-Type, Idempotency-Key, X-CSRF-Token, X-Request-ID, traceparent" {
				t.Fatalf("expected OpenAPI headers, got %q", got)
			}
		})
	}
}

func TestLocalDevelopmentCORSRejectsUnknownOrigin(t *testing.T) {
	setExplicitDevelopmentEnvironment(t)
	request := httptest.NewRequest(http.MethodOptions, "/api/v1/portfolios", nil)
	request.Header.Set("Origin", "https://evil.example")

	response, err := newApp().Test(request)
	if err != nil {
		t.Fatalf("request preflight: %v", err)
	}
	defer response.Body.Close()

	if got := response.Header.Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected unknown origin to be rejected, got %q", got)
	}
	if got := response.Header.Get("Access-Control-Allow-Credentials"); got != "" {
		t.Fatalf("expected unknown origin credentials to be rejected, got %q", got)
	}
}

func TestValidateRuntimeSafetyRejectsUnsafeAuthFlagsOutsideDevelopment(t *testing.T) {
	t.Setenv("OPENINVEST_ENV", "production")
	t.Setenv("OPENINVEST_DEV_AUTH_BYPASS", "true")

	if err := validateRuntimeSafety("postgres://openinvest@example/openinvest"); err == nil {
		t.Fatalf("expected unsafe development auth bypass to be rejected outside development")
	}
}

func TestValidateRuntimeSafetyAllowsUnsafeAuthFlagsOnlyInDevelopment(t *testing.T) {
	t.Setenv("OPENINVEST_ENV", "development")
	t.Setenv("OPENINVEST_DEV_AUTH_BYPASS", "true")
	t.Setenv("OPENINVEST_REFRESH_COOKIE_INSECURE", "true")

	if err := validateRuntimeSafety("postgres://openinvest@example/openinvest"); err != nil {
		t.Fatalf("expected explicit development mode to allow local auth flags: %v", err)
	}
}

func TestValidateRuntimeSafetyRejectsMissingDatabaseOutsideExplicitDevelopment(t *testing.T) {
	t.Setenv("OPENINVEST_ENV", "production")

	if err := validateRuntimeSafety(""); err == nil {
		t.Fatal("expected missing database URL to be rejected outside explicit development")
	}
}

func TestValidateRuntimeSafetyAllowsMissingDatabaseOnlyInExplicitDevelopment(t *testing.T) {
	t.Setenv("OPENINVEST_ENV", "local")

	if err := validateRuntimeSafety(""); err != nil {
		t.Fatalf("expected explicit local mode to permit the unavailable development store: %v", err)
	}
}

func TestConfiguredImportReviewTokenSecretUsesFallbackOnlyInExplicitDevelopment(t *testing.T) {
	t.Setenv("OPENINVEST_IMPORT_REVIEW_TOKEN_SECRET", "")
	t.Setenv("OPENINVEST_ENV", "development")
	if got := string(configuredImportReviewTokenSecret()); got != developmentImportReviewTokenSecret {
		t.Fatalf("expected development-only fallback secret, got %q", got)
	}

	t.Setenv("OPENINVEST_ENV", "production")
	if got := configuredImportReviewTokenSecret(); got != nil {
		t.Fatalf("expected no production fallback secret, got %q", string(got))
	}

	t.Setenv("OPENINVEST_IMPORT_REVIEW_TOKEN_SECRET", "configured-import-review-token-secret-32bytes")
	if got := string(configuredImportReviewTokenSecret()); got != "configured-import-review-token-secret-32bytes" {
		t.Fatalf("expected configured import review secret, got %q", got)
	}
}

func setExplicitDevelopmentEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("DATABASE_URL", "")
	t.Setenv("OPENINVEST_ENV", "development")
}

func TestAppendTransactionRequiresSettlementDateField(t *testing.T) {
	setExplicitDevelopmentEnvironment(t)
	body := []byte(`{
		"transactionType":"BUY",
		"ticker":"SBER",
		"quantity":"1.00000000",
		"unitPrice":{"amount":"100.00000000","currency":"RUB"},
		"commission":{"amount":"0.00000000","currency":"RUB"},
		"tax":{"amount":"0.00000000","currency":"RUB"},
		"tradeDate":"2026-01-10"
	}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/portfolios/00000000-0000-4000-8000-000000000002/transactions", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "transaction-key-000001")

	response, err := newApp().Test(request)
	if err != nil {
		t.Fatalf("request append transaction endpoint: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.StatusCode)
	}
}
