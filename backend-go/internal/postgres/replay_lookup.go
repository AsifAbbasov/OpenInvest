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

// LookupReplayArtifact is intentionally read-only. It is used only to recover an exact completed
// response before re-evaluating time-sensitive proofs that were required for the original write.
func (s *Store) LookupReplayArtifact(
	ctx context.Context,
	command verticalslice.CommandContext,
	method string,
) (verticalslice.CommandReplayArtifact, bool, error) {
	var existingHash string
	var terminalStatus sql.NullString
	var responseVersion sql.NullInt64
	var responseStatus sql.NullInt64
	var responseBody []byte
	var responseRequestID sql.NullString
	var responseTraceID sql.NullString
	var responseHash sql.NullString

	err := s.db.QueryRowContext(ctx, `
		SELECT
			request_hash,
			terminal_status,
			response_version,
			response_status,
			response_body,
			response_request_id::text,
			response_trace_id,
			response_hash
		FROM investment.command_deduplication
		WHERE principal_id = $1
			AND method = $2
			AND canonical_path = $3
			AND idempotency_key = $4
	`, command.SubjectID, method, command.RequestPath, command.IdempotencyKey).Scan(
		&existingHash,
		&terminalStatus,
		&responseVersion,
		&responseStatus,
		&responseBody,
		&responseRequestID,
		&responseTraceID,
		&responseHash,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return verticalslice.CommandReplayArtifact{}, false, nil
	}
	if err != nil {
		return verticalslice.CommandReplayArtifact{}, false, err
	}
	if existingHash != command.RequestHash {
		return verticalslice.CommandReplayArtifact{}, false, ErrIdempotencyConflict
	}
	if !terminalStatus.Valid {
		return verticalslice.CommandReplayArtifact{}, false, ErrIdempotencyInFlight
	}
	if terminalStatus.String != "success" || !responseVersion.Valid {
		return verticalslice.CommandReplayArtifact{}, false, ErrUnsupportedDuplicate
	}
	if responseVersion.Int64 != commandReplayVersion ||
		!responseStatus.Valid || responseStatus.Int64 < 100 || responseStatus.Int64 > 599 ||
		len(responseBody) == 0 || len(responseBody) > maxCommandReplayBodyBytes ||
		!responseRequestID.Valid || strings.TrimSpace(responseRequestID.String) == "" ||
		!responseTraceID.Valid || strings.TrimSpace(responseTraceID.String) == "" ||
		!responseHash.Valid {
		return verticalslice.CommandReplayArtifact{}, false, fmt.Errorf("corrupt idempotency replay artifact")
	}
	if _, err := uuid.Parse(responseRequestID.String); err != nil {
		return verticalslice.CommandReplayArtifact{}, false, fmt.Errorf("corrupt idempotency replay request id")
	}
	hash := sha256.Sum256(responseBody)
	if hex.EncodeToString(hash[:]) != responseHash.String {
		return verticalslice.CommandReplayArtifact{}, false, fmt.Errorf("corrupt idempotency replay body hash")
	}

	return verticalslice.CommandReplayArtifact{
		StatusCode: int(responseStatus.Int64),
		Body:       append([]byte(nil), responseBody...),
		RequestID:  responseRequestID.String,
		TraceID:    responseTraceID.String,
	}, true, nil
}
