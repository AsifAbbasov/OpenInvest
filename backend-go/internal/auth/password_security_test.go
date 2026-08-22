package auth

import (
	"encoding/base64"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestArgonWorkLimiterFailsFastAtConfiguredCapacity(t *testing.T) {
	const limit = 2
	limiter := newArgonWorkLimiter(limit)

	started := make(chan struct{}, limit)
	release := make(chan struct{})
	done := make(chan error, limit)
	var ready sync.WaitGroup
	ready.Add(limit)

	for i := 0; i < limit; i++ {
		go func() {
			ready.Done()
			done <- limiter.run(func() {
				started <- struct{}{}
				<-release
			})
		}()
	}
	ready.Wait()
	for i := 0; i < limit; i++ {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("expected Argon2 work slot to be occupied")
		}
	}

	begin := time.Now()
	err := limiter.run(func() { t.Fatal("capacity-exhausted work must not execute") })
	if !errors.Is(err, ErrAuthCapacity) {
		t.Fatalf("expected ErrAuthCapacity, got %v", err)
	}
	if elapsed := time.Since(begin); elapsed > 100*time.Millisecond {
		t.Fatalf("capacity rejection blocked for %s", elapsed)
	}

	close(release)
	for i := 0; i < limit; i++ {
		select {
		case err := <-done:
			if err != nil {
				t.Fatalf("occupied Argon2 work failed: %v", err)
			}
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for Argon2 work to finish")
		}
	}

	if err := limiter.run(func() {}); err != nil {
		t.Fatalf("expected capacity to recover after slots are released: %v", err)
	}
}

func TestVerifyPasswordRejectsArgonParametersOutsideApprovedBudget(t *testing.T) {
	salt := base64.RawStdEncoding.EncodeToString(make([]byte, argonSaltLen))
	hash := base64.RawStdEncoding.EncodeToString(make([]byte, argonKeyLen))
	encoded := "argon2id$v=19$m=131072,t=3,p=1$" + salt + "$" + hash

	verified, err := verifyPassword("password", encoded)
	if err != nil {
		t.Fatalf("over-budget encoded hash must fail validation before Argon2 work: %v", err)
	}
	if verified {
		t.Fatal("expected over-budget Argon2 parameters to be rejected")
	}
}
