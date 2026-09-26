package importflow

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/openinvest/openinvest/backend-go/internal/importer"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

type ReplayAppender interface {
	AppendImportedTransactionsWithOutcomeReplay(
		ctx context.Context,
		requestContext verticalslice.RequestContext,
		subjectID string,
		idempotencyKey string,
		requestPath string,
		request verticalslice.AppendImportBatchRequest,
		build verticalslice.ImportedTransactionsOutcomeReplayBuilder,
	) ([]verticalslice.Transaction, verticalslice.CommandReplayArtifact, error)
}

type ResultReplayBuilder func(Result) (verticalslice.CommandReplayArtifact, error)

// PreparedReplay is the immutable current-parser import preparation shared by the replay lookup
// and write continuation. It deliberately leaves signed-proof validation to the HTTP boundary.
type PreparedReplay struct {
	Review        importer.Review
	AppendRequest verticalslice.AppendImportBatchRequest
	decisionCount int
	appendCount   int
}

func PrepareReplay(request Request) (PreparedReplay, error) {
	if strings.TrimSpace(request.SubjectID) == "" {
		return PreparedReplay{}, fmt.Errorf("%w: subjectId is required", ErrInvalidFlowInput)
	}
	if strings.TrimSpace(request.PortfolioID) == "" {
		return PreparedReplay{}, fmt.Errorf("%w: portfolioId is required", ErrInvalidFlowInput)
	}
	if request.Reader == nil {
		return PreparedReplay{}, fmt.Errorf("%w: reader is required", ErrInvalidFlowInput)
	}

	payload, err := io.ReadAll(io.LimitReader(request.Reader, maxImportPayloadBytes+1))
	if err != nil {
		return PreparedReplay{}, fmt.Errorf("%w: read import payload", ErrInvalidFlowInput)
	}
	if int64(len(payload)) > maxImportPayloadBytes {
		return PreparedReplay{}, fmt.Errorf("%w: import payload exceeds %d bytes", ErrInvalidFlowInput, maxImportPayloadBytes)
	}
	sourceFileHash := strings.TrimSpace(request.SourceFileHash)
	if sourceFileHash == "" {
		return PreparedReplay{}, fmt.Errorf("%w: sourceFileHash is required", ErrInvalidFlowInput)
	}
	hash := sha256.Sum256(payload)
	actualFileHash := hex.EncodeToString(hash[:])
	if !strings.EqualFold(sourceFileHash, actualFileHash) {
		return PreparedReplay{}, fmt.Errorf("%w: sourceFileHash does not match import payload", importer.ErrUnsafeAppend)
	}

	review, err := importer.ReviewCSV(importer.ReviewRequest{
		SubjectID:          request.SubjectID,
		PortfolioID:        request.PortfolioID,
		SourceKind:         request.SourceKind,
		SourceAccountLabel: request.SourceAccountLabel,
		FileHash:           sourceFileHash,
		Existing:           request.Existing,
		Reader:             bytes.NewReader(payload),
	})
	if err != nil {
		return PreparedReplay{}, err
	}

	appendRequests, err := importer.BuildAppendRequests(review, request.Decisions)
	if err != nil {
		return PreparedReplay{}, err
	}
	if len(appendRequests) == 0 {
		return PreparedReplay{}, ErrNoApprovedRows
	}

	return PreparedReplay{
		Review: review,
		AppendRequest: verticalslice.AppendImportBatchRequest{
			PortfolioID:        request.PortfolioID,
			Transactions:       appendRequests,
			SourceKind:         review.SourceKind,
			SourceAccountLabel: review.SourceAccountLabel,
			SourceFileHash:     review.FileHash,
			Decisions:          importDecisions(request.Decisions),
		},
		decisionCount: len(request.Decisions),
		appendCount:   len(appendRequests),
	}, nil
}

func (prepared PreparedReplay) result(outcome verticalslice.ImportAppendOutcome) Result {
	return Result{
		ParsedRowCount:         prepared.Review.Summary.TotalRows,
		AcceptedRowCount:       prepared.appendCount,
		NonAppendedRowCount:    prepared.Review.Summary.TotalRows - prepared.appendCount,
		AppendedTransactionIDs: transactionIDs(outcome.Transactions),
		SnapshotDatesRebuilt:   append([]string(nil), outcome.SnapshotDatesRebuilt...),
		AuditActionCode:        "IMPORT_APPEND_BATCH",
		NonSensitiveWarnings:   nonSensitiveWarnings(prepared.Review, prepared.decisionCount, prepared.appendCount),
	}
}

// AppendPreparedWithReplay continues a previously prepared current-parser command through the
// existing atomic write and exact-response persistence boundary without parsing it again.
func AppendPreparedWithReplay(
	ctx context.Context,
	appender ReplayAppender,
	requestContext verticalslice.RequestContext,
	subjectID string,
	idempotencyKey string,
	requestPath string,
	prepared PreparedReplay,
	build ResultReplayBuilder,
) (Result, verticalslice.CommandReplayArtifact, error) {
	if appender == nil {
		return Result{}, verticalslice.CommandReplayArtifact{}, fmt.Errorf("%w: appender is required", ErrInvalidFlowInput)
	}
	if build == nil {
		return Result{}, verticalslice.CommandReplayArtifact{}, fmt.Errorf("%w: replay builder is required", ErrInvalidFlowInput)
	}
	if strings.TrimSpace(subjectID) == "" {
		return Result{}, verticalslice.CommandReplayArtifact{}, fmt.Errorf("%w: subjectId is required", ErrInvalidFlowInput)
	}

	var result Result
	transactions, artifact, err := appender.AppendImportedTransactionsWithOutcomeReplay(
		ctx,
		requestContext,
		subjectID,
		idempotencyKey,
		requestPath,
		prepared.AppendRequest,
		func(outcome verticalslice.ImportAppendOutcome) (verticalslice.CommandReplayArtifact, error) {
			result = prepared.result(outcome)
			return build(result)
		},
	)
	if err != nil {
		return Result{}, verticalslice.CommandReplayArtifact{}, err
	}
	if len(artifact.Body) == 0 {
		return Result{}, verticalslice.CommandReplayArtifact{}, fmt.Errorf("%w: replay artifact is empty", ErrInvalidFlowInput)
	}
	if len(transactions) == 0 {
		// Duplicate replay: the original result is already encoded in artifact.Body and must not be
		// reconstructed from mutable database state.
		return Result{}, artifact, nil
	}
	return result, artifact, nil
}

func ReviewAndAppendWithReplay(
	ctx context.Context,
	appender ReplayAppender,
	request Request,
	build ResultReplayBuilder,
) (Result, verticalslice.CommandReplayArtifact, error) {
	prepared, err := PrepareReplay(request)
	if err != nil {
		return Result{}, verticalslice.CommandReplayArtifact{}, err
	}
	return AppendPreparedWithReplay(
		ctx, appender, request.RequestContext, request.SubjectID, request.IdempotencyKey, request.RequestPath, prepared, build,
	)
}
