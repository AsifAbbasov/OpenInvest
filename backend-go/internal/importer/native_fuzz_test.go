package importer

// Provenance: PR #229 at 657c34f813ac7faca1a9aa567b664dee5bc01be8; backend-go/internal/importer/importer_fuzz_test.go.
// Original source and seed evidence remain retained in Issue #280.
import (
	"reflect"
	"strings"
	"testing"
)

func FuzzReviewCSV(f *testing.F) {
	valid := "transaction_type,ticker,quantity,unit_price,gross_amount,commission,tax,trade_date,settlement_date,currency,broker_operation_id,note\n" +
		"BUY,SBER,1.00000000,100.00000000,100.00000000,0.00000000,0.00000000,2026-01-01,2026-01-02,RUB,op-1,note\n"
	f.Add([]byte(valid))
	f.Add([]byte(fuzzCanonicalHeader + "DEPOSIT,,,,100.00000000,0.00000000,0.00000000,2026-08-27,,RUB,deposit-1,note\n"))
	f.Add([]byte(fuzzCanonicalHeader + "\"unterminated"))
	f.Add([]byte(""))
	f.Add([]byte("a,b,c\n1,2,3\n"))
	f.Add([]byte("transaction_type,ticker,quantity,unit_price,gross_amount,commission,tax,trade_date,settlement_date,currency,broker_operation_id,note\n"))

	// Exact #225–228 bytes; camelCase is retained as historical invalid-header input.
	legacyHeader := "transactionType,ticker,quantity,unitPrice,grossAmount,commission,tax,tradeDate,settlementDate,note,brokerOperationId,currency\n"
	legacyValid := legacyHeader + "BUY,SBER,1.00000000,100.00000000,100.00000000,0.00000000,0.00000000,2026-01-10,2026-01-12,test,op-1,RUB\n"
	for _, seed := range []string{
		fuzzCanonicalHeader + "DEPOSIT,,,,1000.00000000,0.00000000,0.00000000,2026-06-19,,RUB,op-1,cash in\n",
		fuzzCanonicalHeader + "BUY,SBER,2.00000000,100.00000000,200.00000000,1.00000000,0.00000000,2026-06-20,2026-06-21,RUB,op-2,buy\n",
		"not,csv\n", "\x00\xff\n",
		fuzzCanonicalHeader + "BUY,SBER,2.00000000,100.00000000,200.00000000,1.00000000,0.00000000,2026-01-10,2026-01-13,RUB,broker-row-1,Imported buy\n",
		legacyValid, legacyValid + legacyValid,
		legacyHeader + "BUY,SBER,not-a-decimal,100,100,0,0,2026-99-99,,,op-1,RUB\n",
		fuzzCanonicalHeader + "BUY,SBER,2.00000000,100.00000000,200.00000000,1.00000000,0.00000000,2026-06-20,,RUB,op-2,buy\n",
		"transaction_type\nBUY\n", "\x00\xff\xfe",
	} {
		f.Add([]byte(seed))
	}
	if _, err := ReviewCSV(ReviewRequest{SubjectID: "fuzz-subject", PortfolioID: "fuzz-portfolio", Reader: strings.NewReader(legacyValid)}); err == nil {
		f.Fatal("obsolete camelCase fixture unexpectedly accepted as current CSV")
	}
	positive, err := ReviewCSV(ReviewRequest{SubjectID: "fuzz-subject", PortfolioID: "fuzz-portfolio", Reader: strings.NewReader(valid)})
	if err != nil || len(positive.Rows) != 1 || positive.Rows[0].Status != ReviewStatusAppendable {
		f.Fatalf("canonical BUY positive control: review=%+v err=%v", positive, err)
	}

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
		second, secondErr := ReviewCSV(ReviewRequest{
			SubjectID:          "fuzz-subject",
			PortfolioID:        "fuzz-portfolio",
			SourceKind:         SourceKindUserUploadedFile,
			SourceAccountLabel: "fuzz-account",
			FileHash:           strings.Repeat("0", 64),
			Reader:             strings.NewReader(string(data)),
		})
		if (err == nil) != (secondErr == nil) {
			t.Fatal("CSV success/error nondeterminism")
		}
		if err != nil {
			if err.Error() != secondErr.Error() {
				t.Fatal("CSV error-message nondeterminism")
			}
			return
		}
		if !reflect.DeepEqual(review.Summary, second.Summary) {
			t.Fatal("CSV classification nondeterminism")
		}
		for i, row := range review.Rows {
			if row.RowNumber < 2 {
				t.Fatalf("invalid CSV RowNumber %d", row.RowNumber)
			}
			switch row.Status {
			case ReviewStatusAppendable, ReviewStatusDuplicate, ReviewStatusConflict, ReviewStatusInvalid:
			default:
				t.Fatalf("unknown row status %q", row.Status)
			}
			if i >= len(second.Rows) || row.Status != second.Rows[i].Status {
				t.Fatal("CSV row classification nondeterminism")
			}
		}
		firstDigest, e1 := ReviewSemanticDigest(review)
		secondDigest, e2 := ReviewSemanticDigest(second)
		if e1 != nil || e2 != nil || firstDigest != secondDigest {
			t.Fatal("CSV semantic digest nondeterminism")
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
