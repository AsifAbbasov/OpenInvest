package position

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
)

func m11IntegerDecimal(value int) decimal.Decimal {
	return decimal.Must(fmt.Sprintf("%d.00000000", value))
}

func m11StateEqual(a, b State) bool {
	return a.Open == b.Open &&
		a.Quantity.Equal(b.Quantity) &&
		a.WeightedAverageCost.Equal(b.WeightedAverageCost) &&
		a.AcquisitionBasis.Equal(b.AcquisitionBasis)
}

func TestM11StatefulPositionRebuildEquivalence(t *testing.T) {
	const (
		seed      int64 = 110011
		sequences      = 1000
		steps          = 50
	)
	rng := rand.New(rand.NewSource(seed))
	totalOperations := 0
	for sequence := 0; sequence < sequences; sequence++ {
		state := Empty()
		history := make([]Trade, 0, steps)
		units := 0
		for step := 0; step < steps; step++ {
			var trade Trade
			if units == 0 || rng.Intn(100) < 62 {
				quantity := 1 + rng.Intn(20)
				trade = Trade{
					Type:      "BUY",
					Quantity:  m11IntegerDecimal(quantity),
					UnitPrice: m11IntegerDecimal(1 + rng.Intn(1000)),
				}
				units += quantity
			} else {
				quantity := 1 + rng.Intn(20)
				if quantity > units {
					quantity = units
				}
				trade = Trade{
					Type:      "SELL",
					Quantity:  m11IntegerDecimal(quantity),
					UnitPrice: m11IntegerDecimal(1 + rng.Intn(1000)),
				}
				units -= quantity
			}
			next, err := Apply(state, trade)
			if err != nil {
				t.Fatalf("sequence=%d step=%d incremental apply: %v", sequence, step, err)
			}
			history = append(history, trade)
			rebuilt, err := Rebuild(history)
			if err != nil {
				t.Fatalf("sequence=%d step=%d rebuild: %v", sequence, step, err)
			}
			if !m11StateEqual(next, rebuilt) {
				t.Fatalf("sequence=%d step=%d incremental=%+v rebuilt=%+v", sequence, step, next, rebuilt)
			}
			if next.Quantity.String() != m11IntegerDecimal(units).String() {
				t.Fatalf("sequence=%d step=%d tracked units=%d state=%s", sequence, step, units, next.Quantity.String())
			}
			state = next
			totalOperations++
		}
	}
	t.Logf("M11_STATEFUL_POSITION_SEED=%d", seed)
	t.Logf("M11_STATEFUL_SEQUENCES=%d", sequences)
	t.Logf("M11_STATEFUL_OPERATIONS=%d", totalOperations)
	t.Log("M11_CM05_INCREMENTAL_REBUILD_EQUIVALENCE=PASS")
}

func FuzzM11PositionPipeline(f *testing.F) {
	for _, seed := range []struct {
		q1, p1, q2, p2, sell uint16
	}{
		{10, 100, 5, 200, 3},
		{1, 1, 1, 1, 2},
		{500, 999, 400, 101, 899},
	} {
		f.Add(seed.q1, seed.p1, seed.q2, seed.p2, seed.sell)
	}
	f.Fuzz(func(t *testing.T, q1, p1, q2, p2, sell uint16) {
		q1 = q1%1000 + 1
		q2 = q2%1000 + 1
		p1 = p1%5000 + 1
		p2 = p2%5000 + 1
		total := int(q1) + int(q2)
		sellUnits := int(sell)%total + 1
		history := []Trade{
			{Type: "BUY", Quantity: m11IntegerDecimal(int(q1)), UnitPrice: m11IntegerDecimal(int(p1))},
			{Type: "BUY", Quantity: m11IntegerDecimal(int(q2)), UnitPrice: m11IntegerDecimal(int(p2))},
			{Type: "SELL", Quantity: m11IntegerDecimal(sellUnits), UnitPrice: m11IntegerDecimal(int(p2))},
		}
		state := Empty()
		for _, trade := range history {
			next, err := Apply(state, trade)
			if err != nil {
				t.Fatalf("incremental apply: %v", err)
			}
			state = next
		}
		rebuilt, err := Rebuild(history)
		if err != nil {
			t.Fatalf("rebuild: %v", err)
		}
		if !m11StateEqual(state, rebuilt) {
			t.Fatalf("incremental=%+v rebuilt=%+v", state, rebuilt)
		}
	})
}
