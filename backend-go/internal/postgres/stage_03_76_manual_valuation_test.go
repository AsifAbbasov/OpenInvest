package postgres

import (
	"errors"
	"testing"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/position"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestStage376ProjectionDerivesManualValuationWithoutInventingZeroBasisReturn(t *testing.T) {
	projection, err := buildPortfolioPositionsProjectionWithValuations(
		[]rebuiltPortfolioPosition{{
			AssetID:            "asset-zero-basis",
			Ticker:             "ZERO",
			AssetType:          "stock",
			PositionGeneration: 7,
			State: position.State{
				Quantity:            decimal.Must("0.00000001"),
				WeightedAverageCost: decimal.Must("0.00000001"),
				AcquisitionBasis:    decimal.Zero(),
				Open:                true,
			},
		}},
		nil,
		"",
		map[string]manualValuationRecord{
			"asset-zero-basis": {
				AssetID:            "asset-zero-basis",
				PositionGeneration: 7,
				Price:              decimal.Must("2.00000000"),
				Currency:           verticalslice.RUB,
				AsOfDate:           "2026-09-08",
				Source:             verticalslice.ManualValuationSourceUserSupplied,
			},
		},
		decimal.Must("10.00000000"),
	)
	if err != nil {
		t.Fatalf("build Stage 3.76 zero-basis projection: %v", err)
	}
	if len(projection.Items) != 1 {
		t.Fatalf("expected one position, got %d", len(projection.Items))
	}
	market := projection.Items[0].MarketValuation
	if market.Status != verticalslice.MarketValuationAvailableStatus || market.UnrealizedReturn != nil {
		t.Fatalf("zero acquisition basis must keep AVAILABLE valuation with null return: %+v", market)
	}
	if market.MarketValue == nil || market.MarketValue.Amount.String() != "0.00000002" {
		t.Fatalf("market value mismatch: %+v", market.MarketValue)
	}
	if market.UnrealizedGain == nil || market.UnrealizedGain.Amount.String() != "0.00000002" {
		t.Fatalf("unrealized gain mismatch: %+v", market.UnrealizedGain)
	}
	if projection.ValuationSummary.Status != verticalslice.ValuationCoverageCompleteStatus ||
		projection.ValuationSummary.CurrentPortfolioValue == nil ||
		projection.ValuationSummary.CurrentPortfolioValue.Amount.String() != "10.00000002" {
		t.Fatalf("complete zero-basis valuation summary mismatch: %+v", projection.ValuationSummary)
	}
}

func TestStage376ProjectionUsesCanonicalHalfEvenForMarketValue(t *testing.T) {
	projection, err := buildPortfolioPositionsProjectionWithValuations(
		[]rebuiltPortfolioPosition{{
			AssetID:            "asset-rounding",
			Ticker:             "ROUND",
			AssetType:          "stock",
			PositionGeneration: 9,
			State: position.State{
				Quantity:            decimal.Must("0.00000003"),
				WeightedAverageCost: decimal.Must("0.33333333"),
				AcquisitionBasis:    decimal.Must("0.00000001"),
				Open:                true,
			},
		}},
		nil,
		"",
		map[string]manualValuationRecord{
			"asset-rounding": {
				AssetID:            "asset-rounding",
				PositionGeneration: 9,
				Price:              decimal.Must("0.50000000"),
				Currency:           verticalslice.RUB,
				AsOfDate:           "2026-09-08",
				Source:             verticalslice.ManualValuationSourceUserSupplied,
			},
		},
		decimal.Zero(),
	)
	if err != nil {
		t.Fatalf("build Stage 3.76 rounding projection: %v", err)
	}
	market := projection.Items[0].MarketValuation
	if market.MarketValue == nil || market.MarketValue.Amount.String() != "0.00000002" {
		t.Fatalf("0.00000003 * 0.50000000 must Half-Even to 0.00000002, got %+v", market.MarketValue)
	}
	if market.UnrealizedGain == nil || market.UnrealizedGain.Amount.String() != "0.00000001" {
		t.Fatalf("rounded unrealized gain mismatch: %+v", market.UnrealizedGain)
	}
}

func TestStage376ProjectionFailsClosedOnDerivedMarketOverflow(t *testing.T) {
	_, err := buildPortfolioPositionsProjectionWithValuations(
		[]rebuiltPortfolioPosition{{
			AssetID:            "asset-overflow",
			Ticker:             "OVERFLOW",
			AssetType:          "stock",
			PositionGeneration: 13,
			State: position.State{
				Quantity:            decimal.Must("99999999999999999999.00000000"),
				WeightedAverageCost: decimal.Must("1.00000000"),
				AcquisitionBasis:    decimal.Must("99999999999999999999.00000000"),
				Open:                true,
			},
		}},
		nil,
		"",
		map[string]manualValuationRecord{
			"asset-overflow": {
				AssetID:            "asset-overflow",
				PositionGeneration: 13,
				Price:              decimal.Must("2.00000000"),
				Currency:           verticalslice.RUB,
				AsOfDate:           "2026-09-08",
				Source:             verticalslice.ManualValuationSourceUserSupplied,
			},
		},
		decimal.Zero(),
	)
	if !errors.Is(err, verticalslice.ErrInvalidInput) {
		t.Fatalf("derived market overflow must fail closed with invalid input, got %v", err)
	}
}

func TestStage376HistoricalManualValuationRequiresExactDateAndGeneration(t *testing.T) {
	rebuilt := []rebuiltPortfolioPosition{{
		AssetID:            "asset-sber",
		Ticker:             "SBER",
		AssetType:          "stock",
		PositionGeneration: 11,
		State: position.State{
			Quantity:            decimal.Must("10.00000000"),
			WeightedAverageCost: decimal.Must("100.00000000"),
			AcquisitionBasis:    decimal.Must("1000.00000000"),
			Open:                true,
		},
	}}
	records := map[string]manualValuationRecord{
		"asset-sber": {
			AssetID:            "asset-sber",
			PositionGeneration: 11,
			Price:              decimal.Must("120.00000000"),
			Currency:           verticalslice.RUB,
			AsOfDate:           "2026-09-08",
			Source:             verticalslice.ManualValuationSourceUserSupplied,
		},
	}

	exact, err := buildPortfolioPositionsProjectionWithValuations(rebuilt, stringPtr376("2026-09-08"), "2026-09-08", records, decimal.Zero())
	if err != nil || exact.Items[0].MarketValuation.Status != verticalslice.MarketValuationAvailableStatus {
		t.Fatalf("exact historical valuation must be available: projection=%+v err=%v", exact, err)
	}

	mismatch, err := buildPortfolioPositionsProjectionWithValuations(rebuilt, stringPtr376("2026-09-07"), "2026-09-07", records, decimal.Zero())
	if err != nil || mismatch.Items[0].MarketValuation.Status != verticalslice.MarketValuationUnavailableStatus {
		t.Fatalf("date mismatch must remain unavailable: projection=%+v err=%v", mismatch, err)
	}

	rebuilt[0].PositionGeneration = 12
	generationMismatch, err := buildPortfolioPositionsProjectionWithValuations(rebuilt, stringPtr376("2026-09-08"), "2026-09-08", records, decimal.Zero())
	if err != nil || generationMismatch.Items[0].MarketValuation.Status != verticalslice.MarketValuationUnavailableStatus {
		t.Fatalf("generation mismatch must remain unavailable: projection=%+v err=%v", generationMismatch, err)
	}
}

func stringPtr376(value string) *string { return &value }
