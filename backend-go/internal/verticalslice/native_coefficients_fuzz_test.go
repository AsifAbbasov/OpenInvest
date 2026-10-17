package verticalslice

// Provenance: PR #228 at 7c09245c3c92cfa16bc40f839b0d6ce34b6b471a; backend-go/internal/verticalslice/stage_03_77_fuzz_test.go.
// Original source and seed evidence remain retained in Issue #280.
import (
	"math"
	"testing"
)

func FuzzXIRRCoefficients(f *testing.F) {
	f.Add(float64(0), float64(-100), float64(1), float64(110), float64(0.5), float64(10))
	f.Add(float64(0), float64(1), float64(1), float64(-1e-28), float64(2), float64(1))
	f.Add(float64(0), float64(-1), float64(1), float64(1e-320), float64(2), float64(0.5))
	f.Add(float64(0), float64(1), float64(1), float64(-2), float64(2), float64(1))

	f.Fuzz(func(t *testing.T, y1, c1, y2, c2, y3, c3 float64) {
		terms := []xirrTerm{
			{years: y1, coefficient: c1},
			{years: y2, coefficient: c2},
			{years: y3, coefficient: c3},
		}

		roots, err := isolateExponentialRoots(terms)
		if err != nil {
			return
		}
		for _, root := range roots {
			if math.IsNaN(root) || math.IsInf(root, 0) {
				t.Fatalf("successful XIRR isolation returned non-finite root: terms=%+v roots=%v", terms, roots)
			}
			normalized := normalizeExponentialTerms(terms)
			if len(normalized) == 0 {
				t.Fatalf("XIRR returned a root for an empty normalized term set: terms=%+v root=%g", terms, root)
			}
			value, scale, evalErr := scaledExponentialValue(normalized, root)
			if evalErr != nil {
				t.Fatalf("successful XIRR root cannot be evaluated: terms=%+v root=%g err=%v", terms, root, evalErr)
			}
			if math.IsNaN(value) || math.IsInf(value, 0) || math.IsNaN(scale) || math.IsInf(scale, 0) {
				t.Fatalf("successful XIRR root produced non-finite evaluation: terms=%+v root=%g value=%g scale=%g", terms, root, value, scale)
			}
		}
	})
}
