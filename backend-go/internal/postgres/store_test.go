package postgres

import "testing"

func TestApprovedAssetFixtureUsesFrozenStage2Identity(t *testing.T) {
	expected := map[string]assetFixture{
		"SBER": {
			ID:        "00000000-0000-4000-8000-00000000a001",
			Ticker:    "SBER",
			AssetType: "stock",
			Name:      "Sberbank ordinary shares",
			ISIN:      stringPtr("RU0009029540"),
			LotSize:   "10.00000000",
		},
		"GAZP": {
			ID:        "00000000-0000-4000-8000-00000000a002",
			Ticker:    "GAZP",
			AssetType: "stock",
			Name:      "Gazprom ordinary shares",
			ISIN:      stringPtr("RU0007661625"),
			LotSize:   "10.00000000",
		},
		"SU26238RMFS4": {
			ID:        "00000000-0000-4000-8000-00000000b001",
			Ticker:    "SU26238RMFS4",
			AssetType: "bond",
			Name:      "OFZ 26238",
			ISIN:      stringPtr("RU000A1038V6"),
			LotSize:   "1.00000000",
		},
	}
	if len(approvedAssetFixtures) != len(expected) {
		t.Fatalf("expected %d fixtures, got %d", len(expected), len(approvedAssetFixtures))
	}
	for ticker, want := range expected {
		got, ok := approvedAssetFixture(ticker)
		if !ok {
			t.Fatalf("expected %s fixture", ticker)
		}
		if got.ID != want.ID ||
			got.Ticker != want.Ticker ||
			got.AssetType != want.AssetType ||
			got.Name != want.Name ||
			got.LotSize != want.LotSize ||
			(got.ISIN == nil || want.ISIN == nil || *got.ISIN != *want.ISIN) {
			t.Fatalf("unexpected %s fixture: got %+v want %+v", ticker, got, want)
		}
	}
}

func TestApprovedAssetFixtureRejectsUnknownTicker(t *testing.T) {
	if _, ok := approvedAssetFixture("UNKNOWN1"); ok {
		t.Fatal("expected unknown ticker to be outside the approved fixture catalog")
	}
	if _, ok := approvedAssetFixture(" SBER "); ok {
		t.Fatal("expected noncanonical ticker spelling to be outside the approved fixture catalog")
	}
}
