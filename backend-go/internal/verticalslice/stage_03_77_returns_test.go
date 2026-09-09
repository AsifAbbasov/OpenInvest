package verticalslice

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
)

func stage377Decimal(v string) decimal.Decimal { return decimal.Must(v) }
func stage377Money(v string) *Money            { x := Money{Amount: stage377Decimal(v), Currency: RUB}; return &x }
func stage377Day(base string, days int) string {
	v, _ := time.Parse("2006-01-02", base)
	return v.AddDate(0, 0, days).Format("2006-01-02")
}

func TestXIRRSimplePositiveAndNegative(t *testing.T) {
	for _, tc := range []struct{ name, terminal, want string }{
		{"positive", "110.00000000", "0.10000000"},
		{"negative", "80.00000000", "-0.20000000"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := BuildPortfolioReturnProjection("p", stage377Day("2026-01-01", 365), []PortfolioReturnCashFlow{{Date: "2026-01-01", Amount: stage377Decimal("-100.00000000")}}, stage377Money(tc.terminal), true)
			if err != nil {
				t.Fatal(err)
			}
			if got.Status != PortfolioReturnAvailableStatus || got.XIRR == nil || got.XIRR.String() != tc.want {
				t.Fatalf("got status=%s reason=%s xirr=%v", got.Status, got.Reason, got.XIRR)
			}
		})
	}
}

func TestXIRRIrregularACT365Reference(t *testing.T) {
	got, err := BuildPortfolioReturnProjection("p", "2026-02-17", []PortfolioReturnCashFlow{
		{Date: "2025-01-05", Amount: stage377Decimal("-1000.00000000")},
		{Date: "2025-04-20", Amount: stage377Decimal("-250.00000000")},
		{Date: "2025-09-10", Amount: stage377Decimal("120.00000000")},
	}, stage377Money("1284.22996102"), true)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != PortfolioReturnAvailableStatus || got.XIRR == nil || got.XIRR.String() != "0.12048717" {
		t.Fatalf("status=%s reason=%s xirr=%v", got.Status, got.Reason, got.XIRR)
	}
}

func TestXIRRNoSignChangeAndLowerBoundaryFailClosed(t *testing.T) {
	asOf := stage377Day("2026-01-01", 365)
	noRoot, err := BuildPortfolioReturnProjection("p", asOf, []PortfolioReturnCashFlow{
		{Date: "2026-01-01", Amount: stage377Decimal("-100.00000000")},
		{Date: asOf, Amount: stage377Decimal("-50.00000000")},
	}, stage377Money("40.00000000"), true)
	if err != nil {
		t.Fatal(err)
	}
	if noRoot.Reason != PortfolioReturnReasonNoSignChange {
		t.Fatalf("no-root reason=%s", noRoot.Reason)
	}

	boundary, err := BuildPortfolioReturnProjection("p", asOf, []PortfolioReturnCashFlow{
		{Date: "2026-01-01", Amount: stage377Decimal("-100.00000000")},
	}, stage377Money("0.00000001"), true)
	if err != nil {
		t.Fatal(err)
	}
	if boundary.Reason != PortfolioReturnReasonNumericalSolutionFailed {
		t.Fatalf("boundary status=%s reason=%s xirr=%v", boundary.Status, boundary.Reason, boundary.XIRR)
	}
}

func TestXIRRMultipleRoots(t *testing.T) {
	asOf := stage377Day("2025-01-01", 730)
	got, err := BuildPortfolioReturnProjection("p", asOf, []PortfolioReturnCashFlow{
		{Date: "2025-01-01", Amount: stage377Decimal("-100.00000000")},
		{Date: stage377Day("2025-01-01", 365), Amount: stage377Decimal("230.00000000")},
		{Date: asOf, Amount: stage377Decimal("-200.00000000")},
	}, stage377Money("68.00000000"), true)
	if err != nil {
		t.Fatal(err)
	}
	if got.Reason != PortfolioReturnReasonAmbiguousMultipleRoots {
		t.Fatalf("got %s", got.Reason)
	}
}

