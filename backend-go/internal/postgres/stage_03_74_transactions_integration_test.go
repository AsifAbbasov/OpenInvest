package postgres_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func stage374Artifact(requestContext verticalslice.RequestContext, transaction verticalslice.Transaction) (verticalslice.CommandReplayArtifact, error) {
	return verticalslice.CommandReplayArtifact{
		StatusCode: 200,
		Body:       []byte(transaction.ID),
		RequestID:  requestContext.RequestID,
		TraceID:    requestContext.TraceID,
	}, nil
}

func correctStage374(t *testing.T, h stage371Harness, transaction verticalslice.Transaction, quantity, price, tradeDate, key string) (verticalslice.Transaction, error) {
	t.Helper()
	q := decimal.Must(quantity)
	p := verticalslice.Money{Amount: decimal.Must(price), Currency: verticalslice.RUB}
	corrected := verticalslice.AppendTransactionRequest{
		PortfolioID:     h.portfolioID,
		TransactionType: transaction.TransactionType,
		Ticker:          transaction.Ticker,
		Quantity:        &q,
		UnitPrice:       &p,
		Commission:      transaction.Commission,
		Tax:             transaction.Tax,
		TradeDate:       tradeDate,
		SettlementDate:  transaction.SettlementDate,
		Note:            transaction.Note,
	}
	requestContext := verticalslice.RequestContext{RequestID: uuid.NewString(), TraceID: "stage-03-74-test"}
	result, _, err := h.service.CorrectTransactionWithReplay(
		h.ctx, requestContext, h.subjectID, key,
		"/api/v1/portfolios/"+h.portfolioID+"/transactions/"+transaction.ID,
		verticalslice.CorrectTransactionRequest{
			PortfolioID:      h.portfolioID,
			TransactionID:    transaction.ID,
			ExpectedRevision: transaction.Revision,
			Reason:           "Incorrect broker value",
			Corrected:        corrected,
		},
		func(result verticalslice.Transaction) (verticalslice.CommandReplayArtifact, error) {
			return stage374Artifact(requestContext, result)
		},
	)
	return result, err
}

func reverseStage374(t *testing.T, h stage371Harness, transaction verticalslice.Transaction, effectiveDate, key string) (verticalslice.TransactionReversal, error) {
	t.Helper()
	requestContext := verticalslice.RequestContext{RequestID: uuid.NewString(), TraceID: "stage-03-74-test"}
	result, _, err := h.service.ReverseTransactionWithReplay(
		h.ctx, requestContext, h.subjectID, key,
		"/api/v1/portfolios/"+h.portfolioID+"/transactions/"+transaction.ID,
		verticalslice.ReverseTransactionRequest{
			PortfolioID:      h.portfolioID,
			TransactionID:    transaction.ID,
			ExpectedRevision: transaction.Revision,
			Reason:           "Broker cancelled transaction",
			EffectiveDate:    effectiveDate,
		},
		func(result verticalslice.TransactionReversal) (verticalslice.CommandReplayArtifact, error) {
			return verticalslice.CommandReplayArtifact{
				StatusCode: 200,
				Body:       []byte(result.ReversalTransactionID),
				RequestID:  requestContext.RequestID,
				TraceID:    requestContext.TraceID,
			}, nil
		},
	)
	return result, err
}

func TestStage374CorrectionRebuildsCurrentAndHistoricalPositions(t *testing.T) {
	h := newStage371Harness(t, "Stage 3.74 correction")
	original := appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "100.00000000", "300.00000000", "2026-03-01"))
	corrected, err := correctStage374(t, h, original, "100.00000000", "280.00000000", "2026-03-01", uuid.NewString())
	if err != nil {
		t.Fatalf("correct transaction: %v", err)
	}
	if corrected.ID != original.ID || corrected.Revision != 2 || corrected.Status != "CORRECTED" {
		t.Fatalf("correction projection mismatch: %+v", corrected)
	}
	current, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "")
	if err != nil {
		t.Fatalf("current positions: %v", err)
	}
	if len(current.Items) != 1 || current.Items[0].WeightedAverageCost.Amount.String() != "280.00000000" || current.Items[0].AcquisitionBasis.Amount.String() != "28000.00000000" {
		t.Fatalf("corrected current position mismatch: %+v", current.Items)
	}
	before, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "2026-02-01")
	if err != nil {
		t.Fatalf("historical before trade: %v", err)
	}
	if len(before.Items) != 0 {
		t.Fatalf("position must not exist before corrected trade date: %+v", before.Items)
	}
	after, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "2026-03-15")
	if err != nil {
		t.Fatalf("historical after trade: %v", err)
	}
	if len(after.Items) != 1 || after.Items[0].WeightedAverageCost.Amount.String() != "280.00000000" {
		t.Fatalf("historical correction not reflected: %+v", after.Items)
	}
	assertContiguousLedgerSequence(t, h, 2)
}

