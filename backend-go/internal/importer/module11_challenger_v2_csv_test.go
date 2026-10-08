package importer

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestM11V2CSVAdversarialCampaign(t *testing.T) {
	header := "transaction_type,ticker,quantity,unit_price,gross_amount,commission,tax,trade_date,settlement_date,currency,broker_operation_id,note\n"
	cases := []struct {
		name string
		payload string
		mustReject bool
	}{
		{"valid buy", header+"BUY,SBER,2.00000000,100.00000000,200.00000000,0.00000000,0.00000000,2026-01-01,,RUB,op-1,note\n", false},
		{"reordered rows", header+"DEPOSIT,,,,1000.00000000,0.00000000,0.00000000,2026-01-02,,RUB,op-2,note\nBUY,SBER,1.00000000,100.00000000,100.00000000,0.00000000,0.00000000,2026-01-01,,RUB,op-1,note\n", false},
		{"duplicate rows", header+"BUY,SBER,1.00000000,100.00000000,100.00000000,0.00000000,0.00000000,2026-01-01,,RUB,dup-1,note\nBUY,SBER,1.00000000,100.00000000,100.00000000,0.00000000,0.00000000,2026-01-01,,RUB,dup-1,note\n", false},
		{"same identity changed amount", header+"DEPOSIT,,,,100.00000000,0.00000000,0.00000000,2026-01-01,,RUB,same-1,note\nDEPOSIT,,,,101.00000000,0.00000000,0.00000000,2026-01-01,,RUB,same-1,note\n", false},
		{"empty optionals", header+"DEPOSIT,,,,10.00000000,0.00000000,0.00000000,2026-01-01,,RUB,,\n", false},
		{"quoted separator", header+"DEPOSIT,,,,10.00000000,0.00000000,0.00000000,2026-01-01,,RUB,quote-1,\"hello,world\"\n", false},
		{"quoted newline", header+"DEPOSIT,,,,10.00000000,0.00000000,0.00000000,2026-01-01,,RUB,quote-2,\"hello\nworld\"\n", false},
		{"unicode note", header+"DEPOSIT,,,,10.00000000,0.00000000,0.00000000,2026-01-01,,RUB,unicode-1,Привет世界\n", false},
		{"extreme valid decimal", header+"DEPOSIT,,,,99999999999999999999.99999999,0.00000000,0.00000000,2026-01-01,,RUB,max-1,note\n", false},
		{"changed date", header+"DEPOSIT,,,,10.00000000,0.00000000,0.00000000,2099-12-31,,RUB,date-1,note\n", false},
		{"malformed width", header+"BUY,SBER,1.00000000\n", true},
		{"duplicate header", "transaction_type,ticker,ticker,quantity,unit_price,gross_amount,commission,tax,trade_date,settlement_date,currency,broker_operation_id,note\nBUY,SBER,SBER,1.00000000,100.00000000,100.00000000,0.00000000,0.00000000,2026-01-01,,RUB,h-1,note\n", true},
		{"unknown header", strings.Replace(header, "note", "unknown_column", 1)+"DEPOSIT,,,,10.00000000,0.00000000,0.00000000,2026-01-01,,RUB,u-1,note\n", true},
		{"invalid decimal", header+"DEPOSIT,,,,1e100,0.00000000,0.00000000,2026-01-01,,RUB,bad-dec-1,note\n", true},
	}
	accepted, rejected := 0, 0
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sum := sha256.Sum256([]byte(tc.payload))
			review, err := ReviewCSV(ReviewRequest{
				SubjectID: "00000000-0000-4000-8000-000000000011",
				PortfolioID: "00000000-0000-4000-8000-000000000012",
				SourceKind: SourceKindUserUploadedFile,
				SourceAccountLabel: "m11-v2",
				FileHash: hex.EncodeToString(sum[:]),
				Reader: strings.NewReader(tc.payload),
			})
			if err != nil {
				rejected++
				if !tc.mustReject {
					t.Logf("M11_V2_CSV_CASE=%s RESULT=REJECTED ERR=%v", tc.name, err)
				}
				return
			}
			if tc.mustReject {
				t.Fatalf("case %s unexpectedly accepted", tc.name)
			}
			accepted++
			first, err := ReviewSemanticDigest(review)
			if err != nil { t.Fatal(err) }
			second, err := ReviewSemanticDigest(review)
			if err != nil { t.Fatal(err) }
			if first != second { t.Fatalf("semantic digest drift %s != %s", first, second) }
			ids := make([]DecisionIdentity, 0, len(review.Rows))
			for _, row := range review.Rows { ids = append(ids, DecisionIdentity{RowNumber:row.RowNumber, RowHash:row.RowHash}) }
			if err := VerifyDecisionIdentities(review, ids); err != nil { t.Fatalf("self identities rejected: %v", err) }
			t.Logf("M11_V2_CSV_CASE=%s RESULT=ACCEPTED ROWS=%d", tc.name, len(review.Rows))
		})
	}
	t.Logf("M11_V2_CSV_CAMPAIGN_CASES=%d", len(cases))
	t.Logf("M11_V2_CSV_CAMPAIGN_ACCEPTED=%d", accepted)
	t.Logf("M11_V2_CSV_CAMPAIGN_REJECTED=%d", rejected)
}
