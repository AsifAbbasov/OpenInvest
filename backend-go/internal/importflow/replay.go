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
	AppendImportedTransactionsWithReplay(
		ctx context.Context,
		requestContext verticalslice.RequestContext,
		subjectID string,
		idempotencyKey string,
		requestPath string,
		request verticalslice.AppendImportBatchRequest,
		build verticalslice.ImportedTransactionsReplayBuilder,
	) ([]verticalslice.Transaction, verticalslice.CommandReplayArtifact, error)
}

type ResultReplayBuilder func(Result) (verticalslice.CommandReplayArtifact, error)

func ReviewAndAppendWithReplay(
	ctx context.Context,
	appender ReplayAppender,
	request Request,
	build ResultReplayBuilder,
) (Result, verticalslice.CommandReplayArtifact, error) {
	if appender == nil {
		return Result{}, verticalslice.CommandReplayArtifact{}, fmt.Errorf("%w: appender is required", ErrInvalidFlowInput)
	}
	if build == nil {
		return Result{}, verticalslice.CommandReplayArtifact{}, fmt.Errorf("%w: replay builder is required", ErrInvalidFlowInput)
	}
	if strings.TrimSpace(request.SubjectID) == "" {
		return Result{}, verticalslice.CommandReplayArtifact{}, fmt.Errorf("%w: subjectId is required", ErrInvalidFlowInput)
	}
	if strings.TrimSpace(request.PortfolioID) == "" {
		return Result{}, verticalslice.CommandReplayArtifact{}, fmt.Errorf("%w: portfolioId is required", ErrInvalidFlowInput)
	}
	if request.Reader == nil {
		return Result{}, verticalslice.CommandReplayArtifact{}, fmt.Errorf("%w: reader is required", ErrInvalidFlowInput)
	}

	payload, err := io.ReadAll(io.LimitReader(request.Reader, maxImportPayloadBytes+1))
	if err != nil {
		return Result{}, verticalslice.CommandReplayArtifact{}, fmt.Errorf("%w: read import payload", ErrInvalidFlowInput)
	}
	if int64(len(payload)) > maxImportPayloadBytes {
		return Result{}, verticalslice.CommandReplayArtifact{}, fmt.Errorf("%w: import payload exceeds %d bytes", ErrInvalidFlowInput, maxImportPayloadBytes)
	}
	sourceFileHash := strings.TrimSpace(request.SourceFileHash)
	if sourceFileHash == "" {
		return Result{}, verticalslice.CommandReplayArtifact{}, fmt.Errorf("%w: sourceFileHash is required", ErrInvalidFlowInput)
	}
	hash := sha256.Sum256(payload)
	actualFileHash := hex.EncodeToString(hash[:])
	if !strings.EqualFold(sourceFileHash, actualFileHash) {
		return Result{}, verticalslice.CommandReplayArtifact{}, fmt.Errorf("%w: sourceFileHash does not match import payload", importer.ErrUnsafeAppend)
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
		return Result{}, verticalslice.CommandReplayArtifact{}, err
	}

	appendRequests, err := importer.BuildAppendRequests(review, request.Decisions)
	if err != nil {
		return Result{}, verticalslice.CommandReplayArtifact{}, err
	}
	if len(appendRequests) == 0 {
		return Result{}, verticalslice.CommandReplayArtifact{}, ErrNoApprovedRows
	}

	var result Result
	transactions, artifact, err := appender.AppendImportedTransactionsWithReplay(
		ctx,
		request.RequestContext,
		request.SubjectID,
		request.IdempotencyKey,
		request.RequestPath,
		verticalslice.AppendImportBatchRequest{
			PortfolioID:        request.PortfolioID,
			Transactions:       appendRequests,
			SourceKind:         review.SourceKind,
			SourceAccountLabel: review.SourceAccountLabel,
			SourceFileHash:     review.FileHash,
			Decisions:          importDecisions(request.Decisions),
		},
		func(transactions []verticalslice.Transaction) (verticalslice.CommandReplayArtifact, error) {
			result = Result{
				ParsedRowCount:         review.Summary.TotalRows,
				AcceptedRowCount:       len(appendRequests),
				NonAppendedRowCount:    review.Summary.TotalRows - len(appendRequests),
				AppendedTransactionIDs: transactionIDs(transactions),
				SnapshotDatesRebuilt:   snapshotDates(appendRequests),
				AuditActionCode:        "IMPORT_APPEND_BATCH",
				NonSensitiveWarnings:   nonSensitiveWarnings(review, len(request.Decisions), len(appendRequests)),
			}
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
