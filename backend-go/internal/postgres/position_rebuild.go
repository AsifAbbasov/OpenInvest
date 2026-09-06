package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/position"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

var (
	ErrLedgerSequenceUnavailable = errors.New("ledger sequence population is incomplete")
	ErrUnsupportedPositionLedger = errors.New("position ledger contains unsupported correction or reversal state")
)

type acquisitionValueTotals struct {
	Stock decimal.Decimal
	Bond  decimal.Decimal
}

type positionLedgerRow struct {
	AssetID         string
	AssetType       string
	TransactionType string
	Quantity        decimal.Decimal
	UnitPrice       decimal.Decimal
	LedgerSequence  int64
}

func ensureStage371Ready(ctx context.Context, db *sql.DB) error {
	var indexReady bool
	var invalidRows int64
	var duplicateGroups int64
	err := db.QueryRowContext(ctx, `
		SELECT
			EXISTS (
				SELECT 1
				FROM pg_class index_class
				JOIN pg_namespace index_namespace ON index_namespace.oid = index_class.relnamespace
				JOIN pg_index index_meta ON index_meta.indexrelid = index_class.oid
				WHERE index_namespace.nspname = 'investment'
					AND index_class.relname = 'transaction_entries_portfolio_ledger_sequence_uidx'
					AND index_meta.indrelid = 'investment.transaction_entries'::regclass
					AND index_meta.indisunique
					AND index_meta.indisvalid
					AND index_meta.indisready
					AND index_meta.indpred IS NULL
					AND index_meta.indexprs IS NULL
					AND index_meta.indnkeyatts = 2
					AND index_meta.indnatts = 2
					AND (
						SELECT array_agg(attribute.attname::text ORDER BY key.ordinality)
						FROM unnest(index_meta.indkey) WITH ORDINALITY AS key(attnum, ordinality)
						JOIN pg_attribute attribute
							ON attribute.attrelid = index_meta.indrelid
							AND attribute.attnum = key.attnum
					) = ARRAY['portfolio_id', 'ledger_sequence']::text[]
			),
			(SELECT COUNT(*) FROM investment.transaction_entries WHERE ledger_sequence IS NULL OR ledger_sequence <= 0),
			(
				SELECT COUNT(*)
				FROM (
					SELECT portfolio_id, ledger_sequence
					FROM investment.transaction_entries
					WHERE ledger_sequence IS NOT NULL
					GROUP BY portfolio_id, ledger_sequence
					HAVING COUNT(*) > 1
				) duplicates
			)
	`).Scan(&indexReady, &invalidRows, &duplicateGroups)
	if err != nil {
		return err
	}
	if !indexReady || invalidRows != 0 || duplicateGroups != 0 {
		return ErrLedgerSequenceUnavailable
	}
	return nil
}

func nextLedgerSequenceTx(ctx context.Context, tx *sql.Tx, portfolioID string) (int64, error) {
	var maxSequence sql.NullInt64
	var invalidRows int64
	if err := tx.QueryRowContext(ctx, `
		SELECT MAX(ledger_sequence), COUNT(*) FILTER (WHERE ledger_sequence IS NULL OR ledger_sequence <= 0)
		FROM investment.transaction_entries
		WHERE portfolio_id = $1
	`, portfolioID).Scan(&maxSequence, &invalidRows); err != nil {
		return 0, err
	}
	if invalidRows != 0 {
		return 0, ErrLedgerSequenceUnavailable
	}
	if !maxSequence.Valid {
		return 1, nil
	}
	if maxSequence.Int64 == math.MaxInt64 {
		return 0, fmt.Errorf("%w: portfolio ledger sequence exhausted", verticalslice.ErrInvalidInput)
	}
	return maxSequence.Int64 + 1, nil
}

