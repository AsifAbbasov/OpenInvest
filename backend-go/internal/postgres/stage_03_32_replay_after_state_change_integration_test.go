package postgres

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestStage0332CompletedTransactionReplaysAfterPortfolioStateChanges(t *testing.T) {
	store := openStage0332Store(t)
	ctx := context.Background()
	subjectID := uuid.NewString()
	cleanupStage0332Subject(t, store, subjectID)
	service := verticalslice.NewService(store, verticalslice.SystemClock{})

	portfolioContext := verticalslice.RequestContext{RequestID: uuid.NewString(), TraceID: "11111111111111111111111111111111"}
	portfolio, _, err := service.CreatePortfolioWithReplay(
		ctx,
		portfolioContext,
		subjectID,
		"stage-03-32-state-portfolio-key",
		"/api/v1/portfolios",
		verticalslice.CreatePortfolioRequest{Name: "State-change replay", BaseCurrency: verticalslice.RUB},
		stage0332ArtifactBuilder(portfolioContext, []byte(`{"portfolio":"created"}`)),
	)
	if err != nil {
		t.Fatalf("create portfolio: %v", err)
	}

	gross := verticalslice.Money{Amount: decimal.Must("500.00000000"), Currency: verticalslice.RUB}
	request := verticalslice.AppendTransactionRequest{
		PortfolioID:     portfolio.ID,
		TransactionType: "DEPOSIT",
		GrossAmount:     &gross,
		Commission:      verticalslice.ZeroMoney(),
		Tax:             verticalslice.ZeroMoney(),
		TradeDate:       "2026-08-23",
	}
	key := "stage-03-32-state-transaction-key"
	path := "/api/v1/portfolios/" + portfolio.ID + "/transactions"
	firstContext := verticalslice.RequestContext{RequestID: uuid.NewString(), TraceID: "22222222222222222222222222222222"}
	builderCalls := 0
	_, firstArtifact, err := service.AppendTransactionWithReplay(
		ctx,
		firstContext,
		subjectID,
		key,
		path,
		request,
		func(verticalslice.Transaction) (verticalslice.CommandReplayArtifact, error) {
			builderCalls++
			return verticalslice.CommandReplayArtifact{
				StatusCode: 201,
				Body:       []byte(`{"transaction":"original"}`),
				RequestID:  firstContext.RequestID,
				TraceID:    firstContext.TraceID,
			}, nil
		},
	)
	if err != nil {
		t.Fatalf("first append: %v", err)
	}

	if _, err := store.db.ExecContext(ctx, `
		UPDATE investment.portfolios
		SET portfolio_state = 'removed_from_active_use', removed_at = now(), updated_at = now()
		WHERE id = $1
	`, portfolio.ID); err != nil {
		t.Fatalf("remove portfolio from active use: %v", err)
	}

	_, replayArtifact, err := service.AppendTransactionWithReplay(
		ctx,
		verticalslice.RequestContext{RequestID: uuid.NewString(), TraceID: "33333333333333333333333333333333"},
		subjectID,
		key,
		path,
		request,
		func(verticalslice.Transaction) (verticalslice.CommandReplayArtifact, error) {
			builderCalls++
			return verticalslice.CommandReplayArtifact{}, errors.New("completed replay must not consult current portfolio state")
		},
	)
	if err != nil {
		t.Fatalf("replay after portfolio state change: %v", err)
	}
	if builderCalls != 1 {
		t.Fatalf("expected response builder once, ran %d times", builderCalls)
	}
	if replayArtifact.StatusCode != firstArtifact.StatusCode ||
		replayArtifact.RequestID != firstArtifact.RequestID ||
		replayArtifact.TraceID != firstArtifact.TraceID ||
		!bytes.Equal(replayArtifact.Body, firstArtifact.Body) {
		t.Fatalf("replay changed after mutable portfolio state change: first=%+v replay=%+v", firstArtifact, replayArtifact)
	}
	assertStage0332Count(t, store, `SELECT count(*) FROM investment.transaction_entries WHERE portfolio_id = $1`, 1, portfolio.ID)
}
