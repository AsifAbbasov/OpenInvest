package importer

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func FuzzM11CSVReviewSemantics(f *testing.F) {
	seeds := []string{
		"transaction_type,ticker,quantity,unit_price,gross_amount,commission,tax,trade_date,settlement_date,currency,broker_operation_id,note\nBUY,SBER,2.00000000,100.00000000,200.00000000,0.00000000,0.00000000,2026-01-01,,RUB,op-1,seed\n",
		"transaction_type,ticker,quantity,unit_price,gross_amount,commission,tax,trade_date,settlement_date,currency,broker_operation_id,note\nDEPOSIT,,,,1000.00000000,0.00000000,0.00000000,2026-01-01,,RUB,op-2,seed\n",
		"transaction_type,ticker,quantity,unit_price,gross_amount,commission,tax,trade_date,settlement_date,currency,broker_operation_id,note\n",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, payload string) {
		if len(payload) > 16*1024 {
			t.Skip()
		}
		sum := sha256.Sum256([]byte(payload))
		request := ReviewRequest{
			SubjectID:          "00000000-0000-4000-8000-000000000011",
			PortfolioID:        "00000000-0000-4000-8000-000000000012",
			SourceKind:         SourceKindUserUploadedFile,
			SourceAccountLabel: "m11-fuzz",
			FileHash:           hex.EncodeToString(sum[:]),
			Reader:             strings.NewReader(payload),
		}
		review, err := ReviewCSV(request)
		if err != nil {
			return
		}
		first, err := ReviewSemanticDigest(review)
		if err != nil {
			t.Fatalf("first semantic digest: %v", err)
		}
		second, err := ReviewSemanticDigest(review)
		if err != nil {
			t.Fatalf("second semantic digest: %v", err)
		}
		if first != second {
			t.Fatalf("semantic digest is nondeterministic: %s != %s", first, second)
		}
		identities := make([]DecisionIdentity, 0, len(review.Rows))
		for _, row := range review.Rows {
			identities = append(identities, DecisionIdentity{RowNumber: row.RowNumber, RowHash: row.RowHash})
		}
		if err := VerifyDecisionIdentities(review, identities); err != nil {
			t.Fatalf("self-generated review identities rejected: %v", err)
		}
	})
}
