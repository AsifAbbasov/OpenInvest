package postgres

import (
	"context"
	"database/sql"
	"strings"

	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func (s *Store) Stage371Ready(ctx context.Context) error {
	return ensureStage371Ready(ctx, s.db)
}

func (s *Store) AppendTransactionStage371(
	ctx context.Context,
	command verticalslice.CommandContext,
	request verticalslice.AppendTransactionRequest,
) (verticalslice.Transaction, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return verticalslice.Transaction{}, err
	}
	defer rollback(tx)

	if err := lockPortfolioTx(ctx, tx, command.SubjectID, request.PortfolioID); err != nil {
		return verticalslice.Transaction{}, err
	}

	entryID, duplicate, err := reserveCommand(ctx, tx, command, "POST")
	if err != nil {
		return verticalslice.Transaction{}, err
	}
	if duplicate {
		transaction, err := getTransactionByEntryTx(ctx, tx, request.PortfolioID, entryID)
		if err != nil {
			return verticalslice.Transaction{}, err
		}
		return transaction, tx.Commit()
	}

	transaction, err := appendTransactionStage371Mutation(ctx, tx, command, request, entryID)
	if err != nil {
		return verticalslice.Transaction{}, err
	}
	if err := completeCommand(ctx, tx, entryID); err != nil {
		return verticalslice.Transaction{}, err
	}
	if err := tx.Commit(); err != nil {
		return verticalslice.Transaction{}, err
	}
	return transaction, nil
}

func (s *Store) AppendTransactionWithReplayStage371(
	ctx context.Context,
	command verticalslice.CommandContext,
	request verticalslice.AppendTransactionRequest,
	build verticalslice.TransactionReplayBuilder,
) (verticalslice.Transaction, verticalslice.CommandReplayArtifact, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err
	}
	defer rollback(tx)

	reservation, err := reserveReplayCommand(ctx, tx, command, "POST")
	if err != nil {
		return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err
	}
	if reservation.Duplicate {
		if err := tx.Commit(); err != nil {
			return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err
		}
		return verticalslice.Transaction{}, reservation.Artifact, nil
	}

	if err := lockPortfolioTx(ctx, tx, command.SubjectID, request.PortfolioID); err != nil {
		return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err
	}

	transaction, err := appendTransactionStage371Mutation(ctx, tx, command, request, reservation.ID)
	if err != nil {
		return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err
	}
	artifact, err := build(transaction)
	if err != nil {
		return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err
	}
	if err := completeReplayCommand(ctx, tx, reservation.ID, artifact); err != nil {
		return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err
	}
	if err := tx.Commit(); err != nil {
		return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err
	}
	return transaction, artifact, nil
}

func appendTransactionStage371Mutation(
	ctx context.Context,
	tx *sql.Tx,
	command verticalslice.CommandContext,
	request verticalslice.AppendTransactionRequest,
	entryID string,
) (verticalslice.Transaction, error) {
	if _, err := verticalslice.GrossFor(request); err != nil {
		return verticalslice.Transaction{}, err
	}
	equivalentDuplicate, err := equivalentTransactionExists(ctx, tx, request, true)
	if err != nil {
		return verticalslice.Transaction{}, err
	}
	if equivalentDuplicate {
		return verticalslice.Transaction{}, verticalslice.ErrInvalidInput
	}
	assetID, err := ensureAsset(ctx, tx, request)
	if err != nil {
		return verticalslice.Transaction{}, err
	}
	if err := insertTransactionEntryStage371WithID(
		ctx,
		tx,
		command,
		request,
		entryID,
		"MANUAL",
		"",
		"",
	); err != nil {
		return verticalslice.Transaction{}, err
	}
	if request.TransactionType == "BUY" || request.TransactionType == "SELL" {
		if err := validatePositionHistoryTx(ctx, tx, request.PortfolioID, assetID); err != nil {
			return verticalslice.Transaction{}, err
		}
	}

	affectedDates, err := planAffectedSnapshotDates(ctx, tx, request.PortfolioID, []string{request.TradeDate})
	if err != nil {
		return verticalslice.Transaction{}, err
	}
	if err := rebuildSnapshotPlan(ctx, tx, request.PortfolioID, affectedDates, command.Now); err != nil {
		return verticalslice.Transaction{}, err
	}
	return getTransactionByEntryTx(ctx, tx, request.PortfolioID, entryID)
}

