package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
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
	_, _, err := rebuildPortfolioPositionsTx(ctx, tx, portfolioID, "")
	return err
}

func portfolioAcquisitionValuesTx(ctx context.Context, tx *sql.Tx, portfolioID string, snapshotDate string) (acquisitionValueTotals, error) {
	rebuilt, _, err := rebuildPortfolioPositionsTx(ctx, tx, portfolioID, snapshotDate)
	if err != nil {
		return acquisitionValueTotals{}, err
	}

	totals := acquisitionValueTotals{Stock: decimal.Zero(), Bond: decimal.Zero()}
	for _, rebuiltPosition := range rebuilt {
		if !rebuiltPosition.State.Open {
			continue
		}
		switch rebuiltPosition.AssetType {
		case "stock":
			totals.Stock = totals.Stock.Add(rebuiltPosition.State.AcquisitionBasis)
			if !totals.Stock.FitsStorage() {
				return acquisitionValueTotals{}, fmt.Errorf("%w: stock acquisition basis exceeds NUMERIC(28,8)", verticalslice.ErrInvalidInput)
			}
		case "bond":
			totals.Bond = totals.Bond.Add(rebuiltPosition.State.AcquisitionBasis)
			if !totals.Bond.FitsStorage() {
				return acquisitionValueTotals{}, fmt.Errorf("%w: bond acquisition basis exceeds NUMERIC(28,8)", verticalslice.ErrInvalidInput)
			}
		default:
			return acquisitionValueTotals{}, fmt.Errorf("unsupported asset type %q in position rebuild", rebuiltPosition.AssetType)
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