func TestStage374SecondCorrectionAndStaleRevisionConflict(t *testing.T) {
	h := newStage371Harness(t, "Stage 3.74 revisions")
	original := appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "10.00000000", "100.00000000", "2026-01-01"))
	revision2, err := correctStage374(t, h, original, "10.00000000", "110.00000000", "2026-01-01", uuid.NewString())
	if err != nil {
		t.Fatalf("first correction: %v", err)
	}
	revision3, err := correctStage374(t, h, revision2, "10.00000000", "120.00000000", "2026-01-01", uuid.NewString())
	if err != nil {
		t.Fatalf("second correction: %v", err)
	}
	if revision3.Revision != 3 {
		t.Fatalf("revision got %d want 3", revision3.Revision)
	}
	_, err = correctStage374(t, h, original, "10.00000000", "130.00000000", "2026-01-01", uuid.NewString())
	if !errors.Is(err, verticalslice.ErrTransactionConflict) {
		t.Fatalf("stale revision error = %v", err)
	}
}

func TestStage374CorrectionOversellRollsBackAtomically(t *testing.T) {
	h := newStage371Harness(t, "Stage 3.74 oversell correction")
	buy := appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "100.00000000", "100.00000000", "2026-01-01"))
	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "SELL", "SBER", "80.00000000", "150.00000000", "2026-02-01"))
	_, err := correctStage374(t, h, buy, "50.00000000", "100.00000000", "2026-01-01", uuid.NewString())
	if !errors.Is(err, verticalslice.ErrInsufficientPositionQuantity) {
		t.Fatalf("oversell correction error = %v", err)
	}
	var revisions int
	if err := h.db.QueryRowContext(h.ctx, `SELECT COUNT(*) FROM investment.transaction_entries WHERE portfolio_id=$1 AND transaction_id=$2`, h.portfolioID, buy.ID).Scan(&revisions); err != nil {
		t.Fatal(err)
	}
	if revisions != 1 {
		t.Fatalf("invalid correction must rollback, rows=%d", revisions)
	}
}

func TestStage374ReversalUsesEffectiveDateForTimeMachine(t *testing.T) {
	h := newStage371Harness(t, "Stage 3.74 reversal time machine")
	buy := appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "10.00000000", "100.00000000", "2026-01-10"))
	reversal, err := reverseStage374(t, h, buy, "2026-03-01", uuid.NewString())
	if err != nil {
		t.Fatalf("reverse transaction: %v", err)
	}
	if reversal.Status != "REVERSED" || reversal.TransactionID != buy.ID {
		t.Fatalf("reversal mismatch: %+v", reversal)
	}
	before, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "2026-02-15")
	if err != nil {
		t.Fatalf("before reversal: %v", err)
	}
	if len(before.Items) != 1 || before.Items[0].Quantity.String() != "10.00000000" {
		t.Fatalf("historically active before reversal: %+v", before.Items)
	}
	after, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "2026-03-01")
	if err != nil {
		t.Fatalf("after reversal: %v", err)
	}
	if len(after.Items) != 0 {
		t.Fatalf("reversal must be effective on selected date: %+v", after.Items)
	}
	current, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "")
	if err != nil {
		t.Fatalf("current after reversal: %v", err)
	}
	if len(current.Items) != 0 {
		t.Fatalf("current must exclude reversed transaction: %+v", current.Items)
	}
}

func TestStage374ReversingBuyThatCreatesOversellRollsBack(t *testing.T) {
	h := newStage371Harness(t, "Stage 3.74 oversell reversal")
	buy := appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "100.00000000", "100.00000000", "2026-01-01"))
	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "SELL", "SBER", "80.00000000", "150.00000000", "2026-02-01"))
	_, err := reverseStage374(t, h, buy, "2026-01-15", uuid.NewString())
	if !errors.Is(err, verticalslice.ErrInsufficientPositionQuantity) {
		t.Fatalf("oversell reversal error = %v", err)
	}
	var reversals int
	if err := h.db.QueryRowContext(h.ctx, `SELECT COUNT(*) FROM investment.transaction_entries WHERE portfolio_id=$1 AND reverses_transaction_id=$2`, h.portfolioID, buy.ID).Scan(&reversals); err != nil {
		t.Fatal(err)
	}
	if reversals != 0 {
		t.Fatalf("invalid reversal must rollback, rows=%d", reversals)
	}
}

func TestStage374ConcurrentCorrectionsOnlyOneWins(t *testing.T) {
	h := newStage371Harness(t, "Stage 3.74 concurrent correction")
	original := appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "10.00000000", "100.00000000", "2026-01-01"))
	start := make(chan struct{})
	errorsOut := make(chan error, 2)
	var wg sync.WaitGroup
	for _, price := range []string{"110.00000000", "120.00000000"} {
		wg.Add(1)
		go func(price string) {
			defer wg.Done()
			<-start
			_, err := correctStage374(t, h, original, "10.00000000", price, "2026-01-01", uuid.NewString())
			errorsOut <- err
		}(price)
	}
	close(start)
	wg.Wait()
	close(errorsOut)
	success, conflicts := 0, 0
	for err := range errorsOut {
		if err == nil {
			success++
		} else if errors.Is(err, verticalslice.ErrTransactionConflict) {
			conflicts++
		} else {
			t.Fatalf("unexpected concurrent error: %v", err)
		}
	}
	if success != 1 || conflicts != 1 {
		t.Fatalf("concurrent results success=%d conflicts=%d", success, conflicts)
	}
}
