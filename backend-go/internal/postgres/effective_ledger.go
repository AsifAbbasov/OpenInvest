package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
)

type effectiveLedgerRow struct {
	EntryID         string
	TransactionID   string
	AssetID         *string
	Ticker          *string
	AssetType       *string
	TransactionType string
	Quantity        *decimal.Decimal
	UnitPrice       *decimal.Decimal
	GrossAmount     decimal.Decimal
	Commission      decimal.Decimal
	Tax             decimal.Decimal
	TradeDate       string
	LedgerSequence  int64
	Revision        int
	UpdatedAt       time.Time
}

func validateEffectiveLedgerShapeTx(ctx context.Context, tx *sql.Tx, portfolioID string) error {
	if err := assertPortfolioLedgerSequenceCompleteTx(ctx, tx, portfolioID); err != nil {
		return err
	}
	var malformedRevisions int64
	if err := tx.QueryRowContext(ctx, `
        WITH ordered AS (
            SELECT
                transaction_id,
                revision,
                prior_entry_id,
                lag(entry_id) OVER (PARTITION BY transaction_id ORDER BY revision) AS previous_entry_id,
                row_number() OVER (PARTITION BY transaction_id ORDER BY revision) AS expected_revision
            FROM investment.transaction_entries
            WHERE portfolio_id = $1
              AND reverses_transaction_id IS NULL
        )
        SELECT COUNT(*)
        FROM ordered
        WHERE revision <> expected_revision
           OR (revision = 1 AND prior_entry_id IS NOT NULL)
           OR (revision > 1 AND prior_entry_id IS DISTINCT FROM previous_entry_id)
    `, portfolioID).Scan(&malformedRevisions); err != nil {
		return err
	}

	var malformedReversals int64
	if err := tx.QueryRowContext(ctx, `
        SELECT COUNT(*)
        FROM investment.transaction_entries reversal
        WHERE reversal.portfolio_id = $1
          AND reversal.reverses_transaction_id IS NOT NULL
          AND (
              reversal.revision <> 1
              OR reversal.prior_entry_id IS NOT NULL
              OR NOT EXISTS (
                  SELECT 1
                  FROM investment.transaction_entries target
                  WHERE target.portfolio_id = reversal.portfolio_id
                    AND target.transaction_id = reversal.reverses_transaction_id
                    AND target.reverses_transaction_id IS NULL
              )
          )
    `, portfolioID).Scan(&malformedReversals); err != nil {
		return err
	}

	var duplicateReversals int64
	if err := tx.QueryRowContext(ctx, `
        SELECT COUNT(*)
        FROM (
            SELECT reverses_transaction_id
            FROM investment.transaction_entries
            WHERE portfolio_id = $1
              AND reverses_transaction_id IS NOT NULL
            GROUP BY reverses_transaction_id
            HAVING COUNT(*) > 1
        ) duplicate_reversals
    `, portfolioID).Scan(&duplicateReversals); err != nil {
		return err
	}
	if malformedRevisions != 0 || malformedReversals != 0 || duplicateReversals != 0 {
		return ErrUnsupportedPositionLedger
	}
	return nil
}

