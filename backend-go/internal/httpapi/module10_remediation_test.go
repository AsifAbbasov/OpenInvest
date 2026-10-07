package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/valyala/fasthttp"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

type module10RemediationProvider struct {
	calls   atomic.Int64
	active  atomic.Int64
	max     atomic.Int64
	block   <-chan struct{}
	started chan<- struct{}
	fail    atomic.Bool
}

func (p *module10RemediationProvider) CorporateActions(
	ctx context.Context,
	query verticalslice.CorporateActionQuery,
) ([]verticalslice.CorporateActionEvent, error) {
	p.calls.Add(1)
	active := p.active.Add(1)
	for {
		current := p.max.Load()
		if active <= current || p.max.CompareAndSwap(current, active) {
			break
		}
	}
	defer p.active.Add(-1)
	if p.started != nil {
		select {
		case p.started <- struct{}{}:
		default:
		}
	}
	if p.block != nil {
		select {
		case <-p.block:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if p.fail.Load() {
		return nil, verticalslice.ErrCorporateActionsProviderUnavailable
	}
	return nil, nil
}

func module10RemediationApp(provider verticalslice.CorporateActionProvider) (*API, *fiber.App) {
	api := &API{
		corporateActionProvider:  provider,
		corporateActionLimiter:   newCorporateActionProjectionRateLimiter(),
		corporateActionCoalescer: newCorporateActionProjectionCoalescer(),
	}
	return api, newReplayApp(api)
}

type module10RemediationResponse struct {
	StatusCode int
	RetryAfter string
}

func module10RemediationRequest(
	t *testing.T,
	app *fiber.App,
	peer string,
	path string,
	xff string,
) module10RemediationResponse {
	t.Helper()
	remoteIP := net.ParseIP(peer)
	if remoteIP == nil {
		t.Fatalf("invalid test peer IP %q", peer)
	}
	var request fasthttp.Request
	request.Header.SetMethod(http.MethodGet)
	request.SetRequestURI(path)
	if xff != "" {
		request.Header.Set(fiber.HeaderXForwardedFor, xff)
	}
	var requestCtx fasthttp.RequestCtx
	requestCtx.Init(&request, &net.TCPAddr{IP: remoteIP, Port: 5000}, nil)
	app.Handler()(&requestCtx)
	return module10RemediationResponse{
		StatusCode: requestCtx.Response.StatusCode(),
		RetryAfter: string(requestCtx.Response.Header.Peek("Retry-After")),
	}
}

func module10RemediationPath(instruments, from, to string) string {
	return fmt.Sprintf("/api/v1/corporate-actions/projection?instrumentId=%s&from=%s&to=%s", instruments, from, to)
}

func TestModule10RemediationSequentialAnonymousAttackAndLegitimateSurvival(t *testing.T) {
	provider := &module10RemediationProvider{}
	_, app := module10RemediationApp(provider)
	path := module10RemediationPath("SBER", "2026-01-01", "2026-12-31")
	counts := map[int]int{}
	for i := 0; i < 60; i++ {
		res := module10RemediationRequest(t, app, "203.0.113.10", path, "")
		counts[res.StatusCode]++
		if res.StatusCode == http.StatusTooManyRequests && res.RetryAfter != "60" {
			t.Fatalf("Retry-After=%q want=60", res.RetryAfter)
		}
	}
	if got := provider.calls.Load(); got != defaultCorporateActionProjectionPerClientLimit {
		t.Fatalf("provider calls after 60 anonymous requests=%d want=%d", got, defaultCorporateActionProjectionPerClientLimit)
	}
	before := provider.calls.Load()
	legit := module10RemediationRequest(
		t,
		app,
		"198.51.100.20",
		module10RemediationPath("GAZP", "2026-01-01", "2026-12-31"),
		"",
	)
	if legit.StatusCode != http.StatusOK {
		t.Fatalf("legitimate status=%d want=200", legit.StatusCode)
	}
	if provider.calls.Load() != before+1 {
		t.Fatalf("legitimate provider call missing: before=%d after=%d", before, provider.calls.Load())
	}
	t.Logf("MODULE10_REMEDIATION_ATTACK anonymous=60 http_200=%d http_429=%d http_503=%d provider_calls=%d legitimate_status=%d legitimate_provider_call=true",
		counts[http.StatusOK], counts[http.StatusTooManyRequests], counts[http.StatusServiceUnavailable],
		provider.calls.Load()-1, legit.StatusCode)
}

func TestModule10RemediationIdenticalConcurrentCoalescing(t *testing.T) {
	for _, clients := range []int{10, 50, 100} {
		t.Run(fmt.Sprintf("clients_%d", clients), func(t *testing.T) {
			release := make(chan struct{})
			started := make(chan struct{}, 1)
			provider := &module10RemediationProvider{block: release, started: started}
			api, app := module10RemediationApp(provider)
			path := module10RemediationPath("SBER", "2026-01-01", "2026-12-31")
			keyQuery := verticalslice.CorporateActionQuery{InstrumentIDs: []string{"SBER"}, From: "2026-01-01", To: "2026-12-31"}
			key := corporateActionProjectionRequestKey(keyQuery)

			statuses := make(chan int, clients)
			var wg sync.WaitGroup
			for i := 0; i < clients; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					res := module10RemediationRequest(t, app, "203.0.113.11", path, "")
					statuses <- res.StatusCode
				}()
			}
			select {
			case <-started:
			case <-time.After(time.Second):
				t.Fatal("provider did not start")
			}

			wantWaiters := clients
			if wantWaiters > defaultCorporateActionProjectionPerClientLimit {
				wantWaiters = defaultCorporateActionProjectionPerClientLimit
			}
			deadline := time.Now().Add(time.Second)
			for {
				api.corporateActionCoalescer.mu.Lock()
				waiters := 0
				if call := api.corporateActionCoalescer.calls[key]; call != nil {
					waiters = call.waiters
				}
				api.corporateActionCoalescer.mu.Unlock()
				if waiters >= wantWaiters {
					break
				}
				if time.Now().After(deadline) {
					t.Fatalf("coalesced waiters=%d want=%d", waiters, wantWaiters)
				}
				time.Sleep(time.Millisecond)
			}
			close(release)
			wg.Wait()
			close(statuses)

			counts := map[int]int{}
			for status := range statuses {
				counts[status]++
			}
			if got := provider.calls.Load(); got != 1 {
				t.Fatalf("provider calls=%d want=1", got)
			}
			if got := api.corporateActionCoalescer.activeCalls(); got != 0 {
				t.Fatalf("active coalescer calls=%d want=0", got)
			}
			t.Logf("MODULE10_REMEDIATION_IDENTICAL clients=%d provider_calls=%d http_200=%d http_429=%d max_active=%d",
				clients, provider.calls.Load(), counts[http.StatusOK], counts[http.StatusTooManyRequests], provider.max.Load())
		})
	}
}