func TestXIRRCloseRoots(t *testing.T) {
	base := "2024-01-01"
	asOf := stage377Day(base, 4*365)
	got, err := BuildPortfolioReturnProjection("p", asOf, []PortfolioReturnCashFlow{
		{Date: base, Amount: stage377Decimal("123.00000000")},
		{Date: stage377Day(base, 365), Amount: stage377Decimal("403.00000000")},
		{Date: stage377Day(base, 730), Amount: stage377Decimal("-429.00000000")},
		{Date: stage377Day(base, 1095), Amount: stage377Decimal("-175.00000000")},
		{Date: asOf, Amount: stage377Decimal("100.00000000")},
	}, stage377Money("65.00000000"), true)
	if err != nil {
		t.Fatal(err)
	}
	if got.Reason != PortfolioReturnReasonAmbiguousMultipleRoots {
		t.Fatalf("got %s", got.Reason)
	}
}

func TestXIRRUnavailableBoundaries(t *testing.T) {
	asOf := stage377Day("2026-01-01", 365)
	cases := []struct {
		name     string
		flows    []PortfolioReturnCashFlow
		terminal *Money
		complete bool
		want     string
	}{
		{"incomplete", []PortfolioReturnCashFlow{{Date: "2026-01-01", Amount: stage377Decimal("-100.00000000")}}, nil, false, PortfolioReturnReasonIncompleteValuation},
		{"no contribution", []PortfolioReturnCashFlow{{Date: "2026-01-01", Amount: stage377Decimal("20.00000000")}}, stage377Money("100.00000000"), true, PortfolioReturnReasonNoExternalContributions},
		{"same date", []PortfolioReturnCashFlow{{Date: asOf, Amount: stage377Decimal("-100.00000000")}}, stage377Money("110.00000000"), true, PortfolioReturnReasonInsufficientDateSpan},
		{"nonpositive", []PortfolioReturnCashFlow{{Date: "2026-01-01", Amount: stage377Decimal("-100.00000000")}}, stage377Money("0.00000000"), true, PortfolioReturnReasonNonPositiveTerminalValue},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := BuildPortfolioReturnProjection("p", asOf, tc.flows, tc.terminal, tc.complete)
			if err != nil {
				t.Fatal(err)
			}
			if got.Reason != tc.want {
				t.Fatalf("got %s", got.Reason)
			}
		})
	}
}

func TestXIRRSameDayAggregationAndFutureCutoff(t *testing.T) {
	asOf := stage377Day("2026-01-01", 365)
	got, err := BuildPortfolioReturnProjection("p", asOf, []PortfolioReturnCashFlow{
		{Date: "2026-01-01", Amount: stage377Decimal("-60.00000000")},
		{Date: "2026-01-01", Amount: stage377Decimal("-40.00000000")},
		{Date: stage377Day(asOf, 1), Amount: stage377Decimal("-900.00000000")},
	}, stage377Money("110.00000000"), true)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != PortfolioReturnAvailableStatus || got.XIRR == nil || got.XIRR.String() != "0.10000000" {
		t.Fatalf("status=%s reason=%s xirr=%v flows=%v", got.Status, got.Reason, got.XIRR, got.ExternalCashFlows)
	}
	if len(got.ExternalCashFlows) != 1 || got.ExternalCashFlows[0].Amount.String() != "-100.00000000" {
		t.Fatalf("flows=%v", got.ExternalCashFlows)
	}
}

func TestXIRRHalfEvenMidpoints(t *testing.T) {
	asOf := stage377Day("2026-01-01", 365)
	for _, tc := range []struct{ terminal, want string }{
		{"112.34567850", "0.12345678"},
		{"112.34567950", "0.12345680"},
	} {
		got, err := BuildPortfolioReturnProjection("p", asOf, []PortfolioReturnCashFlow{{Date: "2026-01-01", Amount: stage377Decimal("-100.00000000")}}, stage377Money(tc.terminal), true)
		if err != nil {
			t.Fatal(err)
		}
		if got.XIRR == nil || got.XIRR.String() != tc.want {
			t.Fatalf("terminal=%s got status=%s reason=%s xirr=%v want=%s", tc.terminal, got.Status, got.Reason, got.XIRR, tc.want)
		}
	}
}

