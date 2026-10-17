package decimal

import "testing"

func FuzzFromString(f *testing.F) {
	for _, seed := range []string{
		"0",
		"-0",
		"1",
		"-1",
		"0.00000001",
		"99999999999999999999.99999999",
		"+1",
		"01",
		"1.",
		".1",
		"1e2",
		" 1 ",
		"\x00",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		value, err := FromString(input)
		if err != nil {
			return
		}
		if !value.FitsStorage() {
			t.Fatalf("accepted Decimal does not fit storage: %q -> %s", input, value.String())
		}

		canonical := value.String()
		roundTrip, err := FromString(canonical)
		if err != nil {
			t.Fatalf("canonical Decimal failed to parse: input=%q canonical=%q err=%v", input, canonical, err)
		}
		if !value.Equal(roundTrip) {
			t.Fatalf("Decimal round trip changed value: input=%q canonical=%q", input, canonical)
		}
	})
}
