package verticalslice

import (
	"context"
	"errors"
	"strings"
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
	assetFilter AssetSearchFilter
	assets      []AssetSummary
}

func (store *recordingStore) Ping(context.Context) error { return nil }

func (store *recordingStore) SearchAssets(_ context.Context, filter AssetSearchFilter) ([]AssetSummary, error) {
	store.assetFilter = filter
	if len(store.assets) == 0 {
		return []AssetSummary{}, nil
	}
	start := 0
	if filter.AfterTicker != "" {
		for index, asset := range store.assets {
			if asset.Ticker > filter.AfterTicker {
				start = index
				break
			}
			start = index + 1
		}
	}
	end := start + filter.Limit
	if end > len(store.assets) {
		end = len(store.assets)
	}
	return store.assets[start:end], nil
}

func (store *recordingStore) ListPortfolios(context.Context, string, PortfolioFilter) ([]Portfolio, error) {
	return nil, nil
}

func (store *recordingStore) CreatePortfolio(_ context.Context, command CommandContext, _ CreatePortfolioRequest) (Portfolio, error) {
	store.requestHash = command.RequestHash
	return Portfolio{ID: "portfolio-id", BaseCurrency: RUB, Version: 1}, nil
}

func (store *recordingStore) GetPortfolio(context.Context, string, string) (Portfolio, error) {
	return Portfolio{}, nil
}

func (store *recordingStore) ListTransactions(context.Context, string, string, TransactionFilter) ([]Transaction, error) {
	return nil, nil
}

func (store *recordingStore) AppendTransaction(_ context.Context, command CommandContext, _ AppendTransactionRequest) (Transaction, error) {
	store.requestHash = command.RequestHash
	return Transaction{}, nil
}

func (store *recordingStore) AppendImportedTransactions(_ context.Context, command CommandContext, _ AppendImportBatchRequest) ([]Transaction, error) {
	store.requestHash = command.RequestHash
	return []Transaction{}, nil
}

func (store *recordingStore) GetPortfolioSummary(context.Context, string, string, string) (PortfolioSummary, error) {
	return PortfolioSummary{}, nil
}

func TestSearchAssetsRequiresQuery(t *testing.T) {
	service := NewService(&recordingStore{}, fixedClock{})

	_, err := service.SearchAssets(context.Background(), AssetSearchFilter{Query: " "})

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}

func TestSearchAssetsRejectsAssetTypeOutsideFrozenEnum(t *testing.T) {
	service := NewService(&recordingStore{}, fixedClock{})

	_, err := service.SearchAssets(context.Background(), AssetSearchFilter{Query: "SBER", AssetType: "ETF"})

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}

func TestSearchAssetsAllowsOneHundredUnicodeCharacters(t *testing.T) {
	store := &recordingStore{}
	service := NewService(store, fixedClock{})

	query := strings.Repeat("Ж", 100)
	if _, err := service.SearchAssets(context.Background(), AssetSearchFilter{Query: query}); err != nil {
		t.Fatalf("expected 100 unicode characters to be accepted, got %v", err)
	}
}

func TestSearchAssetsRejectsMoreThanOneHundredUnicodeCharacters(t *testing.T) {
	service := NewService(&recordingStore{}, fixedClock{})

	query := strings.Repeat("Ж", 101)
	_, err := service.SearchAssets(context.Background(), AssetSearchFilter{Query: query})

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}

func TestSearchAssetsNormalizesLimit(t *testing.T) {
	store := &recordingStore{}
	service := NewService(store, fixedClock{})

	result, err := service.SearchAssets(context.Background(), AssetSearchFilter{Query: "SBER", Limit: 1000})
	if err != nil {
		t.Fatalf("search assets: %v", err)
	}

	if store.assetFilter.Limit != 101 {
		t.Fatalf("expected over-fetch limit 101, got %d", store.assetFilter.Limit)
	}
	if result.Limit != 100 {
		t.Fatalf("expected response limit 100, got %d", result.Limit)
	}
}

