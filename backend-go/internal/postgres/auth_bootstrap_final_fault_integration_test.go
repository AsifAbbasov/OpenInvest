package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/openinvest/openinvest/backend-go/internal/auth"
)

// These tests are ONLY for the disposable OpenInvest PostgreSQL instance.
// The runner verifies the endpoint and requires an explicit opt-in before
// installing scoped, removable, test-only fault-injection triggers.
func stageA2RequireFaultOptIn(t *testing.T) {
	t.Helper()
	if os.Getenv("OPENINVEST_A2_TEST_DDL") != "1" {
		t.Skip("TEMP_DDL_NOT_AUTHORIZED: set OPENINVEST_A2_TEST_DDL=1 only for disposable PostgreSQL")
	}
}

func stageA2StoredCSRF(t *testing.T, db *sql.DB, token string) (string, string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var csrf, state string
	err := db.QueryRowContext(ctx, `
        SELECT csrf_token_hash, session_state
        FROM identity.sessions WHERE refresh_token_hash = $1
    `, bootstrapTokenHash(token)).Scan(&csrf, &state)
	if err != nil {
		t.Fatalf("read session state: %v", err)
	}
	return csrf, state
}

func stageA2NoCredentials(t *testing.T, result auth.AuthResult) {
	t.Helper()
	if result.RefreshToken != "" || result.Session.AccessToken != "" || result.Session.CSRFToken != "" {
		t.Fatal("failed Bootstrap returned credential material")
	}
}

// A freshly named trigger guards ONLY the uniquely generated session hash.
// A deferred constraint trigger raises at COMMIT; an ordinary AFTER trigger
// raises after the UPDATE statement but before the user lookup. In both cases
// the production method must roll back the uncommitted CSRF replacement.
func stageA2InstallFaultTrigger(t *testing.T, db *sql.DB, token string, deferred bool) func() {
	t.Helper()
	stageA2RequireFaultOptIn(t)
	suffix := strings.ReplaceAll(uuid.NewString(), "-", "")
	functionName := "oi_a2_fault_function_" + suffix
	triggerName := "oi_a2_fault_trigger_" + suffix
	hash := bootstrapTokenHash(token) // hex produced by SHA-256, never user SQL
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	createFunction := fmt.Sprintf(`
        CREATE FUNCTION public.%s() RETURNS trigger LANGUAGE plpgsql AS $fn$
        BEGIN
            RAISE EXCEPTION 'STAGE_A2_SCOPED_FAULT' USING ERRCODE = 'P0001';
        END
        $fn$`, functionName)
	if _, err := db.ExecContext(ctx, createFunction); err != nil {
		t.Fatalf("install scoped fault function: %v", err)
	}
	cleanup := func() {
		// Cleanup uses a fresh context, including if the test context was cancelled.
		c, stop := context.WithTimeout(context.Background(), 12*time.Second)
		defer stop()
		if _, err := db.ExecContext(c, fmt.Sprintf(`DROP TRIGGER IF EXISTS %s ON identity.sessions`, triggerName)); err != nil {
			t.Errorf("remove test trigger: %v", err)
		}
		if _, err := db.ExecContext(c, fmt.Sprintf(`DROP FUNCTION IF EXISTS public.%s()`, functionName)); err != nil {
			t.Errorf("remove test fault function: %v", err)
		}
	}
	t.Cleanup(cleanup)
	var ddl string
	if deferred {
		ddl = fmt.Sprintf(`CREATE CONSTRAINT TRIGGER %s
            AFTER UPDATE OF csrf_token_hash ON identity.sessions
            DEFERRABLE INITIALLY DEFERRED
            FOR EACH ROW WHEN (NEW.refresh_token_hash = '%s')
            EXECUTE FUNCTION public.%s()`, triggerName, hash, functionName)
	} else {
		ddl = fmt.Sprintf(`CREATE TRIGGER %s
            AFTER UPDATE OF csrf_token_hash ON identity.sessions
            FOR EACH ROW WHEN (NEW.refresh_token_hash = '%s')
            EXECUTE FUNCTION public.%s()`, triggerName, hash, functionName)
	}
	if _, err := db.ExecContext(ctx, ddl); err != nil {
		t.Fatalf("install scoped fault trigger: %v", err)
	}
	return cleanup
}

