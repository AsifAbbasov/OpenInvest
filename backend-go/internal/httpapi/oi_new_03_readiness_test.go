package httpapi

import (
	"context"
	"errors"
	"io"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

type oiNew03ReadyStore struct {
	verticalslice.Store
	mu                 sync.Mutex
	pingCalls          int
	stage371Calls      int
	stage376Calls      int
	pingErr            error
	blockPing          bool
	lastPingContextErr error
}

func (s *oiNew03ReadyStore) Ping(ctx context.Context) error {
	s.mu.Lock()
	s.pingCalls++
	block := s.blockPing
	err := s.pingErr
	s.mu.Unlock()
	if !block {
		return err
	}
	<-ctx.Done()
	s.mu.Lock()
	s.lastPingContextErr = ctx.Err()
	s.mu.Unlock()
	return ctx.Err()
}
func (s *oiNew03ReadyStore) Stage371Ready(context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stage371Calls++
	return nil
}
func (s *oiNew03ReadyStore) Stage376Ready(context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stage376Calls++
	return nil
}
func (s *oiNew03ReadyStore) counts() (int, int, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pingCalls, s.stage371Calls, s.stage376Calls, s.lastPingContextErr
}
func oiNew03App(store *oiNew03ReadyStore) *fiber.App {
	return newApp(&API{service: verticalslice.NewService(store, verticalslice.SystemClock{})})
}

func TestOINew03ReadyUsesCheapPingOnly(t *testing.T) {
	s := &oiNew03ReadyStore{}
	app := oiNew03App(s)
	resp, err := app.Test(httptest.NewRequest("GET", "/api/v1/ready", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status=%d", resp.StatusCode)
	}
	p, a, b, _ := s.counts()
	if p != 1 || a != 0 || b != 0 {
		t.Fatalf("calls=%d/%d/%d", p, a, b)
	}
}
func TestOINew03RepeatedReadyNeverRunsDeepIntegrity(t *testing.T) {
	s := &oiNew03ReadyStore{}
	app := oiNew03App(s)
	for i := 0; i < 100; i++ {
		resp, err := app.Test(httptest.NewRequest("GET", "/api/v1/ready", nil))
		if err != nil {
			t.Fatalf("request %d: %v", i, err)
		}
		resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Fatalf("request %d status=%d", i, resp.StatusCode)
		}
	}
	p, a, b, _ := s.counts()
	if p != 100 || a != 0 || b != 0 {
		t.Fatalf("calls=%d/%d/%d", p, a, b)
	}
}
func TestOINew03ReadyPingFailureReturnsGeneric503(t *testing.T) {
	s := &oiNew03ReadyStore{pingErr: errors.New("secret postgres host db.internal table investment.transaction_entries")}
	app := oiNew03App(s)
	resp, err := app.Test(httptest.NewRequest("GET", "/api/v1/ready", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 503 {
		t.Fatalf("status=%d body=%s", resp.StatusCode, body)
	}
	text := string(body)
	if !strings.Contains(text, "SERVICE_NOT_READY") {
		t.Fatalf("missing generic code: %s", text)
	}
	if strings.Contains(text, "db.internal") || strings.Contains(text, "transaction_entries") {
		t.Fatalf("internal error leaked: %s", text)
	}
	p, a, b, _ := s.counts()
	if p != 1 || a != 0 || b != 0 {
		t.Fatalf("calls=%d/%d/%d", p, a, b)
	}
}
func TestOINew03ReadyPingDeadlineIsBounded(t *testing.T) {
	s := &oiNew03ReadyStore{blockPing: true}
	app := oiNew03App(s)
	start := time.Now()
	resp, err := app.Test(httptest.NewRequest("GET", "/api/v1/ready", nil), fiber.TestConfig{Timeout: 5 * time.Second, FailOnTimeout: true})
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 503 {
		t.Fatalf("status=%d", resp.StatusCode)
	}
	p, a, b, cerr := s.counts()
	if p != 1 || a != 0 || b != 0 {
		t.Fatalf("calls=%d/%d/%d", p, a, b)
	}
	if !errors.Is(cerr, context.DeadlineExceeded) {
		t.Fatalf("ctx err=%v", cerr)
	}
	if elapsed < time.Second || elapsed >= 5*time.Second {
		t.Fatalf("readiness elapsed outside broad bound: %s", elapsed)
	}
}
func TestOINew03HealthRemainsLivenessOnly(t *testing.T) {
	s := &oiNew03ReadyStore{pingErr: errors.New("db down")}
	app := oiNew03App(s)
	resp, err := app.Test(httptest.NewRequest("GET", "/api/v1/health", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status=%d", resp.StatusCode)
	}
	p, a, b, _ := s.counts()
	if p != 0 || a != 0 || b != 0 {
		t.Fatalf("health touched db: %d/%d/%d", p, a, b)
	}
}
