package tinvest

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

type module10AuditClock struct{ now time.Time }

func (c module10AuditClock) Now() time.Time { return c.now }

type module10AuditRoundTrip func(*http.Request) (*http.Response, error)

func (fn module10AuditRoundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return fn(r) }

func module10AuditQuery() verticalslice.CorporateActionQuery {
	return verticalslice.CorporateActionQuery{
		InstrumentIDs: []string{"SBER"},
		From: "2026-01-01",
		To: "2026-12-31",
	}
}

func module10AuditProvider(t *testing.T, h http.Handler) (*Provider, *httptest.Server) {
	t.Helper()
	s := httptest.NewServer(h)
	p, err := newCorporateActionProvider(
		s.Client(),
		module10AuditClock{now: time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC)},
		"audit-readonly-token",
		s.URL+"/rest",
	)
	if err != nil {
		t.Fatal(err)
	}
	return p, s
}

func TestModule10SlowProviderLifecycle(t *testing.T) {
	t.Run("never_headers", func(t *testing.T) {
		release := make(chan struct{})
		p, s := module10AuditProvider(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			<-release
		}))
		defer func() { s.CloseClientConnections(); s.Close() }()

		before := runtime.NumGoroutine()
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
		defer cancel()
		_, err := p.CorporateActions(ctx, module10AuditQuery())
		close(release)
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("deadline not preserved: %v", err)
		}
		if len(p.semaphore) != 0 {
			t.Fatal("concurrency permit leaked")
		}
		runtime.GC()
		time.Sleep(50 * time.Millisecond)
		if after := runtime.NumGoroutine(); after > before+10 {
			t.Fatalf("goroutine growth before=%d after=%d", before, after)
		}
		t.Log("MODULE10_SLOW never_headers caller_deadline=true permit_released=true")
	})

	t.Run("body_never_completes", func(t *testing.T) {
		release := make(chan struct{})
		p, s := module10AuditProvider(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, "{\"dividends\":[")
			if f, ok := w.(http.Flusher); ok { f.Flush() }
			<-release
		}))
		defer func() { s.CloseClientConnections(); s.Close() }()

		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
		defer cancel()
		_, err := p.CorporateActions(ctx, module10AuditQuery())
		close(release)
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("deadline not preserved: %v", err)
		}
		if len(p.semaphore) != 0 {
			t.Fatal("concurrency permit leaked")
		}
		t.Log("MODULE10_SLOW body_never_completes caller_deadline=true permit_released=true")
	})

	t.Run("slow_chunked", func(t *testing.T) {
		p, s := module10AuditProvider(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			for i := 0; i < 100; i++ {
				select {
				case <-r.Context().Done():
					return
				case <-time.After(20 * time.Millisecond):
				}
				_, _ = io.WriteString(w, " ")
				if f, ok := w.(http.Flusher); ok { f.Flush() }
			}
		}))
		defer func() { s.CloseClientConnections(); s.Close() }()

		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
		defer cancel()
		_, err := p.CorporateActions(ctx, module10AuditQuery())
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("deadline not preserved: %v", err)
		}
		if len(p.semaphore) != 0 {
			t.Fatal("concurrency permit leaked")
		}
	})

	t.Run("connection_reset_mid_body", func(t *testing.T) {
		p, s := module10AuditProvider(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h, ok := w.(http.Hijacker)
			if !ok { t.Fatal("hijack unsupported") }
			conn, rw, err := h.Hijack()
			if err != nil { t.Fatal(err) }
			_, _ = rw.WriteString("HTTP/1.1 200 OK\r\nContent-Length: 1000\r\n\r\n{\"dividends\":[")
			_ = rw.Flush()
			_ = conn.Close()
		}))
		defer func() { s.CloseClientConnections(); s.Close() }()
		if _, err := p.CorporateActions(context.Background(), module10AuditQuery()); err == nil {
			t.Fatal("reset mid-body accepted")
		}
		if len(p.semaphore) != 0 {
			t.Fatal("concurrency permit leaked")
		}
	})

	t.Run("tcp_refusal_and_dns_like_failure", func(t *testing.T) {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil { t.Fatal(err) }
		addr := ln.Addr().String()
		_ = ln.Close()
		p, err := newCorporateActionProvider(&http.Client{}, module10AuditClock{now: time.Now().UTC()}, "audit-readonly-token", "http://"+addr+"/rest")
		if err != nil { t.Fatal(err) }
		if _, err := p.CorporateActions(context.Background(), module10AuditQuery()); err == nil {
			t.Fatal("TCP refusal accepted")
		}

		client := &http.Client{Transport: module10AuditRoundTrip(func(*http.Request) (*http.Response, error) {
			return nil, &net.DNSError{Err: "no such host", Name: "audit.invalid", IsNotFound: true}
		})}
		p, err = newCorporateActionProvider(client, module10AuditClock{now: time.Now().UTC()}, "audit-readonly-token", "http://audit.invalid/rest")
		if err != nil { t.Fatal(err) }
		if _, err := p.CorporateActions(context.Background(), module10AuditQuery()); err == nil {
			t.Fatal("DNS-like failure accepted")
		}
	})
}

