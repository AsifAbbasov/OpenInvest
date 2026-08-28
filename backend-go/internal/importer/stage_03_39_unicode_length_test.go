package importer

import (
	"errors"
	"strings"
	"testing"
)

const stage39UnicodeCSV = "transaction_type,ticker,quantity,unit_price,gross_amount,commission,tax,trade_date,settlement_date,currency,broker_operation_id,note\n" +
	"DEPOSIT,,,,100.00000000,0.00000000,0.00000000,2026-08-27,,RUB,stage39-op-1,unicode label test\n"

func TestStage339ImporterSourceLabelUsesUnicodeCodePoints(t *testing.T) {
	tests := []struct {
		name      string
		label     string
		wantLabel string
		wantError bool
	}{
		{name: "120 Cyrillic", label: strings.Repeat("Ж", 120), wantLabel: strings.Repeat("Ж", 120)},
		{name: "121 Cyrillic", label: strings.Repeat("Ж", 121), wantError: true},
		{name: "120 supplementary", label: strings.Repeat("😀", 120), wantLabel: strings.Repeat("😀", 120)},
		{name: "121 supplementary", label: strings.Repeat("😀", 121), wantError: true},
		{name: "trim identity unchanged", label: " Broker account A ", wantLabel: "Broker account A"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			review, err := ReviewCSV(stage39ReviewRequest(tc.label))
			if tc.wantError {
				if !errors.Is(err, ErrInvalidImport) {
					t.Fatalf("expected invalid import, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("review CSV: %v", err)
			}
			if review.SourceAccountLabel != tc.wantLabel {
				t.Fatalf("source label trim/identity changed: got %q want %q", review.SourceAccountLabel, tc.wantLabel)
			}
		})
	}
}

func TestStage339ImporterRejectsMalformedSourceLabelWithoutParserVersionChange(t *testing.T) {
	_, err := ReviewCSV(stage39ReviewRequest(string([]byte{0xff, 'A'})))
	if !errors.Is(err, ErrInvalidImport) {
		t.Fatalf("expected malformed source label to fail closed, got %v", err)
	}
	if ReviewParserVersion != 2 {
		t.Fatalf("P3-04 must not bump ReviewParserVersion, got %d", ReviewParserVersion)
	}
}

func TestStage339ImporterRejectsMalformedCSVNoteBeforeCodePointCounting(t *testing.T) {
	payload := "transaction_type,ticker,quantity,unit_price,gross_amount,commission,tax,trade_date,settlement_date,currency,broker_operation_id,note\n" +
		"DEPOSIT,,,,100.00000000,0.00000000,0.00000000,2026-08-27,,RUB,stage39-op-invalid,=" +
		string([]byte{0xff}) + "\n"

	review, err := ReviewCSV(ReviewRequest{
		SubjectID:          "subject",
		PortfolioID:        "portfolio-id",
		SourceKind:         SourceKindUserUploadedFile,
		SourceAccountLabel: "Broker account A",
		FileHash:           strings.Repeat("b", 64),
		Reader:             strings.NewReader(payload),
	})
	if err != nil {
		t.Fatalf("review malformed note CSV: %v", err)
	}
	if len(review.Rows) != 1 {
		t.Fatalf("expected one reviewed row, got %d", len(review.Rows))
	}
	row := review.Rows[0]
	if row.Status != ReviewStatusInvalid || row.Candidate != nil {
		t.Fatalf("malformed UTF-8 note must be INVALID/non-appendable, got %+v", row)
	}

	found := false
	for _, reason := range row.ReasonCodes {
		if reason == "NOTE_INVALID_UTF8" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected NOTE_INVALID_UTF8 reason, got %v", row.ReasonCodes)
	}
	if ReviewParserVersion != 2 {
		t.Fatalf("malformed note guard must not bump ReviewParserVersion, got %d", ReviewParserVersion)
	}
}

func stage39ReviewRequest(label string) ReviewRequest {
	return ReviewRequest{
		SubjectID:          "subject",
		PortfolioID:        "portfolio-id",
		SourceKind:         SourceKindUserUploadedFile,
		SourceAccountLabel: label,
		FileHash:           strings.Repeat("a", 64),
		Reader:             strings.NewReader(stage39UnicodeCSV),
	}
}
