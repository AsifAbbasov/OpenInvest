package tinvest

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestProviderResilienceRepeatedDeadlineStormDrainsAllCapacityAndRecovers(t *testing.T) {
	var active atomic.Int32
	var maximum atomic.Int32

	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		current := active.Add(1)
		defer active.Add(-1)
		for {
			seen := maximum.Load()
			if current <= seen || maximum.CompareAndSwap(seen, current) {
				break
			}
		}
		<-request.Context().Done()
		return nil, request.Context().Err()
	})
	provider, err := newCorporateActionProvider(
		&http.Client{Transport: transport},
		fixedClock{now: testNow},
		testToken,
		"https://example.invalid/rest",
	)
	if err != nil {
		t.Fatal(err)
	}

	const waves = 100
	const contenders = 64
	for wave := 0; wave < waves; wave++ {
		var wg sync.WaitGroup
		errs := make(chan error, contenders)
		for i := 0; i < contenders; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
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

		deadline := time.Now().Add(time.Second)
		for (len(provider.semaphore) != 0 || active.Load() != 0) && time.Now().Before(deadline) {
			time.Sleep(time.Millisecond)
		}
		if got := len(provider.semaphore); got != 0 {
			t.Fatalf("wave %d leaked provider semaphore slots: %d", wave, got)
		}
		if got := active.Load(); got != 0 {
			t.Fatalf("wave %d retained %d transport calls", wave, got)
		}
	}

	if got := maximum.Load(); got > maxConcurrency {
		t.Fatalf("deadline storm exceeded provider concurrency bound: got=%d max=%d", got, maxConcurrency)
	}

	successTransport := roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"dividends":[]}`)),
		}, nil
	})
	recovered, err := newCorporateActionProvider(
		&http.Client{Transport: successTransport},
		fixedClock{now: testNow},
		testToken,
		"https://example.invalid/rest",
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := recovered.CorporateActions(context.Background(), testQuery("SBER")); err != nil {
		t.Fatalf("provider did not recover after deadline storm: %v", err)
	}
}

func TestProviderResilienceConcurrencyAdmissionIsFailFastUnderSaturation(t *testing.T) {
	release := make(chan struct{})
	started := make(chan struct{}, maxConcurrency)
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		started <- struct{}{}
		select {
		case <-release:
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"dividends":[]}`)),
			}, nil
		case <-request.Context().Done():
			return nil, request.Context().Err()
		}
	})

	provider, err := newCorporateActionProvider(
		&http.Client{Transport: transport},
		fixedClock{now: testNow},
		testToken,
		"https://example.invalid/rest",
	)
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

	const overflow = 512
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
	if got := len(provider.semaphore); got != 0 {
		t.Fatalf("provider capacity did not fully drain after release: %d", got)
	}
}
