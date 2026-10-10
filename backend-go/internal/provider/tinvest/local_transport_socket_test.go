package tinvest

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/httpapi"
)

func TestRawDownstreamDisconnectRetainsLocalOperationOwnership(t *testing.T) {
	release := make(chan struct{})
	var released atomic.Bool
	defer func() {
		if released.CompareAndSwap(false, true) {
			close(release)
		}
	}()
	entered := make(chan struct{}, 4)
	var remoteActive atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		remoteActive.Add(1)
		defer remoteActive.Add(-1)
		entered <- struct{}{}
		select {
		case <-release:
		case <-r.Context().Done():
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"dividends":[]}`))
	}))
	defer upstream.Close()
	provider, err := newCorporateActionProvider(upstream.Client(), fixedClock{now: testNow}, testToken, upstream.URL+"/rest")
	if err != nil {
		t.Fatal(err)
	}
	app := httpapi.NewDevelopmentReplayWithCorporateActionProvider(nil, provider)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	serving := make(chan error, 1)
	go func() { serving <- app.Listener(listener) }()
	defer func() {
		if released.CompareAndSwap(false, true) {
			close(release)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = app.ShutdownWithContext(ctx)
		_ = listener.Close()
	}()
	for i := 0; i < 4; i++ {
		connection, err := net.DialTimeout("tcp", listener.Addr().String(), time.Second)
		if err != nil {
			t.Fatal(err)
		}
		_, err = fmt.Fprintf(connection, "GET /api/v1/corporate-actions/projection?instrumentId=SBER&from=2026-09-%02d&to=2026-09-30 HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n", i+1)
		if err != nil {
			_ = connection.Close()
			t.Fatal(err)
		}
		select {
		case <-entered:
		case <-time.After(2 * time.Second):
			_ = connection.Close()
			t.Fatal("provider transport did not start")
		}
		_ = connection.Close()
	}
	// Measure occupancy at the actual route boundary while operations are owned.
	slots := len(provider.semaphore)
	if slots > maxConcurrency {
		t.Fatalf("slots=%d bound=%d", slots, maxConcurrency)
	}
	client := &http.Client{Timeout: time.Second}
	response, err := client.Get("http://" + listener.Addr().String() + "/api/v1/corporate-actions/projection?instrumentId=SBER&from=2026-09-10&to=2026-09-30")
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()
	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("occupied provider status=%d", response.StatusCode)
	}
	propagation := "NO"
	if slots == 0 {
		propagation = "OBSERVED"
	}
	t.Logf("DOWNSTREAM_DISCONNECTS=4 PROVIDER_CONCURRENCY_SLOTS_IN_USE=%d UPSTREAM_HANDLERS_ACTIVE=%d DOWNSTREAM_DISCONNECT_CONTEXT_PROPAGATION=%s SUBSEQUENT_REQUEST_STATUS=%d", slots, remoteActive.Load(), propagation, response.StatusCode)
	if released.CompareAndSwap(false, true) {
		close(release)
	}
	deadline := time.Now().Add(2 * time.Second)
	for len(provider.semaphore) != 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if len(provider.semaphore) != 0 {
		t.Fatal("transport completion did not release slots")
	}
	response, err = client.Get("http://" + listener.Addr().String() + "/api/v1/corporate-actions/projection?instrumentId=SBER&from=2026-09-11&to=2026-09-30")
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("recovery status=%d", response.StatusCode)
	}
}
