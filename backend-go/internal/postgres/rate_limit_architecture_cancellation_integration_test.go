package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	architectureHardeningCancellationIterations = 25
	architectureHardeningNaturalSleep           = 8 * time.Second
	architectureHardeningPollInterval           = 50 * time.Millisecond
)

func TestArchitectureHardeningPostgresCancellationProof(t *testing.T) {
	runtimeURL := os.Getenv("OPENINVEST_DATABASE_RUNTIME_TEST_URL")
	ownerURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if runtimeURL == "" || ownerURL == "" {
		t.Skip("PostgreSQL runtime/owner test URLs are not set")
	}
	runtimeStore, err := OpenRuntime(runtimeURL)
	if err != nil {
		t.Fatalf("open runtime store: %v", err)
	}
	defer func() { _ = runtimeStore.Close() }()
	ownerStore, err := Open(ownerURL)
	if err != nil {
		t.Fatalf("open owner store: %v", err)
	}
	defer func() { _ = ownerStore.Close() }()

	var statementTimeout, lockTimeout, idleTimeout string
	if err := runtimeStore.db.QueryRowContext(context.Background(),
		`SELECT current_setting('statement_timeout'), current_setting('lock_timeout'), current_setting('idle_in_transaction_session_timeout')`,
	).Scan(&statementTimeout, &lockTimeout, &idleTimeout); err != nil {
		t.Fatalf("read runtime timeouts: %v", err)
	}
	if statementTimeout != "30s" || lockTimeout != "5s" || idleTimeout != "30s" {
		t.Fatalf("runtime timeout envelope changed: statement=%s lock=%s idle=%s", statementTimeout, lockTimeout, idleTimeout)
	}

	naturalStart := time.Now()
	if _, err := runtimeStore.db.ExecContext(context.Background(), `SELECT pg_sleep(8)`); err != nil {
		t.Fatalf("natural control query: %v", err)
	}
	naturalDuration := time.Since(naturalStart)
	if naturalDuration < 7900*time.Millisecond {
		t.Fatalf("natural pg_sleep duration=%s, expected approximately 8s", naturalDuration)
	}
	t.Logf("CANCEL_PROOF_NATURAL_DURATION_MS=%d", naturalDuration.Milliseconds())

	statsBefore := runtimeStore.db.Stats()
	var terminationTotal time.Duration
	for i := 0; i < architectureHardeningCancellationIterations; i++ {
		evidence := runArchitectureHardeningCancellationIteration(t, runtimeStore.db, ownerStore.db, i)
		terminationTotal += evidence.terminationLatency
		t.Logf(
			"CANCEL_PROOF_ITERATION=%d BACKEND_PID=%d QUERY_ACTIVE_BEFORE_CANCEL=YES CLIENT_CANCEL_SENT=YES REQUEST_CONTEXT_DONE=YES DATABASE_CALL_ERROR=%q POSTGRES_QUERY_TERMINATED=YES TERMINATION_LATENCY_MS=%d POST_CANCEL_SELECT_1=SUCCESS CONNECTION_REUSABLE=YES",
			i+1,
			evidence.backendPID,
			evidence.databaseError,
			evidence.terminationLatency.Milliseconds(),
		)
	}
	statsAfter := runtimeStore.db.Stats()
	t.Logf(
		"CANCEL_PROOF_REPEAT ITERATIONS=%d SUCCESS=%d SUCCESS_RATE=100%% POOL_RECOVERY_RATE=100%% AVG_TERMINATION_LATENCY_MS=%d DBSTATS_WAIT_COUNT_BEFORE=%d DBSTATS_WAIT_COUNT_AFTER=%d DBSTATS_WAIT_DURATION_BEFORE=%s DBSTATS_WAIT_DURATION_AFTER=%s",
		architectureHardeningCancellationIterations,
		architectureHardeningCancellationIterations,
		(terminationTotal / architectureHardeningCancellationIterations).Milliseconds(),
		statsBefore.WaitCount,
		statsAfter.WaitCount,
		statsBefore.WaitDuration,
		statsAfter.WaitDuration,
	)

	runArchitectureHardeningCancellationStorm(t, runtimeStore.db, ownerStore.db)
}

type architectureHardeningCancellationEvidence struct {
	backendPID         int
	databaseError      string
	terminationLatency time.Duration
}

