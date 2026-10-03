package importer

import (
	"bytes"
	"testing"
)

func FuzzReviewCSV(f *testing.F) {
	seeds := [][]byte{
		[]byte(csvHeader + "DEPOSIT,,,,1000.00000000,0.00000000,0.00000000,2026-06-19,,RUB,op-1,cash in\n"),
		[]byte(csvHeader + "BUY,SBER,2.00000000,100.00000000,200.00000000,1.00000000,0.00000000,2026-06-20,,RUB,op-2,buy\n"),
		[]byte(""),
		[]byte("a,b,c\n1,2,3\n"),
		[]byte{0xff, 0xfe, 0xfd},
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, raw []byte) {
		review, err := ReviewCSV(ReviewRequest{
			SubjectID:  "fuzz-subject",
			PortfolioID: "fuzz-portfolio",
			Reader:      bytes.NewReader(raw),
		})
		if err != nil {
			return
		}

		if len(review.Rows) > MaxReviewRows {
			t.Fatalf("review escaped row bound: got=%d max=%d", len(review.Rows), MaxReviewRows)
		}
		total := review.Summary.AppendableRows +
			review.Summary.DuplicateRows +
			review.Summary.ConflictRows +
			review.Summary.InvalidRows
		if total != review.Summary.TotalRows || total != len(review.Rows) {
			t.Fatalf("inconsistent review summary: summary=%+v rows=%d", review.Summary, len(review.Rows))
		}

		decisions := make([]Decision, 0, len(review.Rows))
		for _, row := range review.Rows {
			switch row.Status {
			case ReviewStatusAppendable:
				decisions = append(decisions, Decision{RowNumber: row.RowNumber, RowHash: row.RowHash, Action: DecisionApprove})
			case ReviewStatusDuplicate, ReviewStatusConflict, ReviewStatusInvalid:
				decisions = append(decisions, Decision{RowNumber: row.RowNumber, RowHash: row.RowHash, Action: DecisionIgnore})
			default:
				t.Fatalf("unknown review status %q", row.Status)
			}
		}
		if _, err := BuildAppendRequests(review, decisions); err != nil {
			t.Fatalf("safe decisions derived from successful review must build: %v", err)
		}
	})
}
