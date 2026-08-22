package importer

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

const csvHeader = "transaction_type,ticker,quantity,unit_price,gross_amount,commission,tax,trade_date,settlement_date,currency,broker_operation_id,note\n"

func TestReviewCSVClassifiesAppendableRows(t *testing.T) {
	review := mustReview(t, csvHeader+
		"BUY,SBER,2.00000000,100.00000000,200.00000000,1.00000000,0.00000000,2026-06-20,2026-06-21,RUB,op-1,ordinary buy\n"+
		"DEPOSIT,,,,1000.00000000,0.00000000,0.00000000,2026-06-19,,RUB,op-2,cash in\n"+
		"WITHDRAWAL,,,,250.00000000,0.00000000,0.00000000,2026-06-22,,RUB,op-3,cash out\n", nil)

	if review.Summary.TotalRows != 3 || review.Summary.AppendableRows != 3 {
		t.Fatalf("expected 3 appendable rows, got %+v", review.Summary)
	}

	appendRequests, err := BuildAppendRequests(review, []Decision{
		{RowNumber: 2, RowHash: review.Rows[0].RowHash, Action: DecisionApprove},
		{RowNumber: 3, RowHash: review.Rows[1].RowHash, Action: DecisionApprove},
		{RowNumber: 4, RowHash: review.Rows[2].RowHash, Action: DecisionIgnore},
	})
	if err != nil {
		t.Fatalf("build append requests: %v", err)
	}
	if len(appendRequests) != 2 {
		t.Fatalf("expected 2 append requests, got %d", len(appendRequests))
	}
	if appendRequests[0].TransactionType != "BUY" || appendRequests[1].TransactionType != "DEPOSIT" {
		t.Fatalf("unexpected append request types: %+v", appendRequests)
	}
}

func TestReviewCSVDetectsExactExistingDuplicate(t *testing.T) {
	ticker := "SBER"
	quantity := decimal.Must("2.00000000")
	unitPrice := verticalslice.Money{Amount: decimal.Must("100.00000000"), Currency: verticalslice.RUB}
	existing := []verticalslice.Transaction{{
		TransactionType: "BUY",
		Ticker:          &ticker,
		Quantity:        &quantity,
		UnitPrice:       &unitPrice,
		GrossAmount:     verticalslice.Money{Amount: decimal.Must("200.00000000"), Currency: verticalslice.RUB},
		Commission:      verticalslice.Money{Amount: decimal.Must("1.00000000"), Currency: verticalslice.RUB},
		Tax:             verticalslice.ZeroMoney(),
		TradeDate:       "2026-06-20",
	}}

	review := mustReview(t, csvHeader+
		"BUY,SBER,2.00000000,100.00000000,200.00000000,1.00000000,0.00000000,2026-06-20,,RUB,op-1,duplicate\n", existing)

	if got := review.Rows[0].Status; got != ReviewStatusDuplicate {
		t.Fatalf("expected duplicate, got %s with reasons %v", got, review.Rows[0].ReasonCodes)
	}
}

func TestReviewCSVDetectsScopedBrokerOperationDuplicateInsideFile(t *testing.T) {
	review := mustReview(t, csvHeader+
		"DEPOSIT,,,,1000.00000000,0.00000000,0.00000000,2026-06-19,,RUB,broker-op-1,cash in\n"+
		"DEPOSIT,,,,1000.00000000,0.00000000,0.00000000,2026-06-19,,RUB,broker-op-1,cash in duplicate\n", nil)

	if got := review.Rows[1].Status; got != ReviewStatusDuplicate {
		t.Fatalf("expected broker operation duplicate, got %s with reasons %v", got, review.Rows[1].ReasonCodes)
	}
}

