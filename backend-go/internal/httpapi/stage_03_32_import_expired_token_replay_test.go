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

func TestStage0332CompletedImportReplaysAfterReviewTokenExpires(t *testing.T) {
	store := &stage32ImportReplayStore{}
	service := verticalslice.NewService(store, fixedHTTPClock{})
	secret, err := normalizedImportReviewSecret([]byte("stage-03-32-import-replay-test-secret-32-bytes"))
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
	key := "stage-03-32-import-expiry-key-01"

	firstRequest := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	firstRequest.Header.Set("Content-Type", "application/json")
	firstRequest.Header.Set("Idempotency-Key", key)
	firstRequest.Header.Set("X-Request-ID", "33333333-3333-4333-8333-333333333333")
	firstRequest.Header.Set("traceparent", "00-11111111111111111111111111111111-2222222222222222-01")
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
	if store.appendReplayCalls != 1 {
		t.Fatalf("expected one financial append, got %d", store.appendReplayCalls)
	}
	if store.lookupCalls != 0 {
		t.Fatalf("fresh valid import must not perform recovery lookup, got %d calls", store.lookupCalls)
	}

	// The signed review token has a 15-minute lifetime. Advance beyond it while keeping the
	// command inside the 24-hour idempotency replay window.
	now = now.Add(20 * time.Minute)
	secondRequest := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	secondRequest.Header.Set("Content-Type", "application/json")
	secondRequest.Header.Set("Idempotency-Key", key)
	secondRequest.Header.Set("X-Request-ID", "44444444-4444-4444-8444-444444444444")
	secondRequest.Header.Set("traceparent", "00-33333333333333333333333333333333-4444444444444444-01")
	secondResponse, err := app.Test(secondRequest)
	if err != nil {
		t.Fatalf("expired-token import replay: %v", err)
	}
	defer secondResponse.Body.Close()
	secondBody, err := io.ReadAll(secondResponse.Body)
	if err != nil {
		t.Fatalf("read expired-token replay response: %v", err)
	}

	if secondResponse.StatusCode != firstResponse.StatusCode {
		t.Fatalf("replayed import status changed: first=%d replay=%d", firstResponse.StatusCode, secondResponse.StatusCode)
	}
	if !bytes.Equal(secondBody, firstBody) {
		t.Fatalf("replayed import body changed:\nfirst=%s\nreplay=%s", firstBody, secondBody)
	}
	if secondResponse.Header.Get("X-Request-ID") != firstResponse.Header.Get("X-Request-ID") ||
		secondResponse.Header.Get("X-Trace-ID") != firstResponse.Header.Get("X-Trace-ID") {
		t.Fatalf("replayed import technical identity changed")
	}
	if store.appendReplayCalls != 1 {
		t.Fatalf("expired-token replay executed financial append again: calls=%d", store.appendReplayCalls)
	}
	if store.lookupCalls != 1 {
		t.Fatalf("expected one recovery lookup only after proof expiry, got %d", store.lookupCalls)
	}
}

type stage32ImportReplayStore struct {
	importAPITestStore
	requestHash       string
	subjectID         string
	idempotencyKey    string
	requestPath       string
	artifact          verticalslice.CommandReplayArtifact
	appendReplayCalls int
	lookupCalls       int
}

func (store *stage32ImportReplayStore) LookupReplayArtifact(
	_ context.Context,
	command verticalslice.CommandContext,
	_ string,
) (verticalslice.CommandReplayArtifact, bool, error) {
	store.lookupCalls++
	if store.requestHash == "" {
		return verticalslice.CommandReplayArtifact{}, false, nil
	}
	if store.requestHash != command.RequestHash ||
		store.subjectID != command.SubjectID ||
		store.idempotencyKey != command.IdempotencyKey ||
		store.requestPath != command.RequestPath {
		return verticalslice.CommandReplayArtifact{}, false, nil
	}
	return cloneStage0332ReplayArtifact(store.artifact), true, nil
}

func (store *stage32ImportReplayStore) CreatePortfolioWithReplay(
	context.Context,
	verticalslice.CommandContext,
	verticalslice.CreatePortfolioRequest,
	verticalslice.PortfolioReplayBuilder,
) (verticalslice.Portfolio, verticalslice.CommandReplayArtifact, error) {
	return verticalslice.Portfolio{}, verticalslice.CommandReplayArtifact{}, errors.New("not used")
}

func (store *stage32ImportReplayStore) AppendTransactionWithReplay(
	context.Context,
	verticalslice.CommandContext,
	verticalslice.AppendTransactionRequest,
	verticalslice.TransactionReplayBuilder,
) (verticalslice.Transaction, verticalslice.CommandReplayArtifact, error) {
	return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, errors.New("not used")
}

func (store *stage32ImportReplayStore) AppendImportedTransactionsWithReplay(
	ctx context.Context,
	command verticalslice.CommandContext,
	request verticalslice.AppendImportBatchRequest,
	build verticalslice.ImportedTransactionsReplayBuilder,
) ([]verticalslice.Transaction, verticalslice.CommandReplayArtifact, error) {
	return store.AppendImportedTransactionsWithOutcomeReplay(
		ctx,
		command,
		request,
		func(outcome verticalslice.ImportAppendOutcome) (verticalslice.CommandReplayArtifact, error) {
			return build(outcome.Transactions)
		},
	)
}

func (store *stage32ImportReplayStore) AppendImportedTransactionsWithOutcomeReplay(
	_ context.Context,
	command verticalslice.CommandContext,
	request verticalslice.AppendImportBatchRequest,
	build verticalslice.ImportedTransactionsOutcomeReplayBuilder,
) ([]verticalslice.Transaction, verticalslice.CommandReplayArtifact, error) {
	if store.requestHash != "" {
		return nil, verticalslice.CommandReplayArtifact{}, errors.New("financial append was executed more than once")
	}
	store.appendReplayCalls++
	transactions := make([]verticalslice.Transaction, 0, len(request.Transactions))
	snapshotDates := make([]string, 0, len(request.Transactions))
	for index, item := range request.Transactions {
		transactions = append(transactions, verticalslice.Transaction{
			ID:              "stage-03-32-import-transaction-" + string(rune('a'+index)),
			PortfolioID:     request.PortfolioID,
			TransactionType: item.TransactionType,
			TradeDate:       item.TradeDate,
		})
		snapshotDates = append(snapshotDates, item.TradeDate)
	}
	artifact, err := build(verticalslice.ImportAppendOutcome{
		Transactions:         transactions,
		SnapshotDatesRebuilt: snapshotDates,
	})
	if err != nil {
		return nil, verticalslice.CommandReplayArtifact{}, err
	}
	store.requestHash = command.RequestHash
	store.subjectID = command.SubjectID
	store.idempotencyKey = command.IdempotencyKey
	store.requestPath = command.RequestPath
	store.artifact = cloneStage0332ReplayArtifact(artifact)
	return transactions, artifact, nil
}
