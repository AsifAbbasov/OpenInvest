package decimal

import "testing"

func FuzzFromString(f *testing.F) {
	for _, seed := range []string{
		"0",
		"0.00000000",
		"1",
		"-1",
		"99999999999999999999.99999999",
		"-99999999999999999999.99999999",
		"1.23456789",
		"",
		"+1",
		"01",
		"1.",
		"1e2",
		" 1 ",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		got, err := FromString(input)
		if err != nil {
			return
		}
		if !got.FitsStorage() {
			t.Fatalf("accepted decimal does not fit storage: input=%q value=%q", input, got.String())
		}

		canonical := got.String()
		reparsed, err := FromString(canonical)
		if err != nil {
			t.Fatalf("canonical decimal failed to reparse: input=%q canonical=%q err=%v", input, canonical, err)
		}
		if !got.Equal(reparsed) {
			t.Fatalf("decimal round-trip changed value: input=%q canonical=%q", input, canonical)
		}
	})
}
