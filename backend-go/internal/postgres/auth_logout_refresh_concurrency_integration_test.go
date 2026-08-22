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

func TestLogoutRacingRefreshCannotLeaveDescendantActive(t *testing.T) {
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
		Email:    "stage-03-28-logout-race-" + uuid.NewString() + "@example.com",
		Password: "correct horse battery staple",
		Language: auth.LanguageEN,
		Theme:    auth.ThemeSystem,
		Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	type refreshOutcome struct {
		result auth.AuthResult
		err    error
	}
	type logoutOutcome struct {
		revoked bool
		err     error
	}
	refreshResults := make(chan refreshOutcome, 1)
	logoutResults := make(chan logoutOutcome, 1)
	start := make(chan struct{})
	var ready sync.WaitGroup
	ready.Add(2)

	go func() {
		ready.Done()
		<-start
		result, err := service.Refresh(ctx, root.RefreshToken, root.Session.CSRFToken)
		refreshResults <- refreshOutcome{result: result, err: err}
	}()
	go func() {
		ready.Done()
		<-start
		revoked, err := service.Logout(ctx, root.RefreshToken, root.Session.CSRFToken, false)
		logoutResults <- logoutOutcome{revoked: revoked, err: err}
	}()
	ready.Wait()
	close(start)

	refresh := <-refreshResults
	logout := <-logoutResults
	if logout.err != nil || !logout.revoked {
		t.Fatalf("logout must succeed despite concurrent refresh: revoked=%t err=%v", logout.revoked, logout.err)
	}
	if refresh.err != nil && !errors.Is(refresh.err, auth.ErrInvalidSession) {
		t.Fatalf("unexpected refresh race error: %v", refresh.err)
	}
	if refresh.err == nil {
		if _, err := service.Refresh(ctx, refresh.result.RefreshToken, refresh.result.Session.CSRFToken); !errors.Is(err, auth.ErrInvalidSession) {
			t.Fatalf("refresh descendant survived concurrent logout containment: %v", err)
		}
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
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
		t.Fatalf("logout/refresh race left %d active session(s)", active)
	}
}
