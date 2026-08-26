package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

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
	if err := lockReplayCommandScope(ctx, tx, command, method); err != nil {
		return replayReservation{}, err
	}

	for {
		var existingID string
		var existingHash string
		var terminalStatus sql.NullString
		var responseVersion sql.NullInt64
		var responseStatus sql.NullInt64
		var responseBody []byte
		var responseRequestID sql.NullString
		var responseTraceID sql.NullString
		var responseHash sql.NullString
		var expiresAt time.Time

		err := tx.QueryRowContext(ctx, `
			SELECT
				id,
				request_hash,
				terminal_status,
				response_version,
				response_status,
				response_body,
				response_request_id::text,
				response_trace_id,
				response_hash,
				expires_at
			FROM investment.command_deduplication
			WHERE principal_id = $1
				AND method = $2
				AND canonical_path = $3
				AND idempotency_key = $4
			FOR UPDATE
		`, command.SubjectID, method, command.RequestPath, command.IdempotencyKey).Scan(
			&existingID,
			&existingHash,
			&terminalStatus,
			&responseVersion,
			&responseStatus,
			&responseBody,
			&responseRequestID,
			&responseTraceID,
			&responseHash,
			&expiresAt,
		)
		if errors.Is(err, sql.ErrNoRows) {
			// The INSERT can still block on a mixed-version/non-cooperating writer that does not
			// honor the Stage 3.38 advisory lock. Use a provisional non-null timestamp for the
			// invisible in-transaction row, then anchor persisted admission time only after the
			// potentially blocking INSERT has returned as the winner.
			provisionalTime, clockErr := databaseWallClock(ctx, tx)
			if clockErr != nil {
				return replayReservation{}, clockErr
			}
			commandID := uuid.NewString()
			result, insertErr := tx.ExecContext(ctx, `
				INSERT INTO investment.command_deduplication (
					id, principal_id, method, canonical_path, idempotency_key, request_hash,
					created_at, expires_at
				)
				VALUES (
					$1, $2, $3, $4, $5, $6,
					$7::timestamptz,
					$7::timestamptz + interval '24 hours'
				)
				ON CONFLICT (principal_id, method, canonical_path, idempotency_key) DO NOTHING
			`, commandID, command.SubjectID, method, command.RequestPath, command.IdempotencyKey, command.RequestHash, provisionalTime)
			if insertErr != nil {
				return replayReservation{}, insertErr
			}
			rowsAffected, rowsErr := result.RowsAffected()
			if rowsErr != nil {
				return replayReservation{}, rowsErr
			}
			if rowsAffected == 1 {
				admissionTime, admissionErr := databaseWallClock(ctx, tx)
				if admissionErr != nil {
					return replayReservation{}, admissionErr
				}
				updateResult, updateErr := tx.ExecContext(ctx, `
					UPDATE investment.command_deduplication
					SET
						created_at = $2::timestamptz,
						expires_at = $2::timestamptz + interval '24 hours'
					WHERE id = $1
				`, commandID, admissionTime)
				if updateErr != nil {
					return replayReservation{}, updateErr
				}
				updatedRows, updatedRowsErr := updateResult.RowsAffected()
				if updatedRowsErr != nil {
					return replayReservation{}, updatedRowsErr
				}
				if updatedRows != 1 {
					return replayReservation{}, errors.New("fresh idempotency admission timestamp was not finalized exactly once")
				}
				if _, cleanupErr := cleanupExpiredReplayCommands(ctx, tx); cleanupErr != nil {
					return replayReservation{}, cleanupErr
				}
				return replayReservation{ID: commandID}, nil
			}

			// A mixed-version or otherwise non-cooperating writer won the unique-index race.
			// Loop and acquire that now-visible exact row before deciding replay/conflict/expiry.
			continue
		}
		if err != nil {
			return replayReservation{}, err
		}

		decisionTime, err := databaseWallClock(ctx, tx)
		if err != nil {
			return replayReservation{}, err
		}
		if expiredAt(expiresAt, decisionTime) {
			commandID := uuid.NewString()
			result, updateErr := tx.ExecContext(ctx, `
				UPDATE investment.command_deduplication
				SET
					id = $2,
					request_hash = $3,
					terminal_status = NULL,
					response_hash = NULL,
					response_version = NULL,
					response_status = NULL,
					response_body = NULL,
					response_request_id = NULL,
					response_trace_id = NULL,
					created_at = $4::timestamptz,
					expires_at = $4::timestamptz + interval '24 hours'
				WHERE id = $1
			`, existingID, commandID, command.RequestHash, decisionTime)
			if updateErr != nil {
				return replayReservation{}, updateErr
			}
			rowsAffected, rowsErr := result.RowsAffected()
			if rowsErr != nil {
				return replayReservation{}, rowsErr
			}
			if rowsAffected != 1 {
				return replayReservation{}, errors.New("expired idempotency generation was not reclaimed exactly once")
			}
			if _, cleanupErr := cleanupExpiredReplayCommands(ctx, tx); cleanupErr != nil {
				return replayReservation{}, cleanupErr
			}
			return replayReservation{ID: commandID}, nil
		}

		var reservation replayReservation
		var reservationErr error
		switch {
		case existingHash != command.RequestHash:
			reservationErr = ErrIdempotencyConflict
		case !terminalStatus.Valid:
			reservationErr = ErrIdempotencyInFlight
		case terminalStatus.String != "success" || !responseVersion.Valid:
			reservationErr = ErrUnsupportedDuplicate
		case responseVersion.Int64 != commandReplayVersion ||
			!responseStatus.Valid || responseStatus.Int64 < 100 || responseStatus.Int64 > 599 ||
			len(responseBody) == 0 || len(responseBody) > maxCommandReplayBodyBytes ||
			!responseRequestID.Valid || strings.TrimSpace(responseRequestID.String) == "" ||
			!responseTraceID.Valid || strings.TrimSpace(responseTraceID.String) == "" ||
			!responseHash.Valid:
			reservationErr = fmt.Errorf("corrupt idempotency replay artifact for command %s", existingID)
		default:
			if _, parseErr := uuid.Parse(responseRequestID.String); parseErr != nil {
				reservationErr = fmt.Errorf("corrupt idempotency replay request id for command %s", existingID)
				break
			}
			hash := sha256.Sum256(responseBody)
			if hex.EncodeToString(hash[:]) != responseHash.String {
				reservationErr = fmt.Errorf("corrupt idempotency replay body hash for command %s", existingID)
				break
			}
			reservation = replayReservation{
				ID:        existingID,
				Duplicate: true,
				Artifact: verticalslice.CommandReplayArtifact{
					StatusCode: int(responseStatus.Int64),
					Body:       append([]byte(nil), responseBody...),
					RequestID:  responseRequestID.String,
					TraceID:    responseTraceID.String,
				},
			}
		}

		if _, cleanupErr := cleanupExpiredReplayCommands(ctx, tx); cleanupErr != nil {
			return replayReservation{}, cleanupErr
		}
		return reservation, reservationErr
	}
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
