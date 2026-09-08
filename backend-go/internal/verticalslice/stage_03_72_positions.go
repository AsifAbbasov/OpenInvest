package verticalslice

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
)

const (
	PortfolioPositionProjectionMethodologyVersion = "portfolio-position-projection-v2-manual-valuation"
	MarketValuationUnavailableStatus              = "UNAVAILABLE"
	MarketValuationAvailableStatus                = "AVAILABLE"
	MarketValuationUnavailableReason              = "NO_APPROVED_MARKET_PRICE_SOURCE"
	ValuationCoveragePartialStatus                = "PARTIAL"
	ValuationCoverageCompleteStatus               = "COMPLETE"
	ManualValuationSourceUserSupplied             = "USER_SUPPLIED"
)

var (
	ErrPositionProjectionUnavailable = errors.New("portfolio position projection is unavailable")
	ErrManualValuationUnavailable    = errors.New("manual valuation persistence is unavailable")
)

type PortfolioPositionsStore interface {
	GetPortfolioPositions(
		ctx context.Context,
		subjectID string,
		portfolioID string,
		asOfDate string,
	) (PortfolioPositionsProjection, error)
}

type ManualValuationRequest struct {
	PortfolioID string
	Ticker      string
	Price       Money
	AsOfDate    string
}

type ManualValuationStore interface {
	UpsertManualValuation(ctx context.Context, subjectID string, request ManualValuationRequest) error
	ClearManualValuation(ctx context.Context, subjectID string, portfolioID string, ticker string) error
}

type Stage376ReadyStore interface {
	Stage376Ready(ctx context.Context) error
}

// MarketValuationUnavailable keeps the Stage 3.72 type name for source compatibility while
// Stage 3.76 extends the state to AVAILABLE / USER_SUPPLIED. Empty strings/pointers are omitted
// by the HTTP mapper rather than fabricated into the public contract.
type MarketValuationUnavailable struct {
	Status           string
	Reason           string
	Source           string
	MarketPrice      *Money
	MarketValue      *Money
	UnrealizedGain   *Money
	UnrealizedReturn *decimal.Decimal
	MarketWeight     *decimal.Decimal
	AsOf             *string
}

type PortfolioValuationSummary struct {
	Status                          string
	ValuedPositions                 int
	TotalOpenPositions              int
	ValuedPositionsMarketValue      Money
	ValuedPositionsAcquisitionBasis Money
	TotalAcquisitionBasis           Money
	UnrealizedGain                  Money
	CashValue                       Money
	CurrentPortfolioValue           *Money
}

type PortfolioPositionProjection struct {
	Ticker                 string
	AssetType              string
	Quantity               decimal.Decimal
	WeightedAverageCost    Money
	AcquisitionBasis       Money
	AcquisitionBasisWeight *decimal.Decimal
	MarketValuation        MarketValuationUnavailable
}

type PortfolioPositionsProjection struct {
	Items                 []PortfolioPositionProjection
	TotalAcquisitionBasis Money
	ValuationSummary      PortfolioValuationSummary
	InputsAsOf            *string
	MethodologyVersion    string
}

func (s *Service) GetPortfolioPositions(
	ctx context.Context,
	subjectID string,
	portfolioID string,
	asOfDate string,
) (PortfolioPositionsProjection, error) {
	if asOfDate != "" {
		if _, err := time.Parse("2006-01-02", asOfDate); err != nil {
			return PortfolioPositionsProjection{}, fmt.Errorf("%w: asOfDate must be YYYY-MM-DD", ErrInvalidInput)
		}
	}
	store, ok := s.store.(PortfolioPositionsStore)
	if !ok {
		return PortfolioPositionsProjection{}, ErrPositionProjectionUnavailable
	}
	return store.GetPortfolioPositions(ctx, subjectID, portfolioID, asOfDate)
}

func (s *Service) UpsertManualValuation(
	ctx context.Context,
	subjectID string,
	request ManualValuationRequest,
) (PortfolioPositionsProjection, error) {
	request.PortfolioID = strings.TrimSpace(request.PortfolioID)
	request.Ticker = strings.TrimSpace(request.Ticker)
	request.AsOfDate = strings.TrimSpace(request.AsOfDate)
	if request.PortfolioID == "" || !tickerPattern.MatchString(request.Ticker) {
		return PortfolioPositionsProjection{}, ErrNotFound
	}
	if request.Price.Currency != RUB || !request.Price.Amount.IsPositive() || !request.Price.Amount.FitsStorage() {
		return PortfolioPositionsProjection{}, fmt.Errorf("%w: manual price must be positive RUB within NUMERIC(28,8)", ErrInvalidInput)
	}
	if _, err := time.Parse("2006-01-02", request.AsOfDate); err != nil {
		return PortfolioPositionsProjection{}, fmt.Errorf("%w: asOfDate must be YYYY-MM-DD", ErrInvalidInput)
	}
	store, ok := s.store.(ManualValuationStore)
	if !ok {
		return PortfolioPositionsProjection{}, ErrManualValuationUnavailable
	}
	if err := store.UpsertManualValuation(ctx, subjectID, request); err != nil {
		return PortfolioPositionsProjection{}, err
	}
	return s.GetPortfolioPositions(ctx, subjectID, request.PortfolioID, "")
}

func (s *Service) ClearManualValuation(
	ctx context.Context,
	subjectID string,
	portfolioID string,
	ticker string,
) (PortfolioPositionsProjection, error) {
	portfolioID = strings.TrimSpace(portfolioID)
	ticker = strings.TrimSpace(ticker)
	if portfolioID == "" || !tickerPattern.MatchString(ticker) {
		return PortfolioPositionsProjection{}, ErrNotFound
	}
	store, ok := s.store.(ManualValuationStore)
	if !ok {
		return PortfolioPositionsProjection{}, ErrManualValuationUnavailable
	}
	if err := store.ClearManualValuation(ctx, subjectID, portfolioID, ticker); err != nil {
		return PortfolioPositionsProjection{}, err
	}
	return s.GetPortfolioPositions(ctx, subjectID, portfolioID, "")
}
