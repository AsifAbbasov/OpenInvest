package httpapi

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/openinvest/openinvest/backend-go/internal/sharedbudget"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

type sharedBudgetCountingProvider struct {
	calls atomic.Int64
}

func (provider *sharedBudgetCountingProvider) CorporateActions(context.Context, verticalslice.CorporateActionQuery) ([]verticalslice.CorporateActionEvent, error) {
	provider.calls.Add(1)
	return nil, nil
}

type sharedBudgetHTTPReplica struct {
	app       *fiber.App
	baseURL   string
	provider  *sharedBudgetCountingProvider
	authority sharedbudget.Authority
}

func newSharedBudgetHTTPReplica(t *testing.T, redisURL, namespace string, config HTTPNetworkConfig) *sharedBudgetHTTPReplica {
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
	provider := &sharedBudgetCountingProvider{}
	api := &API{
		corporateActionProvider:  provider,
		corporateActionBudget:    authority,
		corporateActionLimiter:   newCorporateActionProjectionRateLimiter(),
		corporateActionCoalescer: newCorporateActionProjectionCoalescer(),
		httpNetworkConfig:        config,
	}
	app := newFiberApp(config)
	app.Get("/api/v1/corporate-actions/projection", api.getCorporateActionProjection)
	baseURL := startSharedBudgetHTTPServer(t, app)
	replica := &sharedBudgetHTTPReplica{
		app:       app,
		baseURL:   baseURL,
		provider:  provider,
		authority: authority,
	}
	t.Cleanup(func() {
		_ = authority.Close()
	})
	return replica
}

func startSharedBudgetHTTPServer(t *testing.T, app *fiber.App) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- app.Listener(listener)
	}()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = app.ShutdownWithContext(ctx)
		_ = listener.Close()
		select {
		case <-errCh:
		case <-time.After(500 * time.Millisecond):
		}
	})
	return "http://" + listener.Addr().String()
}