func TestModule10RemediationDistinctAndCanonicalEquivalentAttacks(t *testing.T) {
	t.Run("distinct_valid", func(t *testing.T) {
		provider := &module10RemediationProvider{}
		_, app := module10RemediationApp(provider)
		counts := map[int]int{}
		base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		for i := 0; i < 60; i++ {
			day := base.AddDate(0, 0, i)
			date := day.Format("2006-01-02")
			res := module10RemediationRequest(t, app, "203.0.113.12", module10RemediationPath("SBER", date, date), "")
			counts[res.StatusCode]++
		}
		if got := provider.calls.Load(); got != defaultCorporateActionProjectionPerClientLimit {
			t.Fatalf("distinct-query provider calls=%d want=%d", got, defaultCorporateActionProjectionPerClientLimit)
		}
		t.Logf("MODULE10_REMEDIATION_DISTINCT requests=60 provider_calls=%d http_200=%d http_429=%d",
			provider.calls.Load(), counts[http.StatusOK], counts[http.StatusTooManyRequests])
	})

	t.Run("canonical_equivalent_order", func(t *testing.T) {
		release := make(chan struct{})
		started := make(chan struct{}, 1)
		provider := &module10RemediationProvider{block: release, started: started}
		_, app := module10RemediationApp(provider)
		paths := []string{
			module10RemediationPath("SBER,GAZP", "2026-01-01", "2026-12-31"),
			module10RemediationPath("GAZP,SBER", "2026-01-01", "2026-12-31"),
		}
		var wg sync.WaitGroup
		statuses := make(chan int, 10)
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				res := module10RemediationRequest(t, app, "203.0.113.13", paths[i%2], "")
				statuses <- res.StatusCode
			}(i)
		}
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("provider did not start")
		}
		time.Sleep(20 * time.Millisecond)
		close(release)
		wg.Wait()
		close(statuses)
		for status := range statuses {
			if status != http.StatusOK {
				t.Fatalf("canonical-equivalent status=%d want=200", status)
			}
		}
		if got := provider.calls.Load(); got != 1 {
			t.Fatalf("canonical-equivalent provider calls=%d want=1", got)
		}
		t.Logf("MODULE10_REMEDIATION_CANONICAL_EQUIVALENT clients=10 provider_calls=%d", provider.calls.Load())
	})
}

