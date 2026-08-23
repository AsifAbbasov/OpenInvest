package postgres

import (
	"bytes"
	"context"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestStage0332CreatePortfolioReplayIsExactAndDoesNotReexecute(t *testing.T) {
	store := openStage0332Store(t)
	ctx := context.Background()
	subjectID := uuid.NewString()
	cleanupStage0332Subject(t, store, subjectID)
	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	request := verticalslice.CreatePortfolioRequest{Name: "Replay portfolio", BaseCurrency: verticalslice.RUB}
	key := "stage-03-32-create-key-000001"
	path := "/api/v1/portfolios"
	firstRequestID := uuid.NewString()
	firstTraceID := "11111111111111111111111111111111"
	firstBody := []byte(`{"data":{"id":"original"},"meta":{"requestId":"original"}}`)
	builderCalls := 0

	portfolio, firstArtifact, err := service.CreatePortfolioWithReplay(
		ctx,
		verticalslice.RequestContext{RequestID: firstRequestID, TraceID: firstTraceID},
		subjectID,
		key,
		path,
		request,
		func(verticalslice.Portfolio) (verticalslice.CommandReplayArtifact, error) {
			builderCalls++
			return verticalslice.CommandReplayArtifact{
				StatusCode: 201,
				Body:       append([]byte(nil), firstBody...),
				RequestID:  firstRequestID,
				TraceID:    firstTraceID,
			}, nil
		},
	)
	if err != nil {
		t.Fatalf("first create with replay: %v", err)
	}
	if portfolio.ID == "" {
		t.Fatal("first create returned an empty portfolio id")
	}
	if !bytes.Equal(firstArtifact.Body, firstBody) {
		t.Fatalf("first replay body mismatch: got %s want %s", firstArtifact.Body, firstBody)
	}

	secondRequestID := uuid.NewString()
	secondTraceID := "22222222222222222222222222222222"
	_, replayArtifact, err := service.CreatePortfolioWithReplay(
		ctx,
		verticalslice.RequestContext{RequestID: secondRequestID, TraceID: secondTraceID},
		subjectID,
		key,
		path,
		request,
		func(verticalslice.Portfolio) (verticalslice.CommandReplayArtifact, error) {
			builderCalls++
			return verticalslice.CommandReplayArtifact{
				StatusCode: 201,
				Body:       []byte(`{"unexpected":"recomputed"}`),
				RequestID:  secondRequestID,
				TraceID:    secondTraceID,
			}, nil
		},
	)
	if err != nil {
		t.Fatalf("replay create: %v", err)
	}
	if builderCalls != 1 {
		t.Fatalf("expected response builder to run once, ran %d times", builderCalls)
	}
	if replayArtifact.StatusCode != firstArtifact.StatusCode ||
		replayArtifact.RequestID != firstArtifact.RequestID ||
		replayArtifact.TraceID != firstArtifact.TraceID ||
		!bytes.Equal(replayArtifact.Body, firstArtifact.Body) {
		t.Fatalf("replay artifact changed: first=%+v replay=%+v", firstArtifact, replayArtifact)
	}

	assertStage0332Count(t, store, `SELECT count(*) FROM investment.portfolios WHERE subject_id = $1`, 1, subjectID)
	assertStage0332Count(t, store, `SELECT count(*) FROM investment.command_deduplication WHERE principal_id = $1 AND idempotency_key = $2`, 1, subjectID, key)
}

func TestStage0332ReplayRejectsDifferentPayload(t *testing.T) {
	store := openStage0332Store(t)
	ctx := context.Background()
	subjectID := uuid.NewString()
	cleanupStage0332Subject(t, store, subjectID)
	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	key := "stage-03-32-conflict-key-0001"
	requestContext := verticalslice.RequestContext{RequestID: uuid.NewString(), TraceID: "33333333333333333333333333333333"}

	_, _, err := service.CreatePortfolioWithReplay(
		ctx,
		requestContext,
		subjectID,
		key,
		"/api/v1/portfolios",
		verticalslice.CreatePortfolioRequest{Name: "First", BaseCurrency: verticalslice.RUB},
		stage0332ArtifactBuilder(requestContext, []byte(`{"first":true}`)),
	)
	if err != nil {
		t.Fatalf("first create: %v", err)
	}

	_, _, err = service.CreatePortfolioWithReplay(
		ctx,
		verticalslice.RequestContext{RequestID: uuid.NewString(), TraceID: "44444444444444444444444444444444"},
		subjectID,
		key,
		"/api/v1/portfolios",
		verticalslice.CreatePortfolioRequest{Name: "Different payload", BaseCurrency: verticalslice.RUB},
		stage0332ArtifactBuilder(requestContext, []byte(`{"unexpected":true}`)),
	)
	if !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("expected ErrIdempotencyConflict, got %v", err)
	}
	assertStage0332Count(t, store, `SELECT count(*) FROM investment.portfolios WHERE subject_id = $1`, 1, subjectID)
}