func sharedBudgetProjectionStatus(t *testing.T, baseURL, clientIP string) int {
	t.Helper()
	request, err := http.NewRequest(
		http.MethodGet,
		baseURL+"/api/v1/corporate-actions/projection?instrumentId=SBER&from=2026-09-01&to=2026-09-30",
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("X-Forwarded-For", clientIP)
	response, err := (&http.Client{Timeout: 3 * time.Second}).Do(request)
	if err != nil {
		t.Fatalf("projection request: %v", err)
	}
	defer response.Body.Close()
	return response.StatusCode
}

func requireSharedBudgetRedisURL(t *testing.T) string {
	t.Helper()
	url := os.Getenv("OPENINVEST_SHARED_BUDGET_TEST_REDIS_URL")
	if url == "" {
		t.Skip("OPENINVEST_SHARED_BUDGET_TEST_REDIS_URL is required")
	}
	return url
}

func TestCorporateActionSharedBudgetAcrossReplicasAndRestart(t *testing.T) {
	redisURL := requireSharedBudgetRedisURL(t)
	trusted, err := NewHTTPNetworkConfig(true, []string{"127.0.0.1/32"})
	if err != nil {
		t.Fatal(err)
	}

	for _, instances := range []int{1, 2, 4} {
		t.Run(fmt.Sprintf("instances-%d", instances), func(t *testing.T) {
			namespace := fmt.Sprintf("openinvest:test:m12:endpoint:%d:%d", instances, time.Now().UnixNano())
			replicas := make([]*sharedBudgetHTTPReplica, 0, instances)
			for i := 0; i < instances; i++ {
				replicas = append(replicas, newSharedBudgetHTTPReplica(t, redisURL, namespace, trusted))
			}

			var ok, limited, failed int
			const totalRequests = 64
			for i := 0; i < totalRequests; i++ {
				replica := replicas[i%len(replicas)]
				clientIP := fmt.Sprintf("198.51.%d.%d", instances, i+1)
				status := sharedBudgetProjectionStatus(t, replica.baseURL, clientIP)
				switch {
				case status >= 200 && status < 300:
					ok++
				case status == http.StatusTooManyRequests:
					limited++
				default:
					failed++
				}
			}
			var providerCalls int64
			for _, replica := range replicas {
				providerCalls += replica.provider.calls.Load()
			}
			if ok != defaultCorporateActionProjectionGlobalLimit || providerCalls != defaultCorporateActionProjectionGlobalLimit {
				t.Fatalf("instances=%d ok=%d provider_calls=%d want=%d", instances, ok, providerCalls, defaultCorporateActionProjectionGlobalLimit)
			}
			if limited != totalRequests-defaultCorporateActionProjectionGlobalLimit || failed != 0 {
				t.Fatalf("instances=%d limited=%d failed=%d", instances, limited, failed)
			}

			restarted := newSharedBudgetHTTPReplica(t, redisURL, namespace, trusted)
			restartStatus := sharedBudgetProjectionStatus(t, restarted.baseURL, "203.0.113.250")
			if restartStatus != http.StatusTooManyRequests {
				t.Fatalf("instances=%d restarted process reset shared budget: status=%d", instances, restartStatus)
			}
			if restarted.provider.calls.Load() != 0 {
				t.Fatalf("instances=%d restarted process reached provider", instances)
			}
			t.Logf("M12_SHARED_ENDPOINT instances=%d total_http=%d http_2xx=%d http_429=%d http_5xx=%d total_provider_calls=%d process_restart_shared_budget_reset=NO",
				instances, totalRequests+1, ok, limited+1, failed, providerCalls)
		})
	}
}

func TestCorporateActionSharedPerClientLimitAcrossReplicas(t *testing.T) {
	redisURL := requireSharedBudgetRedisURL(t)
	trusted, err := NewHTTPNetworkConfig(true, []string{"127.0.0.1/32"})
	if err != nil {
		t.Fatal(err)
	}
	namespace := fmt.Sprintf("openinvest:test:m12:client:%d", time.Now().UnixNano())
	replicas := make([]*sharedBudgetHTTPReplica, 0, 4)
	for i := 0; i < 4; i++ {
		replicas = append(replicas, newSharedBudgetHTTPReplica(t, redisURL, namespace, trusted))
	}
	var accepted int
	for i := 0; i < 20; i++ {
		status := sharedBudgetProjectionStatus(t, replicas[i%len(replicas)].baseURL, "198.51.100.77")
		if status == http.StatusOK {
			accepted++
		} else if status != http.StatusTooManyRequests {
			t.Fatalf("unexpected status=%d", status)
		}
	}
	if accepted != defaultCorporateActionProjectionPerClientLimit {
		t.Fatalf("same-client accepted=%d want=%d", accepted, defaultCorporateActionProjectionPerClientLimit)
	}
	var calls int64
	for _, replica := range replicas {
		calls += replica.provider.calls.Load()
	}
	if calls != defaultCorporateActionProjectionPerClientLimit {
		t.Fatalf("same-client provider calls=%d want=%d", calls, defaultCorporateActionProjectionPerClientLimit)
	}
	t.Logf("M12_SHARED_PER_CLIENT_LIMIT_ACROSS_REPLICAS=%d", accepted)
}

func TestCorporateActionSharedBudgetBackendFailureFailsClosed(t *testing.T) {
	redisURL := requireSharedBudgetRedisURL(t)
	authority, err := sharedbudget.NewRedisAuthority(redisURL, fmt.Sprintf("openinvest:test:m12:failure:%d", time.Now().UnixNano()))
	if err != nil {
		t.Fatal(err)
	}
	if err := authority.Close(); err != nil {
		t.Fatal(err)
	}
	provider := &sharedBudgetCountingProvider{}
	api := &API{
		corporateActionProvider:  provider,
		corporateActionBudget:    authority,
		corporateActionLimiter:   newCorporateActionProjectionRateLimiter(),
		corporateActionCoalescer: newCorporateActionProjectionCoalescer(),
	}
	app := newFiberApp(HTTPNetworkConfig{})
	app.Get("/api/v1/corporate-actions/projection", api.getCorporateActionProjection)
	baseURL := startSharedBudgetHTTPServer(t, app)
	status := sharedBudgetProjectionStatus(t, baseURL, "127.0.0.1")
	if status != http.StatusServiceUnavailable {
		t.Fatalf("shared budget outage status=%d want=%d", status, http.StatusServiceUnavailable)
	}
	if provider.calls.Load() != 0 {
		t.Fatalf("shared budget outage provider calls=%d want=0", provider.calls.Load())
	}
	t.Log("M12_SHARED_BUDGET_BACKEND_UNAVAILABLE=FAIL_CLOSED provider_calls=0")
}
