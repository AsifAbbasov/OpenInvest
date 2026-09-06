package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

const stage371SnapshotMethodology = "stage-03-71-position-cost-snapshot-v1"

func rebuildSnapshotStage371(ctx context.Context, tx *sql.Tx, portfolioID string, snapshotDate string, now time.Time) error {
	positionValues, err := portfolioAcquisitionValuesTx(ctx, tx, portfolioID, snapshotDate)
	if err != nil {
		return err
	}

	snapshotID := uuid.NewString()
	result, err := tx.ExecContext(ctx, `
		INSERT INTO analytics.portfolio_snapshots (
			id, portfolio_id, snapshot_date,
			total_value_amount, cash_value_amount, stock_value_amount, bond_value_amount,
			invested_capital_amount, nominal_return_rate, real_return_rate,
			snapshot_version, methodology_version, input_watermark, calculated_at
		)
		WITH ledger AS (
			SELECT
				COALESCE(SUM(CASE
					WHEN te.transaction_type = 'DEPOSIT' THEN te.gross_amount
					WHEN te.transaction_type = 'WITHDRAWAL' THEN -te.gross_amount
					WHEN te.transaction_type = 'BUY' THEN -(te.gross_amount + te.commission_amount + te.tax_amount)
					WHEN te.transaction_type = 'SELL' THEN te.gross_amount - te.commission_amount - te.tax_amount
					ELSE 0
				END), 0) AS cash_value,
				COALESCE(SUM(CASE
					WHEN te.transaction_type = 'BUY' THEN te.gross_amount + te.commission_amount + te.tax_amount
					ELSE 0
				END), 0) AS invested_capital,
				COALESCE(MAX(te.created_at)::text, 'empty') AS watermark
			FROM investment.transaction_entries te
			WHERE te.portfolio_id = $2
				AND te.trade_date <= $3::date
		), computed AS (
			SELECT
				cash_value,
				$5::numeric AS stock_value,
				$6::numeric AS bond_value,
				invested_capital,
				cash_value + $5::numeric + $6::numeric AS total_value,
				CASE WHEN invested_capital > 0
					THEN ((cash_value + $5::numeric + $6::numeric) - invested_capital) / invested_capital
					ELSE 0
				END AS nominal_return_rate,
				watermark
			FROM ledger
		)
		SELECT
			$1, $2, $3::date,
			total_value, cash_value, stock_value, bond_value,
			invested_capital, nominal_return_rate, nominal_return_rate,
			COALESCE((
				SELECT MAX(snapshot_version) + 1
				FROM analytics.portfolio_snapshots
				WHERE portfolio_id = $2
					AND snapshot_date = $3::date
					AND methodology_version = $7
			), 1),
			$7, watermark, $4
		FROM computed
		WHERE
			round(total_value, 8) BETWEEN -99999999999999999999.99999999 AND 99999999999999999999.99999999
			AND round(cash_value, 8) BETWEEN -99999999999999999999.99999999 AND 99999999999999999999.99999999
			AND round(stock_value, 8) BETWEEN -99999999999999999999.99999999 AND 99999999999999999999.99999999
			AND round(bond_value, 8) BETWEEN -99999999999999999999.99999999 AND 99999999999999999999.99999999
			AND round(invested_capital, 8) BETWEEN -99999999999999999999.99999999 AND 99999999999999999999.99999999
			AND round(nominal_return_rate, 8) BETWEEN -99999999999999999999.99999999 AND 99999999999999999999.99999999
	`, snapshotID, portfolioID, snapshotDate, now,
		positionValues.Stock.String(), positionValues.Bond.String(), stage371SnapshotMethodology)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return fmt.Errorf("%w: snapshot financial values exceed NUMERIC(28,8) storage precision", verticalslice.ErrInvalidInput)
	}
	return nil
}
