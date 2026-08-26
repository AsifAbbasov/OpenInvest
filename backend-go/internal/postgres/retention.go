package postgres

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

const operationalRetentionCleanupBatchSize = 128

func databaseWallClock(ctx context.Context, tx *sql.Tx) (time.Time, error) {
	var now time.Time
	if err := tx.QueryRowContext(ctx, `SELECT clock_timestamp()`).Scan(&now); err != nil {
		return time.Time{}, err
	}
	return now.UTC(), nil
}

func expiredAt(expiresAt time.Time, decisionTime time.Time) bool {
	return !expiresAt.After(decisionTime)
}

func replayCommandLockScope(command verticalslice.CommandContext, method string) string {
	parts := []string{command.SubjectID, method, command.RequestPath, command.IdempotencyKey}
	var scope strings.Builder
	scope.WriteString("openinvest/idempotency/")
	for _, part := range parts {
		scope.WriteString(strconv.Itoa(len(part)))
		scope.WriteByte(':')
		scope.WriteString(part)
		scope.WriteByte('|')
	}
	return scope.String()
}

func lockReplayCommandScope(
	ctx context.Context,
	tx *sql.Tx,
	command verticalslice.CommandContext,
	method string,
) error {
	_, err := tx.ExecContext(ctx, `
		SELECT pg_advisory_xact_lock(hashtextextended($1, 0))
	`, replayCommandLockScope(command, method))
	return err
}

func cleanupExpiredReplayCommands(ctx context.Context, tx *sql.Tx) (int64, error) {
	result, err := tx.ExecContext(ctx, `
		WITH cleanup_clock AS MATERIALIZED (
			SELECT clock_timestamp() AS at
		),
		candidates AS MATERIALIZED (
			SELECT command_row.id
			FROM investment.command_deduplication command_row
			CROSS JOIN cleanup_clock
			WHERE command_row.expires_at <= cleanup_clock.at
			ORDER BY command_row.expires_at, command_row.id
			LIMIT $1
			FOR UPDATE OF command_row SKIP LOCKED
		)
		DELETE FROM investment.command_deduplication command_row
		USING candidates
		WHERE command_row.id = candidates.id
	`, operationalRetentionCleanupBatchSize)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func cleanupExpiredSessions(ctx context.Context, tx *sql.Tx) (int64, error) {
	result, err := tx.ExecContext(ctx, `
		WITH cleanup_clock AS MATERIALIZED (
			SELECT clock_timestamp() AS at
		),
		candidates AS MATERIALIZED (
			SELECT session_row.id
			FROM identity.sessions session_row
			CROSS JOIN cleanup_clock
			WHERE session_row.expires_at <= cleanup_clock.at
			ORDER BY session_row.expires_at, session_row.id
			LIMIT $1
			FOR UPDATE OF session_row SKIP LOCKED
		)
		DELETE FROM identity.sessions session_row
		USING candidates
		WHERE session_row.id = candidates.id
	`, operationalRetentionCleanupBatchSize)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
