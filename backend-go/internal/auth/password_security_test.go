package auth

import (
	"encoding/base64"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestArgonWorkLimiterCapsConcurrentWork(t *testing.T) {
	const limit = 2
	limiter := newArgonWorkLimiter(limit)

	begin := make(chan struct{})
	started := make(chan struct{}, limit+1)
	release := make(chan struct{}, limit+1)
	done := make(chan struct{}, limit+1)

	var ready sync.WaitGroup
	ready.Add(limit + 1)
	var active atomic.Int32
	var maximum atomic.Int32

	worker := func() {
		ready.Done()
		<-begin
		limiter.run(func() {
			current := active.Add(1)
			for {
				observed := maximum.Load()
				if current <= observed || maximum.CompareAndSwap(observed, current) {
					break
				}
			}
			started <- struct{}{}
			<-release
			active.Add(-1)
		})
		done <- struct{}{}
	}

	for i := 0; i < limit+1; i++ {
		go worker()
	}
	ready.Wait()
	close(begin)

	for i := 0; i < limit; i++ {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("expected bounded Argon2 work to start")
		}
	}

	select {
	case <-started:
		t.Fatal("more Argon2 work started than the configured process-wide limit")
	case <-time.After(100 * time.Millisecond):
	}

	release <- struct{}{}

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("expected queued Argon2 work to start after a slot was released")
	}

	for i := 0; i < limit; i++ {
		release <- struct{}{}
	}
	for i := 0; i < limit+1; i++ {
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for bounded Argon2 work")
		}
	}

	if got := maximum.Load(); got > limit {
		t.Fatalf("maximum concurrent Argon2 work = %d, want <= %d", got, limit)
	}
}

func TestVerifyPasswordRejectsArgonParametersOutsideApprovedBudget(t *testing.T) {
	salt := base64.RawStdEncoding.EncodeToString(make([]byte, argonSaltLen))
	hash := base64.RawStdEncoding.EncodeToString(make([]byte, argonKeyLen))
	encoded := "argon2id$v=19$m=131072,t=3,p=1$" + salt + "$" + hash

	if verifyPassword("password", encoded) {
		t.Fatal("expected over-budget Argon2 parameters to be rejected")
	}
}