func TestModule10RecoveryAndNoRetry(t *testing.T) {
	var slow atomic.Bool
	slow.Store(true)
	slowRelease := make(chan struct{})
	var calls atomic.Int64
	p, s := module10AuditProvider(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if slow.Load() {
			<-slowRelease
			return
		}
		_, _ = io.WriteString(w, "{\"dividends\":[]}")
	}))
	defer func() { s.CloseClientConnections(); s.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	_, err := p.CorporateActions(ctx, module10AuditQuery())
	cancel()
	close(slowRelease)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("timeout error=%v", err)
	}
	slow.Store(false)
	if _, err := p.CorporateActions(context.Background(), module10AuditQuery()); err != nil {
		t.Fatalf("recovery failed: %v", err)
	}

	for _, status := range []int{429, 500, 502, 503, 504} {
		t.Run(fmt.Sprintf("status_%d", status), func(t *testing.T) {
			var attempts atomic.Int64
			p, s := module10AuditProvider(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				attempts.Add(1)
				w.WriteHeader(status)
			}))
			defer func() { s.CloseClientConnections(); s.Close() }()
			_, err := p.CorporateActions(context.Background(), module10AuditQuery())
			if err == nil { t.Fatal("expected provider error") }
			if attempts.Load() != 1 { t.Fatalf("automatic retry observed: %d", attempts.Load()) }
		})
	}
}

func TestModule10DistributedProviderBudget(t *testing.T) {
	for _, instances := range []int{1, 2, 4} {
		t.Run(fmt.Sprintf("instances_%d", instances), func(t *testing.T) {
			var upstream atomic.Int64
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				upstream.Add(1)
				_, _ = io.WriteString(w, "{}")
			}))
			defer func() { s.CloseClientConnections(); s.Close() }()

			for n := 0; n < instances; n++ {
				p, err := newCorporateActionProvider(s.Client(), module10AuditClock{now: time.Now().UTC()}, "audit-readonly-token", s.URL+"/rest")
				if err != nil { t.Fatal(err) }
				for i := 0; i < maxRequestsPerMinute+10; i++ {
					_, _ = p.post(context.Background(), getDividendsMethod, corporateActionRequest{})
				}
			}
			want := int64(instances * maxRequestsPerMinute)
			if got := upstream.Load(); got != want {
				t.Fatalf("aggregate admitted=%d want=%d", got, want)
			}
			t.Logf("MODULE10_BUDGET instances=%d observed=%d nominal_per_process=%d", instances, upstream.Load(), maxRequestsPerMinute)
		})
	}
}

