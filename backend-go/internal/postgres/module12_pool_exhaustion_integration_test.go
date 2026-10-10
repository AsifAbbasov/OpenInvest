package postgres

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"
)

func TestM12PostgresPoolExhaustionAndRecovery(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_M12_POSTGRES_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_M12_POSTGRES_URL is required")
	}
	store, err := Open(databaseURL)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	const capacity = 10
	held := make([]interface{ Close() error }, 0, capacity)
	for i := 0; i < capacity; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		conn, err := store.db.Conn(ctx)
		if err == nil {
			err = conn.PingContext(ctx)
		}
		cancel()
		if err != nil {
			t.Fatalf("hold connection %d: %v", i+1, err)
		}
		held = append(held, conn)
	}
	peak := store.db.Stats()
	if peak.InUse != capacity {
		t.Fatalf("pool in-use=%d want=%d", peak.InUse, capacity)
	}

	blockedCtx, blockedCancel := context.WithTimeout(context.Background(), 400*time.Millisecond)
	start := time.Now()
	blockedConn, blockedErr := store.db.Conn(blockedCtx)
	elapsed := time.Since(start)
	blockedCancel()
	if blockedConn != nil {
		_ = blockedConn.Close()
	}
	if !errors.Is(blockedErr, context.DeadlineExceeded) {
		t.Fatalf("11th connection error=%v want deadline exceeded", blockedErr)
	}

	if err := held[0].Close(); err != nil {
		t.Fatalf("release held connection: %v", err)
	}
	held = held[1:]

	recoveryCtx, recoveryCancel := context.WithTimeout(context.Background(), 2*time.Second)
	recovered, err := store.db.Conn(recoveryCtx)
	if err == nil {
		err = recovered.PingContext(recoveryCtx)
	}
	recoveryCancel()
	if err != nil {
		t.Fatalf("pool did not recover after release: %v", err)
	}
	_ = recovered.Close()
	for _, conn := range held {
		_ = conn.Close()
	}
	time.Sleep(100 * time.Millisecond)
	recovery := store.db.Stats()

	t.Logf("M12_POSTGRES_POOL max_open=%d peak_open=%d peak_in_use=%d wait_count=%d blocked_elapsed=%s blocked_error=DEADLINE_EXCEEDED recovery_open=%d recovery_in_use=%d recovery_idle=%d",
		store.db.Stats().MaxOpenConnections, peak.OpenConnections, peak.InUse, recovery.WaitCount, elapsed.Round(time.Millisecond), recovery.OpenConnections, recovery.InUse, recovery.Idle)
}
