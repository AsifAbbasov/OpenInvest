package auth

import (
	"context"
	"errors"
	"testing"
)

func TestRegisterFailsFastWhenArgonCapacityIsExhausted(t *testing.T) {
	restore := exhaustProcessArgonCapacity(t)
	defer restore()

	store := &memoryStore{}
	service := newTestService(t, store)
	_, err := service.Register(context.Background(), RegistrationRequest{
		Email:    "capacity@example.com",
		Password: "correct horse battery staple",
		Language: LanguageEN,
		Theme:    ThemeSystem,
		Timezone: "UTC",
	})
	if !errors.Is(err, ErrAuthCapacity) {
		t.Fatalf("expected ErrAuthCapacity, got %v", err)
	}
	if store.registered.UserID != "" {
		t.Fatal("registration must not reach persistence while Argon2 capacity is exhausted")
	}
}

func TestUnknownLoginFailsFastWhenDummyArgonCapacityIsExhausted(t *testing.T) {
	restore := exhaustProcessArgonCapacity(t)
	defer restore()

	service := newTestService(t, &memoryStore{})
	_, err := service.Login(context.Background(), LoginRequest{
		Email:    "missing@example.com",
		Password: "correct horse battery staple",
	})
	if !errors.Is(err, ErrAuthCapacity) {
		t.Fatalf("expected ErrAuthCapacity from dummy verification path, got %v", err)
	}
}

func exhaustProcessArgonCapacity(t *testing.T) func() {
	t.Helper()
	previous := processArgonWorkLimiter
	limiter := newArgonWorkLimiter(1)
	limiter.slots <- struct{}{}
	processArgonWorkLimiter = limiter
	return func() {
		<-limiter.slots
		processArgonWorkLimiter = previous
	}
}
