package main

import (
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