func validatePositionHistoryTx(ctx context.Context, tx *sql.Tx, portfolioID string, assetID *string) error {
	if assetID == nil {
		return nil
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT
			te.transaction_type,
			te.quantity::text,
			te.unit_price_amount::text,
			te.ledger_sequence,
			te.revision,
			te.prior_entry_id IS NOT NULL,
			te.reverses_transaction_id IS NOT NULL
		FROM investment.transaction_entries te
		WHERE te.portfolio_id = $1
			AND te.asset_id = $2
			AND te.transaction_type IN ('BUY', 'SELL')
		ORDER BY te.trade_date ASC, te.ledger_sequence ASC
	`, portfolioID, *assetID)
	if err != nil {
		return err
	}
	defer rows.Close()

	trades := []position.Trade{}
	for rows.Next() {
		var transactionType string
		var quantityText string
		var unitPriceText string
		var ledgerSequence sql.NullInt64
		var revision int
		var corrected bool
		var reversal bool
		if err := rows.Scan(
			&transactionType,
			&quantityText,
			&unitPriceText,
			&ledgerSequence,
			&revision,
			&corrected,
			&reversal,
		); err != nil {
			return err
		}
		if !ledgerSequence.Valid || ledgerSequence.Int64 <= 0 {
			return ErrLedgerSequenceUnavailable
		}
		if revision != 1 || corrected || reversal {
			return ErrUnsupportedPositionLedger
		}
		quantity, err := decimal.FromString(quantityText)
		if err != nil {
			return err
		}
		unitPrice, err := decimal.FromString(unitPriceText)
		if err != nil {
			return err
		}
		trades = append(trades, position.Trade{Type: transactionType, Quantity: quantity, UnitPrice: unitPrice})
	}
	if err := rows.Err(); err != nil {
		return err
	}

	_, err = position.Rebuild(trades)
	switch {
	case errors.Is(err, position.ErrInsufficientQuantity):
		return verticalslice.ErrInsufficientPositionQuantity
	case errors.Is(err, position.ErrDerivedOverflow), errors.Is(err, position.ErrInvalidTrade):
		return fmt.Errorf("%w: position rebuild exceeds canonical Decimal constraints", verticalslice.ErrInvalidInput)
	default:
		return err
	}
}

func portfolioAcquisitionValuesTx(ctx context.Context, tx *sql.Tx, portfolioID string, snapshotDate string) (acquisitionValueTotals, error) {
	if err := assertPortfolioLedgerSequenceCompleteTx(ctx, tx, portfolioID); err != nil {
		return acquisitionValueTotals{}, err
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT
			te.asset_id::text,
			a.asset_type,
			te.transaction_type,
			te.quantity::text,
			te.unit_price_amount::text,
			te.ledger_sequence,
			te.revision,
			te.prior_entry_id IS NOT NULL,
			te.reverses_transaction_id IS NOT NULL
		FROM investment.transaction_entries te
		JOIN investment.assets a ON a.id = te.asset_id
		WHERE te.portfolio_id = $1
			AND te.trade_date <= $2::date
			AND te.transaction_type IN ('BUY', 'SELL')
		ORDER BY te.asset_id ASC, te.trade_date ASC, te.ledger_sequence ASC
	`, portfolioID, snapshotDate)
	if err != nil {
		return acquisitionValueTotals{}, err
	}
	defer rows.Close()

	states := map[string]position.State{}
	assetTypes := map[string]string{}
	for rows.Next() {
		var assetID string
		var assetType string
		var transactionType string
		var quantityText string
		var unitPriceText string
		var ledgerSequence int64
		var revision int
		var corrected bool
		var reversal bool
		if err := rows.Scan(
			&assetID,
			&assetType,
			&transactionType,
			&quantityText,
			&unitPriceText,
			&ledgerSequence,
			&revision,
			&corrected,
			&reversal,
		); err != nil {
			return acquisitionValueTotals{}, err
		}
		if ledgerSequence <= 0 {
			return acquisitionValueTotals{}, ErrLedgerSequenceUnavailable
		}
		if revision != 1 || corrected || reversal {
			return acquisitionValueTotals{}, ErrUnsupportedPositionLedger
		}
		quantity, err := decimal.FromString(quantityText)
		if err != nil {
			return acquisitionValueTotals{}, err
		}
		unitPrice, err := decimal.FromString(unitPriceText)
		if err != nil {
			return acquisitionValueTotals{}, err
		}
		next, err := position.Apply(states[assetID], position.Trade{
			Type:      transactionType,
			Quantity:  quantity,
			UnitPrice: unitPrice,
		})
		if errors.Is(err, position.ErrInsufficientQuantity) {
			return acquisitionValueTotals{}, verticalslice.ErrInsufficientPositionQuantity
		}
		if errors.Is(err, position.ErrDerivedOverflow) || errors.Is(err, position.ErrInvalidTrade) {
			return acquisitionValueTotals{}, fmt.Errorf("%w: position rebuild exceeds canonical Decimal constraints", verticalslice.ErrInvalidInput)
		}
		if err != nil {
			return acquisitionValueTotals{}, err
		}
		states[assetID] = next
		assetTypes[assetID] = assetType
	}
	if err := rows.Err(); err != nil {
		return acquisitionValueTotals{}, err
	}

	totals := acquisitionValueTotals{Stock: decimal.Zero(), Bond: decimal.Zero()}
	for assetID, state := range states {
		if !state.Open {
			continue
		}
		switch assetTypes[assetID] {
		case "stock":
			totals.Stock = totals.Stock.Add(state.AcquisitionBasis)
			if !totals.Stock.FitsStorage() {
				return acquisitionValueTotals{}, fmt.Errorf("%w: stock acquisition basis exceeds NUMERIC(28,8)", verticalslice.ErrInvalidInput)
			}
		case "bond":
			totals.Bond = totals.Bond.Add(state.AcquisitionBasis)
			if !totals.Bond.FitsStorage() {
				return acquisitionValueTotals{}, fmt.Errorf("%w: bond acquisition basis exceeds NUMERIC(28,8)", verticalslice.ErrInvalidInput)
			}
		default:
			return acquisitionValueTotals{}, fmt.Errorf("unsupported asset type %q in position rebuild", assetTypes[assetID])
		}
	}
	return totals, nil
}

func assertPortfolioLedgerSequenceCompleteTx(ctx context.Context, tx *sql.Tx, portfolioID string) error {
	var invalidRows int64
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM investment.transaction_entries
		WHERE portfolio_id = $1
			AND (ledger_sequence IS NULL OR ledger_sequence <= 0)
	`, portfolioID).Scan(&invalidRows); err != nil {
		return err
	}
	if invalidRows != 0 {
		return ErrLedgerSequenceUnavailable
	}
	return nil
}
