package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
)

func TestArchitectureHardeningExpensiveReadAdmissionIsFailFastAndBounded(t *testing.T) {
	admission := newExpensiveReadAdmission(2, 3)
	firstA, err := admission.acquire("subject-a")
	if err != nil {
		t.Fatalf("first subject-a acquire: %v", err)
	}
	defer firstA()
	secondA, err := admission.acquire("subject-a")
	if err != nil {
		t.Fatalf("second subject-a acquire: %v", err)
	}
	defer secondA()

	if _, err := admission.acquire("subject-a"); !errors.Is(err, errExpensiveReadSubjectLimited) {
		t.Fatalf("third subject-a acquire error=%v, want subject limit", err)
	}

	firstB, err := admission.acquire("subject-b")
	if err != nil {
		t.Fatalf("first subject-b acquire: %v", err)
	}
	defer firstB()
	if _, err := admission.acquire("subject-c"); !errors.Is(err, errExpensiveReadCapacityExhausted) {
		t.Fatalf("global capacity acquire error=%v, want global exhaustion", err)
	}

	firstA()
	replacement, err := admission.acquire("subject-c")
	if err != nil {
		t.Fatalf("capacity did not recover after release: %v", err)
	}
	replacement()

	admission.mu.Lock()
	defer admission.mu.Unlock()
	if _, exists := admission.perSubjectActive["subject-c"]; exists {
		t.Fatal("released subject entry remained in active-state map")
	}
	if len(admission.perSubjectActive) > cap(admission.capacity) {
		t.Fatalf("active-state map grew beyond global capacity: %d > %d", len(admission.perSubjectActive), cap(admission.capacity))
	}
}


func TestArchitectureHardeningNormalPortfolioUIFanoutFitsAdmission(t *testing.T) {
	const normalFrontendMaxExpensiveReadFanout = 5
	if defaultExpensiveReadPerSubjectCapacity < normalFrontendMaxExpensiveReadFanout {
		t.Fatalf("per-subject capacity=%d is below verified frontend fanout=%d", defaultExpensiveReadPerSubjectCapacity, normalFrontendMaxExpensiveReadFanout)
	}
	if defaultExpensiveReadPerSubjectCapacity > defaultExpensiveReadGlobalCapacity {
		t.Fatalf("per-subject capacity=%d exceeds global capacity=%d", defaultExpensiveReadPerSubjectCapacity, defaultExpensiveReadGlobalCapacity)
	}

	admission := newDefaultExpensiveReadAdmission()
	operations := []string{
		"summary",
		"current-positions",
		"cash-flow",
		"historical-positions",
		"returns",
	}
	releases := make([]func(), 0, len(operations))
	for _, operation := range operations {
		release, err := admission.acquire("normal-portfolio-ui-subject")
		if err != nil {
			t.Fatalf("normal UI operation %s was rejected at verified fanout: %v", operation, err)
		}
		releases = append(releases, release)
	}
	if _, err := admission.acquire("normal-portfolio-ui-subject"); !errors.Is(err, errExpensiveReadSubjectLimited) {
		t.Fatalf("sixth same-subject expensive read error=%v, want subject bound", err)
	}

	otherReleases := make([]func(), 0, defaultExpensiveReadGlobalCapacity-normalFrontendMaxExpensiveReadFanout)
	for i := 0; i < cap(admission.capacity)-normalFrontendMaxExpensiveReadFanout; i++ {
		release, err := admission.acquire(fmt.Sprintf("other-subject-%d", i))
		if err != nil {
			t.Fatalf("global spare slot %d rejected: %v", i, err)
		}
		otherReleases = append(otherReleases, release)
	}
	if _, err := admission.acquire("another-subject"); !errors.Is(err, errExpensiveReadCapacityExhausted) {
		t.Fatalf("ninth process-wide expensive read error=%v, want global capacity exhaustion", err)
	}

	for _, release := range otherReleases {
		release()
	}
	for _, release := range releases {
		release()
	}
}

func TestArchitectureHardeningExpensiveReadAdmissionConcurrentBound(t *testing.T) {
	const global = 8
	const perSubject = 5
	admission := newExpensiveReadAdmission(perSubject, global)

	var active atomic.Int64
	var maxActive atomic.Int64
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 128; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			<-start
			subject := fmt.Sprintf("subject-%d", index%16)
			release, err := admission.acquire(subject)
			if err != nil {
				if !errors.Is(err, errExpensiveReadSubjectLimited) && !errors.Is(err, errExpensiveReadCapacityExhausted) {
					t.Errorf("unexpected admission error: %v", err)
				}
				return
			}
			current := active.Add(1)
			for {
				previous := maxActive.Load()
				if current <= previous || maxActive.CompareAndSwap(previous, current) {
					break
				}
			}
			time.Sleep(2 * time.Millisecond)
			active.Add(-1)
			release()
		}(i)
	}
	close(start)
	wg.Wait()

	if got := maxActive.Load(); got > global {
		t.Fatalf("observed %d active reads, want <= %d", got, global)
	}
	if got := active.Load(); got != 0 {
		t.Fatalf("active reads after completion=%d", got)
	}
	admission.mu.Lock()
	defer admission.mu.Unlock()
	if len(admission.perSubjectActive) != 0 {
		t.Fatalf("subject active map leaked %d entries", len(admission.perSubjectActive))
	}
}

