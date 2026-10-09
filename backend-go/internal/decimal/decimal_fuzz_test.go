package decimal

import "testing"

func FuzzDecimalFromString(f *testing.F) {
	for _, seed := range []string{
		"0",
		"-0",
		"1",
		"-1",
		"0.5",
		"12.34000000",
		"99999999999999999999.99999999",
		"+1",
		"001",
		"1.",
		"1e2",
		"",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		value, err := FromString(input)
		if err != nil {
			return
		}
		if !value.FitsStorage() {
			t.Fatalf("accepted Decimal does not fit canonical storage: input=%q value=%q", input, value.String())
		}

		canonical := value.String()
		roundTrip, err := FromString(canonical)
		if err != nil {
			t.Fatalf("canonical Decimal failed to parse: input=%q canonical=%q err=%v", input, canonical, err)
		}
		if !value.Equal(roundTrip) {
			t.Fatalf("Decimal round-trip changed value: input=%q canonical=%q", input, canonical)
		}
		if roundTrip.String() != canonical {
			t.Fatalf("Decimal canonicalization is not idempotent: first=%q second=%q", canonical, roundTrip.String())
		}
	})
}
