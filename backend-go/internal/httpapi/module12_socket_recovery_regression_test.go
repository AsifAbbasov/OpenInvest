package httpapi

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

// Preserve the current-contract recovery assertion from historical PR #246.
// Exercise the actual API listener with bounded, deliberately incomplete wire
// requests, then verify that an ordinary request is still served. This does
// not assert production edge, proxy, or distributed admission behavior.
func TestModule12MalformedSocketRequestsDoNotDisableAPILiveness(t *testing.T) {
	app := newHTTPAuthApp(t, newHTTPAuthService(t, &httpAuthTestStore{}))
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	served := make(chan error, 1)
	go func() { served <- app.Listener(listener) }()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := app.ShutdownWithContext(ctx); err != nil {
			t.Errorf("shutdown: %v", err)
		}
		_ = listener.Close()
		select {
		case <-served:
		case <-ctx.Done():
			t.Error("listener did not terminate")
		}
	})

	for _, tc := range []struct {
		name      string
		payload   string
		closeWrite bool
		reset     bool
	}{
		{name: "idle_connection"},
		{name: "partial_request_line", payload: "GE"},
		{name: "partial_header_name", payload: "GET /api/v1/health HTTP/1.1\r\nHo"},
		{name: "partial_header_value", payload: "GET /api/v1/health HTTP/1.1\r\nHost:"},
		{name: "incomplete_request_body", payload: "POST /api/v1/auth/register HTTP/1.1\r\nHost: localhost\r\nContent-Length: 100\r\nContent-Type: application/json\r\n\r\n{\"email\":"},
		{name: "truncated_chunked_body", payload: "POST /api/v1/auth/register HTTP/1.1\r\nHost: localhost\r\nTransfer-Encoding: chunked\r\n\r\n5\r\nabc"},
		{name: "conflicting_content_lengths", payload: "POST /api/v1/auth/register HTTP/1.1\r\nHost: localhost\r\nContent-Length: 2\r\nContent-Length: 3\r\n\r\n{}"},
		{name: "large_header", payload: "GET /api/v1/health HTTP/1.1\r\nHost: localhost\r\nX-Fixture: " + strings.Repeat("x", 64*1024) + "\r\n\r\n"},
		{name: "half_closed_request", payload: "GET /api/v1/health HTTP/1.1\r\nHost: localhost\r\n\r\n", closeWrite: true},
		{name: "reset_connection", payload: "GET /api/v1/health HTTP/1.1\r\nHost: localhost\r\n\r\n", reset: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			conn, err := net.DialTimeout("tcp", listener.Addr().String(), time.Second)
			if err != nil {
				t.Fatal(err)
			}
			if err := conn.SetDeadline(time.Now().Add(time.Second)); err != nil {
				conn.Close()
				t.Fatal(err)
			}
			if tc.payload != "" {
				_, _ = io.WriteString(conn, tc.payload)
			}
			if tcp, ok := conn.(*net.TCPConn); ok {
				if tc.closeWrite {
					_ = tcp.CloseWrite()
				}
				if tc.reset {
					_ = tcp.SetLinger(0)
				}
			}
			_ = conn.Close()

			client := &http.Client{Timeout: 2 * time.Second}
			resp, err := client.Get("http://" + listener.Addr().String() + "/api/v1/health")
			if err != nil {
				t.Fatalf("liveness request after %s failed: %v", tc.name, err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("liveness after %s returned HTTP %d", tc.name, resp.StatusCode)
			}
		})
	}
}
