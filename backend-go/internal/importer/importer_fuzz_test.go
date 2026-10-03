package importer

import (
	"strings"
	"testing"
)

func FuzzReviewCSV(f *testing.F) {
	for _, seed := range []string{
		csvHeader + "DEPOSIT,,,,1000.00000000,0.00000000,0.00000000,2026-06-19,,RUB,op-1,cash in\n",
		csvHeader + "BUY,SBER,2.00000000,100.00000000,200.00000000,1.00000000,0.00000000,2026-06-20,2026-06-21,RUB,op-2,buy\n",
		csvHeader,
		"not,csv\n",
		"\x00\xff\n",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, payload string) {
		if len(payload) > 256*1024 {
			t.Skip()
		}

		review, err := ReviewCSV(ReviewRequest{
			SubjectID:          "fuzz-subject",
			PortfolioID:        "fuzz-portfolio",
			SourceKind:         SourceKindUserUploadedFile,
			SourceAccountLabel: "fuzz",
			FileHash:           "fuzz-hash",
			Reader:             strings.NewReader(payload),
		})
		if err != nil {
			return
		}

		if review.Summary.TotalRows != len(review.Rows) {
			t.Fatalf("summary total does not match row count: summary=%+v rows=%d", review.Summary, len(review.Rows))
		}
		classified := review.Summary.AppendableRows +
			review.Summary.DuplicateRows +
			review.Summary.ConflictRows +
			review.Summary.InvalidRows
		if classified != review.Summary.TotalRows {
			t.Fatalf("summary classifications do not add up: summary=%+v", review.Summary)
		}
		if review.Summary.TotalRows > MaxReviewRows {
			t.Fatalf("successful review exceeded MaxReviewRows: got=%d max=%d", review.Summary.TotalRows, MaxReviewRows)
		}

		for _, row := range review.Rows {
			switch row.Status {
			case ReviewStatusAppendable, ReviewStatusDuplicate, ReviewStatusConflict, ReviewStatusInvalid:
			default:
				t.Fatalf("unknown review status %q at row %d", row.Status, row.RowNumber)
			}
			if row.RowNumber < 2 {
				t.Fatalf("invalid CSV data row number: %d", row.RowNumber)
			}
		}
	})
}
