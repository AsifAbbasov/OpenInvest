package postgres

import (
	"errors"
	"testing"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/position"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestStage372BuildProjectionCanonicalPartialSell(t *testing.T) {
	state := rebuildState(t,
		position.Trade{Type: "BUY", Quantity: decimal.Must("100.00000000"), UnitPrice: decimal.Must("250.00000000")},
		position.Trade{Type: "BUY", Quantity: decimal.Must("100.00000000"), UnitPrice: decimal.Must("300.00000000")},
		position.Trade{Type: "SELL", Quantity: decimal.Must("50.00000000"), UnitPrice: decimal.Must("999.00000000")},
	)
	inputsAsOf := "2026-09-03"
	projection, err := buildPortfolioPositionsProjection([]rebuiltPortfolioPosition{{
		AssetID:   "asset-sber",
		Ticker:    "SBER",
		AssetType: "stock",
		State:     state,
	}}, &inputsAsOf)
	if err != nil {
		t.Fatalf("build canonical projection: %v", err)
	}
	if len(projection.Items) != 1 {
		t.Fatalf("expected one open position, got %d", len(projection.Items))
	}
	item := projection.Items[0]
	if item.Quantity.String() != "150.00000000" ||
		item.WeightedAverageCost.Amount.String() != "275.00000000" ||
		item.AcquisitionBasis.Amount.String() != "41250.00000000" {
		t.Fatalf("canonical vector drifted: quantity=%s wac=%s basis=%s",
			item.Quantity.String(), item.WeightedAverageCost.Amount.String(), item.AcquisitionBasis.Amount.String())
	}
	if projection.TotalAcquisitionBasis.Amount.String() != "41250.00000000" {
		t.Fatalf("total basis mismatch: %s", projection.TotalAcquisitionBasis.Amount.String())
	}
	if item.AcquisitionBasisWeight == nil || item.AcquisitionBasisWeight.String() != "1.00000000" {
		t.Fatalf("single open position weight must be 1.00000000, got %v", item.AcquisitionBasisWeight)
	}
	assertUnavailableMarketValuation(t, item)
	if projection.InputsAsOf == nil || *projection.InputsAsOf != inputsAsOf {
		t.Fatalf("inputsAsOf mismatch: %v", projection.InputsAsOf)
	}
	if projection.MethodologyVersion != verticalslice.PortfolioPositionProjectionMethodologyVersion {
		t.Fatalf("methodology mismatch: %s", projection.MethodologyVersion)
	}
}

func TestStage372BuildProjectionZeroBasisLeavesAllocationUndefined(t *testing.T) {
	state := rebuildState(t, position.Trade{
		Type:      "BUY",
		Quantity:  decimal.Must("0.00000001"),
		UnitPrice: decimal.Must("0.00000001"),
	})
	if state.AcquisitionBasis.String() != "0.00000000" || !state.Open {
		t.Fatalf("expected valid open zero-basis witness, state=%+v", state)
	}

	projection, err := buildPortfolioPositionsProjection([]rebuiltPortfolioPosition{{
		AssetID:   "asset-zero",
		Ticker:    "ZERO",
		AssetType: "stock",
		State:     state,
	}}, nil)
	if err != nil {
		t.Fatalf("build zero-basis projection: %v", err)
	}
	if projection.TotalAcquisitionBasis.Amount.String() != "0.00000000" {
		t.Fatalf("zero-basis total mismatch: %s", projection.TotalAcquisitionBasis.Amount.String())
	}
	if len(projection.Items) != 1 || projection.Items[0].AcquisitionBasisWeight != nil {
		t.Fatalf("0/0 acquisition-basis allocation must be null: %+v", projection.Items)
	}
}

