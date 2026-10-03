package verticalslice

import (
	"testing"
)

func FuzzParseBusinessDate(f *testing.F) {
	for _, seed := range []string{
		"2026-06-20",
		"2000-02-29",
		"1900-02-29",
		"2026-13-01",
		"2026-00-00",
		"2026-6-20",
		"",
		"\x00",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		parsed, err := parseBusinessDate(input)
		if err != nil {
			return
		}
		if got := parsed.Format("2006-01-02"); got != input {
			t.Fatalf("accepted non-canonical BusinessDate: input=%q canonical=%q", input, got)
		}
	})
}

func FuzzXIRRTwoFlowProjection(f *testing.F) {
	f.Add("2025-01-01", "2026-01-01", "100.00000000", "110.00000000")
	f.Add("2025-01-01", "2026-01-01", "100.00000000", "0.00000000")
	f.Add("2026-01-01", "2025-01-01", "100.00000000", "110.00000000")

	f.Fuzz(func(t *testing.T, firstDate, secondDate, contributionText, terminalText string) {
		contribution, err := parseFuzzDecimal(contributionText)
		if err != nil || !contribution.IsPositive() {
			return
		}
		terminal, err := parseFuzzDecimal(terminalText)
		if err != nil || terminal.IsNegative() {
			return
		}
		if _, err := parseBusinessDate(firstDate); err != nil {
			return
		}
		if _, err := parseBusinessDate(secondDate); err != nil {
			return
		}

		flows := []PortfolioReturnCashFlow{
			{Date: firstDate, Amount: contribution.Mul(decimalMinusOne)},
			{Date: secondDate, Amount: terminal},
		}
		terms, err := xirrTerms(flows)
		if err != nil {
			return
		}
		_, _ = isolateExponentialRoots(terms)
	})
}
