package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"sort"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func (s *Store) GetPortfolioPositions(
	ctx context.Context,
	subjectID string,
	portfolioID string,
	asOfDate string,
) (verticalslice.PortfolioPositionsProjection, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return verticalslice.PortfolioPositionsProjection{}, err
	}
	defer rollback(tx)

	// Preserve the existing anti-enumeration boundary before reading any ledger or asset row.
	if _, err := getPortfolioTx(ctx, tx, subjectID, portfolioID); err != nil {
		return verticalslice.PortfolioPositionsProjection{}, err
	}

	rebuilt, latestIncludedTradeDate, err := rebuildPortfolioPositionsTx(ctx, tx, portfolioID, asOfDate)
	if err != nil {
		return verticalslice.PortfolioPositionsProjection{}, err
	}

	inputsAsOf := latestIncludedTradeDate
	if asOfDate != "" {
		value := asOfDate
		inputsAsOf = &value
	}
	result, err := buildPortfolioPositionsProjection(rebuilt, inputsAsOf)
	if err != nil {
		return verticalslice.PortfolioPositionsProjection{}, err
	}
	if err := tx.Commit(); err != nil {
		return verticalslice.PortfolioPositionsProjection{}, err
	}
	return result, nil
}

func buildPortfolioPositionsProjection(
	rebuilt []rebuiltPortfolioPosition,
	inputsAsOf *string,
) (verticalslice.PortfolioPositionsProjection, error) {
	// Do not let SQL physical order or future helper changes become a public ordering input.
	sort.SliceStable(rebuilt, func(left, right int) bool {
		if rebuilt[left].Ticker == rebuilt[right].Ticker {
			return rebuilt[left].AssetID < rebuilt[right].AssetID
		}
		return rebuilt[left].Ticker < rebuilt[right].Ticker
	})

	items := make([]verticalslice.PortfolioPositionProjection, 0, len(rebuilt))
	totalAcquisitionBasis := decimal.Zero()
	for _, rebuiltPosition := range rebuilt {
		assetType, err := publicAssetType(rebuiltPosition.AssetType)
		if err != nil {
			return verticalslice.PortfolioPositionsProjection{}, err
		}
		if !rebuiltPosition.State.Open {
			continue
		}

		totalAcquisitionBasis = totalAcquisitionBasis.Add(rebuiltPosition.State.AcquisitionBasis)
		if !totalAcquisitionBasis.FitsStorage() {
			return verticalslice.PortfolioPositionsProjection{}, fmt.Errorf("%w: total acquisition basis exceeds NUMERIC(28,8)", verticalslice.ErrInvalidInput)
		}

		items = append(items, verticalslice.PortfolioPositionProjection{
			Ticker:    rebuiltPosition.Ticker,
			AssetType: assetType,
			Quantity:  rebuiltPosition.State.Quantity,
			WeightedAverageCost: verticalslice.Money{
				Amount:   rebuiltPosition.State.WeightedAverageCost,
				Currency: verticalslice.RUB,
			},
			AcquisitionBasis: verticalslice.Money{
				Amount:   rebuiltPosition.State.AcquisitionBasis,
				Currency: verticalslice.RUB,
			},
			MarketValuation: verticalslice.MarketValuationUnavailable{
				Status: verticalslice.MarketValuationUnavailableStatus,
				Reason: verticalslice.MarketValuationUnavailableReason,
			},
		})
	}

	if totalAcquisitionBasis.IsPositive() {
		for index := range items {
			weight, err := items[index].AcquisitionBasis.Amount.Div(totalAcquisitionBasis)
			if err != nil || !weight.FitsStorage() {
				return verticalslice.PortfolioPositionsProjection{}, fmt.Errorf("%w: acquisition basis weight exceeds canonical Decimal constraints", verticalslice.ErrInvalidInput)
			}
			items[index].AcquisitionBasisWeight = decimalPointer(weight)
		}
	}

	return verticalslice.PortfolioPositionsProjection{
		Items: items,
		TotalAcquisitionBasis: verticalslice.Money{
			Amount:   totalAcquisitionBasis,
			Currency: verticalslice.RUB,
		},
		InputsAsOf:         copyStringPointer(inputsAsOf),
		MethodologyVersion: verticalslice.PortfolioPositionProjectionMethodologyVersion,
	}, nil
}

func publicAssetType(assetType string) (string, error) {
	switch assetType {
	case "stock":
		return "STOCK", nil
	case "bond":
		return "BOND", nil
	default:
		return "", fmt.Errorf("%w: unsupported asset type %q in position projection", verticalslice.ErrInvalidInput, assetType)
	}
}

func decimalPointer(value decimal.Decimal) *decimal.Decimal {
	copy := value
	return &copy
}

func copyStringPointer(value *string) *string {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}
