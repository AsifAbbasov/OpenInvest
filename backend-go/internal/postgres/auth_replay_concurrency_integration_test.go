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

func TestReplayRacingDescendantRefreshLeavesNoActiveFamilyMember(t *testing.T) {
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
		Email:    "stage-03-28-family-race-" + uuid.NewString() + "@example.com",
		Password: "correct horse battery staple",
		Language: auth.LanguageEN,
		Theme:    auth.ThemeSystem,
		Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	descendant, err := service.Refresh(ctx, root.RefreshToken, root.Session.CSRFToken)
	if err != nil {
		t.Fatalf("initial rotation: %v", err)
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	var familyID string
	if err := db.QueryRowContext(ctx, `
		SELECT session_family_id
		FROM identity.sessions
		WHERE user_id = $1
		ORDER BY created_at
		LIMIT 1
	`, root.User.ID).Scan(&familyID); err != nil {
		t.Fatalf("query family id: %v", err)
	}

	type result struct {
		auth auth.AuthResult
		err  error
	}
	outcomes := make(chan result, 2)
	start := make(chan struct{})
	var ready sync.WaitGroup
	ready.Add(2)

	go func() {
		ready.Done()
		<-start
		got, err := service.Refresh(ctx, root.RefreshToken, root.Session.CSRFToken)
		outcomes <- result{auth: got, err: err}
	}()
	go func() {
		ready.Done()
		<-start
		got, err := service.Refresh(ctx, descendant.RefreshToken, descendant.Session.CSRFToken)
		outcomes <- result{auth: got, err: err}
	}()
	ready.Wait()
	close(start)

	var descendantWinner auth.AuthResult
	for i := 0; i < 2; i++ {
		outcome := <-outcomes
		if outcome.err == nil {
			descendantWinner = outcome.auth
			continue
		}
		if !errors.Is(outcome.err, auth.ErrInvalidSession) {
			t.Fatalf("unexpected refresh race error: %v", outcome.err)
		}
	}

	var active int
	if err := db.QueryRowContext(ctx, `
		SELECT count(*)
		FROM identity.sessions
		WHERE session_family_id = $1
			AND session_state = 'active'
	`, familyID).Scan(&active); err != nil {
		t.Fatalf("query active family sessions: %v", err)
	}
	if active != 0 {
		t.Fatalf("replay containment left %d active session(s) in the compromised family", active)
	}

	if descendantWinner.RefreshToken != "" {
		if _, err := service.Refresh(ctx, descendantWinner.RefreshToken, descendantWinner.Session.CSRFToken); !errors.Is(err, auth.ErrInvalidSession) {
			t.Fatalf("descendant created during replay race survived containment: %v", err)
		}
	}
}
