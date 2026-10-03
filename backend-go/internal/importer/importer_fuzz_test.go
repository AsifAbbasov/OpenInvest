package importer

import (
	"strings"
	"testing"
)

func FuzzReviewCSV(f *testing.F) {
	valid := "transaction_type,ticker,quantity,unit_price,gross_amount,commission,tax,trade_date,settlement_date,currency,broker_operation_id,note\n" +
		"BUY,SBER,1.00000000,100.00000000,100.00000000,0.00000000,0.00000000,2026-01-01,2026-01-02,RUB,op-1,note\n"
	f.Add([]byte(valid))
	f.Add([]byte(""))
	f.Add([]byte("a,b,c\n1,2,3\n"))
	f.Add([]byte("transaction_type,ticker,quantity,unit_price,gross_amount,commission,tax,trade_date,settlement_date,currency,broker_operation_id,note\n"))

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1<<20 {
			t.Skip()
		}
		review, err := ReviewCSV(ReviewRequest{
			SubjectID:          "fuzz-subject",
			PortfolioID:        "fuzz-portfolio",
			SourceKind:         SourceKindUserUploadedFile,
			SourceAccountLabel: "fuzz-account",
			FileHash:           strings.Repeat("0", 64),
			Reader:             strings.NewReader(string(data)),
		})
		if err != nil {
			return
		}
		if len(review.Rows) > MaxReviewRows {
			t.Fatalf("successful review exceeded MaxReviewRows: got=%d max=%d", len(review.Rows), MaxReviewRows)
		}
		if review.Summary.TotalRows != len(review.Rows) {
			t.Fatalf("review summary drift: totalRows=%d rows=%d", review.Summary.TotalRows, len(review.Rows))
		}
		classified := review.Summary.AppendableRows +
			review.Summary.DuplicateRows +
			review.Summary.ConflictRows +
			review.Summary.InvalidRows
		if classified != review.Summary.TotalRows {
			t.Fatalf("review classification drift: classified=%d total=%d", classified, review.Summary.TotalRows)
		}
	})
}

func FuzzValidateCSVHeaderShape(f *testing.F) {
	for _, seed := range []string{
		"transaction_type,ticker,quantity",
		"",
		"a,b,c,d,e,f,g,h,i,j,k,l",
		"a,b,c,d,e,f,g,h,i,j,k,l,m",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, raw string) {
		if len(raw) > 1<<20 {
			t.Skip()
		}
		header := strings.Split(raw, ",")
		err := validateCSVHeaderShape(header)
		if err != nil {
			return
		}
		if len(header) > MaxCSVHeaderColumns {
			t.Fatalf("validator accepted %d columns; max=%d", len(header), MaxCSVHeaderColumns)
		}
		for index, field := range header {
			if len(field) > MaxCSVHeaderFieldBytes {
				t.Fatalf("validator accepted oversized header field %d: bytes=%d max=%d", index, len(field), MaxCSVHeaderFieldBytes)
			}
		}
	})
}
