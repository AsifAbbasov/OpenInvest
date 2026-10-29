package position

import (
	"fmt"
	"os"
	"testing"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
)

var benchmarkRebuiltPosition State

// BenchmarkCanonicalPositionRebuild measures the existing WAC engine using
// deterministic BUY/SELL trades. Fixtures are built outside the timed region.
func BenchmarkCanonicalPositionRebuild(b *testing.B) {
	for _, count := range []int{1_000, 10_000, 100_000} {
		b.Run(fmt.Sprintf("trades=%d", count), func(b *testing.B) {
			if count > 1_000 && os.Getenv("OPENINVEST_BENCH_LARGE") != "1" {
				b.Skip("set OPENINVEST_BENCH_LARGE=1 for 10k/100k synthetic workloads")
			}
			one := decimal.Must("1.00000000")
			prices := make([]decimal.Decimal, 17)
			for i := range prices {
				prices[i] = decimal.Must(fmt.Sprintf("%d.00000000", 100+i))
			}
			trades := make([]Trade, count)
			for i := range trades {
				kind := "BUY"
				if i%5 == 4 {
					kind = "SELL"
				}
				trades[i] = Trade{Type: kind, Quantity: one, UnitPrice: prices[i%len(prices)]}
			}
			if state, err := Rebuild(trades); err != nil || !state.Open {
				b.Fatalf("invalid benchmark fixture: state=%+v err=%v", state, err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				state, err := Rebuild(trades)
				if err != nil {
					b.Fatal(err)
				}
				benchmarkRebuiltPosition = state
			}
		})
	}
}
