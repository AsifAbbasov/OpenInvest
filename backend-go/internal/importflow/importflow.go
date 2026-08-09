package importflow

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/openinvest/openinvest/backend-go/internal/importer"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

var (
	ErrInvalidFlowInput = errors.New("invalid import flow input")
	ErrNoApprovedRows   = errors.New("no approved import rows")
)

const maxImportPayloadBytes int64 = 2 * 1024 * 1024

type Appender interface {
	AppendImportedTransactions(ctx context.Context, requestContext verticalslice.RequestContext, subjectID string, idempotencyKey string, requestPath string, request verticalslice.AppendImportBatchRequest) ([]verticalslice.Transaction, error)
}

type Request struct {
	RequestContext     verticalslice.RequestContext
	SubjectID          string
	PortfolioID        string
	IdempotencyKey     string
	RequestPath        string
	SourceKind         string
	SourceAccountLabel string
	SourceFileHash     string
	Existing           []verticalslice.Transaction
	Reader             io.Reader
	Decisions          []importer.Decision
}

type Result struct {
	ParsedRowCount         int
	AcceptedRowCount       int
	NonAppendedRowCount    int
	AppendedTransactionIDs []string
	SnapshotDatesRebuilt   []string
	AuditActionCode        string
	NonSensitiveWarnings   []string
}

func ReviewAndAppend(ctx context.Context, appender Appender, request Request) (Result, error) {
	if appender == nil {
		return Result{}, fmt.Errorf("%w: appender is required", ErrInvalidFlowInput)
	}
	if strings.TrimSpace(request.SubjectID) == "" {
		return Result{}, fmt.Errorf("%w: subjectId is required", ErrInvalidFlowInput)
	}
	if strings.TrimSpace(request.PortfolioID) == "" {
		return Result{}, fmt.Errorf("%w: portfolioId is required", ErrInvalidFlowInput)
	}
	if request.Reader == nil {
		return Result{}, fmt.Errorf("%w: reader is required", ErrInvalidFlowInput)
	}

	payload, err := io.ReadAll(io.LimitReader(request.Reader, maxImportPayloadBytes+1))
	if err != nil {
		return Result{}, fmt.Errorf("%w: read import payload", ErrInvalidFlowInput)
	}
	if int64(len(payload)) > maxImportPayloadBytes {
		return Result{}, fmt.Errorf("%w: import payload exceeds %d bytes", ErrInvalidFlowInput, maxImportPayloadBytes)
	}
	sourceFileHash := strings.TrimSpace(request.SourceFileHash)
	if sourceFileHash == "" {
		return Result{}, fmt.Errorf("%w: sourceFileHash is required", ErrInvalidFlowInput)
	}
	hash := sha256.Sum256(payload)
	actualFileHash := hex.EncodeToString(hash[:])
	if !strings.EqualFold(sourceFileHash, actualFileHash) {
		return Result{}, fmt.Errorf("%w: sourceFileHash does not match import payload", importer.ErrUnsafeAppend)
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
		return Result{}, err
	}

	appendRequests, err := importer.BuildAppendRequests(review, request.Decisions)
	if err != nil {
		return Result{}, err
	}
	if len(appendRequests) == 0 {
		return Result{}, ErrNoApprovedRows
	}

	transactions, err := appender.AppendImportedTransactions(ctx, request.RequestContext, request.SubjectID, request.IdempotencyKey, request.RequestPath, verticalslice.AppendImportBatchRequest{
		PortfolioID:        request.PortfolioID,
		Transactions:       appendRequests,
		SourceKind:         review.SourceKind,
		SourceAccountLabel: review.SourceAccountLabel,
		SourceFileHash:     review.FileHash,
		Decisions:          importDecisions(request.Decisions),
	})
	if err != nil {
		return Result{}, err
	}

	result := Result{
		ParsedRowCount:         review.Summary.TotalRows,
		AcceptedRowCount:       len(appendRequests),
		NonAppendedRowCount:    review.Summary.TotalRows - len(appendRequests),
		AppendedTransactionIDs: transactionIDs(transactions),
		SnapshotDatesRebuilt:   snapshotDates(appendRequests),
		AuditActionCode:        "IMPORT_APPEND_BATCH",
		NonSensitiveWarnings:   nonSensitiveWarnings(review, len(request.Decisions), len(appendRequests)),
	}
	return result, nil
}

func importDecisions(decisions []importer.Decision) []verticalslice.AppendImportDecision {
	result := make([]verticalslice.AppendImportDecision, 0, len(decisions))
	for _, decision := range decisions {
		result = append(result, verticalslice.AppendImportDecision{
			RowNumber: decision.RowNumber,
			Action:    decision.Action,
		})
	}
	return result
}

func transactionIDs(transactions []verticalslice.Transaction) []string {
	ids := make([]string, 0, len(transactions))
	for _, transaction := range transactions {
		ids = append(ids, transaction.ID)
	}
	return ids
}

func snapshotDates(requests []verticalslice.AppendTransactionRequest) []string {
	seen := map[string]struct{}{}
	for _, request := range requests {
		seen[request.TradeDate] = struct{}{}
	}
	dates := make([]string, 0, len(seen))
	for date := range seen {
		dates = append(dates, date)
	}
	sort.Strings(dates)
	return dates
}

func nonSensitiveWarnings(review importer.Review, decisionCount int, appendCount int) []string {
	warnings := []string{}
	if review.Summary.DuplicateRows > 0 {
		warnings = append(warnings, "DUPLICATE_ROWS_PRESENT")
	}
	if review.Summary.ConflictRows > 0 {
		warnings = append(warnings, "CONFLICT_ROWS_PRESENT")
	}
	if review.Summary.InvalidRows > 0 {
		warnings = append(warnings, "INVALID_ROWS_PRESENT")
	}
	if decisionCount < review.Summary.TotalRows {
		warnings = append(warnings, "UNDECIDED_ROWS_PRESENT")
	}
	if review.Summary.AppendableRows > appendCount {
		warnings = append(warnings, "UNAPPROVED_APPENDABLE_ROWS_PRESENT")
	}
	return warnings
}