func TestReviewCSVRejectsBrokerIdentityCollisionInsideFile(t *testing.T) {
	review := mustReview(t, csvHeader+
		"DEPOSIT,,,,1000.00000000,0.00000000,0.00000000,2026-06-19,,RUB,broker-op-1,first cash in\n"+
		"DEPOSIT,,,,2000.00000000,0.00000000,0.00000000,2026-06-19,,RUB,broker-op-1,mutated cash in\n", nil)

	if got := review.Rows[1].Status; got != ReviewStatusConflict {
		t.Fatalf("expected broker identity collision to be a conflict, got %s with reasons %v", got, review.Rows[1].ReasonCodes)
	}
	assertHasReason(t, review.Rows[1], "BROKER_OPERATION_IDENTITY_CONFLICT_ROW_2")
}

func TestReviewCSVKeepsDistinctBrokerOperationsWithIdenticalEconomics(t *testing.T) {
	review := mustReview(t, csvHeader+
		"BUY,SBER,2.00000000,100.00000000,200.00000000,1.00000000,0.00000000,2026-06-20,,RUB,broker-op-a,first execution\n"+
		"BUY,SBER,2.00000000,100.00000000,200.00000000,1.00000000,0.00000000,2026-06-20,,RUB,broker-op-b,second execution\n", nil)

	if review.Summary.AppendableRows != 2 {
		t.Fatalf("expected both independent broker operations to remain appendable, got %+v", review.Summary)
	}

	requests, err := BuildAppendRequests(review, []Decision{
		{RowNumber: 2, RowHash: review.Rows[0].RowHash, Action: DecisionApprove},
		{RowNumber: 3, RowHash: review.Rows[1].RowHash, Action: DecisionApprove},
	})
	if err != nil {
		t.Fatalf("build append requests: %v", err)
	}
	if len(requests) != 2 || requests[0].ImportProvenance == nil || requests[1].ImportProvenance == nil {
		t.Fatalf("expected per-row import provenance, got %+v", requests)
	}
	if requests[0].ImportProvenance.SourceFingerprint != requests[1].ImportProvenance.SourceFingerprint {
		t.Fatal("expected identical economics to share the normalized source fingerprint")
	}
	if requests[0].ImportProvenance.BrokerOperationKey == requests[1].ImportProvenance.BrokerOperationKey {
		t.Fatal("expected distinct broker operations to have distinct persisted identity keys")
	}
}

func TestReviewCSVAllowsDifferentCashAmountsOnSameDate(t *testing.T) {
	review := mustReview(t, csvHeader+
		"DEPOSIT,,,,1000.00000000,0.00000000,0.00000000,2026-06-19,,RUB,broker-op-a,first cash in\n"+
		"DEPOSIT,,,,2000.00000000,0.00000000,0.00000000,2026-06-19,,RUB,broker-op-b,second cash in\n", nil)

	if review.Summary.AppendableRows != 2 || review.Summary.ConflictRows != 0 {
		t.Fatalf("expected different same-day cash amounts to be independent, got %+v", review.Summary)
	}
}

func TestReviewCSVRejectsCashFlowFeesUntilMethodologyExists(t *testing.T) {
	review := mustReview(t, csvHeader+
		"DEPOSIT,,,,1000.00000000,1.00000000,0.00000000,2026-06-19,,RUB,broker-op-a,cash fee unsupported\n", nil)

	if got := review.Rows[0].Status; got != ReviewStatusConflict {
		t.Fatalf("expected unsupported cash-flow fee conflict, got %s with reasons %v", got, review.Rows[0].ReasonCodes)
	}
	assertHasReason(t, review.Rows[0], "CASH_FLOW_FEES_UNSUPPORTED")
}

