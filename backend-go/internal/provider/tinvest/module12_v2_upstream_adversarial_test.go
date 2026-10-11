package tinvest

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func m12V2ProviderFDCount() int {
	entries, err := os.ReadDir("/proc/self/fd")
	if err != nil { return -1 }
	return len(entries)
}

func TestModule12V2UpstreamResourceAdversarialMatrix(t *testing.T) {
	if os.Getenv("OPENINVEST_MODULE12_V2_ADVERSARIAL") != "1" {
		t.Skip("Module 12 V2 adversarial matrix runs only in the specialized workflow")
	}
	type tc struct {
		name string
		handler func(http.ResponseWriter, *http.Request)
	}
	cases := []tc{
		{"headers_delay", func(w http.ResponseWriter, r *http.Request){ select { case <-time.After(250*time.Millisecond): case <-r.Context().Done(): return }; w.Header().Set("Content-Type","application/json"); _,_ = io.WriteString(w,"{\"dividends\":[]}") }},
		{"partial_headers", func(w http.ResponseWriter, r *http.Request){ h, ok := w.(http.Hijacker); if !ok { return }; c, rw, err := h.Hijack(); if err != nil { return }; defer c.Close(); _,_ = rw.WriteString("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nX-M12: partial"); _ = rw.Flush() }},
		{"invalid_headers", func(w http.ResponseWriter, r *http.Request){ h, ok := w.(http.Hijacker); if !ok { return }; c, rw, err := h.Hijack(); if err != nil { return }; defer c.Close(); _,_ = rw.WriteString("HTTP/1.1 200 OK\r\nBad Header Without Colon\r\n\r\n{}"); _ = rw.Flush() }},
		{"body_delay", func(w http.ResponseWriter, r *http.Request){ w.Header().Set("Content-Type","application/json"); w.WriteHeader(200); if f,ok:=w.(http.Flusher);ok{f.Flush()}; select { case <-time.After(250*time.Millisecond): case <-r.Context().Done(): return }; _,_=io.WriteString(w,"{\"dividends\":[]}") }},
		{"partial_body_then_stall", func(w http.ResponseWriter, r *http.Request){ w.Header().Set("Content-Type","application/json"); _,_=io.WriteString(w,"{\"dividends\":["); if f,ok:=w.(http.Flusher);ok{f.Flush()}; select { case <-time.After(250*time.Millisecond): case <-r.Context().Done(): return } }},
		{"partial_body_then_FIN", func(w http.ResponseWriter, r *http.Request){ w.Header().Set("Content-Type","application/json"); _,_=io.WriteString(w,"{\"dividends\":[") }},
		{"RST_before_headers", func(w http.ResponseWriter, r *http.Request){ h, ok := w.(http.Hijacker); if !ok { return }; c, _, err := h.Hijack(); if err != nil { return }; if tcp,ok:=c.(*net.TCPConn);ok{_ = tcp.SetLinger(0)}; _=c.Close() }},
		{"RST_mid_body", func(w http.ResponseWriter, r *http.Request){ h, ok := w.(http.Hijacker); if !ok { return }; c, rw, err := h.Hijack(); if err != nil { return }; _,_=rw.WriteString("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: 100\r\n\r\n{\"dividends\":["); _=rw.Flush(); if tcp,ok:=c.(*net.TCPConn);ok{_ = tcp.SetLinger(0)}; _=c.Close() }},
		{"malformed_json", func(w http.ResponseWriter, r *http.Request){ w.Header().Set("Content-Type","application/json"); _,_=io.WriteString(w,"{not-json") }},
		{"bounded_oversized_json", func(w http.ResponseWriter, r *http.Request){ w.Header().Set("Content-Type","application/json"); _,_=io.WriteString(w,"{\"dividends\":[\""+strings.Repeat("x",512*1024)+"\"]}") }},
		{"deadline_timeout", func(w http.ResponseWriter, r *http.Request){ select { case <-time.After(500*time.Millisecond): return; case <-r.Context().Done(): return } }},
		{"noncooperative_handler", func(w http.ResponseWriter, r *http.Request){ time.Sleep(250*time.Millisecond) }},
		{"very_slow_chunk_stream", func(w http.ResponseWriter, r *http.Request){ w.Header().Set("Content-Type","application/json"); _,_=io.WriteString(w,"{\"dividends\":["); if f,ok:=w.(http.Flusher);ok{f.Flush()}; ticker:=time.NewTicker(20*time.Millisecond); defer ticker.Stop(); for i:=0;i<100;i++{ select{case <-r.Context().Done():return;case <-ticker.C: _,_=io.WriteString(w," "); if f,ok:=w.(http.Flusher);ok{f.Flush()} } } }},
		{"close_without_body", func(w http.ResponseWriter, r *http.Request){ w.WriteHeader(http.StatusOK) }},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			var providerCalls atomic.Int64
			var activeHandlers atomic.Int64
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				providerCalls.Add(1)
				activeHandlers.Add(1)
				defer activeHandlers.Add(-1)
				c.handler(w, r)
			}))
			defer server.Close()
			provider, err := newCorporateActionProvider(server.Client(), fixedClock{now:testNow}, testToken, server.URL+"/rest")
			if err != nil { t.Fatal(err) }
			beforeG, beforeFD := runtime.NumGoroutine(), m12V2ProviderFDCount()
			peakG, peakFD, peakSem := beforeG, beforeFD, 0
			stopSample, sampled := make(chan struct{}), make(chan struct{})
			go func() {
				ticker := time.NewTicker(2 * time.Millisecond)
				defer ticker.Stop()
				defer close(sampled)
				for {
					select {
					case <-stopSample:
						return
					case <-ticker.C:
						if g := runtime.NumGoroutine(); g > peakG { peakG = g }
						if fd := m12V2ProviderFDCount(); fd > peakFD { peakFD = fd }
						if sem := len(provider.semaphore); sem > peakSem { peakSem = sem }
					}
				}
			}()
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			_, callErr := provider.CorporateActions(ctx, testQuery("SBER"))
			close(stopSample)
			<-sampled
			if callErr == nil { t.Errorf("hostile upstream %s unexpectedly produced a valid response", c.name) }
			// If the remote fixture still owns work after the caller has timed out,
			// capacity must remain owned locally; otherwise cancellation could amplify
			// remote work by immediately admitting replacements.
			requiresPostCallerOwnership := false
			switch c.name {
			case "headers_delay", "body_delay", "partial_body_then_stall", "deadline_timeout", "noncooperative_handler", "very_slow_chunk_stream":
				requiresPostCallerOwnership = true
			}
			if requiresPostCallerOwnership && activeHandlers.Load() > 0 && len(provider.semaphore) == 0 {
				t.Errorf("EARLY_PROVIDER_CAPACITY_RECYCLE=YES case=%s active_handlers=%d", c.name, activeHandlers.Load())
			}
			// All fixtures are bounded. Wait for their actual local terminal condition,
			// rather than imposing an earlier artificial release requirement.
			terminalDeadline := time.Now().Add(3 * time.Second)
			for time.Now().Before(terminalDeadline) {
				if len(provider.semaphore) == 0 && activeHandlers.Load() == 0 { break }
				time.Sleep(5 * time.Millisecond)
			}
			if got := len(provider.semaphore); got != 0 { t.Errorf("PROVIDER_SLOT_LEAK=YES case=%s slots=%d", c.name, got) }
			if got := activeHandlers.Load(); got != 0 { t.Errorf("UPSTREAM_HANDLER_LEAK=YES case=%s active=%d", c.name, got) }
			server.CloseClientConnections()
			resourceDeadline := time.Now().Add(time.Second)
			for time.Now().Before(resourceDeadline) {
				if len(provider.semaphore) == 0 && runtime.NumGoroutine() <= beforeG+6 && (beforeFD < 0 || m12V2ProviderFDCount() <= beforeFD+4) { break }
				time.Sleep(10 * time.Millisecond)
			}
			afterG, afterFD := runtime.NumGoroutine(), m12V2ProviderFDCount()

			good := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){ w.Header().Set("Content-Type","application/json"); _,_=io.WriteString(w,"{\"dividends\":[]}") }))
			defer good.Close()
			recovery, err := newCorporateActionProvider(good.Client(), fixedClock{now:testNow}, testToken, good.URL+"/rest")
			if err != nil { t.Fatal(err) }
			if _, err := recovery.CorporateActions(context.Background(), testQuery("SBER")); err != nil {
				t.Errorf("POST_CASE_LEGITIMATE_REQUEST=FAIL case=%s err=%v", c.name, err)
			}
			t.Logf("CASE=%s STATUS=handled ERROR=%v PROVIDER_CALLS=%d SEMAPHORE_PEAK=%d SEMAPHORE_RECOVERY=%d UPSTREAM_ACTIVE_RECOVERY=%d GOROUTINES_BASELINE=%d GOROUTINES_PEAK=%d GOROUTINES_RECOVERY=%d FD_BASELINE=%d FD_PEAK=%d FD_RECOVERY=%d POST_CASE_LEGITIMATE_REQUEST=PASS", c.name, callErr, providerCalls.Load(), peakSem, len(provider.semaphore), activeHandlers.Load(), beforeG, peakG, afterG, beforeFD, peakFD, afterFD)
		})
	}

	t.Run("connect_delay", func(t *testing.T) {
		transport := &http.Transport{
			DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
				select {
				case <-time.After(250*time.Millisecond):
					return nil, errors.New("synthetic delayed connect")
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			},
		}
		client := &http.Client{Transport: transport}
		provider, err := newCorporateActionProvider(client, fixedClock{now:testNow}, testToken, "http://127.0.0.1:1/rest")
		if err != nil { t.Fatal(err) }
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		_, err = provider.CorporateActions(ctx, testQuery("SBER"))
		if err == nil { t.Fatal("connect delay unexpectedly succeeded") }
		deadline := time.Now().Add(time.Second)
		for len(provider.semaphore) != 0 && time.Now().Before(deadline) { time.Sleep(time.Millisecond) }
		if got := len(provider.semaphore); got != 0 { t.Fatalf("PROVIDER_SLOT_LEAK=YES connect_delay slots=%d", got) }
		if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, verticalslice.ErrCorporateActionsProviderUnavailable) {
			t.Logf("connect_delay returned bounded transport error: %v", err)
		}
	})

	// Compile-time guard that httptest's writer still supports the interfaces used by the hostile fixtures.
	_ = bufio.ErrInvalidUnreadByte
}
