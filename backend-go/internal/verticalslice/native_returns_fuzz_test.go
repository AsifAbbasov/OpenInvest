package verticalslice

// Provenance: PR #229 at 657c34f813ac7faca1a9aa567b664dee5bc01be8; backend-go/internal/verticalslice/returns_fuzz_test.go.
// Original source and seed evidence remain retained in Issue #280.
import (
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
)

func FuzzParseBusinessDate(f *testing.F) {
	for _, seed := range []string{
		"2026-06-20",
		"2000-02-29",
		"1900-02-29",
		"2026-13-01",
		"2026-00-00",
		"2026-6-20",
		"",
		"\x00",
		"2026-01-01",
		"2026-01-32",
		" 2026-01-01 ",
		"2026-1-1",
		"2026-02-30",
		"0000-00-00",
		"9999-12-31",
		"2024-02-29",
		"2023-02-29",
		" 2026-01-01",
		"2026-01-01 ",
		"2025-02-29",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		value, err := parseBusinessDate(input)
		repeated, repeatedErr := parseBusinessDate(input)
		if (err == nil) != (repeatedErr == nil) {
			t.Fatal("BusinessDate success/error nondeterminism")
		}
		if err != nil {
			if err.Error() != repeatedErr.Error() {
				t.Fatal("BusinessDate error nondeterminism")
			}
			return
		}
		if !value.Equal(repeated) || value.Format("2006-01-02") != repeated.Format("2006-01-02") || repeated.Location() != time.UTC {
			t.Fatal("BusinessDate repeated value/format/location drift")
		}
		if got := value.Format("2006-01-02"); got != input {
			t.Fatalf("BusinessDate normalization drift: input=%q got=%q", input, got)
		}
		if value.Location() != time.UTC {
			t.Fatal("BusinessDate is not UTC")
		}
		if standard, err := time.Parse("2006-01-02", input); err != nil || !standard.Equal(value) {
			t.Fatalf("parseBusinessDate accepted value rejected by time.Parse: %q: %v", input, err)
		}
	})
}

func FuzzBuildPortfolioReturnProjection(f *testing.F) {
	seeds := []struct {
		asOf, flowDate, flowAmount, terminal string
	}{
		{"2027-01-01", "2026-01-01", "-100.00000000", "110.00000000"},
		{"2027-01-01", "2026-01-01", "-100.00000000", "80.00000000"},
		{"2026-01-01", "2026-01-01", "-100.00000000", "110.00000000"},
		{"bad-date", "2026-01-01", "-100.00000000", "110.00000000"},
	}
	for _, seed := range seeds {
		f.Add(seed.asOf, seed.flowDate, seed.flowAmount, seed.terminal)
	}

	f.Fuzz(func(t *testing.T, asOf, flowDate, flowAmountText, terminalText string) {
		flowAmount, err := decimal.FromString(flowAmountText)
		if err != nil {
			return
		}
		terminalAmount, err := decimal.FromString(terminalText)
		if err != nil {
			return
		}

		terminal := Money{Amount: terminalAmount, Currency: RUB}
		projection, err := BuildPortfolioReturnProjection(
			"fuzz-portfolio",
			asOf,
			[]PortfolioReturnCashFlow{{Date: flowDate, Amount: flowAmount}},
			&terminal,
			true,
		)
		if err != nil {
			return
		}

		switch projection.Status {
		case PortfolioReturnAvailableStatus:
			if projection.XIRR == nil {
				t.Fatalf("AVAILABLE projection has nil XIRR: asOf=%q flowDate=%q", asOf, flowDate)
			}
			if !projection.XIRR.FitsStorage() {
				t.Fatalf("AVAILABLE projection has non-storable XIRR: %s", projection.XIRR.String())
			}
			if projection.Reason != "" {
				t.Fatalf("AVAILABLE projection unexpectedly has reason=%q", projection.Reason)
			}
		case PortfolioReturnUnavailableStatus:
			if projection.XIRR != nil {
				t.Fatalf("UNAVAILABLE projection unexpectedly has XIRR=%s", projection.XIRR.String())
			}
			if projection.Reason == "" {
				t.Fatalf("UNAVAILABLE projection has empty reason")
			}
		default:
			t.Fatalf("unexpected projection status=%q", projection.Status)
		}
	})
}
