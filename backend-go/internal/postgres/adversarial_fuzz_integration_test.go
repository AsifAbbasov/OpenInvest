package postgres_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func fuzzAppendArtifact(transaction verticalslice.Transaction) (verticalslice.CommandReplayArtifact, error) {
	return verticalslice.CommandReplayArtifact{
		StatusCode: 201,
		Body:       []byte(transaction.ID),
		RequestID:  uuid.NewString(),
		TraceID:    "fuzz-postgres-ledger",
	}, nil
}

func FuzzPostgresConcurrentLedgerReplay(f *testing.F) {
	f.Add(uint8(10), uint8(3), uint8(7), uint8(4))
	f.Add(uint8(1), uint8(20), uint8(20), uint8(6))

	f.Fuzz(func(t *testing.T, initialRaw, buyRaw, sellRaw, workersRaw uint8) {
		initialQty := int(initialRaw%20) + 1
		buyQty := int(buyRaw%20) + 1
		sellQty := int(sellRaw%20) + 1
		workers := int(workersRaw%5) + 2

		h := newStage371Harness(t, "fuzz postgres concurrent ledger")
		initial := appendStage371Trade(t, h, stage371Trade(
			h.portfolioID, "BUY", "SBER",
			fmt.Sprintf("%d.00000000", initialQty),
			"100.00000000", "2026-01-01",
		))

		start := make(chan struct{})
		errs := make([]error, workers)
		var wg sync.WaitGroup
		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				<-start
				kind := "BUY"
				qty := buyQty
				if index%2 == 1 {
					kind = "SELL"
					qty = sellQty
				}
				_, errs[index] = h.service.AppendTransaction(
					h.ctx,
					verticalslice.RequestContext{RequestID: uuid.NewString()},
					h.subjectID,
					uuid.NewString(),
					"/api/v1/portfolios/"+h.portfolioID+"/transactions",
					stage371Trade(
						h.portfolioID, kind, "SBER",
						fmt.Sprintf("%d.00000000", qty),
						"101.00000000", "2026-01-02",
					),
				)
			}(i)
		}
		close(start)
		wg.Wait()

		successes := 0
		for _, err := range errs {
			switch {
			case err == nil:
				successes++
			case errors.Is(err, verticalslice.ErrInsufficientPositionQuantity):
			default:
				t.Fatalf("unexpected concurrent BUY/SELL error: %v", err)
			}
		}

		projection, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "")
		if err != nil {
			t.Fatalf("read positions after concurrent mutations: %v", err)
		}
		if len(projection.Items) > 1 {
			t.Fatalf("single-asset fuzz produced multiple open positions: %+v", projection.Items)
		}
		if len(projection.Items) == 1 && (!projection.Items[0].Quantity.IsPositive() || !projection.Items[0].Quantity.FitsStorage()) {
			t.Fatalf("concurrent ledger produced invalid open quantity: %s", projection.Items[0].Quantity.String())
		}
		assertContiguousLedgerSequence(t, h, 1+successes)

		// Same-request concurrent idempotency: regardless of replay/in-flight outcomes,
		// exactly one immutable transaction may be appended for the shared key.
		idemKey := uuid.NewString()
		idemRequest := stage371Trade(h.portfolioID, "BUY", "SBER", "1.00000000", "102.00000000", "2026-01-03")
		beforeIdem := 1 + successes
		idemErrs := make([]error, workers)
		start = make(chan struct{})
		wg = sync.WaitGroup{}
		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				<-start
				_, _, idemErrs[index] = h.service.AppendTransactionWithReplay(
					h.ctx,
					verticalslice.RequestContext{RequestID: uuid.NewString()},
					h.subjectID,
					idemKey,
					"/api/v1/portfolios/"+h.portfolioID+"/transactions",
					idemRequest,
					fuzzAppendArtifact,
				)
			}(i)
		}
		close(start)
		wg.Wait()

		for _, err := range idemErrs {
			if err != nil && !errors.Is(err, postgres.ErrIdempotencyInFlight) {
				t.Fatalf("unexpected concurrent idempotency error: %v", err)
			}
		}
		// A deterministic retry after all contenders finish must replay/complete successfully.
		_, _, err = h.service.AppendTransactionWithReplay(
			h.ctx,
			verticalslice.RequestContext{RequestID: uuid.NewString()},
			h.subjectID,
			idemKey,
			"/api/v1/portfolios/"+h.portfolioID+"/transactions",
			idemRequest,
			fuzzAppendArtifact,
		)
		if err != nil {
			t.Fatalf("idempotency retry after contenders failed: %v", err)
		}
		assertContiguousLedgerSequence(t, h, beforeIdem+1)

		// Concurrent reversal of one still-independent BUY: exactly one reversal wins.
		reversalStart := make(chan struct{})
		reversalErrs := make([]error, workers)
		wg = sync.WaitGroup{}
		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				<-reversalStart
				_, reversalErrs[index] = reverseStage374(t, h, initial, "2026-01-04", uuid.NewString())
			}(i)
		}
		close(reversalStart)
		wg.Wait()

		reversalSuccesses := 0
		for _, err := range reversalErrs {
			switch {
			case err == nil:
				reversalSuccesses++
			case errors.Is(err, verticalslice.ErrTransactionConflict),
				errors.Is(err, verticalslice.ErrInsufficientPositionQuantity):
			default:
				t.Fatalf("unexpected concurrent reversal error: %v", err)
			}
		}
		if reversalSuccesses > 1 {
			t.Fatalf("multiple concurrent reversals committed: %d", reversalSuccesses)
		}
		assertContiguousLedgerSequence(t, h, beforeIdem+1+reversalSuccesses)
	})
}

