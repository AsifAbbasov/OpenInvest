package httpapi

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/openinvest/openinvest/backend-go/internal/auth"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

type module12CountingProvider struct {
	calls atomic.Int64
}

func (p *module12CountingProvider) CorporateActions(context.Context, verticalslice.CorporateActionQuery) ([]verticalslice.CorporateActionEvent, error) {
	p.calls.Add(1)
	return nil, nil
}

type module12Replica struct {
	app      *fiber.App
	baseURL  string
	provider *module12CountingProvider
	nowNanos atomic.Int64
}

func module12NewReplica(t *testing.T, networkConfig HTTPNetworkConfig) *module12Replica {
	t.Helper()
	provider := &module12CountingProvider{}
	replica := &module12Replica{provider: provider}
	replica.nowNanos.Store(time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC).UnixNano())
	api := &API{
		corporateActionProvider:  provider,
		corporateActionLimiter:   newCorporateActionProjectionRateLimiter(),
		corporateActionCoalescer: newCorporateActionProjectionCoalescer(),
		httpNetworkConfig:        networkConfig,
		now: func() time.Time {
			return time.Unix(0, replica.nowNanos.Load()).UTC()
		},
	}
	replica.app = newApp(api)
	replica.baseURL = module12StartFiberApp(t, replica.app)
	return replica
}

func module12StartFiberApp(t *testing.T, app *fiber.App) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- app.Listener(listener)
	}()
	baseURL := "http://" + listener.Addr().String()
	client := &http.Client{Timeout: 500 * time.Millisecond}
	deadline := time.Now().Add(3 * time.Second)
	for {
		resp, requestErr := client.Get(baseURL + "/api/v1/health")
		if requestErr == nil {
			_ = resp.Body.Close()
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("server did not become ready: %v", requestErr)
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = app.ShutdownWithContext(ctx)
		_ = listener.Close()
		select {
		case <-errCh:
		case <-time.After(500 * time.Millisecond):
		}
	})
	return baseURL
}

