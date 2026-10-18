package verticalslice

import (
	"math"
	"testing"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
)

// Expected values are independently quantized rational rates, not produced by
// either the numerical XIRR solver or the production Decimal.Div operation.
func TestXIRRExactYearHalfEvenRationalRegressions(t *testing.T) {
	const principal = "200000000.00000000"
	for _, tc := range []struct {
		name     string
		terminal string
		want     string
	}{
		{"positive below tie", "200000048.00000000", "0.00000024"},
		{"positive half even down original corpus", "200000049.00000000", "0.00000024"},
		{"positive next quantum", "200000050.00000000", "0.00000025"},
		{"positive half even up", "200000051.00000000", "0.00000026"},
		{"positive above tie", "200000052.00000000", "0.00000026"},
		{"negative below tie", "199999952.00000000", "-0.00000024"},
		{"negative half even down", "199999951.00000000", "-0.00000024"},
		{"negative next quantum", "199999950.00000000", "-0.00000025"},
		{"negative half even up", "199999949.00000000", "-0.00000026"},
		{"negative above tie", "199999948.00000000", "-0.00000026"},
		{"zero return", principal, "0.00000000"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := BuildPortfolioReturnProjection(
				"p", "2000-12-31",
				[]PortfolioReturnCashFlow{{Date: "2000-01-01", Amount: decimal.Must("-" + principal)}},
				stage377Money(tc.terminal), true,
			)
			if err != nil {
				t.Fatal(err)
			}
			if got.Status != PortfolioReturnAvailableStatus ||
				got.Reason != "" || got.XIRR == nil || got.XIRR.String() != tc.want {
				t.Fatalf("terminal=%s status=%s reason=%s xirr=%v want=%s",
					tc.terminal, got.Status, got.Reason, got.XIRR, tc.want)
			}
		})
	}
}

func TestXIRRExactYearFractionalMonetaryAmounts(t *testing.T) {
	got, err := BuildPortfolioReturnProjection(
		"p", "2000-12-31",
		[]PortfolioReturnCashFlow{{Date: "2000-01-01", Amount: decimal.Must("-0.12345678")}},
		stage377Money("0.12345679"), true,
	)
	if err != nil {
		t.Fatal(err)
	}
	// Exact rate = 0.00000001 / 0.12345678 = 0.000000081000006642...
	if got.Status != PortfolioReturnAvailableStatus || got.XIRR == nil ||
		got.XIRR.String() != "0.00000008" {
		t.Fatalf("fractional one-year rate: status=%s reason=%s xirr=%v", got.Status, got.Reason, got.XIRR)
	}
}

func TestXIRRExactYearPublicNumericalRange(t *testing.T) {
	for _, tc := range []struct {
		name      string
		principal string
		terminal  string
		want      string
		status    string
		reason    string
	}{
		{"near minus one representable", "100000000.00000000", "1.00000000", "-0.99999999", PortfolioReturnAvailableStatus, ""},
		{"near minus one quantizes outside domain", "100000000.00000000", "0.00000001", "", PortfolioReturnUnavailableStatus, PortfolioReturnReasonNumericalSolutionFailed},
		{"high reliable rate", "100.00000000", "10000100.00000000", "100000.00000000", PortfolioReturnAvailableStatus, ""},
		{"existing float resolution boundary", "100.00000000", "9000000100.00000000", "", PortfolioReturnUnavailableStatus, PortfolioReturnReasonNumericalSolutionFailed},
		// XIRR-284-01: the exact rate lies below 2^19 while its Half-Even
		// quantization lies on 2^19, whose upper float64 ULP is larger.
		{"just below float exponent boundary by 2 nanos", "10.00000000", "5242889.99999998", "524288.00000000", PortfolioReturnAvailableStatus, ""},
		{"Gatekeeper float exponent boundary original", "10.00000000", "5242889.99999999", "524288.00000000", PortfolioReturnAvailableStatus, ""},
		{"exact float exponent boundary rejected by legacy policy", "10.00000000", "5242890.00000000", "", PortfolioReturnUnavailableStatus, PortfolioReturnReasonNumericalSolutionFailed},
		{"just above float exponent boundary rejected by legacy policy", "10.00000000", "5242890.00000001", "", PortfolioReturnUnavailableStatus, PortfolioReturnReasonNumericalSolutionFailed},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := BuildPortfolioReturnProjection(
				"p", "2000-12-31",
				[]PortfolioReturnCashFlow{{Date: "2000-01-01", Amount: decimal.Must("-" + tc.principal)}},
				stage377Money(tc.terminal), true,
			)
			if err != nil {
				t.Fatal(err)
			}
			if got.Status != tc.status || got.Reason != tc.reason {
				t.Fatalf("status=%s reason=%s want=%s/%s", got.Status, got.Reason, tc.status, tc.reason)
			}
			if tc.status == PortfolioReturnUnavailableStatus {
				if got.XIRR != nil {
					t.Fatalf("unavailable XIRR must be nil: %v", got.XIRR)
				}
			} else if got.XIRR == nil || got.XIRR.String() != tc.want {
				t.Fatalf("available XIRR=%v want=%s", got.XIRR, tc.want)
			}
		})
	}
}

