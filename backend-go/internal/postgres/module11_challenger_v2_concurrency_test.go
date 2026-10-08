package postgres_test

import (
	"errors"
	"os"
	"sync"
	"testing"

	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func m11V2BuyVsSell(t *testing.T) {
	h := newStage371Harness(t, "M11 V2 buy vs sell")
	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "10.00000000", "100.00000000", "2026-01-01"))
	start := make(chan struct{})
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done(); <-start
		_, err := h.service.AppendTransaction(h.ctx, verticalslice.RequestContext{}, h.subjectID, uuid.NewString(),
			"/api/v1/portfolios/"+h.portfolioID+"/transactions",
			stage371Trade(h.portfolioID, "BUY", "SBER", "5.00000000", "110.00000000", "2026-01-02"))
		errs <- err
	}()
	go func() {
		defer wg.Done(); <-start
		_, err := h.service.AppendTransaction(h.ctx, verticalslice.RequestContext{}, h.subjectID, uuid.NewString(),
			"/api/v1/portfolios/"+h.portfolioID+"/transactions",
			stage371Trade(h.portfolioID, "SELL", "SBER", "8.00000000", "120.00000000", "2026-01-03"))
		errs <- err
	}()
	close(start); wg.Wait(); close(errs)
	for err := range errs { if err != nil { t.Fatalf("buy/sell concurrency: %v", err) } }
	projection, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "")
	if err != nil { t.Fatal(err) }
	if len(projection.Items) != 1 || projection.Items[0].Quantity.String() != "7.00000000" {
		t.Fatalf("final position=%+v", projection.Items)
	}
}

func m11V2ReversalVsReversal(t *testing.T) {
	h := newStage371Harness(t, "M11 V2 reversal vs reversal")
	tx := appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "5.00000000", "100.00000000", "2026-01-01"))
	start := make(chan struct{})
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done(); <-start
			_, err := reverseStage374(t, h, tx, "2026-01-02", uuid.NewString())
			errs <- err
		}()
	}
	close(start); wg.Wait(); close(errs)
	success, conflicts := 0, 0
	for err := range errs {
		if err == nil { success++ } else if errors.Is(err, verticalslice.ErrTransactionConflict) { conflicts++ } else { t.Fatalf("unexpected reversal error: %v", err) }
	}
	if success != 1 || conflicts != 1 { t.Fatalf("reversal results success=%d conflicts=%d", success, conflicts) }
	projection, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "")
	if err != nil { t.Fatal(err) }
	if len(projection.Items) != 0 { t.Fatalf("reversed position resurrected: %+v", projection.Items) }
}

func m11V2SameIdempotencyConcurrentRetry(t *testing.T) {
	h := newStage371Harness(t, "M11 V2 same idempotency")
	key := uuid.NewString()
	path := "/api/v1/portfolios/"+h.portfolioID+"/transactions"
	req := stage375Request(h.portfolioID, "DEPOSIT", "", "10.00000000", "0.00000000", "0.00000000", "2026-02-01")
	start := make(chan struct{})
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done(); <-start
			rc := verticalslice.RequestContext{RequestID: uuid.NewString(), TraceID: "m11-v2-same-key"}
			_, _, err := h.service.AppendTransactionWithReplay(h.ctx, rc, h.subjectID, key, path, req,
				func(tx verticalslice.Transaction) (verticalslice.CommandReplayArtifact, error) { return m11Artifact(rc, tx) })
			errs <- err
		}()
	}
	close(start); wg.Wait(); close(errs)
	for err := range errs {
		if err != nil && !errors.Is(err, postgres.ErrIdempotencyInFlight) { t.Fatalf("same-key concurrency error=%v", err) }
	}
	var rows int
	if err := h.db.QueryRowContext(h.ctx, `SELECT count(*) FROM investment.transaction_entries WHERE portfolio_id=$1`, h.portfolioID).Scan(&rows); err != nil { t.Fatal(err) }
	if rows != 1 { t.Fatalf("same-key business effects=%d want=1", rows) }
	rc := verticalslice.RequestContext{RequestID: uuid.NewString(), TraceID: "m11-v2-same-key-retry"}
	if _, _, err := h.service.AppendTransactionWithReplay(h.ctx, rc, h.subjectID, key, path, req,
		func(tx verticalslice.Transaction) (verticalslice.CommandReplayArtifact, error) { return m11Artifact(rc, tx) }); err != nil {
		t.Fatalf("settled retry failed: %v", err)
	}
}

func m11V2DifferentKeysSameIntent(t *testing.T) {
	h := newStage371Harness(t, "M11 V2 different idempotency keys")
	req := stage375Request(h.portfolioID, "DEPOSIT", "", "10.00000000", "0.00000000", "0.00000000", "2026-02-01")
	start := make(chan struct{})
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done(); <-start
			_, err := h.service.AppendTransaction(h.ctx, verticalslice.RequestContext{}, h.subjectID, uuid.NewString(),
				"/api/v1/portfolios/"+h.portfolioID+"/transactions", req)
			errs <- err
		}()
	}
	close(start); wg.Wait(); close(errs)
	for err := range errs { if err != nil { t.Fatalf("different-key command error=%v", err) } }
	var rows int
	if err := h.db.QueryRowContext(h.ctx, `SELECT count(*) FROM investment.transaction_entries WHERE portfolio_id=$1`, h.portfolioID).Scan(&rows); err != nil { t.Fatal(err) }
	if rows != 2 { t.Fatalf("different keys expected two distinct commands, rows=%d", rows) }
	t.Log("M11_V2_DIFFERENT_KEYS_SAME_INTENT=DISTINCT_COMMANDS_BY_CONTRACT")
}

func TestM11V2RepeatedConcurrencyCampaign(t *testing.T) {
	if os.Getenv("OPENINVEST_M11_V2_EXPENSIVE") != "1" { t.Skip("specialized Module 11 V2 campaign only") }
	const iterations = 12
	scenarios := []struct {
		name string
		run func(*testing.T)
	}{
		{"buy_vs_sell", m11V2BuyVsSell},
		{"sell_vs_sell", TestM11ConcurrentSellVsSellNoOversell},
		{"correction_vs_correction", TestStage374ConcurrentCorrectionsOnlyOneWins},
		{"correction_vs_reversal", TestM11ConcurrentCorrectionVsReversalSerializes},
		{"reversal_vs_reversal", m11V2ReversalVsReversal},
		{"xirr_read_vs_correction", TestM11XIRRReadDuringCorrectionCommittedStatesOnly},
		{"same_idempotency_retry", m11V2SameIdempotencyConcurrentRetry},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			for i := 0; i < iterations; i++ {
				t.Run(uuid.NewString(), scenario.run)
			}
		})
	}
	t.Run("different_keys_same_intent_contract_observation", m11V2DifferentKeysSameIntent)
	t.Logf("M11_V2_DB_CONCURRENCY_SCENARIOS=%d", len(scenarios))
	t.Logf("M11_V2_DB_CONCURRENCY_ITERATIONS_PER_SCENARIO=%d", iterations)
	t.Logf("M11_V2_DB_CONCURRENCY_TOTAL_ITERATIONS=%d", len(scenarios)*iterations)
}
