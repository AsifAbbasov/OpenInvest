package importflow

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/openinvest/openinvest/backend-go/internal/importer"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

const csvHeader = "transaction_type,ticker,quantity,unit_price,gross_amount,commission,tax,trade_date,settlement_date,currency,broker_operation_id,note\n"

type recordingAppender struct {
	request verticalslice.AppendImportBatchRequest
	called  bool
}

func (appender *recordingAppender) AppendImportedTransactions(_ context.Context, _ verticalslice.RequestContext, _ string, _ string, _ string, request verticalslice.AppendImportBatchRequest) ([]verticalslice.Transaction, error) {
	appender.called = true
	appender.request = request
	return []verticalslice.Transaction{
		{ID: "tx-1", TradeDate: request.Transactions[0].TradeDate},
	}, nil
}

func TestReviewAndAppendAppendsOnlyExplicitlyApprovedRows(t *testing.T) {
	appender := &recordingAppender{}

	result, err := ReviewAndAppend(context.Background(), appender, Request{
		SubjectID:          "subject-1",
		PortfolioID:        "portfolio-1",
		IdempotencyKey:     "import-flow-key-0001",
		RequestPath:        "/internal/imports/review-append",
		SourceAccountLabel: "manual-broker-label",
		Reader: strings.NewReader(csvHeader +
			"DEPOSIT,,,,1000.00000000,0.00000000,0.00000000,2026-06-19,,RUB,op-1,cash in\n" +
			"BUY,SBER,2.00000000,100.00000000,200.00000000,1.00000000,0.00000000,2026-06-20,2026-06-21,RUB,op-2,buy\n"),
		Decisions: []importer.Decision{
			{RowNumber: 2, Action: importer.DecisionApprove},
			{RowNumber: 3, Action: importer.DecisionIgnore},
		},
	})
	if err != nil {
		t.Fatalf("review and append: %v", err)
	}
	if !appender.called {
		t.Fatal("expected appender to be called")
	}
	if len(appender.request.Transactions) != 1 {
		t.Fatalf("expected one approved append request, got %d", len(appender.request.Transactions))
	}
	if appender.request.SourceKind != importer.SourceKindUserUploadedFile {
		t.Fatalf("unexpected source kind: %s", appender.request.SourceKind)
	}
	if appender.request.SourceFileHash == "" {
		t.Fatal("expected computed source file hash")
	}
	if result.ParsedRowCount != 2 || result.AcceptedRowCount != 1 || result.NonAppendedRowCount != 1 {
		t.Fatalf("unexpected result counts: %+v", result)
	}
	if len(result.AppendedTransactionIDs) != 1 || result.AppendedTransactionIDs[0] != "tx-1" {
		t.Fatalf("unexpected transaction ids: %v", result.AppendedTransactionIDs)
	}
	if len(result.SnapshotDatesRebuilt) != 1 || result.SnapshotDatesRebuilt[0] != "2026-06-19" {
		t.Fatalf("unexpected snapshot dates: %v", result.SnapshotDatesRebuilt)
	}
	if result.AuditActionCode != "IMPORT_APPEND_BATCH" {
		t.Fatalf("unexpected audit action code: %s", result.AuditActionCode)
	}
	if len(result.NonSensitiveWarnings) == 0 {
		t.Fatal("expected non-sensitive warning for unapproved appendable row")
	}
}

func TestReviewAndAppendRejectsUnsafeApproval(t *testing.T) {
	appender := &recordingAppender{}

	_, err := ReviewAndAppend(context.Background(), appender, Request{
		SubjectID:      "subject-1",
		PortfolioID:    "portfolio-1",
		IdempotencyKey: "import-flow-key-0002",
		RequestPath:    "/internal/imports/review-append",
		Reader: strings.NewReader(csvHeader +
			"SELL,SBER,1.00000000,100.00000000,100.00000000,0.00000000,0.00000000,2026-06-20,,RUB,op-1,sell later\n"),
		Decisions: []importer.Decision{{RowNumber: 2, Action: importer.DecisionApprove}},
	})
	if !errors.Is(err, importer.ErrUnsafeAppend) {
		t.Fatalf("expected unsafe append error, got %v", err)
	}
	if appender.called {
		t.Fatal("appender must not be called for unsafe approval")
	}
}

func TestReviewAndAppendRequiresApprovedRows(t *testing.T) {
	appender := &recordingAppender{}

	_, err := ReviewAndAppend(context.Background(), appender, Request{
		SubjectID:      "subject-1",
		PortfolioID:    "portfolio-1",
		IdempotencyKey: "import-flow-key-0003",
		RequestPath:    "/internal/imports/review-append",
		Reader: strings.NewReader(csvHeader +
			"DEPOSIT,,,,1000.00000000,0.00000000,0.00000000,2026-06-19,,RUB,op-1,cash in\n"),
		Decisions: []importer.Decision{{RowNumber: 2, Action: importer.DecisionIgnore}},
	})
	if !errors.Is(err, ErrNoApprovedRows) {
		t.Fatalf("expected no approved rows error, got %v", err)
	}
	if appender.called {
		t.Fatal("appender must not be called when no rows are approved")
	}
}

func TestReviewAndAppendRejectsOversizedPayload(t *testing.T) {
	appender := &recordingAppender{}

	_, err := ReviewAndAppend(context.Background(), appender, Request{
		SubjectID:      "subject-1",
		PortfolioID:    "portfolio-1",
		IdempotencyKey: "import-flow-key-0004",
		RequestPath:    "/internal/imports/review-append",
		Reader:         strings.NewReader(strings.Repeat("x", int(maxImportPayloadBytes)+1)),
		Decisions:      []importer.Decision{{RowNumber: 2, Action: importer.DecisionApprove}},
	})
	if !errors.Is(err, ErrInvalidFlowInput) {
		t.Fatalf("expected invalid flow input for oversized payload, got %v", err)
	}
	if appender.called {
		t.Fatal("appender must not be called for oversized payload")
	}
}
