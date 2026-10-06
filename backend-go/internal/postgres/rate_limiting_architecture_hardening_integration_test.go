package postgres

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestRateLimitingHardeningFinancialReadsMaterializeEffectiveLedgerOnce(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}

	store, err := Open(databaseURL)
	if err != nil {
		t.Fatalf("open postgres store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	ctx := context.Background()
	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	subjectID := uuid.NewString()
	portfolio, err := service.CreatePortfolio(
		ctx,
		verticalslice.RequestContext{},
		subjectID,
		"rlsa04-portfolio-"+uuid.NewString(),
		"/api/v1/portfolios",
		verticalslice.CreatePortfolioRequest{Name: "RLSA-04 projection reuse", BaseCurrency: verticalslice.RUB},
	)
	if err != nil {
		t.Fatalf("create portfolio: %v", err)
	}
	_, err = service.AppendTransaction(
		ctx,
		verticalslice.RequestContext{},
		subjectID,
		"rlsa04-deposit-"+uuid.NewString(),
		"/api/v1/portfolios/"+portfolio.ID+"/transactions",
		verticalslice.AppendTransactionRequest{
			PortfolioID:     portfolio.ID,
			TransactionType: "DEPOSIT",
			GrossAmount:     &verticalslice.Money{Amount: decimal.Must("1000.00000000"), Currency: verticalslice.RUB},
			Commission:      verticalslice.ZeroMoney(),
			Tax:             verticalslice.ZeroMoney(),
			TradeDate:       "2026-10-01",
		},
	)
	if err != nil {
		t.Fatalf("append deposit: %v", err)
	}

	run := func(name string, fn func(context.Context) error) {
		t.Helper()
		instrumentation := &effectiveLedgerInstrumentation{}
		readCtx := withEffectiveLedgerInstrumentation(ctx, instrumentation)
		if err := fn(readCtx); err != nil {
			t.Fatalf("%s read: %v", name, err)
		}
		if got := instrumentation.materializations.Load(); got != 1 {
			t.Fatalf("%s materialized/validated the effective ledger %d times, want 1", name, got)
		}
		t.Logf("RLSA04_MATERIALIZATION endpoint=%s count=1", name)
	}

	run("positions", func(readCtx context.Context) error {
		_, err := store.GetPortfolioPositions(readCtx, subjectID, portfolio.ID, "2026-10-01")
		return err
	})
	run("cash-flow", func(readCtx context.Context) error {
		_, err := store.GetPortfolioCashFlow(readCtx, subjectID, portfolio.ID, "", "2026-10-01")
		return err
	})
	run("returns", func(readCtx context.Context) error {
		_, err := store.GetPortfolioReturns(readCtx, subjectID, portfolio.ID, "2026-10-01")
		return err
	})
	run("summary", func(readCtx context.Context) error {
		_, err := store.GetPortfolioSummaryStage371(readCtx, subjectID, portfolio.ID, "2026-10-01")
		return err
	})
}
