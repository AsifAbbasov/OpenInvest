package verticalslice

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
)

const (
	PortfolioPositionProjectionMethodologyVersion = "portfolio-position-projection-v1"
	MarketValuationUnavailableStatus              = "UNAVAILABLE"
	MarketValuationUnavailableReason              = "NO_APPROVED_MARKET_PRICE_SOURCE"
)

var ErrPositionProjectionUnavailable = errors.New("portfolio position projection is unavailable")

type PortfolioPositionsStore interface {
	GetPortfolioPositions(
		ctx context.Context,
		subjectID string,
		portfolioID string,
		asOfDate string,
	) (PortfolioPositionsProjection, error)
}

type MarketValuationUnavailable struct {
	Status string
	Reason string
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