func TestModule10ConcurrencyExhaustion(t *testing.T) {
	for _, c := range []int{1, maxConcurrency, maxConcurrency + 1, 2 * maxConcurrency, 10 * maxConcurrency} {
		t.Run(fmt.Sprintf("C_%d", c), func(t *testing.T) {
			var active, maxActive, upstream atomic.Int64
			release := make(chan struct{})
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				cur := active.Add(1)
				upstream.Add(1)
				for {
					old := maxActive.Load()
					if cur <= old || maxActive.CompareAndSwap(old, cur) { break }
				}
				defer active.Add(-1)
				select {
				case <-release:
				case <-r.Context().Done():
					return
				}
				_, _ = io.WriteString(w, "{}")
			}))
			defer func() { s.CloseClientConnections(); s.Close() }()
			p, err := newCorporateActionProvider(s.Client(), module10AuditClock{now: time.Now().UTC()}, "audit-readonly-token", s.URL+"/rest")
			if err != nil { t.Fatal(err) }

			var wg sync.WaitGroup
			wg.Add(c)
			for i := 0; i < c; i++ {
				go func() {
					defer wg.Done()
					_, _ = p.post(context.Background(), getDividendsMethod, corporateActionRequest{})
				}()
			}
			time.Sleep(60 * time.Millisecond)
			close(release)
			wg.Wait()

			if maxActive.Load() > maxConcurrency {
				t.Fatalf("max active=%d limit=%d", maxActive.Load(), maxConcurrency)
			}
			if len(p.semaphore) != 0 {
				t.Fatal("permit leak")
			}
			if _, err := p.post(context.Background(), getDividendsMethod, corporateActionRequest{}); err != nil {
				t.Fatalf("capacity did not recover: %v", err)
			}
			t.Logf("MODULE10_CONCURRENCY C=%d upstream=%d max_active=%d", c, upstream.Load(), maxActive.Load())
		})
	}
}

func TestModule10ParserAndProviderDataAbuse(t *testing.T) {
	cases := []struct {
		name string
		body string
		ok   bool
	}{
		{"malformed", "{", false},
		{"truncated", "{\"dividends\":[", false},
		{"duplicate_fields", "{\"dividends\":[],\"dividends\":[]}", true},
		{"unknown_100kb", "{\"dividends\":[],\"padding\":\"" + strings.Repeat("a", 100*1024) + "\"}", true},
		{"huge_enum", "{\"dividends\":[{\"dividendType\":\"" + strings.Repeat("X", 100*1024) + "\"}]}", false},
		{"invalid_date", "{\"dividends\":[{\"dividendType\":\"Regular Cash\",\"recordDate\":\"bad\"}]}", false},
		{"negative_money", "{\"dividends\":[{\"dividendType\":\"Regular Cash\",\"dividendNet\":{\"currency\":\"rub\",\"units\":\"-1\",\"nano\":0}}]}", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p, s := module10AuditProvider(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.WriteString(w, tc.body)
			}))
			defer func() { s.CloseClientConnections(); s.Close() }()
			_, err := p.CorporateActions(context.Background(), module10AuditQuery())
			if tc.ok && err != nil { t.Fatalf("unexpected error: %v", err) }
			if !tc.ok && err == nil { t.Fatal("hostile provider response accepted") }
		})
	}
	for _, size := range []int{1 << 20, 10 << 20} {
		t.Run(fmt.Sprintf("oversized_%d", size), func(t *testing.T) {
			body := "{\"dividends\":[],\"padding\":\"" + strings.Repeat("a", size) + "\"}"
			p, s := module10AuditProvider(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.WriteString(w, body)
			}))
			defer func() { s.CloseClientConnections(); s.Close() }()
			if _, err := p.CorporateActions(context.Background(), module10AuditQuery()); err == nil {
				t.Fatal("oversized body accepted")
			}
		})
	}
}

