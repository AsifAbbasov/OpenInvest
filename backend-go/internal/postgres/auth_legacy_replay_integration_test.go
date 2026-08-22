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

func TestLegacyRefreshReplayWithUnknownLineageFailsClosed(t *testing.T) {
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
	root, err := service.Register(ctx, auth.RegistrationRequest{
		Email:    "stage-03-28-legacy-" + uuid.NewString() + "@example.com",
		Password: "correct horse battery staple",
		Language: auth.LanguageEN,
		Theme:    auth.ThemeSystem,
		Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if _, err := db.ExecContext(ctx, `
		UPDATE identity.sessions
		SET session_family_id = NULL
		WHERE user_id = $1
	`, root.User.ID); err != nil {
		t.Fatalf("simulate pre-Stage-3.28 session: %v", err)
	}

	descendant, err := service.Refresh(ctx, root.RefreshToken, root.Session.CSRFToken)
	if err != nil {
		t.Fatalf("rotate legacy active session: %v", err)
	}
	independent, err := service.Login(ctx, auth.LoginRequest{
		Email:    root.User.Email,
		Password: "correct horse battery staple",
	})
	if err != nil {
		t.Fatalf("independent login: %v", err)
	}

	if _, err := service.Refresh(ctx, root.RefreshToken, root.Session.CSRFToken); !errors.Is(err, auth.ErrInvalidSession) {
		t.Fatalf("expected legacy replay rejection, got %v", err)
	}
	if _, err := service.Refresh(ctx, descendant.RefreshToken, descendant.Session.CSRFToken); !errors.Is(err, auth.ErrInvalidSession) {
		t.Fatalf("legacy replay must revoke descendant with reconstructed family: %v", err)
	}
	if _, err := service.Refresh(ctx, independent.RefreshToken, independent.Session.CSRFToken); !errors.Is(err, auth.ErrInvalidSession) {
		t.Fatalf("unknown legacy lineage must fail closed across active user sessions: %v", err)
	}

	var active int
	if err := db.QueryRowContext(ctx, `
		SELECT count(*)
		FROM identity.sessions
		WHERE user_id = $1
			AND session_state = 'active'
	`, root.User.ID).Scan(&active); err != nil {
		t.Fatalf("query active sessions: %v", err)
	}
	if active != 0 {
		t.Fatalf("legacy replay containment left %d active session(s)", active)
	}
}
