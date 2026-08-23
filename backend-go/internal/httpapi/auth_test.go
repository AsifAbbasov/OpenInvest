package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/openinvest/openinvest/backend-go/internal/auth"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestAuthRegisterSetsHttpOnlyRefreshCookieWithoutBodyLeak(t *testing.T) {
	authStore := &httpAuthTestStore{}
	authService := newHTTPAuthService(t, authStore)
	app := newHTTPAuthApp(t, authService)

	response := authRequest(t, app, http.MethodPost, "/api/v1/auth/register", `{
		"email":"Investor@Example.com",
		"password":"correct horse battery staple",
		"language":"en",
		"theme":"system",
		"timezone":"UTC"
	}`, "", "")
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, response.StatusCode)
	}
	refreshCookie := requireCookie(t, response, auth.RefreshCookieName)
	if !refreshCookie.HttpOnly {
		t.Fatalf("expected refresh cookie to be HttpOnly")
	}
	if refreshCookie.Value == "" {
		t.Fatalf("expected non-empty refresh cookie")
	}
	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode register response: %v", err)
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	text := string(encoded)
	if strings.Contains(text, refreshCookie.Value) || strings.Contains(text, "refreshToken") {
		t.Fatalf("refresh token leaked in response body: %s", text)
	}
	if !strings.Contains(text, `"privacyMode":true`) {
		t.Fatalf("privacy mode default missing from response: %s", text)
	}
}

func TestAuthRefreshRequiresCSRFAndRejectsReplay(t *testing.T) {
	authService := newHTTPAuthService(t, &httpAuthTestStore{})
	app := newHTTPAuthApp(t, authService)

	registerResponse := authRequest(t, app, http.MethodPost, "/api/v1/auth/register", `{
		"email":"investor@example.com",
		"password":"correct horse battery staple",
		"language":"en",
		"theme":"system",
		"timezone":"UTC"
	}`, "", "")
	defer registerResponse.Body.Close()
	refreshCookie := requireCookie(t, registerResponse, auth.RefreshCookieName)
	csrfToken := readCSRFToken(t, registerResponse)

	missingCSRF := authRequest(t, app, http.MethodPost, "/api/v1/auth/refresh", "", refreshCookie.Value, "")
	defer missingCSRF.Body.Close()
	if missingCSRF.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected missing csrf status %d, got %d", http.StatusUnauthorized, missingCSRF.StatusCode)
	}

	rotated := authRequest(t, app, http.MethodPost, "/api/v1/auth/refresh", "", refreshCookie.Value, csrfToken)
	defer rotated.Body.Close()
	if rotated.StatusCode != http.StatusOK {
		t.Fatalf("expected refresh status %d, got %d", http.StatusOK, rotated.StatusCode)
	}
	rotatedCookie := requireCookie(t, rotated, auth.RefreshCookieName)
	if rotatedCookie.Value == refreshCookie.Value {
		t.Fatalf("expected refresh token rotation")
	}

	replay := authRequest(t, app, http.MethodPost, "/api/v1/auth/refresh", "", refreshCookie.Value, csrfToken)
	defer replay.Body.Close()
	if replay.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected replay status %d, got %d", http.StatusUnauthorized, replay.StatusCode)
	}
}

func TestAuthRegisterRejectsUnknownSensitiveFields(t *testing.T) {
	authService := newHTTPAuthService(t, &httpAuthTestStore{})
	app := newHTTPAuthApp(t, authService)

	response := authRequest(t, app, http.MethodPost, "/api/v1/auth/register", `{
		"email":"investor@example.com",
		"password":"correct horse battery staple",
		"language":"en",
		"theme":"system",
		"timezone":"UTC",
		"inn":"123456789012"
	}`, "", "")
	defer response.Body.Close()

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected unknown sensitive field to be rejected with %d, got %d", http.StatusBadRequest, response.StatusCode)
	}
}