func runArchitectureHardeningCancellationIteration(
	t *testing.T,
	runtimeDB *sql.DB,
	ownerDB *sql.DB,
	iteration int,
) architectureHardeningCancellationEvidence {
	t.Helper()
	conn, err := runtimeDB.Conn(context.Background())
	if err != nil {
		t.Fatalf("iteration %d acquire runtime connection: %v", iteration+1, err)
	}
	defer conn.Close()

	var backendPID int
	if err := conn.QueryRowContext(context.Background(), `SELECT pg_backend_pid()`).Scan(&backendPID); err != nil {
		t.Fatalf("iteration %d backend pid: %v", iteration+1, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, queryErr := conn.ExecContext(ctx, `SELECT pg_sleep(8)`)
		done <- queryErr
	}()

	waitArchitectureHardeningQueryActive(t, ownerDB, backendPID, architectureHardeningPollInterval)
	cancelAt := time.Now()
	cancel()
	queryErr := <-done
	terminationLatency := time.Since(cancelAt)
	if queryErr == nil {
		t.Fatalf("iteration %d long query completed naturally after cancellation", iteration+1)
	}
	if ctx.Err() != context.Canceled {
		t.Fatalf("iteration %d request context error=%v, want canceled", iteration+1, ctx.Err())
	}
	if terminationLatency >= architectureHardeningNaturalSleep {
		t.Fatalf("iteration %d cancellation took %s, not materially before 8s completion", iteration+1, terminationLatency)
	}
	waitArchitectureHardeningQueryGone(t, ownerDB, backendPID, 2*time.Second)

	var one int
	if err := conn.QueryRowContext(context.Background(), `SELECT 1`).Scan(&one); err != nil || one != 1 {
		t.Fatalf("iteration %d post-cancel SELECT 1 value=%d err=%v", iteration+1, one, err)
	}
	return architectureHardeningCancellationEvidence{
		backendPID: backendPID,
		databaseError: queryErr.Error(),
		terminationLatency: terminationLatency,
	}
}

func waitArchitectureHardeningQueryActive(t *testing.T, ownerDB *sql.DB, backendPID int, interval time.Duration) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		var state, query string
		err := ownerDB.QueryRowContext(context.Background(), `
			SELECT state::text, query::text
			FROM pg_stat_activity
			WHERE pid=$1
		`, backendPID).Scan(&state, &query)
		if err == nil && state == "active" && strings.Contains(query, "pg_sleep(8)") {
			return
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("inspect pg_stat_activity pid=%d: %v", backendPID, err)
		}
		time.Sleep(interval)
	}
	t.Fatalf("backend pid=%d did not become active with controlled pg_sleep query", backendPID)
}

func waitArchitectureHardeningQueryGone(t *testing.T, ownerDB *sql.DB, backendPID int, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		var state, query string
		err := ownerDB.QueryRowContext(context.Background(), `
			SELECT state::text, query::text
			FROM pg_stat_activity
			WHERE pid=$1
		`, backendPID).Scan(&state, &query)
		if errors.Is(err, sql.ErrNoRows) {
			return
		}
		if err != nil {
			t.Fatalf("inspect terminated query pid=%d: %v", backendPID, err)
		}
		if state != "active" || !strings.Contains(query, "pg_sleep(8)") {
			return
		}
		time.Sleep(architectureHardeningPollInterval)
	}
	t.Fatalf("backend pid=%d still executing pg_sleep after cancellation", backendPID)
}

