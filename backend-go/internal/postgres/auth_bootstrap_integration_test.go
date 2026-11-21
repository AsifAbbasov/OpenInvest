package postgres_test

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/openinvest/openinvest/backend-go/internal/auth"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func bootstrapTokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func newBootstrapIntegrationFixture(
	t *testing.T,
) (*auth.Service, *sql.DB) {
	t.Helper()

	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}

	store, err := postgres.Open(databaseURL)
	if err != nil {
		t.Fatalf("open PostgreSQL store: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open verification connection: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping PostgreSQL: %v", err)
	}

	service, err := auth.NewService(
		store,
		verticalslice.SystemClock{},
		auth.Config{
			AccessTokenSecret: []byte(
				"01234567890123456789012345678901",
			),
			RefreshCookieSecure: true,
		},
	)
	if err != nil {
		t.Fatalf("create auth service: %v", err)
	}

	return service, db
}

func registerBootstrapIntegrationUser(
	t *testing.T,
	service *auth.Service,
) auth.AuthResult {
	t.Helper()

	result, err := service.Register(
		context.Background(),
		auth.RegistrationRequest{
			Email: "bootstrap-" +
				uuid.NewString() +
				"@example.com",
			Password: "correct horse battery staple",
			Language: auth.LanguageEN,
			Theme:    auth.ThemeSystem,
			Timezone: "UTC",
		},
	)
	if err != nil {
		t.Fatalf("register test user: %v", err)
	}

	return result
}

func TestBootstrapPostgresPreservesRefreshSession(t *testing.T) {
	service, db := newBootstrapIntegrationFixture(t)
	ctx := context.Background()

	initial := registerBootstrapIntegrationUser(t, service)
	refreshHash := bootstrapTokenHash(initial.RefreshToken)

	var sessionID string
	var familyBefore string
	var refreshBefore string
	var csrfBefore string
	var stateBefore string
	var expiryBefore time.Time

	err := db.QueryRowContext(ctx, `
		SELECT
			id,
			session_family_id,
			refresh_token_hash,
			csrf_token_hash,
			session_state,
			expires_at
		FROM identity.sessions
		WHERE refresh_token_hash = $1
	`, refreshHash).Scan(
		&sessionID,
		&familyBefore,
		&refreshBefore,
		&csrfBefore,
		&stateBefore,
		&expiryBefore,
	)
	if err != nil {
		t.Fatalf("read original session: %v", err)
	}

	restored, err := service.Bootstrap(
		ctx,
		initial.RefreshToken,
	)
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	if restored.User.ID != initial.User.ID {
		t.Fatal("bootstrap restored a different user")
	}

	if restored.RefreshToken != "" {
		t.Fatal("bootstrap returned a refresh token")
	}

	if restored.Session.AccessToken == "" {
		t.Fatal("bootstrap did not return an access token")
	}

	if restored.Session.CSRFToken == "" {
		t.Fatal("bootstrap did not return a CSRF token")
	}

	if restored.Session.CSRFToken == initial.Session.CSRFToken {
		t.Fatal("bootstrap reused the original CSRF token")
	}

	var sessionAfter string
	var familyAfter string
	var refreshAfter string
	var csrfAfter string
	var stateAfter string
	var expiryAfter time.Time

	err = db.QueryRowContext(ctx, `
		SELECT
			id,
			session_family_id,
			refresh_token_hash,
			csrf_token_hash,
			session_state,
			expires_at
		FROM identity.sessions
		WHERE id = $1
	`, sessionID).Scan(
		&sessionAfter,
		&familyAfter,
		&refreshAfter,
		&csrfAfter,
		&stateAfter,
		&expiryAfter,
	)
	if err != nil {
		t.Fatalf("read restored session: %v", err)
	}

	if sessionAfter != sessionID ||
		familyAfter != familyBefore ||
		refreshAfter != refreshBefore ||
		stateAfter != stateBefore ||
		!expiryAfter.Equal(expiryBefore) {
		t.Fatal("bootstrap changed protected session fields")
	}

	if csrfAfter == csrfBefore {
		t.Fatal("CSRF hash was not replaced")
	}

	if csrfAfter != bootstrapTokenHash(restored.Session.CSRFToken) {
		t.Fatal("PostgreSQL CSRF hash does not match returned token")
	}

	_, err = service.Refresh(
		ctx,
		initial.RefreshToken,
		initial.Session.CSRFToken,
	)
	if !errors.Is(err, auth.ErrInvalidSession) {
		t.Fatalf("old CSRF was accepted: %v", err)
	}

	rotated, err := service.Refresh(
		ctx,
		initial.RefreshToken,
		restored.Session.CSRFToken,
	)
	if err != nil {
		t.Fatalf("refresh after bootstrap: %v", err)
	}

	if rotated.RefreshToken == initial.RefreshToken {
		t.Fatal("explicit refresh did not rotate refresh token")
	}
}

