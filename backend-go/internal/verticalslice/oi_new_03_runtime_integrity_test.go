package verticalslice

import (
	"context"
	"errors"
	"testing"
)

type oiNew03IntegrityStore struct {
	Store
	pingCalls     int
	stage371Calls int
	stage376Calls int
	oiNew04Calls  int
	pingErr       error
	stage371Err   error
	stage376Err   error
	oiNew04Err    error
}

func (s *oiNew03IntegrityStore) Ping(context.Context) error { s.pingCalls++; return s.pingErr }
func (s *oiNew03IntegrityStore) Stage371Ready(context.Context) error {
	s.stage371Calls++
	return s.stage371Err
}
func (s *oiNew03IntegrityStore) Stage376Ready(context.Context) error {
	s.stage376Calls++
	return s.stage376Err
}
func (s *oiNew03IntegrityStore) StageOINew04Ready(context.Context) error {
	s.oiNew04Calls++
	return s.oiNew04Err
}

type oiNew03CheapOnlyStore struct{ Store }

func (s *oiNew03CheapOnlyStore) Ping(context.Context) error { return nil }

func TestOINew03ServiceReadyUsesCheapPingOnly(t *testing.T) {
	store := &oiNew03IntegrityStore{}
	svc := NewService(store, SystemClock{})
	if err := svc.Ready(context.Background()); err != nil {
		t.Fatalf("ready: %v", err)
	}
	if store.pingCalls != 1 || store.stage371Calls != 0 || store.stage376Calls != 0 || store.oiNew04Calls != 0 {
		t.Fatalf("unexpected calls ping=%d stage371=%d stage376=%d oiNew04=%d", store.pingCalls, store.stage371Calls, store.stage376Calls, store.oiNew04Calls)
	}
}
func TestOINew03RuntimeIntegrityRunsStage371AndStage376(t *testing.T) {
	store := &oiNew03IntegrityStore{}
	svc := NewService(store, SystemClock{})
	if err := svc.ValidateRuntimeIntegrity(context.Background()); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if store.stage371Calls != 1 || store.stage376Calls != 1 || store.oiNew04Calls != 1 {
		t.Fatalf("unexpected calls: %d %d %d", store.stage371Calls, store.stage376Calls, store.oiNew04Calls)
	}
}
func TestOINew03RuntimeIntegrityFailsClosedOnStage371(t *testing.T) {
	boom := errors.New("stage371 failed")
	store := &oiNew03IntegrityStore{stage371Err: boom}
	svc := NewService(store, SystemClock{})
	if err := svc.ValidateRuntimeIntegrity(context.Background()); !errors.Is(err, boom) {
		t.Fatalf("expected stage371 failure, got %v", err)
	}
	if store.stage371Calls != 1 || store.stage376Calls != 0 || store.oiNew04Calls != 0 {
		t.Fatalf("unexpected calls: %d %d %d", store.stage371Calls, store.stage376Calls, store.oiNew04Calls)
	}
}
func TestOINew03RuntimeIntegrityFailsClosedOnStage376(t *testing.T) {
	boom := errors.New("stage376 failed")
	store := &oiNew03IntegrityStore{stage376Err: boom}
	svc := NewService(store, SystemClock{})
	if err := svc.ValidateRuntimeIntegrity(context.Background()); !errors.Is(err, boom) {
		t.Fatalf("expected stage376 failure, got %v", err)
	}
	if store.stage371Calls != 1 || store.stage376Calls != 1 || store.oiNew04Calls != 0 {
		t.Fatalf("unexpected calls: %d %d %d", store.stage371Calls, store.stage376Calls, store.oiNew04Calls)
	}
}
func TestOINew04RuntimeIntegrityFailsClosedOnReplayState(t *testing.T) {
	boom := errors.New("oi-new-04 replay state failed")
	store := &oiNew03IntegrityStore{oiNew04Err: boom}
	svc := NewService(store, SystemClock{})
	if err := svc.ValidateRuntimeIntegrity(context.Background()); !errors.Is(err, boom) {
		t.Fatalf("expected OI-NEW-04 failure, got %v", err)
	}
	if store.stage371Calls != 1 || store.stage376Calls != 1 || store.oiNew04Calls != 1 {
		t.Fatalf("unexpected calls: %d %d %d", store.stage371Calls, store.stage376Calls, store.oiNew04Calls)
	}
}
func TestOINew03RuntimeIntegrityFailsClosedWithoutCapabilities(t *testing.T) {
	svc := NewService(&oiNew03CheapOnlyStore{}, SystemClock{})
	if err := svc.ValidateRuntimeIntegrity(context.Background()); !errors.Is(err, ErrRuntimeIntegrityUnavailable) {
		t.Fatalf("expected unavailable, got %v", err)
	}
}