func runArchitectureHardeningCancellationStorm(t *testing.T, runtimeDB *sql.DB, ownerDB *sql.DB) {
	t.Helper()
	const queries = 10
	before := runtimeDB.Stats()

	type stormQuery struct {
		conn   *sql.Conn
		pid    int
		cancel context.CancelFunc
		done   chan error
	}
	storm := make([]stormQuery, 0, queries)
	for i := 0; i < queries; i++ {
		conn, err := runtimeDB.Conn(context.Background())
		if err != nil {
			t.Fatalf("storm acquire connection %d: %v", i, err)
		}
		var pid int
		if err := conn.QueryRowContext(context.Background(), `SELECT pg_backend_pid()`).Scan(&pid); err != nil {
			conn.Close()
			t.Fatalf("storm backend pid %d: %v", i, err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan error, 1)
		go func(c *sql.Conn, queryCtx context.Context, result chan<- error) {
			_, queryErr := c.ExecContext(queryCtx, `SELECT pg_sleep(8)`)
			result <- queryErr
		}(conn, ctx, done)
		storm = append(storm, stormQuery{conn: conn, pid: pid, cancel: cancel, done: done})
	}

	deadline := time.Now().Add(3 * time.Second)
	for {
		active := architectureHardeningActiveSleepQueries(t, ownerDB)
		if active >= queries {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("storm active PostgreSQL queries=%d, want %d", active, queries)
		}
		time.Sleep(architectureHardeningPollInterval)
	}
	inFlight := runtimeDB.Stats()
	cancelAt := time.Now()
	for i := range storm {
		storm[i].cancel()
	}
	cancelled := 0
	natural := 0
	for i := range storm {
		err := <-storm[i].done
		if err != nil {
			cancelled++
		} else {
			natural++
		}
		storm[i].cancel()
		if closeErr := storm[i].conn.Close(); closeErr != nil {
			t.Fatalf("storm close conn %d: %v", i, closeErr)
		}
	}
	termination := time.Since(cancelAt)
	if cancelled != queries || natural != 0 {
		t.Fatalf("storm cancelled=%d natural=%d", cancelled, natural)
	}
	if termination >= architectureHardeningNaturalSleep {
		t.Fatalf("storm termination=%s, expected materially before natural completion", termination)
	}

	deadline = time.Now().Add(2 * time.Second)
	for architectureHardeningActiveSleepQueries(t, ownerDB) != 0 {
		if time.Now().After(deadline) {
			t.Fatal("storm PostgreSQL sleep queries remained active")
		}
		time.Sleep(architectureHardeningPollInterval)
	}
	after := runtimeDB.Stats()
	if after.InUse != 0 {
		t.Fatalf("storm pool retained in-use connections: %+v", after)
	}
	var one int
	if err := runtimeDB.QueryRowContext(context.Background(), `SELECT 1`).Scan(&one); err != nil || one != 1 {
		t.Fatalf("storm post-cancel SELECT 1=%d err=%v", one, err)
	}
	t.Logf(
		"CANCEL_STORM CONCURRENT_LONG_QUERIES=%d ACTIVE_POSTGRES_QUERIES_BEFORE_CANCEL=%d CANCELLED_QUERIES=%d NATURAL_COMPLETIONS=%d TERMINATION_LATENCY_MS=%d DBSTATS_MAX_OPEN=%d DBSTATS_OPEN_BEFORE=%d DBSTATS_IN_USE_BEFORE=%d DBSTATS_IDLE_BEFORE=%d DBSTATS_WAIT_COUNT_BEFORE=%d DBSTATS_WAIT_DURATION_BEFORE=%s DBSTATS_OPEN_INFLIGHT=%d DBSTATS_IN_USE_INFLIGHT=%d DBSTATS_IDLE_INFLIGHT=%d DBSTATS_OPEN_AFTER=%d DBSTATS_IN_USE_AFTER=%d DBSTATS_IDLE_AFTER=%d DBSTATS_WAIT_COUNT_AFTER=%d DBSTATS_WAIT_DURATION_AFTER=%s POOL_RECOVERS=YES NO_CONNECTION_LEAK=YES",
		queries, queries, cancelled, natural, termination.Milliseconds(),
		before.MaxOpenConnections, before.OpenConnections, before.InUse, before.Idle, before.WaitCount, before.WaitDuration,
		inFlight.OpenConnections, inFlight.InUse, inFlight.Idle,
		after.OpenConnections, after.InUse, after.Idle, after.WaitCount, after.WaitDuration,
	)
}

func architectureHardeningActiveSleepQueries(t *testing.T, ownerDB *sql.DB) int {
	t.Helper()
	var count int
	if err := ownerDB.QueryRowContext(context.Background(), `
		SELECT count(*)
		FROM pg_stat_activity
		WHERE state='active'
		  AND query LIKE '%pg_sleep(8)%'
		  AND application_name <> current_setting('application_name')
	`).Scan(&count); err != nil {
		t.Fatalf("count active sleep queries: %v", err)
	}
	return count
}

func TestArchitectureHardeningCancellationConstants(t *testing.T) {
	if architectureHardeningCancellationIterations < 25 {
		t.Fatalf("cancellation iterations=%d, want >=25", architectureHardeningCancellationIterations)
	}
	if architectureHardeningNaturalSleep < 8*time.Second {
		t.Fatalf("natural sleep=%s, want >=8s", architectureHardeningNaturalSleep)
	}
	if architectureHardeningPollInterval < 50*time.Millisecond || architectureHardeningPollInterval > 100*time.Millisecond {
		t.Fatalf("pg_stat_activity sampling=%s, want 50-100ms", architectureHardeningPollInterval)
	}
	_ = fmt.Sprintf
	_ = sync.Once{}
}