func TestBootstrapPostgresConcurrentRequests(t *testing.T) {
	service, db := newBootstrapIntegrationFixture(t)
	ctx := context.Background()

	initial := registerBootstrapIntegrationUser(t, service)
	refreshHash := bootstrapTokenHash(initial.RefreshToken)

	type outcome struct {
		result auth.AuthResult
		err    error
	}

	const workers = 2

	start := make(chan struct{})
	results := make(chan outcome, workers)

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			<-start

			result, err := service.Bootstrap(
				ctx,
				initial.RefreshToken,
			)

			results <- outcome{
				result: result,
				err:    err,
			}
		}()
	}

	close(start)
	wg.Wait()
	close(results)

	var csrfTokens []string

	for outcome := range results {
		if outcome.err != nil {
			t.Fatalf(
				"concurrent bootstrap failed: %v",
				outcome.err,
			)
		}

		if outcome.result.User.ID != initial.User.ID {
			t.Fatal("concurrent bootstrap restored wrong user")
		}

		if outcome.result.RefreshToken != "" {
			t.Fatal("concurrent bootstrap rotated refresh token")
		}

		csrfTokens = append(
			csrfTokens,
			outcome.result.Session.CSRFToken,
		)
	}

	if len(csrfTokens) != workers {
		t.Fatalf(
			"expected %d results, got %d",
			workers,
			len(csrfTokens),
		)
	}

	if csrfTokens[0] == csrfTokens[1] {
		t.Fatal("concurrent bootstrap returned duplicate CSRF")
	}

	var storedCSRF string
	var state string

	err := db.QueryRowContext(ctx, `
		SELECT csrf_token_hash, session_state
		FROM identity.sessions
		WHERE refresh_token_hash = $1
	`, refreshHash).Scan(&storedCSRF, &state)
	if err != nil {
		t.Fatalf("read concurrent session: %v", err)
	}

	if state != "active" {
		t.Fatalf(
			"concurrent bootstrap changed state to %q",
			state,
		)
	}

	firstMatches := storedCSRF ==
		bootstrapTokenHash(csrfTokens[0])

	secondMatches := storedCSRF ==
		bootstrapTokenHash(csrfTokens[1])

	if firstMatches == secondMatches {
		t.Fatal(
			"database must contain exactly one returned CSRF hash",
		)
	}

	currentCSRF := csrfTokens[0]
	if secondMatches {
		currentCSRF = csrfTokens[1]
	}

	rotated, err := service.Refresh(
		ctx,
		initial.RefreshToken,
		currentCSRF,
	)
	if err != nil {
		t.Fatalf(
			"refresh after concurrent bootstrap: %v",
			err,
		)
	}

	if rotated.RefreshToken == "" {
		t.Fatal("refresh did not issue a successor token")
	}
}

func TestBootstrapPostgresRejectsRevokedSession(t *testing.T) {
	service, _ := newBootstrapIntegrationFixture(t)
	ctx := context.Background()

	initial := registerBootstrapIntegrationUser(t, service)

	revoked, err := service.Logout(
		ctx,
		initial.RefreshToken,
		initial.Session.CSRFToken,
		false,
	)
	if err != nil {
		t.Fatalf("logout: %v", err)
	}
	if !revoked {
		t.Fatal("logout did not revoke session")
	}

	_, err = service.Bootstrap(
		ctx,
		initial.RefreshToken,
	)
	if !errors.Is(err, auth.ErrInvalidSession) {
		t.Fatalf(
			"revoked session was restored: %v",
			err,
		)
	}
}

