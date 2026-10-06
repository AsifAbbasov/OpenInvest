package tinvest

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestProviderResilienceRepeatedDeadlineStormDrainsAllCapacityAndRecovers(t *testing.T) {
	var active atomic.Int32
	var maximum atomic.Int32
	started := make(chan struct{}, 1024)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := active.Add(1)
		defer active.Add(-1)
		for {
			seen := maximum.Load()
			if current <= seen || maximum.CompareAndSwap(seen, current) {
				break
			}
		}
		started <- struct{}{}
		<-r.Context().Done()
	}))
	defer server.Close()

	provider, err := newCorporateActionProvider(server.Client(), fixedClock{now: testNow}, testToken, server.URL+"/rest")
	if err != nil {
		t.Fatal(err)
	}

	const waves = 40
	const contenders = 64
	for wave := 0; wave < waves; wave++ {
		var wg sync.WaitGroup
		errs := make(chan error, contenders)
		for i := 0; i < contenders; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
				defer cancel()
				_, err := provider.CorporateActions(ctx, testQuery("SBER"))
				errs <- err
			}()
		}
		wg.Wait()
		close(errs)

		for err := range errs {
			if err == nil {
				t.Fatalf("wave %d unexpectedly succeeded against non-responsive provider", wave)
			}
			if !errors.Is(err, context.DeadlineExceeded) &&
				!errors.Is(err, verticalslice.ErrCorporateActionsProviderUnavailable) {
				t.Fatalf("wave %d unexpected error class: %v", wave, err)
			}
		}

		deadline := time.Now().Add(2 * time.Second)
		for (len(provider.semaphore) != 0 || active.Load() != 0) && time.Now().Before(deadline) {
			time.Sleep(5 * time.Millisecond)
		}
		if got := len(provider.semaphore); got != 0 {
			t.Fatalf("wave %d leaked provider semaphore slots: %d", wave, got)
		}
		if got := active.Load(); got != 0 {
			t.Fatalf("wave %d left %d upstream handlers active", wave, got)
		}
	}
	if maximum.Load() > maxConcurrency {
		t.Fatalf("deadline storm exceeded provider concurrency bound: got=%d max=%d", maximum.Load(), maxConcurrency)
	}

	recovery := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"dividends":[]}`)
	}))
	defer recovery.Close()
	recovered, err := newCorporateActionProvider(recovery.Client(), fixedClock{now: testNow}, testToken, recovery.URL+"/rest")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := recovered.CorporateActions(context.Background(), testQuery("SBER")); err != nil {
		t.Fatalf("provider did not recover after deadline storm: %v", err)
	}
}

func TestProviderResilienceConcurrencyAdmissionIsFailFastUnderSaturation(t *testing.T) {
	started := make(chan struct{}, maxConcurrency)
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		started <- struct{}{}
		<-release
		_, _ = io.WriteString(w, `{"dividends":[]}`)
	}))
	defer server.Close()

	provider, err := newCorporateActionProvider(server.Client(), fixedClock{now: testNow}, testToken, server.URL+"/rest")
	if err != nil {
		t.Fatal(err)
	}

	primary := make(chan error, maxConcurrency)
	for i := 0; i < maxConcurrency; i++ {
		go func() {
			_, err := provider.CorporateActions(context.Background(), testQuery("SBER"))
			primary <- err
		}()
	}
	for i := 0; i < maxConcurrency; i++ {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("provider failed to saturate expected concurrency")
		}
	}

	const overflow = 256
	start := time.Now()
	for i := 0; i < overflow; i++ {
		_, err := provider.CorporateActions(context.Background(), testQuery("SBER"))
		if !errors.Is(err, verticalslice.ErrCorporateActionsProviderUnavailable) {
			t.Fatalf("overflow caller %d did not fail closed: %v", i, err)
		}
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("overflow callers queued instead of failing fast: elapsed=%s", elapsed)
	}

	close(release)
	for i := 0; i < maxConcurrency; i++ {
		if err := <-primary; err != nil {
			t.Fatalf("primary request failed after release: %v", err)
		}
	}
}
