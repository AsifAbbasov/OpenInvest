package importer

import (
	"strings"
	"testing"
)

func FuzzReviewCSV(f *testing.F) {
	valid := "transactionType,ticker,quantity,unitPrice,grossAmount,commission,tax,tradeDate,settlementDate,note,brokerOperationId,currency\n" +
		"BUY,SBER,1.00000000,100.00000000,100.00000000,0.00000000,0.00000000,2026-01-10,2026-01-12,test,op-1,RUB\n"

	for _, seed := range []string{
		valid,
		"",
		"a,b,c\n1,2,3\n",
		valid + valid,
		"transactionType,ticker,quantity,unitPrice,grossAmount,commission,tax,tradeDate,settlementDate,note,brokerOperationId,currency\n" +
			"BUY,SBER,not-a-decimal,100,100,0,0,2026-99-99,,,op-1,RUB\n",
	} {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		// Parser fuzzing and bulk/resource testing are separate concerns. Keep the
		// coverage-guided corpus bounded so one oversized mutation cannot dominate
		// the fuzz worker.
		if len(data) > 256*1024 {
			t.Skip()
		}

		request := ReviewRequest{
			SubjectID:          "00000000-0000-4000-8000-000000000001",
			PortfolioID:        "00000000-0000-4000-8000-000000000002",
			SourceKind:         SourceKindUserUploadedFile,
			SourceAccountLabel: "fuzz",
			Reader:             strings.NewReader(string(data)),
		}

		first, firstErr := ReviewCSV(request)

		request.Reader = strings.NewReader(string(data))
		second, secondErr := ReviewCSV(request)

		if (firstErr == nil) != (secondErr == nil) {
			t.Fatalf("ReviewCSV is non-deterministic for identical input: firstErr=%v secondErr=%v", firstErr, secondErr)
		}
		if firstErr != nil {
			if firstErr.Error() != secondErr.Error() {
				t.Fatalf("ReviewCSV error changed for identical input: first=%q second=%q", firstErr.Error(), secondErr.Error())
			}
			return
		}

		if len(first.Rows) > MaxReviewRows {
			t.Fatalf("ReviewCSV exceeded MaxReviewRows: got=%d max=%d", len(first.Rows), MaxReviewRows)
		}
		if first.Summary.TotalRows != len(first.Rows) {
			t.Fatalf("summary/row mismatch: summary=%d rows=%d", first.Summary.TotalRows, len(first.Rows))
		}

		firstDigest, err := ReviewSemanticDigest(first)
		if err != nil {
			t.Fatalf("digest successful review: %v", err)
		}
		secondDigest, err := ReviewSemanticDigest(second)
		if err != nil {
			t.Fatalf("digest repeated successful review: %v", err)
		}
		if firstDigest != secondDigest {
			t.Fatalf("ReviewCSV semantic digest is non-deterministic: first=%s second=%s", firstDigest, secondDigest)
		}
	})
}
