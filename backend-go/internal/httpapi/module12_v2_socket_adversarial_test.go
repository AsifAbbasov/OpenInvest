package httpapi

import (
	"io"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"
)

func m12V2FDCount() int {
	entries, err := os.ReadDir("/proc/self/fd")
	if err != nil { return -1 }
	return len(entries)
}

func TestModule12V2BoundedRealSocketAdversarialMatrix(t *testing.T) {
	requireModule12V2Adversarial(t)
	app := newHTTPAuthApp(t, newHTTPAuthService(t, &httpAuthTestStore{}))
	s := startM12V2Listener(t, app)
	addr := s.ln.Addr().String()

	type socketCase struct {
		name string
		send func(net.Conn)
		linger time.Duration
	}
	raw := func(payload string) func(net.Conn) {
		return func(c net.Conn) { _, _ = io.WriteString(c, payload) }
	}
	cases := []socketCase{
		{"connect_then_idle", func(net.Conn){}, 120*time.Millisecond},
		{"partial_request_line", raw("GE"), 120*time.Millisecond},
		{"partial_header_name", raw("GET /api/v1/health HTTP/1.1\r\nHo"), 120*time.Millisecond},
		{"partial_header_value", raw("GET /api/v1/health HTTP/1.1\r\nHost:"), 120*time.Millisecond},
		{"headers_complete_no_body", raw("POST /api/v1/auth/register HTTP/1.1\r\nHost: localhost\r\nContent-Length: 100\r\nContent-Type: application/json\r\n\r\n"), 120*time.Millisecond},
		{"partial_body_then_idle", raw("POST /api/v1/auth/register HTTP/1.1\r\nHost: localhost\r\nContent-Length: 100\r\nContent-Type: application/json\r\n\r\n{\"email\":"), 120*time.Millisecond},
		{"partial_body_then_FIN", func(c net.Conn){ raw("POST /api/v1/auth/register HTTP/1.1\r\nHost: localhost\r\nContent-Length: 100\r\n\r\nabc")(c); if tcp, ok := c.(*net.TCPConn); ok { _ = tcp.CloseWrite() } }, 50*time.Millisecond},
		{"partial_body_then_RST", func(c net.Conn){ raw("POST /api/v1/auth/register HTTP/1.1\r\nHost: localhost\r\nContent-Length: 100\r\n\r\nabc")(c); if tcp, ok := c.(*net.TCPConn); ok { _ = tcp.SetLinger(0) } }, 0},
		{"complete_body_disconnect_before_response", raw("POST /api/v1/auth/register HTTP/1.1\r\nHost: localhost\r\nContent-Type: application/json\r\nContent-Length: 2\r\n\r\n{}"), 0},
		{"tcp_rst", func(c net.Conn){ if tcp, ok := c.(*net.TCPConn); ok { _ = tcp.SetLinger(0) } }, 0},
		{"half_close_write", func(c net.Conn){ raw("GET /api/v1/health HTTP/1.1\r\nHost: localhost\r\n\r\n")(c); if tcp, ok := c.(*net.TCPConn); ok { _ = tcp.CloseWrite() } }, 50*time.Millisecond},
		{"slow_byte_drip", func(c net.Conn){ for _, b := range []byte("GET /api/v1/health HTTP/1.1\r\nHost: localhost\r\n\r\n") { _, _ = c.Write([]byte{b}); time.Sleep(2*time.Millisecond) } }, 0},
		{"duplicate_content_length", raw("POST /api/v1/auth/register HTTP/1.1\r\nHost: localhost\r\nContent-Length: 2\r\nContent-Length: 2\r\n\r\n{}"), 0},
		{"conflicting_content_length", raw("POST /api/v1/auth/register HTTP/1.1\r\nHost: localhost\r\nContent-Length: 2\r\nContent-Length: 3\r\n\r\n{}"), 0},
		{"partial_chunked_body", raw("POST /api/v1/auth/register HTTP/1.1\r\nHost: localhost\r\nTransfer-Encoding: chunked\r\n\r\n5\r\nabc"), 100*time.Millisecond},
		{"malformed_chunk", raw("POST /api/v1/auth/register HTTP/1.1\r\nHost: localhost\r\nTransfer-Encoding: chunked\r\n\r\nZZ\r\nabc\r\n0\r\n\r\n"), 0},
		{"oversized_header_with_bounded_fixture", raw("GET /api/v1/health HTTP/1.1\r\nHost: localhost\r\nX-M12: "+strings.Repeat("a",64*1024)+"\r\n\r\n"), 0},
		{"oversized_body_with_bounded_fixture", raw("POST /api/v1/auth/register HTTP/1.1\r\nHost: localhost\r\nContent-Length: 1048576\r\nContent-Type: application/json\r\n\r\n"+strings.Repeat("x",1024*1024)), 0},
	}
	if len(cases) > 32 { t.Fatal("MAX_CONCURRENT_ATTACK_CONNECTIONS ceiling violated") }

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			beforeG, beforeFD := runtime.NumGoroutine(), m12V2FDCount()
			conn, err := net.DialTimeout("tcp", addr, time.Second)
			if err != nil { t.Fatal(err) }
			_ = conn.SetDeadline(time.Now().Add(700*time.Millisecond))
			tc.send(conn)
			if tc.linger > 0 { time.Sleep(tc.linger) }
			_ = conn.Close()

			deadline := time.Now().Add(2*time.Second)
			for time.Now().Before(deadline) {
				g, fd := runtime.NumGoroutine(), m12V2FDCount()
				if g <= beforeG+12 && (beforeFD < 0 || fd <= beforeFD+4) { break }
				time.Sleep(20*time.Millisecond)
			}
			afterG, afterFD := runtime.NumGoroutine(), m12V2FDCount()
			resp := m12V2Do(t, http.MethodGet, s.url+"/api/v1/health", nil, "")
			status := resp.StatusCode
			resp.Body.Close()
			t.Logf("CASE=%s STATUS=bounded GOROUTINES_BASELINE=%d GOROUTINES_PEAK_BOUND=%d GOROUTINES_RECOVERY=%d FD_BASELINE=%d FD_RECOVERY=%d POST_CASE_LEGITIMATE_REQUEST=%d", tc.name, beforeG, beforeG+12, afterG, beforeFD, afterFD, status)
			if afterG > beforeG+12 { t.Errorf("UNBOUNDED_GOROUTINE_GROWTH=YES before=%d after=%d", beforeG, afterG) }
			if beforeFD >= 0 && afterFD > beforeFD+4 { t.Errorf("UNBOUNDED_FD_GROWTH=YES before=%d after=%d", beforeFD, afterFD) }
			if status != http.StatusOK { t.Errorf("POST_ATTACK_SERVICE_RECOVERY=FAIL status=%d", status) }
		})
	}
}