func TestStage372BuildProjectionHalfEvenWeightAndNoForceNormalization(t *testing.T) {
	projection, err := buildPortfolioPositionsProjection([]rebuiltPortfolioPosition{
		openRebuilt("asset-a", "AAA", "stock", "1.00000000", "1.00000000", "1.00000000"),
		openRebuilt("asset-b", "BBB", "stock", "1.00000000", "1.00000000", "1.00000000"),
		openRebuilt("asset-c", "CCC", "bond", "1.00000000", "1.00000000", "1.00000000"),
	}, nil)
	if err != nil {
		t.Fatalf("build thirds projection: %v", err)
	}
	for _, item := range projection.Items {
		if item.AcquisitionBasisWeight == nil || item.AcquisitionBasisWeight.String() != "0.33333333" {
			t.Fatalf("expected independent 8dp weight, got %v", item.AcquisitionBasisWeight)
		}
	}
	// Three independently rounded thirds sum to 0.99999999. The final item must not be altered to 0.33333334.
	if projection.Items[2].AcquisitionBasisWeight.String() != "0.33333333" {
		t.Fatalf("final weight was force-normalized: %s", projection.Items[2].AcquisitionBasisWeight.String())
	}

	halfEven, err := buildPortfolioPositionsProjection([]rebuiltPortfolioPosition{
		openRebuilt("asset-a", "AAA", "stock", "1.00000000", "24691357.00000000", "24691357.00000000"),
		openRebuilt("asset-b", "BBB", "stock", "1.00000000", "175308643.00000000", "175308643.00000000"),
	}, nil)
	if err != nil {
		t.Fatalf("build Half-Even projection: %v", err)
	}
	if halfEven.Items[0].AcquisitionBasisWeight == nil || halfEven.Items[0].AcquisitionBasisWeight.String() != "0.12345678" {
		t.Fatalf("Half-Even tie must keep even final digit, got %v", halfEven.Items[0].AcquisitionBasisWeight)
	}
}

func TestStage372BuildProjectionSortsTickerAndExcludesClosedPositions(t *testing.T) {
	closed := position.Empty()
	projection, err := buildPortfolioPositionsProjection([]rebuiltPortfolioPosition{
		openRebuilt("asset-z", "ZZZ", "bond", "1.00000000", "10.00000000", "10.00000000"),
		{AssetID: "asset-closed", Ticker: "CCC", AssetType: "stock", State: closed},
		openRebuilt("asset-a", "AAA", "stock", "2.00000000", "10.00000000", "20.00000000"),
	}, nil)
	if err != nil {
		t.Fatalf("build ordered projection: %v", err)
	}
	if len(projection.Items) != 2 || projection.Items[0].Ticker != "AAA" || projection.Items[1].Ticker != "ZZZ" {
		t.Fatalf("public items must be ticker ASC with closed positions excluded: %+v", projection.Items)
	}
}

func TestStage372BuildProjectionSingleBuyMultipleBondsAndMaximumDecimal(t *testing.T) {
	single := rebuildState(t, position.Trade{
		Type: "BUY", Quantity: decimal.Must("7.50000000"), UnitPrice: decimal.Must("123.45678901"),
	})
	projection, err := buildPortfolioPositionsProjection([]rebuiltPortfolioPosition{{
		AssetID: "asset-single", Ticker: "SBER", AssetType: "stock", State: single,
	}}, nil)
	if err != nil {
		t.Fatalf("build single-BUY projection: %v", err)
	}
	if len(projection.Items) != 1 || projection.Items[0].Quantity.String() != "7.50000000" ||
		projection.Items[0].WeightedAverageCost.Amount.String() != "123.45678901" ||
		projection.Items[0].AcquisitionBasis.Amount.String() != "925.92591758" {
		t.Fatalf("single BUY projection drifted: %+v", projection.Items)
	}

	multipleBonds, err := buildPortfolioPositionsProjection([]rebuiltPortfolioPosition{
		openRebuilt("bond-b", "BOND2", "bond", "2.00000000", "1000.00000000", "2000.00000000"),
		openRebuilt("bond-a", "BOND1", "bond", "1.00000000", "500.00000000", "500.00000000"),
	}, nil)
	if err != nil {
		t.Fatalf("build multiple-BOND projection: %v", err)
	}
	if len(multipleBonds.Items) != 2 || multipleBonds.Items[0].Ticker != "BOND1" || multipleBonds.Items[1].Ticker != "BOND2" ||
		multipleBonds.TotalAcquisitionBasis.Amount.String() != "2500.00000000" {
		t.Fatalf("multiple BOND projection drifted: %+v", multipleBonds)
	}

	maximum := decimal.Must("99999999999999999999.99999999")
	maximumProjection, err := buildPortfolioPositionsProjection([]rebuiltPortfolioPosition{{
		AssetID: "asset-max", Ticker: "MAX", AssetType: "stock", State: position.State{
			Quantity: decimal.Must("1.00000000"), WeightedAverageCost: maximum, AcquisitionBasis: maximum, Open: true,
		},
	}}, nil)
	if err != nil {
		t.Fatalf("maximum storage-compatible position must remain representable: %v", err)
	}
	if maximumProjection.TotalAcquisitionBasis.Amount.String() != maximum.String() ||
		maximumProjection.Items[0].AcquisitionBasisWeight == nil ||
		maximumProjection.Items[0].AcquisitionBasisWeight.String() != "1.00000000" {
		t.Fatalf("maximum Decimal projection drifted: %+v", maximumProjection)
	}
}

