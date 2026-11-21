package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/auth"
)

// These tests depend only on the existing Stage A.2 integration fixture.
// They use fresh disposable accounts in OPENINVEST_DATABASE_TEST_URL.
// They must not run against development or production databases.

func stageA2ActiveSessionCount(t *testing.T, ctx context.Context, db *sql.DB, userID string) int {
	t.Helper()
	var count int
	if err := db.QueryRowContext(ctx, `
        SELECT count(*) FROM identity.sessions
        WHERE user_id = $1 AND session_state = 'active'
    `, userID).Scan(&count); err != nil {
		t.Fatalf("count active sessions: %v", err)
	}
	return count
}

// B — A valid session cannot be assigned to a nonexistent owner. This
// verifies the database constraint rather than manufacturing invalid rows.
func TestBootstrapA2OwnerForeignKeyBlocksOrphanSession(t *testing.T) {
	service, db := newBootstrapIntegrationFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	initial := registerBootstrapIntegrationUser(t, service)
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin isolated owner probe: %v", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
        UPDATE identity.sessions
        SET user_id = $1
        WHERE refresh_token_hash = $2
    `, uuid.NewString(), bootstrapTokenHash(initial.RefreshToken))
	if err == nil {
		t.Fatal("orphan session owner was permitted: missing immediate owner integrity constraint")
	}
	if !stageA2SQLState(err, "23503") {
		t.Fatalf("owner mutation rejected for unexpected reason: %v", err)
	}
	t.Log("B_OWNER_ORPHAN=CONSTRAINT_PREVENTED")
}

// B — Test a missing required relationship only if it can be constructed
// while retaining the unique test session. A constraint-protected case is
// reported as such and does not weaken the production schema.
func TestBootstrapA2MissingAuthoritativeRelations(t *testing.T) {
	for _, tc := range []struct {
		name, table string
	}{
		{name: "privacy_settings", table: "identity.privacy_settings"},
		{name: "investment_link", table: "identity.user_investment_links"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service, db := newBootstrapIntegrationFixture(t)
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			initial := registerBootstrapIntegrationUser(t, service)
			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				t.Fatalf("begin relation probe: %v", err)
			}
			defer tx.Rollback()

			// Table identifiers are compile-time constants, not user input.
			statement := "DELETE FROM " + tc.table + " WHERE user_id = $1"
			affected, err := tx.ExecContext(ctx, statement, initial.User.ID)
			if err != nil {
				if stageA2SQLState(err, "23503") {
					t.Log("B_RELATION_MISSING=CONSTRAINT_PREVENTED")
					return
				}
				t.Fatalf("relation deletion failed unexpectedly: %v", err)
			}
			rows, err := affected.RowsAffected()
			if err != nil || rows != 1 {
				t.Fatalf("expected one isolated user relationship, rows=%d err=%v", rows, err)
			}
			var remaining int
			err = tx.QueryRowContext(ctx, `
                SELECT count(*) FROM identity.sessions WHERE refresh_token_hash = $1
            `, bootstrapTokenHash(initial.RefreshToken)).Scan(&remaining)
			if err != nil {
				t.Fatalf("check session after relation probe: %v", err)
			}
			if remaining == 0 {
				t.Log("B_RELATION_MISSING=SESSION_CASCADE_PROTECTED")
				return // Transaction is rolled back by defer.
			}
			if err := tx.Commit(); err != nil {
				if stageA2SQLState(err, "23503") {
					t.Log("B_RELATION_MISSING=DEFERRED_CONSTRAINT_PREVENTED")
					return
				}
				t.Fatalf("commit isolated broken relation: %v", err)
			}
			got, err := service.Bootstrap(ctx, initial.RefreshToken)
			if !errors.Is(err, auth.ErrInvalidSession) && !errors.Is(err, auth.ErrInvalidCredentials) {
				t.Fatalf("missing %s relation not classified as invalid credential/session: %v", tc.name, err)
			}
			if got.RefreshToken != "" || got.Session.CSRFToken != "" || got.Session.AccessToken != "" {
				t.Fatal("rejected relation returned credential material")
			}
			t.Log("B_RELATION_MISSING=REJECTED_WITHOUT_CREDENTIALS")
		})
	}
}

func stageA2SQLState(err error, code string) bool {
	var state interface{ SQLState() string }
	return errors.As(err, &state) && state.SQLState() == code
}

type stageA2AuthOutcome struct {
	result auth.AuthResult
	err    error
}

// C1 — Execute competing operations at one barrier and assert that one
// durable decision wins. The pre-existing Refresh/Refresh test additionally
// covers full family replay revocation.
func TestBootstrapA2RacingRefresh(t *testing.T) {
	service, db := newBootstrapIntegrationFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	initial := registerBootstrapIntegrationUser(t, service)

	start := make(chan struct{})
	var wg sync.WaitGroup
	var bootstrap, refresh stageA2AuthOutcome
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		bootstrap.result, bootstrap.err = service.Bootstrap(ctx, initial.RefreshToken)
	}()
	go func() {
		defer wg.Done()
		<-start
		refresh.result, refresh.err = service.Refresh(ctx, initial.RefreshToken, initial.Session.CSRFToken)
	}()
	close(start)
	wg.Wait()

	switch {
	case bootstrap.err == nil && errors.Is(refresh.err, auth.ErrInvalidSession):
		if bootstrap.result.RefreshToken != "" || bootstrap.result.Session.CSRFToken == "" {
			t.Fatal("Bootstrap winner returned unexpected credentials")
		}
		t.Log("C_BOOTSTRAP_FIRST=PASS")
	case refresh.err == nil && errors.Is(bootstrap.err, auth.ErrInvalidSession):
		if refresh.result.RefreshToken == "" || refresh.result.RefreshToken == initial.RefreshToken {
			t.Fatal("Refresh winner did not rotate its cookie")
		}
		t.Log("C_REFRESH_FIRST=PASS")
	default:
		t.Fatalf("unexpected race outcomes bootstrap=%v refresh=%v", bootstrap.err, refresh.err)
	}
	if count := stageA2ActiveSessionCount(t, ctx, db, initial.User.ID); count != 1 {
		t.Fatalf("expected exactly one active session after race; got %d", count)
	}
	t.Log("C_BOOTSTRAP_VS_REFRESH_FINAL_STATE=PASS")
}

// C2 — Bootstrap and Logout using the original CSRF cannot both succeed;
// the database must never restore a revoked session.
func TestBootstrapA2RacingLogout(t *testing.T) {
	service, db := newBootstrapIntegrationFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	initial := registerBootstrapIntegrationUser(t, service)

	start := make(chan struct{})
	var wg sync.WaitGroup
	var bootstrap stageA2AuthOutcome
	var revoked bool
	var logoutErr error
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		bootstrap.result, bootstrap.err = service.Bootstrap(ctx, initial.RefreshToken)
	}()
	go func() {
		defer wg.Done()
		<-start
		revoked, logoutErr = service.Logout(ctx, initial.RefreshToken, initial.Session.CSRFToken, false)
	}()
	close(start)
	wg.Wait()

	switch {
	case bootstrap.err == nil && errors.Is(logoutErr, auth.ErrInvalidSession):
		if revoked {
			t.Fatal("rejected stale-CSRF logout incorrectly reported success")
		}
		if count := stageA2ActiveSessionCount(t, ctx, db, initial.User.ID); count != 1 {
			t.Fatalf("Bootstrap winner should leave one active session, got %d", count)
		}
		// Clean up by proving new CSRF can revoke the winner's session.
		success, err := service.Logout(ctx, initial.RefreshToken, bootstrap.result.Session.CSRFToken, false)
		if err != nil || !success {
			t.Fatalf("logout using winning bootstrap CSRF: success=%v err=%v", success, err)
		}
		t.Log("C_BOOTSTRAP_BEFORE_LOGOUT=PASS")
	case errors.Is(bootstrap.err, auth.ErrInvalidSession) && logoutErr == nil && revoked:
		t.Log("C_LOGOUT_BEFORE_BOOTSTRAP=PASS")
	default:
		t.Fatalf("unexpected logout race bootstrap=%v logout=%v revoked=%v", bootstrap.err, logoutErr, revoked)
	}
	if count := stageA2ActiveSessionCount(t, ctx, db, initial.User.ID); count != 0 {
		t.Fatalf("logout race left %d active sessions", count)
	}
	result, err := service.Bootstrap(ctx, initial.RefreshToken)
	if !errors.Is(err, auth.ErrInvalidSession) || result.Session.AccessToken != "" {
		t.Fatalf("revoked session restored after logout race: %v", err)
	}
	t.Log("C_BOOTSTRAP_VS_LOGOUT_FINAL_STATE=PASS")
}

// D — A replayed pre-rotation token must revoke its active successor family
// according to the existing replay contract. Bootstrap must not revive it.
func TestBootstrapA2RefreshReplayStillRevokesDescendant(t *testing.T) {
	service, db := newBootstrapIntegrationFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	initial := registerBootstrapIntegrationUser(t, service)
	rotated, err := service.Refresh(ctx, initial.RefreshToken, initial.Session.CSRFToken)
	if err != nil {
		t.Fatalf("rotate original session: %v", err)
	}
	if rotated.RefreshToken == "" || rotated.RefreshToken == initial.RefreshToken {
		t.Fatal("refresh token did not rotate")
	}
	_, err = service.Refresh(ctx, initial.RefreshToken, initial.Session.CSRFToken)
	if !errors.Is(err, auth.ErrInvalidSession) {
		t.Fatalf("replayed old refresh token was not rejected: %v", err)
	}
	got, err := service.Bootstrap(ctx, rotated.RefreshToken)
	if !errors.Is(err, auth.ErrInvalidSession) || got.Session.AccessToken != "" || got.Session.CSRFToken != "" {
		t.Fatalf("replayed session family was restored: %v", err)
	}
	if count := stageA2ActiveSessionCount(t, ctx, db, initial.User.ID); count != 0 {
		t.Fatalf("replay left active session descendants: %d", count)
	}
	t.Log("D_REPLAY_REVOKES_SUCCESSOR_FAMILY=PASS")
}

// E — Lower-layer evidence only. Intentionally fail a PostgreSQL transaction
// after changing the test session's CSRF hash and verify a rollback restores
// the original hash. This does NOT prove fault handling inside Bootstrap.
func TestBootstrapA2DatabaseRollbackLowerLayer(t *testing.T) {
	service, db := newBootstrapIntegrationFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	initial := registerBootstrapIntegrationUser(t, service)
	before := bootstrapTokenHash(initial.Session.CSRFToken)
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin transaction: %v", err)
	}
	defer tx.Rollback()
	update, err := tx.ExecContext(ctx, `
        UPDATE identity.sessions SET csrf_token_hash = $1
        WHERE refresh_token_hash = $2
    `, bootstrapTokenHash(uuid.NewString()), bootstrapTokenHash(initial.RefreshToken))
	if err != nil {
		t.Fatalf("update isolated CSRF proof: %v", err)
	}
	rows, err := update.RowsAffected()
	if err != nil || rows != 1 {
		t.Fatalf("expected exactly one update: rows=%d err=%v", rows, err)
	}
	var sentinel int
	err = tx.QueryRowContext(ctx, `SELECT 1 / 0`).Scan(&sentinel)
	if !stageA2SQLState(err, "22012") {
		t.Fatalf("expected PostgreSQL division-by-zero transaction fault; got %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback injected fault: %v", err)
	}
	var stored string
	err = db.QueryRowContext(ctx, `
        SELECT csrf_token_hash FROM identity.sessions WHERE refresh_token_hash = $1
    `, bootstrapTokenHash(initial.RefreshToken)).Scan(&stored)
	if err != nil {
		t.Fatalf("read rollback outcome: %v", err)
	}
	if stored != before {
		t.Fatal("partial CSRF mutation survived PostgreSQL transaction rollback")
	}
	t.Log("E_POSTGRESQL_ROLLBACK_LOWER_LAYER=PASS; BOOTSTRAP_FAULT_INJECTION=NOT_VERIFIED")
}

// These four deterministic order probes complement the simultaneous races.
// The channel/return boundary establishes a committed operation before
// starting its competitor. They are order-invariant evidence, not proof of
// both transaction lock schedules under simultaneous contention.
func TestBootstrapA2OrderedBootstrapThenRefresh(t *testing.T) {
	service, db := newBootstrapIntegrationFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	initial := registerBootstrapIntegrationUser(t, service)
	restored, err := service.Bootstrap(ctx, initial.RefreshToken)
	if err != nil {
		t.Fatalf("bootstrap first: %v", err)
	}
	_, err = service.Refresh(ctx, initial.RefreshToken, initial.Session.CSRFToken)
	if !errors.Is(err, auth.ErrInvalidSession) {
		t.Fatalf("pre-bootstrap CSRF accepted after bootstrap: %v", err)
	}
	successor, err := service.Refresh(ctx, initial.RefreshToken, restored.Session.CSRFToken)
	if err != nil || successor.RefreshToken == "" {
		t.Fatalf("refresh with latest CSRF: %v", err)
	}
	if count := stageA2ActiveSessionCount(t, ctx, db, initial.User.ID); count != 1 {
		t.Fatalf("ordered Bootstrap/Refresh produced %d active sessions", count)
	}
	t.Log("C_BOOTSTRAP_COMMITTED_BEFORE_REFRESH=PASS")
}

func TestBootstrapA2OrderedRefreshThenBootstrap(t *testing.T) {
	service, db := newBootstrapIntegrationFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	initial := registerBootstrapIntegrationUser(t, service)
	successor, err := service.Refresh(ctx, initial.RefreshToken, initial.Session.CSRFToken)
	if err != nil {
		t.Fatalf("refresh first: %v", err)
	}
	_, err = service.Bootstrap(ctx, initial.RefreshToken)
	if !errors.Is(err, auth.ErrInvalidSession) {
		t.Fatalf("old refresh token revived after Refresh: %v", err)
	}
	restored, err := service.Bootstrap(ctx, successor.RefreshToken)
	if err != nil || restored.User.ID != initial.User.ID || restored.RefreshToken != "" {
		t.Fatalf("successor Bootstrap invalid after refresh: %v", err)
	}
	if count := stageA2ActiveSessionCount(t, ctx, db, initial.User.ID); count != 1 {
		t.Fatalf("ordered Refresh/Bootstrap produced %d active sessions", count)
	}
	t.Log("C_REFRESH_COMMITTED_BEFORE_BOOTSTRAP=PASS")
}

func TestBootstrapA2OrderedBootstrapThenLogout(t *testing.T) {
	service, db := newBootstrapIntegrationFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	initial := registerBootstrapIntegrationUser(t, service)
	restored, err := service.Bootstrap(ctx, initial.RefreshToken)
	if err != nil {
		t.Fatalf("bootstrap first: %v", err)
	}
	_, err = service.Logout(ctx, initial.RefreshToken, initial.Session.CSRFToken, false)
	if !errors.Is(err, auth.ErrInvalidSession) {
		t.Fatalf("old CSRF logged out post-bootstrap session: %v", err)
	}
	revoked, err := service.Logout(ctx, initial.RefreshToken, restored.Session.CSRFToken, false)
	if err != nil || !revoked {
		t.Fatalf("logout new CSRF: revoked=%v err=%v", revoked, err)
	}
	if count := stageA2ActiveSessionCount(t, ctx, db, initial.User.ID); count != 0 {
		t.Fatalf("Bootstrap/Logout left %d active sessions", count)
	}
	t.Log("C_BOOTSTRAP_COMMITTED_BEFORE_LOGOUT=PASS")
}

func TestBootstrapA2OrderedLogoutThenBootstrap(t *testing.T) {
	service, db := newBootstrapIntegrationFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	initial := registerBootstrapIntegrationUser(t, service)
	revoked, err := service.Logout(ctx, initial.RefreshToken, initial.Session.CSRFToken, false)
	if err != nil || !revoked {
		t.Fatalf("logout first: revoked=%v err=%v", revoked, err)
	}
	got, err := service.Bootstrap(ctx, initial.RefreshToken)
	if !errors.Is(err, auth.ErrInvalidSession) || got.Session.AccessToken != "" {
		t.Fatalf("bootstrap revived logged-out session: %v", err)
	}
	if count := stageA2ActiveSessionCount(t, ctx, db, initial.User.ID); count != 0 {
		t.Fatalf("Logout/Bootstrap left %d active sessions", count)
	}
	t.Log("C_LOGOUT_COMMITTED_BEFORE_BOOTSTRAP=PASS")
}
