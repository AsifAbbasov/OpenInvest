package decimal

import "testing"

func TestDecimalPreservesEightFractionalDigits(t *testing.T) {
	value, err := FromString("12.34")
	if err != nil {
		t.Fatalf("parse decimal: %v", err)
	}

	if got := value.String(); got != "12.34000000" {
		t.Fatalf("expected fixed scale decimal, got %q", got)
	}
}

func TestDecimalRejectsMoreThanEightFractionalDigits(t *testing.T) {
	if _, err := FromString("1.123456789"); err == nil {
		t.Fatal("expected decimal precision error")
	}
}

func TestDecimalMultipliesWithoutFloat(t *testing.T) {
	quantity := Must("100.00000000")
	price := Must("280.00000000")

	if got := quantity.Mul(price).String(); got != "28000.00000000" {
		t.Fatalf("expected 28000.00000000, got %s", got)
	}
}

func TestDecimalMulUsesHalfEvenRounding(t *testing.T) {
	tests := map[string]struct {
		left  string
		right string
		want  string
	}{
		"below half":        {"0.00000002", "0.60000000", "0.00000001"},
		"above half":        {"0.00000003", "0.60000000", "0.00000002"},
		"half to even down": {"0.00000005", "0.50000000", "0.00000002"},
		"half to even up":   {"0.00000007", "0.50000000", "0.00000004"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := Must(test.left).Mul(Must(test.right)).String()
			if got != test.want {
				t.Fatalf("expected %s, got %s", test.want, got)
			}
		})
	}
}

func TestDecimalMarshalJSONUsesStringValue(t *testing.T) {
	value := Must("123.45000000")
	encoded, err := value.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal decimal: %v", err)
	}
	if got := string(encoded); got != `"123.45000000"` {
		t.Fatalf("expected decimal JSON string, got %s", got)
	}
}
