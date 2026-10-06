package postgres

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const cancellationNaturalDuration = 8 * time.Second

func TestPostgresContextCancellationProofRepeated(t *testing.T) {
	ownerURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	runtimeURL := os.Getenv("OPENINVEST_DATABASE_RUNTIME_TEST_URL")
	if ownerURL == "" || runtimeURL == "" {
		t.Skip("PostgreSQL test URLs are not configured")
	}
	owner, err := sql.Open("pgx", ownerURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = owner.Close() })

	store, err := OpenRuntime(runtimeURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })

	const iterations = 25
	for iteration := 0; iteration < iterations; iteration++ {
		conn, err := store.db.Conn(context.Background())
		if err != nil {
			t.Fatalf("iteration %d acquire runtime connection: %v", iteration, err)
		}
		var pid int
		if err := conn.QueryRowContext(context.Background(), "SELECT pg_backend_pid()").Scan(&pid); err != nil {
			_ = conn.Close()
			t.Fatalf("iteration %d backend pid: %v", iteration, err)
		}

		queryCtx, cancel := context.WithCancel(context.Background())
		done := make(chan error, 1)
		started := time.Now()
		go func() {
			_, err := conn.ExecContext(queryCtx, "SELECT pg_sleep(8)")
			done <- err
		}()

		if !waitForActiveSleepQuery(t, owner, pid, 2*time.Second) {
			cancel()
			_ = conn.Close()
			t.Fatalf("iteration %d backend %d query never became active", iteration, pid)
		}
		cancelAt := time.Now()
		cancel()
		err = waitCancellationResult(t, done, 2*time.Second)
		terminationLatency := time.Since(cancelAt)
		totalDuration := time.Since(started)
		if err == nil {
			_ = conn.Close()
			t.Fatalf("iteration %d long query completed naturally", iteration)
		}
		if totalDuration >= cancellationNaturalDuration {
			_ = conn.Close()
			t.Fatalf("iteration %d cancellation did not beat natural duration: %s", iteration, totalDuration)
		}
		if !errors.Is(queryCtx.Err(), context.Canceled) {
			_ = conn.Close()
			t.Fatalf("iteration %d request context error=%v", iteration, queryCtx.Err())
		}
		if !waitForSleepQueryGone(t, owner, pid, 2*time.Second) {
			_ = conn.Close()
			t.Fatalf("iteration %d backend %d remained active after cancel", iteration, pid)
		}
		var one int
		if err := conn.QueryRowContext(context.Background(), "SELECT 1").Scan(&one); err != nil || one != 1 {
			_ = conn.Close()
			t.Fatalf("iteration %d connection not reusable: one=%d err=%v", iteration, one, err)
		}
		if err := conn.Close(); err != nil {
			t.Fatalf("iteration %d close connection: %v", iteration, err)
		}
		t.Logf("CANCEL_PROOF iteration=%d backend_pid=%d query_active_before_cancel=YES context_done=YES postgres_query_terminated=YES natural_ms=%d termination_latency_ms=%d post_cancel_select_1=SUCCESS",
			iteration+1, pid, cancellationNaturalDuration.Milliseconds(), terminationLatency.Milliseconds())
	}
	t.Logf("CANCEL_SUCCESS_RATE=100%% iterations=%d", iterations)
}

