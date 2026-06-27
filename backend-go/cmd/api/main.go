package main

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/gofiber/fiber/v3"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/httpapi"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func newApp() *fiber.App {
	var store verticalslice.Store
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		store = unavailableStore{}
		return httpapi.New(verticalslice.NewService(store, verticalslice.SystemClock{}))
	}
	store, err := postgres.Open(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	return httpapi.New(verticalslice.NewService(store, verticalslice.SystemClock{}))
}

func main() {
	if err := newApp().Listen(":8080"); err != nil {
		log.Fatal(err)
	}
}

type unavailableStore struct{}

func (unavailableStore) Ping(context.Context) error {
	return errors.New("database url is not configured")
}

func (unavailableStore) ListPortfolios(context.Context, string, int) ([]verticalslice.Portfolio, error) {
	return nil, errors.New("database url is not configured")
}

func (unavailableStore) CreatePortfolio(context.Context, verticalslice.CommandContext, verticalslice.CreatePortfolioRequest) (verticalslice.Portfolio, error) {
	return verticalslice.Portfolio{}, errors.New("database url is not configured")
}

func (unavailableStore) GetPortfolio(context.Context, string, string) (verticalslice.Portfolio, error) {
	return verticalslice.Portfolio{}, errors.New("database url is not configured")
}

func (unavailableStore) ListTransactions(context.Context, string, string, verticalslice.TransactionFilter) ([]verticalslice.Transaction, error) {
	return nil, errors.New("database url is not configured")
}

func (unavailableStore) AppendTransaction(context.Context, verticalslice.CommandContext, verticalslice.AppendTransactionRequest) (verticalslice.Transaction, error) {
	return verticalslice.Transaction{}, errors.New("database url is not configured")
}

func (unavailableStore) GetPortfolioSummary(context.Context, string, string, string) (verticalslice.PortfolioSummary, error) {
	return verticalslice.PortfolioSummary{
		TotalValue:        verticalslice.Money{Amount: decimal.Zero(), Currency: verticalslice.RUB},
		CashValue:         verticalslice.Money{Amount: decimal.Zero(), Currency: verticalslice.RUB},
		StockValue:        verticalslice.Money{Amount: decimal.Zero(), Currency: verticalslice.RUB},
		BondValue:         verticalslice.Money{Amount: decimal.Zero(), Currency: verticalslice.RUB},
		InvestedCapital:   verticalslice.Money{Amount: decimal.Zero(), Currency: verticalslice.RUB},
		DividendsReceived: verticalslice.Money{Amount: decimal.Zero(), Currency: verticalslice.RUB},
		CouponsReceived:   verticalslice.Money{Amount: decimal.Zero(), Currency: verticalslice.RUB},
	}, errors.New("database url is not configured")
}
