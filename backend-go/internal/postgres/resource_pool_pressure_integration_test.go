package postgres_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestResourceExhaustionPostgresPoolSaturationCancelsAndRecovers(t *testing.T) {
	h := newStage371Harness(t, "postgres pool saturation")

	// Hold the portfolio row lock from an independent verification connection.
	blocker, err := h.db.BeginTx(h.ctx, nil)
	if err != nil {
		t.Fatalf("begin blocker transaction: %v", err)
	}
	defer blocker.Rollback()
	if _, err := blocker.ExecContext(
		h.ctx,
		"SELECT id FROM investment.portfolios WHERE id=$1 FOR UPDATE",
		h.portfolioID,
	); err != nil {
		t.Fatalf("lock portfolio row: %v", err)
	}

	// The production Store pool is capped at 10 open connections. Occupy all of them
	// with legitimate requests waiting on the same portfolio lock.
	const workers = 10
	start := make(chan struct{})
	errs := make([]error, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			<-start
			_, errs[index] = h.service.AppendTransaction(
				h.ctx,
				verticalslice.RequestContext{RequestID: uuid.NewString()},
				h.subjectID,
				uuid.NewString(),
				"/api/v1/portfolios/"+h.portfolioID+"/transactions",
				stage371Trade(
					h.portfolioID,
					"BUY",
					"SBER",
					"1.00000000",
					"100.00000000",
					"2026-03-01",
				),
			)
		}(i)
	}
	close(start)

	// Give the contenders time to consume the Store pool while blocked on the row lock.
	time.Sleep(250 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	_, err = h.service.AppendTransaction(
		ctx,
		verticalslice.RequestContext{RequestID: uuid.NewString()},
		h.subjectID,
		uuid.NewString(),
		"/api/v1/portfolios/"+h.portfolioID+"/transactions",
		stage371Trade(
			h.portfolioID,
			"BUY",
			"SBER",
			"1.00000000",
			"100.00000000",
			"2026-03-01",
		),
	)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("pool-saturated request did not honor context deadline: %v", err)
	}

	// Release pressure. All legitimate blocked requests must drain and the pool must
	// remain reusable rather than leaking connections or wedging the service.
	if err := blocker.Rollback(); err != nil {
		t.Fatalf("release blocker transaction: %v", err)
	}
	wg.Wait()
	for index, err := range errs {
		if err != nil {
			t.Fatalf("blocked request %d failed after pressure release: %v", index, err)
		}
	}

	_, err = h.service.AppendTransaction(
		h.ctx,
		verticalslice.RequestContext{RequestID: uuid.NewString()},
		h.subjectID,
		uuid.NewString(),
		"/api/v1/portfolios/"+h.portfolioID+"/transactions",
		stage371Trade(
			h.portfolioID,
			"BUY",
			"SBER",
			"1.00000000",
			"101.00000000",
			"2026-03-02",
		),
	)
	if err != nil {
		t.Fatalf("service did not recover after pool saturation: %v", err)
	}

	assertContiguousLedgerSequence(t, h, workers+1)
}