func (s *Store) AppendImportedTransactionsStage371(
	ctx context.Context,
	command verticalslice.CommandContext,
	request verticalslice.AppendImportBatchRequest,
) ([]verticalslice.Transaction, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollback(tx)

	if err := lockPortfolioTx(ctx, tx, command.SubjectID, request.PortfolioID); err != nil {
		return nil, err
	}
	commandID, duplicate, err := reserveCommand(ctx, tx, command, "POST")
	if err != nil {
		return nil, err
	}
	if duplicate {
		transactions, err := getImportedTransactionsByCommandTx(ctx, tx, commandID, request)
		if err != nil {
			return nil, err
		}
		return transactions, tx.Commit()
	}

	outcome, err := appendImportedTransactionsStage371Mutation(ctx, tx, command, commandID, request)
	if err != nil {
		return nil, err
	}
	if err := completeCommand(ctx, tx, commandID); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return outcome.Transactions, nil
}

func (s *Store) AppendImportedTransactionsWithReplayStage371(
	ctx context.Context,
	command verticalslice.CommandContext,
	request verticalslice.AppendImportBatchRequest,
	build verticalslice.ImportedTransactionsReplayBuilder,
) ([]verticalslice.Transaction, verticalslice.CommandReplayArtifact, error) {
	return s.AppendImportedTransactionsWithOutcomeReplayStage371(
		ctx,
		command,
		request,
		func(outcome verticalslice.ImportAppendOutcome) (verticalslice.CommandReplayArtifact, error) {
			return build(outcome.Transactions)
		},
	)
}

func (s *Store) AppendImportedTransactionsWithOutcomeStage371(
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
		return verticalslice.ImportAppendOutcome{}, ErrUnsupportedDuplicate
	}

	outcome, err := appendImportedTransactionsStage371Mutation(ctx, tx, command, commandID, request)
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

func (s *Store) AppendImportedTransactionsWithOutcomeReplayStage371(
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
	outcome, err := appendImportedTransactionsStage371Mutation(ctx, tx, command, reservation.ID, request)
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

func appendImportedTransactionsStage371Mutation(
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
		if err := insertTransactionEntryStage371WithID(
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

func insertTransactionEntryStage371WithID(
	ctx context.Context,
	tx *sql.Tx,
	command verticalslice.CommandContext,
	request verticalslice.AppendTransactionRequest,
	entryID string,
	sourceKind string,
	sourceFileHash string,
	sourceAccountLabel string,
) error {
	gross, err := verticalslice.GrossFor(request)
	if err != nil {
		return err
	}
	assetID, err := activeAssetID(ctx, tx, request)
	if err != nil {
		return err
	}
	ledgerSequence, err := nextLedgerSequenceTx(ctx, tx, request.PortfolioID)
	if err != nil {
		return err
	}

	transactionID := uuid.NewString()
	brokerOperationKey := ""
	sourceFingerprint := ""
	var sourceIdentityVersion any
	if request.ImportProvenance != nil {
		brokerOperationKey = request.ImportProvenance.BrokerOperationKey
		sourceFingerprint = request.ImportProvenance.SourceFingerprint
		sourceIdentityVersion = request.ImportProvenance.IdentityVersion
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO investment.transaction_entries (
			entry_id, transaction_id, portfolio_id, asset_id, revision, transaction_type,
			quantity, unit_price_amount, unit_price_currency,
			gross_amount, gross_currency, commission_amount, commission_currency,
			tax_amount, tax_currency, trade_date, settlement_date, note,
			source_kind, source_file_hash, source_account_label,
			source_broker_operation_key, source_fingerprint, source_identity_version,
			created_at, request_id, trace_id, ledger_sequence
		)
		VALUES (
			$1, $2, $3, $4, 1, $5,
			$6, $7, $8,
			$9, 'RUB', $10, 'RUB',
			$11, 'RUB', $12, $13, $14,
			$15, NULLIF($16, ''), $17,
			NULLIF($18, ''), NULLIF($19, ''), $20,
			$21, NULLIF($22, '')::uuid, NULLIF($23, ''), $24
		)
	`, entryID, transactionID, request.PortfolioID, assetID, request.TransactionType,
		decimalString(request.Quantity), moneyAmount(request.UnitPrice), moneyCurrency(request.UnitPrice),
		gross.Amount.String(), request.Commission.Amount.String(),
		request.Tax.Amount.String(), request.TradeDate, request.SettlementDate, request.Note,
		sourceKind, sourceFileHash, strings.TrimSpace(sourceAccountLabel),
		brokerOperationKey, sourceFingerprint, sourceIdentityVersion,
		command.Now, command.RequestID, command.TraceID, ledgerSequence)
	return err
}