func TestXIRRTangentFailsClosed(t *testing.T) {
	base := "2026-01-01"
	asOf := stage377Day(base, 730)
	// Polynomial in x=1/(1+r): (x-1)^2 = 1 - 2x + x^2.
	// Build external flows + terminal so the combined series has the double root r=0.
	got, err := BuildPortfolioReturnProjection("p", asOf, []PortfolioReturnCashFlow{
		{Date: base, Amount: stage377Decimal("1.00000000")},
		{Date: stage377Day(base, 365), Amount: stage377Decimal("-2.00000000")},
		{Date: asOf, Amount: stage377Decimal("-1.00000000")},
	}, stage377Money("2.00000000"), true)
	if err != nil {
		t.Fatal(err)
	}
	if got.Reason != PortfolioReturnReasonAmbiguousMultipleRoots {
		t.Fatalf("status=%s reason=%s xirr=%v", got.Status, got.Reason, got.XIRR)
	}
}

func TestXIRRTwoTermExtremeRatioUsesLogSpace(t *testing.T) {
	roots, err := isolateExponentialRoots([]xirrTerm{
		{years: 0, coefficient: -1},
		{years: 1, coefficient: 1e-320},
	})
	if err != nil {
		t.Fatalf("extreme two-term ratio must remain finite in log-space: %v", err)
	}
	if len(roots) != 1 || math.IsNaN(roots[0]) || math.IsInf(roots[0], 0) {
		t.Fatalf("extreme two-term ratio root mismatch: %v", roots)
	}
}

func TestXIRRScale8QuantizationFailsClosedBeyondFloatExactIntegerRange(t *testing.T) {
	asOf := stage377Day("2026-01-01", 365)
	got, err := BuildPortfolioReturnProjection(
		"p",
		asOf,
		[]PortfolioReturnCashFlow{{Date: "2026-01-01", Amount: stage377Decimal("-100.00000000")}},
		stage377Money("9000000100.00000000"),
		true,
	)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != PortfolioReturnUnavailableStatus || got.Reason != PortfolioReturnReasonNumericalSolutionFailed || got.XIRR != nil {
		t.Fatalf("extreme rate must fail closed at scale-8 boundary: status=%s reason=%s xirr=%v", got.Status, got.Reason, got.XIRR)
	}
}

func TestXIRRHighButReliableScale8UsesExactIntegerFormatting(t *testing.T) {
	asOf := stage377Day("2026-01-01", 365)
	got, err := BuildPortfolioReturnProjection(
		"p",
		asOf,
		[]PortfolioReturnCashFlow{{Date: "2026-01-01", Amount: stage377Decimal("-100.00000000")}},
		stage377Money("10000100.00000000"),
		true,
	)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != PortfolioReturnAvailableStatus || got.XIRR == nil || got.XIRR.String() != "100000.00000000" {
		t.Fatalf("high reliable scale-8 rate formatting drifted: status=%s reason=%s xirr=%v", got.Status, got.Reason, got.XIRR)
	}
}

func TestXIRRScaledEvaluationIncludesCoefficientMagnitude(t *testing.T) {
	terms := []xirrTerm{
		{years: 0, coefficient: 1},
		{years: 1, coefficient: -1e-28},
	}
	// At y=-70 the tiny coefficient is exponentially amplified and must dominate.
	value, scale, err := scaledExponentialValue(terms, -70)
	if err != nil {
		t.Fatal(err)
	}
	if !(value < 0) || !(scale > 0) {
		t.Fatalf("coefficient-aware scaling lost the amplified small term: value=%g scale=%g", value, scale)
	}
}

func TestXIRRRejectsExternalAggregateBeyondCanonicalDecimal(t *testing.T) {
	asOf := stage377Day("2026-01-01", 365)
	_, err := BuildPortfolioReturnProjection(
		"p",
		asOf,
		[]PortfolioReturnCashFlow{
			{Date: "2026-01-01", Amount: stage377Decimal("-99999999999999999999.00000000")},
			{Date: "2026-01-01", Amount: stage377Decimal("-1.00000001")},
		},
		stage377Money("1.00000000"),
		true,
	)
	if err == nil || !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("external aggregate beyond NUMERIC(28,8) must fail closed, err=%v", err)
	}
}

