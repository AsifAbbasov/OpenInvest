package decimal

import "testing"

func FuzzDecimalFixedScaleAlgebra(f *testing.F) {
	f.Add("1.00000000", "2.00000000", "3.00000000")
	f.Add("0.00000001", "0.50000000", "-1.00000000")
	f.Add("99999999999999999999.99999999", "0.00000001", "1.00000000")

	f.Fuzz(func(t *testing.T, aText, bText, cText string) {
		a, err := FromString(aText)
		if err != nil {
			return
		}
		b, err := FromString(bText)
		if err != nil {
			return
		}
		c, err := FromString(cText)
		if err != nil {
			return
		}

		// Addition/subtraction are exact in the fixed-scale representation even when
		// the transient derived value grows beyond persistence precision.
		if got := a.Add(b).Sub(b); !got.Equal(a) {
			t.Fatalf("(a+b)-b != a: a=%s b=%s got=%s", a.String(), b.String(), got.String())
		}
		if got := a.Sub(b).Add(b); !got.Equal(a) {
			t.Fatalf("(a-b)+b != a: a=%s b=%s got=%s", a.String(), b.String(), got.String())
		}
		if !a.Add(b).Equal(b.Add(a)) {
			t.Fatalf("addition is not commutative: a=%s b=%s", a.String(), b.String())
		}
		if !a.Mul(b).Equal(b.Mul(a)) {
			t.Fatalf("Half-Even multiplication is not commutative: a=%s b=%s", a.String(), b.String())
		}

		zero := Zero()
		if !a.Add(zero).Equal(a) || !a.Sub(zero).Equal(a) {
			t.Fatalf("zero identity drift: a=%s", a.String())
		}

		// Do not assert full distributivity/invertibility: each Mul/Div is intentionally
		// quantized to scale 8. Instead assert determinism plus the actual persistence
		// contract: storage-fitting values must round-trip through canonical ingress,
		// while derived values outside NUMERIC(28,8) must not be re-admitted as if they
		// were persistence-safe.
		left1 := a.Mul(b).Add(a.Mul(c))
		left2 := a.Mul(b).Add(a.Mul(c))
		if !left1.Equal(left2) {
			t.Fatalf("fixed-scale arithmetic is nondeterministic")
		}

		for _, value := range []Decimal{a, b, c, a.Add(b), a.Sub(b), a.Mul(b), left1} {
			text := value.String()
			roundTrip, parseErr := FromString(text)
			if value.FitsStorage() {
				if parseErr != nil || !roundTrip.Equal(value) {
					t.Fatalf("storage-fitting canonical arithmetic value failed round-trip: %s err=%v", text, parseErr)
				}
				continue
			}
			if parseErr == nil {
				t.Fatalf("derived value outside NUMERIC(28,8) was re-admitted by canonical ingress: %s", text)
			}
		}

		if !b.IsZero() {
			q1, err1 := a.Div(b)
			q2, err2 := a.Div(b)
			if (err1 == nil) != (err2 == nil) {
				t.Fatalf("division error nondeterminism: %v vs %v", err1, err2)
			}
			if err1 == nil {
				if !q1.Equal(q2) {
					t.Fatalf("division result nondeterminism: %s vs %s", q1.String(), q2.String())
				}
				text := q1.String()
				roundTrip, parseErr := FromString(text)
				if q1.FitsStorage() {
					if parseErr != nil || !roundTrip.Equal(q1) {
						t.Fatalf("storage-fitting division result failed canonical round-trip: %s err=%v", text, parseErr)
					}
				} else if parseErr == nil {
					t.Fatalf("division result outside NUMERIC(28,8) was re-admitted by canonical ingress: %s", text)
				}
			}
		}
	})
}
