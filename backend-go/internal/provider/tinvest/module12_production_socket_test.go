package tinvest

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func module12ProviderQuery() verticalslice.CorporateActionQuery {
	return verticalslice.CorporateActionQuery{
		InstrumentIDs: []string{"SBER"},
		From:          "2026-09-01",
		To:            "2026-09-30",
	}
}

func module12Provider(t *testing.T, baseURL string, transport *http.Transport) *Provider {
	t.Helper()
	provider, err := newCorporateActionProvider(
		&http.Client{Transport: transport},
		verticalslice.SystemClock{},
		"module12-readonly-token",
		baseURL+"/rest",
	)
	if err != nil {
		t.Fatal(err)
	}
	return provider
}

func TestM12ProviderBudgetIsProcessLocalAcrossInstances(t *testing.T) {
	var upstream atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		upstream.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"dividends":[]}`))
	}))
	defer server.Close()

	for _, instances := range []int{1, 2, 4} {
		t.Run(string(rune('0'+instances))+"-instances", func(t *testing.T) {
			before := upstream.Load()
			providers := make([]*Provider, 0, instances)
			transports := make([]*http.Transport, 0, instances)
			for i := 0; i < instances; i++ {
				transport := &http.Transport{}
				transports = append(transports, transport)
				providers = append(providers, module12Provider(t, server.URL, transport))
			}
			for _, provider := range providers {
				for i := 0; i < maxRequestsPerMinute; i++ {
					if _, err := provider.CorporateActions(context.Background(), module12ProviderQuery()); err != nil {
						t.Fatalf("provider request %d: %v", i+1, err)
					}
				}
				if _, err := provider.CorporateActions(context.Background(), module12ProviderQuery()); !errors.Is(err, verticalslice.ErrCorporateActionsProviderUnavailable) {
					t.Fatalf("61st provider request did not fail closed: %v", err)
				}
			}
			delta := upstream.Load() - before
			want := int64(instances * maxRequestsPerMinute)
			if delta != want {
				t.Fatalf("upstream calls=%d want=%d", delta, want)
			}
			for _, transport := range transports {
				transport.CloseIdleConnections()
			}
			t.Logf("M12_PROVIDER_INSTANCE_BUDGET instances=%d per_instance_limit=%d total_upstream_calls=%d multiplication=%dx", instances, maxRequestsPerMinute, delta, instances)
		})
	}
}

type module12SocketMode int32

const (
	module12SocketSuccess module12SocketMode = iota
	module12SocketBlockUntilCancel
	module12SocketPartialResponse
	module12SocketReset
	module12SocketTimeout
)

type module12SocketState struct {
	mode    atomic.Int32
	active  atomic.Int32
	maximum atomic.Int32
	entered chan struct{}
}

func (s *module12SocketState) updateMaximum(current int32) {
	for {
		seen := s.maximum.Load()
		if current <= seen || s.maximum.CompareAndSwap(seen, current) {
			return
		}
	}
}

func (s *module12SocketState) handler(w http.ResponseWriter, r *http.Request) {
	current := s.active.Add(1)
	s.updateMaximum(current)
	defer s.active.Add(-1)
	select {
	case s.entered <- struct{}{}:
	default:
	}
	switch module12SocketMode(s.mode.Load()) {
	case module12SocketBlockUntilCancel:
		<-r.Context().Done()
	case module12SocketPartialResponse:
		hijacker, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "hijack unavailable", http.StatusInternalServerError)
			return
		}
		conn, rw, err := hijacker.Hijack()
		if err != nil {
			return
		}
		_, _ = rw.WriteString("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: 100\r\n\r\n{\"dividends\":[")
		_ = rw.Flush()
		_ = conn.Close()
	case module12SocketReset:
		hijacker, ok := w.(http.Hijacker)
		if !ok {
			return
		}
		conn, _, err := hijacker.Hijack()
		if err == nil {
			_ = conn.Close()
		}
	case module12SocketTimeout:
		select {
		case <-r.Context().Done():
		case <-time.After(7 * time.Second):
		}
	default:
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"dividends":[]}`))
	}
}

func module12FDs() int {
	entries, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		return -1
	}
	return len(entries)
}

