package main

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

type oiNew03StartupStore struct {
	verticalslice.Store
	stage371Calls  int
	stage376Calls  int
	stage371Err    error
	stage376Err    error
	sawDeadline371 bool
	sawDeadline376 bool
	remaining371   time.Duration
}

func (s *oiNew03StartupStore) Ping(context.Context) error { return nil }
func (s *oiNew03StartupStore) Stage371Ready(ctx context.Context) error {
	s.stage371Calls++
	if d, ok := ctx.Deadline(); ok {
		s.sawDeadline371 = true
		s.remaining371 = time.Until(d)
	}
	return s.stage371Err
}
func (s *oiNew03StartupStore) Stage376Ready(ctx context.Context) error {
	s.stage376Calls++
	if _, ok := ctx.Deadline(); ok {
		s.sawDeadline376 = true
	}
	return s.stage376Err
}

func TestOINew03RuntimeIntegrityStartupValidationRunsBeforeServing(t *testing.T) {
	s := &oiNew03StartupStore{}
	svc, err := newValidatedRuntimeService(s)
	if err != nil {
		t.Fatal(err)
	}
	if svc == nil {
		t.Fatal("nil service")
	}
	if s.stage371Calls != 1 || s.stage376Calls != 1 {
		t.Fatalf("calls=%d/%d", s.stage371Calls, s.stage376Calls)
	}
	if !s.sawDeadline371 || !s.sawDeadline376 {
		t.Fatal("startup validation missing deadline")
	}
	if s.remaining371 <= 0 || s.remaining371 > runtimeIntegrityStartupTimeout {
		t.Fatalf("remaining=%s", s.remaining371)
	}
}
func TestOINew03RuntimeIntegrityStartupFailsClosedOnStage371(t *testing.T) {
	boom := errors.New("stage371 failure")
	s := &oiNew03StartupStore{stage371Err: boom}
	svc, err := newValidatedRuntimeService(s)
	if svc != nil || !errors.Is(err, boom) {
		t.Fatalf("svc=%v err=%v", svc, err)
	}
	if s.stage371Calls != 1 || s.stage376Calls != 0 {
		t.Fatalf("calls=%d/%d", s.stage371Calls, s.stage376Calls)
	}
}
func TestOINew03RuntimeIntegrityStartupFailsClosedOnStage376(t *testing.T) {
	boom := errors.New("stage376 failure")
	s := &oiNew03StartupStore{stage376Err: boom}
	svc, err := newValidatedRuntimeService(s)
	if svc != nil || !errors.Is(err, boom) {
		t.Fatalf("svc=%v err=%v", svc, err)
	}
	if s.stage371Calls != 1 || s.stage376Calls != 1 {
		t.Fatalf("calls=%d/%d", s.stage371Calls, s.stage376Calls)
	}
}

func TestOINew03RuntimeIntegrityStartupAgainstPostgres(t *testing.T) {
	dsn := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if dsn == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}
	store, err := postgres.Open(dsn)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			t.Errorf("close postgres: %v", err)
		}
	}()
	svc, err := newValidatedRuntimeService(store)
	if err != nil {
		t.Fatalf("validate migrated postgres runtime integrity: %v", err)
	}
	if svc == nil {
		t.Fatal("nil validated service")
	}
}
