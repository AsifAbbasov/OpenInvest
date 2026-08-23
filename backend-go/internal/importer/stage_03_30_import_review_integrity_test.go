package importer

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

const stage330Header = "transaction_type,ticker,quantity,unit_price,gross_amount,commission,tax,trade_date,settlement_date,currency,broker_operation_id,note\n"

func stage330Rows(count int) string {
	var builder strings.Builder
	builder.WriteString(stage330Header)
	for row := 0; row < count; row++ {
		builder.WriteString("DEPOSIT,,,,")
		builder.WriteString(fmt.Sprintf("%d.00000000", 100+row))
		builder.WriteString(",0.00000000,0.00000000,2026-08-23,,RUB,broker-")
		builder.WriteString(fmt.Sprintf("%03d", row))
		builder.WriteString(",note\n")
	}
	return builder.String()
}

func TestStage330ReviewCSVEnforcesRowBoundDuringParse(t *testing.T) {
	if _, err := ReviewCSV(ReviewRequest{
		SubjectID:   "subject-a",
		PortfolioID: "portfolio-a",
		Reader:      strings.NewReader(stage330Rows(MaxReviewRows)),
	}); err != nil {
		t.Fatalf("expected exactly %d rows to be accepted: %v", MaxReviewRows, err)
	}

	_, err := ReviewCSV(ReviewRequest{
		SubjectID:   "subject-a",
		PortfolioID: "portfolio-a",
		Reader:      strings.NewReader(stage330Rows(MaxReviewRows + 1)),
	})
	if !errors.Is(err, ErrInvalidImport) {
		t.Fatalf("expected row 101 to fail during parse, got %v", err)
	}
}

func TestStage330SemanticDigestBindsNormalizedMeaningAndStatus(t *testing.T) {
	review, err := ReviewCSV(ReviewRequest{
		SubjectID:          "subject-a",
		PortfolioID:        "portfolio-a",
		SourceAccountLabel: "Broker A",
		Reader: strings.NewReader(stage330Header +
			"DEPOSIT,,,,100.00000000,0.00000000,0.00000000,2026-08-23,,RUB,broker-1,note\n"),
	})
	if err != nil {
		t.Fatalf("review CSV: %v", err)
	}
	first, err := ReviewSemanticDigest(review)
	if err != nil {
		t.Fatalf("digest first review: %v", err)
	}

	mutated := review
	mutated.Rows = append([]RowReview(nil), review.Rows...)
	mutated.Rows[0] = review.Rows[0]
	mutated.Rows[0].Status = ReviewStatusDuplicate
	second, err := ReviewSemanticDigest(mutated)
	if err != nil {
		t.Fatalf("digest mutated review: %v", err)
	}
	if first == second {
		t.Fatal("expected review status change to alter semantic digest")
	}

	mutated = review
	mutated.Rows = append([]RowReview(nil), review.Rows...)
	row := mutated.Rows[0]
	candidate := *row.Candidate
	note := "changed normalized note"
	candidate.SafeNote = &note
	row.Candidate = &candidate
	mutated.Rows[0] = row
	third, err := ReviewSemanticDigest(mutated)
	if err != nil {
		t.Fatalf("digest candidate change: %v", err)
	}
	if first == third {
		t.Fatal("expected normalized candidate change to alter semantic digest")
	}
}

func TestStage330ReviewHistoryFilterCollectsOnlyRelevantKeys(t *testing.T) {
	review, err := ReviewCSV(ReviewRequest{
		SubjectID:          "subject-a",
		PortfolioID:        "portfolio-a",
		SourceAccountLabel: " Broker A ",
		Reader: strings.NewReader(stage330Header +
			"DEPOSIT,,,,100.00000000,0.00000000,0.00000000,2026-08-23,,RUB,broker-1,note\n" +
			"DEPOSIT,,,,200.00000000,0.00000000,0.00000000,2026-08-23,,RUB,,note\n"),
	})
	if err != nil {
		t.Fatalf("review CSV: %v", err)
	}
	filter := ReviewHistoryFilter(review)
	if filter.SourceAccountLabel != "Broker A" {
		t.Fatalf("expected normalized source label, got %q", filter.SourceAccountLabel)
	}
	if len(filter.TradeDates) != 1 || filter.TradeDates[0] != "2026-08-23" {
		t.Fatalf("expected one unique trade date, got %v", filter.TradeDates)
	}
	if len(filter.BrokerOperationKeys) != 1 {
		t.Fatalf("expected one broker operation key, got %v", filter.BrokerOperationKeys)
	}
	if len(filter.SourceFingerprints) != 2 {
		t.Fatalf("expected two distinct financial fingerprints, got %v", filter.SourceFingerprints)
	}
}
