package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
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
	request := httptest.NewRequest(http.MethodOptions, "/api/v1/portfolios", nil)
	request.Header.Set("Origin", "http://localhost:3000")
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
	if got := response.Header.Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("expected allowed local origin, got %q", got)
	}
	if got := response.Header.Get("Access-Control-Allow-Headers"); got != "Accept, Authorization, Content-Type, Idempotency-Key, X-CSRF-Token, X-Request-ID, traceparent" {
		t.Fatalf("expected OpenAPI headers, got %q", got)
	}
}

func TestLocalDevelopmentCORSRejectsUnknownOrigin(t *testing.T) {
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
}

func TestAppendTransactionRequiresSettlementDateField(t *testing.T) {
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
