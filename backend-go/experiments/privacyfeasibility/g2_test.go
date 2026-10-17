// Package privacyfeasibility contains synthetic architecture experiments only.
// Nothing in the application imports this package or applies this policy.
package privacyfeasibility

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestG2SyntheticPostgresLifecyclePolicy(t *testing.T) {
	raw := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if raw == "" {
		if os.Getenv("GITHUB_ACTIONS") == "true" {
			t.Fatal("G2_POSTGRES=NOT_EXECUTED: CI must supply its disposable PostgreSQL service URL")
		}
		t.Skip("G2_POSTGRES=NOT_EXECUTED: no disposable CI PostgreSQL URL")
	}
	if os.Getenv("GITHUB_ACTIONS") != "true" && os.Getenv("OPENINVEST_PRIVACY_FEASIBILITY_DISPOSABLE") != "1" {
		t.Fatal("set OPENINVEST_PRIVACY_FEASIBILITY_DISPOSABLE=1 only for an explicitly disposable localhost cluster")
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Hostname() != "127.0.0.1" && u.Hostname() != "localhost") {
		t.Fatal("experiment requires explicit loopback disposable PostgreSQL")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()
	admin, err := sql.Open("pgx", raw)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	suffix := strings.ReplaceAll(uuid.NewString(), "-", "")
	dbname, role := "oi_privacy_feas_"+suffix, "oi_pf_"+suffix
	_, err = admin.ExecContext(ctx, `CREATE DATABASE `+dbname)
	if err != nil {
		t.Fatal("create isolated experiment database: ", err)
	}
	defer func() {
		cleanup, stop := context.WithTimeout(context.Background(), 10*time.Second)
		defer stop()
		if _, e := admin.ExecContext(cleanup, `DROP DATABASE `+dbname+` WITH (FORCE)`); e != nil {
			t.Error(e)
		}
		if _, e := admin.ExecContext(cleanup, `DROP ROLE IF EXISTS `+role); e != nil {
			t.Error(e)
		}
	}()
	u.Path = "/" + dbname
	owner, err := sql.Open("pgx", u.String())
	if err != nil {
		t.Fatal(err)
	}
	defer owner.Close()
	var version string
	if err = owner.QueryRowContext(ctx, `SHOW server_version_num`).Scan(&version); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(version, "18") {
		t.Fatalf("PostgreSQL 18 required, got %s", version)
	}
	var fullVersion string
	if err = owner.QueryRowContext(ctx, `SELECT version()`).Scan(&fullVersion); err != nil {
		t.Fatal(err)
	}
	t.Logf("G2_ENVIRONMENT=DISPOSABLE_DATABASE version_num=%s version=%s", version, fullVersion)
	_, err = owner.ExecContext(ctx, fmt.Sprintf(`
CREATE ROLE %s LOGIN PASSWORD 'synthetic-local-only' NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE NOREPLICATION;
REVOKE CREATE ON SCHEMA public FROM PUBLIC;
CREATE SCHEMA pf;
CREATE TABLE pf.manifest(role_name text PRIMARY KEY, epoch int NOT NULL);
INSERT INTO pf.manifest VALUES ('%s', 2);
CREATE TABLE pf.lifecycle(subject int PRIMARY KEY, state text NOT NULL, lock_token int NOT NULL DEFAULT 0 CHECK(lock_token=0));
INSERT INTO pf.lifecycle VALUES(1,'active',0),(2,'active',0);
CREATE TABLE pf.ledger(entry int PRIMARY KEY, subject int NOT NULL REFERENCES pf.lifecycle(subject), amount numeric(28,8) NOT NULL);
INSERT INTO pf.ledger VALUES(1,1,10.00000000);
CREATE FUNCTION pf.permit(s int) RETURNS boolean LANGUAGE plpgsql VOLATILE SECURITY INVOKER SET search_path=pg_catalog,pf AS $body$
DECLARE current_state text;
BEGIN
 IF NOT EXISTS(SELECT 1 FROM pf.manifest WHERE role_name=session_user AND epoch=2) THEN RETURN false; END IF;
 SELECT state INTO current_state FROM pf.lifecycle WHERE subject=s FOR SHARE;
 RETURN COALESCE(current_state='active',false);
END $body$;
REVOKE ALL ON FUNCTION pf.permit(int) FROM PUBLIC;
ALTER TABLE pf.ledger ENABLE ROW LEVEL SECURITY;
ALTER TABLE pf.ledger FORCE ROW LEVEL SECURITY;
CREATE POLICY lifecycle ON pf.ledger USING(pf.permit(subject)) WITH CHECK(pf.permit(subject));
GRANT USAGE ON SCHEMA pf TO %s;
GRANT SELECT ON pf.lifecycle,pf.manifest TO %s;
GRANT UPDATE(lock_token) ON pf.lifecycle TO %s;
GRANT EXECUTE ON FUNCTION pf.permit(int) TO %s;
GRANT SELECT,INSERT ON pf.ledger TO %s;
`, role, role, role, role, role, role, role))
	if err != nil {
		t.Fatal(err)
	}
	u.User = url.UserPassword(role, "synthetic-local-only")
	runtime, err := sql.Open("pgx", u.String())
	if err != nil {
		t.Fatal(err)
	}
	defer runtime.Close()
	runtime.SetMaxOpenConns(4)
	count := func(want int) {
		t.Helper()
		var n int
		if e := runtime.QueryRowContext(ctx, `SELECT count(*) FROM pf.ledger`).Scan(&n); e != nil || n != want {
			t.Fatalf("rows=%d want=%d err=%v", n, want, e)
		}
	}
	denied := func(query, code string) {
		t.Helper()
		_, e := runtime.ExecContext(ctx, query)
		var pe *pgconn.PgError
		if !errors.As(e, &pe) || pe.Code != code {
			t.Fatalf("expected SQLSTATE %s, got %v", code, e)
		}
	}
	count(1)
	denied(`UPDATE pf.lifecycle SET state='active' WHERE subject=1`, "42501")
	denied(`UPDATE pf.ledger SET amount=0`, "42501")
	denied(`DELETE FROM pf.ledger`, "42501")
	denied(`TRUNCATE pf.ledger`, "42501")
	// Ordinary SELECT holds a lifecycle lock until its transaction ends.
	reader, e := runtime.BeginTx(ctx, nil)
	if e != nil {
		t.Fatal(e)
	}
	defer reader.Rollback()
	var amount string
	if e = reader.QueryRowContext(ctx, `SELECT amount::text FROM pf.ledger WHERE entry=1`).Scan(&amount); e != nil {
		t.Fatal(e)
	}
	if amount != "10.00000000" {
		t.Fatal("financial amount changed: ", amount)
	}
	transition, e := owner.BeginTx(ctx, nil)
	if e != nil {
		t.Fatal(e)
	}
	defer transition.Rollback()
	if _, e = transition.ExecContext(ctx, `SET LOCAL lock_timeout='150ms'`); e != nil {
		t.Fatal(e)
	}
	_, e = transition.ExecContext(ctx, `UPDATE pf.lifecycle SET state='pending' WHERE subject=1`)
	var pe *pgconn.PgError
	if !errors.As(e, &pe) || pe.Code != "55P03" {
		t.Fatalf("expected reader/transition lock conflict, got %v", e)
	}
	transition.Rollback()
	reader.Rollback()
	// A repeatable-read snapshot cannot use an earlier active lifecycle version.
	stale, e := runtime.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if e != nil {
		t.Fatal(e)
	}
	defer stale.Rollback()
	var state string
	if e = stale.QueryRowContext(ctx, `SELECT state FROM pf.lifecycle WHERE subject=1`).Scan(&state); e != nil {
		t.Fatal(e)
	}
	if _, e = owner.ExecContext(ctx, `UPDATE pf.lifecycle SET state='pending' WHERE subject=1`); e != nil {
		t.Fatal(e)
	}
	e = stale.QueryRowContext(ctx, `SELECT amount::text FROM pf.ledger WHERE entry=1`).Scan(&amount)
	if !errors.As(e, &pe) || pe.Code != "40001" {
		t.Fatalf("expected stale snapshot serialization failure, got %v", e)
	}
	stale.Rollback()
	count(0)
	denied(`INSERT INTO pf.ledger VALUES(2,1,1.00000000)`, "42501")
	if _, e = runtime.ExecContext(ctx, `INSERT INTO pf.ledger VALUES(3,2,2.00000000)`); e != nil {
		t.Fatal("unrelated active subject: ", e)
	}
	count(1)
	// A pool is insufficient evidence of revocation on one backend: pin it explicitly.
	pinned, e := runtime.Conn(ctx)
	if e != nil {
		t.Fatal(e)
	}
	defer pinned.Close()
	pid := func(q interface {
		QueryRowContext(context.Context, string, ...any) *sql.Row
	}) int {
		t.Helper()
		var value int
		if e := q.QueryRowContext(ctx, `SELECT pg_backend_pid()`).Scan(&value); e != nil {
			t.Fatal(e)
		}
		return value
	}
	revoker, e := owner.Conn(ctx)
	if e != nil {
		t.Fatal(e)
	}
	defer revoker.Close()
	beforePID := pid(pinned)
	var ownerPID int
	if e = revoker.QueryRowContext(ctx, `SELECT pg_backend_pid()`).Scan(&ownerPID); e != nil {
		t.Fatal(e)
	}
	if ownerPID == beforePID {
		t.Fatal("revocation must use a different privileged backend")
	}
	var n int
	if e = pinned.QueryRowContext(ctx, `SELECT count(*) FROM pf.ledger WHERE subject=2`).Scan(&n); e != nil || n != 1 {
		t.Fatalf("pinned authorized read rows=%d err=%v", n, e)
	}
	if pid(pinned) != beforePID {
		t.Fatal("backend changed before revocation")
	}
	if _, e = revoker.ExecContext(ctx, `DELETE FROM pf.manifest`); e != nil {
		t.Fatal(e)
	}
	if e = pinned.QueryRowContext(ctx, `SELECT count(*) FROM pf.ledger WHERE subject=2`).Scan(&n); e != nil || n != 0 {
		t.Fatalf("pinned revoked read rows=%d err=%v", n, e)
	}
	if pid(pinned) != beforePID {
		t.Fatal("backend changed after revoked SELECT")
	}
	_, e = pinned.ExecContext(ctx, `INSERT INTO pf.ledger VALUES(4,2,1.00000000)`)
	if !errors.As(e, &pe) || pe.Code != "42501" {
		t.Fatalf("pinned revoked INSERT: %v", e)
	}
	if pid(pinned) != beforePID {
		t.Fatal("backend changed after revoked INSERT")
	}
	t.Logf("EPOCH_REVOCATION_EXISTING_SESSION=PASS SCOPE=AUTOCOMMIT PINNED_PID_BEFORE=%d PINNED_PID_AFTER=%d PRIVILEGED_PID=%d SELECT_ROWS=0 INSERT_SQLSTATE=42501", beforePID, pid(pinned), ownerPID)

	// This is an observation of snapshot semantics, not a weakened universal-revocation oracle.
	// Successful stale-epoch access is an explicit unresolved architecture counterexample.
	for index, scenario := range []struct {
		name      string
		isolation sql.IsolationLevel
	}{
		{"READ_COMMITTED", sql.LevelReadCommitted},
		{"REPEATABLE_READ", sql.LevelRepeatableRead},
		{"SERIALIZABLE", sql.LevelSerializable},
	} {
		t.Run("epoch_transaction_"+scenario.name, func(t *testing.T) {
			if _, e := revoker.ExecContext(ctx, `INSERT INTO pf.manifest VALUES($1,2)`, role); e != nil {
				t.Fatal(e)
			}
			defer func() {
				if _, e := revoker.ExecContext(ctx, `DELETE FROM pf.manifest`); e != nil {
					t.Error(e)
				}
			}()
			tx, e := pinned.BeginTx(ctx, &sql.TxOptions{Isolation: scenario.isolation})
			if e != nil {
				t.Fatal(e)
			}
			defer tx.Rollback()
			if _, e = tx.ExecContext(ctx, `SET LOCAL lock_timeout='250ms'; SET LOCAL statement_timeout='3000ms'`); e != nil {
				t.Fatal(e)
			}
			var actualIsolation string
			if e = tx.QueryRowContext(ctx, `SHOW transaction_isolation`).Scan(&actualIsolation); e != nil {
				t.Fatal(e)
			}
			if pid(tx) != beforePID {
				t.Fatal("transaction backend changed")
			}
			var count int
			if e = tx.QueryRowContext(ctx, `SELECT count(*) FROM pf.ledger WHERE subject=2`).Scan(&count); e != nil || count != 1 {
				t.Fatalf("pre-revocation rows=%d err=%v", count, e)
			}
			if _, e = revoker.ExecContext(ctx, `DELETE FROM pf.manifest`); e != nil {
				t.Fatal(e)
			}
			selectErr := tx.QueryRowContext(ctx, `SELECT count(*) FROM pf.ledger WHERE subject=2`).Scan(&count)
			selectState, insertState := "00000", "NOT_ATTEMPTED_AFTER_TRANSACTION_ABORT"
			if selectErr != nil {
				count = -1 // no post-revocation row observation when the statement failed
				var pgerr *pgconn.PgError
				if !errors.As(selectErr, &pgerr) {
					t.Fatal(selectErr)
				}
				selectState = pgerr.Code
				if selectState != "40001" && selectState != "55P03" && selectState != "57014" {
					t.Fatal(selectErr)
				}
			} else {
				if count != 0 && count != 1 {
					t.Fatalf("unexpected row count %d", count)
				}
				_, insertErr := tx.ExecContext(ctx, `INSERT INTO pf.ledger VALUES($1,2,1.00000000)`, 10+index)
				insertState = "00000"
				if insertErr != nil {
					var pgerr *pgconn.PgError
					if !errors.As(insertErr, &pgerr) {
						t.Fatal(insertErr)
					}
					insertState = pgerr.Code
					if insertState != "42501" && insertState != "40001" && insertState != "55P03" && insertState != "57014" {
						t.Fatal(insertErr)
					}
				}
			}
			if scenario.isolation == sql.LevelReadCommitted && (selectErr != nil || count != 0 || insertState != "42501") {
				t.Fatalf("Read Committed must deny after revoke: rows=%d SELECT=%s INSERT=%s", count, selectState, insertState)
			}
			property := "NOT_PROVEN"
			if selectErr == nil && count == 0 && insertState == "42501" {
				property = "PASS_TESTED_TRANSACTION"
			}
			if (selectErr == nil && count == 1) || insertState == "00000" {
				property = "BLOCKED_STALE_SNAPSHOT_ACCESS"
			}
			// Roll back all synthetic writes; success here is still access admitted before rollback.
			tx.Rollback()
			if pid(pinned) != beforePID {
				t.Fatal("backend changed after transaction")
			}
			t.Logf("EPOCH_TRANSACTION isolation=%s pid=%d SELECT_ROWS=%d SELECT_SQLSTATE=%s INSERT_SQLSTATE=%s PROPERTY=%s INSERTS_ROLLED_BACK=YES", actualIsolation, beforePID, count, selectState, insertState, property)
		})
	}
	var secdef bool
	if e = owner.QueryRowContext(ctx, `SELECT prosecdef FROM pg_proc WHERE oid='pf.permit(integer)'::regprocedure`).Scan(&secdef); e != nil || secdef {
		t.Fatal("helper must be invoker", e)
	}
	t.Logf("G2_SYNTHETIC_POSTGRES=PASS version=%s LOCK_CONFLICT=55P03 STALE_SNAPSHOT=40001 APPEND_ONLY=PASS UNIVERSAL_TRANSACTION_EPOCH_REVOCATION=NOT_PROVEN", version)
	t.Log("G2_SCOPE=ONE_SYNTHETIC_LEDGER_ONLY ACTUAL_APPLICATION_OLD_BINARY_WORKER_PROVIDER=NOT_EXECUTED")
}
