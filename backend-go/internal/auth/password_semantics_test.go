package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestStage335RegistrationPasswordCodePointPolicy(t *testing.T) {
	tests := []struct {
		name     string
		password string
		valid    bool
	}{
		{name: "11 ascii", password: strings.Repeat("a", 11), valid: false},
		{name: "12 ascii", password: strings.Repeat("a", 12), valid: true},
		{name: "11 cyrillic", password: strings.Repeat("я", 11), valid: false},
		{name: "12 cyrillic", password: strings.Repeat("я", 12), valid: true},
		{name: "256 supplementary", password: strings.Repeat("😀", 256), valid: true},
		{name: "257 supplementary", password: strings.Repeat("😀", 257), valid: false},
		{name: "malformed utf8", password: string([]byte{0xff, 'a', 'b'}), valid: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateRegistration(stage335Registration(test.password))
			if test.valid && err != nil {
				t.Fatalf("expected password to be valid: %v", err)
			}
			if !test.valid && !IsInvalidInput(err) {
				t.Fatalf("expected invalid input, got %v", err)
			}
		})
	}
}

func TestStage335LoginRejectsMalformedOrOversizedPasswordBeforeStoreAndArgon2(t *testing.T) {
	for _, password := range []string{string([]byte{0xff}), strings.Repeat("😀", 257)} {
		store := &memoryStore{}
		service := newTestService(t, store)
		_, err := service.Login(context.Background(), LoginRequest{Email: "investor@example.com", Password: password})
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("expected generic invalid credentials, got %v", err)
		}
		if store.findCalls != 0 {
			t.Fatalf("expected rejection before store lookup/Argon2 path, got %d lookups", store.findCalls)
		}
	}
}

func TestStage335LegacyMultibytePasswordStillAuthenticates(t *testing.T) {
	password := "абвгде"
	encoded, err := hashPassword(password)
	if err != nil {
		t.Fatalf("hash legacy password: %v", err)
	}
	store := &memoryStore{
		user: StoredUser{
			ID:                  "10000000-0000-4000-8000-000000000001",
			InvestmentSubjectID: "10000000-0000-4000-8000-000000000002",
			Email:               "investor@example.com",
			Language:            LanguageEN,
			Theme:               ThemeSystem,
			Timezone:            "UTC",
			PrivacyMode:         true,
			CreatedAt:           fixedClock{}.Now(),
		},
		password: encoded,
	}
	service := newTestService(t, store)
	if _, err := service.Login(context.Background(), LoginRequest{Email: "investor@example.com", Password: password}); err != nil {
		t.Fatalf("legacy multibyte login: %v", err)
	}
	if _, err := service.Register(context.Background(), stage335Registration(password)); !IsInvalidInput(err) {
		t.Fatalf("expected corrected registration minimum to reject legacy 6-code-point secret, got %v", err)
	}
}

func TestStage335WhitespaceOnlyPasswordRoundTripsExactly(t *testing.T) {
	store := &memoryStore{}
	service := newTestService(t, store)
	password := strings.Repeat(" ", 12)
	if _, err := service.Register(context.Background(), stage335Registration(password)); err != nil {
		t.Fatalf("register whitespace-only exact secret: %v", err)
	}
	if _, err := service.Login(context.Background(), LoginRequest{Email: "investor@example.com", Password: password}); err != nil {
		t.Fatalf("login whitespace-only exact secret: %v", err)
	}
}

func TestStage335PasswordIdentityDoesNotNormalizeUnicode(t *testing.T) {
	store := &memoryStore{}
	service := newTestService(t, store)
	composed := strings.Repeat("a", 11) + "é"
	decomposed := strings.Repeat("a", 11) + "e\u0301"
	if _, err := service.Register(context.Background(), stage335Registration(composed)); err != nil {
		t.Fatalf("register composed secret: %v", err)
	}
	if _, err := service.Login(context.Background(), LoginRequest{Email: "investor@example.com", Password: decomposed}); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected byte-distinct canonically equivalent secret to fail, got %v", err)
	}
	if _, err := service.Login(context.Background(), LoginRequest{Email: "investor@example.com", Password: composed}); err != nil {
		t.Fatalf("expected exact original secret to authenticate: %v", err)
	}
}

func stage335Registration(password string) RegistrationRequest {
	return RegistrationRequest{
		Email:    "investor@example.com",
		Password: password,
		Language: LanguageEN,
		Theme:    ThemeSystem,
		Timezone: "UTC",
	}
}