func TestPostgresCancellationStormRecoversCanonicalPool(t *testing.T) {
	ownerURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	runtimeURL := os.Getenv("OPENINVEST_DATABASE_RUNTIME_TEST_URL")
	if ownerURL == "" || runtimeURL == "" {
		t.Skip("PostgreSQL test URLs are not configured")
	}
	owner, err := sql.Open("pgx", ownerURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = owner.Close() })
	store, err := OpenRuntime(runtimeURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })

	before := store.db.Stats()
	const concurrentQueries = 10
	conns := make([]*sql.Conn, 0, concurrentQueries)
	pids := make([]int, 0, concurrentQueries)
	cancels := make([]context.CancelFunc, 0, concurrentQueries)
	results := make([]<-chan error, 0, concurrentQueries)

	for i := 0; i < concurrentQueries; i++ {
		conn, err := store.db.Conn(context.Background())
		if err != nil {
			t.Fatalf("acquire connection %d: %v", i, err)
		}
		conns = append(conns, conn)
		var pid int
		if err := conn.QueryRowContext(context.Background(), "SELECT pg_backend_pid()").Scan(&pid); err != nil {
			t.Fatalf("backend pid %d: %v", i, err)
		}
		pids = append(pids, pid)
		ctx, cancel := context.WithCancel(context.Background())
		cancels = append(cancels, cancel)
		done := make(chan error, 1)
		results = append(results, done)
		go func(conn *sql.Conn, ctx context.Context, done chan<- error) {
			_, err := conn.ExecContext(ctx, "SELECT pg_sleep(8)")
			done <- err
		}(conn, ctx, done)
	}

	if got := waitForActiveSleepCount(t, owner, pids, concurrentQueries, 3*time.Second); got != concurrentQueries {
		t.Fatalf("active sleep queries=%d want=%d", got, concurrentQueries)
	}
	atPeak := store.db.Stats()
	if atPeak.InUse != concurrentQueries || atPeak.MaxOpenConnections != concurrentQueries {
		t.Fatalf("pool peak in_use=%d max=%d want=%d", atPeak.InUse, atPeak.MaxOpenConnections, concurrentQueries)
	}

	cancelAt := time.Now()
	for _, cancel := range cancels {
		cancel()
	}
	cancelled := 0
	natural := 0
	for _, result := range results {
		err := waitCancellationResult(t, result, 3*time.Second)
		if err == nil {
			natural++
		} else {
			cancelled++
		}
	}
	for _, conn := range conns {
		if err := conn.Close(); err != nil {
			t.Fatalf("close storm connection: %v", err)
		}
	}
	if cancelled != concurrentQueries || natural != 0 {
		t.Fatalf("storm cancelled=%d natural=%d", cancelled, natural)
	}
	if !waitForAllSleepQueriesGone(t, owner, pids, 3*time.Second) {
		t.Fatal("one or more PostgreSQL sleep queries remained active after storm cancellation")
	}

	var one int
	if err := store.db.QueryRowContext(context.Background(), "SELECT 1").Scan(&one); err != nil || one != 1 {
		t.Fatalf("post-storm SELECT 1 one=%d err=%v", one, err)
	}
	after := store.db.Stats()
	if after.InUse != 0 {
		t.Fatalf("permanent pool occupancy after storm: %+v", after)
	}
	if after.OpenConnections > after.MaxIdleClosed+5 && after.OpenConnections > 5 {
		t.Fatalf("unexpected open connections after recovery: %+v", after)
	}
	t.Logf("CANCEL_STORM concurrent=%d cancelled=%d natural=%d termination_latency_ms=%d", concurrentQueries, cancelled, natural, time.Since(cancelAt).Milliseconds())
	t.Logf("DBSTATS before_max=%d before_open=%d before_in_use=%d before_idle=%d before_wait_count=%d before_wait_duration=%s before_max_idle_closed=%d before_max_idle_time_closed=%d before_max_lifetime_closed=%d",
		before.MaxOpenConnections, before.OpenConnections, before.InUse, before.Idle, before.WaitCount, before.WaitDuration, before.MaxIdleClosed, before.MaxIdleTimeClosed, before.MaxLifetimeClosed)
	t.Logf("DBSTATS peak_max=%d peak_open=%d peak_in_use=%d peak_idle=%d peak_wait_count=%d peak_wait_duration=%s",
		atPeak.MaxOpenConnections, atPeak.OpenConnections, atPeak.InUse, atPeak.Idle, atPeak.WaitCount, atPeak.WaitDuration)
	t.Logf("DBSTATS after_max=%d after_open=%d after_in_use=%d after_idle=%d after_wait_count=%d after_wait_duration=%s after_max_idle_closed=%d after_max_idle_time_closed=%d after_max_lifetime_closed=%d",
		after.MaxOpenConnections, after.OpenConnections, after.InUse, after.Idle, after.WaitCount, after.WaitDuration, after.MaxIdleClosed, after.MaxIdleTimeClosed, after.MaxLifetimeClosed)
}

func waitCancellationResult(t *testing.T, done <-chan error, timeout time.Duration) error {
	t.Helper()
	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		t.Fatal("timed out waiting for cancelled database call")
		return nil
	}
}

func waitForActiveSleepQuery(t *testing.T, owner *sql.DB, pid int, timeout time.Duration) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		var state, query string
		err := owner.QueryRowContext(context.Background(), "SELECT state, query FROM pg_stat_activity WHERE pid=$1", pid).Scan(&state, &query)
		if err == nil && state == "active" && strings.Contains(query, "pg_sleep(8)") {
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return false
}

func waitForSleepQueryGone(t *testing.T, owner *sql.DB, pid int, timeout time.Duration) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		var active int
		if err := owner.QueryRowContext(context.Background(), "SELECT count(*) FROM pg_stat_activity WHERE pid=$1 AND state='active' AND query LIKE '%pg_sleep(8)%'", pid).Scan(&active); err == nil && active == 0 {
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return false
}

func waitForActiveSleepCount(t *testing.T, owner *sql.DB, pids []int, want int, timeout time.Duration) int {
	t.Helper()
	deadline := time.Now().Add(timeout)
	last := 0
	for time.Now().Before(deadline) {
		last = activeSleepCount(t, owner, pids)
		if last == want {
			return last
		}
		time.Sleep(50 * time.Millisecond)
	}
	return last
}

func waitForAllSleepQueriesGone(t *testing.T, owner *sql.DB, pids []int, timeout time.Duration) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if activeSleepCount(t, owner, pids) == 0 {
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return false
}

func activeSleepCount(t *testing.T, owner *sql.DB, pids []int) int {
	t.Helper()
	if len(pids) == 0 {
		return 0
	}
	placeholders := make([]string, len(pids))
	args := make([]any, len(pids))
	for i, pid := range pids {
		placeholders[i] = "$" + strconv.Itoa(i+1)
		args[i] = pid
	}
	query := "SELECT count(*) FROM pg_stat_activity WHERE pid IN (" + strings.Join(placeholders, ",") + ") AND state='active' AND query LIKE '%pg_sleep(8)%'"
	var count int
	if err := owner.QueryRowContext(context.Background(), query, args...).Scan(&count); err != nil {
		t.Fatalf("sample pg_stat_activity: %v", err)
	}
	return count
}