func TestModule10DedupRedirectStampedeAndCircuitBound(t *testing.T) {
	duplicate := "{\"dividends\":[" +
		"{\"dividendType\":\"Regular Cash\",\"recordDate\":\"2026-10-10T00:00:00Z\",\"dividendNet\":{\"currency\":\"rub\",\"units\":\"1\",\"nano\":0}}," +
		"{\"dividendType\":\"Regular Cash\",\"recordDate\":\"2026-10-10T00:00:00Z\",\"dividendNet\":{\"currency\":\"rub\",\"units\":\"1\",\"nano\":0}}" +
		"]}"
	p, s := module10AuditProvider(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, duplicate)
	}))
	events, err := p.CorporateActions(context.Background(), module10AuditQuery())
	s.Close()
	if err != nil || len(events) != 1 {
		t.Fatalf("dedup failed len=%d err=%v", len(events), err)
	}

	var redirectHits atomic.Int64
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirectHits.Add(1)
		t.Errorf("redirect followed with auth=%q", r.Header.Get("Authorization"))
	}))
	defer target.Close()
	p, source := module10AuditProvider(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusFound)
	}))
	_, _ = p.CorporateActions(context.Background(), module10AuditQuery())
	source.Close()
	if redirectHits.Load() != 0 {
		t.Fatal("redirect target reached")
	}

	var deadCalls atomic.Int64
	p, dead := module10AuditProvider(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deadCalls.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer dead.Close()
	for i := 0; i < 100; i++ {
		_, _ = p.CorporateActions(context.Background(), module10AuditQuery())
	}
	if deadCalls.Load() != maxRequestsPerMinute {
		t.Fatalf("dead-provider calls=%d want=%d", deadCalls.Load(), maxRequestsPerMinute)
	}
	t.Logf("MODULE10_CIRCUIT_BREAKER incoming=100 upstream=%d circuit_breaker=absent local_budget_bound=true", deadCalls.Load())

	for _, clients := range []int{10, 50, 100} {
		t.Run(fmt.Sprintf("stampede_%d", clients), func(t *testing.T) {
			var upstream atomic.Int64
			release := make(chan struct{})
			p, s := module10AuditProvider(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				upstream.Add(1)
				select {
				case <-release:
				case <-r.Context().Done():
					return
				}
				_, _ = io.WriteString(w, "{\"dividends\":[]}")
			}))
			defer func() { s.CloseClientConnections(); s.Close() }()

			var wg sync.WaitGroup
			wg.Add(clients)
			for i := 0; i < clients; i++ {
				go func() {
					defer wg.Done()
					_, _ = p.CorporateActions(context.Background(), module10AuditQuery())
				}()
			}
			time.Sleep(60 * time.Millisecond)
			close(release)
			wg.Wait()
			if upstream.Load() > maxConcurrency {
				t.Fatalf("stampede upstream=%d", upstream.Load())
			}
			t.Logf("MODULE10_STAMPEDE clients=%d upstream=%d coalesced=0 fail_fast_bound=%d", clients, upstream.Load(), maxConcurrency)
		})
	}
}

func TestModule10FailureIsolationAndResourceLeak(t *testing.T) {
	beforeG := runtime.NumGoroutine()
	beforeFD := -1
	if entries, err := os.ReadDir("/proc/self/fd"); err == nil { beforeFD = len(entries) }

	var mode atomic.Int64
	timeoutRelease := make(chan struct{})
	p, s := module10AuditProvider(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch mode.Load() {
		case 0, 5:
			_, _ = io.WriteString(w, "{\"dividends\":[]}")
		case 1, 2:
			w.WriteHeader(http.StatusServiceUnavailable)
		case 3:
			<-timeoutRelease
		case 4:
			_, _ = io.WriteString(w, "{")
		}
	}))
	for i := int64(0); i < 6; i++ {
		mode.Store(i)
		ctx := context.Background()
		var cancel context.CancelFunc
		if i == 3 { ctx, cancel = context.WithTimeout(ctx, 60*time.Millisecond) }
		_, err := p.CorporateActions(ctx, module10AuditQuery())
		if cancel != nil {
			cancel()
			close(timeoutRelease)
		}
		if (i == 0 || i == 5) && err != nil { t.Fatalf("success phase %d: %v", i, err) }
		if i > 0 && i < 5 && err == nil { t.Fatalf("failure phase %d succeeded", i) }
	}
	s.CloseClientConnections()
	s.Close()
	if len(p.semaphore) != 0 { t.Fatal("permit stuck after failure isolation sequence") }

	client := &http.Client{Transport: module10AuditRoundTrip(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("synthetic network failure")
	})}
	p2, err := newCorporateActionProvider(client, module10AuditClock{now: time.Now().UTC()}, "audit-readonly-token", "http://audit.invalid/rest")
	if err != nil { t.Fatal(err) }
	p2.requestGate = newRequestBudget(20000, time.Minute, time.Now)
	for i := 0; i < 10000; i++ {
		_, _ = p2.CorporateActions(context.Background(), module10AuditQuery())
	}
	if len(p2.semaphore) != 0 { t.Fatal("permit leaked in 10000 failure campaign") }

	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	afterG := runtime.NumGoroutine()
	afterFD := -1
	if entries, err := os.ReadDir("/proc/self/fd"); err == nil { afterFD = len(entries) }
	if afterG > beforeG+20 { t.Fatalf("goroutine growth before=%d after=%d", beforeG, afterG) }
	if beforeFD >= 0 && afterFD > beforeFD+20 { t.Fatalf("fd growth before=%d after=%d", beforeFD, afterFD) }
	t.Logf("MODULE10_RESOURCE_LEAK before_g=%d after_g=%d before_fd=%d after_fd=%d failures=10000", beforeG, afterG, beforeFD, afterFD)
}


