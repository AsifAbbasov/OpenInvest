package httpapi

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
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

func TestArchitectureHardeningExpensiveReadAdmissionConcurrentBound(t *testing.T) {
	const global = 8
	const perSubject = 2
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
