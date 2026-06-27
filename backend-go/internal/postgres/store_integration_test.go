package postgres_test

import (
	"context"
	"errors"
	"os"
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
