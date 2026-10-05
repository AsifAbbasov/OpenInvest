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

func newAuthTheftPostgresService(t *testing.T) (*auth.Service, *postgres.Store, *sql.DB) {
	t.Helper()
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}
	store, err := postgres.Open(databaseURL)
	if err != nil {
		t.Fatalf("open postgres store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	service, err := auth.NewService(store, verticalslice.SystemClock{}, auth.Config{
		AccessTokenSecret:   []byte("01234567890123456789012345678901"),
		RefreshCookieSecure: true,
	})
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open verification db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return service, store, db
}

func registerAuthTheftUser(t *testing.T, service *auth.Service, prefix string) auth.AuthResult {
	t.Helper()
	result, err := service.Register(context.Background(), auth.RegistrationRequest{
		Email:    prefix + "-" + uuid.NewString() + "@example.com",
		Password: "correct horse battery staple",
		Language: auth.LanguageEN,
		Theme:    auth.ThemeSystem,
		Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("register auth theft user: %v", err)
	}
	return result
}

func TestAuthTheftPostgresCookieOnlyAndWrongCSRFFailWithoutDestroyingLegitimateSession(t *testing.T) {
	service, _, _ := newAuthTheftPostgresService(t)
	root := registerAuthTheftUser(t, service, "auth-theft-csrf")

	for i := 0; i < 100; i++ {
		_, err := service.Refresh(context.Background(), root.RefreshToken, "wrong-csrf-"+uuid.NewString())
		if !errors.Is(err, auth.ErrInvalidSession) {
			t.Fatalf("wrong csrf attempt %d: got %v want invalid session", i, err)
		}
	}

	legitimate, err := service.Refresh(context.Background(), root.RefreshToken, root.Session.CSRFToken)
	if err != nil {
		t.Fatalf("legitimate refresh failed after wrong-CSRF pressure: %v", err)
	}
	if legitimate.RefreshToken == root.RefreshToken {
		t.Fatal("legitimate refresh did not rotate token after theft pressure")
	}
}

func TestAuthTheftPostgresThirtyTwoConcurrentUsesOfStolenPairContainFamily(t *testing.T) {
	service, _, db := newAuthTheftPostgresService(t)
	ctx := context.Background()
	root := registerAuthTheftUser(t, service, "auth-theft-race")
	independent, err := service.Login(ctx, auth.LoginRequest{
		Email:    root.User.Email,
		Password: "correct horse battery staple",
	})
	if err != nil {
		t.Fatalf("create independent family: %v", err)
	}

	type outcome struct {
		result auth.AuthResult
		err    error
	}
	const contenders = 32
	start := make(chan struct{})
	results := make(chan outcome, contenders)
	var ready sync.WaitGroup
	ready.Add(contenders)
	for i := 0; i < contenders; i++ {
		go func() {
			ready.Done()
			<-start
			got, err := service.Refresh(ctx, root.RefreshToken, root.Session.CSRFToken)
			results <- outcome{result: got, err: err}
		}()
	}
	ready.Wait()
	close(start)

	successes := 0
	rejections := 0
	var winner auth.AuthResult
	for i := 0; i < contenders; i++ {
		got := <-results
		switch {
		case got.err == nil:
			successes++
			winner = got.result
		case errors.Is(got.err, auth.ErrInvalidSession):
			rejections++
		default:
			t.Fatalf("unexpected stolen-pair race result: %v", got.err)
		}
	}
	if successes != 1 || rejections != contenders-1 {
		t.Fatalf("stolen-pair race escaped single-winner serialization: successes=%d rejections=%d", successes, rejections)
	}

	// The first replay after the winner must revoke the entire compromised family.
	if winner.RefreshToken != "" {
		if _, err := service.Refresh(ctx, winner.RefreshToken, winner.Session.CSRFToken); !errors.Is(err, auth.ErrInvalidSession) {
			t.Fatalf("winner descendant survived replay containment: %v", err)
		}
	}

	var activeCompromisedFamily int
	if err := db.QueryRowContext(ctx, `
		SELECT count(*)
		FROM identity.sessions
		WHERE user_id = $1
		  AND session_family_id = (
		    SELECT session_family_id
		    FROM identity.sessions
		    WHERE user_id = $1
		    ORDER BY created_at ASC
		    LIMIT 1
		  )
		  AND session_state = 'active'
	`, root.User.ID).Scan(&activeCompromisedFamily); err != nil {
		t.Fatalf("query compromised family: %v", err)
	}
	if activeCompromisedFamily != 0 {
		t.Fatalf("compromised session family retained %d active descendants", activeCompromisedFamily)
	}

	if _, err := service.Refresh(ctx, independent.RefreshToken, independent.Session.CSRFToken); err != nil {
		t.Fatalf("independent session family was collateral-damaged by replay containment: %v", err)
	}
}

func TestAuthTheftPostgresAllSessionsLogoutRevokesEveryIndependentFamily(t *testing.T) {
	service, _, db := newAuthTheftPostgresService(t)
	ctx := context.Background()
	root := registerAuthTheftUser(t, service, "auth-theft-all-logout")
	first, err := service.Login(ctx, auth.LoginRequest{Email: root.User.Email, Password: "correct horse battery staple"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.Login(ctx, auth.LoginRequest{Email: root.User.Email, Password: "correct horse battery staple"})
	if err != nil {
		t.Fatal(err)
	}

	revoked, err := service.Logout(ctx, first.RefreshToken, first.Session.CSRFToken, true)
	if err != nil || !revoked {
		t.Fatalf("all-sessions logout failed: revoked=%t err=%v", revoked, err)
	}

	for index, session := range []auth.AuthResult{root, first, second} {
		if _, err := service.Refresh(ctx, session.RefreshToken, session.Session.CSRFToken); !errors.Is(err, auth.ErrInvalidSession) {
			t.Fatalf("session family %d survived all-sessions logout: %v", index, err)
		}
	}

	var active int
	if err := db.QueryRowContext(ctx,
		"SELECT count(*) FROM identity.sessions WHERE user_id=$1 AND session_state='active'",
		root.User.ID,
	).Scan(&active); err != nil {
		t.Fatal(err)
	}
	if active != 0 {
		t.Fatalf("all-sessions logout left %d active sessions", active)
	}
}

func TestAuthTheftPostgresAllSessionsLogoutRacingStolenRefreshLeavesNoSessionActive(t *testing.T) {
	service, _, db := newAuthTheftPostgresService(t)
	ctx := context.Background()
	root := registerAuthTheftUser(t, service, "auth-theft-all-race")
	legitimate, err := service.Login(ctx, auth.LoginRequest{
		Email:    root.User.Email,
		Password: "correct horse battery staple",
	})
	if err != nil {
		t.Fatal(err)
	}

	start := make(chan struct{})
	var ready sync.WaitGroup
	ready.Add(2)
	refreshDone := make(chan error, 1)
	logoutDone := make(chan error, 1)

	go func() {
		ready.Done()
		<-start
		_, err := service.Refresh(ctx, root.RefreshToken, root.Session.CSRFToken)
		refreshDone <- err
	}()
	go func() {
		ready.Done()
		<-start
		revoked, err := service.Logout(ctx, legitimate.RefreshToken, legitimate.Session.CSRFToken, true)
		if err == nil && !revoked {
			err = errors.New("all-sessions logout returned revoked=false")
		}
		logoutDone <- err
	}()
	ready.Wait()
	close(start)

	refreshErr := <-refreshDone
	logoutErr := <-logoutDone
	if refreshErr != nil && !errors.Is(refreshErr, auth.ErrInvalidSession) {
		t.Fatalf("unexpected stolen refresh race error: %v", refreshErr)
	}
	if logoutErr != nil {
		t.Fatalf("all-sessions logout lost race: %v", logoutErr)
	}

	var active int
	if err := db.QueryRowContext(ctx,
		"SELECT count(*) FROM identity.sessions WHERE user_id=$1 AND session_state='active'",
		root.User.ID,
	).Scan(&active); err != nil {
		t.Fatal(err)
	}
	if active != 0 {
		t.Fatalf("logout/refresh race left %d active session(s)", active)
	}
}
