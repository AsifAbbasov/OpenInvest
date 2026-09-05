package verticalslice

import (
	"testing"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
)

func TestCalculateDividendCanonicalVector(t *testing.T) {
	cost := Money{Amount: decimal.Must("280000.00000000"), Currency: RUB}
	result, err := CalculateDividend(DividendCalculationRequest{
		Ticker:          "SBER",
		Quantity:        decimal.Must("1000.00000000"),
		DividendPerUnit: Money{Amount: decimal.Must("34.84000000"), Currency: RUB},
		PositionCost:    &cost,
	})
	if err != nil {
		t.Fatalf("CalculateDividend: %v", err)
	}
	if got := result.GrossDividend.Amount.String(); got != "34840.00000000" {
		t.Fatalf("gross dividend = %s", got)
	}
	if result.GrossYield == nil || result.GrossYield.String() != "0.12442857" {
		t.Fatalf("gross yield = %v", result.GrossYield)
	}
	if result.TaxIncluded {
		t.Fatal("taxIncluded must be false")
	}
	if result.MethodologyVersion != DividendCalculatorMethodologyVersion {
		t.Fatalf("methodology = %q", result.MethodologyVersion)
	}
}

func TestCalculateDividendWithoutPositionCostReturnsNullYield(t *testing.T) {
	result, err := CalculateDividend(DividendCalculationRequest{
		Ticker:          "SBER",
		Quantity:        decimal.Must("1000.00000000"),
		DividendPerUnit: Money{Amount: decimal.Must("34.84000000"), Currency: RUB},
	})
	if err != nil {
		t.Fatalf("CalculateDividend: %v", err)
	}
	if result.GrossYield != nil {
		t.Fatalf("gross yield must be nil, got %s", result.GrossYield.String())
	}
}

func TestCalculateDividendRejectsInvalidFinancialInputs(t *testing.T) {
	positiveCost := Money{Amount: decimal.Must("100.00000000"), Currency: RUB}
	zeroCost := Money{Amount: decimal.Zero(), Currency: RUB}
	negativeCost := Money{Amount: decimal.Must("-1.00000000"), Currency: RUB}

	tests := []struct {
		name    string
		request DividendCalculationRequest
	}{
		{"zero quantity", DividendCalculationRequest{Ticker: "SBER", Quantity: decimal.Zero(), DividendPerUnit: Money{Amount: decimal.Must("1.00000000"), Currency: RUB}, PositionCost: &positiveCost}},
		{"negative quantity", DividendCalculationRequest{Ticker: "SBER", Quantity: decimal.Must("-1.00000000"), DividendPerUnit: Money{Amount: decimal.Must("1.00000000"), Currency: RUB}, PositionCost: &positiveCost}},
		{"negative dividend", DividendCalculationRequest{Ticker: "SBER", Quantity: decimal.Must("1.00000000"), DividendPerUnit: Money{Amount: decimal.Must("-1.00000000"), Currency: RUB}, PositionCost: &positiveCost}},
		{"wrong dividend currency", DividendCalculationRequest{Ticker: "SBER", Quantity: decimal.Must("1.00000000"), DividendPerUnit: Money{Amount: decimal.Must("1.00000000"), Currency: "USD"}, PositionCost: &positiveCost}},
		{"zero position cost", DividendCalculationRequest{Ticker: "SBER", Quantity: decimal.Must("1.00000000"), DividendPerUnit: Money{Amount: decimal.Must("1.00000000"), Currency: RUB}, PositionCost: &zeroCost}},
		{"negative position cost", DividendCalculationRequest{Ticker: "SBER", Quantity: decimal.Must("1.00000000"), DividendPerUnit: Money{Amount: decimal.Must("1.00000000"), Currency: RUB}, PositionCost: &negativeCost}},
		{"wrong position currency", DividendCalculationRequest{Ticker: "SBER", Quantity: decimal.Must("1.00000000"), DividendPerUnit: Money{Amount: decimal.Must("1.00000000"), Currency: RUB}, PositionCost: &Money{Amount: decimal.Must("100.00000000"), Currency: "USD"}}},
		{"invalid ticker", DividendCalculationRequest{Ticker: "sber", Quantity: decimal.Must("1.00000000"), DividendPerUnit: Money{Amount: decimal.Must("1.00000000"), Currency: RUB}, PositionCost: &positiveCost}},
		{"ticker surrounding whitespace", DividendCalculationRequest{Ticker: " SBER ", Quantity: decimal.Must("1.00000000"), DividendPerUnit: Money{Amount: decimal.Must("1.00000000"), Currency: RUB}, PositionCost: &positiveCost}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := CalculateDividend(tt.request); err == nil {
				t.Fatal("expected rejection")
			}
		})
	}
}

func TestCalculateDividendRejectsUninitializedDecimalsWithoutPanic(t *testing.T) {
	tests := []DividendCalculationRequest{
		{Ticker: "SBER", DividendPerUnit: Money{Amount: decimal.Must("1.00000000"), Currency: RUB}},
		{Ticker: "SBER", Quantity: decimal.Must("1.00000000"), DividendPerUnit: Money{Currency: RUB}},
		{Ticker: "SBER", Quantity: decimal.Must("1.00000000"), DividendPerUnit: Money{Amount: decimal.Must("1.00000000"), Currency: RUB}, PositionCost: &Money{Currency: RUB}},
	}
	for index, request := range tests {
		if _, err := CalculateDividend(request); err == nil {
			t.Fatalf("case %d: expected fail-closed rejection", index+1)
		}
	}
}

func TestCalculateDividendRejectsDerivedOverflow(t *testing.T) {
	_, err := CalculateDividend(DividendCalculationRequest{
		Ticker:          "SBER",
		Quantity:        decimal.Must("99999999999999999999.99999999"),
		DividendPerUnit: Money{Amount: decimal.Must("2.00000000"), Currency: RUB},
	})
	if err == nil {
		t.Fatal("expected gross dividend overflow rejection")
	}

	tinyCost := Money{Amount: decimal.Must("0.00000001"), Currency: RUB}
	_, err = CalculateDividend(DividendCalculationRequest{
		Ticker:          "SBER",
		Quantity:        decimal.Must("99999999999999999999.99999999"),
		DividendPerUnit: Money{Amount: decimal.Must("1.00000000"), Currency: RUB},
		PositionCost:    &tinyCost,
	})
	if err == nil {
		t.Fatal("expected gross yield overflow rejection")
	}
}

func TestCalculateDividendUsesHalfEvenArithmetic(t *testing.T) {
	result, err := CalculateDividend(DividendCalculationRequest{
		Ticker:          "TEST",
		Quantity:        decimal.Must("0.00000001"),
		DividendPerUnit: Money{Amount: decimal.Must("0.50000000"), Currency: RUB},
	})
	if err != nil {
		t.Fatalf("CalculateDividend: %v", err)
	}
	if got := result.GrossDividend.Amount.String(); got != "0.00000000" {
		t.Fatalf("half-even tie should round to even zero, got %s", got)
	}
}