func FuzzBackdatedCorrectionReversalPositionGeneration(f *testing.F) {
	f.Add(uint8(10), uint8(5), uint8(3), uint8(20))
	f.Add(uint8(2), uint8(2), uint8(1), uint8(100))

	f.Fuzz(func(t *testing.T, openRaw, sellRaw, reopenRaw, priceRaw uint8) {
		openQty := int(openRaw%20) + 1
		sellQty := int(sellRaw%uint8(openQty)) + 1
		reopenQty := int(reopenRaw%20) + 1
		price := int(priceRaw%200) + 1

		h := newStage371Harness(t, "fuzz backdated generation")
		first := appendStage371Trade(t, h, stage371Trade(
			h.portfolioID, "BUY", "SBER",
			fmt.Sprintf("%d.00000000", openQty),
			fmt.Sprintf("%d.00000000", price),
			"2026-02-02",
		))

		// Bind a manual valuation to the first position generation.
		_, err := h.service.UpsertManualValuation(h.ctx, h.subjectID, verticalslice.ManualValuationRequest{
			PortfolioID: h.portfolioID,
			Ticker:      "SBER",
			Price:       verticalslice.Money{Amount: stage377Decimal("150.00000000"), Currency: verticalslice.RUB},
			AsOfDate:    "2026-02-02",
		})
		if err != nil {
			t.Fatalf("bind generation-1 valuation: %v", err)
		}

		appendStage371Trade(t, h, stage371Trade(
			h.portfolioID, "SELL", "SBER",
			fmt.Sprintf("%d.00000000", sellQty),
			"160.00000000", "2026-02-04",
		))
		remaining := openQty - sellQty
		if remaining > 0 {
			appendStage371Trade(t, h, stage371Trade(
				h.portfolioID, "SELL", "SBER",
				fmt.Sprintf("%d.00000000", remaining),
				"160.00000000", "2026-02-05",
			))
		}
		reopened := appendStage371Trade(t, h, stage371Trade(
			h.portfolioID, "BUY", "SBER",
			fmt.Sprintf("%d.00000000", reopenQty),
			"170.00000000", "2026-02-06",
		))

		current, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "")
		if err != nil {
			t.Fatalf("current projection after reopen: %v", err)
		}
		if len(current.Items) != 1 {
			t.Fatalf("reopen must produce one open position: %+v", current.Items)
		}
		if current.Items[0].MarketValuation.Status == verticalslice.MarketValuationAvailableStatus {
			t.Fatalf("valuation from prior position generation leaked into reopened generation")
		}

		// Backdate the reopened transaction and verify all historical/current reads remain coherent.
		corrected, err := correctStage374(
			t, h, reopened,
			fmt.Sprintf("%d.00000000", reopenQty),
			"175.00000000",
			"2026-02-01",
			uuid.NewString(),
		)
		if err != nil {
			t.Fatalf("backdated correction rejected unexpectedly: %v", err)
		}

		for _, asOf := range []string{"2026-01-31", "2026-02-01", "2026-02-03", "2026-02-06"} {
			firstProjection, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, asOf)
			if err != nil {
				t.Fatalf("historical projection %s: %v", asOf, err)
			}
			secondProjection, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, asOf)
			if err != nil {
				t.Fatalf("repeat historical projection %s: %v", asOf, err)
			}
			if len(firstProjection.Items) != len(secondProjection.Items) {
				t.Fatalf("historical projection nondeterminism at %s", asOf)
			}
			for i := range firstProjection.Items {
				if firstProjection.Items[i].Quantity.String() != secondProjection.Items[i].Quantity.String() ||
					firstProjection.Items[i].WeightedAverageCost.Amount.String() != secondProjection.Items[i].WeightedAverageCost.Amount.String() {
					t.Fatalf("historical financial nondeterminism at %s", asOf)
				}
				if !firstProjection.Items[i].Quantity.IsPositive() {
					t.Fatalf("historical projection exposed non-positive open quantity at %s", asOf)
				}
			}
		}

		_, err = reverseStage374(t, h, corrected, "2026-02-07", uuid.NewString())
		if err != nil {
			t.Fatalf("reversal after backdated correction: %v", err)
		}
		finalProjection, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "")
		if err != nil {
			t.Fatalf("final projection: %v", err)
		}
		if len(finalProjection.Items) != 0 {
			t.Fatalf("reversed reopened generation remained open: %+v", finalProjection.Items)
		}

		// The original transaction identity must remain append-only and untouched.
		var originalRows int
		if err := h.db.QueryRowContext(h.ctx,
			"SELECT COUNT(*) FROM investment.transaction_entries WHERE portfolio_id=$1 AND transaction_id=$2",
			h.portfolioID, first.ID,
		).Scan(&originalRows); err != nil {
			t.Fatalf("count original generation rows: %v", err)
		}
		if originalRows != 1 {
			t.Fatalf("unrelated original generation was rewritten: rows=%d", originalRows)
		}
	})
}