func TestAuthRateLimitedResponseIncludesRetryAfter(t *testing.T) {
	authService := newHTTPAuthService(t, &httpAuthTestStore{})
	app := newApp(&API{
		service:     verticalslice.NewService(&importAPITestStore{}, fixedHTTPClock{}),
		auth:        authService,
		authLimiter: newAuthRateLimiter(1, time.Minute),
	})
	body := `{
		"email":"investor@example.com",
		"password":"correct horse battery staple",
		"language":"en",
		"theme":"system",
		"timezone":"UTC"
	}`

	first := authRequest(t, app, http.MethodPost, "/api/v1/auth/register", body, "", "")
	defer first.Body.Close()
	second := authRequest(t, app, http.MethodPost, "/api/v1/auth/register", body, "", "")
	defer second.Body.Close()

	if second.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected status %d, got %d", http.StatusTooManyRequests, second.StatusCode)
	}
	if got := second.Header.Get("Retry-After"); got != authRateLimitRetryAfterSeconds {
		t.Fatalf("expected Retry-After %q, got %q", authRateLimitRetryAfterSeconds, got)
	}
}

func TestAuthLogoutRateLimitBoundsRejectedAuditWrites(t *testing.T) {
	store := &httpAuthTestStore{}
	authService := newHTTPAuthService(t, store)
	app := newApp(&API{
		service:     verticalslice.NewService(&importAPITestStore{}, fixedHTTPClock{}),
		auth:        authService,
		authLimiter: newBoundedAuthRateLimiter(1, 1, 8, time.Minute),
		now:         func() time.Time { return time.Date(2026, 8, 23, 10, 0, 0, 0, time.UTC) },
	})

	first := authRequest(t, app, http.MethodPost, "/api/v1/auth/logout", `{"allSessions":false}`, "", "")
	defer first.Body.Close()
	if first.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected first invalid logout to return %d, got %d", http.StatusUnauthorized, first.StatusCode)
	}
	if len(store.authEvents) != 1 || store.authEvents[0].ActionCode != "AUTH_LOGOUT_REJECTED" {
		t.Fatalf("expected one rejected logout audit event, got %+v", store.authEvents)
	}

	second := authRequest(t, app, http.MethodPost, "/api/v1/auth/logout", `{"allSessions":false}`, "", "")
	defer second.Body.Close()
	if second.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected repeated invalid logout to be rate limited with %d, got %d", http.StatusTooManyRequests, second.StatusCode)
	}
	if got := second.Header.Get("Retry-After"); got != authRateLimitRetryAfterSeconds {
		t.Fatalf("expected Retry-After %q, got %q", authRateLimitRetryAfterSeconds, got)
	}
	if len(store.authEvents) != 1 {
		t.Fatalf("rate-limited logout must not create another audit write, got %d events", len(store.authEvents))
	}
}

func TestAuthRateLimiterBoundsUniqueKeyCardinalityAndReclaimsExpiredBuckets(t *testing.T) {
	limiter := newBoundedAuthRateLimiter(2, 100, 2, time.Minute)
	now := time.Date(2026, 8, 23, 10, 0, 0, 0, time.UTC)

	if !limiter.allow("/login|192.0.2.1", now) || !limiter.allow("/login|192.0.2.2", now) {
		t.Fatal("expected first two keys to be admitted")
	}
	if limiter.allow("/login|192.0.2.3", now) {
		t.Fatal("expected new key to fail closed at active-key capacity")
	}
	if got := len(limiter.attempts); got != 2 {
		t.Fatalf("expected key map bounded at 2, got %d", got)
	}

	afterExpiry := now.Add(2 * time.Minute)
	if !limiter.allow("/login|192.0.2.3", afterExpiry) {
		t.Fatal("expected expired buckets to be reclaimed")
	}
	if got := len(limiter.attempts); got != 1 {
		t.Fatalf("expected one active key after reclamation, got %d", got)
	}
}

