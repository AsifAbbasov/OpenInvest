package httpapi

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestStage0332PortfolioHTTPReplayReturnsOriginalStatusBodyAndTechnicalIdentity(t *testing.T) {
	store := &stage32HTTPReplayStore{}
	app := NewDevelopmentReplay(verticalslice.NewService(store, fixedHTTPClock{}))
	body := []byte(`{"name":"Replay portfolio","baseCurrency":"RUB"}`)
	key := "stage-03-32-http-key-000001"

	firstRequestID := "11111111-1111-4111-8111-111111111111"
	firstTraceID := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	firstRequest := httptest.NewRequest(http.MethodPost, "/api/v1/portfolios", bytes.NewReader(body))
	firstRequest.Header.Set("Content-Type", "application/json")
	firstRequest.Header.Set("Idempotency-Key", key)
	firstRequest.Header.Set("X-Request-ID", firstRequestID)
	firstRequest.Header.Set("traceparent", "00-"+firstTraceID+"-bbbbbbbbbbbbbbbb-01")

	firstResponse, err := app.Test(firstRequest)
	if err != nil {
		t.Fatalf("first portfolio request: %v", err)
	}
	defer firstResponse.Body.Close()
	firstResponseBody, err := io.ReadAll(firstResponse.Body)
	if err != nil {
		t.Fatalf("read first response: %v", err)
	}
	if firstResponse.StatusCode != http.StatusCreated {
		t.Fatalf("first status: got %d want %d body=%s", firstResponse.StatusCode, http.StatusCreated, firstResponseBody)
	}
	if firstResponse.Header.Get("X-Request-ID") != firstRequestID || firstResponse.Header.Get("X-Trace-ID") != firstTraceID {
		t.Fatalf("first technical identity mismatch: request=%q trace=%q", firstResponse.Header.Get("X-Request-ID"), firstResponse.Header.Get("X-Trace-ID"))
	}

	secondRequest := httptest.NewRequest(http.MethodPost, "/api/v1/portfolios", bytes.NewReader(body))
	secondRequest.Header.Set("Content-Type", "application/json")
	secondRequest.Header.Set("Idempotency-Key", key)
	secondRequest.Header.Set("X-Request-ID", "22222222-2222-4222-8222-222222222222")
	secondRequest.Header.Set("traceparent", "00-cccccccccccccccccccccccccccccccc-dddddddddddddddd-01")

	secondResponse, err := app.Test(secondRequest)
	if err != nil {
		t.Fatalf("replay portfolio request: %v", err)
	}
	defer secondResponse.Body.Close()
	secondResponseBody, err := io.ReadAll(secondResponse.Body)
	if err != nil {
		t.Fatalf("read replay response: %v", err)
	}

	if secondResponse.StatusCode != firstResponse.StatusCode {
		t.Fatalf("replay status changed: first=%d replay=%d", firstResponse.StatusCode, secondResponse.StatusCode)
	}
	if !bytes.Equal(secondResponseBody, firstResponseBody) {
		t.Fatalf("replay body changed:\nfirst=%s\nreplay=%s", firstResponseBody, secondResponseBody)
	}
	if secondResponse.Header.Get("X-Request-ID") != firstResponse.Header.Get("X-Request-ID") ||
		secondResponse.Header.Get("X-Trace-ID") != firstResponse.Header.Get("X-Trace-ID") {
		t.Fatalf(
			"replay technical identity changed: first request=%q trace=%q replay request=%q trace=%q",
			firstResponse.Header.Get("X-Request-ID"),
			firstResponse.Header.Get("X-Trace-ID"),
			secondResponse.Header.Get("X-Request-ID"),
			secondResponse.Header.Get("X-Trace-ID"),
		)
	}
	if store.createBuilderCalls != 1 {
		t.Fatalf("expected exact response builder once, got %d calls", store.createBuilderCalls)
	}
}

type stage32HTTPReplayStore struct {
	importAPITestStore
	createRequestHash  string
	createArtifact     verticalslice.CommandReplayArtifact
	createBuilderCalls int
}

func (store *stage32HTTPReplayStore) CreatePortfolioWithReplay(
	_ context.Context,
	command verticalslice.CommandContext,
	request verticalslice.CreatePortfolioRequest,
	build verticalslice.PortfolioReplayBuilder,
) (verticalslice.Portfolio, verticalslice.CommandReplayArtifact, error) {
	if store.createRequestHash != "" {
		if store.createRequestHash != command.RequestHash {
			return verticalslice.Portfolio{}, verticalslice.CommandReplayArtifact{}, errors.New("idempotency conflict")
		}
		return verticalslice.Portfolio{}, store.createArtifact, nil
	}
	portfolio := verticalslice.Portfolio{
		ID:           "00000000-0000-4000-8000-000000000032",
		Name:         request.Name,
		BaseCurrency: request.BaseCurrency,
		Version:      1,
		CreatedAt:    time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC),
	}
	store.createBuilderCalls++
	artifact, err := build(portfolio)
	if err != nil {
		return verticalslice.Portfolio{}, verticalslice.CommandReplayArtifact{}, err
	}
	store.createRequestHash = command.RequestHash
	store.createArtifact = artifact
	return portfolio, artifact, nil
}

func (store *stage32HTTPReplayStore) AppendTransactionWithReplay(
	context.Context,
	verticalslice.CommandContext,
	verticalslice.AppendTransactionRequest,
	verticalslice.TransactionReplayBuilder,
) (verticalslice.Transaction, verticalslice.CommandReplayArtifact, error) {
	return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, errors.New("not used")
}

func (store *stage32HTTPReplayStore) AppendImportedTransactionsWithReplay(
	context.Context,
	verticalslice.CommandContext,
	verticalslice.AppendImportBatchRequest,
	verticalslice.ImportedTransactionsReplayBuilder,
) ([]verticalslice.Transaction, verticalslice.CommandReplayArtifact, error) {
	return nil, verticalslice.CommandReplayArtifact{}, errors.New("not used")
}
