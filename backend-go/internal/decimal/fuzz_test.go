package decimal

import "testing"

func FuzzDecimalFromString(f *testing.F) {
	for _, seed := range []string{
		"0",
		"0.00000000",
		"1",
		"-1",
		"99999999999999999999.99999999",
		"-99999999999999999999.99999999",
		"01",
		"+1",
		"1.",
		"1.000000000",
		"",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		got, err := FromString(input)
		if err != nil {
			return
		}

		if !got.FitsStorage() {
			t.Fatalf("FromString accepted value that does not fit NUMERIC(%d,%d): %q -> %s", Precision, Scale, input, got.String())
		}

		canonical := got.String()
		roundTrip, err := FromString(canonical)
		if err != nil {
			t.Fatalf("canonical round-trip rejected %q derived from %q: %v", canonical, input, err)
		}
		if !got.Equal(roundTrip) {
			t.Fatalf("round-trip changed value: input=%q canonical=%q got=%s roundTrip=%s", input, canonical, got.String(), roundTrip.String())
		}
	})
}
