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
	cashValue, investedCapital, watermark, err := effectiveSnapshotCashTx(ctx, tx, portfolioID, snapshotDate)
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
        WITH computed AS (
            SELECT
                $5::numeric AS cash_value,
                $6::numeric AS stock_value,
                $7::numeric AS bond_value,
                $8::numeric AS invested_capital,
                $5::numeric + $6::numeric + $7::numeric AS total_value,
                CASE WHEN $8::numeric > 0
                    THEN (($5::numeric + $6::numeric + $7::numeric) - $8::numeric) / $8::numeric
                    ELSE 0
                END AS nominal_return_rate
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
                  AND methodology_version = $9
            ), 1),
            $9, $10, $4
        FROM computed
        WHERE
            round(total_value, 8) BETWEEN -99999999999999999999.99999999 AND 99999999999999999999.99999999
            AND round(cash_value, 8) BETWEEN -99999999999999999999.99999999 AND 99999999999999999999.99999999
            AND round(stock_value, 8) BETWEEN -99999999999999999999.99999999 AND 99999999999999999999.99999999
            AND round(bond_value, 8) BETWEEN -99999999999999999999.99999999 AND 99999999999999999999.99999999
            AND round(invested_capital, 8) BETWEEN -99999999999999999999.99999999 AND 99999999999999999999.99999999
            AND round(nominal_return_rate, 8) BETWEEN -99999999999999999999.99999999 AND 99999999999999999999.99999999
    `, snapshotID, portfolioID, snapshotDate, now,
		cashValue.String(), positionValues.Stock.String(), positionValues.Bond.String(), investedCapital.String(),
		stage371SnapshotMethodology, watermark)
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
