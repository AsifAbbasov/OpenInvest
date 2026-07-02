package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"sync"
	"testing"

	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestStoreVerticalSlice(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}

	store, err := postgres.Open(databaseURL)
	if err != nil {
		t.Fatalf("open postgres store: %v", err)
	}
	defer store.Close()

	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	ctx := context.Background()
	subjectID := uuid.NewString()

	portfolio, err := service.CreatePortfolio(ctx, verticalslice.RequestContext{}, subjectID, "portfolio-key-000001", "/api/v1/portfolios", verticalslice.CreatePortfolioRequest{
		Name:         "Long-term capital",
		BaseCurrency: verticalslice.RUB,
	})
	if err != nil {
		t.Fatalf("create portfolio: %v", err)
	}

	ticker := "SBER"
	quantity := decimal.Must("100.00000000")
	unitPrice := verticalslice.Money{Amount: decimal.Must("280.00000000"), Currency: verticalslice.RUB}
	_, err = service.AppendTransaction(ctx, verticalslice.RequestContext{}, subjectID, "transaction-key-000001", "/api/v1/portfolios/"+portfolio.ID+"/transactions", verticalslice.AppendTransactionRequest{
		PortfolioID:     portfolio.ID,
		TransactionType: "BUY",
		Ticker:          &ticker,
		Quantity:        &quantity,
		UnitPrice:       &unitPrice,
		Commission:      verticalslice.Money{Amount: decimal.Must("28.00000000"), Currency: verticalslice.RUB},
		Tax:             verticalslice.ZeroMoney(),
		TradeDate:       "2026-01-10",
	})
	if err != nil {
		t.Fatalf("append transaction: %v", err)
	}

	_, err = service.AppendTransaction(ctx, verticalslice.RequestContext{}, subjectID, "transaction-key-000002", "/api/v1/portfolios/"+portfolio.ID+"/transactions", verticalslice.AppendTransactionRequest{
		PortfolioID:     portfolio.ID,
		TransactionType: "BUY",
		Ticker:          &ticker,
		Quantity:        &quantity,
		UnitPrice:       &unitPrice,
		Commission:      verticalslice.Money{Amount: decimal.Must("28.00000000"), Currency: verticalslice.RUB},
		Tax:             verticalslice.ZeroMoney(),
		TradeDate:       "2026-01-10",
	})
	if err != nil {
		t.Fatalf("append second same-day transaction: %v", err)
	}

	summary, err := service.GetPortfolioSummary(ctx, subjectID, portfolio.ID, "")
	if err != nil {
		t.Fatalf("get portfolio summary: %v", err)
	}
	if got := summary.StockValue.Amount.String(); got != "56000.00000000" {
		t.Fatalf("expected stock value 56000.00000000, got %s", got)
	}

	_, err = service.GetPortfolioSummary(ctx, subjectID, portfolio.ID, "2026-01-09")
	if !errors.Is(err, postgres.ErrNotFound) {
		t.Fatalf("expected no summary before first trade business date, got %v", err)
	}

	backdatedQuantity := decimal.Must("10.00000000")
	backdatedPrice := verticalslice.Money{Amount: decimal.Must("100.00000000"), Currency: verticalslice.RUB}
	_, err = service.AppendTransaction(ctx, verticalslice.RequestContext{}, subjectID, "transaction-key-000003", "/api/v1/portfolios/"+portfolio.ID+"/transactions", verticalslice.AppendTransactionRequest{
		PortfolioID:     portfolio.ID,
		TransactionType: "BUY",
		Ticker:          &ticker,
		Quantity:        &backdatedQuantity,
		UnitPrice:       &backdatedPrice,
		Commission:      verticalslice.ZeroMoney(),
		Tax:             verticalslice.ZeroMoney(),
		TradeDate:       "2026-01-05",
	})
	if err != nil {
		t.Fatalf("append backdated transaction: %v", err)
	}

	summary, err = service.GetPortfolioSummary(ctx, subjectID, portfolio.ID, "2026-01-05")
	if err != nil {
		t.Fatalf("get backdated portfolio summary: %v", err)
	}
	if got := summary.StockValue.Amount.String(); got != "1000.00000000" {
		t.Fatalf("expected stock value 1000.00000000 for 2026-01-05, got %s", got)
	}

	summary, err = service.GetPortfolioSummary(ctx, subjectID, portfolio.ID, "2026-01-10")
	if err != nil {
		t.Fatalf("get portfolio summary after backdated transaction: %v", err)
	}
	if got := summary.StockValue.Amount.String(); got != "57000.00000000" {
		t.Fatalf("expected stock value 57000.00000000 after backdated rebuild, got %s", got)
	}

	filtered, err := service.ListTransactions(ctx, subjectID, portfolio.ID, verticalslice.TransactionFilter{
		TransactionType: "BUY",
		FromDate:        "2026-01-06",
		ToDate:          "2026-01-10",
		Limit:           10,
	})
	if err != nil {
		t.Fatalf("list filtered transactions: %v", err)
	}
	if len(filtered) != 2 {
		t.Fatalf("expected 2 filtered transactions, got %d", len(filtered))
	}
}

