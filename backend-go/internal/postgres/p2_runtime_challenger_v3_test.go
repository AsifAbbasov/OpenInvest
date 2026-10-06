package postgres

import (
	"context"
	"os"
	"testing"
)

func TestP2RuntimeChallengerV3Materializations(t *testing.T) {
	if os.Getenv("OPENINVEST_P2_V3") != "1" {
		t.Skip("OPENINVEST_P2_V3 is not enabled")
	}
	runtimeURL := os.Getenv("OPENINVEST_P2_V3_RUNTIME_URL")
	subjectID := os.Getenv("OPENINVEST_P2_V3_SUBJECT_ID")
	portfolioID := os.Getenv("OPENINVEST_P2_V3_PORTFOLIO_ID")
	if runtimeURL == "" || subjectID == "" || portfolioID == "" {
		t.Fatal("runtime URL, subject ID and portfolio ID are required")
	}

	store, err := OpenRuntime(runtimeURL)
	if err != nil {
		t.Fatalf("open runtime store: %v", err)
	}
	defer func() { _ = store.Close() }()

	run := func(name string, want int64, fn func(context.Context) error) {
		t.Helper()
		instrumentation := &effectiveLedgerInstrumentation{}
		ctx := withEffectiveLedgerInstrumentation(context.Background(), instrumentation)
		if err := fn(ctx); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		got := instrumentation.materializations.Load()
		t.Logf("P2V3_MATERIALIZATION endpoint=%s count=%d", name, got)
		if got != want {
			t.Fatalf("%s materializations=%d want=%d", name, got, want)
		}
	}

	run("positions", 1, func(ctx context.Context) error {
		_, err := store.GetPortfolioPositions(ctx, subjectID, portfolioID, "2010-01-01")
		return err
	})
	run("cash-flow", 1, func(ctx context.Context) error {
		_, err := store.GetPortfolioCashFlow(ctx, subjectID, portfolioID, "2000-01-01", "2010-01-01")
		return err
	})
	run("returns", 2, func(ctx context.Context) error {
		_, err := store.GetPortfolioReturns(ctx, subjectID, portfolioID, "2010-01-01")
		return err
	})
	run("summary", 3, func(ctx context.Context) error {
		_, err := store.GetPortfolioSummaryStage371(ctx, subjectID, portfolioID, "2010-01-01")
		return err
	})
}
