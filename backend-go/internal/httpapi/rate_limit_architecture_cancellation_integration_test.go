package httpapi

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

func TestArchitectureHardeningHTTPForcedRequestContextCancelsPostgres(t *testing.T) {
	runtimeURL := os.Getenv("OPENINVEST_DATABASE_RUNTIME_TEST_URL")
	ownerURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if runtimeURL == "" || ownerURL == "" {
		t.Skip("PostgreSQL runtime/owner test URLs are not set")
	}

	runtimeConfig, err := pgx.ParseConfig(runtimeURL)
	if err != nil {
		t.Fatalf("parse runtime URL: %v", err)
	}
	if runtimeConfig.RuntimeParams == nil {
		runtimeConfig.RuntimeParams = map[string]string{}
	}
	runtimeConfig.RuntimeParams["statement_timeout"] = "30s"
	runtimeConfig.RuntimeParams["lock_timeout"] = "5s"
	runtimeConfig.RuntimeParams["idle_in_transaction_session_timeout"] = "30s"
	runtimeConfig.RuntimeParams["application_name"] = "oi-module9-http-cancel"
	runtimeDB := stdlib.OpenDB(*runtimeConfig)
	runtimeDB.SetMaxOpenConns(10)
	runtimeDB.SetMaxIdleConns(5)
	defer runtimeDB.Close()

	ownerConfig, err := pgx.ParseConfig(ownerURL)
	if err != nil {
		t.Fatalf("parse owner URL: %v", err)
	}
	ownerDB := stdlib.OpenDB(*ownerConfig)
	defer ownerDB.Close()

	lifecycle := NewRequestLifecycle()
	app := fiber.New()
	app.Use(lifecycle.Middleware)

	backendPID := make(chan int, 1)
	databaseResult := make(chan error, 1)
	app.Get("/audit/cancel", func(c fiber.Ctx) error {
		conn, err := runtimeDB.Conn(c.Context())
		if err != nil {
			databaseResult <- err
			return err
		}
		defer conn.Close()

		var pid int
		if err := conn.QueryRowContext(c.Context(), `SELECT pg_backend_pid()`).Scan(&pid); err != nil {
			databaseResult <- err
			return err
		}
		backendPID <- pid
		_, err = conn.ExecContext(c.Context(), `SELECT pg_sleep(8)`)
		databaseResult <- err
		return err
	})

	requestDone := make(chan error, 1)
	go func() {
		response, requestErr := app.Test(httptest.NewRequest(http.MethodGet, "/audit/cancel", nil))
		if response != nil {
			_ = response.Body.Close()
		}
		requestDone <- requestErr
	}()

	var pid int
	select {
	case pid = <-backendPID:
	case <-time.After(3 * time.Second):
		t.Fatal("HTTP handler did not acquire PostgreSQL backend")
	}
	waitHTTPArchitectureHardeningQueryActive(t, ownerDB, pid)

	before := runtimeDB.Stats()
	cancelAt := time.Now()
	lifecycle.ForceCancel()

	var dbErr error
	select {
	case dbErr = <-databaseResult:
	case <-time.After(2 * time.Second):
		t.Fatal("database call did not return after forced HTTP request-context cancellation")
	}
	termination := time.Since(cancelAt)
	if dbErr == nil {
		t.Fatal("database long query completed without cancellation error")
	}
	if termination >= 8*time.Second {
		t.Fatalf("HTTP-to-PostgreSQL cancellation latency=%s, expected materially below 8s", termination)
	}

	select {
	case <-requestDone:
	case <-time.After(2 * time.Second):
		t.Fatal("HTTP request did not unwind after cancellation")
	}
	lifecycle.Wait()
	waitHTTPArchitectureHardeningQueryGone(t, ownerDB, pid)

	var one int
	if err := runtimeDB.QueryRowContext(context.Background(), `SELECT 1`).Scan(&one); err != nil || one != 1 {
		t.Fatalf("post-cancel SELECT 1=%d err=%v", one, err)
	}
	after := runtimeDB.Stats()
	if after.InUse != 0 {
		t.Fatalf("HTTP cancellation left pool slot in use: %+v", after)
	}

	t.Logf(
		"HTTP_CANCEL_PROOF BACKEND_PID=%d QUERY_ACTIVE_BEFORE_CANCEL=YES CLIENT_CANCEL_SENT=YES REQUEST_CONTEXT_DONE=YES DATABASE_CALL_ERROR=%q POSTGRES_QUERY_TERMINATED=YES QUERY_NATURAL_DURATION_MS=8000 TERMINATION_LATENCY_MS=%d POST_CANCEL_SELECT_1=SUCCESS CONNECTION_REUSABLE=YES DBSTATS_IN_USE_BEFORE=%d DBSTATS_IN_USE_AFTER=%d",
		pid,
		dbErr.Error(),
		termination.Milliseconds(),
		before.InUse,
		after.InUse,
	)
}

func waitHTTPArchitectureHardeningQueryActive(t *testing.T, db *sql.DB, pid int) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		var state, query string
		err := db.QueryRowContext(context.Background(), `
			SELECT state::text, query::text
			FROM pg_stat_activity
			WHERE pid=$1
		`, pid).Scan(&state, &query)
		if err == nil && state == "active" && strings.Contains(query, "pg_sleep(8)") {
			return
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("inspect HTTP cancellation backend pid=%d: %v", pid, err)
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("HTTP cancellation backend pid=%d never became active", pid)
}

func waitHTTPArchitectureHardeningQueryGone(t *testing.T, db *sql.DB, pid int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		var state, query string
		err := db.QueryRowContext(context.Background(), `
			SELECT state::text, query::text
			FROM pg_stat_activity
			WHERE pid=$1
		`, pid).Scan(&state, &query)
		if errors.Is(err, sql.ErrNoRows) {
			return
		}
		if err != nil {
			t.Fatalf("inspect HTTP terminated backend pid=%d: %v", pid, err)
		}
		if state != "active" || !strings.Contains(query, "pg_sleep(8)") {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("HTTP cancellation backend pid=%d remained active", pid)
}
