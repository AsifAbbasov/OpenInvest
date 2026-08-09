package verticalslice

import (
	"context"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
)

const RUB = "RUB"

type Clock interface {
	Now() time.Time
}

type SystemClock struct{}

func (SystemClock) Now() time.Time {
	return time.Now().UTC()
}

type Store interface {
	Ping(ctx context.Context) error
	SearchAssets(ctx context.Context, filter AssetSearchFilter) ([]AssetSummary, error)
	ListPortfolios(ctx context.Context, subjectID string, filter PortfolioFilter) ([]Portfolio, error)
	CreatePortfolio(ctx context.Context, command CommandContext, request CreatePortfolioRequest) (Portfolio, error)
	GetPortfolio(ctx context.Context, subjectID string, portfolioID string) (Portfolio, error)
	ListTransactions(ctx context.Context, subjectID string, portfolioID string, filter TransactionFilter) ([]Transaction, error)
	AppendTransaction(ctx context.Context, command CommandContext, request AppendTransactionRequest) (Transaction, error)
	AppendImportedTransactions(ctx context.Context, command CommandContext, request AppendImportBatchRequest) ([]Transaction, error)
	GetPortfolioSummary(ctx context.Context, subjectID string, portfolioID string, asOfDate string) (PortfolioSummary, error)
}

type CommandContext struct {
	SubjectID      string
	IdempotencyKey string
	RequestHash    string
	RequestPath    string
	RequestID      string
	TraceID        string
	Now            time.Time
}

type RequestContext struct {
	RequestID string
	TraceID   string
}

type AssetSearchFilter struct {
	Query       string
	AssetType   string
	AfterTicker string
	Limit       int
}

type AssetSearchResult struct {
	Items      []AssetSummary
	NextTicker *string
	HasMore    bool
	Limit      int
}

type PortfolioFilter struct {
	Limit           int
	BeforeUpdatedAt *time.Time
	BeforeID        string
}

type AssetSummary struct {
	Ticker    string
	Name      string
	AssetType string
	Currency  string
	LotSize   decimal.Decimal
	LastPrice *Money
}

type CreatePortfolioRequest struct {
	Name         string
	BaseCurrency string
}

type AppendTransactionRequest struct {
	PortfolioID     string
	TransactionType string
	Ticker          *string
	Quantity        *decimal.Decimal
	UnitPrice       *Money
	GrossAmount     *Money
	Commission      Money
	Tax             Money
	TradeDate       string
	SettlementDate  *string
	Note            *string
}

type AppendImportBatchRequest struct {
	PortfolioID        string
	Transactions       []AppendTransactionRequest
	SourceKind         string
	SourceAccountLabel string
	SourceFileHash     string
	Decisions          []AppendImportDecision
}

type AppendImportDecision struct {
	RowNumber int
	Action    string
}

type TransactionFilter struct {
	TransactionType string
	FromDate        string
	ToDate          string
	Limit           int
	BeforeTradeDate string
	BeforeEntryID   string
}

type Money struct {
	Amount   decimal.Decimal
	Currency string
}

func (m Money) Sub(other Money) Money {
	return Money{Amount: m.Amount.Sub(other.Amount), Currency: RUB}
}

type Portfolio struct {
	ID           string
	Name         string
	BaseCurrency string
	Version      int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Transaction struct {
	ID              string
	EntryID         string
	PortfolioID     string
	TransactionType string
	Status          string
	Ticker          *string
	Quantity        *decimal.Decimal
	UnitPrice       *Money
	GrossAmount     Money
	Commission      Money
	Tax             Money
	TradeDate       string
	SettlementDate  *string
	Note            *string
	Revision        int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type PortfolioSummary struct {
	PortfolioID        string
	AsOfDate           string
	TotalValue         Money
	CashValue          Money
	StockValue         Money
	BondValue          Money
	InvestedCapital    Money
	DividendsReceived  Money
	CouponsReceived    Money
	NominalReturnRate  decimal.Decimal
	XIRR               *decimal.Decimal
	RealReturn         RealReturn
	PurchasingPower    PurchasingPower
	Positions          []PortfolioPosition
	MethodologyVersion string
	CalculatedAt       time.Time
}

type RealReturn struct {
	NominalReturnRate decimal.Decimal
	InflationRate     decimal.Decimal
	RealReturnRate    decimal.Decimal
	NominalGain       Money
	RealGain          Money
	FromDate          string
	ToDate            string
	Methodology       string
}

type PurchasingPower struct {
	PortfolioValue Money
	AsOfDate       string
	Equivalents    []PurchasingPowerEquivalent
}

type PurchasingPowerEquivalent struct {
	Category  string
	Label     string
	UnitPrice Money
	Quantity  decimal.Decimal
}

type PortfolioPosition struct {
	Ticker              string
	AssetType           string
	Quantity            decimal.Decimal
	WeightedAverageCost Money
	MarketPrice         Money
	MarketValue         Money
	UnrealizedGain      Money
	Weight              decimal.Decimal
}
