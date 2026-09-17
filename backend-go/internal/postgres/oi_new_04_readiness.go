package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

// StageOINew04Ready is a controlled-startup integrity check, not request readiness.
// R0 deliberately does not touch replay relations so pre-migration/pre-activation deployments remain compatible.
// R1/R2 prove that the replay policy relations are readable; an active generation additionally requires R2.
func (s *Store) StageOINew04Ready(ctx context.Context) error {
	if s == nil || s.db == nil || !s.runtimeCapabilityProfile.valid() {
		return fmt.Errorf("%w: invalid OI-NEW-04 application capability", ErrReplayStateStale)
	}
	if s.runtimeCapabilityProfile == RuntimeCapabilityR0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return err
	}
	defer rollback(tx)
	policy, err := currentReplayPolicyTx(ctx, tx)
	if err != nil {
		return err
	}
	if policy.BlockedAfterInvalidation {
		return fmt.Errorf("%w: invalidated replay generation requires drain/revoke to R0 or activation of a new N+1 generation", ErrReplayStateStale)
	}
	if policy.Active && s.runtimeCapabilityProfile != RuntimeCapabilityR2 {
		return fmt.Errorf("%w: active replay policy requires R2 bounded-capable application", ErrReplayStateStale)
	}
	return tx.Commit()
}