func stageA2AssertFailedBootstrapRollback(t *testing.T, deferred bool) {
	t.Helper()
	service, db := newBootstrapIntegrationFixture(t)
	initial := registerBootstrapIntegrationUser(t, service)
	before, oldState := stageA2StoredCSRF(t, db, initial.RefreshToken)
	if oldState != "active" {
		t.Fatalf("initial session not active: %q", oldState)
	}
	drop := stageA2InstallFaultTrigger(t, db, initial.RefreshToken, deferred)
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	result, err := service.Bootstrap(ctx, initial.RefreshToken)
	if err == nil {
		t.Fatal("Bootstrap unexpectedly committed with injected transaction fault")
	}
	if !stageA2SQLState(err, "P0001") {
		t.Fatalf("Bootstrap failed for another reason, not scoped injection: %v", err)
	}
	stageA2NoCredentials(t, result)
	after, state := stageA2StoredCSRF(t, db, initial.RefreshToken)
	if after != before || state != "active" {
		t.Fatalf("Bootstrap partially changed session: hashSame=%t state=%q", after == before, state)
	}
	// Verify that removal of the injector allows an ordinary Bootstrap again.
	// The deferred cleanup is retained for failures in earlier assertions.
	drop()
	restored, err := service.Bootstrap(ctx, initial.RefreshToken)
	if err != nil || restored.Session.AccessToken == "" || restored.Session.CSRFToken == "" {
		t.Fatalf("Bootstrap not functional after removing injector: %v", err)
	}
}

func TestBootstrapA2FaultAfterCSRFUpdateRollsBack(t *testing.T) {
	stageA2AssertFailedBootstrapRollback(t, false)
	t.Log("E_BOOTSTRAP_POST_UPDATE_ROLLBACK=DIRECT_PASS")
}

func TestBootstrapA2DeferredCommitFailureRollsBack(t *testing.T) {
	stageA2AssertFailedBootstrapRollback(t, true)
	t.Log("E_BOOTSTRAP_COMMIT_FAILURE_ROLLBACK=DIRECT_PASS")
}

func TestBootstrapA2AuthoritativeLookupFailureRollsBackCSRF(t *testing.T) {
	service, db := newBootstrapIntegrationFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	initial := registerBootstrapIntegrationUser(t, service)
	before, state := stageA2StoredCSRF(t, db, initial.RefreshToken)
	if state != "active" {
		t.Fatalf("initial session state: %q", state)
	}
	// Existing A2 tests prove privacy_settings can be removed on this disposable DB.
	rows, err := db.ExecContext(ctx, `DELETE FROM identity.privacy_settings WHERE user_id = $1`, initial.User.ID)
	if err != nil {
		t.Fatalf("remove disposable privacy relationship: %v", err)
	}
	n, err := rows.RowsAffected()
	if err != nil || n != 1 {
		t.Fatalf("expected one privacy row removed: n=%d err=%v", n, err)
	}
	got, err := service.Bootstrap(ctx, initial.RefreshToken)
	if !errors.Is(err, auth.ErrInvalidSession) {
		t.Fatalf("missing user relationship: unexpected error: %v", err)
	}
	stageA2NoCredentials(t, got)
	after, afterState := stageA2StoredCSRF(t, db, initial.RefreshToken)
	if after != before || afterState != "active" {
		t.Fatalf("authoritative lookup failure did not roll back CSRF: same=%t state=%q", after == before, afterState)
	}
	t.Log("E_AUTH_USER_LOOKUP_ROLLBACK=DIRECT_PASS")
}

// For controlled tests we hold the same per-owner advisory lock as Bootstrap,
// Refresh, and Logout, then observe PostgreSQL waiters via pg_locks.
type stageA2AdvisoryLock struct {
	tx      *sql.Tx
	classID string
	objID   string
	subID   int
}

func stageA2HoldUserLock(t *testing.T, ctx context.Context, db *sql.DB, userID string) *stageA2AdvisoryLock {
	t.Helper()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin lock-holder tx: %v", err)
	}
	t.Cleanup(func() { _ = tx.Rollback() })
	_, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, "openinvest/auth-refresh/"+userID)
	if err != nil {
		t.Fatalf("hold auth advisory lock: %v", err)
	}
	lock := &stageA2AdvisoryLock{tx: tx}
	err = tx.QueryRowContext(ctx, `
        SELECT classid::text, objid::text, objsubid FROM pg_locks
        WHERE pid = pg_backend_pid() AND locktype = 'advisory' AND granted
        LIMIT 1
    `).Scan(&lock.classID, &lock.objID, &lock.subID)
	if err != nil {
		t.Fatalf("read held advisory lock identity: %v", err)
	}
	return lock
}
func stageA2AwaitWaiters(t *testing.T, db *sql.DB, l *stageA2AdvisoryLock, n int) {
	t.Helper()
	deadline := time.Now().Add(9 * time.Second)
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		var waiting int
		err := db.QueryRowContext(ctx, `
            SELECT count(*) FROM pg_locks
            WHERE locktype = 'advisory' AND NOT granted
              AND classid::text = $1 AND objid::text = $2 AND objsubid = $3
        `, l.classID, l.objID, l.subID).Scan(&waiting)
		cancel()
		if err != nil {
			t.Fatalf("read PostgreSQL waiters: %v", err)
		}
		if waiting >= n {
			return
		}
		time.Sleep(25 * time.Millisecond) // Polling only; not used to impose operation order.
	}
	t.Fatalf("expected %d PostgreSQL advisory-lock waiters; not observed", n)
}

