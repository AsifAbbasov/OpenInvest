package importer

import (
	"bytes"
	"testing"
)

const fuzzCSVHeader = "transaction_type,ticker,quantity,unit_price,gross_amount,commission,tax,trade_date,settlement_date,currency,broker_operation_id,note\n"

func FuzzReviewCSV(f *testing.F) {
	for _, seed := range [][]byte{
		[]byte(fuzzCSVHeader + "BUY,SBER,2.00000000,100.00000000,200.00000000,1.00000000,0.00000000,2026-01-10,2026-01-13,RUB,broker-row-1,Imported buy\n"),
		[]byte(fuzzCSVHeader + "DEPOSIT,,,,100.00000000,0.00000000,0.00000000,2026-08-27,,RUB,deposit-1,note\n"),
		[]byte(""),
		[]byte("a,b,c\n1,2,3\n"),
		[]byte(fuzzCSVHeader + "\"unterminated"),
	} {
		f.Add(seed)
	}

	review := func(payload []byte) (Review, error) {
		return ReviewCSV(ReviewRequest{
			SubjectID:          "fuzz-subject",
			PortfolioID:        "fuzz-portfolio",
			SourceKind:         SourceKindUserUploadedFile,
			SourceAccountLabel: "fuzz-account",
			FileHash:           "fuzz",
			Reader:             bytes.NewReader(payload),
		})
	}

	f.Fuzz(func(t *testing.T, payload []byte) {
		first, firstErr := review(payload)
		second, secondErr := review(payload)

		if (firstErr == nil) != (secondErr == nil) {
			t.Fatalf("same CSV input produced non-deterministic success/error result")
		}
		if firstErr != nil {
			return
		}

		if len(first.Rows) > MaxReviewRows {
			t.Fatalf("accepted review exceeds MaxReviewRows: got=%d max=%d", len(first.Rows), MaxReviewRows)
		}
		if first.Summary.TotalRows != len(first.Rows) {
			t.Fatalf("summary totalRows mismatch: got=%d rows=%d", first.Summary.TotalRows, len(first.Rows))
		}

		firstDigest, err := ReviewSemanticDigest(first)
		if err != nil {
			t.Fatalf("digest first review: %v", err)
		}
		secondDigest, err := ReviewSemanticDigest(second)
		if err != nil {
			t.Fatalf("digest second review: %v", err)
		}
		if firstDigest != secondDigest {
			t.Fatalf("same CSV input produced non-deterministic semantic digest")
		}
	})
}
