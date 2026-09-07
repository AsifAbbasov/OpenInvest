package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
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
	if err := assertPortfolioLedgerSequenceCompleteTx(ctx, tx, portfolioID); err != nil {
		return nil, nil, err
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT
			te.asset_id::text,
			a.ticker,
			a.asset_type,
			te.transaction_type,
			te.quantity::text,
			te.unit_price_amount::text,
			te.trade_date::text,
			te.ledger_sequence,
			te.revision,
			te.prior_entry_id IS NOT NULL,
			te.reverses_transaction_id IS NOT NULL
		FROM investment.transaction_entries te
		JOIN investment.assets a ON a.id = te.asset_id
		WHERE te.portfolio_id = $1
			AND te.transaction_type IN ('BUY', 'SELL')
			AND ($2::text = '' OR te.trade_date <= NULLIF($2::text, '')::date)
		ORDER BY a.ticker ASC, te.asset_id ASC, te.trade_date ASC, te.ledger_sequence ASC
	`, portfolioID, asOfDate)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	positions := make([]rebuiltPortfolioPosition, 0)
	positionIndex := map[string]int{}
	var latestIncludedTradeDate *string

	for rows.Next() {
		var assetID string
		var ticker string
		var assetType string
		var transactionType string
		var quantityText string
		var unitPriceText string
		var tradeDate string
		var ledgerSequence int64
		var revision int
		var corrected bool
		var reversal bool
		if err := rows.Scan(
			&assetID,
			&ticker,
			&assetType,
			&transactionType,
			&quantityText,
			&unitPriceText,
			&tradeDate,
			&ledgerSequence,
			&revision,
			&corrected,
			&reversal,
		); err != nil {
			return nil, nil, err
		}
		if ledgerSequence <= 0 {
			return nil, nil, ErrLedgerSequenceUnavailable
		}
		if revision != 1 || corrected || reversal {
			return nil, nil, ErrUnsupportedPositionLedger
		}

		quantity, err := decimal.FromString(quantityText)
		if err != nil {
			return nil, nil, err
		}
		unitPrice, err := decimal.FromString(unitPriceText)
		if err != nil {
			return nil, nil, err
		}

		index, ok := positionIndex[assetID]
		if !ok {
			index = len(positions)
			positionIndex[assetID] = index
			positions = append(positions, rebuiltPortfolioPosition{
				AssetID:   assetID,
				Ticker:    ticker,
				AssetType: assetType,
				State:     position.Empty(),
			})
		}
		next, err := position.Apply(positions[index].State, position.Trade{
			Type:      transactionType,
			Quantity:  quantity,
			UnitPrice: unitPrice,
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

		if latestIncludedTradeDate == nil || tradeDate > *latestIncludedTradeDate {
			value := tradeDate
			latestIncludedTradeDate = &value
		}
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	return positions, latestIncludedTradeDate, nil
}
