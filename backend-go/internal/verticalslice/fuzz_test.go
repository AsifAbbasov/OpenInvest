package verticalslice

import (
	"fmt"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
)

func FuzzParseBusinessDate(f *testing.F) {
	for _, seed := range []string{
		"2026-01-01",
		"2024-02-29",
		"2023-02-29",
		"2026-13-01",
		"2026-01-32",
		" 2026-01-01",
		"2026-01-01 ",
		"2026-1-1",
		"",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		parsed, err := parseBusinessDate(input)
		if err != nil {
			return
		}

		if len(input) != len("2006-01-02") {
			t.Fatalf("parseBusinessDate accepted non-canonical length: %q", input)
		}
		if parsed.Format("2006-01-02") != input {
			t.Fatalf("parseBusinessDate accepted non-canonical date: input=%q parsed=%s", input, parsed.Format("2006-01-02"))
		}
		if parsed.Location() != time.UTC {
			t.Fatalf("business date did not parse in UTC: input=%q location=%s", input, parsed.Location())
		}
	})
}

func FuzzPortfolioXIRR(f *testing.F) {
	f.Add(int64(-1000), int64(100), int64(1200), uint16(0), uint16(180))
	f.Add(int64(-1000), int64(-500), int64(2000), uint16(0), uint16(365))
	f.Add(int64(0), int64(0), int64(1), uint16(0), uint16(1))
	f.Add(int64(1000), int64(1000), int64(1000), uint16(10), uint16(20))

	f.Fuzz(func(t *testing.T, firstRaw, secondRaw, terminalRaw int64, dayOneRaw, dayTwoRaw uint16) {
		first := decimal.Must(fmt.Sprintf("%d.00000000", firstRaw%1_000_000_000))
		second := decimal.Must(fmt.Sprintf("%d.00000000", secondRaw%1_000_000_000))

		terminalUnits := int64(uint64(terminalRaw)%1_000_000_000) + 1
		terminal := Money{
			Amount:   decimal.Must(fmt.Sprintf("%d.00000000", terminalUnits)),
			Currency: RUB,
		}

		base := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		dayOne := int(dayOneRaw % 3000)
		dayTwo := int(dayTwoRaw % 3000)
		asOf := base.AddDate(0, 0, 3001).Format("2006-01-02")
		flows := []PortfolioReturnCashFlow{
			{Date: base.AddDate(0, 0, dayOne).Format("2006-01-02"), Amount: first},
			{Date: base.AddDate(0, 0, dayTwo).Format("2006-01-02"), Amount: second},
		}

		firstProjection, firstErr := BuildPortfolioReturnProjection("fuzz-portfolio", asOf, flows, &terminal, true)
		secondProjection, secondErr := BuildPortfolioReturnProjection("fuzz-portfolio", asOf, flows, &terminal, true)

		if (firstErr == nil) != (secondErr == nil) {
			t.Fatalf("XIRR projection is non-deterministic: firstErr=%v secondErr=%v", firstErr, secondErr)
		}
		if firstErr != nil {
			if firstErr.Error() != secondErr.Error() {
				t.Fatalf("XIRR error changed across identical input: first=%q second=%q", firstErr.Error(), secondErr.Error())
			}
			return
		}

		if firstProjection.Status != secondProjection.Status || firstProjection.Reason != secondProjection.Reason {
			t.Fatalf("XIRR status changed across identical input: first=%s/%s second=%s/%s",
				firstProjection.Status, firstProjection.Reason, secondProjection.Status, secondProjection.Reason)
		}

		switch firstProjection.Status {
		case PortfolioReturnAvailableStatus:
			if firstProjection.XIRR == nil {
				t.Fatalf("AVAILABLE projection has nil XIRR")
			}
			if !firstProjection.XIRR.FitsStorage() {
				t.Fatalf("AVAILABLE XIRR does not fit canonical Decimal storage: %s", firstProjection.XIRR.String())
			}
			if firstProjection.Reason != "" {
				t.Fatalf("AVAILABLE projection unexpectedly has reason %q", firstProjection.Reason)
			}
			if secondProjection.XIRR == nil || firstProjection.XIRR.String() != secondProjection.XIRR.String() {
				t.Fatalf("XIRR value changed across identical input")
			}
		case PortfolioReturnUnavailableStatus:
			if firstProjection.XIRR != nil {
				t.Fatalf("UNAVAILABLE projection unexpectedly exposes XIRR=%s", firstProjection.XIRR.String())
			}
			if firstProjection.Reason == "" {
				t.Fatalf("UNAVAILABLE projection has no reason")
			}
		default:
			t.Fatalf("unexpected XIRR status %q", firstProjection.Status)
		}
	})
}