func TestSearchAssetsUsesTickerKeysetAnchor(t *testing.T) {
	store := &recordingStore{
		assets: []AssetSummary{
			{Ticker: "SBER", Name: "Sberbank ordinary shares", AssetType: "STOCK", Currency: RUB, LotSize: decimal.Must("10.00000000")},
			{Ticker: "SU26238RMFS4", Name: "OFZ 26238", AssetType: "BOND", Currency: RUB, LotSize: decimal.Must("1.00000000")},
		},
	}
	service := NewService(store, fixedClock{})

	first, err := service.SearchAssets(context.Background(), AssetSearchFilter{Query: "S", Limit: 1})
	if err != nil {
		t.Fatalf("search first page: %v", err)
	}
	if !first.HasMore || first.NextTicker == nil || len(first.Items) != 1 || first.Items[0].Ticker != "SBER" {
		t.Fatalf("unexpected first page: %+v", first)
	}

	second, err := service.SearchAssets(context.Background(), AssetSearchFilter{Query: "S", AfterTicker: *first.NextTicker, Limit: 1})
	if err != nil {
		t.Fatalf("search second page: %v", err)
	}
	if second.HasMore || second.NextTicker != nil || len(second.Items) != 1 || second.Items[0].Ticker != "SU26238RMFS4" {
		t.Fatalf("unexpected second page: %+v", second)
	}
}

func TestSearchAssetsRejectsInvalidKeysetAnchor(t *testing.T) {
	service := NewService(&recordingStore{}, fixedClock{})

	for _, anchor := range []string{
		"sber",
		"SBER!",
		strings.Repeat("A", 33),
	} {
		_, err := service.SearchAssets(context.Background(), AssetSearchFilter{Query: "SBER", AfterTicker: anchor, Limit: 1})
		if !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("expected invalid input for keyset anchor %q, got %v", anchor, err)
		}
	}
}

func TestGetAssetDefersDetailUntilRuntimeSourceAndRequiredFieldsExist(t *testing.T) {
	service := NewService(&recordingStore{}, fixedClock{})

	_, err := service.GetAsset(context.Background(), "SBER")

	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found while asset card detail is deferred, got %v", err)
	}
}

