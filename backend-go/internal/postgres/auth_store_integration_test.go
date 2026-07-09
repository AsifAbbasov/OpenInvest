package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
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

	revoked, err := service.Logout(ctx, rotated.RefreshToken, rotated.Session.CSRFToken, false)
	if err != nil {
		t.Fatalf("logout: %v", err)
	}
	if !revoked {
		t.Fatalf("expected logout to revoke session")
	}
	if _, err := service.Refresh(ctx, rotated.RefreshToken, rotated.Session.CSRFToken); !errors.Is(err, auth.ErrInvalidSession) {
		t.Fatalf("expected revoked refresh token to be rejected, got %v", err)
	}
}
