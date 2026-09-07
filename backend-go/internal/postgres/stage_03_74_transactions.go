package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

type stage374LogicalEntry struct {
	EntryID         string
	TransactionID   string
	Revision        int
	TransactionType string
	AssetID         sql.NullString
	Quantity        sql.NullString
	UnitPrice       sql.NullString
	GrossAmount     string
	Commission      string
	Tax             string
	TradeDate       string
	SettlementDate  sql.NullString
	Note            sql.NullString
}

func (s *Store) CorrectTransactionWithReplayStage374(
	ctx context.Context,
	command verticalslice.CommandContext,
	request verticalslice.CorrectTransactionRequest,
	build verticalslice.TransactionReplayBuilder,
) (verticalslice.Transaction, verticalslice.CommandReplayArtifact, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err
	}
	defer rollback(tx)

	reservation, err := reserveReplayCommand(ctx, tx, command, "PATCH")
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
	if err := validateEffectiveLedgerShapeTx(ctx, tx, request.PortfolioID); err != nil {
		return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err
	}
	current, err := latestLogicalEntryTx(ctx, tx, request.PortfolioID, request.TransactionID)
	if err != nil {
		return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err
	}
	reversed, err := logicalTransactionReversedTx(ctx, tx, request.PortfolioID, request.TransactionID)
	if err != nil {
		return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err
	}
	if reversed || current.Revision != request.ExpectedRevision {
		return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, verticalslice.ErrTransactionConflict
	}
	gross, err := verticalslice.GrossFor(request.Corrected)
	if err != nil {
		return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err
	}
	assetID, err := ensureAsset(ctx, tx, request.Corrected)
	if err != nil {
		return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err
	}
	ledgerSequence, err := nextLedgerSequenceTx(ctx, tx, request.PortfolioID)
	if err != nil {
		return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err
	}

	_, err = tx.ExecContext(ctx, `
        INSERT INTO investment.transaction_entries (
            entry_id, transaction_id, portfolio_id, asset_id, revision, transaction_type,
            quantity, unit_price_amount, unit_price_currency,
            gross_amount, gross_currency, commission_amount, commission_currency,
            tax_amount, tax_currency, trade_date, settlement_date, note,
            correction_reason, prior_entry_id, reverses_transaction_id,
            source_kind, source_file_hash, source_account_label,
            source_broker_operation_key, source_fingerprint, source_identity_version,
            created_at, request_id, trace_id, ledger_sequence
        ) VALUES (
            $1, $2, $3, $4, $5, $6,
            $7, $8, $9,
            $10, 'RUB', $11, 'RUB',
            $12, 'RUB', $13, $14, $15,
            $16, $17, NULL,
            'MANUAL', NULL, '', NULL, NULL, NULL,
            $18, NULLIF($19, '')::uuid, NULLIF($20, ''), $21
        )
    `, reservation.ID, request.TransactionID, request.PortfolioID, assetID, current.Revision+1, request.Corrected.TransactionType,
		decimalString(request.Corrected.Quantity), moneyAmount(request.Corrected.UnitPrice), moneyCurrency(request.Corrected.UnitPrice),
		gross.Amount.String(), request.Corrected.Commission.Amount.String(), request.Corrected.Tax.Amount.String(),
		request.Corrected.TradeDate, request.Corrected.SettlementDate, request.Corrected.Note,
		request.Reason, current.EntryID, command.Now, command.RequestID, command.TraceID, ledgerSequence)
	if err != nil {
		return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err
	}

	if _, _, err := rebuildPortfolioPositionsTx(ctx, tx, request.PortfolioID, ""); err != nil {
		return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err
	}
	affectedDates, err := planAffectedSnapshotDates(ctx, tx, request.PortfolioID, []string{current.TradeDate, request.Corrected.TradeDate})
	if err != nil {
		return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err
	}
	if err := rebuildSnapshotPlan(ctx, tx, request.PortfolioID, affectedDates, command.Now); err != nil {
		return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err
	}
	transaction, err := getTransactionByEntryTx(ctx, tx, request.PortfolioID, reservation.ID)
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

