package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestServiceRegisterCreatesPrivacyFirstSession(t *testing.T) {
	store := &memoryStore{}
	service := newTestService(t, store)

	result, err := service.Register(context.Background(), RegistrationRequest{
		Email:    " Investor@Example.COM ",
		Password: "correct horse battery staple",
		Language: LanguageEN,
		Theme:    ThemeSystem,
		Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if result.User.Email != "investor@example.com" {
		t.Fatalf("expected normalized email, got %q", result.User.Email)
	}
	if !result.User.PrivacyMode {
		t.Fatalf("expected privacy mode enabled by default")
	}
	if result.RefreshToken == "" || result.Session.AccessToken == "" || result.Session.CSRFToken == "" {
		t.Fatalf("expected access, refresh and csrf tokens")
	}
	if store.registered.Session.RefreshTokenHash == "" || store.registered.Session.CSRFTokenHash == "" {
		t.Fatalf("expected only hashed refresh/csrf tokens to reach the store")
	}
	if store.registered.Session.RefreshTokenHash == result.RefreshToken || store.registered.Session.CSRFTokenHash == result.Session.CSRFToken {
		t.Fatalf("raw refresh/csrf token leaked into store record")
	}
}

func TestServiceRefreshRequiresCSRFBeforeRotation(t *testing.T) {
	store := &memoryStore{}
	service := newTestService(t, store)
	result, err := service.Register(context.Background(), RegistrationRequest{
		Email:    "investor@example.com",
		Password: "correct horse battery staple",
		Language: LanguageEN,
		Theme:    ThemeSystem,
		Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	_, err = service.Refresh(context.Background(), result.RefreshToken, "wrong-csrf-token")
	if !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("expected invalid session for csrf mismatch, got %v", err)
	}
	if store.rotateCalls != 1 {
		t.Fatalf("expected one attempted rotation, got %d", store.rotateCalls)
	}
	if store.revoked {
		t.Fatalf("csrf mismatch must not revoke or rotate the current session")
	}

	rotated, err := service.Refresh(context.Background(), result.RefreshToken, result.Session.CSRFToken)
	if err != nil {
		t.Fatalf("refresh with correct csrf: %v", err)
	}
	if rotated.RefreshToken == result.RefreshToken || rotated.Session.CSRFToken == result.Session.CSRFToken {
		t.Fatalf("expected refresh and csrf token rotation")
	}
}

func TestServiceAuditsEarlyRefreshAndLogoutRejections(t *testing.T) {
	store := &memoryStore{}
	service := newTestService(t, store)

	if _, err := service.Refresh(context.Background(), "", "csrf-token"); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("expected missing refresh token to be invalid session, got %v", err)
	}
	if _, err := service.Logout(context.Background(), "refresh-token", "", false); !errors.Is(err, ErrInvalidCSRF) {
		t.Fatalf("expected missing csrf token to be invalid csrf, got %v", err)
	}

	if len(store.auditEvents) != 2 {
		t.Fatalf("expected two rejected auth audit events, got %d", len(store.auditEvents))
	}
	if store.auditEvents[0].ActionCode != "AUTH_REFRESH_REJECTED" || store.auditEvents[0].Outcome != "failure" {
		t.Fatalf("unexpected refresh audit event: %#v", store.auditEvents[0])
	}
	if store.auditEvents[1].ActionCode != "AUTH_LOGOUT_REJECTED" || store.auditEvents[1].Outcome != "failure" {
		t.Fatalf("unexpected logout audit event: %#v", store.auditEvents[1])
	}
}

func TestServiceRejectsRefreshReplayAfterRotation(t *testing.T) {
	service := newTestService(t, &memoryStore{})
	result, err := service.Register(context.Background(), RegistrationRequest{
		Email:    "investor@example.com",
		Password: "correct horse battery staple",
		Language: LanguageEN,
		Theme:    ThemeSystem,
		Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if _, err := service.Refresh(context.Background(), result.RefreshToken, result.Session.CSRFToken); err != nil {
		t.Fatalf("first refresh: %v", err)
	}
	if _, err := service.Refresh(context.Background(), result.RefreshToken, result.Session.CSRFToken); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("expected refresh replay to be rejected, got %v", err)
	}
}

func TestServiceLogoutRevokesSession(t *testing.T) {
	service := newTestService(t, &memoryStore{})
	result, err := service.Register(context.Background(), RegistrationRequest{
		Email:    "investor@example.com",
		Password: "correct horse battery staple",
		Language: LanguageEN,
		Theme:    ThemeSystem,
		Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	revoked, err := service.Logout(context.Background(), result.RefreshToken, result.Session.CSRFToken, false)
	if err != nil {
		t.Fatalf("logout: %v", err)
	}
	if !revoked {
		t.Fatalf("expected logout to revoke the active session")
	}
	if _, err := service.Refresh(context.Background(), result.RefreshToken, result.Session.CSRFToken); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("expected logged-out refresh token to be rejected, got %v", err)
	}
}

func TestServiceRequiresExplicitAccessTokenSecretOrEphemeralFlag(t *testing.T) {
	_, err := NewService(&memoryStore{}, fixedClock{}, Config{})
	if !IsInvalidInput(err) {
		t.Fatalf("expected invalid input for missing access token secret, got %v", err)
	}
	if _, err := NewService(&memoryStore{}, fixedClock{}, Config{AllowEphemeralAccessTokenSecret: true}); err != nil {
		t.Fatalf("expected explicit ephemeral secret flag to be accepted: %v", err)
	}
}

func TestServiceRejectsDisplayNameEmailForms(t *testing.T) {
	service := newTestService(t, &memoryStore{})

	if _, err := service.Register(context.Background(), RegistrationRequest{
		Email:    `Investor <investor@example.com>`,
		Password: "correct horse battery staple",
		Language: LanguageEN,
		Theme:    ThemeSystem,
		Timezone: "UTC",
	}); !IsInvalidInput(err) {
		t.Fatalf("expected display-name email form to be rejected, got %v", err)
	}
}

func TestServiceRejectsOverlongEmail(t *testing.T) {
	service := newTestService(t, &memoryStore{})
	overlong := strings.Repeat("a", 245) + "@example.com"

	if _, err := service.Register(context.Background(), RegistrationRequest{
		Email:    overlong,
		Password: "correct horse battery staple",
		Language: LanguageEN,
		Theme:    ThemeSystem,
		Timezone: "UTC",
	}); !IsInvalidInput(err) {
		t.Fatalf("expected overlong email to be rejected, got %v", err)
	}
}

func newTestService(t *testing.T, store *memoryStore) *Service {
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

type fixedClock struct{}

func (fixedClock) Now() time.Time {
	return time.Date(2026, 7, 9, 12, 0, 0, 0, time.UTC)
}

type memoryStore struct {
	registered  RegistrationRecord
	user        StoredUser
	password    string
	sessions    map[string]SessionRecord
	auditEvents []AuthAuditRecord
	revoked     bool
	rotateCalls int
}

func (store *memoryStore) RegisterUser(_ context.Context, record RegistrationRecord) (StoredUser, error) {
	store.registered = record
	store.password = record.PasswordHash
	store.user = StoredUser{
		ID:                  record.UserID,
		InvestmentSubjectID: record.InvestmentSubjectID,
		Email:               record.EmailNormalized,
		Language:            record.Language,
		Theme:               record.Theme,
		Timezone:            record.Timezone,
		PrivacyMode:         true,
		CreatedAt:           record.Now,
	}
	store.sessions = map[string]SessionRecord{record.Session.RefreshTokenHash: record.Session}
	return store.user, nil
}

func (store *memoryStore) FindUserByEmail(_ context.Context, email string) (StoredUser, string, error) {
	if email != store.user.Email {
		return StoredUser{}, "", ErrInvalidCredentials
	}
	return store.user, store.password, nil
}

func (store *memoryStore) CreateSession(_ context.Context, record SessionRecord) error {
	if store.sessions == nil {
		store.sessions = map[string]SessionRecord{}
	}
	store.sessions[record.RefreshTokenHash] = record
	return nil
}

func (store *memoryStore) RotateSession(_ context.Context, currentRefreshTokenHash string, currentCSRFTokenHash string, next SessionRecord, _ time.Time) (StoredUser, error) {
	store.rotateCalls++
	current, ok := store.sessions[currentRefreshTokenHash]
	if !ok {
		return StoredUser{}, ErrInvalidSession
	}
	if current.CSRFTokenHash != currentCSRFTokenHash {
		return StoredUser{}, ErrInvalidSession
	}
	delete(store.sessions, currentRefreshTokenHash)
	next.UserID = current.UserID
	store.sessions[next.RefreshTokenHash] = next
	store.revoked = true
	return store.user, nil
}

func (store *memoryStore) RevokeSession(_ context.Context, refreshTokenHash string, csrfTokenHash string, allSessions bool, _ time.Time) (bool, error) {
	current, ok := store.sessions[refreshTokenHash]
	if !ok || current.CSRFTokenHash != csrfTokenHash {
		return false, ErrInvalidSession
	}
	if allSessions {
		store.sessions = map[string]SessionRecord{}
	} else {
		delete(store.sessions, refreshTokenHash)
	}
	return true, nil
}

func (store *memoryStore) RecordAuthEvent(_ context.Context, record AuthAuditRecord) error {
	store.auditEvents = append(store.auditEvents, record)
	return nil
}
