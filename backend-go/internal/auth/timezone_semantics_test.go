package auth

import (
	"context"
	"strings"
	"testing"
)

func TestStage337RegistrationTimezoneResolverContract(t *testing.T) {
	tests := []struct {
		name     string
		timezone string
		valid    bool
	}{
		{name: "utc compatibility", timezone: "UTC", valid: true},
		{name: "baku", timezone: "Asia/Baku", valid: true},
		{name: "berlin", timezone: "Europe/Berlin", valid: true},
		{name: "new york", timezone: "America/New_York", valid: true},
		{name: "loadable fixed-offset tzdb identifier", timezone: "Etc/GMT+4", valid: true},
		{name: "empty", timezone: "", valid: false},
		{name: "whitespace only", timezone: " ", valid: false},
		{name: "host local pseudo-zone", timezone: "Local", valid: false},
		{name: "unknown", timezone: "Not/AZone", valid: false},
		{name: "invented", timezone: "Mars/Olympus", valid: false},
		{name: "raw offset", timezone: "+04:00", valid: false},
		{name: "raw utc offset", timezone: "UTC+04:00", valid: false},
		{name: "leading whitespace", timezone: " Asia/Baku", valid: false},
		{name: "trailing whitespace", timezone: "Asia/Baku ", valid: false},
		// Case sensitivity is delegated to the active LoadLocation source; no case normalization is performed here.
		{name: "parent traversal", timezone: "../UTC", valid: false},
		{name: "absolute path", timezone: "/etc/localtime", valid: false},
		{name: "over existing byte bound", timezone: strings.Repeat("A", 65), valid: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateRegistration(stage337Registration(test.timezone))
			if test.valid && err != nil {
				t.Fatalf("expected timezone %q to be valid: %v", test.timezone, err)
			}
			if !test.valid && !IsInvalidInput(err) {
				t.Fatalf("expected timezone %q to be invalid input, got %v", test.timezone, err)
			}
		})
	}
}

func TestStage337TimezonePreResolverSyntaxGuards(t *testing.T) {
	tests := []struct {
		name     string
		timezone string
		invalid  bool
	}{
		{name: "empty", timezone: "", invalid: true},
		{name: "whitespace only", timezone: " ", invalid: true},
		{name: "leading whitespace", timezone: " Asia/Baku", invalid: true},
		{name: "trailing whitespace", timezone: "Asia/Baku ", invalid: true},
		{name: "positive raw offset", timezone: "+04:00", invalid: true},
		{name: "negative raw offset", timezone: "-04:00", invalid: true},
		{name: "positive UTC raw offset", timezone: "UTC+04:00", invalid: true},
		{name: "negative UTC raw offset", timezone: "UTC-04:00", invalid: true},
		{name: "normal IANA identifier", timezone: "Asia/Baku", invalid: false},
		{name: "UTC compatibility", timezone: "UTC", invalid: false},
		{name: "loadable fixed-offset tzdb identifier", timezone: "Etc/GMT+4", invalid: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := invalidRegistrationTimezoneSyntax(test.timezone); got != test.invalid {
				t.Fatalf("invalidRegistrationTimezoneSyntax(%q): expected %t, got %t", test.timezone, test.invalid, got)
			}
		})
	}
}

func TestStage337RegistrationPersistsAcceptedTimezoneExactly(t *testing.T) {
	store := &stage337Store{}
	service := newStage337Service(t, store)
	timezone := "Etc/GMT+4"

	if _, err := service.Register(context.Background(), stage337Registration(timezone)); err != nil {
		t.Fatalf("register: %v", err)
	}
	if store.registerCalls != 1 {
		t.Fatalf("expected one store registration, got %d", store.registerCalls)
	}
	if store.registered.Timezone != timezone {
		t.Fatalf("expected exact timezone %q, got %q", timezone, store.registered.Timezone)
	}
}

func TestStage337InvalidTimezoneNeverReachesRegistrationStore(t *testing.T) {
	store := &stage337Store{}
	service := newStage337Service(t, store)

	if _, err := service.Register(context.Background(), stage337Registration("Not/AZone")); !IsInvalidInput(err) {
		t.Fatalf("expected invalid timezone to fail validation, got %v", err)
	}
	if store.registerCalls != 0 {
		t.Fatalf("invalid timezone must not reach RegisterUser, got %d calls", store.registerCalls)
	}
}

func stage337Registration(timezone string) RegistrationRequest {
	return RegistrationRequest{
		Email:    "investor@example.com",
		Password: "correct horse battery staple",
		Language: LanguageEN,
		Theme:    ThemeSystem,
		Timezone: timezone,
	}
}

type stage337Store struct {
	memoryStore
	registerCalls int
}

func (store *stage337Store) RegisterUser(ctx context.Context, record RegistrationRecord) (StoredUser, error) {
	store.registerCalls++
	return store.memoryStore.RegisterUser(ctx, record)
}

func newStage337Service(t *testing.T, store Store) *Service {
	t.Helper()
	service, err := NewService(store, fixedClock{}, Config{
		AccessTokenSecret:   []byte("01234567890123456789012345678901"),
		RefreshCookieSecure: true,
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	return service
}
