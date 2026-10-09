package tinvest

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/sharedbudget"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func requireProviderSharedBudgetRedisURL(t *testing.T) string {
	t.Helper()
	url := os.Getenv("OPENINVEST_SHARED_BUDGET_TEST_REDIS_URL")
	if url == "" {
		t.Skip("OPENINVEST_SHARED_BUDGET_TEST_REDIS_URL is required")
	}
	return url
}

func newProviderSharedAuthority(t *testing.T, redisURL, namespace string) sharedbudget.Authority {
	t.Helper()
	authority, err := sharedbudget.NewRedisAuthority(redisURL, namespace)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := authority.Ping(ctx); err != nil {
		_ = authority.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = authority.Close() })
	return authority
}

func sharedProviderQuery() verticalslice.CorporateActionQuery {
	return verticalslice.CorporateActionQuery{
		InstrumentIDs: []string{"SBER"},
		From:          "2026-09-01",
		To:            "2026-09-30",
	}
}

func TestSharedProviderBudgetAcrossInstancesAndRestart(t *testing.T) {
	redisURL := requireProviderSharedBudgetRedisURL(t)
	var upstream atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		upstream.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"dividends":[]}`))
	}))
	defer server.Close()

	for _, instances := range []int{1, 2, 4} {
		t.Run(fmt.Sprintf("instances-%d", instances), func(t *testing.T) {
			namespace := fmt.Sprintf("openinvest:test:m12:provider:%d:%d", instances, time.Now().UnixNano())
			before := upstream.Load()
			providers := make([]*Provider, 0, instances)
			for i := 0; i < instances; i++ {
				authority := newProviderSharedAuthority(t, redisURL, namespace)
				provider, err := newCorporateActionProviderWithSharedBudget(
					&http.Client{},
					fixedClock{now: testNow},
					testToken,
					server.URL+"/rest",
					authority,
				)
				if err != nil {
					t.Fatal(err)
				}
				providers = append(providers, provider)
			}

			var accepted, limited int
			for i := 0; i < 80; i++ {
				_, err := providers[i%len(providers)].CorporateActions(context.Background(), sharedProviderQuery())
				if err == nil {
					accepted++
					continue
				}
				if errors.Is(err, verticalslice.ErrCorporateActionsProviderUnavailable) {
					limited++
					continue
				}
				t.Fatalf("provider request %d: %v", i+1, err)
			}
			delta := upstream.Load() - before
			if accepted != maxRequestsPerMinute || delta != maxRequestsPerMinute {
				t.Fatalf("instances=%d accepted=%d upstream=%d want=%d", instances, accepted, delta, maxRequestsPerMinute)
			}
			if limited != 20 {
				t.Fatalf("instances=%d limited=%d want=20", instances, limited)
			}

			restartAuthority := newProviderSharedAuthority(t, redisURL, namespace)
			restarted, err := newCorporateActionProviderWithSharedBudget(
				&http.Client{},
				fixedClock{now: testNow},
				testToken,
				server.URL+"/rest",
				restartAuthority,
			)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := restarted.CorporateActions(context.Background(), sharedProviderQuery()); !errors.Is(err, verticalslice.ErrCorporateActionsProviderUnavailable) {
				t.Fatalf("restarted provider received fresh budget: %v", err)
			}
			if got := upstream.Load() - before; got != maxRequestsPerMinute {
				t.Fatalf("restarted provider reached upstream: %d", got)
			}
			t.Logf("M12_SHARED_PROVIDER instances=%d total_provider_calls=%d provider_budget_multiplication=NO process_restart_shared_budget_reset=NO",
				instances, delta)
		})
	}
}

func TestSharedProviderBudgetBackendFailureFailsClosed(t *testing.T) {
	redisURL := requireProviderSharedBudgetRedisURL(t)
	authority, err := sharedbudget.NewRedisAuthority(redisURL, fmt.Sprintf("openinvest:test:m12:provider-failure:%d", time.Now().UnixNano()))
	if err != nil {
		t.Fatal(err)
	}
	if err := authority.Close(); err != nil {
		t.Fatal(err)
	}
	var upstream atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		upstream.Add(1)
		_, _ = w.Write([]byte(`{"dividends":[]}`))
	}))
	defer server.Close()
	provider, err := newCorporateActionProviderWithSharedBudget(
		&http.Client{},
		fixedClock{now: testNow},
		testToken,
		server.URL+"/rest",
		authority,
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.CorporateActions(context.Background(), sharedProviderQuery()); !errors.Is(err, verticalslice.ErrCorporateActionsProviderUnavailable) {
		t.Fatalf("shared backend outage error=%v", err)
	}
	if upstream.Load() != 0 {
		t.Fatalf("shared backend outage upstream calls=%d want=0", upstream.Load())
	}
	t.Log("M12_SHARED_PROVIDER_BACKEND_UNAVAILABLE=FAIL_CLOSED provider_calls=0")
}

func TestSharedProviderAllowanceObservationIsGlobal(t *testing.T) {
	redisURL := requireProviderSharedBudgetRedisURL(t)
	namespace := fmt.Sprintf("openinvest:test:m12:provider-headers:%d", time.Now().UnixNano())
	var upstream atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		upstream.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Reset", "60")
		_, _ = w.Write([]byte(`{"dividends":[]}`))
	}))
	defer server.Close()

	firstAuthority := newProviderSharedAuthority(t, redisURL, namespace)
	secondAuthority := newProviderSharedAuthority(t, redisURL, namespace)
	first, err := newCorporateActionProviderWithSharedBudget(&http.Client{}, fixedClock{now: testNow}, testToken, server.URL+"/rest", firstAuthority)
	if err != nil {
		t.Fatal(err)
	}
	second, err := newCorporateActionProviderWithSharedBudget(&http.Client{}, fixedClock{now: testNow}, testToken, server.URL+"/rest", secondAuthority)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := first.CorporateActions(context.Background(), sharedProviderQuery()); err != nil {
		t.Fatalf("first provider call: %v", err)
	}
	if _, err := second.CorporateActions(context.Background(), sharedProviderQuery()); !errors.Is(err, verticalslice.ErrCorporateActionsProviderUnavailable) {
		t.Fatalf("second provider ignored shared remote allowance: %v", err)
	}
	if upstream.Load() != 1 {
		t.Fatalf("shared provider allowance upstream calls=%d want=1", upstream.Load())
	}
	t.Log("M12_SHARED_PROVIDER_REMOTE_ALLOWANCE=GLOBAL_CONSERVATIVE")
}

// These checks exercise repository-owned operation lifetime, not termination
// of computations performed by an external server after a connection closes.
func TestCancelledProviderWorkRemainsBoundedByConcurrency(t *testing.T) {
	baselineG := runtime.NumGoroutine()
	baselineFD := providerRegressionFDCount()
	peakG, peakFD := baselineG, baselineFD
	maximumSlots := 0
	const rounds = 10
	for round := 0; round < rounds; round++ {
		entered := make(chan struct{}, maxConcurrency)
		release := make(chan struct{})
		var active atomic.Int32
		transport := operationLifetimeTransport{entered: entered, release: release, active: &active}
		provider, err := newCorporateActionProvider(&http.Client{Transport: transport}, fixedClock{now: testNow}, testToken, "http://operation.test/rest")
		if err != nil {
			t.Fatal(err)
		}
		results := make(chan error, maxConcurrency)
		cancels := make([]context.CancelFunc, maxConcurrency)
		for i := 0; i < maxConcurrency; i++ {
			ctx, cancel := context.WithCancel(context.Background())
			cancels[i] = cancel
			go func() { _, err := provider.CorporateActions(ctx, sharedProviderQuery()); results <- err }()
		}
		for i := 0; i < maxConcurrency; i++ {
			select {
			case <-entered:
			case <-time.After(2 * time.Second):
				close(release)
				t.Fatal("operation did not enter transport")
			}
		}
		for _, cancel := range cancels {
			cancel()
		}
		for i := 0; i < maxConcurrency; i++ {
			select {
			case err := <-results:
				if !errors.Is(err, context.Canceled) {
					close(release)
					t.Fatalf("caller cancellation: %v", err)
				}
			case <-time.After(time.Second):
				close(release)
				t.Fatal("caller cancellation blocked")
			}
		}
		if got := len(provider.semaphore); got != maxConcurrency {
			close(release)
			t.Fatalf("slot released before RoundTrip terminal: %d", got)
		}
		if round == 0 {
			time.Sleep(requestTimeout + time.Second)
			if len(provider.semaphore) != maxConcurrency {
				close(release)
				t.Fatal("elapsed timeout released an unterminated local operation")
			}
		}
		maximumSlots = len(provider.semaphore)
		if active.Load() != maxConcurrency {
			close(release)
			t.Fatal("local operation accounting mismatch")
		}
		if _, err := provider.CorporateActions(context.Background(), sharedProviderQuery()); !errors.Is(err, verticalslice.ErrCorporateActionsProviderUnavailable) {
			close(release)
			t.Fatalf("occupied transport accepted replacement: %v", err)
		}
		if got := runtime.NumGoroutine(); got > peakG {
			peakG = got
		}
		if got := providerRegressionFDCount(); got > peakFD {
			peakFD = got
		}
		close(release) // Explicit, observed RoundTrip terminal condition.
		deadline := time.Now().Add(time.Second)
		for len(provider.semaphore) != 0 && time.Now().Before(deadline) {
			time.Sleep(time.Millisecond)
		}
		if len(provider.semaphore) != 0 || active.Load() != 0 {
			t.Fatal("local transport did not recover")
		}
	}
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	if got := runtime.NumGoroutine(); got > baselineG+maxConcurrency+4 {
		t.Fatalf("goroutine recovery grew: baseline=%d recovery=%d", baselineG, got)
	}
	if got := providerRegressionFDCount(); baselineFD >= 0 && got > baselineFD+2 {
		t.Fatalf("FD recovery grew: baseline=%d recovery=%d", baselineFD, got)
	}
	t.Logf("TOTAL_CANCELLATIONS=%d MAX_LOCAL_PROVIDER_TRANSPORTS=%d MAX_SEMAPHORE_IN_USE=%d GOROUTINES_BASELINE=%d GOROUTINES_PEAK=%d GOROUTINES_RECOVERY=%d FD_BASELINE=%d FD_PEAK=%d FD_RECOVERY=%d", rounds*maxConcurrency, maxConcurrency, maximumSlots, baselineG, peakG, runtime.NumGoroutine(), baselineFD, peakFD, providerRegressionFDCount())
}

type operationLifetimeTransport struct {
	entered chan<- struct{}
	release <-chan struct{}
	active  *atomic.Int32
}

func (transport operationLifetimeTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	transport.active.Add(1)
	defer transport.active.Add(-1)
	transport.entered <- struct{}{}
	<-request.Context().Done()
	<-transport.release
	return nil, request.Context().Err()
}

func providerRegressionFDCount() int {
	entries, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		return -1
	}
	return len(entries)
}

type transportFailureMode int32

const (
	transportSuccess transportFailureMode = iota
	transportPartial
	transportReset
	transportTimeout
)

func TestProviderPartialResetTimeoutStillFailClosedAndRecover(t *testing.T) {
	var mode atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch transportFailureMode(mode.Load()) {
		case transportPartial:
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
		case transportReset:
			hijacker, ok := w.(http.Hijacker)
			if !ok {
				return
			}
			conn, _, err := hijacker.Hijack()
			if err == nil {
				_ = conn.Close()
			}
		case transportTimeout:
			select {
			case <-r.Context().Done():
			case <-time.After(requestTimeout + time.Second):
			}
		default:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"dividends":[]}`))
		}
	}))
	defer server.Close()
	provider, err := newCorporateActionProvider(&http.Client{}, fixedClock{now: testNow}, testToken, server.URL+"/rest")
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		name string
		mode transportFailureMode
	}{
		{name: "partial", mode: transportPartial},
		{name: "reset", mode: transportReset},
		{name: "timeout", mode: transportTimeout},
	} {
		t.Run(test.name, func(t *testing.T) {
			mode.Store(int32(test.mode))
			if _, err := provider.CorporateActions(context.Background(), sharedProviderQuery()); !errors.Is(err, verticalslice.ErrCorporateActionsProviderUnavailable) {
				t.Fatalf("%s error=%v", test.name, err)
			}
			mode.Store(int32(transportSuccess))
			if _, err := provider.CorporateActions(context.Background(), sharedProviderQuery()); err != nil {
				t.Fatalf("%s recovery: %v", test.name, err)
			}
		})
	}
	t.Log("M12_PARTIAL_RESPONSE_RECOVERY=PASS M12_CONNECTION_RESET_RECOVERY=PASS M12_UPSTREAM_TIMEOUT_RECOVERY=PASS")
}
