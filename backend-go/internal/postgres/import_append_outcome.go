package postgres

import (
	"context"
	"database/sql"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func (s *Store) AppendImportedTransactionsWithOutcome(
	ctx context.Context,
	command verticalslice.CommandContext,
	request verticalslice.AppendImportBatchRequest,
) (verticalslice.ImportAppendOutcome, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return verticalslice.ImportAppendOutcome{}, err
	}
	defer rollback(tx)

	if err := lockPortfolioTx(ctx, tx, command.SubjectID, request.PortfolioID); err != nil {
		return verticalslice.ImportAppendOutcome{}, err
	}

	commandID, duplicate, err := reserveCommand(ctx, tx, command, "POST")
	if err != nil {
		return verticalslice.ImportAppendOutcome{}, err
	}
	if duplicate {
		// This outcome API is intentionally exact. Legacy command rows do not persist the original
		// snapshot-date result, so reconstructing it from current projection state would be unsafe.
		return verticalslice.ImportAppendOutcome{}, ErrUnsupportedDuplicate
	}

	outcome, err := appendImportedTransactionsMutation(ctx, tx, command, commandID, request)
	if err != nil {
		return verticalslice.ImportAppendOutcome{}, err
	}
	if err := completeCommand(ctx, tx, commandID); err != nil {
		return verticalslice.ImportAppendOutcome{}, err
	}
	if err := tx.Commit(); err != nil {
		return verticalslice.ImportAppendOutcome{}, err
	}
	return outcome, nil
}

func (s *Store) AppendImportedTransactionsWithOutcomeReplay(
	ctx context.Context,
	command verticalslice.CommandContext,
	request verticalslice.AppendImportBatchRequest,
	build verticalslice.ImportedTransactionsOutcomeReplayBuilder,
) ([]verticalslice.Transaction, verticalslice.CommandReplayArtifact, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, verticalslice.CommandReplayArtifact{}, err
	}
	defer rollback(tx)

	// Completed replay is resolved before mutable portfolio state. New commands remain fully atomic:
	// reservation, ledger inserts, exact snapshot plan, rebuilds, response artifact, and commit.
	reservation, err := reserveReplayCommand(ctx, tx, command, "POST")
	if err != nil {
		return nil, verticalslice.CommandReplayArtifact{}, err
	}
	if reservation.Duplicate {
		if err := tx.Commit(); err != nil {
			return nil, verticalslice.CommandReplayArtifact{}, err
		}
		return nil, reservation.Artifact, nil
	}

	if err := lockPortfolioTx(ctx, tx, command.SubjectID, request.PortfolioID); err != nil {
		return nil, verticalslice.CommandReplayArtifact{}, err
	}

	outcome, err := appendImportedTransactionsMutation(ctx, tx, command, reservation.ID, request)
	if err != nil {
		return nil, verticalslice.CommandReplayArtifact{}, err
	}
	artifact, err := build(outcome)
	if err != nil {
		return nil, verticalslice.CommandReplayArtifact{}, err
	}
	if err := completeReplayCommand(ctx, tx, reservation.ID, artifact); err != nil {
		return nil, verticalslice.CommandReplayArtifact{}, err
	}
	if err := tx.Commit(); err != nil {
		return nil, verticalslice.CommandReplayArtifact{}, err
	}
	return outcome.Transactions, artifact, nil
}

func appendImportedTransactionsMutation(
	ctx context.Context,
	tx *sql.Tx,
	command verticalslice.CommandContext,
	commandID string,
	request verticalslice.AppendImportBatchRequest,
) (verticalslice.ImportAppendOutcome, error) {
	for _, transactionRequest := range request.Transactions {
		duplicate, err := importIdentityExists(ctx, tx, transactionRequest, request.SourceAccountLabel)
		if err != nil {
			return verticalslice.ImportAppendOutcome{}, err
		}
		if duplicate {
			return verticalslice.ImportAppendOutcome{}, verticalslice.ErrInvalidInput
		}
		legacyOrManualDuplicate, err := legacyOrManualEquivalentTransactionExists(ctx, tx, transactionRequest)
		if err != nil {
			return verticalslice.ImportAppendOutcome{}, err
		}
		if legacyOrManualDuplicate {
			return verticalslice.ImportAppendOutcome{}, verticalslice.ErrInvalidInput
		}
		conflict, err := nearConflictTransactionExists(ctx, tx, transactionRequest)
		if err != nil {
			return verticalslice.ImportAppendOutcome{}, err
		}
		if conflict {
			return verticalslice.ImportAppendOutcome{}, verticalslice.ErrInvalidInput
		}
	}

	if err := ensureAssets(ctx, tx, request.Transactions); err != nil {
		return verticalslice.ImportAppendOutcome{}, err
	}

	entryIDs := make([]string, 0, len(request.Transactions))
	tradeDates := make([]string, 0, len(request.Transactions))
	for index, transactionRequest := range request.Transactions {
		entryID, err := importEntryID(commandID, index)
		if err != nil {
			return verticalslice.ImportAppendOutcome{}, err
		}
		if err := insertTransactionEntryWithID(
			ctx,
			tx,
			command,
			transactionRequest,
			entryID,
			request.SourceKind,
			request.SourceFileHash,
			request.SourceAccountLabel,
		); err != nil {
			return verticalslice.ImportAppendOutcome{}, err
		}
		entryIDs = append(entryIDs, entryID)
		tradeDates = append(tradeDates, transactionRequest.TradeDate)
	}

	affectedDates, err := planAffectedSnapshotDates(ctx, tx, request.PortfolioID, tradeDates)
	if err != nil {
		return verticalslice.ImportAppendOutcome{}, err
	}
	if err := rebuildSnapshotPlan(ctx, tx, request.PortfolioID, affectedDates, command.Now); err != nil {
		return verticalslice.ImportAppendOutcome{}, err
	}
	if err := recordImportAppendAudit(ctx, tx, command, request.PortfolioID); err != nil {
		return verticalslice.ImportAppendOutcome{}, err
	}

	transactions := make([]verticalslice.Transaction, 0, len(entryIDs))
	for _, entryID := range entryIDs {
		transaction, err := getTransactionByEntryTx(ctx, tx, request.PortfolioID, entryID)
		if err != nil {
			return verticalslice.ImportAppendOutcome{}, err
		}
		transactions = append(transactions, transaction)
	}

	return verticalslice.ImportAppendOutcome{
		Transactions:         transactions,
		SnapshotDatesRebuilt: append([]string(nil), affectedDates...),
	}, nil
}
