package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestStage0332ExpiredTamperedImportTokenCannotRecoverCompletedCommand(t *testing.T) {
	store := &stage32ImportReplayStore{}
	service := verticalslice.NewService(store, fixedHTTPClock{})
	secret, err := normalizedImportReviewSecret([]byte("stage-03-32-import-tamper-test-secret-32-bytes"))
	if err != nil {
		t.Fatalf("normalize import review secret: %v", err)
	}
	now := time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)
	api := &API{
		service:                 service,
		allowDevelopmentSubject: true,
		authLimiter:             newAuthRateLimiter(20, time.Minute),
		importReviewSecret:      secret,
		paginationCursorSecret:  derivePaginationCursorSecret(secret),
		now:                     func() time.Time { return now },
	}
	app := newReplayApp(api)
	body := validImportAppendBody(t, app, importCSV)
	path := "/api/v1/portfolios/00000000-0000-4000-8000-000000000002/imports/append"
	key := "stage-03-32-import-tamper-key-01"

	firstRequest := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	firstRequest.Header.Set("Content-Type", "application/json")
	firstRequest.Header.Set("Idempotency-Key", key)
	firstResponse, err := app.Test(firstRequest)
	if err != nil {
		t.Fatalf("first import append: %v", err)
	}
	defer firstResponse.Body.Close()
	if firstResponse.StatusCode != http.StatusCreated {
		t.Fatalf("first import status: got %d want %d", firstResponse.StatusCode, http.StatusCreated)
	}
	if store.appendReplayCalls != 1 || store.lookupCalls != 0 {
		t.Fatalf("unexpected first-request calls: append=%d lookup=%d", store.appendReplayCalls, store.lookupCalls)
	}

	var request map[string]any
	if err := json.Unmarshal(body, &request); err != nil {
		t.Fatalf("decode append body: %v", err)
	}
	token, ok := request["reviewToken"].(string)
	if !ok || token == "" {
		t.Fatal("append body is missing reviewToken")
	}
	request["reviewToken"] = token + "x"
	tamperedBody, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("encode tampered append body: %v", err)
	}

	// Move beyond the valid token lifetime so a loose "any expired token" recovery rule would be
	// vulnerable. The HMAC/context verifier must still reject this request before replay lookup.
	now = now.Add(20 * time.Minute)
	secondRequest := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(tamperedBody))
	secondRequest.Header.Set("Content-Type", "application/json")
	secondRequest.Header.Set("Idempotency-Key", key)
	secondResponse, err := app.Test(secondRequest)
	if err != nil {
		t.Fatalf("tampered expired-token request: %v", err)
	}
	defer secondResponse.Body.Close()
	if secondResponse.StatusCode != http.StatusBadRequest {
		t.Fatalf("tampered expired token recovered a command: got status %d want %d", secondResponse.StatusCode, http.StatusBadRequest)
	}
	if store.appendReplayCalls != 1 {
		t.Fatalf("tampered expired token executed another financial append: %d", store.appendReplayCalls)
	}
	if store.lookupCalls != 0 {
		t.Fatalf("tampered expired token reached replay lookup: %d", store.lookupCalls)
	}
}
