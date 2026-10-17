package position

import (
	"testing"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
)

func fuzzPositiveDecimal(raw uint64) decimal.Decimal {
	integer := raw%1000000 + 1
	fraction := (raw / 1000000) % 100000000
	return decimal.Must(formatFuzzDecimal(integer, fraction))
}

func formatFuzzDecimal(integer uint64, fraction uint64) string {
	const digits = "00000000"
	frac := digits + uintToString(fraction)
	frac = frac[len(frac)-8:]
	return uintToString(integer) + "." + frac
}

func uintToString(v uint64) string {
	if v == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[i:])
}

func FuzzStatefulWACLedgerSequence(f *testing.F) {
	f.Add([]byte{0, 1, 10, 100, 0, 1, 20, 200, 1, 1, 5, 150})
	f.Add([]byte{0, 1, 1, 1, 1, 1, 2, 1})
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 4096 {
			t.Skip()
		}

		state := Empty()
		trades := make([]Trade, 0, len(data)/4)
		for i := 0; i+3 < len(data); i += 4 {
			kind := "BUY"
			if data[i]&1 == 1 {
				kind = "SELL"
			}
			qRaw := uint64(data[i+1])<<8 | uint64(data[i+2])
			pRaw := uint64(data[i+3]) + 1
			trade := Trade{
				Type:      kind,
				Quantity:  fuzzPositiveDecimal(qRaw),
				UnitPrice: fuzzPositiveDecimal(pRaw),
			}
			trades = append(trades, trade)

			before := state
			next, err := Apply(state, trade)
			if err != nil {
				if kind == "SELL" {
					// Rejection must be atomic from the caller's perspective: the input state is immutable.
					if state.Quantity.String() != before.Quantity.String() ||
						state.WeightedAverageCost.String() != before.WeightedAverageCost.String() ||
						state.AcquisitionBasis.String() != before.AcquisitionBasis.String() ||
						state.Open != before.Open {
						t.Fatalf("rejected SELL mutated prior state")
					}
				}
				continue
			}
			state = next

			if state.Open {
				if !state.Quantity.IsPositive() {
					t.Fatalf("open position has non-positive quantity: %s", state.Quantity.String())
				}
				if !state.WeightedAverageCost.IsPositive() {
					t.Fatalf("open position has non-positive WAC: %s", state.WeightedAverageCost.String())
				}
				if !state.AcquisitionBasis.IsPositive() {
					t.Fatalf("open position has non-positive acquisition basis: %s", state.AcquisitionBasis.String())
				}
				if !state.Quantity.FitsStorage() || !state.WeightedAverageCost.FitsStorage() || !state.AcquisitionBasis.FitsStorage() {
					t.Fatalf("open position escaped canonical NUMERIC(28,8) bounds: %+v", state)
				}
				wantBasis := state.Quantity.Mul(state.WeightedAverageCost)
				if !wantBasis.Equal(state.AcquisitionBasis) {
					t.Fatalf("basis invariant drift: qty=%s wac=%s basis=%s recomputed=%s",
						state.Quantity.String(), state.WeightedAverageCost.String(),
						state.AcquisitionBasis.String(), wantBasis.String())
				}
			} else if !state.Quantity.IsZero() || !state.WeightedAverageCost.IsZero() || !state.AcquisitionBasis.IsZero() {
				t.Fatalf("closed position retained financial state: %+v", state)
			}
		}

		rebuilt, err := Rebuild(trades)
		if err == nil {
			if rebuilt.Quantity.String() != state.Quantity.String() ||
				rebuilt.WeightedAverageCost.String() != state.WeightedAverageCost.String() ||
				rebuilt.AcquisitionBasis.String() != state.AcquisitionBasis.String() ||
				rebuilt.Open != state.Open {
				t.Fatalf("incremental/rebuild divergence: incremental=%+v rebuilt=%+v", state, rebuilt)
			}
		}
	})
}
