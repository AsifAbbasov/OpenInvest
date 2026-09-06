package postgres_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestStage371SummaryPrefersSerializedVersionOverCalculatedAt(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}

	store, err := postgres.Open(databaseURL)
	if err != nil {
		t.Fatalf("open postgres store: %v", err)
	}
	defer store.Close()

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open summary verification database: %v", err)
	}
	closeDBOnCleanup(t, db, "Stage 3.71 summary selection")

	ctx := context.Background()
	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	subjectID := uuid.NewString()
	portfolio, err := service.CreatePortfolio(
		ctx,
		verticalslice.RequestContext{},
		subjectID,
		"stage-03-71-summary-portfolio-key",
		"/api/v1/portfolios",
		verticalslice.CreatePortfolioRequest{Name: "Stage 3.71 summary ordering", BaseCurrency: verticalslice.RUB},
	)
	if err != nil {
		t.Fatalf("create portfolio: %v", err)
	}
	t.Cleanup(func() { cleanupPortfolioRows(t, ctx, db, portfolio.ID) })

	ticker := "SBER"
	quantity := decimal.Must("1.00000000")
	prices := []string{"100.00000000", "200.00000000"}
	for index, price := range prices {
		unitPrice := verticalslice.Money{Amount: decimal.Must(price), Currency: verticalslice.RUB}
		_, err := service.AppendTransaction(
			ctx,
			verticalslice.RequestContext{},
			subjectID,
			[]string{"stage-03-71-summary-buy-key-01", "stage-03-71-summary-buy-key-02"}[index],
			"/api/v1/portfolios/"+portfolio.ID+"/transactions",
			verticalslice.AppendTransactionRequest{
				PortfolioID:     portfolio.ID,
				TransactionType: "BUY",
				Ticker:          &ticker,
				Quantity:        &quantity,
				UnitPrice:       &unitPrice,
				Commission:      verticalslice.ZeroMoney(),
				Tax:             verticalslice.ZeroMoney(),
				TradeDate:       "2026-09-06",
			},
		)
		if err != nil {
			t.Fatalf("append BUY %d: %v", index+1, err)
		}
	}

	var stage371Count int
	var maxVersion int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*), MAX(snapshot_version)
		FROM analytics.portfolio_snapshots
		WHERE portfolio_id = $1
			AND snapshot_date = '2026-09-06'::date
			AND methodology_version = 'stage-03-71-position-cost-snapshot-v1'
	`, portfolio.ID).Scan(&stage371Count, &maxVersion); err != nil {
		t.Fatalf("query Stage 3.71 snapshot versions: %v", err)
	}
	if stage371Count != 2 || maxVersion != 2 {
		t.Fatalf("expected exactly two serialized Stage 3.71 snapshots ending at version 2, got count=%d max=%d", stage371Count, maxVersion)
	}

	if _, err := db.ExecContext(ctx, `
		UPDATE analytics.portfolio_snapshots
		SET calculated_at = CASE snapshot_version
			WHEN 1 THEN '2099-01-01T00:00:00Z'::timestamptz
			WHEN 2 THEN '2000-01-01T00:00:00Z'::timestamptz
			ELSE calculated_at
		END
		WHERE portfolio_id = $1
			AND snapshot_date = '2026-09-06'::date
			AND methodology_version = 'stage-03-71-position-cost-snapshot-v1'
	`, portfolio.ID); err != nil {
		t.Fatalf("invert Stage 3.71 calculated_at chronology: %v", err)
	}

	if _, err := db.ExecContext(ctx, `
		INSERT INTO analytics.portfolio_snapshots (
			id, portfolio_id, snapshot_date,
			total_value_amount, cash_value_amount, stock_value_amount, bond_value_amount,
			invested_capital_amount, nominal_return_rate, real_return_rate,
			snapshot_version, methodology_version, input_watermark, calculated_at
		)
		SELECT
			$2::uuid, portfolio_id, snapshot_date,
			999.00000000, cash_value_amount, 999.00000000, bond_value_amount,
			invested_capital_amount, nominal_return_rate, real_return_rate,
			99, 'stage-03-02-local-cost-snapshot-v1', input_watermark, '2100-01-01T00:00:00Z'::timestamptz
		FROM analytics.portfolio_snapshots
		WHERE portfolio_id = $1
			AND snapshot_date = '2026-09-06'::date
			AND methodology_version = 'stage-03-71-position-cost-snapshot-v1'
			AND snapshot_version = 2
	`, portfolio.ID, uuid.NewString()); err != nil {
		t.Fatalf("insert preserved Stage 3.02 competing snapshot: %v", err)
	}

	summary, err := service.GetPortfolioSummary(ctx, subjectID, portfolio.ID, "2026-09-06")
	if err != nil {
		t.Fatalf("get deterministic Stage 3.71 summary: %v", err)
	}
	if summary.MethodologyVersion != "stage-03-71-position-cost-snapshot-v1" {
		t.Fatalf("expected Stage 3.71 methodology to outrank preserved Stage 3.02 snapshot, got %s", summary.MethodologyVersion)
	}
	if got := summary.StockValue.Amount.String(); got != "300.00000000" {
		t.Fatalf("expected serialized version 2 stock basis 300.00000000 despite inverted timestamps, got %s", got)
	}
}