func TestModule10RemediationNATForwardingAndCardinalityBounds(t *testing.T) {
	t.Run("shared_nat_normal_flow", func(t *testing.T) {
		provider := &module10RemediationProvider{}
		_, app := module10RemediationApp(provider)
		path := module10RemediationPath("SBER", "2026-01-01", "2026-12-31")
		for i := 0; i < 12; i++ {
			res := module10RemediationRequest(t, app, "203.0.113.14", path, "")
			if res.StatusCode != http.StatusOK {
				t.Fatalf("normal shared-NAT request %d status=%d", i+1, res.StatusCode)
			}
		}
		res := module10RemediationRequest(t, app, "203.0.113.14", path, "")
		if res.StatusCode != http.StatusTooManyRequests {
			t.Fatalf("shared-NAT request 13 status=%d want=429", res.StatusCode)
		}
	})

	t.Run("direct_mode_ignores_forwarded_header_rotation", func(t *testing.T) {
		provider := &module10RemediationProvider{}
		_, app := module10RemediationApp(provider)
		path := module10RemediationPath("SBER", "2026-01-01", "2026-12-31")
		for i := 0; i < 13; i++ {
			res := module10RemediationRequest(t, app, "203.0.113.15", path, fmt.Sprintf("198.51.100.%d", i+1))
			if i < 12 && res.StatusCode != http.StatusOK {
				t.Fatalf("forwarded rotation request %d status=%d want=200", i+1, res.StatusCode)
			}
			if i == 12 && res.StatusCode != http.StatusTooManyRequests {
				t.Fatalf("forwarded rotation request %d status=%d want=429", i+1, res.StatusCode)
			}
		}
	})

	t.Run("bounded_high_cardinality_state", func(t *testing.T) {
		limiter := newCorporateActionProjectionRateLimiter()
		now := time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC)
		for i := 0; i < 10000; i++ {
			_ = limiter.allow(fmt.Sprintf("/api/v1/corporate-actions/projection|2001:db8::%x", i+1), now)
		}
		if got := len(limiter.attempts); got > defaultCorporateActionProjectionGlobalLimit {
			t.Fatalf("active limiter keys=%d exceed global admitted bound=%d", got, defaultCorporateActionProjectionGlobalLimit)
		}
		before := len(limiter.attempts)
		now = now.Add(defaultCorporateActionProjectionWindow + time.Second)
		if !limiter.allow("/api/v1/corporate-actions/projection|2001:db8::ffff", now) {
			t.Fatal("limiter did not recover after expiry")
		}
		after := len(limiter.attempts)
		if after != 1 {
			t.Fatalf("expired limiter entries after sweep=%d want=1", after)
		}
		t.Logf("MODULE10_REMEDIATION_CARDINALITY attempts=10000 active_before_expiry=%d active_after_expiry=%d max_keys=%d",
			before, after, defaultCorporateActionProjectionMaxClientKeys)
	})
}

func TestModule10RemediationCoalescerErrorAndCancellationRecovery(t *testing.T) {
	provider := &module10RemediationProvider{}
	api, app := module10RemediationApp(provider)
	path := module10RemediationPath("SBER", "2026-01-01", "2026-12-31")

	provider.fail.Store(true)
	failed := module10RemediationRequest(t, app, "203.0.113.16", path, "")
	if failed.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("provider failure status=%d want=503", failed.StatusCode)
	}
	if got := api.corporateActionCoalescer.activeCalls(); got != 0 {
		t.Fatalf("singleflight stuck after error: %d", got)
	}

	provider.fail.Store(false)
	recovered := module10RemediationRequest(t, app, "203.0.113.16", path, "")
	if recovered.StatusCode != http.StatusOK {
		t.Fatalf("recovery status=%d want=200", recovered.StatusCode)
	}
	if got := api.corporateActionCoalescer.activeCalls(); got != 0 {
		t.Fatalf("singleflight stuck after recovery: %d", got)
	}

	coalescer := newCorporateActionProjectionCoalescer()
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{})
	finished := make(chan error, 1)
	go func() {
		_, err := coalescer.do(ctx, "cancel-key", func(shared context.Context) (corporateActionProjectionDTO, error) {
			close(started)
			<-shared.Done()
			return corporateActionProjectionDTO{}, shared.Err()
		})
		finished <- err
	}()
	<-started
	cancel()
	select {
	case err := <-finished:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("cancel result=%v want context canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("cancelled coalesced request did not finish")
	}
	deadline := time.Now().Add(time.Second)
	for coalescer.activeCalls() != 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if got := coalescer.activeCalls(); got != 0 {
		t.Fatalf("singleflight stuck after cancellation: %d", got)
	}
}