func TestModule10RawTCPCancellationProof(t *testing.T) {
	for _, tc := range []struct {
		name        string
		partialBody bool
	}{
		{name: "never_headers"},
		{name: "partial_body", partialBody: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ln, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			defer ln.Close()

			connectionResult := make(chan error, 1)
			go func() {
				conn, acceptErr := ln.Accept()
				if acceptErr != nil {
					connectionResult <- acceptErr
					return
				}
				defer conn.Close()

				reader := bufio.NewReader(conn)
				request, readErr := http.ReadRequest(reader)
				if readErr != nil {
					connectionResult <- readErr
					return
				}
				if request.Body != nil {
					_, _ = io.Copy(io.Discard, request.Body)
					_ = request.Body.Close()
				}
				if tc.partialBody {
					_, _ = io.WriteString(conn, "HTTP/1.1 200 OK\r\nContent-Length: 1000\r\n\r\n{\"dividends\":[")
				}
				_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
				_, readErr = reader.ReadByte()
				connectionResult <- readErr
			}()

			p, err := newCorporateActionProvider(
				&http.Client{},
				module10AuditClock{now: time.Now().UTC()},
				"audit-readonly-token",
				"http://"+ln.Addr().String()+"/rest",
			)
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
			defer cancel()
			_, callErr := p.CorporateActions(ctx, module10AuditQuery())
			if !errors.Is(callErr, context.DeadlineExceeded) {
				t.Fatalf("deadline not preserved: %v", callErr)
			}

			select {
			case readErr := <-connectionResult:
				if readErr == nil {
					t.Fatal("upstream connection produced unexpected post-request byte")
				}
				if networkErr, ok := readErr.(net.Error); ok && networkErr.Timeout() {
					t.Fatalf("upstream TCP connection remained open after caller cancellation: %v", readErr)
				}
				t.Logf("MODULE10_RAW_TCP case=%s caller_deadline=true upstream_connection_closed=true read_error=%T", tc.name, readErr)
			case <-time.After(3 * time.Second):
				t.Fatal("raw TCP cancellation witness did not complete")
			}
		})
	}
}

func TestModule10TokenHandling(t *testing.T) {
	secret := "SECRET-AUDIT-TOKEN-123"
	client := &http.Client{Transport: module10AuditRoundTrip(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("transport detail includes " + secret)
	})}
	p, err := newCorporateActionProvider(client, module10AuditClock{now: time.Now().UTC()}, secret, "http://audit.invalid/rest")
	if err != nil { t.Fatal(err) }
	_, err = p.CorporateActions(context.Background(), module10AuditQuery())
	if err == nil { t.Fatal("expected transport error") }
	if strings.Contains(err.Error(), secret) { t.Fatalf("secret echoed: %v", err) }

	for _, token := range []string{"", " token", "token ", "tok\nen"} {
		if _, err := newCorporateActionProvider(&http.Client{}, module10AuditClock{now: time.Now().UTC()}, token, "http://audit.invalid/rest"); err == nil {
			t.Fatalf("invalid token accepted: %q", token)
		}
	}
}
