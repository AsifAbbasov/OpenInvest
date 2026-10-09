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
			t.Fatalf("accepted Decimal does not fit storage: %q -> %s", input, value.String())
		}
		canonical := value.String()
		reparsed, err := FromString(canonical)
		if err != nil {
			t.Fatalf("canonical Decimal failed to reparse: input=%q canonical=%q err=%v", input, canonical, err)
		}
		if !value.Equal(reparsed) {
			t.Fatalf("Decimal round-trip drift: input=%q canonical=%q reparsed=%q", input, canonical, reparsed.String())
		}
	})
}

func FuzzDecimalArithmetic(f *testing.F) {
	for _, seed := range [][2]string{
		{"1", "2"},
		{"-1", "3"},
		{"0.00000001", "0.50000000"},
		{"99999999999999999999.99999999", "1"},
		{"0", "0"},
	} {
		f.Add(seed[0], seed[1])
	}

	f.Fuzz(func(t *testing.T, leftText, rightText string) {
		left, err := FromString(leftText)
		if err != nil {
			return
		}
		right, err := FromString(rightText)
		if err != nil {
			return
		}

		leftBefore := left.String()
		rightBefore := right.String()

		_ = left.Add(right).String()
		_ = left.Sub(right).String()
		_ = left.Mul(right).String()

		quotient, divErr := left.Div(right)
		if right.IsZero() {
			if divErr == nil {
				t.Fatalf("division by zero unexpectedly succeeded: %q / %q", leftText, rightText)
			}
		} else if divErr != nil {
			t.Fatalf("division by non-zero Decimal failed: %q / %q: %v", leftText, rightText, divErr)
		} else {
			_ = quotient.String()
		}

		if left.String() != leftBefore || right.String() != rightBefore {
			t.Fatalf("Decimal arithmetic mutated an operand: left=%q right=%q", leftText, rightText)
		}
	})
}
