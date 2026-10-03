package importer

import (
	"strings"
	"testing"
)

func FuzzReviewCSV(f *testing.F) {
	for _, seed := range []string{
		csvHeader,
		csvHeader + "DEPOSIT,,,,1000.00000000,0.00000000,0.00000000,2026-06-19,,RUB,op-1,cash in\n",
		csvHeader + "BUY,SBER,2.00000000,100.00000000,200.00000000,1.00000000,0.00000000,2026-06-20,,RUB,op-2,buy\n",
		"transaction_type\nBUY\n",
		"\x00\xff\xfe",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, payload string) {
		// Keep one fuzz execution bounded. Large-body admission is tested separately
		// at the HTTP/resource layer; this target is for parser/state-machine behavior.
		if len(payload) > 256*1024 {
			return
		}

		review, err := ReviewCSV(ReviewRequest{
			SubjectID:   "fuzz-subject",
			PortfolioID: "fuzz-portfolio",
			SourceKind:  SourceKindUserUploadedFile,
			Reader:      strings.NewReader(payload),
		})
		if err != nil {
			return
		}

		if len(review.Rows) > MaxReviewRows {
			t.Fatalf("parser admitted %d rows, max is %d", len(review.Rows), MaxReviewRows)
		}
		if review.Summary.TotalRows != len(review.Rows) {
			t.Fatalf("summary total mismatch: total=%d rows=%d", review.Summary.TotalRows, len(review.Rows))
		}

		classified := review.Summary.AppendableRows +
			review.Summary.DuplicateRows +
			review.Summary.ConflictRows +
			review.Summary.InvalidRows
		if classified != review.Summary.TotalRows {
			t.Fatalf("summary classification mismatch: classified=%d total=%d", classified, review.Summary.TotalRows)
		}
	})
}