func TestStoreAppendImportedTransactionsIsAtomicAndIdempotent(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}

	store, err := postgres.Open(databaseURL)
	if err != nil {
		t.Fatalf("open postgres store: %v", err)
	}
	defer store.Close()

	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	ctx := context.Background()
	subjectID := uuid.NewString()

	portfolio, err := service.CreatePortfolio(ctx, verticalslice.RequestContext{}, subjectID, "portfolio-key-import-01", "/api/v1/portfolios", verticalslice.CreatePortfolioRequest{
		Name:         "Imported capital",
		BaseCurrency: verticalslice.RUB,
	})
	if err != nil {
		t.Fatalf("create portfolio: %v", err)
	}

	depositGross := verticalslice.Money{Amount: decimal.Must("1000.00000000"), Currency: verticalslice.RUB}
	ticker := "SBER"
	quantity := decimal.Must("2.00000000")
	unitPrice := verticalslice.Money{Amount: decimal.Must("100.00000000"), Currency: verticalslice.RUB}
	batch := verticalslice.AppendImportBatchRequest{
		PortfolioID: portfolio.ID,
		Transactions: []verticalslice.AppendTransactionRequest{
			{
				PortfolioID:     portfolio.ID,
				TransactionType: "DEPOSIT",
				GrossAmount:     &depositGross,
				Commission:      verticalslice.ZeroMoney(),
				Tax:             verticalslice.ZeroMoney(),
				TradeDate:       "2026-06-19",
			},
			{
				PortfolioID:     portfolio.ID,
				TransactionType: "BUY",
				Ticker:          &ticker,
				Quantity:        &quantity,
				UnitPrice:       &unitPrice,
				Commission:      verticalslice.Money{Amount: decimal.Must("1.00000000"), Currency: verticalslice.RUB},
				Tax:             verticalslice.ZeroMoney(),
				TradeDate:       "2026-06-20",
			},
		},
		SourceKind:     "USER_UPLOADED_FILE",
		SourceFileHash: "file-hash",
	}

	transactions, err := service.AppendImportedTransactions(ctx, verticalslice.RequestContext{}, subjectID, "import-batch-key-001", "/internal/imports/append", batch)
	if err != nil {
		t.Fatalf("append imported transactions: %v", err)
	}
	if len(transactions) != 2 {
		t.Fatalf("expected 2 imported transactions, got %d", len(transactions))
	}

	summary, err := service.GetPortfolioSummary(ctx, subjectID, portfolio.ID, "2026-06-20")
	if err != nil {
		t.Fatalf("get portfolio summary: %v", err)
	}
	if got := summary.StockValue.Amount.String(); got != "200.00000000" {
		t.Fatalf("expected stock value 200.00000000, got %s", got)
	}
	if got := summary.CashValue.Amount.String(); got != "799.00000000" {
		t.Fatalf("expected cash value 799.00000000, got %s", got)
	}

	_, err = service.AppendImportedTransactions(ctx, verticalslice.RequestContext{}, subjectID, "import-batch-key-001", "/internal/imports/append", batch)
	if !errors.Is(err, postgres.ErrUnsupportedDuplicate) {
		t.Fatalf("expected unsupported duplicate for repeated import batch, got %v", err)
	}

	secondDeposit := verticalslice.Money{Amount: decimal.Must("500.00000000"), Currency: verticalslice.RUB}
	conflictingBatch := verticalslice.AppendImportBatchRequest{
		PortfolioID: portfolio.ID,
		Transactions: []verticalslice.AppendTransactionRequest{
			{
				PortfolioID:     portfolio.ID,
				TransactionType: "DEPOSIT",
				GrossAmount:     &secondDeposit,
				Commission:      verticalslice.ZeroMoney(),
				Tax:             verticalslice.ZeroMoney(),
				TradeDate:       "2026-06-21",
			},
			batch.Transactions[1],
		},
		SourceKind:     "USER_UPLOADED_FILE",
		SourceFileHash: "file-hash-2",
	}

	_, err = service.AppendImportedTransactions(ctx, verticalslice.RequestContext{}, subjectID, "import-batch-key-002", "/internal/imports/append", conflictingBatch)
	if !errors.Is(err, verticalslice.ErrInvalidInput) {
		t.Fatalf("expected duplicate conflict for import batch, got %v", err)
	}

	listed, err := service.ListTransactions(ctx, subjectID, portfolio.ID, verticalslice.TransactionFilter{Limit: 10})
	if err != nil {
		t.Fatalf("list transactions: %v", err)
	}
	if len(listed) != 2 {
		t.Fatalf("expected rollback to keep 2 transactions, got %d", len(listed))
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open audit check db: %v", err)
	}
	defer db.Close()

	var auditCount int
	err = db.QueryRowContext(ctx, `
		SELECT count(*)
		FROM audit.events
		WHERE target_id = $1 AND action_code = 'IMPORT_APPEND_BATCH' AND outcome = 'success'
	`, portfolio.ID).Scan(&auditCount)
	if err != nil {
		t.Fatalf("query audit evidence: %v", err)
	}
	if auditCount != 1 {
		t.Fatalf("expected one import append audit event, got %d", auditCount)
	}
}

