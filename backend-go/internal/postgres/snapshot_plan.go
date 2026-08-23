package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

// planAffectedSnapshotDates returns the exact deterministic union of snapshot dates affected by
// one financial command. Every transaction trade date is rebuilt even when no snapshot existed yet,
// and every existing snapshot on or after the earliest affected trade date is rebuilt because its
// ledger prefix changed. The returned dates are unique and ascending.
func planAffectedSnapshotDates(
	ctx context.Context,
	tx *sql.Tx,
	portfolioID string,
	tradeDates []string,
) ([]string, error) {
	dateSet := make(map[string]struct{}, len(tradeDates))
	for _, raw := range tradeDates {
		date := strings.TrimSpace(raw)
		if date == "" {
			return nil, fmt.Errorf("%w: snapshot rebuild trade date is required", verticalslice.ErrInvalidInput)
		}
		dateSet[date] = struct{}{}
	}
	if len(dateSet) == 0 {
		return []string{}, nil
	}

	inputDates := make([]string, 0, len(dateSet))
	for date := range dateSet {
		inputDates = append(inputDates, date)
	}
	sort.Strings(inputDates)
	earliestDate := inputDates[0]

	rows, err := tx.QueryContext(ctx, `
		SELECT snapshot_date::text
		FROM analytics.portfolio_snapshots
		WHERE portfolio_id = $1
			AND snapshot_date >= $2::date
			AND methodology_version = 'stage-03-02-local-cost-snapshot-v1'
		ORDER BY snapshot_date ASC
	`, portfolioID, earliestDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var snapshotDate string
		if err := rows.Scan(&snapshotDate); err != nil {
			return nil, err
		}
		dateSet[snapshotDate] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	affectedDates := make([]string, 0, len(dateSet))
	for date := range dateSet {
		affectedDates = append(affectedDates, date)
	}
	sort.Strings(affectedDates)
	return affectedDates, nil
}

// rebuildSnapshotPlan rebuilds every planned date exactly once and in deterministic order.
func rebuildSnapshotPlan(
	ctx context.Context,
	tx *sql.Tx,
	portfolioID string,
	affectedDates []string,
	now time.Time,
) error {
	for _, snapshotDate := range affectedDates {
		if err := rebuildSnapshot(ctx, tx, portfolioID, snapshotDate, now); err != nil {
			return err
		}
	}
	return nil
}
