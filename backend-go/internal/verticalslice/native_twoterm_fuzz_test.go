package verticalslice

// Provenance: PR #225 at 1b1c03a0bf0fb65bee5ad5a18b8c602fbe08a71b; stage_03_77_returns_fuzz_test.go.
import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
)

func FuzzXIRRTwoTermProjection(f *testing.F) {
	f.Add(uint64(100), uint64(110), uint16(365))
	f.Add(uint64(100), uint64(80), uint16(365))
	f.Add(uint64(1), uint64(1), uint16(1))
	f.Add(uint64(1_000_000), uint64(2_000_000), uint16(730))

	// 365-day positive/negative and exact Half-Even tie controls (raw values increment by one).
	f.Add(uint64(99), uint64(109), uint16(364))
	f.Add(uint64(99), uint64(79), uint16(364))
	f.Add(uint64(199999999), uint64(200000000), uint16(364))
	f.Add(uint64(199999999), uint64(200000002), uint16(364))
	f.Add(uint64(199999999), uint64(199999998), uint16(364))
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
		if days == 365 {
			// Closed-form one-year reference using exact integer arithmetic,
			// independent of the production XIRR solver and Decimal rounding.
			want := nativeOneYearRate(principal, terminal)
			if first.Status != PortfolioReturnAvailableStatus || first.XIRR == nil || first.XIRR.String() != want {
				t.Fatalf("one-year financial oracle: projection=%+v want=%s", first, want)
			}
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

// nativeOneYearRate quantizes (terminal-principal)/principal to scale 8, Half-Even.
func nativeOneYearRate(principal, terminal uint64) string {
	n := new(big.Int).Sub(new(big.Int).SetUint64(terminal), new(big.Int).SetUint64(principal))
	negative := n.Sign() < 0
	n.Abs(n).Mul(n, big.NewInt(100000000))
	denominator := new(big.Int).SetUint64(principal)
	q, rem := new(big.Int), new(big.Int)
	q.QuoRem(n, denominator, rem)
	cmp := new(big.Int).Lsh(rem, 1).Cmp(denominator)
	if cmp > 0 || (cmp == 0 && q.Bit(0) == 1) {
		q.Add(q, big.NewInt(1))
	}
	whole, fraction := new(big.Int), new(big.Int)
	whole.QuoRem(q, big.NewInt(100000000), fraction)
	sign := ""
	if negative && q.Sign() != 0 {
		sign = "-"
	}
	return fmt.Sprintf("%s%s.%08d", sign, whole.String(), fraction.Int64())
}