func TestAuthRateLimiterAppliesGlobalBudgetAcrossUniqueKeys(t *testing.T) {
	limiter := newBoundedAuthRateLimiter(20, 2, 16, time.Minute)
	now := time.Date(2026, 8, 23, 10, 0, 0, 0, time.UTC)

	if !limiter.allow("/logout|192.0.2.1", now) || !limiter.allow("/logout|192.0.2.2", now) {
		t.Fatal("expected first two global attempts to be admitted")
	}
	if limiter.allow("/logout|192.0.2.3", now) {
		t.Fatal("expected third unique-key attempt to fail at global budget")
	}
	if got := len(limiter.globalAttempts); got != 2 {
		t.Fatalf("expected bounded global memory of 2, got %d", got)
	}

	if !limiter.allow("/logout|192.0.2.3", now.Add(2*time.Minute)) {
		t.Fatal("expected global budget reset after window")
	}
}

func TestAuthLogoutRequiresExplicitScopeBody(t *testing.T) {
	authService := newHTTPAuthService(t, &httpAuthTestStore{})
	app := newHTTPAuthApp(t, authService)

	registerResponse := authRequest(t, app, http.MethodPost, "/api/v1/auth/register", `{
		"email":"investor@example.com",
		"password":"correct horse battery staple",
		"language":"en",
		"theme":"system",
		"timezone":"UTC"
	}`, "", "")
	defer registerResponse.Body.Close()
	refreshCookie := requireCookie(t, registerResponse, auth.RefreshCookieName)
	csrfToken := readCSRFToken(t, registerResponse)

	logout := authRequest(t, app, http.MethodPost, "/api/v1/auth/logout", "", refreshCookie.Value, csrfToken)
	defer logout.Body.Close()
	if logout.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected empty logout body to be rejected with %d, got %d", http.StatusBadRequest, logout.StatusCode)
	}

	logout = authRequest(t, app, http.MethodPost, "/api/v1/auth/logout", `{}`, refreshCookie.Value, csrfToken)
	defer logout.Body.Close()
	if logout.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected missing allSessions to be rejected with %d, got %d", http.StatusBadRequest, logout.StatusCode)
	}
}

func TestAuthLogoutClearsRefreshCookie(t *testing.T) {
	authService := newHTTPAuthService(t, &httpAuthTestStore{})
	app := newHTTPAuthApp(t, authService)

	registerResponse := authRequest(t, app, http.MethodPost, "/api/v1/auth/register", `{
		"email":"investor@example.com",
		"password":"correct horse battery staple",
		"language":"en",
		"theme":"system",
		"timezone":"UTC"
	}`, "", "")
	defer registerResponse.Body.Close()
	refreshCookie := requireCookie(t, registerResponse, auth.RefreshCookieName)
	csrfToken := readCSRFToken(t, registerResponse)

	logout := authRequest(t, app, http.MethodPost, "/api/v1/auth/logout", `{"allSessions":false}`, refreshCookie.Value, csrfToken)
	defer logout.Body.Close()
	if logout.StatusCode != http.StatusOK {
		t.Fatalf("expected logout status %d, got %d", http.StatusOK, logout.StatusCode)
	}
	cleared := requireCookie(t, logout, auth.RefreshCookieName)
	if cleared.MaxAge >= 0 {
		t.Fatalf("expected refresh cookie to be expired, got max-age %d", cleared.MaxAge)
	}
}

func newHTTPAuthService(t *testing.T, store *httpAuthTestStore) *auth.Service {
	t.Helper()
	service, err := auth.NewService(store, fixedHTTPClock{}, auth.Config{
		AccessTokenSecret:   []byte("01234567890123456789012345678901"),
		RefreshCookieSecure: true,
	})
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}
	return service
}

func newHTTPAuthApp(t *testing.T, authService *auth.Service) *fiber.App {
	t.Helper()
	app, err := New(
		verticalslice.NewService(&importAPITestStore{}, fixedHTTPClock{}),
		authService,
		[]byte("test-import-review-token-secret-32-bytes"),
	)
	if err != nil {
		t.Fatalf("new authenticated HTTP app: %v", err)
	}
	return app
}

