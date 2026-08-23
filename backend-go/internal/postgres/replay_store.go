package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

const commandReplayVersion = 1
const maxCommandReplayBodyBytes = 256 * 1024

type replayReservation struct {
	ID        string
	Duplicate bool
	Artifact  verticalslice.CommandReplayArtifact
}

func (s *Store) CreatePortfolioWithReplay(
	ctx context.Context,
	command verticalslice.CommandContext,
	request verticalslice.CreatePortfolioRequest,
	build verticalslice.PortfolioReplayBuilder,
) (verticalslice.Portfolio, verticalslice.CommandReplayArtifact, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return verticalslice.Portfolio{}, verticalslice.CommandReplayArtifact{}, err
	}
	defer rollback(tx)

	// Resolve an already completed command before consulting or mutating current business state.
	// Exact replay must remain observable even if the resource has changed since the first success.
	reservation, err := reserveReplayCommand(ctx, tx, command, "POST")
	if err != nil {
		return verticalslice.Portfolio{}, verticalslice.CommandReplayArtifact{}, err
	}
	if reservation.Duplicate {
		if err := tx.Commit(); err != nil {
			return verticalslice.Portfolio{}, verticalslice.CommandReplayArtifact{}, err
		}
		return verticalslice.Portfolio{}, reservation.Artifact, nil
	}

	if err := ensureSubject(ctx, tx, command.SubjectID); err != nil {
		return verticalslice.Portfolio{}, verticalslice.CommandReplayArtifact{}, err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO investment.portfolios
			(id, subject_id, name, base_currency, portfolio_state, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'active', 1, $5, $5)
	`, reservation.ID, command.SubjectID, request.Name, request.BaseCurrency, command.Now)
	if err != nil {
		return verticalslice.Portfolio{}, verticalslice.CommandReplayArtifact{}, err
	}

	portfolio, err := getPortfolioTx(ctx, tx, command.SubjectID, reservation.ID)
	if err != nil {
		return verticalslice.Portfolio{}, verticalslice.CommandReplayArtifact{}, err
	}
	artifact, err := build(portfolio)
	if err != nil {
		return verticalslice.Portfolio{}, verticalslice.CommandReplayArtifact{}, err
	}
	if err := completeReplayCommand(ctx, tx, reservation.ID, artifact); err != nil {
		return verticalslice.Portfolio{}, verticalslice.CommandReplayArtifact{}, err
	}
	if err := tx.Commit(); err != nil {
		return verticalslice.Portfolio{}, verticalslice.CommandReplayArtifact{}, err
	}
	return portfolio, artifact, nil
}

func (s *Store) AppendTransactionWithReplay(
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

	// Resolve exact replay before locking or validating mutable portfolio state. For a brand-new
	// command the reservation is part of this same transaction and rolls back with any later error.
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

	if _, err := verticalslice.GrossFor(request); err != nil {
		return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err
	}
	equivalentDuplicate, err := equivalentTransactionExists(ctx, tx, request, true)
	if err != nil {
		return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err
	}
	if equivalentDuplicate {
		return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, verticalslice.ErrInvalidInput
	}
	if _, err := ensureAsset(ctx, tx, request); err != nil {
		return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err
	}
	if err := insertTransactionEntryWithID(ctx, tx, command, request, reservation.ID, "MANUAL", "", ""); err != nil {
		return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err
	}
	if err := rebuildAffectedSnapshots(ctx, tx, request.PortfolioID, request.TradeDate, command.Now); err != nil {
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

func (s *Store) AppendImportedTransactionsWithReplay(
	ctx context.Context,
	command verticalslice.CommandContext,
	request verticalslice.AppendImportBatchRequest,
	build verticalslice.ImportedTransactionsReplayBuilder,
) ([]verticalslice.Transaction, verticalslice.CommandReplayArtifact, error) {
	return s.AppendImportedTransactionsWithOutcomeReplay(
		ctx,
		command,
		request,
		func(outcome verticalslice.ImportAppendOutcome) (verticalslice.CommandReplayArtifact, error) {
			return build(outcome.Transactions)
		},
	)
}

func reserveReplayCommand(ctx context.Context, tx *sql.Tx, command verticalslice.CommandContext, method string) (replayReservation, error) {
	commandID := uuid.NewString()
	var existingID string
	var existingHash string
	var terminalStatus sql.NullString
	var responseVersion sql.NullInt64
	var responseStatus sql.NullInt64
	var responseBody []byte
	var responseRequestID sql.NullString
	var responseTraceID sql.NullString
	var responseHash sql.NullString

	err := tx.QueryRowContext(ctx, `
		INSERT INTO investment.command_deduplication
			(id, principal_id, method, canonical_path, idempotency_key, request_hash, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7::timestamptz, $7::timestamptz + interval '24 hours')
		ON CONFLICT (principal_id, method, canonical_path, idempotency_key)
		DO UPDATE SET idempotency_key = investment.command_deduplication.idempotency_key
		RETURNING
			id,
			request_hash,
			terminal_status,
			response_version,
			response_status,
			response_body,
			response_request_id::text,
			response_trace_id,
			response_hash
	`, commandID, command.SubjectID, method, command.RequestPath, command.IdempotencyKey, command.RequestHash, command.Now).
		Scan(
			&existingID,
			&existingHash,
			&terminalStatus,
			&responseVersion,
			&responseStatus,
			&responseBody,
			&responseRequestID,
			&responseTraceID,
			&responseHash,
		)
	if err != nil {
		return replayReservation{}, err
	}
	if existingID == commandID {
		return replayReservation{ID: commandID}, nil
	}
	if existingHash != command.RequestHash {
		return replayReservation{}, ErrIdempotencyConflict
	}
	if !terminalStatus.Valid {
		return replayReservation{}, ErrIdempotencyInFlight
	}
	if terminalStatus.String != "success" || !responseVersion.Valid {
		return replayReservation{}, ErrUnsupportedDuplicate
	}
	if responseVersion.Int64 != commandReplayVersion ||
		!responseStatus.Valid || responseStatus.Int64 < 100 || responseStatus.Int64 > 599 ||
		len(responseBody) == 0 || len(responseBody) > maxCommandReplayBodyBytes ||
		!responseRequestID.Valid || strings.TrimSpace(responseRequestID.String) == "" ||
		!responseTraceID.Valid || strings.TrimSpace(responseTraceID.String) == "" ||
		!responseHash.Valid {
		return replayReservation{}, fmt.Errorf("corrupt idempotency replay artifact for command %s", existingID)
	}
	if _, err := uuid.Parse(responseRequestID.String); err != nil {
		return replayReservation{}, fmt.Errorf("corrupt idempotency replay request id for command %s", existingID)
	}
	hash := sha256.Sum256(responseBody)
	if hex.EncodeToString(hash[:]) != responseHash.String {
		return replayReservation{}, fmt.Errorf("corrupt idempotency replay body hash for command %s", existingID)
	}

	return replayReservation{
		ID:        existingID,
		Duplicate: true,
		Artifact: verticalslice.CommandReplayArtifact{
			StatusCode: int(responseStatus.Int64),
			Body:       append([]byte(nil), responseBody...),
			RequestID:  responseRequestID.String,
			TraceID:    responseTraceID.String,
		},
	}, nil
}

func completeReplayCommand(ctx context.Context, tx *sql.Tx, commandID string, artifact verticalslice.CommandReplayArtifact) error {
	if artifact.StatusCode < 100 || artifact.StatusCode > 599 {
		return fmt.Errorf("invalid replay response status %d", artifact.StatusCode)
	}
	if len(artifact.Body) == 0 || len(artifact.Body) > maxCommandReplayBodyBytes {
		return fmt.Errorf("invalid replay response body size %d", len(artifact.Body))
	}
	if _, err := uuid.Parse(strings.TrimSpace(artifact.RequestID)); err != nil {
		return fmt.Errorf("invalid replay request id")
	}
	traceID := strings.TrimSpace(artifact.TraceID)
	if traceID == "" || len(traceID) > 128 {
		return fmt.Errorf("invalid replay trace id")
	}

	hash := sha256.Sum256(artifact.Body)
	result, err := tx.ExecContext(ctx, `
		UPDATE investment.command_deduplication
		SET
			terminal_status = 'success',
			response_version = $2,
			response_status = $3,
			response_body = $4,
			response_request_id = $5::uuid,
			response_trace_id = $6,
			response_hash = $7
		WHERE id = $1
			AND terminal_status IS NULL
	`, commandID, commandReplayVersion, artifact.StatusCode, artifact.Body, artifact.RequestID, traceID, hex.EncodeToString(hash[:]))
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return errors.New("idempotency command was not completed exactly once")
	}
	return nil
}
