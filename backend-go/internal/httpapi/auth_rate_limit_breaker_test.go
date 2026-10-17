package httpapi

import (
	"github.com/gofiber/fiber/v3"
	"errors"
	"os"
	"testing"
	"time"
)

func requireSecurityBreaker(t *testing.T) {
	t.Helper()
	if os.Getenv("OPENINVEST_SECURITY_BREAKER_TESTS") != "1" {
		t.Skip("OPENINVEST_SECURITY_BREAKER_TESTS=1 is required")
	}
}

// This test encodes the production security invariant we actually want:
// one abusive client must not receive a fresh auth budget merely because
// load balancing sends the next request to a different application process.
func TestSecurityBreakerAuthRateLimitMustBeSharedAcrossInstances(t *testing.T) {
	requireSecurityBreaker(t)

	const perClientLimit = 2
	networkConfig := HTTPNetworkConfig{}
	apiA, appA := newOINew0506RateLimitApp(perClientLimit, networkConfig)
	apiB, appB := newOINew0506RateLimitApp(perClientLimit, networkConfig)
	peer := "198.51.100.77"

	admitted := 0
	for i := 0; i < 8; i++ {
		app := appA
		api := apiA
		if i%2 == 1 {
			app = appB
			api = apiB
		}
		_, err := oiNew0506RateLimitAttempt(app, api, peer, nil)
		switch {
		case err == nil:
			admitted++
		case errors.Is(err, errAuthRateLimited):
		default:
			t.Fatalf("unexpected limiter error on attempt %d: %v", i+1, err)
		}
	}

	if admitted > perClientLimit {
		t.Fatalf(
			"distributed auth rate-limit bypass reproduced: same client received %d admitted requests across two instances, secure budget=%d",
			admitted,
			perClientLimit,
		)
	}
}

func TestSecurityBreakerAuthRateLimitHeaderRotationCannotCreateFreshBudget(t *testing.T) {
	requireSecurityBreaker(t)

	api, app := newOINew0506RateLimitApp(2, HTTPNetworkConfig{})
	peer := "198.51.100.88"
	headers := []map[string]string{
		{"X-Forwarded-For": "203.0.113.1", "X-Real-IP": "203.0.113.2", "Forwarded": "for=203.0.113.3"},
		{"X-Forwarded-For": "203.0.113.4, 203.0.113.5", "CF-Connecting-IP": "203.0.113.6"},
		{"X-Forwarded-For": "::ffff:203.0.113.7"},
		{"X-Forwarded-For": "garbage"},
		{"X-Forwarded-For": "127.0.0.1"},
	}

	for i, h := range headers {
		_, err := oiNew0506RateLimitAttempt(app, api, peer, h)
		if i < 2 {
			if err != nil {
				t.Fatalf("attempt %d unexpectedly rejected before configured limit: %v", i+1, err)
			}
			continue
		}
		if !errors.Is(err, errAuthRateLimited) {
			t.Fatalf("header rotation created fresh limiter budget on attempt %d: %v", i+1, err)
		}
	}
}

func TestSecurityBreakerAuthRateLimiterWindowBoundaryDoesNotDoubleSpend(t *testing.T) {
	requireSecurityBreaker(t)

	limiter := newBoundedAuthRateLimiter(2, 2, 16, time.Minute)
	key := "/api/v1/auth/login|198.51.100.99"
	start := time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC)

	if !limiter.allow(key, start) || !limiter.allow(key, start.Add(1*time.Nanosecond)) {
		t.Fatal("expected first two requests to be admitted")
	}
	if limiter.allow(key, start.Add(time.Minute-1*time.Nanosecond)) {
		t.Fatal("request before expiry boundary bypassed limiter")
	}
	if !limiter.allow(key, start.Add(time.Minute+1*time.Nanosecond)) {
		t.Fatal("expired request budget was not reclaimed after the window")
	}
}


func TestSecurityBreakerAuthRateLimitAmplificationDoesNotScaleWithInstanceCount(t *testing.T) {
	requireSecurityBreaker(t)

	const (
		perClientLimit = 2
		instances      = 8
		attempts       = 64
	)
	type instance struct {
		api *API
		app *fiber.App
	}
	apps := make([]instance, 0, instances)
	for i := 0; i < instances; i++ {
		api, app := newOINew0506RateLimitApp(perClientLimit, HTTPNetworkConfig{})
		apps = append(apps, instance{api: api, app: app})
	}

	peer := "198.51.100.201"
	admitted := 0
	for i := 0; i < attempts; i++ {
		target := apps[i%len(apps)]
		_, err := oiNew0506RateLimitAttempt(target.app, target.api, peer, nil)
		switch {
		case err == nil:
			admitted++
		case errors.Is(err, errAuthRateLimited):
		default:
			t.Fatalf("unexpected limiter result on attempt %d: %v", i+1, err)
		}
	}
	if admitted > perClientLimit {
		t.Fatalf("distributed limiter budget amplified with instance count: admitted=%d secure_budget=%d instances=%d", admitted, perClientLimit, instances)
	}
}

func TestSecurityBreakerAuthRateLimitConcurrentCrossInstanceBurstCannotMultiplyBudget(t *testing.T) {
	requireSecurityBreaker(t)

	const (
		perClientLimit = 2
		instances      = 4
		contenders     = 64
	)
	type instance struct {
		api *API
		app *fiber.App
	}
	apps := make([]instance, 0, instances)
	for i := 0; i < instances; i++ {
		api, app := newOINew0506RateLimitApp(perClientLimit, HTTPNetworkConfig{})
		apps = append(apps, instance{api: api, app: app})
	}

	peer := "198.51.100.202"
	start := make(chan struct{})
	results := make(chan error, contenders)
	for i := 0; i < contenders; i++ {
		i := i
		go func() {
			<-start
			target := apps[i%len(apps)]
			_, err := oiNew0506RateLimitAttempt(target.app, target.api, peer, nil)
			results <- err
		}()
	}
	close(start)

	admitted := 0
	for i := 0; i < contenders; i++ {
		err := <-results
		switch {
		case err == nil:
			admitted++
		case errors.Is(err, errAuthRateLimited):
		default:
			t.Fatalf("unexpected concurrent limiter result: %v", err)
		}
	}
	if admitted > perClientLimit {
		t.Fatalf("concurrent distributed limiter bypass reproduced: admitted=%d secure_budget=%d instances=%d", admitted, perClientLimit, instances)
	}
}
