package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestStage339ImportDTOUsesCodePointsWhileCSVLimitStaysBytes(t *testing.T) {
	for _, tc := range []struct {
		name      string
		label     string
		wantError bool
	}{
		{name: "120 Cyrillic", label: strings.Repeat("Ж", 120)},
		{name: "121 Cyrillic", label: strings.Repeat("Ж", 121), wantError: true},
		{name: "120 supplementary", label: strings.Repeat("😀", 120)},
		{name: "121 supplementary", label: strings.Repeat("😀", 121), wantError: true},
		{name: "malformed", label: string([]byte{0xff}), wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := (importReviewRequestDTO{
				SourceAccountLabel: tc.label,
				CSVPayload:         "x",
			}).validate()
			if tc.wantError {
				if !errors.Is(err, verticalslice.ErrInvalidInput) {
					t.Fatalf("expected invalid input, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected label to be admitted, got %v", err)
			}
		})
	}

	if err := (importReviewRequestDTO{CSVPayload: strings.Repeat("x", maxHTTPImportPayloadBytes)}).validate(); err != nil {
		t.Fatalf("exact 2 MiB byte payload must be accepted by the DTO gate: %v", err)
	}
	if err := (importReviewRequestDTO{CSVPayload: strings.Repeat("x", maxHTTPImportPayloadBytes+1)}).validate(); !errors.Is(err, verticalslice.ErrInvalidInput) {
		t.Fatalf("payload over 2 MiB bytes must be rejected, got %v", err)
	}
}

func TestStage339AssetSearchEnforcesRawCodePointBoundBeforeTrim(t *testing.T) {
	app := NewDevelopment(verticalslice.NewService(&importAPITestStore{}, fixedHTTPClock{}))
	for _, tc := range []struct {
		name       string
		query      string
		wantStatus int
	}{
		{name: "100 supplementary", query: strings.Repeat("😀", 100), wantStatus: http.StatusOK},
		{name: "101 supplementary", query: strings.Repeat("😀", 101), wantStatus: http.StatusBadRequest},
		{name: "raw 101 trims to 100", query: " " + strings.Repeat("A", 100), wantStatus: http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/api/v1/assets/search?query="+url.QueryEscape(tc.query), nil)
			response, err := app.Test(request)
			if err != nil {
				t.Fatalf("asset search request: %v", err)
			}
			defer response.Body.Close()
			if response.StatusCode != tc.wantStatus {
				body, _ := io.ReadAll(response.Body)
				t.Fatalf("status: got %d want %d body=%s", response.StatusCode, tc.wantStatus, body)
			}
		})
	}
}

type stage39HTTPReplayLookupStore struct {
	stage32HTTPReplayStore
	lookupCalls int
}

func (store *stage39HTTPReplayLookupStore) LookupReplayArtifact(
	_ context.Context,
	command verticalslice.CommandContext,
	method string,
) (verticalslice.CommandReplayArtifact, bool, error) {
	store.lookupCalls++
	if method != http.MethodPost {
		return verticalslice.CommandReplayArtifact{}, false, errors.New("unexpected method")
	}
	if store.createRequestHash == "" {
		return verticalslice.CommandReplayArtifact{}, false, nil
	}
	if command.RequestHash != store.createRequestHash {
		return verticalslice.CommandReplayArtifact{}, false, errors.New("unexpected historical request hash")
	}
	return cloneStage0332ReplayArtifact(store.createArtifact), true, nil
}

func TestStage339PortfolioHTTPHistoricalCompatibilityReplaysExactArtifact(t *testing.T) {
	store := &stage39HTTPReplayLookupStore{}
	app := NewDevelopmentReplay(verticalslice.NewService(store, fixedHTTPClock{}))
	key := "stage-03-39-http-history-key-01"
	normalizedName := strings.Repeat("A", 100)

	firstBody, err := json.Marshal(createPortfolioRequestDTO{Name: normalizedName, BaseCurrency: verticalslice.RUB})
	if err != nil {
		t.Fatalf("marshal first body: %v", err)
	}
	first := httptest.NewRequest(http.MethodPost, "/api/v1/portfolios", bytes.NewReader(firstBody))
	first.Header.Set("Content-Type", "application/json")
	first.Header.Set("Idempotency-Key", key)
	first.Header.Set("X-Request-ID", "11111111-1111-4111-8111-111111111111")
	first.Header.Set("traceparent", "00-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-bbbbbbbbbbbbbbbb-01")
	firstResponse, err := app.Test(first)
	if err != nil {
		t.Fatalf("first request: %v", err)
	}
	defer firstResponse.Body.Close()
	firstResponseBody, err := io.ReadAll(firstResponse.Body)
	if err != nil {
		t.Fatalf("read first body: %v", err)
	}
	if firstResponse.StatusCode != http.StatusCreated {
		t.Fatalf("first status: got %d body=%s", firstResponse.StatusCode, firstResponseBody)
	}

	legacyRawBody, err := json.Marshal(createPortfolioRequestDTO{Name: " " + normalizedName, BaseCurrency: verticalslice.RUB})
	if err != nil {
		t.Fatalf("marshal legacy body: %v", err)
	}
	retry := httptest.NewRequest(http.MethodPost, "/api/v1/portfolios", bytes.NewReader(legacyRawBody))
	retry.Header.Set("Content-Type", "application/json")
	retry.Header.Set("Idempotency-Key", key)
	retry.Header.Set("X-Request-ID", "22222222-2222-4222-8222-222222222222")
	retry.Header.Set("traceparent", "00-cccccccccccccccccccccccccccccccc-dddddddddddddddd-01")
	retryResponse, err := app.Test(retry)
	if err != nil {
		t.Fatalf("historical retry: %v", err)
	}
	defer retryResponse.Body.Close()
	retryBody, err := io.ReadAll(retryResponse.Body)
	if err != nil {
		t.Fatalf("read retry body: %v", err)
	}

	if retryResponse.StatusCode != firstResponse.StatusCode || !bytes.Equal(retryBody, firstResponseBody) {
		t.Fatalf("historical exact response changed: first=%d %s retry=%d %s",
			firstResponse.StatusCode, firstResponseBody, retryResponse.StatusCode, retryBody)
	}
	if retryResponse.Header.Get("X-Request-ID") != firstResponse.Header.Get("X-Request-ID") ||
		retryResponse.Header.Get("X-Trace-ID") != firstResponse.Header.Get("X-Trace-ID") {
		t.Fatalf("historical technical identity changed")
	}
	if store.lookupCalls != 1 {
		t.Fatalf("expected one read-only compatibility lookup, got %d", store.lookupCalls)
	}
	if store.createBuilderCalls != 1 {
		t.Fatalf("historical retry created a second business response/effect; builder calls=%d", store.createBuilderCalls)
	}
}
