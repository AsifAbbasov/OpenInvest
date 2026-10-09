package httpapi

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestAuthRateLimiterRemainsMemoryBoundedUnderHighCardinalityPressure(t *testing.T) {
	const (
		perKeyLimit = 5
		globalLimit = 500
		maxKeys     = 128
	)
	window := time.Minute
	limiter := newBoundedAuthRateLimiter(perKeyLimit, globalLimit, maxKeys, window)
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	allowed := 0
	for i := 0; i < 50000; i++ {
		if limiter.allow(fmt.Sprintf("198.51.100.%d", i), now) {
			allowed++
		}
		if len(limiter.attempts) > maxKeys {
			t.Fatalf("rate limiter key cardinality escaped bound: got=%d max=%d", len(limiter.attempts), maxKeys)
		}
		if len(limiter.globalAttempts) > globalLimit {
			t.Fatalf("global attempt storage escaped bound: got=%d max=%d", len(limiter.globalAttempts), globalLimit)
		}
	}
	if allowed > globalLimit {
		t.Fatalf("global admission bound bypassed: allowed=%d limit=%d", allowed, globalLimit)
	}
	if len(limiter.attempts) > maxKeys {
		t.Fatalf("high-cardinality pressure left unbounded key map: %d", len(limiter.attempts))
	}

	later := now.Add(window + time.Second)
	if !limiter.allow("203.0.113.10", later) {
		t.Fatal("expired high-cardinality pressure permanently exhausted limiter capacity")
	}
	if len(limiter.attempts) > maxKeys {
		t.Fatalf("post-expiry sweep exceeded key bound: %d", len(limiter.attempts))
	}
	if len(limiter.globalAttempts) > globalLimit {
		t.Fatalf("post-expiry global history exceeded bound: %d", len(limiter.globalAttempts))
	}
}

func TestResourceExhaustionGenericJSONBodyHasHardCeiling(t *testing.T) {
	app := newApp(&API{
		service:                 verticalslice.NewService(&importAPITestStore{}, fixedHTTPClock{}),
		allowDevelopmentSubject: true,
	})

	// Fiber's configured/default admission ceiling must reject a multi-megabyte
	// generic JSON body before application JSON decoding allocates/processes it.
	body := bytes.Repeat([]byte("x"), 5*1024*1024)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/portfolios", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "resource-bound-key-000001")

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("oversized request transport failed unexpectedly: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("generic oversized JSON body was not rejected at transport admission: status=%d", response.StatusCode)
	}
}
