package verticalslice

import (
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
)

func FuzzBusinessDate(f *testing.F) {
	for _, seed := range []string{
		"2026-01-01",
		"2000-02-29",
		"1900-02-29",
		"2026-02-30",
		"2026-1-1",
		"",
		"0000-00-00",
		"9999-12-31",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		parsed, err := parseBusinessDate(input)
		if err != nil {
			return
		}
		if got := parsed.Format("2006-01-02"); got != input {
			t.Fatalf("BusinessDate did not round-trip: input=%q got=%q", input, got)
		}
		if parsed.Location() != time.UTC {
			t.Fatalf("BusinessDate parser returned non-UTC location: %v", parsed.Location())
		}
	})
}

func FuzzXIRRCoefficients(f *testing.F) {
	f.Add(int64(-1000), int64(500), int64(700), int32(0), int32(180), int32(365))
	f.Add(int64(-1), int64(0), int64(2), int32(0), int32(1), int32(365))
	f.Add(int64(-100), int64(250), int64(-50), int32(0), int32(100), int32(200))

	f.Fuzz(func(t *testing.T, a, b, c int64, d1, d2, d3 int32) {
		const amountBound = int64(1_000_000_000_000)
		a %= amountBound
		b %= amountBound
		c %= amountBound

		const dayBound = int32(36500)
		d1 %= dayBound
		d2 %= dayBound
		d3 %= dayBound

		base := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		flow := func(amount int64, offset int32) PortfolioReturnCashFlow {
			value, err := decimal.FromString(strconv.FormatInt(amount, 10))
			if err != nil {
				t.Fatalf("generated canonical decimal rejected: %v", err)
			}
			return PortfolioReturnCashFlow{
				Date:   base.AddDate(0, 0, int(offset)).Format("2006-01-02"),
				Amount: value,
			}
		}

		terms, err := xirrTerms([]PortfolioReturnCashFlow{
			flow(a, d1),
			flow(b, d2),
			flow(c, d3),
		})
		if err != nil {
			t.Fatalf("valid generated XIRR terms rejected: %v", err)
		}

		previousYears := math.Inf(-1)
		for _, term := range terms {
			if math.IsNaN(term.years) || math.IsInf(term.years, 0) {
				t.Fatalf("non-finite XIRR exponent: %v", term.years)
			}
			if math.IsNaN(term.coefficient) || math.IsInf(term.coefficient, 0) {
				t.Fatalf("non-finite XIRR coefficient: %v", term.coefficient)
			}
			if term.years < previousYears {
				t.Fatalf("XIRR terms are not normalized in ascending date order")
			}
			previousYears = term.years
		}

		if len(terms) >= 2 {
			roots, err := isolateExponentialRoots(terms)
			if err != nil {
				return
			}
			for _, root := range roots {
				if math.IsNaN(root) || math.IsInf(root, 0) {
					t.Fatalf("non-finite isolated XIRR root: %v", root)
				}
			}
		}
	})
}