func TestStage0332ReplayArtifactFailureRollsBackBusinessWrite(t *testing.T) {
	store := openStage0332Store(t)
	ctx := context.Background()
	subjectID := uuid.NewString()
	cleanupStage0332Subject(t, store, subjectID)
	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	key := "stage-03-32-rollback-key-0001"
	request := verticalslice.CreatePortfolioRequest{Name: "Rollback portfolio", BaseCurrency: verticalslice.RUB}

	_, _, err := service.CreatePortfolioWithReplay(
		ctx,
		verticalslice.RequestContext{RequestID: uuid.NewString(), TraceID: "55555555555555555555555555555555"},
		subjectID,
		key,
		"/api/v1/portfolios",
		request,
		func(verticalslice.Portfolio) (verticalslice.CommandReplayArtifact, error) {
			return verticalslice.CommandReplayArtifact{}, errors.New("response serialization failed")
		},
	)
	if err == nil {
		t.Fatal("expected replay builder failure")
	}
	assertStage0332Count(t, store, `SELECT count(*) FROM investment.portfolios WHERE subject_id = $1`, 0, subjectID)
	assertStage0332Count(t, store, `SELECT count(*) FROM investment.command_deduplication WHERE principal_id = $1`, 0, subjectID)

	requestContext := verticalslice.RequestContext{RequestID: uuid.NewString(), TraceID: "66666666666666666666666666666666"}
	portfolio, _, err := service.CreatePortfolioWithReplay(
		ctx,
		requestContext,
		subjectID,
		key,
		"/api/v1/portfolios",
		request,
		stage0332ArtifactBuilder(requestContext, []byte(`{"retry":"succeeded"}`)),
	)
	if err != nil {
		t.Fatalf("retry after rollback: %v", err)
	}
	if portfolio.ID == "" {
		t.Fatal("retry after rollback did not create portfolio")
	}
}

func TestStage0332TransactionReplayDoesNotDuplicateLedger(t *testing.T) {
	store := openStage0332Store(t)
	ctx := context.Background()
	subjectID := uuid.NewString()
	cleanupStage0332Subject(t, store, subjectID)
	service := verticalslice.NewService(store, verticalslice.SystemClock{})

	portfolioContext := verticalslice.RequestContext{RequestID: uuid.NewString(), TraceID: "77777777777777777777777777777777"}
	portfolio, _, err := service.CreatePortfolioWithReplay(
		ctx,
		portfolioContext,
		subjectID,
		"stage-03-32-portfolio-key-0001",
		"/api/v1/portfolios",
		verticalslice.CreatePortfolioRequest{Name: "Ledger replay", BaseCurrency: verticalslice.RUB},
		stage0332ArtifactBuilder(portfolioContext, []byte(`{"portfolio":true}`)),
	)
	if err != nil {
		t.Fatalf("create portfolio: %v", err)
	}

	gross := verticalslice.Money{Amount: decimal.Must("1000.00000000"), Currency: verticalslice.RUB}
	request := verticalslice.AppendTransactionRequest{
		PortfolioID:     portfolio.ID,
		TransactionType: "DEPOSIT",
		GrossAmount:     &gross,
		Commission:      verticalslice.ZeroMoney(),
		Tax:             verticalslice.ZeroMoney(),
		TradeDate:       "2026-08-23",
		SettlementDate:  nil,
	}
	key := "stage-03-32-transaction-key-01"
	path := "/api/v1/portfolios/" + portfolio.ID + "/transactions"
	firstRequestID := uuid.NewString()
	firstTraceID := "88888888888888888888888888888888"
	builderCalls := 0

	transaction, firstArtifact, err := service.AppendTransactionWithReplay(
		ctx,
		verticalslice.RequestContext{RequestID: firstRequestID, TraceID: firstTraceID},
		subjectID,
		key,
		path,
		request,
		func(verticalslice.Transaction) (verticalslice.CommandReplayArtifact, error) {
			builderCalls++
			return verticalslice.CommandReplayArtifact{
				StatusCode: 201,
				Body:       []byte(`{"transaction":"original"}`),
				RequestID:  firstRequestID,
				TraceID:    firstTraceID,
			}, nil
		},
	)
	if err != nil {
		t.Fatalf("first append transaction: %v", err)
	}
	if transaction.ID == "" {
		t.Fatal("first append returned empty transaction id")
	}

	_, replayArtifact, err := service.AppendTransactionWithReplay(
		ctx,
		verticalslice.RequestContext{RequestID: uuid.NewString(), TraceID: "99999999999999999999999999999999"},
		subjectID,
		key,
		path,
		request,
		func(verticalslice.Transaction) (verticalslice.CommandReplayArtifact, error) {
			builderCalls++
			return verticalslice.CommandReplayArtifact{}, errors.New("duplicate replay must not rebuild response")
		},
	)
	if err != nil {
		t.Fatalf("replay append transaction: %v", err)
	}
	if builderCalls != 1 {
		t.Fatalf("expected transaction response builder once, ran %d times", builderCalls)
	}
	if !bytes.Equal(firstArtifact.Body, replayArtifact.Body) || firstArtifact.RequestID != replayArtifact.RequestID {
		t.Fatalf("transaction replay artifact changed: first=%+v replay=%+v", firstArtifact, replayArtifact)
	}
	assertStage0332Count(t, store, `SELECT count(*) FROM investment.transaction_entries WHERE portfolio_id = $1`, 1, portfolio.ID)
}

