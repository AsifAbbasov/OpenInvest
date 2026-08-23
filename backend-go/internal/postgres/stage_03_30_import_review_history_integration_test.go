package postgres_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func stage330HistoryHarness(t *testing.T) (*postgres.Store, *sql.DB, context.Context, string, verticalslice.Portfolio) {
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
	closeDBOnCleanup(t, db, "stage 3.30 import history")

	ctx := context.Background()
	subjectID := uuid.NewString()
	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	portfolio, err := service.CreatePortfolio(
		ctx,
		verticalslice.RequestContext{},
		subjectID,
		"stage330-portfolio-key-0001",
		"/api/v1/portfolios",
		verticalslice.CreatePortfolioRequest{Name: "Stage 3.30 import history", BaseCurrency: verticalslice.RUB},
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

func stage330InsertManualDeposit(t *testing.T, ctx context.Context, db *sql.DB, portfolioID string, tradeDate string, gross string, createdAt time.Time) string {
	t.Helper()
	entryID := uuid.NewString()
	_, err := db.ExecContext(ctx, `
		INSERT INTO investment.transaction_entries (
			entry_id, transaction_id, portfolio_id, asset_id, revision, transaction_type,
			quantity, unit_price_amount, unit_price_currency,
			gross_amount, gross_currency, commission_amount, commission_currency,
			tax_amount, tax_currency, trade_date, settlement_date, note,
			source_kind, source_file_hash, created_at, request_id, trace_id
		)
		VALUES (
			$1, $2, $3, NULL, 1, 'DEPOSIT',
			NULL, NULL, NULL,
			$4::numeric, 'RUB', 0, 'RUB',
			0, 'RUB', $5::date, NULL, NULL,
			'MANUAL', NULL, $6, NULL, NULL
		)
	`, entryID, uuid.NewString(), portfolioID, gross, tradeDate, createdAt)
	if err != nil {
		t.Fatalf("insert manual deposit: %v", err)
	}
	return entryID
}

func TestStage330TargetedImportHistoryFindsOldRowBeyondLatestHundred(t *testing.T) {
	store, db, ctx, subjectID, portfolio := stage330HistoryHarness(t)
	oldEntryID := stage330InsertManualDeposit(t, ctx, db, portfolio.ID, "2025-01-01", "10.00000000", time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC))

	for index := 0; index < 100; index++ {
		stage330InsertManualDeposit(
			t,
			ctx,
			db,
			portfolio.ID,
			"2026-01-01",
			fmt.Sprintf("%d.00000000", 1000+index),
			time.Date(2026, 1, 1, 12, index%60, 0, 0, time.UTC),
		)
	}

	latest, err := store.ListTransactions(ctx, subjectID, portfolio.ID, verticalslice.TransactionFilter{Limit: 100})
	if err != nil {
		t.Fatalf("list latest transactions: %v", err)
	}
	if len(latest) != 100 {
		t.Fatalf("expected latest public page to contain 100 rows, got %d", len(latest))
	}
	for _, item := range latest {
		if item.EntryID == oldEntryID {
			t.Fatal("test setup failed: old target unexpectedly present in latest-100 page")
		}
	}

	history, err := store.ListImportReviewTransactions(ctx, subjectID, portfolio.ID, verticalslice.ImportReviewHistoryFilter{
		TradeDates: []string{"2025-01-01"},
	})
	if err != nil {
		t.Fatalf("list targeted import review history: %v", err)
	}
	if len(history) != 1 || history[0].EntryID != oldEntryID {
		t.Fatalf("expected targeted full-history lookup to return old target %s, got %+v", oldEntryID, history)
	}
}
