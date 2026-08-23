package decimal

import "testing"

func TestFromStringEnforcesCanonicalNumeric28Scale8Precision(t *testing.T) {
	for _, input := range []string{
		"99999999999999999999.99999999",
		"-99999999999999999999.99999999",
		"0.00000001",
		"0",
	} {
		if _, err := FromString(input); err != nil {
			t.Fatalf("expected %q to fit NUMERIC(28,8): %v", input, err)
		}
	}

	for _, input := range []string{
		"100000000000000000000",
		"100000000000000000000.00000000",
		"-100000000000000000000.00000000",
	} {
		if _, err := FromString(input); err == nil {
			t.Fatalf("expected %q to exceed NUMERIC(28,8)", input)
		}
	}
}

func TestFitsStorageRejectsArithmeticGrowthBeyondNumeric28Scale8(t *testing.T) {
	left := Must("99999999999999999999.00000000")
	right := Must("2.00000000")
	derived := left.Mul(right)
	if derived.FitsStorage() {
		t.Fatalf("expected derived value %s to exceed NUMERIC(28,8)", derived.String())
	}
}
