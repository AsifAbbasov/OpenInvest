package postgres

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestRuntimeSessionUsesBoundedDatabaseTimeouts(t *testing.T) {
	ownerURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	runtimeURL := os.Getenv("OPENINVEST_DATABASE_RUNTIME_TEST_URL")
	if ownerURL == "" || runtimeURL == "" {
		t.Skip("runtime timeout integration URLs are not set")
	}

	ownerDB, err := sql.Open("pgx", ownerURL)
	if err != nil {
		t.Fatalf("open owner database: %v", err)
	}
	t.Cleanup(func() { _ = ownerDB.Close() })
	var ownerStatementTimeout string
	var ownerLockTimeout string
	var ownerIdleTimeout string
	if err := ownerDB.QueryRow(`
		SELECT current_setting('statement_timeout'), current_setting('lock_timeout'), current_setting('idle_in_transaction_session_timeout')
	`).Scan(&ownerStatementTimeout, &ownerLockTimeout, &ownerIdleTimeout); err != nil {
		t.Fatalf("read owner timeouts: %v", err)
	}
	if ownerStatementTimeout != "0" || ownerLockTimeout != "0" || ownerIdleTimeout != "0" {
		t.Fatalf("owner session acquired runtime timeout envelope: statement=%q lock=%q idle=%q", ownerStatementTimeout, ownerLockTimeout, ownerIdleTimeout)
	}

	parsedRuntimeURL, err := url.Parse(runtimeURL)
	if err != nil {
		t.Fatalf("parse runtime database URL: %v", err)
	}
	query := parsedRuntimeURL.Query()
	query.Set("statement_timeout", "0")
	query.Set("lock_timeout", "0")
	query.Set("idle_in_transaction_session_timeout", "0")
	parsedRuntimeURL.RawQuery = query.Encode()

	runtime, err := OpenRuntime(parsedRuntimeURL.String())
	if err != nil {
		t.Fatalf("open bounded runtime session: %v", err)
	}
	t.Cleanup(func() { _ = runtime.Close() })

	ctx := context.Background()
	for i := 0; i < 3; i++ {
		conn, err := runtime.db.Conn(ctx)
		if err != nil {
			t.Fatalf("acquire runtime connection %d: %v", i, err)
		}
		var statementTimeout string
		var lockTimeout string
		var idleTimeout string
		err = conn.QueryRowContext(ctx, `
			SELECT current_setting('statement_timeout'), current_setting('lock_timeout'), current_setting('idle_in_transaction_session_timeout')
		`).Scan(&statementTimeout, &lockTimeout, &idleTimeout)
		_ = conn.Close()
		if err != nil {
			t.Fatalf("read runtime connection %d timeouts: %v", i, err)
		}
		if statementTimeout != runtimeStatementTimeout || lockTimeout != runtimeLockTimeout || idleTimeout != runtimeIdleInTransactionTimeout {
			t.Fatalf("runtime connection %d timeout envelope = statement=%q lock=%q idle=%q", i, statementTimeout, lockTimeout, idleTimeout)
		}
	}
}

func TestRuntimeTimeoutFailuresRollbackAndLeavePoolUsable(t *testing.T) {
	ownerURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	runtimeURL := os.Getenv("OPENINVEST_DATABASE_RUNTIME_TEST_URL")
	if ownerURL == "" || runtimeURL == "" {
		t.Skip("runtime timeout integration URLs are not set")
	}

	ownerDB, err := sql.Open("pgx", ownerURL)
	if err != nil {
		t.Fatalf("open owner database: %v", err)
	}
	t.Cleanup(func() { _ = ownerDB.Close() })
	runtime, err := OpenRuntime(runtimeURL)
	if err != nil {
		t.Fatalf("open runtime store: %v", err)
	}
	t.Cleanup(func() { _ = runtime.Close() })
	runtime.db.SetMaxOpenConns(1)
	runtime.db.SetMaxIdleConns(1)

	assertPoolUsable := func(label string) {
		t.Helper()
		var value int
		if err := runtime.db.QueryRow(`SELECT 1`).Scan(&value); err != nil || value != 1 {
			t.Fatalf("runtime pool was unusable after %s: value=%d err=%v", label, value, err)
		}
	}
	assertAbortedTransaction := func(label, timeout string, statement string, wantCode string) {
		t.Helper()
		tx, err := runtime.db.BeginTx(context.Background(), nil)
		if err != nil {
			t.Fatalf("begin runtime transaction for %s: %v", label, err)
		}
		if _, err := tx.Exec(`SELECT set_config($1, $2, true)`, timeout, "20ms"); err != nil {
			_ = tx.Rollback()
			t.Fatalf("set %s for %s: %v", timeout, label, err)
		}
		_, err = tx.Exec(statement)
		if err == nil {
			_ = tx.Rollback()
			t.Fatalf("%s unexpectedly succeeded", label)
		}
		if code := postgresErrorCode(err); code != wantCode {
			_ = tx.Rollback()
			t.Fatalf("%s SQLSTATE = %q, want %q: %v", label, code, wantCode, err)
		}
		if err := tx.Rollback(); err != nil {
			t.Fatalf("rollback %s transaction: %v", label, err)
		}
		assertPoolUsable(label)
	}

	assertAbortedTransaction("statement timeout", "statement_timeout", `SELECT pg_sleep(0.1)`, "57014")

	const advisoryLockID = 710033
	if _, err := ownerDB.Exec(`SELECT pg_advisory_lock($1)`, advisoryLockID); err != nil {
		t.Fatalf("acquire owner advisory lock: %v", err)
	}
	t.Cleanup(func() { _, _ = ownerDB.Exec(`SELECT pg_advisory_unlock($1)`, advisoryLockID) })
	assertAbortedTransaction("lock timeout", "lock_timeout", `SELECT pg_advisory_xact_lock(710033)`, "55P03")

	conn, err := runtime.db.Conn(context.Background())
	if err != nil {
		t.Fatalf("acquire runtime connection for idle timeout: %v", err)
	}
	if _, err := conn.ExecContext(context.Background(), `BEGIN; SET idle_in_transaction_session_timeout = '50ms'`); err != nil {
		_ = conn.Close()
		t.Fatalf("begin idle timeout transaction: %v", err)
	}
	time.Sleep(150 * time.Millisecond)
	_, err = conn.ExecContext(context.Background(), `SELECT 1`)
	_ = conn.Close()
	if err == nil {
		t.Fatal("idle-in-transaction timeout unexpectedly allowed a query")
	}
	if code := postgresErrorCode(err); code != "25P03" {
		t.Fatalf("idle-in-transaction SQLSTATE = %q, want 25P03: %v", code, err)
	}
	assertPoolUsable("idle-in-transaction timeout")
}

func postgresErrorCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}
