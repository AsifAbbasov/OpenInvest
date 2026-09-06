package position

import (
	"errors"
	"testing"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
)

func testDecimal(value string) decimal.Decimal {
	return decimal.Must(value)
}

func testTrade(kind string, quantity string, price string) Trade {
	return Trade{Type: kind, Quantity: testDecimal(quantity), UnitPrice: testDecimal(price)}
}

func TestCanonicalLifecycle(t *testing.T) {
	state, err := Rebuild([]Trade{
		testTrade("BUY", "100.00000000", "250.00000000"),
		testTrade("BUY", "100.00000000", "300.00000000"),
		testTrade("SELL", "50.00000000", "350.00000000"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := state.Quantity.String(), "150.00000000"; got != want {
		t.Fatalf("quantity %s want %s", got, want)
	}
	if got, want := state.WeightedAverageCost.String(), "275.00000000"; got != want {
		t.Fatalf("WAC %s want %s", got, want)
	}
	if got, want := state.AcquisitionBasis.String(), "41250.00000000"; got != want {
		t.Fatalf("basis %s want %s", got, want)
	}

	state, err = Apply(state, testTrade("SELL", "150.00000000", "360.00000000"))
	if err != nil {
		t.Fatal(err)
	}
	if state.Open || !state.Quantity.IsZero() {
		t.Fatalf("expected closed state, got %+v", state)
	}

	state, err = Apply(state, testTrade("BUY", "20.00000000", "310.00000000"))
	if err != nil {
		t.Fatal(err)
	}
	if state.WeightedAverageCost.String() != "310.00000000" || state.AcquisitionBasis.String() != "6200.00000000" {
		t.Fatalf("unexpected reopen state: %+v", state)
	}
}

func TestPartialSellPreservesAuthoritativeWAC(t *testing.T) {
	state := State{
		Quantity:            testDecimal("0.40000000"),
		WeightedAverageCost: testDecimal("275.12345678"),
		AcquisitionBasis:    testDecimal("110.04938271"),
		Open:                true,
	}

	got, err := Apply(state, testTrade("SELL", "0.20000000", "999.00000000"))
	if err != nil {
		t.Fatal(err)
	}
	if got.WeightedAverageCost.String() != "275.12345678" {
		t.Fatalf("WAC repriced after partial SELL: %s", got.WeightedAverageCost.String())
	}
	if got.AcquisitionBasis.String() != "55.02469136" {
		t.Fatalf("basis %s", got.AcquisitionBasis.String())
	}
	repriced, err := got.AcquisitionBasis.Div(got.Quantity)
	if err != nil {
		t.Fatal(err)
	}
	if repriced.String() != "275.12345680" {
		t.Fatalf("fractional non-invertibility witness changed: %s", repriced.String())
	}
}

func TestOversellFailsClosed(t *testing.T) {
	_, err := Rebuild([]Trade{
		testTrade("BUY", "1.00000000", "100.00000000"),
		testTrade("SELL", "1.00000001", "90.00000000"),
	})
	if !errors.Is(err, ErrInsufficientQuantity) {
		t.Fatalf("got %v", err)
	}
}

func TestDerivedOverflowFailsClosed(t *testing.T) {
	_, err := Rebuild([]Trade{
		testTrade("BUY", "99999999999999999999.99999999", "99999999999999999999.99999999"),
	})
	if !errors.Is(err, ErrDerivedOverflow) {
		t.Fatalf("got %v", err)
	}
}

func TestRebuildIsDeterministic(t *testing.T) {
	trades := []Trade{
		testTrade("BUY", "3.00000000", "100.00000000"),
		testTrade("BUY", "2.00000000", "101.00000000"),
		testTrade("SELL", "1.25000000", "120.00000000"),
	}
	first, err := Rebuild(trades)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Rebuild(trades)
	if err != nil {
		t.Fatal(err)
	}
	if first.Quantity.String() != second.Quantity.String() ||
		first.WeightedAverageCost.String() != second.WeightedAverageCost.String() ||
		first.AcquisitionBasis.String() != second.AcquisitionBasis.String() {
		t.Fatalf("nondeterministic rebuild: first=%+v second=%+v", first, second)
	}
}
