package importer

import (
	"testing"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestReviewCSVRejectsExistingFallbackThenStrongIdentityTransition(t *testing.T) {
	fallback := mustReview(t, csvHeader+
		"DEPOSIT,,,,1000.00000000,0.00000000,0.00000000,2026-06-19,,RUB,,fallback identity\n", nil)
	requests, err := BuildAppendRequests(fallback, []Decision{{
		RowNumber: 2,
		RowHash:   fallback.Rows[0].RowHash,
		Action:    DecisionApprove,
	}})
	if err != nil {
		t.Fatalf("build fallback append request: %v", err)
	}
	existing := persistedImportTransaction(requests[0], "manual-broker-label")

	strong := mustReview(t, csvHeader+
		"DEPOSIT,,,,1000.00000000,0.00000000,0.00000000,2026-06-19,,RUB,broker-op-a,strong identity\n", []verticalslice.Transaction{existing})
	if got := strong.Rows[0].Status; got != ReviewStatusConflict {
		t.Fatalf("expected fallback-to-strong transition to fail closed, got %s with reasons %v", got, strong.Rows[0].ReasonCodes)
	}
	assertHasReason(t, strong.Rows[0], "BROKER_OPERATION_IDENTITY_CONFLICT")
}

func TestReviewCSVRejectsExistingStrongThenFallbackIdentityTransition(t *testing.T) {
	strong := mustReview(t, csvHeader+
		"DEPOSIT,,,,1000.00000000,0.00000000,0.00000000,2026-06-19,,RUB,broker-op-a,strong identity\n", nil)
	requests, err := BuildAppendRequests(strong, []Decision{{
		RowNumber: 2,
		RowHash:   strong.Rows[0].RowHash,
		Action:    DecisionApprove,
	}})
	if err != nil {
		t.Fatalf("build strong append request: %v", err)
	}
	existing := persistedImportTransaction(requests[0], "manual-broker-label")

	fallback := mustReview(t, csvHeader+
		"DEPOSIT,,,,1000.00000000,0.00000000,0.00000000,2026-06-19,,RUB,,fallback identity\n", []verticalslice.Transaction{existing})
	if got := fallback.Rows[0].Status; got != ReviewStatusDuplicate {
		t.Fatalf("expected strong-to-fallback transition to be rejected as duplicate, got %s with reasons %v", got, fallback.Rows[0].ReasonCodes)
	}
}

func TestReviewCSVRejectsFallbackThenStrongIdentityInsideFile(t *testing.T) {
	review := mustReview(t, csvHeader+
		"DEPOSIT,,,,1000.00000000,0.00000000,0.00000000,2026-06-19,,RUB,,fallback identity\n"+
		"DEPOSIT,,,,1000.00000000,0.00000000,0.00000000,2026-06-19,,RUB,broker-op-a,strong identity\n", nil)

	if got := review.Rows[0].Status; got != ReviewStatusAppendable {
		t.Fatalf("expected first fallback row to remain appendable, got %s with reasons %v", got, review.Rows[0].ReasonCodes)
	}
	if got := review.Rows[1].Status; got != ReviewStatusConflict {
		t.Fatalf("expected later strong row with same fingerprint to fail closed, got %s with reasons %v", got, review.Rows[1].ReasonCodes)
	}
	assertHasReason(t, review.Rows[1], "MIXED_IMPORT_IDENTITY_STRENGTH_ROW_2")
}

func TestReviewCSVRejectsStrongThenFallbackIdentityInsideFile(t *testing.T) {
	review := mustReview(t, csvHeader+
		"DEPOSIT,,,,1000.00000000,0.00000000,0.00000000,2026-06-19,,RUB,broker-op-a,strong identity\n"+
		"DEPOSIT,,,,1000.00000000,0.00000000,0.00000000,2026-06-19,,RUB,,fallback identity\n", nil)

	if got := review.Rows[0].Status; got != ReviewStatusAppendable {
		t.Fatalf("expected first strong row to remain appendable, got %s with reasons %v", got, review.Rows[0].ReasonCodes)
	}
	if got := review.Rows[1].Status; got != ReviewStatusConflict {
		t.Fatalf("expected later fallback row with same fingerprint to fail closed, got %s with reasons %v", got, review.Rows[1].ReasonCodes)
	}
	assertHasReason(t, review.Rows[1], "MIXED_IMPORT_IDENTITY_STRENGTH_ROW_2")
}

func TestReviewCSVStillAllowsDistinctStrongIdentitiesWithSameFingerprint(t *testing.T) {
	review := mustReview(t, csvHeader+
		"BUY,SBER,2.00000000,100.00000000,200.00000000,1.00000000,0.00000000,2026-06-20,,RUB,broker-op-a,first execution\n"+
		"BUY,SBER,2.00000000,100.00000000,200.00000000,1.00000000,0.00000000,2026-06-20,,RUB,broker-op-b,second execution\n", nil)

	if review.Summary.AppendableRows != 2 || review.Summary.ConflictRows != 0 || review.Summary.DuplicateRows != 0 {
		t.Fatalf("expected two distinct strong broker identities to remain appendable, got %+v", review.Summary)
	}
}

func persistedImportTransaction(request verticalslice.AppendTransactionRequest, sourceAccountLabel string) verticalslice.Transaction {
	gross := *request.GrossAmount
	provenance := request.ImportProvenance
	return verticalslice.Transaction{
		PortfolioID:              request.PortfolioID,
		TransactionType:          request.TransactionType,
		Ticker:                   request.Ticker,
		Quantity:                 request.Quantity,
		UnitPrice:                request.UnitPrice,
		GrossAmount:              gross,
		Commission:               request.Commission,
		Tax:                      request.Tax,
		TradeDate:                request.TradeDate,
		SettlementDate:           request.SettlementDate,
		SourceKind:               SourceKindUserUploadedFile,
		SourceAccountLabel:       sourceAccountLabel,
		SourceBrokerOperationKey: provenance.BrokerOperationKey,
		SourceFingerprint:        provenance.SourceFingerprint,
		SourceIdentityVersion:    provenance.IdentityVersion,
	}
}
