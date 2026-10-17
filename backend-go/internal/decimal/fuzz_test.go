package decimal

import "testing"

func FuzzDecimalFromString(f *testing.F) {
	for _, seed := range []string{
		"0", "1", "-1", "12.34",
		"99999999999999999999.99999999",
		"-99999999999999999999.99999999",
		"01", "+1", "1.", "1.000000000", " 1 ", "",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		value, err := FromString(input)
		if err != nil {
			return
		}
		if !value.FitsStorage() {
			t.Fatalf("accepted decimal does not fit NUMERIC(%d,%d): %q", Precision, Scale, input)
		}

		canonical := value.String()
		reparsed, err := FromString(canonical)
		if err != nil {
			t.Fatalf("canonical decimal %q from input %q is not reparsable: %v", canonical, input, err)
		}
		if !value.Equal(reparsed) {
			t.Fatalf("decimal round-trip changed value: input=%q canonical=%q", input, canonical)
		}
		if len(canonical) > maxLexicalBytes {
			t.Fatalf("canonical decimal exceeds lexical bound: %q", canonical)
		}
	})
}
