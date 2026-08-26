package httpapi

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

type stage338ExpiredRetentionReplayStore struct {
	stage32ImportReplayStore
}

func (store *stage338ExpiredRetentionReplayStore) LookupReplayArtifact(
	_ context.Context,
	_ verticalslice.CommandContext,
	_ string,
) (verticalslice.CommandReplayArtifact, bool, error) {
	store.lookupCalls++
	return verticalslice.CommandReplayArtifact{}, false, nil
}

func TestStage0338ExpiredImportProofCannotAuthorizeFreshWriteAfterReplayRetention(t *testing.T) {
	store := &stage338ExpiredRetentionReplayStore{}
	service := verticalslice.NewService(store, fixedHTTPClock{})
	secret, err := normalizedImportReviewSecret([]byte("stage-03-38-import-retention-test-secret-32-bytes"))
	if err != nil {
		t.Fatalf("normalize import review secret: %v", err)
	}

	now := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
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
	key := "stage-03-38-import-retention-key-01"

	firstRequest := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	firstRequest.Header.Set("Content-Type", "application/json")
	firstRequest.Header.Set("Idempotency-Key", key)
	firstResponse, err := app.Test(firstRequest)
	if err != nil {
		t.Fatalf("first import append: %v", err)
	}
	defer firstResponse.Body.Close()
	firstBody, err := io.ReadAll(firstResponse.Body)
	if err != nil {
		t.Fatalf("read first import response: %v", err)
	}
	if firstResponse.StatusCode != http.StatusCreated {
		t.Fatalf("first import status: got %d want %d body=%s", firstResponse.StatusCode, http.StatusCreated, firstBody)
	}
	if store.appendReplayCalls != 1 || store.lookupCalls != 0 {
		t.Fatalf("unexpected first-write calls: append=%d lookup=%d", store.appendReplayCalls, store.lookupCalls)
	}

	// The review proof and the 24-hour replay-retention window are both now stale.
	// The store models the PostgreSQL lookup result after command expiry: not found.
	now = now.Add(25 * time.Hour)
	secondRequest := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	secondRequest.Header.Set("Content-Type", "application/json")
	secondRequest.Header.Set("Idempotency-Key", key)
	secondResponse, err := app.Test(secondRequest)
	if err != nil {
		t.Fatalf("expired-proof import request: %v", err)
	}
	defer secondResponse.Body.Close()
	secondBody, err := io.ReadAll(secondResponse.Body)
	if err != nil {
		t.Fatalf("read expired-proof response: %v", err)
	}
	if secondResponse.StatusCode == http.StatusCreated {
		t.Fatalf("expired proof plus missing retained replay authorized a fresh write: body=%s", secondBody)
	}
	if store.lookupCalls != 1 {
		t.Fatalf("expected one read-only recovery lookup, got %d", store.lookupCalls)
	}
	if store.appendReplayCalls != 1 {
		t.Fatalf("expired proof triggered a second financial append: calls=%d", store.appendReplayCalls)
	}
}