func TestBootstrapPostgresRejectsExpiredSession(t *testing.T) {
	service, db := newBootstrapIntegrationFixture(t)
	ctx := context.Background()

	initial := registerBootstrapIntegrationUser(t, service)
	refreshHash := bootstrapTokenHash(initial.RefreshToken)

	// Modify only the uniquely generated test session.
	result, err := db.ExecContext(ctx, `
		UPDATE identity.sessions
		SET expires_at = created_at + INTERVAL '1 microsecond'
		WHERE refresh_token_hash = $1
	`, refreshHash)
	if err != nil {
		t.Fatalf("expire test session: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		t.Fatalf("read affected rows: %v", err)
	}
	if affected != 1 {
		t.Fatalf(
			"expected one expired test session, got %d",
			affected,
		)
	}

	_, err = service.Bootstrap(
		ctx,
		initial.RefreshToken,
	)
	if !errors.Is(err, auth.ErrInvalidSession) {
		t.Fatalf(
			"expired session was restored: %v",
			err,
		)
	}
}

func TestBootstrapPostgresLogoutWithNewCSRF(t *testing.T) {
	service, _ := newBootstrapIntegrationFixture(t)
	ctx := context.Background()

	initial := registerBootstrapIntegrationUser(t, service)

	restored, err := service.Bootstrap(
		ctx,
		initial.RefreshToken,
	)
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	if restored.Session.CSRFToken == "" {
		t.Fatal("bootstrap returned empty CSRF token")
	}
	if restored.Session.CSRFToken == initial.Session.CSRFToken {
		t.Fatal("bootstrap did not replace CSRF token")
	}

	revoked, err := service.Logout(
		ctx,
		initial.RefreshToken,
		restored.Session.CSRFToken,
		false,
	)
	if err != nil {
		t.Fatalf("logout with new CSRF: %v", err)
	}
	if !revoked {
		t.Fatal("logout did not revoke session")
	}

	_, err = service.Bootstrap(ctx, initial.RefreshToken)
	if !errors.Is(err, auth.ErrInvalidSession) {
		t.Fatalf("revoked session was restored: %v", err)
	}
}

// An unknown token hash must not restore a session or produce credentials.
// Mutating the opaque token preserves its length and URL-safe alphabet.
func TestBootstrapPostgresRejectsUnknownRefreshToken(t *testing.T) {
	service, db := newBootstrapIntegrationFixture(t)
	ctx := context.Background()

	initial := registerBootstrapIntegrationUser(t, service)
	unknown := []byte(initial.RefreshToken)
	if len(unknown) == 0 {
		t.Fatal("registration returned an empty refresh token")
	}
	if unknown[0] == 'a' {
		unknown[0] = 'b'
	} else {
		unknown[0] = 'a'
	}
	unknownToken := string(unknown)

	var hashExists bool
	err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM identity.sessions
			WHERE refresh_token_hash = $1
		)
	`, bootstrapTokenHash(unknownToken)).Scan(&hashExists)
	if err != nil {
		t.Fatalf("check unknown token hash: %v", err)
	}
	if hashExists {
		t.Fatal("generated test token unexpectedly exists in the database")
	}

	result, err := service.Bootstrap(ctx, unknownToken)
	if !errors.Is(err, auth.ErrInvalidSession) {
		t.Fatalf("unknown refresh token was not rejected as invalid session: %v", err)
	}
	if result.Session.AccessToken != "" ||
		result.Session.CSRFToken != "" ||
		result.RefreshToken != "" {
		t.Fatal("rejected Bootstrap returned credential material")
	}
}

// After an explicit Refresh, the previous refresh cookie is no longer valid
// for Bootstrap and must not be restored as an active session.
func TestBootstrapPostgresRejectsRotatedRefreshToken(t *testing.T) {
	service, db := newBootstrapIntegrationFixture(t)
	ctx := context.Background()

	initial := registerBootstrapIntegrationUser(t, service)
	oldRefreshHash := bootstrapTokenHash(initial.RefreshToken)

	rotated, err := service.Refresh(
		ctx,
		initial.RefreshToken,
		initial.Session.CSRFToken,
	)
	if err != nil {
		t.Fatalf("initial explicit refresh: %v", err)
	}
	if rotated.RefreshToken == "" || rotated.RefreshToken == initial.RefreshToken {
		t.Fatal("explicit Refresh did not replace the refresh token")
	}

	result, err := service.Bootstrap(ctx, initial.RefreshToken)
	if !errors.Is(err, auth.ErrInvalidSession) {
		t.Fatalf("rotated refresh token was accepted by Bootstrap: %v", err)
	}
	if result.Session.AccessToken != "" ||
		result.Session.CSRFToken != "" ||
		result.RefreshToken != "" {
		t.Fatal("rejected Bootstrap returned credential material")
	}

	var oldTokenActive bool
	err = db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM identity.sessions
			WHERE refresh_token_hash = $1 AND session_state = 'active'
		)
	`, oldRefreshHash).Scan(&oldTokenActive)
	if err != nil {
		t.Fatalf("check old session state: %v", err)
	}
	if oldTokenActive {
		t.Fatal("Bootstrap resurrected a rotated refresh session")
	}
}
