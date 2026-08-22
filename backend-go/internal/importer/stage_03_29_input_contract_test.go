package importer

import (
	"errors"
	"strings"
	"testing"
)

const stage329CSVHeader = "transaction_type,ticker,quantity,unit_price,gross_amount,commission,tax,trade_date,settlement_date,currency,broker_operation_id,note\n"

func TestStage329ReviewCSVRejectsDuplicateNormalizedHeaders(t *testing.T) {
	payload := "transaction_type,ticker, TICKER ,quantity,unit_price,gross_amount,commission,tax,trade_date,settlement_date,currency,broker_operation_id,note\n" +
		"BUY,SBER,SBER,2.00000000,100.00000000,200.00000000,0.00000000,0.00000000,2026-01-10,2026-01-13,RUB,broker-1,note\n"

	_, err := ReviewCSV(ReviewRequest{
		SubjectID:   "subject-a",
		PortfolioID: "portfolio-a",
		Reader:      strings.NewReader(payload),
	})
	if !errors.Is(err, ErrInvalidImport) {
		t.Fatalf("expected duplicate normalized header to fail closed, got %v", err)
	}
}

func TestStage329ReviewCSVRejectsOverlongSafeNoteBeforeAppend(t *testing.T) {
	payload := stage329CSVHeader +
		"BUY,SBER,2.00000000,100.00000000,200.00000000,0.00000000,0.00000000,2026-01-10,2026-01-13,RUB,broker-1," +
		strings.Repeat("Ж", 501) + "\n"

	review, err := ReviewCSV(ReviewRequest{
		SubjectID:   "subject-a",
		PortfolioID: "portfolio-a",
		Reader:      strings.NewReader(payload),
	})
	if err != nil {
		t.Fatalf("review CSV: %v", err)
	}
	if len(review.Rows) != 1 {
		t.Fatalf("expected one reviewed row, got %d", len(review.Rows))
	}
	if review.Rows[0].Status != ReviewStatusInvalid || review.Rows[0].Candidate != nil {
		t.Fatalf("expected overlong note row to be INVALID/non-appendable, got %+v", review.Rows[0])
	}
	found := false
	for _, reason := range review.Rows[0].ReasonCodes {
		if reason == "NOTE_TOO_LONG" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected NOTE_TOO_LONG reason, got %v", review.Rows[0].ReasonCodes)
	}
}

func TestStage329ReviewCSVAcceptsMaximumSafeNoteLength(t *testing.T) {
	payload := stage329CSVHeader +
		"BUY,SBER,2.00000000,100.00000000,200.00000000,0.00000000,0.00000000,2026-01-10,2026-01-13,RUB,broker-1," +
		strings.Repeat("Ж", 500) + "\n"

	review, err := ReviewCSV(ReviewRequest{
		SubjectID:   "subject-a",
		PortfolioID: "portfolio-a",
		Reader:      strings.NewReader(payload),
	})
	if err != nil {
		t.Fatalf("review CSV: %v", err)
	}
	if len(review.Rows) != 1 || review.Rows[0].Status != ReviewStatusAppendable {
		t.Fatalf("expected 500-character note row to remain appendable, got %+v", review.Rows)
	}
}