func (s *Store) ReverseTransactionWithReplayStage374(
	ctx context.Context,
	command verticalslice.CommandContext,
	request verticalslice.ReverseTransactionRequest,
	build verticalslice.TransactionReversalReplayBuilder,
) (verticalslice.TransactionReversal, verticalslice.CommandReplayArtifact, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return verticalslice.TransactionReversal{}, verticalslice.CommandReplayArtifact{}, err
	}
	defer rollback(tx)

	reservation, err := reserveReplayCommand(ctx, tx, command, "DELETE")
	if err != nil {
		return verticalslice.TransactionReversal{}, verticalslice.CommandReplayArtifact{}, err
	}
	if reservation.Duplicate {
		if err := tx.Commit(); err != nil {
			return verticalslice.TransactionReversal{}, verticalslice.CommandReplayArtifact{}, err
		}
		return verticalslice.TransactionReversal{}, reservation.Artifact, nil
	}
	if err := lockPortfolioTx(ctx, tx, command.SubjectID, request.PortfolioID); err != nil {
		return verticalslice.TransactionReversal{}, verticalslice.CommandReplayArtifact{}, err
	}
	if err := validateEffectiveLedgerShapeTx(ctx, tx, request.PortfolioID); err != nil {
		return verticalslice.TransactionReversal{}, verticalslice.CommandReplayArtifact{}, err
	}
	current, err := latestLogicalEntryTx(ctx, tx, request.PortfolioID, request.TransactionID)
	if err != nil {
		return verticalslice.TransactionReversal{}, verticalslice.CommandReplayArtifact{}, err
	}
	reversed, err := logicalTransactionReversedTx(ctx, tx, request.PortfolioID, request.TransactionID)
	if err != nil {
		return verticalslice.TransactionReversal{}, verticalslice.CommandReplayArtifact{}, err
	}
	if reversed || current.Revision != request.ExpectedRevision {
		return verticalslice.TransactionReversal{}, verticalslice.CommandReplayArtifact{}, verticalslice.ErrTransactionConflict
	}
	ledgerSequence, err := nextLedgerSequenceTx(ctx, tx, request.PortfolioID)
	if err != nil {
		return verticalslice.TransactionReversal{}, verticalslice.CommandReplayArtifact{}, err
	}
	reversalTransactionID := uuid.NewString()
	_, err = tx.ExecContext(ctx, `
        INSERT INTO investment.transaction_entries (
            entry_id, transaction_id, portfolio_id, asset_id, revision, transaction_type,
            quantity, unit_price_amount, unit_price_currency,
            gross_amount, gross_currency, commission_amount, commission_currency,
            tax_amount, tax_currency, trade_date, settlement_date, note,
            correction_reason, prior_entry_id, reverses_transaction_id,
            source_kind, source_file_hash, source_account_label,
            source_broker_operation_key, source_fingerprint, source_identity_version,
            created_at, request_id, trace_id, ledger_sequence
        ) VALUES (
            $1, $2, $3, $4, 1, $5,
            $6::numeric, $7::numeric, CASE WHEN $7::numeric IS NULL THEN NULL ELSE 'RUB' END,
            $8::numeric, 'RUB', $9::numeric, 'RUB', $10::numeric, 'RUB',
            $11::date, $12::date, $13,
            $14, NULL, $15,
            'MANUAL', NULL, '', NULL, NULL, NULL,
            $16, NULLIF($17, '')::uuid, NULLIF($18, ''), $19
        )
    `, reservation.ID, reversalTransactionID, request.PortfolioID, nullStringValue(current.AssetID), current.TransactionType,
		nullStringValue(current.Quantity), nullStringValue(current.UnitPrice), current.GrossAmount, current.Commission, current.Tax,
		request.EffectiveDate, nullStringValue(current.SettlementDate), nullStringValue(current.Note), request.Reason, request.TransactionID,
		command.Now, command.RequestID, command.TraceID, ledgerSequence)
	if err != nil {
		return verticalslice.TransactionReversal{}, verticalslice.CommandReplayArtifact{}, err
	}

	if _, _, err := rebuildPortfolioPositionsTx(ctx, tx, request.PortfolioID, ""); err != nil {
		return verticalslice.TransactionReversal{}, verticalslice.CommandReplayArtifact{}, err
	}
	affectedDates, err := planAffectedSnapshotDates(ctx, tx, request.PortfolioID, []string{request.EffectiveDate})
	if err != nil {
		return verticalslice.TransactionReversal{}, verticalslice.CommandReplayArtifact{}, err
	}
	if err := rebuildSnapshotPlan(ctx, tx, request.PortfolioID, affectedDates, command.Now); err != nil {
		return verticalslice.TransactionReversal{}, verticalslice.CommandReplayArtifact{}, err
	}
	result := verticalslice.TransactionReversal{
		TransactionID:         request.TransactionID,
		ReversalTransactionID: reversalTransactionID,
		Status:                "REVERSED",
		EffectiveDate:         request.EffectiveDate,
	}
	artifact, err := build(result)
	if err != nil {
		return verticalslice.TransactionReversal{}, verticalslice.CommandReplayArtifact{}, err
	}
	if err := completeReplayCommand(ctx, tx, reservation.ID, artifact); err != nil {
		return verticalslice.TransactionReversal{}, verticalslice.CommandReplayArtifact{}, err
	}
	if err := tx.Commit(); err != nil {
		return verticalslice.TransactionReversal{}, verticalslice.CommandReplayArtifact{}, err
	}
	return result, artifact, nil
}

func latestLogicalEntryTx(ctx context.Context, tx *sql.Tx, portfolioID string, transactionID string) (stage374LogicalEntry, error) {
	var entry stage374LogicalEntry
	err := tx.QueryRowContext(ctx, `
        SELECT entry_id::text, transaction_id::text, revision, transaction_type,
               asset_id::text, quantity::text, unit_price_amount::text,
               gross_amount::text, commission_amount::text, tax_amount::text,
               trade_date::text, settlement_date::text, note
        FROM investment.transaction_entries
        WHERE portfolio_id = $1
          AND transaction_id = $2
          AND reverses_transaction_id IS NULL
        ORDER BY revision DESC
        LIMIT 1
    `, portfolioID, transactionID).Scan(
		&entry.EntryID, &entry.TransactionID, &entry.Revision, &entry.TransactionType,
		&entry.AssetID, &entry.Quantity, &entry.UnitPrice,
		&entry.GrossAmount, &entry.Commission, &entry.Tax,
		&entry.TradeDate, &entry.SettlementDate, &entry.Note,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return stage374LogicalEntry{}, ErrNotFound
	}
	return entry, err
}

func logicalTransactionReversedTx(ctx context.Context, tx *sql.Tx, portfolioID string, transactionID string) (bool, error) {
	var count int
	if err := tx.QueryRowContext(ctx, `
        SELECT COUNT(*)
        FROM investment.transaction_entries
        WHERE portfolio_id = $1 AND reverses_transaction_id = $2
    `, portfolioID, transactionID).Scan(&count); err != nil {
		return false, err
	}
	if count > 1 {
		return false, ErrUnsupportedPositionLedger
	}
	return count == 1, nil
}

func nullStringValue(value sql.NullString) any {
	if !value.Valid {
		return nil
	}
	return strings.TrimSpace(value.String)
}
