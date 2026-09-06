package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/postgres"
)

func TestStage371RuntimeReadinessExecutesExactIndexGate(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}

	store, err := postgres.Open(databaseURL)
	if err != nil {
		t.Fatalf("open postgres store: %v", err)
	}
	defer store.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := store.Stage371Ready(ctx); err != nil {
		t.Fatalf("Stage 3.71 migrated runtime must satisfy exact readiness gate: %v", err)
	}
}