func TestXIRROneSignChangeFastPathHandlesLargeContributionHistory(t *testing.T) {
	base := "2020-01-01"
	flows := make([]PortfolioReturnCashFlow, 0, 1500)
	for day := 0; day < 1500; day++ {
		flows = append(flows, PortfolioReturnCashFlow{
			Date:   stage377Day(base, day),
			Amount: stage377Decimal("-1.00000000"),
		})
	}
	asOf := stage377Day(base, 1825)
	got, err := BuildPortfolioReturnProjection("p", asOf, flows, stage377Money("2000.00000000"), true)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != PortfolioReturnAvailableStatus || got.XIRR == nil {
		t.Fatalf("large one-sign-change series must solve without derivative recursion: status=%s reason=%s xirr=%v", got.Status, got.Reason, got.XIRR)
	}
	terms, err := xirrTerms(append(append([]PortfolioReturnCashFlow{}, flows...), PortfolioReturnCashFlow{Date: asOf, Amount: stage377Decimal("2000.00000000")}))
	if err != nil {
		t.Fatal(err)
	}
	if changes := exponentialCoefficientSignChanges(terms); changes != 1 {
		t.Fatalf("large conventional series sign changes=%d want 1", changes)
	}
}

func TestXIRRExactPolynomialMidpointDistinguishesNearTieBeyondFloat64(t *testing.T) {
	asOf := stage377Day("2026-01-01", 365)
	cases := []struct {
		name     string
		terminal string
		want     string
	}{
		{"just above", "1123456785000000000.00000001", "0.12345679"},
		{"just below", "1123456784999999999.99999999", "0.12345678"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := BuildPortfolioReturnProjection(
				"p",
				asOf,
				[]PortfolioReturnCashFlow{{Date: "2026-01-01", Amount: stage377Decimal("-1000000000000000000.00000000")}},
				stage377Money(tc.terminal),
				true,
			)
			if err != nil {
				t.Fatal(err)
			}
			if got.Status != PortfolioReturnAvailableStatus || got.XIRR == nil || got.XIRR.String() != tc.want {
				t.Fatalf("terminal=%s status=%s reason=%s xirr=%v want=%s", tc.terminal, got.Status, got.Reason, got.XIRR, tc.want)
			}
		})
	}
}

func TestPortfolioReturnsRejectsWhitespacePaddedBusinessDate(t *testing.T) {
	service := NewService(&stage377ReturnStoreStub{}, nil)
	_, err := service.GetPortfolioReturns(context.Background(), "subject", "portfolio", " 2026-01-01 ")
	if err == nil || !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("whitespace-padded asOfDate must be rejected, err=%v", err)
	}
}

type stage377ReturnStoreStub struct{ Store }

func (*stage377ReturnStoreStub) GetPortfolioReturns(context.Context, string, string, string) (PortfolioReturnProjection, error) {
	return PortfolioReturnProjection{}, nil
}

func TestXIRRMultiSignChangeComplexityFailsClosed(t *testing.T) {
	start, err := time.Parse("2006-01-02", "2020-01-01")
	if err != nil {
		t.Fatal(err)
	}
	flows := make([]PortfolioReturnCashFlow, 0, maxAmbiguousXIRRTerms+1)
	for index := 0; index < maxAmbiguousXIRRTerms+1; index++ {
		amount := "-100.00000000"
		if index%2 == 1 {
			amount = "100.00000000"
		}
		flows = append(flows, PortfolioReturnCashFlow{
			Date:   start.AddDate(0, 0, index).Format("2006-01-02"),
			Amount: decimal.Must(amount),
		})
	}
	terminal := Money{Amount: decimal.Must("110.00000000"), Currency: RUB}
	projection, err := BuildPortfolioReturnProjection(
		"portfolio-1",
		start.AddDate(0, 0, maxAmbiguousXIRRTerms+1).Format("2006-01-02"),
		flows,
		&terminal,
		true,
	)
	if err != nil {
		t.Fatalf("complexity guard must fail closed through projection status, got error: %v", err)
	}
	if projection.Status != PortfolioReturnUnavailableStatus || projection.Reason != PortfolioReturnReasonNumericalSolutionFailed {
		t.Fatalf("complex multi-sign-change series must fail closed, got status=%s reason=%s", projection.Status, projection.Reason)
	}
}
