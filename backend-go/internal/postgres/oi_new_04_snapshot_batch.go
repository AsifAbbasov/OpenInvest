package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

const (
	// replaySnapshotDateBound is the defensive per-command affected-date ceiling for ACTIVE C'.
	// No smaller product bound is independently justified today, so it intentionally matches the
	// already-approved mutable raw-ledger work bound. The separate SQL-statement bound below removes
	// the former one-round-trip-per-date amplification even when D reaches this ceiling.
	replaySnapshotDateBound = replayBoundRawRows

	// ACTIVE C' persists all admitted snapshot versions with one set-based INSERT statement.
	replaySnapshotSQLStatementBound = 1
)

type replaySnapshotBatchRow struct {
	ID              string `json:"id"`
	SnapshotDate    string `json:"snapshot_date"`
	CashValue       string `json:"cash_value"`
	StockValue      string `json:"stock_value"`
	BondValue       string `json:"bond_value"`
	InvestedCapital string `json:"invested_capital"`
	InputWatermark  string `json:"input_watermark"`
}

func insertReplaySnapshotsBatchTx(
	ctx context.Context,
	tx *sql.Tx,
	portfolioID string,
	affectedDates []string,
	now time.Time,
	states []replayWorkingState,
) error {
	if len(affectedDates) != len(states) {
		return ErrReplayStateStale
	}
	if len(affectedDates) == 0 {
		return nil
	}
	if len(affectedDates) > replaySnapshotDateBound {
		return ErrRetroactiveReplayWindowExceeded
	}

	rows := make([]replaySnapshotBatchRow, 0, len(affectedDates))
	for index, snapshotDate := range affectedDates {
		if err := ctx.Err(); err != nil {
			return err
		}
		state := states[index]
		positionValues, err := replayAcquisitionTotals(state)
		if err != nil {
			return err
		}
		cashValue := state.Financial.Amounts.netCashFlow()
		invested := state.Financial.InvestedCapital
		if !cashValue.FitsStorage() || !invested.FitsStorage() {
			return verticalslice.ErrInvalidInput
		}
		rows = append(rows, replaySnapshotBatchRow{
			ID:              uuid.NewString(),
			SnapshotDate:    snapshotDate,
			CashValue:       cashValue.String(),
			StockValue:      positionValues.Stock.String(),
			BondValue:       positionValues.Bond.String(),
			InvestedCapital: invested.String(),
			InputWatermark:  state.Financial.InputWatermark,
		})
	}

	payload, err := json.Marshal(rows)
	if err != nil {
		return err
	}

	countReplaySnapshotSQLStatement(ctx)
	result, err := tx.ExecContext(ctx, `
		WITH input AS (
			SELECT *
			FROM jsonb_to_recordset($2::jsonb) AS row_data(
				id uuid,
				snapshot_date date,
				cash_value numeric,
				stock_value numeric,
				bond_value numeric,
				invested_capital numeric,
				input_watermark text
			)
		), versions AS (
			SELECT
				i.snapshot_date,
				COALESCE(MAX(ps.snapshot_version), 0) + 1 AS snapshot_version
			FROM input i
			LEFT JOIN analytics.portfolio_snapshots ps
			  ON ps.portfolio_id = $1::uuid
			 AND ps.snapshot_date = i.snapshot_date
			 AND ps.methodology_version = $3
			GROUP BY i.snapshot_date
		), computed AS (
			SELECT
				i.id,
				i.snapshot_date,
				i.cash_value,
				i.stock_value,
				i.bond_value,
				i.invested_capital,
				i.cash_value + i.stock_value + i.bond_value AS total_value,
				CASE WHEN i.invested_capital > 0
					THEN ((i.cash_value + i.stock_value + i.bond_value) - i.invested_capital) / i.invested_capital
					ELSE 0
				END AS nominal_return_rate,
				v.snapshot_version,
				i.input_watermark
			FROM input i
			JOIN versions v USING (snapshot_date)
		)
		INSERT INTO analytics.portfolio_snapshots (
			id, portfolio_id, snapshot_date,
			total_value_amount, cash_value_amount, stock_value_amount, bond_value_amount,
			invested_capital_amount, nominal_return_rate, real_return_rate,
			snapshot_version, methodology_version, input_watermark, calculated_at
		)
		SELECT
			id, $1::uuid, snapshot_date,
			total_value, cash_value, stock_value, bond_value,
			invested_capital, nominal_return_rate, nominal_return_rate,
			snapshot_version, $3, input_watermark, $4
		FROM computed
		WHERE
			round(total_value, 8) BETWEEN -99999999999999999999.99999999 AND 99999999999999999999.99999999
			AND round(cash_value, 8) BETWEEN -99999999999999999999.99999999 AND 99999999999999999999.99999999
			AND round(stock_value, 8) BETWEEN -99999999999999999999.99999999 AND 99999999999999999999.99999999
			AND round(bond_value, 8) BETWEEN -99999999999999999999.99999999 AND 99999999999999999999.99999999
			AND round(invested_capital, 8) BETWEEN -99999999999999999999.99999999 AND 99999999999999999999.99999999
			AND round(nominal_return_rate, 8) BETWEEN -99999999999999999999.99999999 AND 99999999999999999999.99999999
		ORDER BY snapshot_date
	`, portfolioID, string(payload), stage371SnapshotMethodology, now)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != int64(len(rows)) {
		return fmt.Errorf("%w: snapshot financial values exceed NUMERIC(28,8) storage precision", verticalslice.ErrInvalidInput)
	}
	countReplaySnapshotWrites(ctx, len(rows))
	if err := maybeReplayTestFaultAfterSnapshotSQL(ctx); err != nil {
		return err
	}
	return nil
}