func TestReviewCSVUsesPersistedBrokerIdentityForExistingRows(t *testing.T) {
	first := mustReview(t, csvHeader+
		"DEPOSIT,,,,1000.00000000,0.00000000,0.00000000,2026-06-19,,RUB,broker-op-stable,first cash in\n", nil)
	requests, err := BuildAppendRequests(first, []Decision{{RowNumber: 2, RowHash: first.Rows[0].RowHash, Action: DecisionApprove}})
	if err != nil {
		t.Fatalf("build first append request: %v", err)
	}
	request := requests[0]
	existing := verticalslice.Transaction{
		PortfolioID:              request.PortfolioID,
		TransactionType:          request.TransactionType,
		GrossAmount:              *request.GrossAmount,
		Commission:               request.Commission,
		Tax:                      request.Tax,
		TradeDate:                request.TradeDate,
		SourceKind:               SourceKindUserUploadedFile,
		SourceAccountLabel:       "manual-broker-label",
		SourceBrokerOperationKey: request.ImportProvenance.BrokerOperationKey,
		SourceFingerprint:        request.ImportProvenance.SourceFingerprint,
		SourceIdentityVersion:    verticalslice.ImportIdentityVersion,
	}

	repeated := mustReview(t, csvHeader+
		"DEPOSIT,,,,1000.00000000,0.00000000,0.00000000,2026-06-19,,RUB,broker-op-stable,repeated cash in\n", []verticalslice.Transaction{existing})
	if got := repeated.Rows[0].Status; got != ReviewStatusDuplicate {
		t.Fatalf("expected persisted broker identity duplicate, got %s with reasons %v", got, repeated.Rows[0].ReasonCodes)
	}

	changed := mustReview(t, csvHeader+
		"DEPOSIT,,,,2000.00000000,0.00000000,0.00000000,2026-06-19,,RUB,broker-op-stable,changed economics\n", []verticalslice.Transaction{existing})
	if got := changed.Rows[0].Status; got != ReviewStatusConflict {
		t.Fatalf("expected broker identity collision to fail closed, got %s with reasons %v", got, changed.Rows[0].ReasonCodes)
	}
	assertHasReason(t, changed.Rows[0], "BROKER_OPERATION_IDENTITY_CONFLICT")
}

func TestReviewCSVScopesPersistedBrokerIdentityBySourceAccount(t *testing.T) {
	first := mustReview(t, csvHeader+
		"DEPOSIT,,,,1000.00000000,0.00000000,0.00000000,2026-06-19,,RUB,broker-op-scoped,cash in\n", nil)
	requests, err := BuildAppendRequests(first, []Decision{{RowNumber: 2, RowHash: first.Rows[0].RowHash, Action: DecisionApprove}})
	if err != nil {
		t.Fatalf("build first append request: %v", err)
	}
	request := requests[0]
	existing := verticalslice.Transaction{
		PortfolioID:              request.PortfolioID,
		TransactionType:          request.TransactionType,
		GrossAmount:              *request.GrossAmount,
		Commission:               request.Commission,
		Tax:                      request.Tax,
		TradeDate:                request.TradeDate,
		SourceKind:               SourceKindUserUploadedFile,
		SourceAccountLabel:       "another-broker-account",
		SourceBrokerOperationKey: request.ImportProvenance.BrokerOperationKey,
		SourceFingerprint:        request.ImportProvenance.SourceFingerprint,
		SourceIdentityVersion:    verticalslice.ImportIdentityVersion,
	}

	review := mustReview(t, csvHeader+
		"DEPOSIT,,,,1000.00000000,0.00000000,0.00000000,2026-06-19,,RUB,broker-op-scoped,cash in on scoped account\n", []verticalslice.Transaction{existing})
	if got := review.Rows[0].Status; got != ReviewStatusAppendable {
		t.Fatalf("expected same broker operation key in another source account to remain appendable, got %s with reasons %v", got, review.Rows[0].ReasonCodes)
	}
}

func TestReviewCSVMarksGrossAmountMismatchForBuyAsConflict(t *testing.T) {
	review := mustReview(t, csvHeader+
		"BUY,SBER,2.00000000,100.00000000,199.99000000,1.00000000,0.00000000,2026-06-20,,RUB,op-1,gross mismatch\n", nil)

	if got := review.Rows[0].Status; got != ReviewStatusConflict {
		t.Fatalf("expected gross mismatch conflict, got %s with reasons %v", got, review.Rows[0].ReasonCodes)
	}
	assertHasReason(t, review.Rows[0], "GROSS_AMOUNT_MISMATCH")
}

