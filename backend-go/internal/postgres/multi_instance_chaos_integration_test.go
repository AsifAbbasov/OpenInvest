package postgres_test

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

type chaosClock struct{ now time.Time }

func (clock chaosClock) Now() time.Time { return clock.now }

func TestChaosMultiInstanceSharedDatabaseSerializesMutations(t *testing.T) {
	h := newStage371Harness(t, "chaos multi-instance")
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}

	secondStore, err := postgres.Open(databaseURL)
	if err != nil {
		t.Fatalf("open second application store: %v", err)
	}
	t.Cleanup(func() { _ = secondStore.Close() })
	secondService := verticalslice.NewService(secondStore, verticalslice.SystemClock{})

	appendStage371Trade(t, h, stage371Trade(
		h.portfolioID, "BUY", "SBER", "10.00000000", "100.00000000", "2026-04-01",
	))

	start := make(chan struct{})
	errs := make([]error, 2)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		_, errs[0] = h.service.AppendTransaction(
			h.ctx,
			verticalslice.RequestContext{RequestID: uuid.NewString()},
			h.subjectID,
			uuid.NewString(),
			"/api/v1/portfolios/"+h.portfolioID+"/transactions",
			stage371Trade(h.portfolioID, "SELL", "SBER", "7.00000000", "110.00000000", "2026-04-02"),
		)
	}()
	go func() {
		defer wg.Done()
		<-start
		_, errs[1] = secondService.AppendTransaction(
			h.ctx,
			verticalslice.RequestContext{RequestID: uuid.NewString()},
			h.subjectID,
			uuid.NewString(),
			"/api/v1/portfolios/"+h.portfolioID+"/transactions",
			stage371Trade(h.portfolioID, "SELL", "SBER", "7.00000000", "120.00000000", "2026-04-02"),
		)
	}()
	close(start)
	wg.Wait()

	successes := 0
	for _, err := range errs {
		if err == nil {
			successes++
			continue
		}
		if err != verticalslice.ErrInsufficientPositionQuantity {
			t.Fatalf("unexpected multi-instance mutation result: %v", err)
		}
	}
	if successes != 1 {
		t.Fatalf("expected exactly one legal SELL across two instances, got successes=%d errors=%v", successes, errs)
	}
	assertContiguousLedgerSequence(t, h, 2)
}

func TestChaosMultiInstanceSharedIdempotencyAllowsOneMutation(t *testing.T) {
	h := newStage371Harness(t, "chaos shared idempotency")
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}

	secondStore, err := postgres.Open(databaseURL)
	if err != nil {
		t.Fatalf("open second application store: %v", err)
	}
	t.Cleanup(func() { _ = secondStore.Close() })
	secondService := verticalslice.NewService(secondStore, verticalslice.SystemClock{})

	key := uuid.NewString()
	path := "/api/v1/portfolios/" + h.portfolioID + "/transactions"
	request := stage371Trade(h.portfolioID, "BUY", "SBER", "1.00000000", "100.00000000", "2026-05-01")

	start := make(chan struct{})
	errs := make([]error, 2)
	var wg sync.WaitGroup
	wg.Add(2)
	for i := 0; i < 2; i++ {
		index := i
		go func() {
			defer wg.Done()
			<-start
			service := h.service
			if index == 1 {
				service = secondService
			}
			_, _, errs[index] = service.AppendTransactionWithReplay(
				h.ctx,
				verticalslice.RequestContext{RequestID: uuid.NewString()},
				h.subjectID,
				key,
				path,
				request,
				fuzzAppendArtifact,
			)
		}()
	}
	close(start)
	wg.Wait()

	// A deterministic retry after both contenders finish must replay the same command
	// without appending another immutable ledger row.
	_, _, err = secondService.AppendTransactionWithReplay(
		h.ctx,
		verticalslice.RequestContext{RequestID: uuid.NewString()},
		h.subjectID,
		key,
		path,
		request,
		fuzzAppendArtifact,
	)
	if err != nil {
		t.Fatalf("shared idempotency retry failed: %v; contenders=%v", err, errs)
	}
	assertContiguousLedgerSequence(t, h, 1)
}

func TestChaosClockSkewDoesNotOverrideLedgerSequenceAuthority(t *testing.T) {
	h := newStage371Harness(t, "chaos clock skew")
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}

	pastStore, err := postgres.Open(databaseURL)
	if err != nil {
		t.Fatalf("open past-skew store: %v", err)
	}
	t.Cleanup(func() { _ = pastStore.Close() })
	futureStore, err := postgres.Open(databaseURL)
	if err != nil {
		t.Fatalf("open future-skew store: %v", err)
	}
	t.Cleanup(func() { _ = futureStore.Close() })

	base := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	past := verticalslice.NewService(pastStore, chaosClock{now: base.Add(-12 * time.Hour)})
	future := verticalslice.NewService(futureStore, chaosClock{now: base.Add(12 * time.Hour)})

	for index, service := range []*verticalslice.Service{future, past, future, past} {
		_, err := service.AppendTransaction(
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
				"2026-06-01",
			),
		)
		if err != nil {
			t.Fatalf("clock-skew append %d: %v", index, err)
		}
	}

	assertContiguousLedgerSequence(t, h, 4)
	projection, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "")
	if err != nil {
		t.Fatalf("projection after clock-skew writes: %v", err)
	}
	if len(projection.Items) != 1 || projection.Items[0].Quantity.String() != "4.00000000" {
		t.Fatalf("clock skew changed financial ordering/state: %+v", projection.Items)
	}
}
