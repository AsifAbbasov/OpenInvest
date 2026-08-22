package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"sync"
	"testing"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/openinvest/openinvest/backend-go/internal/auth"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestStoreAuthPrivacySessionLifecycle(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}

	store, err := postgres.Open(databaseURL)
	if err != nil {
		t.Fatalf("open postgres store: %v", err)
	}
	defer store.Close()

	service, err := auth.NewService(store, verticalslice.SystemClock{}, auth.Config{
		AccessTokenSecret:   []byte("01234567890123456789012345678901"),
		RefreshCookieSecure: true,
	})
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}
	ctx := context.Background()
	result, err := service.Register(ctx, auth.RegistrationRequest{
		Email:    "stage-03-11-auth-" + uuid.NewString() + "@example.com",
		Password: "correct horse battery staple",
		Language: auth.LanguageEN,
		Theme:    auth.ThemeSystem,
		Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if !result.User.PrivacyMode {
		t.Fatalf("expected privacy mode enabled")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open db for privacy evidence: %v", err)
	}
	defer db.Close()
	var privacyMode bool
	var taxProfileEnabled bool
	var notificationsEnabled bool
	var analyticsMode string
	if err := db.QueryRowContext(ctx, `
		SELECT privacy_mode, tax_profile_enabled, notifications_enabled, analytics_mode
		FROM identity.privacy_settings
		WHERE user_id = $1
	`, result.User.ID).Scan(&privacyMode, &taxProfileEnabled, &notificationsEnabled, &analyticsMode); err != nil {
		t.Fatalf("query privacy settings: %v", err)
	}
	if !privacyMode || taxProfileEnabled || notificationsEnabled || analyticsMode != "anonymous" {
		t.Fatalf("unexpected privacy defaults: privacy=%t tax=%t notifications=%t analytics=%q", privacyMode, taxProfileEnabled, notificationsEnabled, analyticsMode)
	}

	independent, err := service.Login(ctx, auth.LoginRequest{Email: result.User.Email, Password: "correct horse battery staple"})
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	rotated, err := service.Refresh(ctx, result.RefreshToken, result.Session.CSRFToken)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if rotated.RefreshToken == result.RefreshToken {
		t.Fatalf("expected refresh token rotation")
	}

	if _, err := service.Refresh(ctx, result.RefreshToken, result.Session.CSRFToken); !errors.Is(err, auth.ErrInvalidSession) {
		t.Fatalf("expected replay to be rejected, got %v", err)
	}
	if _, err := service.Refresh(ctx, rotated.RefreshToken, rotated.Session.CSRFToken); !errors.Is(err, auth.ErrInvalidSession) {
		t.Fatalf("expected replay to revoke the rotated descendant, got %v", err)
	}

	independentRotated, err := service.Refresh(ctx, independent.RefreshToken, independent.Session.CSRFToken)
	if err != nil {
		t.Fatalf("independent login family must survive replay in another family: %v", err)
	}

	if _, err := service.Refresh(ctx, "unknown-refresh-token", "unknown-csrf-token"); !errors.Is(err, auth.ErrInvalidSession) {
		t.Fatalf("expected unknown refresh token to be rejected, got %v", err)
	}

	if _, err := service.Logout(ctx, "", independentRotated.Session.CSRFToken, false); !errors.Is(err, auth.ErrInvalidSession) {
		t.Fatalf("expected missing refresh token logout to be rejected, got %v", err)
	}

	revoked, err := service.Logout(ctx, independentRotated.RefreshToken, independentRotated.Session.CSRFToken, false)
	if err != nil {
		t.Fatalf("logout: %v", err)
	}
	if !revoked {
		t.Fatalf("expected logout to revoke session")
	}
	if _, err := service.Refresh(ctx, independentRotated.RefreshToken, independentRotated.Session.CSRFToken); !errors.Is(err, auth.ErrInvalidSession) {
		t.Fatalf("expected revoked refresh token to be rejected, got %v", err)
	}

	rows, err := db.QueryContext(ctx, `
		SELECT action_code, outcome, count(*)
		FROM audit.events
		WHERE action_code IN (
			'AUTH_REGISTER',
			'AUTH_LOGIN',
			'AUTH_REFRESH',
			'AUTH_REFRESH_REPLAY',
			'AUTH_REFRESH_REJECTED',
			'AUTH_LOGOUT',
			'AUTH_LOGOUT_REJECTED'
		)
		GROUP BY action_code, outcome
	`)
	if err != nil {
		t.Fatalf("query auth audit evidence: %v", err)
	}
	defer rows.Close()
	auditCounts := map[string]int{}
	for rows.Next() {
		var action string
		var outcome string
		var count int
		if err := rows.Scan(&action, &outcome, &count); err != nil {
			t.Fatalf("scan auth audit evidence: %v", err)
		}
		auditCounts[action+"|"+outcome] = count
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate auth audit evidence: %v", err)
	}
	for _, key := range []string{
		"AUTH_REGISTER|success",
		"AUTH_LOGIN|success",
		"AUTH_REFRESH|success",
		"AUTH_REFRESH_REPLAY|failure",
		"AUTH_REFRESH_REJECTED|failure",
		"AUTH_LOGOUT|success",
		"AUTH_LOGOUT_REJECTED|failure",
	} {
		if auditCounts[key] < 1 {
			t.Fatalf("missing auth audit evidence %s in %#v", key, auditCounts)
		}
	}
}

func TestConcurrentRefreshReplayRevokesWinningSessionFamily(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}

	store, err := postgres.Open(databaseURL)
	if err != nil {
		t.Fatalf("open postgres store: %v", err)
	}
	defer store.Close()

	service, err := auth.NewService(store, verticalslice.SystemClock{}, auth.Config{
		AccessTokenSecret:   []byte("01234567890123456789012345678901"),
		RefreshCookieSecure: true,
	})
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}

	ctx := context.Background()
	registered, err := service.Register(ctx, auth.RegistrationRequest{
		Email:    "stage-03-28-replay-" + uuid.NewString() + "@example.com",
		Password: "correct horse battery staple",
		Language: auth.LanguageEN,
		Theme:    auth.ThemeSystem,
		Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	independent, err := service.Login(ctx, auth.LoginRequest{
		Email:    registered.User.Email,
		Password: "correct horse battery staple",
	})
	if err != nil {
		t.Fatalf("independent login: %v", err)
	}

	type refreshResult struct {
		result auth.AuthResult
		err    error
	}
	results := make(chan refreshResult, 2)
	start := make(chan struct{})
	var ready sync.WaitGroup
	ready.Add(2)

	refresh := func() {
		ready.Done()
		<-start
		result, err := service.Refresh(ctx, registered.RefreshToken, registered.Session.CSRFToken)
		results <- refreshResult{result: result, err: err}
	}
	go refresh()
	go refresh()
	ready.Wait()
	close(start)

	var winner auth.AuthResult
	successes := 0
	rejections := 0
	for i := 0; i < 2; i++ {
		outcome := <-results
		switch {
		case outcome.err == nil:
			successes++
			winner = outcome.result
		case errors.Is(outcome.err, auth.ErrInvalidSession):
			rejections++
		default:
			t.Fatalf("unexpected concurrent refresh error: %v", outcome.err)
		}
	}
	if successes != 1 || rejections != 1 {
		t.Fatalf("expected one rotation success and one replay rejection, got successes=%d rejections=%d", successes, rejections)
	}

	if _, err := service.Refresh(ctx, winner.RefreshToken, winner.Session.CSRFToken); !errors.Is(err, auth.ErrInvalidSession) {
		t.Fatalf("expected replay detector to revoke the winning descendant, got %v", err)
	}
	if _, err := service.Refresh(ctx, independent.RefreshToken, independent.Session.CSRFToken); err != nil {
		t.Fatalf("independent login family must remain active after replay containment: %v", err)
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open db for replay evidence: %v", err)
	}
	defer db.Close()

	var activeFamilies int
	if err := db.QueryRowContext(ctx, `
		SELECT count(DISTINCT session_family_id)
		FROM identity.sessions
		WHERE user_id = $1
			AND session_state = 'active'
	`, registered.User.ID).Scan(&activeFamilies); err != nil {
		t.Fatalf("query active session families: %v", err)
	}
	if activeFamilies != 1 {
		t.Fatalf("expected only the independent login family to remain active, got %d active families", activeFamilies)
	}
}