func TestM12ProviderRealSocketCancellationResetAndRecovery(t *testing.T) {
	state := &module12SocketState{entered: make(chan struct{}, 128)}
	server := httptest.NewServer(http.HandlerFunc(state.handler))
	defer server.Close()
	transport := &http.Transport{}
	defer transport.CloseIdleConnections()
	provider := module12Provider(t, server.URL, transport)

	baselineG := runtime.NumGoroutine()
	baselineFD := module12FDs()
	state.mode.Store(int32(module12SocketBlockUntilCancel))

	const rounds = 5
	const concurrent = 4
	for round := 0; round < rounds; round++ {
		errs := make(chan error, concurrent)
		cancels := make([]context.CancelFunc, 0, concurrent)
		for i := 0; i < concurrent; i++ {
			ctx, cancel := context.WithCancel(context.Background())
			cancels = append(cancels, cancel)
			go func(ctx context.Context) {
				_, err := provider.CorporateActions(ctx, module12ProviderQuery())
				errs <- err
			}(ctx)
		}
		for i := 0; i < concurrent; i++ {
			select {
			case <-state.entered:
			case <-time.After(2 * time.Second):
				t.Fatal("provider request did not reach real socket")
			}
		}
		for _, cancel := range cancels {
			cancel()
		}
		for i := 0; i < concurrent; i++ {
			if err := <-errs; !errors.Is(err, context.Canceled) {
				t.Fatalf("cancelled provider request returned %v", err)
			}
		}
	}

	peakG := runtime.NumGoroutine()
	peakFD := module12FDs()
	deadline := time.Now().Add(2 * time.Second)
	for state.active.Load() != 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if state.active.Load() != 0 || len(provider.semaphore) != 0 {
		t.Fatalf("provider resources not recovered: active=%d semaphore=%d", state.active.Load(), len(provider.semaphore))
	}

	state.mode.Store(int32(module12SocketSuccess))
	if _, err := provider.CorporateActions(context.Background(), module12ProviderQuery()); err != nil {
		t.Fatalf("legitimate request after cancellations: %v", err)
	}
	t.Logf("M12_PROVIDER_CONCURRENT_CANCELLATIONS=%d semaphore_recovered=YES subsequent_request=PASS", rounds*concurrent)

	state.mode.Store(int32(module12SocketPartialResponse))
	if _, err := provider.CorporateActions(context.Background(), module12ProviderQuery()); !errors.Is(err, verticalslice.ErrCorporateActionsProviderUnavailable) {
		t.Fatalf("partial response classification: %v", err)
	}
	state.mode.Store(int32(module12SocketSuccess))
	if _, err := provider.CorporateActions(context.Background(), module12ProviderQuery()); err != nil {
		t.Fatalf("recovery after partial response: %v", err)
	}
	t.Log("M12_PROVIDER_PARTIAL_RESPONSE=FAIL_CLOSED RECOVERY=PASS")

	state.mode.Store(int32(module12SocketReset))
	if _, err := provider.CorporateActions(context.Background(), module12ProviderQuery()); !errors.Is(err, verticalslice.ErrCorporateActionsProviderUnavailable) {
		t.Fatalf("reset classification: %v", err)
	}
	state.mode.Store(int32(module12SocketSuccess))
	if _, err := provider.CorporateActions(context.Background(), module12ProviderQuery()); err != nil {
		t.Fatalf("recovery after reset: %v", err)
	}
	t.Log("M12_PROVIDER_CONNECTION_RESET=FAIL_CLOSED RECOVERY=PASS")

	state.mode.Store(int32(module12SocketTimeout))
	timeoutStart := time.Now()
	if _, err := provider.CorporateActions(context.Background(), module12ProviderQuery()); !errors.Is(err, verticalslice.ErrCorporateActionsProviderUnavailable) {
		t.Fatalf("timeout classification: %v", err)
	}
	timeoutElapsed := time.Since(timeoutStart)
	state.mode.Store(int32(module12SocketSuccess))
	if _, err := provider.CorporateActions(context.Background(), module12ProviderQuery()); err != nil {
		t.Fatalf("recovery after timeout: %v", err)
	}
	t.Logf("M12_PROVIDER_UPSTREAM_TIMEOUT elapsed=%s configured=%s RECOVERY=PASS", timeoutElapsed.Round(time.Millisecond), requestTimeout)

	transport.CloseIdleConnections()
	runtime.GC()
	time.Sleep(300 * time.Millisecond)
	t.Logf("M12_PROVIDER_SOCKET_RESOURCES goroutines_baseline=%d goroutines_peak=%d goroutines_recovery=%d fd_baseline=%d fd_peak=%d fd_recovery=%d provider_max_active=%d",
		baselineG, peakG, runtime.NumGoroutine(), baselineFD, peakFD, module12FDs(), state.maximum.Load())
}
