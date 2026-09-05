package postgres

import (
	"context"
	"database/sql"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func (s *Store) CalculateDividendWithReplay(
	ctx context.Context,
	command verticalslice.CommandContext,
	calculation verticalslice.DividendCalculation,
	build verticalslice.DividendReplayBuilder,
) (verticalslice.DividendCalculation, verticalslice.CommandReplayArtifact, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return verticalslice.DividendCalculation{}, verticalslice.CommandReplayArtifact{}, err
	}
	defer rollback(tx)

	reservation, err := reserveReplayCommand(ctx, tx, command, "POST")
	if err != nil {
		return verticalslice.DividendCalculation{}, verticalslice.CommandReplayArtifact{}, err
	}
	if reservation.Duplicate {
		if err := tx.Commit(); err != nil {
			return verticalslice.DividendCalculation{}, verticalslice.CommandReplayArtifact{}, err
		}
		return verticalslice.DividendCalculation{}, reservation.Artifact, nil
	}
	artifact, err := build(calculation)
	if err != nil {
		return verticalslice.DividendCalculation{}, verticalslice.CommandReplayArtifact{}, err
	}
	if err := completeReplayCommand(ctx, tx, reservation.ID, artifact); err != nil {
		return verticalslice.DividendCalculation{}, verticalslice.CommandReplayArtifact{}, err
	}
	if err := tx.Commit(); err != nil {
		return verticalslice.DividendCalculation{}, verticalslice.CommandReplayArtifact{}, err
	}
	return calculation, artifact, nil
}
