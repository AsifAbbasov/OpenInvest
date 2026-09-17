package postgres

import (
	"context"
	"database/sql"
)

// Stable repository-owned advisory-lock namespace for OI-NEW-04 deployment finalization.
// Runtime financial/create transactions hold a shared xact lock; finalization/activation takes
// the exclusive xact lock. pg_try_advisory_xact_lock_shared makes new writes fail closed instead
// of waiting behind S5_FINALIZING, while the exclusive lock waits for already-running writes to drain.
const oiNew04FinalizationAdvisoryLockKey int64 = 0x4f494e45573034 // "OINEW04"

func (s *Store) assertReplayWriteWindowTx(ctx context.Context, tx *sql.Tx) error {
	if s == nil || !s.runtimeCapabilityProfile.valid() {
		return ErrReplayStateStale
	}
	var admitted bool
	if err := tx.QueryRowContext(ctx, `SELECT pg_try_advisory_xact_lock_shared($1)`, oiNew04FinalizationAdvisoryLockKey).Scan(&admitted); err != nil {
		return err
	}
	if !admitted {
		return ErrReplayStateStale
	}
	return nil
}

func lockReplayFinalizationTx(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, oiNew04FinalizationAdvisoryLockKey)
	return err
}
