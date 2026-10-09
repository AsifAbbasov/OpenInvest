package httpapi

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func FuzzPaginationCursorIsolationAndTamperResistance(f *testing.F) {
	f.Add([]byte("subject-a"), []byte("subject-b"), []byte("portfolio-a"), uint16(0), uint8(0))
	f.Add([]byte("subject-a"), []byte("subject-a"), []byte("portfolio-a"), uint16(364), uint8(1))

	f.Fuzz(func(t *testing.T, subjectSeed, otherSubjectSeed, portfolioSeed []byte, dayRaw uint16, mutation uint8) {
		if len(subjectSeed) > 1024 || len(otherSubjectSeed) > 1024 || len(portfolioSeed) > 1024 {
			return
		}

		// Production subjects, portfolios and transaction entry anchors are UUIDs.
		// Derive canonical UUIDs from fuzz bytes so the fuzzer explores reachable
		// cursor states instead of manufacturing server-signed cursors that the
		// production application could never emit.
		subject := uuid.NewSHA1(uuid.Nil, append([]byte("subject:"), subjectSeed...)).String()
		otherSubject := uuid.NewSHA1(uuid.Nil, append([]byte("other-subject:"), otherSubjectSeed...)).String()
		portfolio := uuid.NewSHA1(uuid.Nil, append([]byte("portfolio:"), portfolioSeed...)).String()
		entryID := uuid.NewSHA1(uuid.Nil, append([]byte("entry:"), portfolioSeed...)).String()

		base, _ := time.Parse("2006-01-02", "2026-01-01")
		tradeDate := base.AddDate(0, 0, int(dayRaw%365)).Format("2006-01-02")

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
			t.Fatalf("production-reachable server cursor exceeds verification ceiling: %d", len(raw))
		}

		applied := filter
		if err := api.applyTransactionCursor(raw, subject, portfolio, &applied); err != nil {
			t.Fatalf("valid production-reachable signed cursor rejected: %v", err)
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

		otherPortfolio := uuid.NewSHA1(uuid.Nil, append([]byte("other-portfolio:"), portfolioSeed...)).String()
		crossPortfolio := filter
		if err := api.applyTransactionCursor(raw, subject, otherPortfolio, &crossPortfolio); err == nil {
			t.Fatal("cursor replay across portfolios was accepted")
		}

		changedFilter := filter
		changedFilter.TransactionType = "SELL"
		if err := api.applyTransactionCursor(raw, subject, portfolio, &changedFilter); err == nil {
			t.Fatal("cursor replay across changed filters was accepted")
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
