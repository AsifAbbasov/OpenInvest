package postgres

import (
	"context"
	"os"
	"testing"
)

func TestArchitectureHardeningRuntimeMaterializations(t *testing.T) {
	runtimeURL := os.Getenv("OPENINVEST_ARCH_HARDENING_RUNTIME_URL")
	subjectID := os.Getenv("OPENINVEST_ARCH_HARDENING_SUBJECT_ID")
	portfolioID := os.Getenv("OPENINVEST_ARCH_HARDENING_PORTFOLIO_ID")
	if runtimeURL == "" || subjectID == "" || portfolioID == "" {
		t.Skip("architecture hardening runtime materialization environment is not set")
	}

	store, err := OpenRuntime(runtimeURL)
	if err != nil {
		t.Fatalf("open runtime store: %v", err)
	}
	defer func() { _ = store.Close() }()

	run := func(name string, fn func(context.Context) error) {
		t.Helper()
		instrumentation := &effectiveLedgerInstrumentation{}
		ctx := withEffectiveLedgerInstrumentation(context.Background(), instrumentation)
		if err := fn(ctx); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		got := instrumentation.materializations.Load()
		t.Logf("ARCH_HARDENING_MATERIALIZATION endpoint=%s count=%d", name, got)
		if got != 1 {
			t.Fatalf("%s materializations=%d want=1", name, got)
		}
	}

	run("positions", func(ctx context.Context) error {
		_, err := store.GetPortfolioPositions(ctx, subjectID, portfolioID, "2010-01-01")
		return err
	})
	run("cash-flow", func(ctx context.Context) error {
		_, err := store.GetPortfolioCashFlow(ctx, subjectID, portfolioID, "2000-01-01", "2010-01-01")
		return err
	})
	run("returns", func(ctx context.Context) error {
		_, err := store.GetPortfolioReturns(ctx, subjectID, portfolioID, "2010-01-01")
		return err
	})
	run("summary", func(ctx context.Context) error {
		_, err := store.GetPortfolioSummaryStage371(ctx, subjectID, portfolioID, "2010-01-01")
		return err
	})
}