func TestStage372BuildProjectionFailsClosedOnUnsupportedAssetAndAggregateOverflow(t *testing.T) {
	_, err := buildPortfolioPositionsProjection([]rebuiltPortfolioPosition{
		openRebuilt("asset-future", "FUT", "future", "1.00000000", "1.00000000", "1.00000000"),
	}, nil)
	if !errors.Is(err, verticalslice.ErrInvalidInput) {
		t.Fatalf("unsupported asset type must fail closed, got %v", err)
	}

	maximum := decimal.Must("99999999999999999999.99999999")
	state := position.State{
		Quantity:            decimal.Must("1.00000000"),
		WeightedAverageCost: maximum,
		AcquisitionBasis:    maximum,
		Open:                true,
	}
	_, err = buildPortfolioPositionsProjection([]rebuiltPortfolioPosition{
		{AssetID: "asset-a", Ticker: "AAA", AssetType: "stock", State: state},
		{AssetID: "asset-b", Ticker: "BBB", AssetType: "bond", State: state},
	}, nil)
	if !errors.Is(err, verticalslice.ErrInvalidInput) {
		t.Fatalf("aggregate overflow must fail closed, got %v", err)
	}
}

func TestStage372BuildProjectionKeepsFractionalAuthoritativeWAC(t *testing.T) {
	state := position.State{
		Quantity:            decimal.Must("0.20000000"),
		WeightedAverageCost: decimal.Must("275.12345678"),
		AcquisitionBasis:    decimal.Must("55.02469136"),
		Open:                true,
	}
	projection, err := buildPortfolioPositionsProjection([]rebuiltPortfolioPosition{{
		AssetID: "asset-sber", Ticker: "SBER", AssetType: "stock", State: state,
	}}, nil)
	if err != nil {
		t.Fatalf("build fractional projection: %v", err)
	}
	item := projection.Items[0]
	if item.WeightedAverageCost.Amount.String() != "275.12345678" || item.AcquisitionBasis.Amount.String() != "55.02469136" {
		t.Fatalf("projection must preserve authoritative WAC/basis without division re-derivation: %+v", item)
	}
}

func rebuildState(t *testing.T, trades ...position.Trade) position.State {
	t.Helper()
	state, err := position.Rebuild(trades)
	if err != nil {
		t.Fatalf("rebuild test position: %v", err)
	}
	return state
}

func openRebuilt(assetID, ticker, assetType, quantity, wac, basis string) rebuiltPortfolioPosition {
	return rebuiltPortfolioPosition{
		AssetID:   assetID,
		Ticker:    ticker,
		AssetType: assetType,
		State: position.State{
			Quantity:            decimal.Must(quantity),
			WeightedAverageCost: decimal.Must(wac),
			AcquisitionBasis:    decimal.Must(basis),
			Open:                true,
		},
	}
}

func assertUnavailableMarketValuation(t *testing.T, item verticalslice.PortfolioPositionProjection) {
	t.Helper()
	if item.MarketValuation.Status != verticalslice.MarketValuationUnavailableStatus ||
		item.MarketValuation.Reason != verticalslice.MarketValuationUnavailableReason {
		t.Fatalf("market valuation contract drifted: %+v", item.MarketValuation)
	}
}
