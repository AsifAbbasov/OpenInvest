package tinvest

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

// Post-fix local lifecycle verification using a cooperative real socket fixture.
// This does not reproduce remote computation surviving connection termination.
func TestDetachedOperationRetainsRealSocketCapacity(t *testing.T) {
	var mu sync.Mutex
	release := make(chan struct{})
	var immediate atomic.Bool
	entered := make(chan struct{}, maxConcurrency)
	var remoteActive, remoteMaximum atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		active := remoteActive.Add(1)
		for {
			maximum := remoteMaximum.Load()
			if active <= maximum || remoteMaximum.CompareAndSwap(maximum, active) {
				break
			}
		}
		defer remoteActive.Add(-1)
		if !immediate.Load() {
			mu.Lock()
			gate := release
			mu.Unlock()
			entered <- struct{}{}
			select {
			case <-gate:
			case <-r.Context().Done():
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"dividends":[]}`))
	}))
	defer upstream.Close()
	base := http.DefaultTransport.(*http.Transport).Clone()
	base.DisableKeepAlives = true
	defer base.CloseIdleConnections()
	observed := &observedOperationTransport{base: base}
	provider, err := newCorporateActionProvider(&http.Client{Transport: observed}, fixedClock{now: testNow}, testToken, upstream.URL+"/rest")
	if err != nil {
		t.Fatal(err)
	}
	baselineG, baselineFD := runtime.NumGoroutine(), providerRegressionFDCount()
	peakG, peakFD := baselineG, baselineFD
	maximumSlots := 0
	replacements := 0
	const rounds = 10
	for round := 0; round < rounds; round++ {
		mu.Lock()
		release = make(chan struct{})
		gate := release
		mu.Unlock()
		results := make(chan error, maxConcurrency)
		cancels := make([]context.CancelFunc, maxConcurrency)
		for i := range cancels {
			ctx, cancel := context.WithCancel(context.Background())
			cancels[i] = cancel
			go func() { _, err := provider.CorporateActions(ctx, sharedProviderQuery()); results <- err }()
		}
		for i := 0; i < maxConcurrency; i++ {
			select {
			case <-entered:
			case <-time.After(2 * time.Second):
				close(gate)
				t.Fatal("local socket operation did not start")
			}
		}
		for _, cancel := range cancels {
			cancel()
		}
		for i := 0; i < maxConcurrency; i++ {
			select {
			case err := <-results:
				if !errors.Is(err, context.Canceled) {
					close(gate)
					t.Fatalf("caller cancellation=%v", err)
				}
			case <-time.After(time.Second):
				close(gate)
				t.Fatal("caller cancellation was not prompt")
			}
		}
		time.Sleep(25 * time.Millisecond)
		if len(provider.semaphore) != maxConcurrency || observed.active.Load() != maxConcurrency {
			close(gate)
			t.Fatal("caller cancellation recycled locally owned transport")
		}
		maximumSlots = len(provider.semaphore)
		for i := 0; i < maxConcurrency; i++ {
			_, err := provider.CorporateActions(context.Background(), sharedProviderQuery())
			if err == nil {
				replacements++
			}
			if !errors.Is(err, verticalslice.ErrCorporateActionsProviderUnavailable) {
				close(gate)
				t.Fatalf("occupied capacity error=%v", err)
			}
		}
		if g := runtime.NumGoroutine(); g > peakG {
			peakG = g
		}
		if fd := providerRegressionFDCount(); fd > peakFD {
			peakFD = fd
		}
		close(gate)
		deadline := time.Now().Add(time.Second)
		for (len(provider.semaphore) != 0 || observed.active.Load() != 0 || remoteActive.Load() != 0) && time.Now().Before(deadline) {
			time.Sleep(time.Millisecond)
		}
		if len(provider.semaphore) != 0 || observed.active.Load() != 0 || remoteActive.Load() != 0 {
			t.Fatal("local terminal condition did not recover")
		}
	}
	immediate.Store(true)
	if _, err := provider.CorporateActions(context.Background(), sharedProviderQuery()); err != nil {
		t.Fatalf("subsequent request=%v", err)
	}
	base.CloseIdleConnections()
	runtime.GC()
	time.Sleep(200 * time.Millisecond)
	recoveryG, recoveryFD := runtime.NumGoroutine(), providerRegressionFDCount()
	if recoveryG > baselineG+maxConcurrency+4 {
		t.Fatalf("goroutine recovery=%d baseline=%d", recoveryG, baselineG)
	}
	if baselineFD >= 0 && recoveryFD > baselineFD+2 {
		t.Fatalf("FD recovery=%d baseline=%d", recoveryFD, baselineFD)
	}
	if observed.maximum.Load() > maxConcurrency {
		t.Fatal("local transport maximum exceeded")
	}
	t.Logf("DETACHED_SOCKET_TOTAL_CANCELLATIONS=40 CANCELLATION_DRIVEN_REPLACEMENT_BEFORE_LOCAL_TERMINAL=%d MAX_LOCAL_PROVIDER_TRANSPORTS=%d MAX_SEMAPHORE_IN_USE=%d COOPERATIVE_FIXTURE_MAX_REMOTE_HANDLERS=%d GOROUTINES_BASELINE=%d GOROUTINES_PEAK=%d GOROUTINES_RECOVERY=%d FD_BASELINE=%d FD_PEAK=%d FD_RECOVERY=%d", replacements, observed.maximum.Load(), maximumSlots, remoteMaximum.Load(), baselineG, peakG, recoveryG, baselineFD, peakFD, recoveryFD)
}

type observedOperationTransport struct {
	base    http.RoundTripper
	active  atomic.Int32
	maximum atomic.Int32
}

func (transport *observedOperationTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	active := transport.active.Add(1)
	for {
		maximum := transport.maximum.Load()
		if active <= maximum || transport.maximum.CompareAndSwap(maximum, active) {
			break
		}
	}
	response, err := transport.base.RoundTrip(request)
	if err != nil {
		transport.active.Add(-1)
		return response, err
	}
	response.Body = &observedOperationBody{ReadCloser: response.Body, done: func() { transport.active.Add(-1) }}
	return response, nil
}

type observedOperationBody struct {
	io.ReadCloser
	once sync.Once
	done func()
}

func (body *observedOperationBody) Close() error {
	err := body.ReadCloser.Close()
	body.once.Do(body.done)
	return err
}

func TestDetachedOperationHardTimeoutStillTerminatesLocalSocket(t *testing.T) {
	entered := make(chan struct{}, 1)
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entered <- struct{}{}
		select {
		case <-r.Context().Done():
		case <-release:
		}
	}))
	defer server.Close()
	defer close(release)
	base := http.DefaultTransport.(*http.Transport).Clone()
	defer base.CloseIdleConnections()
	transport := &observedOperationTransport{base: base}
	provider, err := newCorporateActionProvider(&http.Client{Transport: transport}, fixedClock{now: testNow}, testToken, server.URL+"/rest")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan error, 1)
	go func() { _, err := provider.CorporateActions(ctx, sharedProviderQuery()); result <- err }()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("socket did not start")
	}
	cancel()
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("caller error=%v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("caller did not return promptly")
	}
	time.Sleep(100 * time.Millisecond)
	if transport.active.Load() != 1 || len(provider.semaphore) != 1 {
		t.Fatal("caller cancellation terminated detached operation early")
	}
	deadline := time.Now().Add(requestTimeout + time.Second)
	for (len(provider.semaphore) != 0 || transport.active.Load() != 0) && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if len(provider.semaphore) != 0 || transport.active.Load() != 0 {
		t.Fatal("hard timeout failed to terminate locally owned transport")
	}
	t.Log("DETACHED_OPERATION_HARD_TIMEOUT=PASS CALLER_CANCEL_RETURNS_PROMPTLY=YES")
}
