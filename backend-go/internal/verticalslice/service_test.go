package verticalslice

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
)

type fixedClock struct{}

func (fixedClock) Now() time.Time {
	return time.Date(2026, 6, 26, 12, 0, 0, 0, time.UTC)
}

type recordingStore struct {
	requestHash string
}

func (store *recordingStore) Ping(context.Context) error { return nil }

func (store *recordingStore) ListPortfolios(context.Context, string, int) ([]Portfolio, error) {
	return nil, nil
}

func (store *recordingStore) CreatePortfolio(_ context.Context, command CommandContext, _ CreatePortfolioRequest) (Portfolio, error) {
	store.requestHash = command.RequestHash
	return Portfolio{ID: "portfolio-id", BaseCurrency: RUB, Version: 1}, nil
}

func (store *recordingStore) GetPortfolio(context.Context, string, string) (Portfolio, error) {
	return Portfolio{}, nil
}

func (store *recordingStore) ListTransactions(context.Context, string, string, int) ([]Transaction, error) {
	return nil, nil
}

func (store *recordingStore) AppendTransaction(_ context.Context, command CommandContext, _ AppendTransactionRequest) (Transaction, error) {
	store.requestHash = command.RequestHash
	return Transaction{}, nil
}

func (store *recordingStore) GetPortfolioSummary(context.Context, string, string, string) (PortfolioSummary, error) {
	return PortfolioSummary{}, nil
}

func TestCreatePortfolioRequiresIdempotencyKey(t *testing.T) {
	service := NewService(&recordingStore{}, fixedClock{})

	_, err := service.CreatePortfolio(context.Background(), "subject", "", "/api/v1/portfolios", CreatePortfolioRequest{
		Name:         "Long-term capital",
		BaseCurrency: RUB,
	})

	if !errors.Is(err, ErrMissingIdempotency) {
		t.Fatalf("expected missing idempotency key, got %v", err)
	}
}

func TestAppendTransactionRejectsInvalidTicker(t *testing.T) {
	service := NewService(&recordingStore{}, fixedClock{})
	ticker := "sber"
	quantity := decimal.Must("1.00000000")
	unitPrice := Money{Amount: decimal.Must("100.00000000"), Currency: RUB}

	_, err := service.AppendTransaction(context.Background(), "subject", "valid-key-000001", "/path", AppendTransactionRequest{
		PortfolioID:     "portfolio-id",
		TransactionType: "BUY",
		Ticker:          &ticker,
		Quantity:        &quantity,
		UnitPrice:       &unitPrice,
		Commission:      ZeroMoney(),
		Tax:             ZeroMoney(),
		TradeDate:       "2026-06-26",
	})

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}

func TestCreatePortfolioRejectsContractInvalidIdempotencyKey(t *testing.T) {
	service := NewService(&recordingStore{}, fixedClock{})

	_, err := service.CreatePortfolio(context.Background(), "subject", "too-short", "/api/v1/portfolios", CreatePortfolioRequest{
		Name:         "Long-term capital",
		BaseCurrency: RUB,
	})

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}

func TestGrossForDerivesTradeGrossAmount(t *testing.T) {
	quantity := decimal.Must("3.00000000")
	unitPrice := Money{Amount: decimal.Must("10.50000000"), Currency: RUB}

	gross, err := GrossFor(AppendTransactionRequest{Quantity: &quantity, UnitPrice: &unitPrice})
	if err != nil {
		t.Fatalf("derive gross amount: %v", err)
	}

	if got := gross.Amount.String(); got != "31.50000000" {
		t.Fatalf("expected 31.50000000, got %s", got)
	}
}

func TestGrossForRejectsTradeGrossMismatch(t *testing.T) {
	quantity := decimal.Must("3.00000000")
	unitPrice := Money{Amount: decimal.Must("10.50000000"), Currency: RUB}
	wrongGross := Money{Amount: decimal.Must("32.00000000"), Currency: RUB}

	_, err := GrossFor(AppendTransactionRequest{
		Quantity:    &quantity,
		UnitPrice:   &unitPrice,
		GrossAmount: &wrongGross,
	})

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}

func TestAppendTransactionRejectsSellUntilCostBasisExists(t *testing.T) {
	service := NewService(&recordingStore{}, fixedClock{})
	ticker := "SBER"
	quantity := decimal.Must("1.00000000")
	unitPrice := Money{Amount: decimal.Must("110.00000000"), Currency: RUB}

	_, err := service.AppendTransaction(context.Background(), "subject", "valid-key-000001", "/path", AppendTransactionRequest{
		PortfolioID:     "portfolio-id",
		TransactionType: "SELL",
		Ticker:          &ticker,
		Quantity:        &quantity,
		UnitPrice:       &unitPrice,
		Commission:      ZeroMoney(),
		Tax:             ZeroMoney(),
		TradeDate:       "2026-06-26",
	})

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}

func TestAppendTransactionHashIncludesDecimalValues(t *testing.T) {
	ticker := "SBER"
	quantity := decimal.Must("1.00000000")
	firstPrice := Money{Amount: decimal.Must("100.00000000"), Currency: RUB}
	secondPrice := Money{Amount: decimal.Must("101.00000000"), Currency: RUB}

	firstStore := &recordingStore{}
	firstService := NewService(firstStore, fixedClock{})
	_, err := firstService.AppendTransaction(context.Background(), "subject", "valid-key-000001", "/path", AppendTransactionRequest{
		PortfolioID:     "portfolio-id",
		TransactionType: "BUY",
		Ticker:          &ticker,
		Quantity:        &quantity,
		UnitPrice:       &firstPrice,
		Commission:      ZeroMoney(),
		Tax:             ZeroMoney(),
		TradeDate:       "2026-06-26",
	})
	if err != nil {
		t.Fatalf("append first transaction: %v", err)
	}

	secondStore := &recordingStore{}
	secondService := NewService(secondStore, fixedClock{})
	_, err = secondService.AppendTransaction(context.Background(), "subject", "valid-key-000001", "/path", AppendTransactionRequest{
		PortfolioID:     "portfolio-id",
		TransactionType: "BUY",
		Ticker:          &ticker,
		Quantity:        &quantity,
		UnitPrice:       &secondPrice,
		Commission:      ZeroMoney(),
		Tax:             ZeroMoney(),
		TradeDate:       "2026-06-26",
	})
	if err != nil {
		t.Fatalf("append second transaction: %v", err)
	}

	if firstStore.requestHash == secondStore.requestHash {
		t.Fatal("expected different request hashes for different decimal values")
	}
}
