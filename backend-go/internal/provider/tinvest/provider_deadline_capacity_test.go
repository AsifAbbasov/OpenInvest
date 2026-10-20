package tinvest

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestCorporateActionsCallerDeadlineRetainsCapacityUntilUpstreamCompletes(t *testing.T) {
	requestStarted := make(chan struct{})
	releaseUpstream := make(chan struct{})
	upstreamFinished := make(chan struct{})
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(releaseUpstream) }) }

	provider, server := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(requestStarted)
		// Keep the upstream in flight even after the caller's deadline.
		<-releaseUpstream
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"dividends":[]}`))
		close(upstreamFinished)
	}))
	t.Cleanup(func() {
		release() // Avoid leaking the test HTTP handler on an assertion failure.
		server.Close()
	})

	// Only the upstream operation owns the semaphore. Context cancellation
	// ends caller waiting, not upstream work, under context.WithoutCancel.
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	result := make(chan error, 1)
	go func() {
		_, err := provider.CorporateActions(ctx, testQuery("SBER"))
		result <- err
	}()

	select {
	case <-requestStarted:
	case <-time.After(5 * time.Second):
		t.Fatal("upstream HTTP request never started")
	}
	select {
	case err := <-result:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("expected caller deadline, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("caller did not return after its deadline")
	}

	select {
	case <-upstreamFinished:
		t.Fatal("upstream unexpectedly finished before release")
	default:
	}
	if got := len(provider.semaphore); got != 1 {
		t.Fatalf("active upstream must retain exactly one slot; got %d", got)
	}

	release()
	select {
	case <-upstreamFinished:
	case <-time.After(5 * time.Second):
		t.Fatal("upstream HTTP handler did not finish")
	}
	// The handler's completion precedes the client's response consumption.
	// The actual release happens after postOwned() and response-body Close.
	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(time.Millisecond)
	defer tick.Stop()
	for len(provider.semaphore) != 0 {
		select {
		case <-tick.C:
		case <-deadline.C:
			t.Fatal("upstream completed but the provider slot was not released")
		}
	}
}