func TestCreatePortfolioRequiresIdempotencyKey(t *testing.T) {
	service := NewService(&recordingStore{}, fixedClock{})

	_, err := service.CreatePortfolio(context.Background(), RequestContext{}, "subject", "", "/api/v1/portfolios", CreatePortfolioRequest{
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

	_, err := service.AppendTransaction(context.Background(), RequestContext{}, "subject", "valid-key-000001", "/path", AppendTransactionRequest{
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

func TestAppendTransactionRejectsWhitespacePaddedTicker(t *testing.T) {
	service := NewService(&recordingStore{}, fixedClock{})
	ticker := " SBER "
	quantity := decimal.Must("1.00000000")
	unitPrice := Money{Amount: decimal.Must("100.00000000"), Currency: RUB}

	_, err := service.AppendTransaction(context.Background(), RequestContext{}, "subject", "valid-key-000001", "/path", AppendTransactionRequest{
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

	_, err := service.CreatePortfolio(context.Background(), RequestContext{}, "subject", "too-short", "/api/v1/portfolios", CreatePortfolioRequest{
		Name:         "Long-term capital",
		BaseCurrency: RUB,
	})

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}

func TestListTransactionsRejectsTransactionTypeOutsideOpenAPIEnum(t *testing.T) {
	service := NewService(&recordingStore{}, fixedClock{})

	_, err := service.ListTransactions(context.Background(), "subject", "portfolio-id", TransactionFilter{
		TransactionType: "CORRECTION",
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

	_, err := service.AppendTransaction(context.Background(), RequestContext{}, "subject", "valid-key-000001", "/path", AppendTransactionRequest{
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
	_, err := firstService.AppendTransaction(context.Background(), RequestContext{}, "subject", "valid-key-000001", "/path", AppendTransactionRequest{
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
	_, err = secondService.AppendTransaction(context.Background(), RequestContext{}, "subject", "valid-key-000001", "/path", AppendTransactionRequest{
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

func TestAppendImportedTransactionsRejectsDuplicateRows(t *testing.T) {
	service := NewService(&recordingStore{}, fixedClock{})
	gross := Money{Amount: decimal.Must("1000.00000000"), Currency: RUB}
	request := AppendTransactionRequest{
		PortfolioID:     "portfolio-id",
		TransactionType: "DEPOSIT",
		GrossAmount:     &gross,
		Commission:      ZeroMoney(),
		Tax:             ZeroMoney(),
		TradeDate:       "2026-06-26",
	}

	_, err := service.AppendImportedTransactions(context.Background(), RequestContext{}, "subject", "import-batch-key-0001", "/internal/imports/append", AppendImportBatchRequest{
		PortfolioID:    "portfolio-id",
		Transactions:   []AppendTransactionRequest{request, request},
		SourceKind:     "USER_UPLOADED_FILE",
		SourceFileHash: strings.Repeat("a", 64),
	})

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}

func TestAppendImportedTransactionsHashIncludesWholeBatch(t *testing.T) {
	firstGross := Money{Amount: decimal.Must("1000.00000000"), Currency: RUB}
	secondGross := Money{Amount: decimal.Must("2000.00000000"), Currency: RUB}
	firstRequest := AppendImportBatchRequest{
		PortfolioID:        "portfolio-id",
		SourceAccountLabel: "Manual CSV",
		Transactions: []AppendTransactionRequest{{
			PortfolioID:     "portfolio-id",
			TransactionType: "DEPOSIT",
			GrossAmount:     &firstGross,
			Commission:      ZeroMoney(),
			Tax:             ZeroMoney(),
			TradeDate:       "2026-06-26",
		}},
		SourceKind:     "USER_UPLOADED_FILE",
		SourceFileHash: strings.Repeat("a", 64),
		Decisions:      []AppendImportDecision{{RowNumber: 2, Action: "APPROVE"}},
	}
	secondRequest := firstRequest
	secondRequest.Transactions = []AppendTransactionRequest{{
		PortfolioID:     "portfolio-id",
		TransactionType: "DEPOSIT",
		GrossAmount:     &secondGross,
		Commission:      ZeroMoney(),
		Tax:             ZeroMoney(),
		TradeDate:       "2026-06-26",
	}}

	firstStore := &recordingStore{}
	firstService := NewService(firstStore, fixedClock{})
	if _, err := firstService.AppendImportedTransactions(context.Background(), RequestContext{}, "subject", "import-batch-key-0001", "/internal/imports/append", firstRequest); err != nil {
		t.Fatalf("append first import batch: %v", err)
	}

	secondStore := &recordingStore{}
	secondService := NewService(secondStore, fixedClock{})
	if _, err := secondService.AppendImportedTransactions(context.Background(), RequestContext{}, "subject", "import-batch-key-0001", "/internal/imports/append", secondRequest); err != nil {
		t.Fatalf("append second import batch: %v", err)
	}

	if firstStore.requestHash == secondStore.requestHash {
		t.Fatal("expected different request hashes for different import batches")
	}

	thirdRequest := firstRequest
	thirdRequest.SourceAccountLabel = "Another Manual CSV"

	thirdStore := &recordingStore{}
	thirdService := NewService(thirdStore, fixedClock{})
	if _, err := thirdService.AppendImportedTransactions(context.Background(), RequestContext{}, "subject", "import-batch-key-0001", "/internal/imports/append", thirdRequest); err != nil {
		t.Fatalf("append third import batch: %v", err)
	}

	if firstStore.requestHash == thirdStore.requestHash {
		t.Fatal("expected different request hashes for different import source labels")
	}

	fourthRequest := firstRequest
	fourthRequest.Decisions = []AppendImportDecision{{RowNumber: 2, Action: "IGNORE"}}

	fourthStore := &recordingStore{}
	fourthService := NewService(fourthStore, fixedClock{})
	if _, err := fourthService.AppendImportedTransactions(context.Background(), RequestContext{}, "subject", "import-batch-key-0001", "/internal/imports/append", fourthRequest); err != nil {
		t.Fatalf("append fourth import batch: %v", err)
	}

	if firstStore.requestHash == fourthStore.requestHash {
		t.Fatal("expected different request hashes for different import decisions")
	}
}

func TestAppendImportedTransactionsRejectsInvalidSourceFileHash(t *testing.T) {
	service := NewService(&recordingStore{}, fixedClock{})

	_, err := service.AppendImportedTransactions(context.Background(), RequestContext{}, "subject", "import-batch-key-0002", "/internal/imports/append", AppendImportBatchRequest{
		PortfolioID:    "portfolio-id",
		SourceKind:     "USER_UPLOADED_FILE",
		SourceFileHash: "not-a-sha256-digest",
	})

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}
