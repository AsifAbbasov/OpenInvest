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

func TestStage0333RuntimeRoleCanAppendButCannotMutateLedger(t *testing.T) {
	ownerURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	runtimeURL := os.Getenv("OPENINVEST_DATABASE_RUNTIME_TEST_URL")
	if ownerURL == "" || runtimeURL == "" {
		t.Skip("runtime privilege integration URLs are not set")
	}

	if ownerStore, err := postgres.OpenRuntime(ownerURL); err == nil {
		_ = ownerStore.Close()
		t.Fatal("schema owner unexpectedly passed runtime privilege validation")
	} else if !errors.Is(err, postgres.ErrUnsafeRuntimeDatabaseRole) {
		t.Fatalf("expected unsafe runtime role error for owner connection, got %v", err)
	}

	runtimeStore, err := postgres.OpenRuntime(runtimeURL)
	if err != nil {
		t.Fatalf("open least-privilege runtime store: %v", err)
	}
	defer runtimeStore.Close()

	ownerDB, err := sql.Open("pgx", ownerURL)
	if err != nil {
		t.Fatalf("open owner cleanup db: %v", err)
	}
	closeDBOnCleanup(t, ownerDB, "stage 3.33 runtime owner")

	runtimeDB, err := sql.Open("pgx", runtimeURL)
	if err != nil {
		t.Fatalf("open runtime verification db: %v", err)
	}
	closeDBOnCleanup(t, runtimeDB, "stage 3.33 runtime role")

	ctx := context.Background()
	service := verticalslice.NewService(runtimeStore, verticalslice.SystemClock{})
	subjectID := uuid.NewString()
	portfolio, err := service.CreatePortfolio(
		ctx,
		verticalslice.RequestContext{},
		subjectID,
		"stage-03-33-runtime-portfolio-key-01",
		"/api/v1/portfolios",
		verticalslice.CreatePortfolioRequest{Name: "Runtime privilege proof", BaseCurrency: verticalslice.RUB},
	)
	if err != nil {
		t.Fatalf("runtime role create portfolio: %v", err)
	}
	t.Cleanup(func() { cleanupPortfolioRows(t, ctx, ownerDB, portfolio.ID) })

	_, err = service.AppendTransaction(
		ctx,
		verticalslice.RequestContext{},
		subjectID,
		"stage-03-33-runtime-transaction-key-01",
		"/api/v1/portfolios/"+portfolio.ID+"/transactions",
		verticalslice.AppendTransactionRequest{
			PortfolioID:     portfolio.ID,
			TransactionType: "DEPOSIT",
			GrossAmount: &verticalslice.Money{
				Amount:   decimal.Must("100.00000000"),
				Currency: verticalslice.RUB,
			},
			Commission: verticalslice.Money{Amount: decimal.Zero(), Currency: verticalslice.RUB},
			Tax:        verticalslice.Money{Amount: decimal.Zero(), Currency: verticalslice.RUB},
			TradeDate:  "2026-08-23",
		},
	)
	if err != nil {
		t.Fatalf("runtime role append transaction: %v", err)
	}

	if _, err := runtimeDB.ExecContext(ctx, `
		UPDATE investment.transaction_entries
		SET note = note
		WHERE portfolio_id = $1
	`, portfolio.ID); err == nil {
		t.Fatal("runtime role unexpectedly updated append-only transaction_entries")
	}
	if _, err := runtimeDB.ExecContext(ctx, `
		DELETE FROM investment.transaction_entries
		WHERE portfolio_id = $1
	`, portfolio.ID); err == nil {
		t.Fatal("runtime role unexpectedly deleted append-only transaction_entries")
	}

	var canTruncate bool
	if err := runtimeDB.QueryRowContext(ctx, `
		SELECT has_table_privilege(current_user, 'investment.transaction_entries', 'TRUNCATE')
	`).Scan(&canTruncate); err != nil {
		t.Fatalf("inspect runtime truncate privilege: %v", err)
	}
	if canTruncate {
		t.Fatal("runtime role unexpectedly has TRUNCATE on append-only transaction_entries")
	}
}