func TestArchitectureHardeningExpensiveReadHTTPErrorContract(t *testing.T) {
	app := fiber.New()
	app.Get("/subject", func(c fiber.Ctx) error {
		return writeMappedError(c, errExpensiveReadSubjectLimited)
	})
	app.Get("/capacity", func(c fiber.Ctx) error {
		return writeMappedError(c, errExpensiveReadCapacityExhausted)
	})

	tests := []struct {
		path       string
		wantStatus int
		wantCode   string
		wantRetry  string
	}{
		{path: "/subject", wantStatus: http.StatusTooManyRequests, wantCode: "RATE_LIMITED", wantRetry: expensiveReadSubjectRetryAfterSeconds},
		{path: "/capacity", wantStatus: http.StatusServiceUnavailable, wantCode: "READ_CAPACITY_EXHAUSTED", wantRetry: expensiveReadCapacityRetryAfterSeconds},
	}
	for _, test := range tests {
		response, err := app.Test(httptest.NewRequest(http.MethodGet, test.path, nil))
		if err != nil {
			t.Fatalf("%s request: %v", test.path, err)
		}
		if response.StatusCode != test.wantStatus {
			_ = response.Body.Close()
			t.Fatalf("%s status=%d want=%d", test.path, response.StatusCode, test.wantStatus)
		}
		if got := response.Header.Get("Retry-After"); got != test.wantRetry {
			_ = response.Body.Close()
			t.Fatalf("%s Retry-After=%q want=%q", test.path, got, test.wantRetry)
		}
		var payload struct {
			Error struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
			_ = response.Body.Close()
			t.Fatalf("%s decode: %v", test.path, err)
		}
		_ = response.Body.Close()
		if payload.Error.Code != test.wantCode {
			t.Fatalf("%s code=%q want=%q", test.path, payload.Error.Code, test.wantCode)
		}
	}
}

func TestArchitectureHardeningAuthSharedNATCouplingReduced(t *testing.T) {
	now := time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC)
	api := &API{
		authLoginIPLimiter:    newBoundedAuthRateLimiter(defaultAuthLoginIPLimit, 10000, 2048, time.Minute),
		authCredentialLimiter: newBoundedAuthRateLimiter(defaultAuthCredentialLimit, 10000, 4096, time.Minute),
		now:                   func() time.Time { return now },
	}
	ipKey := "/api/v1/auth/login|203.0.113.44"
	for i := 0; i < 50; i++ {
		if !api.authLoginIPLimiter.allow(ipKey, now) {
			t.Fatalf("shared NAT user %d was coupled before emergency ceiling", i)
		}
		email := fmt.Sprintf("user-%d@example.com", i)
		if err := api.checkAuthCredentialRateLimit("/api/v1/auth/login", email); err != nil {
			t.Fatalf("independent credential %d unexpectedly limited: %v", i, err)
		}
	}
}

func TestArchitectureHardeningAuthCredentialBucketSurvivesIPRotationAndCaseVariants(t *testing.T) {
	now := time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC)
	api := &API{
		authLoginIPLimiter:    newBoundedAuthRateLimiter(defaultAuthLoginIPLimit, 10000, 2048, time.Minute),
		authCredentialLimiter: newBoundedAuthRateLimiter(defaultAuthCredentialLimit, 10000, 4096, time.Minute),
		now:                   func() time.Time { return now },
	}
	for i := 0; i < defaultAuthCredentialLimit; i++ {
		ipKey := fmt.Sprintf("/api/v1/auth/login|2001:db8::%x", i+1)
		if !api.authLoginIPLimiter.allow(ipKey, now) {
			t.Fatalf("rotated IP %d unexpectedly hit IP ceiling", i)
		}
		email := "Investor@Example.com"
		if i%2 == 1 {
			email = " investor@example.COM "
		}
		if err := api.checkAuthCredentialRateLimit("/api/v1/auth/login", email); err != nil {
			t.Fatalf("credential attempt %d unexpectedly rejected: %v", i+1, err)
		}
	}
	if err := api.checkAuthCredentialRateLimit("/api/v1/auth/login", "INVESTOR@example.com"); !errors.Is(err, errAuthRateLimited) {
		t.Fatalf("case/space variant escaped normalized credential bucket: %v", err)
	}
}

func TestArchitectureHardeningRotatingCredentialsStillHitIPEmergencyCeiling(t *testing.T) {
	now := time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC)
	ipLimiter := newBoundedAuthRateLimiter(defaultAuthLoginIPLimit, 10000, 2048, time.Minute)
	ipKey := "/api/v1/auth/login|198.51.100.8"
	for i := 0; i < defaultAuthLoginIPLimit; i++ {
		if !ipLimiter.allow(ipKey, now) {
			t.Fatalf("attempt %d rejected before emergency ceiling", i+1)
		}
	}
	if ipLimiter.allow(ipKey, now) {
		t.Fatal("rotating fake credentials could bypass the IP emergency ceiling")
	}
}
