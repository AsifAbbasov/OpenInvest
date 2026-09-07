package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/openinvest/openinvest/backend-go/internal/position"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

type rebuiltPortfolioPosition struct {
	AssetID   string
	Ticker    string
	AssetType string
	State     position.State
}

func rebuildPortfolioPositionsTx(
	ctx context.Context,
	tx *sql.Tx,
	portfolioID string,
	asOfDate string,
) ([]rebuiltPortfolioPosition, *string, error) {
	ledgerRows, err := effectiveLedgerRowsTx(ctx, tx, portfolioID, asOfDate)
	if err != nil {
		return nil, nil, err
	}

	positions := make([]rebuiltPortfolioPosition, 0)
	positionIndex := map[string]int{}
	var latestIncludedTradeDate *string
	for _, row := range ledgerRows {
		if row.TransactionType != "BUY" && row.TransactionType != "SELL" {
			continue
		}
		if row.AssetID == nil || row.Ticker == nil || row.AssetType == nil || row.Quantity == nil || row.UnitPrice == nil {
			return nil, nil, ErrUnsupportedPositionLedger
		}
		index, ok := positionIndex[*row.AssetID]
		if !ok {
			index = len(positions)
			positionIndex[*row.AssetID] = index
			positions = append(positions, rebuiltPortfolioPosition{
				AssetID:   *row.AssetID,
				Ticker:    *row.Ticker,
				AssetType: *row.AssetType,
				State:     position.Empty(),
			})
		}
		next, err := position.Apply(positions[index].State, position.Trade{
			Type:      row.TransactionType,
			Quantity:  *row.Quantity,
			UnitPrice: *row.UnitPrice,
		})
		switch {
		case errors.Is(err, position.ErrInsufficientQuantity):
			return nil, nil, verticalslice.ErrInsufficientPositionQuantity
		case errors.Is(err, position.ErrDerivedOverflow), errors.Is(err, position.ErrInvalidTrade):
			return nil, nil, fmt.Errorf("%w: position rebuild exceeds canonical Decimal constraints", verticalslice.ErrInvalidInput)
		case err != nil:
			return nil, nil, err
		}
		positions[index].State = next
		if latestIncludedTradeDate == nil || row.TradeDate > *latestIncludedTradeDate {
			value := row.TradeDate
			latestIncludedTradeDate = &value
		}
	}
	return positions, latestIncludedTradeDate, nil
}