func authRequest(t *testing.T, app *fiber.App, method string, path string, body string, refreshToken string, csrfToken string) *http.Response {
	t.Helper()
	request := httptest.NewRequest(method, path, bytes.NewReader([]byte(body)))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	if refreshToken != "" {
		request.AddCookie(&http.Cookie{Name: auth.RefreshCookieName, Value: refreshToken})
	}
	if csrfToken != "" {
		request.Header.Set("X-CSRF-Token", csrfToken)
	}
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request %s %s: %v", method, path, err)
	}
	return response
}

func requireCookie(t *testing.T, response *http.Response, name string) *http.Cookie {
	t.Helper()
	for _, cookie := range response.Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	t.Fatalf("missing cookie %q", name)
	return nil
}

func readCSRFToken(t *testing.T, response *http.Response) string {
	t.Helper()
	var payload struct {
		Data struct {
			Session struct {
				CSRFToken string `json:"csrfToken"`
			} `json:"session"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode auth response: %v", err)
	}
	if payload.Data.Session.CSRFToken == "" {
		t.Fatalf("missing csrf token in response")
	}
	return payload.Data.Session.CSRFToken
}

type httpAuthTestStore struct {
	user       auth.StoredUser
	password   string
	sessions   map[string]auth.SessionRecord
	authEvents []auth.AuthAuditRecord
}

func (store *httpAuthTestStore) RegisterUser(_ context.Context, record auth.RegistrationRecord) (auth.StoredUser, error) {
	store.password = record.PasswordHash
	store.user = auth.StoredUser{
		ID:                  record.UserID,
		InvestmentSubjectID: record.InvestmentSubjectID,
		Email:               record.EmailNormalized,
		Language:            record.Language,
		Theme:               record.Theme,
		Timezone:            record.Timezone,
		PrivacyMode:         true,
		CreatedAt:           record.Now,
	}
	store.sessions = map[string]auth.SessionRecord{record.Session.RefreshTokenHash: record.Session}
	return store.user, nil
}

func (store *httpAuthTestStore) FindUserByEmail(_ context.Context, email string) (auth.StoredUser, string, error) {
	if email != store.user.Email {
		return auth.StoredUser{}, "", auth.ErrInvalidCredentials
	}
	return store.user, store.password, nil
}

func (store *httpAuthTestStore) CreateSession(_ context.Context, record auth.SessionRecord) error {
	if store.sessions == nil {
		store.sessions = map[string]auth.SessionRecord{}
	}
	store.sessions[record.RefreshTokenHash] = record
	return nil
}

func (store *httpAuthTestStore) RotateSession(_ context.Context, currentRefreshTokenHash string, currentCSRFTokenHash string, next auth.SessionRecord, _ time.Time) (auth.StoredUser, error) {
	current, ok := store.sessions[currentRefreshTokenHash]
	if !ok || current.CSRFTokenHash != currentCSRFTokenHash {
		return auth.StoredUser{}, auth.ErrInvalidSession
	}
	delete(store.sessions, currentRefreshTokenHash)
	next.UserID = current.UserID
	store.sessions[next.RefreshTokenHash] = next
	return store.user, nil
}

func (store *httpAuthTestStore) RevokeSession(_ context.Context, refreshTokenHash string, csrfTokenHash string, allSessions bool, _ time.Time) (bool, error) {
	current, ok := store.sessions[refreshTokenHash]
	if !ok || current.CSRFTokenHash != csrfTokenHash {
		return false, auth.ErrInvalidSession
	}
	if allSessions {
		store.sessions = map[string]auth.SessionRecord{}
	} else {
		delete(store.sessions, refreshTokenHash)
	}
	return true, nil
}

func (store *httpAuthTestStore) RecordAuthEvent(_ context.Context, record auth.AuthAuditRecord) error {
	store.authEvents = append(store.authEvents, record)
	return nil
}
