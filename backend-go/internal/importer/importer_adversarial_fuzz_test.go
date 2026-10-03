package importer

import (
	"errors"
	"strconv"
	"strings"
	"testing"
)

const fuzzCanonicalHeader = "transaction_type,ticker,quantity,unit_price,gross_amount,commission,tax,trade_date,settlement_date,currency,broker_operation_id,note\n"

func FuzzReviewCSVStructuralBounds(f *testing.F) {
	f.Add(uint16(0), uint16(0), uint16(0))
	f.Add(uint16(MaxReviewRows), uint16(MaxCSVHeaderColumns), uint16(MaxCSVHeaderFieldBytes))
	f.Add(uint16(MaxReviewRows+1), uint16(MaxCSVHeaderColumns+1), uint16(MaxCSVHeaderFieldBytes+1))

	f.Fuzz(func(t *testing.T, rowsRaw, columnsRaw, fieldBytesRaw uint16) {
		rows := int(rowsRaw % 160)
		columns := int(columnsRaw % 32)
		fieldBytes := int(fieldBytesRaw % 512)

		if columns > MaxCSVHeaderColumns || fieldBytes > MaxCSVHeaderFieldBytes {
			header := make([]string, columns)
			if len(header) == 0 {
				header = []string{strings.Repeat("h", fieldBytes)}
			} else {
				for i := range header {
					header[i] = "h" + strconv.Itoa(i)
				}
				header[0] = strings.Repeat("h", fieldBytes)
			}
			err := validateCSVHeaderShape(header)
			if err == nil {
				t.Fatalf("header structural bound bypass: columns=%d bytes=%d", columns, fieldBytes)
			}
			if !errors.Is(err, ErrInvalidImport) {
				t.Fatalf("header bound returned wrong error: %v", err)
			}
		}

		var b strings.Builder
		b.WriteString(fuzzCanonicalHeader)
		for i := 0; i < rows; i++ {
			b.WriteString("BUY,SBER,1.00000000,100.00000000,100.00000000,0.00000000,0.00000000,2026-01-01,2026-01-02,RUB,op-")
			b.WriteString(strconv.Itoa(i))
			b.WriteString(",note\n")
		}
		review, err := ReviewCSV(ReviewRequest{
			SubjectID:          "fuzz-subject",
			PortfolioID:        "fuzz-portfolio",
			SourceKind:         SourceKindUserUploadedFile,
			SourceAccountLabel: "fuzz-account",
			FileHash:           strings.Repeat("a", 64),
			Reader:             strings.NewReader(b.String()),
		})

		if rows > MaxReviewRows {
			if err == nil || !errors.Is(err, ErrInvalidImport) {
				t.Fatalf("row limit bypass: rows=%d err=%v", rows, err)
			}
			return
		}
		if err != nil {
			t.Fatalf("bounded canonical CSV rejected: rows=%d err=%v", rows, err)
		}
		if len(review.Rows) != rows {
			t.Fatalf("bounded CSV row count drift: got=%d want=%d", len(review.Rows), rows)
		}
	})
}
