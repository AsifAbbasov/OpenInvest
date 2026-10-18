package verticalslice

// Exact #226 three typed seed tuples and flow-to-terms oracle; 448666172ac03b0d5a078ac3e00ca80e78121553.
// Independent integer-day aggregation checks normalization without a second solver.
import (
	"math"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
)

func FuzzXIRRThreeFlowTerms(f *testing.F) {
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

		// Independent three-flow reference: aggregate signed integer cash by day,
		// drop zero buckets, sort dates, rebase years and normalize by max magnitude.
		byDay := map[int32]int64{}
		for i, day := range []int32{d1, d2, d3} {
			byDay[day] += []int64{a, b, c}[i]
		}
		var days []int
		maxMagnitude := float64(0)
		for day, amount := range byDay {
			if amount != 0 {
				days = append(days, int(day))
				maxMagnitude = math.Max(maxMagnitude, math.Abs(float64(amount)))
			}
		}
		sort.Ints(days)
		if len(terms) != len(days) {
			t.Fatalf("normalization cardinality: got=%d want=%d", len(terms), len(days))
		}
		for i, day := range days {
			wantYears := float64(day-days[0]) / 365
			wantCoefficient := float64(byDay[int32(day)]) / maxMagnitude
			if math.Abs(terms[i].years-wantYears) > 1e-10 || math.Abs(terms[i].coefficient-wantCoefficient) > 1e-12 {
				t.Fatalf("normalized term mismatch: got=%+v want=(%g,%g)", terms[i], wantYears, wantCoefficient)
			}
			if terms[i].coefficient == 0 || (i > 0 && terms[i].years <= terms[i-1].years) {
				t.Fatal("normalization kept zero or duplicate/nonascending dates")
			}
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