func TestReviewCSVDetectsNearDuplicateInsideImportFile(t *testing.T) {
	review := mustReview(t, csvHeader+
		"BUY,SBER,2.00000000,100.00000000,200.00000000,1.00000000,0.00000000,2026-06-20,,RUB,op-1,first buy\n"+
		"BUY,SBER,2.00000000,100.00000000,200.00000000,2.00000000,0.00000000,2026-06-20,,RUB,op-2,near duplicate commission differs\n", nil)

	if got := review.Rows[0].Status; got != ReviewStatusAppendable {
		t.Fatalf("expected first row appendable, got %s with reasons %v", got, review.Rows[0].ReasonCodes)
	}
	if got := review.Rows[1].Status; got != ReviewStatusConflict {
		t.Fatalf("expected second row near duplicate conflict, got %s with reasons %v", got, review.Rows[1].ReasonCodes)
	}
	assertHasReason(t, review.Rows[1], "NEAR_DUPLICATE_IMPORTED_ROW_2_REQUIRES_REVIEW")
}

func TestReviewCSVMarksUnsupportedAndUnsafeRowsForReview(t *testing.T) {
	review := mustReview(t, csvHeader+
		"SELL,SBER,1.00000000,100.00000000,100.00000000,0.00000000,0.00000000,2026-06-20,,RUB,op-1,sell later\n"+
		"BUY,SBER,1.00000000,100.00000000,100.00000000,0.00000000,0.00000000,2026-06-20,,USD,op-2,wrong currency\n", nil)

	if review.Summary.ConflictRows != 2 {
		t.Fatalf("expected 2 conflict rows, got %+v", review.Summary)
	}
	for _, row := range review.Rows {
		if row.Status != ReviewStatusConflict {
			t.Fatalf("expected conflict, got %s", row.Status)
		}
	}
}

func TestReviewCSVNeutralizesSpreadsheetFormulaPayloads(t *testing.T) {
	review := mustReview(t, csvHeader+
		"DEPOSIT,,,,1000.00000000,0.00000000,0.00000000,2026-06-19,,RUB,\"=BROKEROP()\",\"=HYPERLINK(\"\"https://evil.example\"\")\"\n", nil)

	note := review.Rows[0].Candidate.SafeNote
	if note == nil {
		t.Fatal("expected safe note")
	}
	if !strings.HasPrefix(*note, "'=") {
		t.Fatalf("expected neutralized spreadsheet formula, got %q", *note)
	}
	if !strings.HasPrefix(review.Rows[0].BrokerOperationID, "'=") {
		t.Fatalf("expected neutralized broker operation id, got %q", review.Rows[0].BrokerOperationID)
	}
	requests, err := BuildAppendRequests(review, []Decision{{RowNumber: 2, RowHash: review.Rows[0].RowHash, Action: DecisionApprove}})
	if err != nil {
		t.Fatalf("build append request: %v", err)
	}
	if got := requests[0].ImportProvenance.BrokerOperationKey; got != verticalslice.BrokerOperationKey("=BROKEROP()") {
		t.Fatalf("expected identity key from raw broker operation id, got %q", got)
	}
}

func TestBuildAppendRequestsRejectsUnapprovedUnsafeRows(t *testing.T) {
	review := mustReview(t, csvHeader+
		"SELL,SBER,1.00000000,100.00000000,100.00000000,0.00000000,0.00000000,2026-06-20,,RUB,op-1,sell later\n", nil)

	_, err := BuildAppendRequests(review, []Decision{{RowNumber: 2, RowHash: review.Rows[0].RowHash, Action: DecisionApprove}})
	if !errors.Is(err, ErrUnsafeAppend) {
		t.Fatalf("expected unsafe append error, got %v", err)
	}
}

