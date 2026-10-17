package httpapi

import (
	"strings"
	"testing"
)

func FuzzPaginationCursorRoundTrip(f *testing.F) {
	f.Add("subject", "portfolio", "entry", "2026-06-20")
	f.Add("", "", "", "")
	f.Add("пароль", "portfolio/../x", "entry\x00id", "9999-99-99")

	f.Fuzz(func(t *testing.T, subjectID, portfolioID, entryID, tradeDate string) {
		api := &API{paginationCursorSecret: []byte("0123456789abcdef0123456789abcdef")}
		payload := paginationCursorPayload{
			Version:     1,
			Resource:    "transactions",
			SubjectID:   subjectID,
			PortfolioID: portfolioID,
			TradeDate:   tradeDate,
			EntryID:     entryID,
		}

		token, err := api.signPaginationCursor(payload)
		if err != nil {
			t.Fatalf("sign cursor: %v", err)
		}
		decoded, err := api.verifyPaginationCursor(token)
		if err != nil {
			t.Fatalf("verify signed cursor: %v", err)
		}
		if decoded != payload {
			t.Fatalf("cursor round trip mismatch: want=%+v got=%+v", payload, decoded)
		}

		parts := strings.Split(token, ".")
		if len(parts) != 2 || len(parts[1]) == 0 {
			t.Fatalf("signer emitted malformed cursor")
		}
		first := parts[1][0]
		replacement := byte('A')
		if first == replacement {
			replacement = 'B'
		}
		tampered := parts[0] + "." + string(replacement) + parts[1][1:]
		if _, err := api.verifyPaginationCursor(tampered); err == nil {
			t.Fatalf("tampered cursor signature was accepted")
		}
	})
}