func effectiveLedgerRowsTx(ctx context.Context, tx *sql.Tx, portfolioID string, asOfDate string) ([]effectiveLedgerRow, error) {
	if err := validateEffectiveLedgerShapeTx(ctx, tx, portfolioID); err != nil {
		return nil, err
	}
	rows, err := tx.QueryContext(ctx, `
        WITH revisioned AS (
            SELECT
                te.*,
                min(te.ledger_sequence) OVER (PARTITION BY te.transaction_id) AS logical_sequence,
                row_number() OVER (PARTITION BY te.transaction_id ORDER BY te.revision DESC) AS latest_rank
            FROM investment.transaction_entries te
            WHERE te.portfolio_id = $1
              AND te.reverses_transaction_id IS NULL
        ), reversals AS (
            SELECT reverses_transaction_id, min(trade_date) AS effective_date
            FROM investment.transaction_entries
            WHERE portfolio_id = $1
              AND reverses_transaction_id IS NOT NULL
            GROUP BY reverses_transaction_id
        )
        SELECT
            te.entry_id::text,
            te.transaction_id::text,
            te.asset_id::text,
            a.ticker,
            a.asset_type,
            te.transaction_type,
            te.quantity::text,
            te.unit_price_amount::text,
            te.gross_amount::text,
            te.commission_amount::text,
            te.tax_amount::text,
            te.trade_date::text,
            te.logical_sequence,
            te.revision,
            te.created_at
        FROM revisioned te
        LEFT JOIN reversals reversal ON reversal.reverses_transaction_id = te.transaction_id
        LEFT JOIN investment.assets a ON a.id = te.asset_id
        WHERE te.latest_rank = 1
          AND ($2::text = '' OR te.trade_date <= NULLIF($2::text, '')::date)
          AND (
              reversal.reverses_transaction_id IS NULL
              OR ($2::text <> '' AND reversal.effective_date > NULLIF($2::text, '')::date)
          )
        ORDER BY te.trade_date ASC, te.logical_sequence ASC
    `, portfolioID, asOfDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []effectiveLedgerRow{}
	for rows.Next() {
		var row effectiveLedgerRow
		var assetID, ticker, assetType sql.NullString
		var quantity, unitPrice sql.NullString
		var gross, commission, tax string
		if err := rows.Scan(
			&row.EntryID, &row.TransactionID, &assetID, &ticker, &assetType,
			&row.TransactionType, &quantity, &unitPrice, &gross, &commission, &tax,
			&row.TradeDate, &row.LedgerSequence, &row.Revision, &row.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if row.LedgerSequence <= 0 {
			return nil, ErrLedgerSequenceUnavailable
		}
		if assetID.Valid {
			value := assetID.String
			row.AssetID = &value
		}
		if ticker.Valid {
			value := ticker.String
			row.Ticker = &value
		}
		if assetType.Valid {
			value := assetType.String
			row.AssetType = &value
		}
		if quantity.Valid {
			parsed, err := decimal.FromString(quantity.String)
			if err != nil {
				return nil, err
			}
			row.Quantity = &parsed
		}
		if unitPrice.Valid {
			parsed, err := decimal.FromString(unitPrice.String)
			if err != nil {
				return nil, err
			}
			row.UnitPrice = &parsed
		}
		var parseErr error
		row.GrossAmount, parseErr = decimal.FromString(gross)
		if parseErr != nil {
			return nil, parseErr
		}
		row.Commission, parseErr = decimal.FromString(commission)
		if parseErr != nil {
			return nil, parseErr
		}
		row.Tax, parseErr = decimal.FromString(tax)
		if parseErr != nil {
			return nil, parseErr
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func effectiveSnapshotCashTx(ctx context.Context, tx *sql.Tx, portfolioID string, snapshotDate string) (decimal.Decimal, decimal.Decimal, string, error) {
	rows, err := effectiveLedgerRowsTx(ctx, tx, portfolioID, snapshotDate)
	if err != nil {
		return decimal.Zero(), decimal.Zero(), "", err
	}
	cash := decimal.Zero()
	invested := decimal.Zero()
	for _, row := range rows {
		switch row.TransactionType {
		case "DEPOSIT":
			cash = cash.Add(row.GrossAmount)
		case "WITHDRAWAL":
			cash = cash.Sub(row.GrossAmount)
		case "BUY":
			outflow := row.GrossAmount.Add(row.Commission).Add(row.Tax)
			cash = cash.Sub(outflow)
			invested = invested.Add(outflow)
		case "SELL":
			inflow := row.GrossAmount.Sub(row.Commission).Sub(row.Tax)
			cash = cash.Add(inflow)
		}
		if !cash.FitsStorage() || !invested.FitsStorage() {
			return decimal.Zero(), decimal.Zero(), "", fmt.Errorf("snapshot financial values exceed NUMERIC(28,8)")
		}
	}
	var watermark string
	if err := tx.QueryRowContext(ctx, `
        SELECT COALESCE(MAX(created_at)::text, 'empty')
        FROM investment.transaction_entries
        WHERE portfolio_id = $1 AND trade_date <= $2::date
    `, portfolioID, snapshotDate).Scan(&watermark); err != nil {
		return decimal.Zero(), decimal.Zero(), "", err
	}
	return cash, invested, watermark, nil
}
