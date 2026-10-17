package verticalslice

import (
	"fmt"
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
)

func FuzzParseBusinessDate(f *testing.F) {
	for _, seed := range []string{
		"2026-01-01",
		"2000-02-29",
		"1900-02-29",
		"2026-13-01",
		"2026-01-32",
		" 2026-01-01 ",
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
		if got := parsed.Format("2006-01-02"); got != input {
			t.Fatalf("accepted BusinessDate is not canonical: input=%q formatted=%q", input, got)
		}
		roundTrip, err := parseBusinessDate(parsed.Format("2006-01-02"))
		if err != nil {
			t.Fatalf("canonical BusinessDate failed round-trip: input=%q err=%v", input, err)
		}
		if !roundTrip.Equal(parsed) {
			t.Fatalf("BusinessDate round-trip changed value: input=%q first=%v second=%v", input, parsed, roundTrip)
		}
	})
}

func FuzzXIRRTwoTermProjection(f *testing.F) {
	f.Add(uint64(100), uint64(110), uint16(365))
	f.Add(uint64(100), uint64(80), uint16(365))
	f.Add(uint64(1), uint64(1), uint16(1))
	f.Add(uint64(1_000_000), uint64(2_000_000), uint16(730))

	f.Fuzz(func(t *testing.T, principalRaw uint64, terminalRaw uint64, daysRaw uint16) {
		principal := principalRaw%1_000_000_000 + 1
		terminal := terminalRaw%1_000_000_000 + 1
		days := int(daysRaw%7300) + 1

		start := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		asOf := start.AddDate(0, 0, days).Format("2006-01-02")

		principalDecimal, err := decimal.FromString(fmt.Sprintf("-%d.00000000", principal))
		if err != nil {
			t.Fatalf("fuzz harness generated invalid principal: %v", err)
		}
		terminalDecimal, err := decimal.FromString(fmt.Sprintf("%d.00000000", terminal))
		if err != nil {
			t.Fatalf("fuzz harness generated invalid terminal value: %v", err)
		}
		terminalMoney := Money{Amount: terminalDecimal, Currency: RUB}
		flows := []PortfolioReturnCashFlow{{
			Date:   start.Format("2006-01-02"),
			Amount: principalDecimal,
		}}

		first, err := BuildPortfolioReturnProjection("fuzz-portfolio", asOf, flows, &terminalMoney, true)
		if err != nil {
			t.Fatalf("valid two-term XIRR input returned error: %v", err)
		}
		second, err := BuildPortfolioReturnProjection("fuzz-portfolio", asOf, flows, &terminalMoney, true)
		if err != nil {
			t.Fatalf("identical two-term XIRR input returned error on second run: %v", err)
		}

		if first.Status != second.Status || first.Reason != second.Reason {
			t.Fatalf("XIRR projection is non-deterministic: first=%+v second=%+v", first, second)
		}
		if (first.XIRR == nil) != (second.XIRR == nil) {
			t.Fatalf("XIRR nil/non-nil result changed between identical runs")
		}
		if first.XIRR != nil {
			if !first.XIRR.FitsStorage() {
				t.Fatalf("available XIRR does not fit canonical Decimal storage: %s", first.XIRR.String())
			}
			if second.XIRR == nil || first.XIRR.String() != second.XIRR.String() {
				t.Fatalf("XIRR value changed between identical runs: first=%v second=%v", first.XIRR, second.XIRR)
			}
			rate, err := strconv.ParseFloat(first.XIRR.String(), 64)
			if err != nil || math.IsNaN(rate) || math.IsInf(rate, 0) || rate <= -1 {
				t.Fatalf("available XIRR is outside the valid finite domain: value=%q err=%v", first.XIRR.String(), err)
			}
		}
	})
}
