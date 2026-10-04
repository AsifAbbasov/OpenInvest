package httpapi

import (
	"strings"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func FuzzPaginationCursorIsolationAndTamperResistance(f *testing.F) {
	f.Add("subject-a", "subject-b", "portfolio-a", "2026-01-01", "entry-1", uint8(0))
	f.Add("subject-a", "subject-a", "portfolio-a", "2026-12-31", "entry-2", uint8(1))

	f.Fuzz(func(t *testing.T, subject, otherSubject, portfolio, tradeDate, entryID string, mutation uint8) {
		if len(subject) == 0 || len(subject) > 128 ||
			len(otherSubject) == 0 || len(otherSubject) > 128 ||
			len(portfolio) == 0 || len(portfolio) > 128 ||
			len(tradeDate) > 32 || len(entryID) == 0 || len(entryID) > 128 {
			return
		}
		if _, err := time.Parse("2006-01-02", tradeDate); err != nil {
			return
		}

		api := &API{paginationCursorSecret: []byte("pagination-adversarial-test-secret")}
		filter := verticalslice.TransactionFilter{
			TransactionType: "BUY",
			FromDate:        "2026-01-01",
			ToDate:          "2026-12-31",
		}
		raw, err := api.signPaginationCursor(paginationCursorPayload{
			Version:         1,
			Resource:        "transactions",
			SubjectID:       subject,
			PortfolioID:     portfolio,
			TransactionType: filter.TransactionType,
			FromDate:        filter.FromDate,
			ToDate:          filter.ToDate,
			TradeDate:       tradeDate,
			EntryID:         entryID,
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(raw) > maxPaginationCursorBytes {
			t.Fatalf("server generated a cursor larger than its own verification ceiling: %d", len(raw))
		}

		applied := filter
		if err := api.applyTransactionCursor(raw, subject, portfolio, &applied); err != nil {
			t.Fatalf("valid signed cursor rejected: %v", err)
		}
		if applied.BeforeTradeDate != tradeDate || applied.BeforeEntryID != entryID {
			t.Fatalf("cursor anchor drift: %+v", applied)
		}

		if otherSubject != subject {
			crossSubject := filter
			if err := api.applyTransactionCursor(raw, otherSubject, portfolio, &crossSubject); err == nil {
				t.Fatal("cursor replay across principals was accepted")
			}
		}

		crossPortfolio := filter
		if err := api.applyTransactionCursor(raw, subject, portfolio+"-other", &crossPortfolio); err == nil {
			t.Fatal("cursor replay across portfolios was accepted")
		}

		tampered := []byte(raw)
		if len(tampered) > 0 {
			index := int(mutation) % len(tampered)
			if tampered[index] == 'A' {
				tampered[index] = 'B'
			} else {
				tampered[index] = 'A'
			}
			if _, err := api.verifyPaginationCursor(string(tampered)); err == nil {
				t.Fatal("tampered cursor passed HMAC verification")
			}
		}

		oversized := strings.Repeat("A", maxPaginationCursorBytes+1)
		if _, err := api.verifyPaginationCursor(oversized); err == nil {
			t.Fatal("oversized cursor bypassed the hard length ceiling")
		}
	})
}