func module12ProjectionStatus(t *testing.T, baseURL string, forwardedFor ...string) int {
	t.Helper()
	request, err := http.NewRequest(http.MethodGet, baseURL+"/api/v1/corporate-actions/projection?instrumentId=SBER&from=2026-09-01&to=2026-09-30", nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range forwardedFor {
		request.Header.Add("X-Forwarded-For", value)
	}
	request.Header.Add("Forwarded", "for=192.0.2.200")
	request.Header.Add("CF-Connecting-IP", "192.0.2.201")
	response, err := (&http.Client{Timeout: 3 * time.Second}).Do(request)
	if err != nil {
		t.Fatalf("projection request: %v", err)
	}
	defer response.Body.Close()
	return response.StatusCode
}

func TestM12MultiInstanceHTTPBudgetAndClientIPBoundary(t *testing.T) {
	trusted, err := NewHTTPNetworkConfig(true, []string{"127.0.0.1/32"})
	if err != nil {
		t.Fatal(err)
	}

	for _, instances := range []int{1, 2, 4} {
		t.Run(fmt.Sprintf("instances-%d", instances), func(t *testing.T) {
			replicas := make([]*module12Replica, 0, instances)
			for i := 0; i < instances; i++ {
				replicas = append(replicas, module12NewReplica(t, trusted))
			}
			var ok, limited, failed int
			const requestsPerReplica = 60
			for replicaIndex, replica := range replicas {
				for i := 0; i < requestsPerReplica; i++ {
					clientIP := fmt.Sprintf("198.51.%d.%d", 100+replicaIndex, i+1)
					switch status := module12ProjectionStatus(t, replica.baseURL, clientIP); {
					case status >= 200 && status < 300:
						ok++
					case status == http.StatusTooManyRequests:
						limited++
					default:
						failed++
					}
				}
			}
			var providerCalls int64
			for _, replica := range replicas {
				providerCalls += replica.provider.calls.Load()
			}
			wantCalls := int64(defaultCorporateActionProjectionGlobalLimit * instances)
			if providerCalls != wantCalls || ok != int(wantCalls) || limited != (requestsPerReplica-defaultCorporateActionProjectionGlobalLimit)*instances || failed != 0 {
				t.Fatalf("unexpected multi-instance result: calls=%d ok=%d limited=%d failed=%d", providerCalls, ok, limited, failed)
			}
			recovered := 0
			for _, replica := range replicas {
				replica.nowNanos.Add(int64(61 * time.Second))
				if status := module12ProjectionStatus(t, replica.baseURL, "203.0.113.250"); status == http.StatusOK {
					recovered++
				}
			}
			t.Logf("M12_MULTI_INSTANCE instances=%d total_http=%d http_2xx=%d http_429=%d http_5xx=%d total_provider_calls=%d global_budget_multiplication=%dx recovery_after_window=%d/%d",
				instances, requestsPerReplica*instances, ok, limited, failed, providerCalls, instances, recovered, instances)
		})
	}

	direct := module12NewReplica(t, HTTPNetworkConfig{})
	for i := 0; i < 13; i++ {
		status := module12ProjectionStatus(t, direct.baseURL,
			fmt.Sprintf("203.0.113.%d", i+1),
			fmt.Sprintf("198.51.100.%d", i+1),
		)
		if i < 12 && status != http.StatusOK {
			t.Fatalf("direct mode request %d status=%d", i+1, status)
		}
		if i == 12 && status != http.StatusTooManyRequests {
			t.Fatalf("direct mode rotating forwarding headers bypassed per-client limit: status=%d", status)
		}
	}
	if direct.provider.calls.Load() != 12 {
		t.Fatalf("direct-mode provider calls=%d want=12", direct.provider.calls.Load())
	}
	t.Log("M12_DIRECT_MODE_ROTATING_AND_DUPLICATE_XFF_BYPASS=NO provider_calls=12")
}

func TestM12CookieCORSAndCacheBrowserBoundary(t *testing.T) {
	t.Setenv("OPENINVEST_ALLOWED_WEB_ORIGINS", "https://app.example.test")
	store := &httpAuthTestStore{}
	app := newHTTPAuthApp(t, newHTTPAuthService(t, store))
	register := securityAuthRequest(t, app, http.MethodPost, "/api/v1/auth/register", `{
		"email":"module12-browser@example.com",
		"password":"correct horse battery staple",
		"language":"en",
		"theme":"system",
		"timezone":"UTC"
	}`, "", "")
	defer register.Body.Close()
	cookie := requireCookie(t, register, auth.RefreshCookieName)
	if !cookie.Secure || !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode || cookie.Path != "/api/v1/auth" {
		t.Fatalf("refresh cookie attributes: secure=%t httponly=%t samesite=%v path=%q", cookie.Secure, cookie.HttpOnly, cookie.SameSite, cookie.Path)
	}
	if got := register.Header.Get("Cache-Control"); got != "no-store" {
		t.Fatalf("auth cache-control=%q", got)
	}
	t.Log("M12_REFRESH_COOKIE secure=YES httponly=YES samesite=STRICT path=/api/v1/auth cache_control=NO_STORE")

	preflight := func(origin string) *http.Response {
		request, err := http.NewRequest(http.MethodOptions, "http://example.invalid/api/v1/auth/refresh", nil)
		if err != nil {
			t.Fatal(err)
		}
		request.Header.Set("Origin", origin)
		request.Header.Set("Access-Control-Request-Method", http.MethodPost)
		response, err := app.Test(request)
		if err != nil {
			t.Fatal(err)
		}
		return response
	}
	allowed := preflight("https://app.example.test")
	if allowed.StatusCode != http.StatusNoContent || allowed.Header.Get("Access-Control-Allow-Origin") != "https://app.example.test" || allowed.Header.Get("Access-Control-Allow-Credentials") != "true" {
		t.Fatalf("allowed CORS response: status=%d origin=%q credentials=%q", allowed.StatusCode, allowed.Header.Get("Access-Control-Allow-Origin"), allowed.Header.Get("Access-Control-Allow-Credentials"))
	}
	allowed.Body.Close()
	blocked := preflight("https://evil.example.test")
	if blocked.Header.Get("Access-Control-Allow-Origin") != "" || blocked.Header.Get("Access-Control-Allow-Credentials") != "" {
		t.Fatalf("disallowed origin received CORS trust: origin=%q credentials=%q", blocked.Header.Get("Access-Control-Allow-Origin"), blocked.Header.Get("Access-Control-Allow-Credentials"))
	}
	blocked.Body.Close()
	t.Log("M12_CORS_EXACT_ORIGIN_ALLOWLIST=PASS M12_CORS_DISALLOWED_ORIGIN_CREDENTIALS=NO")

	t.Setenv("OPENINVEST_ALLOWED_WEB_ORIGINS", "")
	defaultLocal := preflight("http://localhost:3000")
	defaultAllowed := defaultLocal.Header.Get("Access-Control-Allow-Origin") == "http://localhost:3000"
	defaultLocal.Body.Close()
	t.Logf("M12_PRODUCTION_CONFIG_MISSING_ALLOWED_ORIGINS_DEFAULTS_LOCALHOST=%t", defaultAllowed)
}

type module12BlockingProvider struct {
	calls      atomic.Int64
	entered    chan struct{}
	release    chan struct{}
	cancelled  chan struct{}
	enterOnce  sync.Once
	cancelOnce sync.Once
}

func (p *module12BlockingProvider) CorporateActions(ctx context.Context, _ verticalslice.CorporateActionQuery) ([]verticalslice.CorporateActionEvent, error) {
	p.calls.Add(1)
	p.enterOnce.Do(func() { close(p.entered) })
	select {
	case <-ctx.Done():
		p.cancelOnce.Do(func() { close(p.cancelled) })
		return nil, ctx.Err()
	case <-p.release:
		return nil, nil
	}
}

func module12FDCount() int {
	entries, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		return -1
	}
	return len(entries)
}

