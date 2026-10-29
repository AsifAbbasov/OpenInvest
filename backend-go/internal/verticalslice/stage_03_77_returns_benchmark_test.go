package verticalslice

import (
	"fmt"
	"os"
	"testing"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
)

var benchmarkReturnProjection PortfolioReturnProjection

// BenchmarkXIRRProjectionCashFlowAggregation measures the existing XIRR path
// including validation, same-day cash-flow aggregation and root calculation.
// Its 28 distinct synthetic input dates avoid presenting 100k rows as 100k
// independent investment dates. No market prices are inferred.
func BenchmarkXIRRProjectionCashFlowAggregation(b *testing.B) {
	for _, count := range []int{1_000, 10_000, 100_000} {
		b.Run(fmt.Sprintf("flows=%d", count), func(b *testing.B) {
			if count > 1_000 && os.Getenv("OPENINVEST_BENCH_LARGE") != "1" {
				b.Skip("set OPENINVEST_BENCH_LARGE=1 for 10k/100k synthetic workloads")
			}
			minusOne := decimal.Must("-1.00000000")
			dates := make([]string, 28)
			for i := range dates {
				dates[i] = fmt.Sprintf("2026-01-%02d", i+1)
			}
			flows := make([]PortfolioReturnCashFlow, count)
			for i := range flows {
				flows[i] = PortfolioReturnCashFlow{Date: dates[i%len(dates)], Amount: minusOne}
			}
			terminal := &Money{Amount: decimal.Must(fmt.Sprintf("%d.00000000", count*12/10)), Currency: RUB}
			if got, err := BuildPortfolioReturnProjection("synthetic-portfolio", "2027-01-01", flows, terminal, true); err != nil || got.Status != PortfolioReturnAvailableStatus || got.XIRR == nil {
				b.Fatalf("invalid benchmark fixture: status=%s reason=%s err=%v", got.Status, got.Reason, err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				got, err := BuildPortfolioReturnProjection("synthetic-portfolio", "2027-01-01", flows, terminal, true)
				if err != nil || got.Status != PortfolioReturnAvailableStatus || got.XIRR == nil {
					b.Fatalf("XIRR became unavailable: status=%s reason=%s err=%v", got.Status, got.Reason, err)
				}
				benchmarkReturnProjection = got
			}
		})
	}
}
