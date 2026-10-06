package postgres

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/openinvest/openinvest/backend-go/internal/httpapi"
)

func TestHTTPRequestCancellationPropagatesToPostgres(t *testing.T) {
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

	lifecycle := httpapi.NewRequestLifecycle()
	app := fiber.New()
	app.Use(lifecycle.Middleware)

	pidCh := make(chan int, 1)
	dbDone := make(chan error, 1)
	app.Get("/audit-cancel", func(c fiber.Ctx) error {
		conn, err := store.db.Conn(c.Context())
		if err != nil {
			dbDone <- err
			return err
		}
		defer conn.Close()
		var pid int
		if err := conn.QueryRowContext(c.Context(), "SELECT pg_backend_pid()").Scan(&pid); err != nil {
			dbDone <- err
			return err
		}
		pidCh <- pid
		_, err = conn.ExecContext(c.Context(), "SELECT pg_sleep(8)")
		dbDone <- err
		return err
	})

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	serveDone := make(chan error, 1)
	go func() { serveDone <- app.Listener(listener) }()
	t.Cleanup(func() {
		_ = app.Shutdown()
		_ = listener.Close()
	})

	requestCtx, cancelClient := context.WithCancel(context.Background())
	req, err := http.NewRequestWithContext(requestCtx, http.MethodGet, "http://"+listener.Addr().String()+"/audit-cancel", nil)
	if err != nil {
		t.Fatal(err)
	}
	clientDone := make(chan error, 1)
	go func() {
		resp, err := http.DefaultClient.Do(req)
		if resp != nil {
			_ = resp.Body.Close()
		}
		clientDone <- err
	}()

	var pid int
	select {
	case pid = <-pidCh:
	case <-time.After(2 * time.Second):
		t.Fatal("HTTP handler did not reach PostgreSQL")
	}
	if !waitForActiveSleepQuery(t, owner, pid, 2*time.Second) {
		t.Fatalf("backend %d query not active before client cancellation", pid)
	}

	cancelAt := time.Now()
	cancelClient()
	select {
	case <-clientDone:
	case <-time.After(2 * time.Second):
		t.Fatal("HTTP client did not observe cancellation")
	}

	select {
	case err := <-dbDone:
		if err == nil {
			t.Fatal("database query completed naturally after HTTP client cancellation")
		}
		if time.Since(cancelAt) >= cancellationNaturalDuration {
			t.Fatalf("database cancellation did not beat natural duration: %s", time.Since(cancelAt))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("HTTP client cancellation did not propagate to PostgreSQL within 2s")
	}

	if !waitForSleepQueryGone(t, owner, pid, 2*time.Second) {
		t.Fatalf("backend %d remained active after HTTP cancellation", pid)
	}
	var one int
	if err := store.db.QueryRowContext(context.Background(), "SELECT 1").Scan(&one); err != nil || one != 1 {
		t.Fatalf("pool not reusable after HTTP cancellation: one=%d err=%v", one, err)
	}
	t.Logf("HTTP_TO_POSTGRES_CANCEL_PROVEN backend_pid=%d query_active_before_cancel=YES client_cancel_sent=YES termination_latency_ms=%d post_cancel_select_1=SUCCESS", pid, time.Since(cancelAt).Milliseconds())

	// The test owns the route; production code gains no audit endpoint or pg_sleep.
	_ = strings.Builder{}
}