func TestM12RealSocketDisconnectAndSlowClientBehavior(t *testing.T) {
	baselineG := runtime.NumGoroutine()
	baselineFD := module12FDCount()

	blocking := &module12BlockingProvider{
		entered:   make(chan struct{}),
		release:   make(chan struct{}),
		cancelled: make(chan struct{}),
	}
	lifecycle := NewRequestLifecycle()
	api := &API{
		corporateActionProvider:  blocking,
		corporateActionLimiter:   newCorporateActionProjectionRateLimiter(),
		corporateActionCoalescer: newCorporateActionProjectionCoalescer(),
		requestLifecycle:         lifecycle,
	}
	app := newReplayApp(api)
	baseURL := module12StartFiberApp(t, app)
	addr := baseURL[len("http://"):]

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = fmt.Fprintf(conn, "GET /api/v1/corporate-actions/projection?instrumentId=SBER&from=2026-09-01&to=2026-09-30 HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n")
	select {
	case <-blocking.entered:
	case <-time.After(2 * time.Second):
		t.Fatal("provider call did not start")
	}
	peakG := runtime.NumGoroutine()
	peakFD := module12FDCount()
	_ = conn.Close()

	cancelObserved := false
	select {
	case <-blocking.cancelled:
		cancelObserved = true
	case <-time.After(750 * time.Millisecond):
	}
	t.Logf("M12_DOWNSTREAM_DISCONNECT_DURING_PROVIDER_PROPAGATES_CONTEXT=%t", cancelObserved)
	close(blocking.release)

	waitCtx, waitCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer waitCancel()
	if err := lifecycle.WaitContext(waitCtx); err != nil {
		t.Fatalf("request lifecycle did not recover: %v", err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for api.corporateActionCoalescer.activeCalls() != 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if api.corporateActionCoalescer.activeCalls() != 0 {
		t.Fatal("coalescer did not recover after disconnected request")
	}

	counting := &module12CountingProvider{}
	partialAPI := &API{
		corporateActionProvider:  counting,
		corporateActionLimiter:   newCorporateActionProjectionRateLimiter(),
		corporateActionCoalescer: newCorporateActionProjectionCoalescer(),
	}
	partialApp := newApp(partialAPI)
	partialURL := module12StartFiberApp(t, partialApp)
	partialAddr := partialURL[len("http://"):]

	beforeConn, err := net.Dial("tcp", partialAddr)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = fmt.Fprint(beforeConn, "GET /api/v1/corporate-actions/projection?instrumentId=SBER&from=2026-09-01&to=2026-09-30 HTTP/1.1\r\nHost: localhost\r\n")
	_ = beforeConn.Close()
	time.Sleep(250 * time.Millisecond)
	t.Logf("M12_DISCONNECT_BEFORE_COMPLETE_HTTP_REQUEST_PROVIDER_CALLS=%d", counting.calls.Load())
	if counting.calls.Load() != 0 {
		t.Fatal("incomplete disconnected request reached provider")
	}

	slowConn, err := net.Dial("tcp", partialAddr)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = fmt.Fprint(slowConn, "GET /api/v1/health HTTP/1.1\r\nHost: localhost\r\n")
	start := time.Now()
	_ = slowConn.SetReadDeadline(time.Now().Add(17 * time.Second))
	var one [1]byte
	_, readErr := slowConn.Read(one[:])
	elapsed := time.Since(start)
	_ = slowConn.Close()
	closedWithinBound := readErr != nil && elapsed < 17*time.Second
	t.Logf("M12_SLOWLORIS_BOUNDED=%t elapsed=%s read_error=%v configured_read_timeout=%s", closedWithinBound, elapsed.Round(time.Millisecond), readErr, defaultHTTPReadTimeout)

	runtime.GC()
	time.Sleep(250 * time.Millisecond)
	t.Logf("M12_SOCKET_RESOURCES goroutines_baseline=%d goroutines_peak=%d goroutines_recovery=%d fd_baseline=%d fd_peak=%d fd_recovery=%d",
		baselineG, peakG, runtime.NumGoroutine(), baselineFD, peakFD, module12FDCount())
}
