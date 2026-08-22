package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/openinvest/openinvest/backend-go/internal/auth"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestDatabaseRejectsNewSessionWithoutFamilyIdentity(t *testing.T) {
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
		Email:    "stage-03-28-db-guard-" + uuid.NewString() + "@example.com",
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

	now := time.Now().UTC()
	_, err = db.ExecContext(ctx, `
		INSERT INTO identity.sessions (
			id, user_id, refresh_token_hash, csrf_token_hash, session_state,
			expires_at, created_at, last_rotated_at
		)
		VALUES ($1, $2, $3, $4, 'active', $5, $6, $6)
	`, uuid.NewString(), registered.User.ID, uuid.NewString(), uuid.NewString(), now.Add(time.Hour), now)
	if err == nil {
		t.Fatal("expected direct database insert without session_family_id to fail")
	}
	var state interface{ SQLState() string }
	if !errors.As(err, &state) || state.SQLState() != "23514" {
		t.Fatalf("expected SQLSTATE 23514 from session family guard, got %v", err)
	}
}
