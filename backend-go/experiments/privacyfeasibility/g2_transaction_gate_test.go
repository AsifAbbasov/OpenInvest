package privacyfeasibility

// These fixtures are isolated architecture experiments, not application policy or a release authority.
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
)

type gateLab struct {
	t                          *testing.T
	ctx                        context.Context
	admin, owner, old, fresh   *sql.DB
	database, oldRole, newRole string
	ownerURL, oldURL           string
}

func newGateLab(t *testing.T) *gateLab {
	t.Helper()
	u, e := url.Parse(os.Getenv("OPENINVEST_DATABASE_TEST_URL"))
	if e != nil || (u.Hostname() != "127.0.0.1" && u.Hostname() != "localhost") {
		t.Fatal("explicit disposable loopback PostgreSQL required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	t.Cleanup(cancel)
	admin, e := sql.Open("pgx", u.String())
	if e != nil {
		t.Fatal(e)
	}
	token := strings.ReplaceAll(uuid.NewString(), "-", "")
	f := &gateLab{t: t, ctx: ctx, admin: admin, database: "oi_gate_" + token, oldRole: "oi_old_" + token, newRole: "oi_new_" + token}
	if _, e = admin.ExecContext(ctx, `CREATE DATABASE `+f.database); e != nil {
		admin.Close()
		t.Fatal(e)
	}
	t.Cleanup(func() {
		cleanup, stop := context.WithTimeout(context.Background(), 10*time.Second)
		defer stop()
		if f.old != nil {
			f.old.Close()
		}
		if f.fresh != nil {
			f.fresh.Close()
		}
		if f.owner != nil {
			rows, e := f.owner.QueryContext(cleanup, `SELECT gid FROM pg_prepared_xacts WHERE database=$1`, f.database)
			if e == nil {
				var gids []string
				for rows.Next() {
					var gid string
					if e = rows.Scan(&gid); e != nil {
						t.Error(e)
						break
					}
					gids = append(gids, gid)
				}
				rows.Close()
				for _, gid := range gids {
					if !strings.HasPrefix(gid, "oi_gate_") {
						t.Error("unexpected prepared owner")
						continue
					}
					if _, e = f.owner.ExecContext(cleanup, `ROLLBACK PREPARED '`+gid+`'`); e != nil {
						t.Error(e)
					}
				}
			} else {
				t.Error(e)
			}
			f.owner.Close()
		}
		if _, e := admin.ExecContext(cleanup, `DROP DATABASE `+f.database+` WITH (FORCE)`); e != nil {
			t.Error(e)
		}
		for _, r := range []string{f.oldRole, f.newRole} {
			if _, e := admin.ExecContext(cleanup, `DROP ROLE IF EXISTS `+r); e != nil {
				t.Error(e)
			}
		}
		admin.Close()
	})
	u.Path = "/" + f.database
	f.ownerURL = u.String()
	f.owner, e = sql.Open("pgx", f.ownerURL)
	if e != nil {
		t.Fatal(e)
	}
	f.owner.SetMaxOpenConns(4)
	var version string
	var prepared int
	if e = f.owner.QueryRowContext(ctx, `SELECT version(),current_setting('max_prepared_transactions')::int`).Scan(&version, &prepared); e != nil {
		t.Fatal(e)
	}
	if !strings.HasPrefix(version, "PostgreSQL 18.") || prepared < 1 {
		t.Fatalf("PostgreSQL 18 with prepared transactions required: %s prepared=%d", version, prepared)
	}
	t.Logf("G2_GATE_ENV database=%s version=%s max_prepared_transactions=%d", f.database, version, prepared)
	f.exec(`CREATE ROLE ` + f.oldRole + ` LOGIN PASSWORD 'synthetic-only' NOSUPERUSER NOBYPASSRLS NOREPLICATION NOCREATEDB NOCREATEROLE`)
	f.exec(`CREATE ROLE ` + f.newRole + ` LOGIN PASSWORD 'synthetic-only' NOSUPERUSER NOBYPASSRLS NOREPLICATION NOCREATEDB NOCREATEROLE`)
	f.exec(`REVOKE CREATE ON SCHEMA public FROM PUBLIC;
 CREATE SCHEMA pf;
 CREATE TABLE pf.epoch(id int PRIMARY KEY,state text NOT NULL,lock_token int NOT NULL DEFAULT 0 CHECK(lock_token=0));
 CREATE TABLE pf.subject(id int PRIMARY KEY,state text NOT NULL,lock_token int NOT NULL DEFAULT 0 CHECK(lock_token=0));
 INSERT INTO pf.epoch VALUES(1,'active',0);INSERT INTO pf.subject VALUES(1,'active',0);
 CREATE TABLE pf.ledger(id int PRIMARY KEY,subject int NOT NULL REFERENCES pf.subject(id),amount numeric(28,8) NOT NULL);
 INSERT INTO pf.ledger VALUES(1,1,10.00000000);
 CREATE TABLE pf.barrier(id int PRIMARY KEY,phase text NOT NULL);INSERT INTO pf.barrier VALUES(1,'FREEZE');
 CREATE TABLE pf.snapshot_anchor(id int PRIMARY KEY);INSERT INTO pf.snapshot_anchor VALUES(1);
 CREATE FUNCTION pf.permit(s int) RETURNS boolean LANGUAGE plpgsql VOLATILE SECURITY INVOKER SET search_path=pg_catalog,pf AS $$
 DECLARE e text;v text;
 BEGIN
  SELECT state INTO e FROM pf.epoch WHERE id=1 FOR SHARE;
  IF e IS DISTINCT FROM 'active' THEN RETURN false;END IF;
  SELECT state INTO v FROM pf.subject WHERE id=s FOR SHARE;
  RETURN COALESCE(v='active',false);
 END $$;
 REVOKE ALL ON FUNCTION pf.permit(int) FROM PUBLIC;
 ALTER TABLE pf.ledger ENABLE ROW LEVEL SECURITY;ALTER TABLE pf.ledger FORCE ROW LEVEL SECURITY;
 CREATE POLICY gate ON pf.ledger USING(pf.permit(subject)) WITH CHECK(pf.permit(subject));`)
	for _, role := range []string{f.oldRole, f.newRole} {
		f.exec(`GRANT USAGE ON SCHEMA pf TO ` + role)
		f.exec(`GRANT SELECT ON pf.epoch,pf.subject,pf.snapshot_anchor TO ` + role)
		f.exec(`GRANT UPDATE(lock_token) ON pf.epoch,pf.subject TO ` + role)
		f.exec(`GRANT SELECT,INSERT ON pf.ledger TO ` + role)
		f.exec(`GRANT EXECUTE ON FUNCTION pf.permit(int) TO ` + role)
	}
	u.User = url.UserPassword(f.oldRole, "synthetic-only")
	f.oldURL = u.String()
	f.old, e = sql.Open("pgx", u.String())
	if e != nil {
		t.Fatal(e)
	}
	f.old.SetMaxOpenConns(4)
	u.User = url.UserPassword(f.newRole, "synthetic-only")
	f.fresh, e = sql.Open("pgx", u.String())
	if e != nil {
		t.Fatal(e)
	}
	f.fresh.SetMaxOpenConns(4)
	return f
}
func (f *gateLab) exec(q string, args ...any) {
	f.t.Helper()
	if _, e := f.owner.ExecContext(f.ctx, q, args...); e != nil {
		f.t.Fatal(e)
	}
}
func (f *gateLab) integer(q string, args ...any) int {
	f.t.Helper()
	var n int
	if e := f.owner.QueryRowContext(f.ctx, q, args...).Scan(&n); e != nil {
		f.t.Fatal(e)
	}
	return n
}
func sqlState(e error) string {
	if e == nil {
		return "00000"
	}
	var p *pgconn.PgError
	if errors.As(e, &p) {
		return p.Code
	}
	return "CONNECTION_ERROR"
}
func requireState(t *testing.T, e error, state string) {
	t.Helper()
	if sqlState(e) != state {
		t.Fatalf("want %s got %s: %v", state, sqlState(e), e)
	}
}
func (f *gateLab) conn() (*sql.Conn, int) {
	f.t.Helper()
	c, e := f.old.Conn(f.ctx)
	if e != nil {
		f.t.Fatal(e)
	}
	var pid int
	if e = c.QueryRowContext(f.ctx, `SELECT pg_backend_pid()`).Scan(&pid); e != nil {
		c.Close()
		f.t.Fatal(e)
	}
	return c, pid
}
func (f *gateLab) freeze() { f.exec(`ALTER ROLE ` + f.oldRole + ` NOLOGIN`) }
func (f *gateLab) sessions() int {
	return f.integer(`SELECT count(*) FROM pg_stat_activity WHERE datname=$1 AND usename IN ($2,$3)`, f.database, f.oldRole, f.newRole)
}
func (f *gateLab) prepared() int {
	return f.integer(`SELECT count(*) FROM pg_prepared_xacts WHERE database=$1`, f.database)
}
func (f *gateLab) revoke() error {
	tx, e := f.owner.BeginTx(f.ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	if _, e = tx.ExecContext(f.ctx, `SET LOCAL lock_timeout='150ms'; SET LOCAL statement_timeout='1000ms'`); e != nil {
		return e
	}
	if _, e = tx.ExecContext(f.ctx, `UPDATE pf.epoch SET state='revoked' WHERE id=1`); e != nil {
		return e
	}
	return tx.Commit()
}
func (f *gateLab) terminate(pid int) {
	f.t.Helper()
	var done bool
	e := f.owner.QueryRowContext(f.ctx, `SELECT pg_terminate_backend(pid,3000) FROM pg_stat_activity WHERE pid=$1 AND datname=$2 AND usename IN ($3,$4)`, pid, f.database, f.oldRole, f.newRole).Scan(&done)
	if e != nil || !done {
		f.t.Fatalf("terminate owned synthetic backend %d: %v", pid, e)
	}
	f.wait(`SELECT count(*) FROM pg_stat_activity WHERE datname=$1 AND pid=$2`, 0, f.database, pid)
}
func (f *gateLab) wait(q string, want int, args ...any) {
	f.t.Helper()
	end := time.Now().Add(3 * time.Second)
	for time.Now().Before(end) {
		if f.integer(q, args...) == want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	f.t.Fatal("bounded observation did not reach expected state")
}
func (f *gateLab) deniedFresh() {
	f.t.Helper()
	var n int
	if e := f.fresh.QueryRowContext(f.ctx, `SELECT count(*) FROM pf.ledger`).Scan(&n); e != nil || n != 0 {
		f.t.Fatalf("post-cut read rows=%d err=%v", n, e)
	}
	_, e := f.fresh.ExecContext(f.ctx, `INSERT INTO pf.ledger VALUES(90,1,1.00000000)`)
	requireState(f.t, e, "42501")
}

func TestG2TransactionGates(t *testing.T) {
	if os.Getenv("OPENINVEST_G2_TRANSACTION_MATRIX") != "1" {
		t.Skip("dedicated disposable-container matrix requires OPENINVEST_G2_TRANSACTION_MATRIX=1")
	}
	// B: mandatory epoch AND subject locking; no application helper call is needed in SELECT/INSERT.
	for _, mode := range []struct {
		name      string
		isolation sql.IsolationLevel
	}{
		{"READ_COMMITTED", sql.LevelReadCommitted}, {"REPEATABLE_READ", sql.LevelRepeatableRead}, {"SERIALIZABLE", sql.LevelSerializable},
	} {
		t.Run("B_existing_transaction_"+mode.name, func(t *testing.T) {
			f := newGateLab(t)
			c, pid := f.conn()
			defer c.Close()
			tx, e := c.BeginTx(f.ctx, &sql.TxOptions{Isolation: mode.isolation})
			if e != nil {
				t.Fatal(e)
			}
			defer tx.Rollback()
			var n int
			if e = tx.QueryRowContext(f.ctx, `SELECT count(*) FROM pf.ledger`).Scan(&n); e != nil || n != 1 {
				t.Fatal(n, e)
			}
			requireState(t, f.revoke(), "55P03") // no revocation ACK while an admitted reader still holds the gate
			if e = tx.QueryRowContext(f.ctx, `SELECT count(*) FROM pf.ledger`).Scan(&n); e != nil || n != 1 {
				t.Fatal(n, e)
			}
			tx.Rollback()
			if e = f.revoke(); e != nil {
				t.Fatal(e)
			}
			f.deniedFresh()
			t.Logf("B_HELD_READER mode=%s pid=%d REVOKE_BEFORE_TX_END=55P03 RELEASE=BLOCKED CUT_AFTER_ROLLBACK=PASS POST_CUT_READ=0 INSERT=42501 PRIOR_READ_NOT_RETRACTED=YES", mode.name, pid)
		})
		t.Run("B_snapshot_without_gate_"+mode.name, func(t *testing.T) {
			f := newGateLab(t)
			c, pid := f.conn()
			defer c.Close()
			tx, e := c.BeginTx(f.ctx, &sql.TxOptions{Isolation: mode.isolation})
			if e != nil {
				t.Fatal(e)
			}
			defer tx.Rollback()
			var n int
			if e = tx.QueryRowContext(f.ctx, `SELECT count(*) FROM pf.snapshot_anchor`).Scan(&n); e != nil {
				t.Fatal(e)
			}
			if e = f.revoke(); e != nil {
				t.Fatal(e)
			}
			e = tx.QueryRowContext(f.ctx, `SELECT count(*) FROM pf.ledger`).Scan(&n)
			if mode.isolation == sql.LevelReadCommitted {
				if e != nil || n != 0 {
					t.Fatal(n, e)
				}
			} else {
				requireState(t, e, "40001")
			}
			t.Logf("B_STALE_UNUSED_GATE mode=%s pid=%d POST_CUT_SELECT_SQLSTATE=%s CONFIDENTIALITY_AFTER_CUT=PASS_TESTED_SCOPE", mode.name, pid, sqlState(e))
		})
		t.Run("A_inflight_insert_"+mode.name, func(t *testing.T) {
			f := newGateLab(t)
			c, pid := f.conn()
			defer c.Close()
			tx, e := c.BeginTx(f.ctx, &sql.TxOptions{Isolation: mode.isolation})
			if e != nil {
				t.Fatal(e)
			}
			defer tx.Rollback()
			if _, e = tx.ExecContext(f.ctx, `INSERT INTO pf.ledger VALUES(50,1,1)`); e != nil {
				t.Fatal(e)
			}
			f.freeze()
			requireState(t, f.revoke(), "55P03")
			f.terminate(pid)
			if f.integer(`SELECT count(*) FROM pf.ledger WHERE id=50`) != 0 {
				t.Fatal("in-flight insert survived termination")
			}
			if e := f.revoke(); e != nil {
				t.Fatal(e)
			}
			if _, e := c.ExecContext(f.ctx, `SELECT 1`); e == nil {
				t.Fatal("terminated connection unexpectedly usable")
			}
			f.deniedFresh()
			t.Logf("A_INSERT mode=%s pid=%d UNCOMMITTED_ROW_AFTER_TERMINATION=0 OLD_CONNECTION_REUSE=DENIED CUT=PASS", mode.name, pid)
		})
	}
	t.Run("B_autocommit_pool_reuse", func(t *testing.T) {
		f := newGateLab(t)
		f.old.SetMaxOpenConns(1)
		f.old.SetMaxIdleConns(1)
		c, pid := f.conn()
		var n int
		if e := c.QueryRowContext(f.ctx, `SELECT count(*) FROM pf.ledger`).Scan(&n); e != nil || n != 1 {
			t.Fatal(n, e)
		}
		c.Close()
		if e := f.revoke(); e != nil {
			t.Fatal(e)
		}
		c, after := f.conn()
		defer c.Close()
		if after != pid {
			t.Fatal("pool did not reuse pinned backend")
		}
		if e := c.QueryRowContext(f.ctx, `SELECT count(*) FROM pf.ledger`).Scan(&n); e != nil || n != 0 {
			t.Fatal(n, e)
		}
		_, e := c.ExecContext(f.ctx, `INSERT INTO pf.ledger VALUES(91,1,1)`)
		requireState(t, e, "42501")
		t.Logf("B_AUTOCOMMIT_POOL pid_before=%d pid_after=%d READ=0 INSERT=42501", pid, after)
	})
	for _, transactional := range []bool{false, true} {
		t.Run(fmt.Sprintf("A_long_select_transactional_%t", transactional), func(t *testing.T) {
			f := newGateLab(t)
			c, pid := f.conn()
			defer c.Close()
			if transactional {
				if _, e := c.ExecContext(f.ctx, `BEGIN`); e != nil {
					t.Fatal(e)
				}
				defer c.ExecContext(f.ctx, `ROLLBACK`)
			}
			done := make(chan error, 1)
			go func() {
				var amount string
				var ignored any
				done <- c.QueryRowContext(f.ctx, `SELECT amount::text,pg_sleep(2) FROM pf.ledger`).Scan(&amount, &ignored)
			}()
			f.wait(`SELECT count(*) FROM pg_stat_activity WHERE datname=$1 AND pid=$2 AND wait_event='PgSleep'`, 1, f.database, pid)
			f.freeze()
			requireState(t, f.revoke(), "55P03")
			var cancelled bool
			if e := f.owner.QueryRowContext(f.ctx, `SELECT pg_cancel_backend(pid) FROM pg_stat_activity WHERE pid=$1 AND datname=$2`, pid, f.database).Scan(&cancelled); e != nil || !cancelled {
				t.Fatal(e)
			}
			select {
			case e := <-done:
				requireState(t, e, "57014")
			case <-f.ctx.Done():
				t.Fatal("cancel observation timed out")
			}
			if f.sessions() == 0 {
				t.Fatal("cancel was incorrectly treated as backend drain")
			}
			if transactional {
				if _, e := c.ExecContext(f.ctx, `ROLLBACK`); e != nil {
					t.Fatal(e)
				}
			}
			f.terminate(pid)
			if e := f.revoke(); e != nil {
				t.Fatal(e)
			}
			f.deniedFresh()
			t.Logf("A_LONG_SELECT transactional=%t pid=%d CANCEL=57014 CANCEL_NOT_SESSION_DRAIN=YES TERMINATION_VERIFIED=YES POST_CUT=DENY", transactional, pid)
		})
	}
	t.Run("A_idle_transaction_cancel_not_sufficient", func(t *testing.T) {
		f := newGateLab(t)
		c, pid := f.conn()
		defer c.Close()
		if _, e := c.ExecContext(f.ctx, `BEGIN`); e != nil {
			t.Fatal(e)
		}
		var n int
		if e := c.QueryRowContext(f.ctx, `SELECT count(*) FROM pf.ledger`).Scan(&n); e != nil {
			t.Fatal(e)
		}
		f.wait(`SELECT count(*) FROM pg_stat_activity WHERE datname=$1 AND pid=$2 AND state='idle in transaction'`, 1, f.database, pid)
		f.freeze()
		var cancelled bool
		if e := f.owner.QueryRowContext(f.ctx, `SELECT pg_cancel_backend(pid) FROM pg_stat_activity WHERE pid=$1 AND datname=$2`, pid, f.database).Scan(&cancelled); e != nil {
			t.Fatal(e)
		}
		requireState(t, f.revoke(), "55P03")
		f.terminate(pid)
		if e := f.revoke(); e != nil {
			t.Fatal(e)
		}
		t.Logf("A_IDLE pid=%d CANCEL_SIGNAL=%t INCOMPLETE_DRAIN_BLOCKS=55P03 TERMINATION_THEN_CUT=PASS", pid, cancelled)
	})
	t.Run("A_prepared_survives_backend_termination", func(t *testing.T) {
		f := newGateLab(t)
		c, pid := f.conn()
		defer c.Close()
		gid := "oi_gate_" + strings.ReplaceAll(uuid.NewString(), "-", "")
		if _, e := c.ExecContext(f.ctx, `BEGIN; INSERT INTO pf.ledger VALUES(51,1,1); PREPARE TRANSACTION '`+gid+`'`); e != nil {
			t.Fatal(e)
		}
		if f.prepared() != 1 {
			t.Fatal("prepared transaction not observed")
		}
		f.freeze()
		f.terminate(pid)
		if f.prepared() != 1 {
			t.Fatal("prepared transaction incorrectly assumed to follow backend lifetime")
		}
		requireState(t, f.revoke(), "55P03")
		f.exec(`ROLLBACK PREPARED '` + gid + `'`)
		if f.prepared() != 0 || f.integer(`SELECT count(*) FROM pf.ledger WHERE id=51`) != 0 {
			t.Fatal("prepared rollback not reconciled")
		}
		if e := f.revoke(); e != nil {
			t.Fatal(e)
		}
		t.Logf("A_PREPARED pid=%d SURVIVES_TERMINATION=YES RELEASE_WITH_PREPARED=BLOCKED EXPLICIT_ROLLBACK=VERIFIED", pid)
	})
	t.Run("C_credentials_are_not_workload_attestation", func(t *testing.T) {
		f := newGateLab(t)
		c, pid := f.conn()
		defer c.Close()
		f.freeze()
		var same int
		if e := c.QueryRowContext(f.ctx, `SELECT pg_backend_pid()`).Scan(&same); e != nil || same != pid {
			t.Fatal("NOLOGIN unexpectedly drained existing backend", e)
		}
		probe, e := sql.Open("pgx", f.oldURL)
		if e != nil {
			t.Fatal(e)
		}
		defer probe.Close()
		e = probe.PingContext(f.ctx)
		requireState(t, e, "28000")
		f.terminate(pid)
		// A label is only metadata: this is an ordinary query under a legitimately admitted new credential.
		fresh, e := f.fresh.Conn(f.ctx)
		if e != nil {
			t.Fatal(e)
		}
		defer fresh.Close()
		if _, e = fresh.ExecContext(f.ctx, `SET application_name='synthetic-legacy-label'`); e != nil {
			t.Fatal(e)
		}
		var n int
		if e = fresh.QueryRowContext(f.ctx, `SELECT count(*) FROM pf.ledger`).Scan(&n); e != nil || n != 1 {
			t.Fatal(n, e)
		}
		t.Logf("C_CREDENTIALS old_pid=%d NEW_OLD_LOGIN=28000 EXISTING_BACKEND_REQUIRED_TERMINATION=YES NEW_CREDENTIAL_NORMAL_SQL=ALLOWED", pid)
		t.Log("C_WORKLOAD_IDENTITY=EXTERNAL_EVIDENCE_REQUIRED ACTUAL_OLD_BINARY_AND_ATTESTATION=NOT_EXECUTED APPLICATION_NAME_NOT_AUTHORITY=YES")
	})

	for _, mode := range []struct {
		name      string
		isolation sql.IsolationLevel
	}{{"READ_COMMITTED", sql.LevelReadCommitted}, {"REPEATABLE_READ", sql.LevelRepeatableRead}, {"SERIALIZABLE", sql.LevelSerializable}} {
		t.Run("C_existing_transaction_"+mode.name, func(t *testing.T) {
			f := newGateLab(t)
			c, pid := f.conn()
			defer c.Close()
			tx, e := c.BeginTx(f.ctx, &sql.TxOptions{Isolation: mode.isolation})
			if e != nil {
				t.Fatal(e)
			}
			defer tx.Rollback()
			var n int
			if e = tx.QueryRowContext(f.ctx, `SELECT count(*) FROM pf.ledger`).Scan(&n); e != nil || n != 1 {
				t.Fatal(n, e)
			}
			f.freeze()
			var after int
			if e = tx.QueryRowContext(f.ctx, `SELECT pg_backend_pid()`).Scan(&after); e != nil || after != pid {
				t.Fatal(after, e)
			}
			probe, e := sql.Open("pgx", f.oldURL)
			if e != nil {
				t.Fatal(e)
			}
			defer probe.Close()
			requireState(t, probe.PingContext(f.ctx), "28000")
			f.terminate(pid)
			if _, e = tx.ExecContext(f.ctx, `SELECT 1`); e == nil {
				t.Fatal("terminated transaction reusable")
			}
			t.Logf("C_VERSIONED_CREDENTIAL mode=%s pid=%d NOLOGIN_NOT_EXISTING_TX_REVOCATION=YES NEW_LEGACY_LOGIN=28000 TERMINATION_CONFIRMED=YES", mode.name, pid)
		})
	}

	t.Run("release_barrier_retry_crash_and_missing_evidence", func(t *testing.T) {
		f := newGateLab(t)
		c, oldPID := f.conn()
		defer c.Close()
		f.freeze()
		// Prefix is persisted, CAS-ordered, and independently re-read. RELEASE intentionally lacks a trusted evidence adapter.
		move := func(from, to string) error {
			phases := []string{"FREEZE", "DRAIN", "VERIFY", "ROTATE_FENCE", "RECONCILE", "RELEASE"}
			valid := false
			for i := 0; i < len(phases)-1; i++ {
				if phases[i] == from && phases[i+1] == to {
					valid = true
				}
			}
			if !valid {
				return errors.New("phase skip")
			}
			if to == "RELEASE" {
				return errors.New("G1/workload/privileged-path evidence missing: RELEASE BLOCKED")
			}
			if to == "VERIFY" && (f.sessions() != 0 || f.prepared() != 0 || f.integer(`SELECT count(*) FROM pg_roles WHERE rolname IN ($1,$2) AND rolcanlogin`, f.oldRole, f.newRole) != 0) {
				return errors.New("incomplete drain")
			}
			tx, e := f.owner.BeginTx(f.ctx, nil)
			if e != nil {
				return e
			}
			defer tx.Rollback()
			var phase string
			if e = tx.QueryRowContext(f.ctx, `SELECT phase FROM pf.barrier WHERE id=1 FOR UPDATE`).Scan(&phase); e != nil {
				return e
			}
			if phase == to {
				return tx.Commit()
			}
			if phase != from {
				return errors.New("stale phase")
			}
			if to == "ROTATE_FENCE" {
				if _, e = tx.ExecContext(f.ctx, `UPDATE pf.epoch SET state='revoked'; REVOKE ALL ON pf.ledger FROM `+f.oldRole); e != nil {
					return e
				}
			}
			if _, e = tx.ExecContext(f.ctx, `UPDATE pf.barrier SET phase=$1 WHERE id=1`, to); e != nil {
				return e
			}
			return tx.Commit()
		}
		if e := move("FREEZE", "VERIFY"); e == nil {
			t.Fatal("phase skip accepted")
		}
		if e := move("FREEZE", "DRAIN"); e != nil {
			t.Fatal(e)
		}
		if e := move("DRAIN", "VERIFY"); e == nil {
			t.Fatal("incomplete drain accepted")
		}
		f.terminate(oldPID)
		if e := move("DRAIN", "VERIFY"); e == nil {
			t.Fatal("incomplete admission freeze accepted")
		}
		f.exec(`ALTER ROLE ` + f.newRole + ` NOLOGIN`)
		controller, e := sql.Open("pgx", f.ownerURL)
		if e != nil {
			t.Fatal(e)
		}
		defer controller.Close()
		var phase string
		if e = controller.QueryRowContext(f.ctx, `SELECT phase FROM pf.barrier`).Scan(&phase); e != nil || phase != "DRAIN" {
			t.Fatal(phase, e)
		}
		if e := move("FREEZE", "DRAIN"); e != nil {
			t.Fatal("idempotent retry", e)
		}
		if e := move("DRAIN", "VERIFY"); e != nil {
			t.Fatal(e)
		}
		failed, e := f.owner.BeginTx(f.ctx, nil)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = failed.ExecContext(f.ctx, `UPDATE pf.epoch SET state='revoked'`); e != nil {
			t.Fatal(e)
		}
		failed.Rollback()
		var state string
		if e = f.owner.QueryRowContext(f.ctx, `SELECT state FROM pf.epoch`).Scan(&state); e != nil || state != "active" {
			t.Fatal("failed cutover did not roll back", e)
		}
		if e = move("VERIFY", "ROTATE_FENCE"); e != nil {
			t.Fatal(e)
		}
		if e = move("ROTATE_FENCE", "RECONCILE"); e != nil {
			t.Fatal(e)
		}
		if e = move("RECONCILE", "RELEASE"); e == nil {
			t.Fatal("fabricated external approval")
		}
		if e = f.owner.QueryRowContext(f.ctx, `SELECT phase FROM pf.barrier`).Scan(&phase); e != nil || phase != "RECONCILE" {
			t.Fatal(phase, e)
		}
		t.Log("RELEASE_BARRIER_PREFIX=PASS PERSISTED_RETRY_AND_DB_ROLLBACK=PASS POSITIVE_RELEASE=NOT_EXECUTED EXTERNAL_EVIDENCE=BLOCKED")
	})
}