func openStage0332Store(t *testing.T) *Store {
	t.Helper()
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}
	store, err := Open(databaseURL)
	if err != nil {
		t.Fatalf("open postgres store: %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Fatalf("close postgres store: %v", err)
		}
	})
	return store
}

func stage0332ArtifactBuilder(requestContext verticalslice.RequestContext, body []byte) verticalslice.PortfolioReplayBuilder {
	return func(verticalslice.Portfolio) (verticalslice.CommandReplayArtifact, error) {
		return verticalslice.CommandReplayArtifact{
			StatusCode: 201,
			Body:       append([]byte(nil), body...),
			RequestID:  requestContext.RequestID,
			TraceID:    requestContext.TraceID,
		}, nil
	}
}

func assertStage0332Count(t *testing.T, store *Store, query string, want int, args ...any) {
	t.Helper()
	var got int
	if err := store.db.QueryRowContext(context.Background(), query, args...).Scan(&got); err != nil {
		t.Fatalf("count query: %v", err)
	}
	if got != want {
		t.Fatalf("count mismatch: got %d want %d", got, want)
	}
}

func cleanupStage0332Subject(t *testing.T, store *Store, subjectID string) {
	t.Helper()
	t.Cleanup(func() {
		ctx := context.Background()
		statements := []string{
			`DELETE FROM analytics.snapshot_positions WHERE snapshot_id IN (SELECT id FROM analytics.portfolio_snapshots WHERE portfolio_id IN (SELECT id FROM investment.portfolios WHERE subject_id = $1))`,
			`DELETE FROM analytics.portfolio_snapshots WHERE portfolio_id IN (SELECT id FROM investment.portfolios WHERE subject_id = $1)`,
			`DELETE FROM analytics.calculation_runs WHERE portfolio_id IN (SELECT id FROM investment.portfolios WHERE subject_id = $1)`,
			`DELETE FROM investment.transaction_entries WHERE portfolio_id IN (SELECT id FROM investment.portfolios WHERE subject_id = $1)`,
			`DELETE FROM audit.events WHERE actor_id = $1`,
			`DELETE FROM audit.actors WHERE id = $1`,
			`DELETE FROM investment.command_deduplication WHERE principal_id = $1`,
			`DELETE FROM investment.portfolios WHERE subject_id = $1`,
			`DELETE FROM investment.subjects WHERE id = $1`,
		}
		for _, statement := range statements {
			if _, err := store.db.ExecContext(ctx, statement, subjectID); err != nil {
				t.Errorf("cleanup Stage 3.32 subject %s: %v", subjectID, err)
			}
		}
	})
}