func TestStoreAppendImportedTransactionsSerializesConcurrentDuplicateBatches(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}

	store, err := postgres.Open(databaseURL)
	if err != nil {
		t.Fatalf("open postgres store: %v", err)
	}
	defer store.Close()

	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	ctx := context.Background()
	subjectID := uuid.NewString()

	portfolio, err := service.CreatePortfolio(ctx, verticalslice.RequestContext{}, subjectID, "portfolio-key-import-02", "/api/v1/portfolios", verticalslice.CreatePortfolioRequest{
		Name:         "Concurrent imported capital",
		BaseCurrency: verticalslice.RUB,
	})
	if err != nil {
		t.Fatalf("create portfolio: %v", err)
	}

	gross := verticalslice.Money{Amount: decimal.Must("750.00000000"), Currency: verticalslice.RUB}
	batch := verticalslice.AppendImportBatchRequest{
		PortfolioID: portfolio.ID,
		Transactions: []verticalslice.AppendTransactionRequest{{
			PortfolioID:     portfolio.ID,
			TransactionType: "DEPOSIT",
			GrossAmount:     &gross,
			Commission:      verticalslice.ZeroMoney(),
			Tax:             verticalslice.ZeroMoney(),
			TradeDate:       "2026-06-23",
		}},
		SourceKind:     "USER_UPLOADED_FILE",
		SourceFileHash: "file-hash-concurrent",
	}

	start := make(chan struct{})
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for _, key := range []string{"import-batch-key-101", "import-batch-key-102"} {
		wg.Add(1)
		go func(idempotencyKey string) {
			defer wg.Done()
			<-start
			_, err := service.AppendImportedTransactions(ctx, verticalslice.RequestContext{}, subjectID, idempotencyKey, "/internal/imports/append", batch)
			errs <- err
		}(key)
	}
	close(start)
	wg.Wait()
	close(errs)

	successes := 0
	duplicates := 0
	for err := range errs {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, verticalslice.ErrInvalidInput):
			duplicates++
		default:
			t.Fatalf("unexpected concurrent append error: %v", err)
		}
	}
	if successes != 1 || duplicates != 1 {
		t.Fatalf("expected one success and one duplicate rejection, got successes=%d duplicates=%d", successes, duplicates)
	}

	listed, err := service.ListTransactions(ctx, subjectID, portfolio.ID, verticalslice.TransactionFilter{Limit: 10})
	if err != nil {
		t.Fatalf("list transactions: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected exactly one imported transaction after concurrent batches, got %d", len(listed))
	}
}
