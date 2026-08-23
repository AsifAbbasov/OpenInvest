package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func stage329SnapshotTestHarness(t *testing.T) (*postgres.Store, *sql.DB, context.Context, string, verticalslice.Portfolio) {
	t.Helper()
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}

	store, err := postgres.Open(databaseURL)
	if err != nil {
		t.Fatalf("open postgres store: %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Fatalf("close postgres store: %v", err)
		}
	})

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open verification db: %v", err)
	}
	closeDBOnCleanup(t, db, "stage 3.29 snapshot verification")

	ctx := context.Background()
	subjectID := uuid.NewString()
	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	portfolio, err := service.CreatePortfolio(
		ctx,
		verticalslice.RequestContext{},
		subjectID,
		"stage329-portfolio-key-0001",
		"/api/v1/portfolios",
		verticalslice.CreatePortfolioRequest{Name: "Stage 3.29 snapshot bounds", BaseCurrency: verticalslice.RUB},
	)
	if err != nil {
		t.Fatalf("create portfolio: %v", err)
	}

	t.Cleanup(func() {
		cleanupPortfolioRows(t, ctx, db, portfolio.ID)
		if _, err := db.ExecContext(ctx, `DELETE FROM investment.command_deduplication WHERE principal_id = $1`, subjectID); err != nil {
			t.Fatalf("cleanup command deduplication: %v", err)
		}
		if _, err := db.ExecContext(ctx, `DELETE FROM investment.subjects WHERE id = $1`, subjectID); err != nil {
			t.Fatalf("cleanup subject: %v", err)
		}
	})

	return store, db, ctx, subjectID, portfolio
}

func stage329Count(t *testing.T, ctx context.Context, db *sql.DB, query string, args ...any) int {
	t.Helper()
	var count int
	if err := db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		t.Fatalf("count rows: %v", err)
	}
	return count
}

func TestStage329CumulativeDepositsFailClosedBeforeSnapshotOverflow(t *testing.T) {
	store, db, ctx, subjectID, portfolio := stage329SnapshotTestHarness(t)
	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	maxMoney := verticalslice.Money{
		Amount:   decimal.Must("99999999999999999999.99999999"),
		Currency: verticalslice.RUB,
	}
	request := verticalslice.AppendTransactionRequest{
		PortfolioID:     portfolio.ID,
		TransactionType: "DEPOSIT",
		GrossAmount:     &maxMoney,
		Commission:      verticalslice.ZeroMoney(),
		Tax:             verticalslice.ZeroMoney(),
		TradeDate:       "2026-08-23",
	}

	if _, err := service.AppendTransaction(
		ctx,
		verticalslice.RequestContext{},
		subjectID,
		"stage329-deposit-key-0001",
		"/api/v1/portfolios/"+portfolio.ID+"/transactions",
		request,
	); err != nil {
		t.Fatalf("append first maximum deposit: %v", err)
	}

	if _, err := service.AppendTransaction(
		ctx,
		verticalslice.RequestContext{},
		subjectID,
		"stage329-deposit-key-0002",
		"/api/v1/portfolios/"+portfolio.ID+"/transactions",
		request,
	); !errors.Is(err, verticalslice.ErrInvalidInput) {
		t.Fatalf("expected controlled invalid-input error for cumulative snapshot overflow, got %v", err)
	}

	if got := stage329Count(t, ctx, db,
		`SELECT COUNT(*) FROM investment.transaction_entries WHERE portfolio_id = $1`, portfolio.ID); got != 1 {
		t.Fatalf("expected failed second deposit to roll back ledger entry, got %d entries", got)
	}
	if got := stage329Count(t, ctx, db,
		`SELECT COUNT(*) FROM analytics.portfolio_snapshots WHERE portfolio_id = $1`, portfolio.ID); got != 1 {
		t.Fatalf("expected failed second deposit to roll back snapshot version, got %d snapshots", got)
	}
	if got := stage329Count(t, ctx, db,
		`SELECT COUNT(*) FROM investment.command_deduplication WHERE principal_id = $1 AND idempotency_key = $2`,
		subjectID, "stage329-deposit-key-0002"); got != 0 {
		t.Fatalf("expected failed second deposit to roll back idempotency reservation, got %d rows", got)
	}

	summary, err := service.GetPortfolioSummary(ctx, subjectID, portfolio.ID, "2026-08-23")
	if err != nil {
		t.Fatalf("get surviving summary: %v", err)
	}
	if got := summary.CashValue.Amount.String(); got != "99999999999999999999.99999999" {
		t.Fatalf("expected first maximum deposit snapshot to remain canonical, got %s", got)
	}
}

func TestStage329BuyComponentSumFailsClosedAndRemainsAtomic(t *testing.T) {
	store, db, ctx, subjectID, portfolio := stage329SnapshotTestHarness(t)
	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	ticker := "SBER"
	quantity := decimal.Must("99999999999999999999.99999999")
	unitPrice := verticalslice.Money{Amount: decimal.Must("1.00000000"), Currency: verticalslice.RUB}
	maxCommission := verticalslice.Money{
		Amount:   decimal.Must("99999999999999999999.99999999"),
		Currency: verticalslice.RUB,
	}

	_, err := service.AppendTransaction(
		ctx,
		verticalslice.RequestContext{},
		subjectID,
		"stage329-buy-key-000001",
		"/api/v1/portfolios/"+portfolio.ID+"/transactions",
		verticalslice.AppendTransactionRequest{
			PortfolioID:     portfolio.ID,
			TransactionType: "BUY",
			Ticker:          &ticker,
			Quantity:        &quantity,
			UnitPrice:       &unitPrice,
			Commission:      maxCommission,
			Tax:             verticalslice.ZeroMoney(),
			TradeDate:       "2026-08-23",
		},
	)
	if !errors.Is(err, verticalslice.ErrInvalidInput) {
		t.Fatalf("expected controlled invalid-input error for BUY snapshot overflow, got %v", err)
	}

	if got := stage329Count(t, ctx, db,
		`SELECT COUNT(*) FROM investment.transaction_entries WHERE portfolio_id = $1`, portfolio.ID); got != 0 {
		t.Fatalf("expected failed BUY to roll back ledger entry, got %d entries", got)
	}
	if got := stage329Count(t, ctx, db,
		`SELECT COUNT(*) FROM analytics.portfolio_snapshots WHERE portfolio_id = $1`, portfolio.ID); got != 0 {
		t.Fatalf("expected failed BUY to leave no snapshot state, got %d snapshots", got)
	}
	if got := stage329Count(t, ctx, db,
		`SELECT COUNT(*) FROM investment.command_deduplication WHERE principal_id = $1 AND idempotency_key = $2`,
		subjectID, "stage329-buy-key-000001"); got != 0 {
		t.Fatalf("expected failed BUY to roll back idempotency reservation, got %d rows", got)
	}
}