func TestXIRRAdjacentDayCountRetainsACT365Solver(t *testing.T) {
	// Independently computed at 90-digit precision from
	// (110/100)^(365/days)-1 and rounded to scale 8 Half-Even.
	for _, tc := range []struct {
		name string
		days int
		want string
	}{
		{"364 days", 364, "0.10028806"},
		{"365 days", 365, "0.10000000"},
		{"366 days", 366, "0.09971359"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := BuildPortfolioReturnProjection(
				"p", stage377Day("2000-01-01", tc.days),
				[]PortfolioReturnCashFlow{{Date: "2000-01-01", Amount: decimal.Must("-100.00000000")}},
				stage377Money("110.00000000"), true,
			)
			if err != nil {
				t.Fatal(err)
			}
			if got.Status != PortfolioReturnAvailableStatus || got.XIRR == nil ||
				got.XIRR.String() != tc.want {
				t.Fatalf("days=%d status=%s reason=%s xirr=%v want=%s",
					tc.days, got.Status, got.Reason, got.XIRR, tc.want)
			}
		})
	}
}

// TestXIRRExactYearLegacyNumericalBoundaryParity verifies the pre-fix
// production call path actually accepts the exact Gatekeeper counterexample.
// The main projection must keep that public status while using a rational
// quotient for Half-Even formatting.
func TestXIRRExactYearLegacyNumericalBoundaryParity(t *testing.T) {
	combined := []PortfolioReturnCashFlow{
		{Date: "2000-01-01", Amount: decimal.Must("-10.00000000")},
		{Date: "2000-12-31", Amount: decimal.Must("5242889.99999999")},
	}
	terms, err := xirrTerms(combined)
	if err != nil {
		t.Fatal(err)
	}
	roots, err := isolateExponentialRoots(terms)
	if err != nil || len(roots) != 1 {
		t.Fatalf("legacy two-term root: roots=%v err=%v", roots, err)
	}
	legacy, err := quantizeXIRR(math.Expm1(roots[0]), terms, combined)
	if err != nil || legacy.String() != "524288.00000000" {
		t.Fatalf("pre-fix public numerical path was not AVAILABLE: legacy=%v err=%v", legacy, err)
	}
	got, err := BuildPortfolioReturnProjection(
		"p", "2000-12-31", combined[:1], stage377Money("5242889.99999999"), true,
	)
	if err != nil || got.Status != PortfolioReturnAvailableStatus ||
		got.XIRR == nil || got.XIRR.String() != legacy.String() {
		t.Fatalf("public boundary drifted from historical AVAILABLE: projection=%+v err=%v", got, err)
	}
}
