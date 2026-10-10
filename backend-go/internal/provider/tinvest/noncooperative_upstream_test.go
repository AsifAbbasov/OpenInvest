package tinvest

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

// This post-remediation check covers only caller cancellation before the
// independent provider deadline. It makes no remote termination guarantee.
func TestNoncooperativeUpstreamCannotRecycleCapacityOnCallerCancellation(t *testing.T) {
	release := make(chan struct{})
	var releaseOnce sync.Once
	releaseHandlers := func() { releaseOnce.Do(func() { close(release) }) }
	var immediate atomic.Bool
	var remoteActive, remoteMaximum, remoteStarted atomic.Int32
	entered := make(chan struct{}, 2*maxConcurrency)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if !immediate.Load() {
			remoteStarted.Add(1)
			active := remoteActive.Add(1)
			for {
				maximum := remoteMaximum.Load()
				if active <= maximum || remoteMaximum.CompareAndSwap(maximum, active) {
					break
				}
			}
			entered <- struct{}{}
			// Deliberately independent of socket closure, request context and time.
			<-release
			remoteActive.Add(-1)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"dividends":[]}`))
	}))
	defer upstream.Close()
	// Failure cleanup cannot run before the successful security measurement.
	defer releaseHandlers()
	base := http.DefaultTransport.(*http.Transport).Clone()
	base.DisableKeepAlives = true
	defer base.CloseIdleConnections()
	transport := &observedOperationTransport{base: base}
	provider, err := newCorporateActionProvider(&http.Client{Transport: transport}, fixedClock{now: testNow}, testToken, upstream.URL+"/rest")
	if err != nil {
		t.Fatal(err)
	}
	baselineG, baselineFD := runtime.NumGoroutine(), providerRegressionFDCount()
	startedAt := time.Now()
	results := make(chan error, maxConcurrency)
	cancels := make([]context.CancelFunc, maxConcurrency)
	for i := range cancels {
		ctx, cancel := context.WithCancel(context.Background())
		cancels[i] = cancel
		defer cancel()
		go func() { _, err := provider.CorporateActions(ctx, sharedProviderQuery()); results <- err }()
	}
	entryDeadline := time.NewTimer(time.Second)
	defer entryDeadline.Stop()
	for i := 0; i < maxConcurrency; i++ {
		select {
		case <-entered:
		case <-entryDeadline.C:
			t.Fatal("four real upstream handlers did not start promptly")
		}
	}
	firstActive := remoteActive.Load()
	if firstActive != maxConcurrency {
		t.Fatalf("first wave remote active=%d", firstActive)
	}
	for _, cancel := range cancels {
		cancel()
	}
	callerDeadline := time.NewTimer(time.Second)
	defer callerDeadline.Stop()
	for i := 0; i < maxConcurrency; i++ {
		select {
		case err := <-results:
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("caller cancellation=%v", err)
			}
		case <-callerDeadline.C:
			t.Fatal("cancelled callers did not return promptly")
		}
	}
	const observationDelay = 500 * time.Millisecond
	time.Sleep(observationDelay)
	slots, localActive := len(provider.semaphore), transport.active.Load()
	if slots != maxConcurrency || localActive != maxConcurrency || remoteActive.Load() != maxConcurrency {
		t.Fatalf("before local deadline: semaphore=%d local=%d remote=%d", slots, localActive, remoteActive.Load())
	}
	for i := 0; i < maxConcurrency; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
		_, err := provider.CorporateActions(ctx, sharedProviderQuery())
		cancel()
		if !errors.Is(err, verticalslice.ErrCorporateActionsProviderUnavailable) {
			t.Fatalf("occupied capacity must fail closed: %v", err)
		}
	}
	// Drain independent handlers during cleanup even if the challenger fails.
	secondWave := remoteStarted.Load() - maxConcurrency
	if time.Since(startedAt) >= requestTimeout-time.Second {
		t.Fatal("measurement approached independent hard timeout; evidence invalid")
	}
	t.Logf("FIRST_WAVE_REMOTE_ACTIVE=%d CALLERS_CANCELLED=4 CALLERS_RETURNED_PROMPTLY=YES OBSERVATION_DELAY_MS=500 PROVIDER_SEMAPHORE_IN_USE_AFTER_CANCEL=%d LOCAL_PROVIDER_OPERATIONS_AFTER_CANCEL=%d SECOND_WAVE_ATTEMPTS=4 SECOND_WAVE_REACHED_REMOTE=%d REMOTE_TOTAL_STARTED=%d REMOTE_MAX_ACTIVE=%d", firstActive, slots, localActive, secondWave, remoteStarted.Load(), remoteMaximum.Load())
	if secondWave != 0 || remoteMaximum.Load() > maxConcurrency {
		t.Fatal("caller cancellation recycled capacity before local terminal condition")
	}
	releaseHandlers()
	deadline := time.Now().Add(time.Second)
	for (len(provider.semaphore) != 0 || transport.active.Load() != 0 || remoteActive.Load() != 0) && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if len(provider.semaphore) != 0 || transport.active.Load() != 0 || remoteActive.Load() != 0 {
		t.Fatal("explicit release did not recover transport ownership")
	}
	immediate.Store(true)
	if _, err := provider.CorporateActions(context.Background(), sharedProviderQuery()); err != nil {
		t.Fatalf("subsequent legitimate request=%v", err)
	}
	base.CloseIdleConnections()
	runtime.GC()
	time.Sleep(200 * time.Millisecond)
	recoveryG, recoveryFD := runtime.NumGoroutine(), providerRegressionFDCount()
	if recoveryG > baselineG+8 || (baselineFD >= 0 && recoveryFD > baselineFD+2) {
		t.Fatalf("resource recovery: goroutines=%d baseline=%d FDs=%d baseline=%d", recoveryG, baselineG, recoveryFD, baselineFD)
	}
	t.Logf("HOSTILE_NONCOOPERATIVE_UPSTREAM_TEST=PASS CALLER_CANCELLATION_CANNOT_RECYCLE_PROVIDER_CAPACITY_EARLY=YES SEMAPHORE_AFTER_RELEASE=%d LOCAL_PROVIDER_OPERATIONS_AFTER_RELEASE=%d SUBSEQUENT_LEGITIMATE_REQUEST=PASS GOROUTINES_BASELINE=%d GOROUTINES_RECOVERY=%d FD_BASELINE=%d FD_RECOVERY=%d", len(provider.semaphore), transport.active.Load(), baselineG, recoveryG, baselineFD, recoveryFD)
}