func TestBootstrapA2CancelledWhileWaitingDoesNotChangeCSRF(t *testing.T) {
	service, db := newBootstrapIntegrationFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	initial := registerBootstrapIntegrationUser(t, service)
	before, _ := stageA2StoredCSRF(t, db, initial.RefreshToken)
	held := stageA2HoldUserLock(t, ctx, db, initial.User.ID)
	operationCtx, abort := context.WithCancel(ctx)
	done := make(chan error, 1)
	go func() {
		got, err := service.Bootstrap(operationCtx, initial.RefreshToken)
		if got.Session.CSRFToken != "" || got.Session.AccessToken != "" || got.RefreshToken != "" {
			done <- errors.New("cancelled Bootstrap returned credentials")
			return
		}
		done <- err
	}()
	stageA2AwaitWaiters(t, db, held, 1)
	abort()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("cancelled Bootstrap unexpectedly succeeded")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Bootstrap did not stop after context cancellation")
	}
	if err := held.tx.Rollback(); err != nil {
		t.Fatalf("release held advisory lock: %v", err)
	}
	after, state := stageA2StoredCSRF(t, db, initial.RefreshToken)
	if before != after || state != "active" {
		t.Fatalf("cancelled Bootstrap changed session: hashSame=%t state=%q", before == after, state)
	}
	t.Log("E_POSTGRES_LOCK_WAIT_CANCELLATION_ROLLBACK=DIRECT_PASS; PROCESS_INTERRUPTION=NOT_RUN")
}

// This test verifies acquisition order under controlled PostgreSQL contention.
// Both operations are observed waiting on the SAME advisory lock before release.
func TestBootstrapA2ControlledAdvisoryLockOrdering(t *testing.T) {
	for _, tc := range []struct{ name, first, second string }{
		{"bootstrap_first_refresh", "bootstrap", "refresh"},
		{"refresh_first_bootstrap", "refresh", "bootstrap"},
		{"bootstrap_first_logout", "bootstrap", "logout"},
		{"logout_first_bootstrap", "logout", "bootstrap"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service, db := newBootstrapIntegrationFixture(t)
			ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
			defer cancel()
			initial := registerBootstrapIntegrationUser(t, service)
			held := stageA2HoldUserLock(t, ctx, db, initial.User.ID)
			type outcome struct {
				name    string
				result  auth.AuthResult
				revoked bool
				err     error
			}
			results := make(chan outcome, 2)
			launch := func(name string) {
				go func() {
					var r outcome
					r.name = name
					switch name {
					case "bootstrap":
						r.result, r.err = service.Bootstrap(ctx, initial.RefreshToken)
					case "refresh":
						r.result, r.err = service.Refresh(ctx, initial.RefreshToken, initial.Session.CSRFToken)
					case "logout":
						r.revoked, r.err = service.Logout(ctx, initial.RefreshToken, initial.Session.CSRFToken, false)
					}
					results <- r
				}()
			}
			launch(tc.first)
			stageA2AwaitWaiters(t, db, held, 1)
			launch(tc.second)
			stageA2AwaitWaiters(t, db, held, 2)
			if err := held.tx.Commit(); err != nil {
				t.Fatalf("release scheduling barrier: %v", err)
			}
			var a, b outcome
			select {
			case a = <-results:
			case <-ctx.Done():
				t.Fatalf("first operation timeout: %v", ctx.Err())
			}
			select {
			case b = <-results:
			case <-ctx.Done():
				t.Fatalf("second operation timeout: %v", ctx.Err())
			}
			if a.name != tc.first || b.name != tc.second {
				t.Fatalf("unexpected PostgreSQL lock order: %s then %s; wanted %s then %s", a.name, b.name, tc.first, tc.second)
			}
			if a.err != nil {
				t.Fatalf("first operation (%s) failed: %v", a.name, a.err)
			}
			if a.name == "logout" && !a.revoked {
				t.Fatal("first logout did not revoke session")
			}
			if a.name == "bootstrap" && a.result.Session.CSRFToken == "" {
				t.Fatal("first bootstrap returned no CSRF")
			}
			if a.name == "refresh" && (a.result.RefreshToken == "" || a.result.RefreshToken == initial.RefreshToken) {
				t.Fatal("first refresh did not rotate token")
			}
			if !errors.Is(b.err, auth.ErrInvalidSession) {
				t.Fatalf("second operation (%s) should reject stale session/CSRF: %v", b.name, b.err)
			}
			stageA2NoCredentials(t, b.result)
			if tc.first == "logout" {
				if _, err := service.Bootstrap(ctx, initial.RefreshToken); !errors.Is(err, auth.ErrInvalidSession) {
					t.Fatalf("revoked session restored after logout: %v", err)
				}
			}
			t.Log("C_CONTROLLED_LOCK_ORDER_DIRECT_PASS: " + tc.first + " before " + tc.second)
		})
	}
}