func TestBuildAppendRequestsRejectsRowHashMismatch(t *testing.T) {
	review := mustReview(t, csvHeader+
		"DEPOSIT,,,,1000.00000000,0.00000000,0.00000000,2026-06-19,,RUB,op-1,cash in\n", nil)

	_, err := BuildAppendRequests(review, []Decision{{RowNumber: 2, RowHash: strings.Repeat("0", 64), Action: DecisionApprove}})
	if !errors.Is(err, ErrUnsafeAppend) {
		t.Fatalf("expected unsafe append error, got %v", err)
	}
}

func TestBuildAppendRequestsRejectsDuplicateDecisions(t *testing.T) {
	review := mustReview(t, csvHeader+
		"DEPOSIT,,,,1000.00000000,0.00000000,0.00000000,2026-06-19,,RUB,op-1,cash in\n", nil)

	_, err := BuildAppendRequests(review, []Decision{
		{RowNumber: 2, RowHash: review.Rows[0].RowHash, Action: DecisionApprove},
		{RowNumber: 2, RowHash: review.Rows[0].RowHash, Action: DecisionApprove},
	})
	if !errors.Is(err, ErrUnsafeAppend) {
		t.Fatalf("expected unsafe append error, got %v", err)
	}
}

func TestReviewCSVRejectsMissingRequiredHeader(t *testing.T) {
	_, err := ReviewCSV(ReviewRequest{
		SubjectID:   "subject",
		PortfolioID: "portfolio",
		SourceKind:  SourceKindUserUploadedFile,
		Reader:      strings.NewReader("transaction_type\nBUY\n"),
	})

	if !errors.Is(err, ErrInvalidImport) {
		t.Fatalf("expected invalid import, got %v", err)
	}
}

func TestReviewCSVReadsCanonicalFinancialFixture(t *testing.T) {
	payload, err := os.ReadFile("../../../tests/financial/import/valid_stage_03_06.csv")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	review := mustReview(t, string(payload), nil)

	if review.Summary.TotalRows != 3 || review.Summary.AppendableRows != 3 {
		t.Fatalf("expected valid fixture to be appendable, got %+v", review.Summary)
	}
}

func TestReviewCSVReadsAllCanonicalFinancialFixtures(t *testing.T) {
	tests := map[string]Summary{
		"../../../tests/financial/import/valid_stage_03_06.csv": {
			TotalRows:      3,
			AppendableRows: 3,
		},
		"../../../tests/financial/import/conflicts_stage_03_06.csv": {
			TotalRows:    4,
			ConflictRows: 4,
		},
		"../../../tests/financial/import/formula_injection_stage_03_06.csv": {
			TotalRows:      2,
			AppendableRows: 2,
		},
	}

	for path, want := range tests {
		t.Run(path, func(t *testing.T) {
			payload, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}

			review := mustReview(t, string(payload), nil)
			if review.Summary != want {
				t.Fatalf("unexpected fixture summary: got %+v want %+v", review.Summary, want)
			}
		})
	}
}

func mustReview(t *testing.T, input string, existing []verticalslice.Transaction) Review {
	t.Helper()
	review, err := ReviewCSV(ReviewRequest{
		SubjectID:          "subject-1",
		PortfolioID:        "portfolio-1",
		SourceKind:         SourceKindUserUploadedFile,
		SourceAccountLabel: "manual-broker-label",
		FileHash:           "file-hash",
		Existing:           existing,
		Reader:             strings.NewReader(input),
	})
	if err != nil {
		t.Fatalf("review csv: %v", err)
	}
	return review
}

func assertHasReason(t *testing.T, row RowReview, reason string) {
	t.Helper()
	for _, got := range row.ReasonCodes {
		if got == reason {
			return
		}
	}
	t.Fatalf("expected row %d to contain reason %q, got %v", row.RowNumber, reason, row.ReasonCodes)
}
