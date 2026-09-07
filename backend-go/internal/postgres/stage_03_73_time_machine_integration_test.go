package postgres_test

import "testing"

func TestStage373PortfolioTimeMachineCanonicalHistoricalVectors(t *testing.T) {
	h := newStage371Harness(t, "Stage 3.73 portfolio time machine vectors")

	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "100.00000000", "250.00000000", "2026-01-10"))
	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "100.00000000", "300.00000000", "2026-03-01"))
	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "SELL", "SBER", "50.00000000", "350.00000000", "2026-06-01"))

	cases := []struct {
		asOfDate string
		quantity string
		wac      string
		basis    string
	}{
		{asOfDate: "2026-02-01", quantity: "100.00000000", wac: "250.00000000", basis: "25000.00000000"},
		{asOfDate: "2026-04-01", quantity: "200.00000000", wac: "275.00000000", basis: "55000.00000000"},
		{asOfDate: "2026-07-01", quantity: "150.00000000", wac: "275.00000000", basis: "41250.00000000"},
	}

	for _, tc := range cases {
		t.Run(tc.asOfDate, func(t *testing.T) {
			projection, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, tc.asOfDate)
			if err != nil {
				t.Fatalf("get historical projection: %v", err)
			}
			if len(projection.Items) != 1 {
				t.Fatalf("expected one historical position, got %d", len(projection.Items))
			}
			position := projection.Items[0]
			if position.Ticker != "SBER" ||
				position.Quantity.String() != tc.quantity ||
				position.WeightedAverageCost.Amount.String() != tc.wac ||
				position.AcquisitionBasis.Amount.String() != tc.basis {
				t.Fatalf("historical vector mismatch for %s: %+v", tc.asOfDate, position)
			}
			if projection.TotalAcquisitionBasis.Amount.String() != tc.basis {
				t.Fatalf("historical total basis got %s want %s", projection.TotalAcquisitionBasis.Amount.String(), tc.basis)
			}
			if projection.InputsAsOf == nil || *projection.InputsAsOf != tc.asOfDate {
				t.Fatalf("historical inputsAsOf got %v want %s", projection.InputsAsOf, tc.asOfDate)
			}
			if position.MarketValuation.Status != "UNAVAILABLE" ||
				position.MarketValuation.Reason != "NO_APPROVED_MARKET_PRICE_SOURCE" {
				t.Fatalf("historical market state must remain unavailable: %+v", position.MarketValuation)
			}

			repeated, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, tc.asOfDate)
			if err != nil {
				t.Fatalf("repeat historical projection: %v", err)
			}
			if len(repeated.Items) != 1 ||
				repeated.Items[0].Quantity.String() != position.Quantity.String() ||
				repeated.Items[0].WeightedAverageCost.Amount.String() != position.WeightedAverageCost.Amount.String() ||
				repeated.Items[0].AcquisitionBasis.Amount.String() != position.AcquisitionBasis.Amount.String() {
				t.Fatalf("repeated historical projection is not deterministic: first=%+v repeated=%+v", position, repeated.Items)
			}
		})
	}
}

func TestStage373PortfolioTimeMachineShowsHistoricallyOpenCurrentlyClosedPosition(t *testing.T) {
	h := newStage371Harness(t, "Stage 3.73 historically open currently closed")

	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "5.00000000", "100.00000000", "2026-01-10"))
	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "SELL", "SBER", "5.00000000", "150.00000000", "2026-05-01"))

	current, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "")
	if err != nil {
		t.Fatalf("get current projection: %v", err)
	}
	if len(current.Items) != 0 {
		t.Fatalf("fully closed current position must be absent: %+v", current.Items)
	}

	historical, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "2026-03-01")
	if err != nil {
		t.Fatalf("get historical projection: %v", err)
	}
	if len(historical.Items) != 1 ||
		historical.Items[0].Ticker != "SBER" ||
		historical.Items[0].Quantity.String() != "5.00000000" ||
		historical.Items[0].WeightedAverageCost.Amount.String() != "100.00000000" ||
		historical.Items[0].AcquisitionBasis.Amount.String() != "500.00000000" {
		t.Fatalf("historically open position must be reconstructed: %+v", historical.Items)
	}
}
