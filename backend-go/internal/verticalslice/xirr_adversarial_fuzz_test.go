package verticalslice

import (
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
)

func FuzzXIRRAdversarialSequence(f *testing.F) {
	f.Add([]byte{0, 10, 1, 20, 2, 30}, uint16(365), uint64(10000))
	f.Add([]byte{0, 255, 1, 1, 2, 255, 3, 1}, uint16(730), uint64(1))
	f.Add([]byte{}, uint16(1), uint64(1))

	f.Fuzz(func(t *testing.T, raw []byte, spanDays uint16, terminalRaw uint64) {
		if len(raw) > maxAmbiguousXIRRTerms*2 {
			raw = raw[:maxAmbiguousXIRRTerms*2]
		}
		count := len(raw) / 2
		if count < 1 {
			return
		}

		span := int(spanDays)
		if span == 0 {
			span = 1
		}

		base, _ := time.Parse("2006-01-02", "2020-01-01")
		flows := make([]PortfolioReturnCashFlow, 0, count)
		for i := 0; i < count; i++ {
			day := int(raw[2*i]) % (span + 1)
			magnitude := uint64(raw[2*i+1]) + 1
			amountText := fuzzXIRRDecimal(magnitude)
			if i%2 == 0 {
				amountText = "-" + amountText
			}
			flows = append(flows, PortfolioReturnCashFlow{
				Date:   base.AddDate(0, 0, day).Format("2006-01-02"),
				Amount: decimal.Must(amountText),
			})
		}

		terminalText := fuzzXIRRDecimal(terminalRaw%1000000 + 1)
		terminal := Money{Amount: decimal.Must(terminalText), Currency: RUB}
		asOf := base.AddDate(0, 0, span).Format("2006-01-02")

		first, err := BuildPortfolioReturnProjection("fuzz", asOf, flows, &terminal, true)
		if err != nil {
			return
		}
		second, err := BuildPortfolioReturnProjection("fuzz", asOf, flows, &terminal, true)
		if err != nil {
			t.Fatalf("XIRR became nondeterministic: first succeeded, second failed: %v", err)
		}

		if first.Status != second.Status || first.Reason != second.Reason {
			t.Fatalf("XIRR status nondeterminism: first=%+v second=%+v", first, second)
		}
		if (first.XIRR == nil) != (second.XIRR == nil) {
			t.Fatalf("XIRR nil-state nondeterminism")
		}
		if first.XIRR != nil {
			if first.XIRR.String() != second.XIRR.String() {
				t.Fatalf("XIRR value nondeterminism: %s vs %s", first.XIRR.String(), second.XIRR.String())
			}
			if !first.XIRR.FitsStorage() {
				t.Fatalf("available XIRR escaped Decimal storage: %s", first.XIRR.String())
			}
			if first.Status != PortfolioReturnAvailableStatus || first.Reason != "" {
				t.Fatalf("XIRR present with inconsistent status/reason: %+v", first)
			}
		} else if first.Status == PortfolioReturnAvailableStatus {
			t.Fatalf("AVAILABLE XIRR projection has nil rate")
		}
	})
}

func fuzzXIRRDecimal(raw uint64) string {
	whole := raw%1000000000 + 1
	return xirrUint(whole) + ".00000000"
}

func xirrUint(v uint64) string {
	if v == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for v > 0 {
		i--
		b[i] = byte('0' + v%10)
		v /= 10
	}
	return string(b[i:])
}
