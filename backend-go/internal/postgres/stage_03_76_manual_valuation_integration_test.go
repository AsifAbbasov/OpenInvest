package postgres_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func cleanupStage376Valuations(t *testing.T, h stage371Harness) {
	t.Helper()
	t.Cleanup(func() {
		if _, err := h.db.ExecContext(h.ctx, `DELETE FROM investment.portfolio_manual_valuations WHERE portfolio_id = $1`, h.portfolioID); err != nil {
			t.Fatalf("cleanup Stage 3.76 manual valuations: %v", err)
		}
	})
}

func appendStage376Deposit(t *testing.T, h stage371Harness, amount, tradeDate string) {
	t.Helper()
	gross := verticalslice.Money{Amount: decimal.Must(amount), Currency: verticalslice.RUB}
	_, err := h.service.AppendTransaction(
		h.ctx,
		verticalslice.RequestContext{},
		h.subjectID,
		uuid.NewString(),
		"/api/v1/portfolios/"+h.portfolioID+"/transactions",
		verticalslice.AppendTransactionRequest{
			PortfolioID:     h.portfolioID,
			TransactionType: "DEPOSIT",
			GrossAmount:     &gross,
			Commission:      verticalslice.ZeroMoney(),
			Tax:             verticalslice.ZeroMoney(),
			TradeDate:       tradeDate,
		},
	)
	if err != nil {
		t.Fatalf("append Stage 3.76 deposit: %v", err)
	}
}

func upsertStage376(t *testing.T, h stage371Harness, ticker, price, asOfDate string) verticalslice.PortfolioPositionsProjection {
	t.Helper()
	projection, err := h.service.UpsertManualValuation(h.ctx, h.subjectID, verticalslice.ManualValuationRequest{
		PortfolioID: h.portfolioID,
		Ticker:      ticker,
		Price:       verticalslice.Money{Amount: decimal.Must(price), Currency: verticalslice.RUB},
		AsOfDate:    asOfDate,
	})
	if err != nil {
		t.Fatalf("upsert Stage 3.76 %s valuation: %v", ticker, err)
	}
	return projection
}

func findStage376Position(t *testing.T, projection verticalslice.PortfolioPositionsProjection, ticker string) verticalslice.PortfolioPositionProjection {
	t.Helper()
	for _, item := range projection.Items {
		if item.Ticker == ticker {
			return item
		}
	}
	t.Fatalf("position %s not found: %+v", ticker, projection.Items)
	return verticalslice.PortfolioPositionProjection{}
}

func TestStage376ManualValuationCoverageCashAndUpdateClear(t *testing.T) {
	h := newStage371Harness(t, "Stage 3.76 valuation coverage")
	cleanupStage376Valuations(t, h)
	appendStage376Deposit(t, h, "100000.00000000", "2026-09-01")
	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "50.00000000", "250.00000000", "2026-09-01"))
	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "100.00000000", "287.50000000", "2026-09-01"))
	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "GAZP", "10.00000000", "100.00000000", "2026-09-02"))

	initial, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "")
	if err != nil {
		t.Fatalf("read initial Stage 3.76 projection: %v", err)
	}
	if initial.ValuationSummary.Status != verticalslice.ValuationCoveragePartialStatus ||
		initial.ValuationSummary.ValuedPositions != 0 || initial.ValuationSummary.TotalOpenPositions != 2 ||
		initial.ValuationSummary.CurrentPortfolioValue != nil {
		t.Fatalf("initial partial coverage mismatch: %+v", initial.ValuationSummary)
	}
	initialSBER := findStage376Position(t, initial, "SBER")
	if initialSBER.Quantity.String() != "150.00000000" ||
		initialSBER.WeightedAverageCost.Amount.String() != "275.00000000" ||
		initialSBER.AcquisitionBasis.Amount.String() != "41250.00000000" {
		t.Fatalf("multi-BUY WAC must remain canonical before valuation: %+v", initialSBER)
	}

	partial := upsertStage376(t, h, "SBER", "318.50000000", "2026-09-08")
	sber := findStage376Position(t, partial, "SBER")
	market := sber.MarketValuation
	if market.Status != verticalslice.MarketValuationAvailableStatus || market.Source != verticalslice.ManualValuationSourceUserSupplied ||
		market.MarketPrice == nil || market.MarketPrice.Amount.String() != "318.50000000" ||
		market.MarketValue == nil || market.MarketValue.Amount.String() != "47775.00000000" ||
		market.UnrealizedGain == nil || market.UnrealizedGain.Amount.String() != "6525.00000000" ||
		market.UnrealizedReturn == nil || market.UnrealizedReturn.String() != "0.15818182" ||
		market.AsOf == nil || *market.AsOf != "2026-09-08" {
		t.Fatalf("SBER manual valuation mismatch: %+v", market)
	}
	if partial.ValuationSummary.Status != verticalslice.ValuationCoveragePartialStatus ||
		partial.ValuationSummary.ValuedPositions != 1 || partial.ValuationSummary.TotalOpenPositions != 2 ||
		partial.ValuationSummary.ValuedPositionsMarketValue.Amount.String() != "47775.00000000" ||
		partial.ValuationSummary.ValuedPositionsAcquisitionBasis.Amount.String() != "41250.00000000" ||
		partial.ValuationSummary.TotalAcquisitionBasis.Amount.String() != "42250.00000000" ||
		partial.ValuationSummary.UnrealizedGain.Amount.String() != "6525.00000000" ||
		partial.ValuationSummary.CashValue.Amount.String() != "57750.00000000" ||
		partial.ValuationSummary.CurrentPortfolioValue != nil {
		t.Fatalf("partial valuation summary mismatch: %+v", partial.ValuationSummary)
	}

	complete := upsertStage376(t, h, "GAZP", "110.00000000", "2026-09-07")
	if complete.ValuationSummary.Status != verticalslice.ValuationCoverageCompleteStatus ||
		complete.ValuationSummary.ValuedPositions != 2 ||
		complete.ValuationSummary.ValuedPositionsMarketValue.Amount.String() != "48875.00000000" ||
		complete.ValuationSummary.UnrealizedGain.Amount.String() != "6625.00000000" ||
		complete.ValuationSummary.CurrentPortfolioValue == nil ||
		complete.ValuationSummary.CurrentPortfolioValue.Amount.String() != "106625.00000000" {
		t.Fatalf("complete valuation summary mismatch: %+v", complete.ValuationSummary)
	}
	if findStage376Position(t, complete, "SBER").MarketValuation.MarketWeight == nil ||
		findStage376Position(t, complete, "SBER").MarketValuation.MarketWeight.String() != "0.97749361" {
		t.Fatalf("SBER valued-position allocation mismatch: %+v", findStage376Position(t, complete, "SBER").MarketValuation)
	}

	updated := upsertStage376(t, h, "SBER", "300.00000000", "2026-09-09")
	updatedSBER := findStage376Position(t, updated, "SBER").MarketValuation
	if updatedSBER.MarketValue == nil || updatedSBER.MarketValue.Amount.String() != "45000.00000000" ||
		updatedSBER.UnrealizedGain == nil || updatedSBER.UnrealizedGain.Amount.String() != "3750.00000000" ||
		updatedSBER.AsOf == nil || *updatedSBER.AsOf != "2026-09-09" {
		t.Fatalf("future-dated manual valuation update must overwrite current state without hidden wall-clock rejection: %+v", updatedSBER)
	}

	cleared, err := h.service.ClearManualValuation(h.ctx, h.subjectID, h.portfolioID, "SBER")
	if err != nil {
		t.Fatalf("clear Stage 3.76 valuation: %v", err)
	}
	if findStage376Position(t, cleared, "SBER").MarketValuation.Status != verticalslice.MarketValuationUnavailableStatus ||
		cleared.ValuationSummary.Status != verticalslice.ValuationCoveragePartialStatus {
		t.Fatalf("clear must restore honest unavailable state: %+v", cleared)
	}
	if _, err := h.service.ClearManualValuation(h.ctx, h.subjectID, h.portfolioID, "SBER"); err != nil {
		t.Fatalf("repeated clear must be idempotent: %v", err)
	}
}

func TestStage376LifecycleCorrectionReversalHistoricalAndReopenGuard(t *testing.T) {
	h := newStage371Harness(t, "Stage 3.76 lifecycle")
	cleanupStage376Valuations(t, h)
	buy := appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "10.00000000", "100.00000000", "2026-09-01"))
	upsertStage376(t, h, "SBER", "120.00000000", "2026-09-08")

	exact, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "2026-09-08")
	if err != nil || findStage376Position(t, exact, "SBER").MarketValuation.Status != verticalslice.MarketValuationAvailableStatus {
		t.Fatalf("exact historical date must use eligible manual valuation: projection=%+v err=%v", exact, err)
	}
	mismatch, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "2026-09-07")
	if err != nil || findStage376Position(t, mismatch, "SBER").MarketValuation.Status != verticalslice.MarketValuationUnavailableStatus {
		t.Fatalf("historical date mismatch must not carry valuation: projection=%+v err=%v", mismatch, err)
	}

	corrected, err := correctStage374(t, h, buy, "10.00000000", "110.00000000", "2026-09-01", uuid.NewString())
	if err != nil {
		t.Fatalf("correct Stage 3.76 BUY: %v", err)
	}
	afterCorrection, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "")
	if err != nil {
		t.Fatalf("projection after correction: %v", err)
	}
	correctedMarket := findStage376Position(t, afterCorrection, "SBER").MarketValuation
	if correctedMarket.Status != verticalslice.MarketValuationAvailableStatus || correctedMarket.MarketValue == nil ||
		correctedMarket.MarketValue.Amount.String() != "1200.00000000" || correctedMarket.UnrealizedGain == nil ||
		correctedMarket.UnrealizedGain.Amount.String() != "100.00000000" {
		t.Fatalf("correction must preserve generation and recompute P/L: %+v", correctedMarket)
	}

	sell := appendStage371Trade(t, h, stage371Trade(h.portfolioID, "SELL", "SBER", "2.00000000", "130.00000000", "2026-09-09"))
	afterSell, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "")
	if err != nil {
		t.Fatalf("projection after partial sell: %v", err)
	}
	partialSellMarket := findStage376Position(t, afterSell, "SBER").MarketValuation
	if partialSellMarket.Status != verticalslice.MarketValuationAvailableStatus || partialSellMarket.MarketValue == nil ||
		partialSellMarket.MarketValue.Amount.String() != "960.00000000" || partialSellMarket.UnrealizedGain == nil ||
		partialSellMarket.UnrealizedGain.Amount.String() != "80.00000000" {
		t.Fatalf("partial sell must preserve generation and recompute P/L: %+v", partialSellMarket)
	}
	if _, err := reverseStage374(t, h, sell, "2026-09-10", uuid.NewString()); err != nil {
		t.Fatalf("reverse partial sell: %v", err)
	}
	afterReversal, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "")
	if err != nil || findStage376Position(t, afterReversal, "SBER").MarketValuation.MarketValue == nil ||
		findStage376Position(t, afterReversal, "SBER").MarketValuation.MarketValue.Amount.String() != "1200.00000000" {
		t.Fatalf("reversal must restore effective quantity under same generation: projection=%+v err=%v", afterReversal, err)
	}

	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "SELL", "SBER", "10.00000000", "125.00000000", "2026-09-11"))
	closed, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "")
	if err != nil || len(closed.Items) != 0 {
		t.Fatalf("full close must remove open position: projection=%+v err=%v", closed, err)
	}
	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "2.00000000", "200.00000000", "2026-09-12"))
	reopened, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "")
	if err != nil {
		t.Fatalf("projection after reopen: %v", err)
	}
	if findStage376Position(t, reopened, "SBER").MarketValuation.Status != verticalslice.MarketValuationUnavailableStatus {
		t.Fatalf("old generation valuation must not reactivate after reopen: %+v", reopened)
	}
	if _, err := h.service.ClearManualValuation(h.ctx, h.subjectID, h.portfolioID, "SBER"); err != nil {
		t.Fatalf("clear stale valuation after reopen must succeed: %v", err)
	}
	fresh := upsertStage376(t, h, "SBER", "250.00000000", "2026-09-12")
	freshMarket := findStage376Position(t, fresh, "SBER").MarketValuation
	if freshMarket.Status != verticalslice.MarketValuationAvailableStatus || freshMarket.MarketValue == nil ||
		freshMarket.MarketValue.Amount.String() != "500.00000000" {
		t.Fatalf("fresh reopened generation valuation mismatch: %+v", freshMarket)
	}

	_ = corrected
}

func TestStage376RejectsInvalidPriceTickerAndForeignPortfolio(t *testing.T) {
	h := newStage371Harness(t, "Stage 3.76 validation")
	cleanupStage376Valuations(t, h)
	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "1.00000000", "100.00000000", "2026-09-01"))

	cases := []verticalslice.ManualValuationRequest{
		{PortfolioID: h.portfolioID, Ticker: "SBER", Price: verticalslice.Money{Amount: decimal.Zero(), Currency: verticalslice.RUB}, AsOfDate: "2026-09-08"},
		{PortfolioID: h.portfolioID, Ticker: "SBER", Price: verticalslice.Money{Amount: decimal.Must("-1.00000000"), Currency: verticalslice.RUB}, AsOfDate: "2026-09-08"},
		{PortfolioID: h.portfolioID, Ticker: "SBER", Price: verticalslice.Money{Amount: decimal.Must("1.00000000"), Currency: "USD"}, AsOfDate: "2026-09-08"},
		{PortfolioID: h.portfolioID, Ticker: "SBER", Price: verticalslice.Money{Amount: decimal.Must("1.00000000"), Currency: verticalslice.RUB}, AsOfDate: "2026-02-30"},
	}
	for _, request := range cases {
		if _, err := h.service.UpsertManualValuation(h.ctx, h.subjectID, request); !errors.Is(err, verticalslice.ErrInvalidInput) {
			t.Fatalf("invalid manual valuation must be rejected: request=%+v err=%v", request, err)
		}
	}
	if _, err := h.service.UpsertManualValuation(h.ctx, h.subjectID, verticalslice.ManualValuationRequest{
		PortfolioID: h.portfolioID,
		Ticker:      "GAZP",
		Price:       verticalslice.Money{Amount: decimal.Must("100.00000000"), Currency: verticalslice.RUB},
		AsOfDate:    "2026-09-08",
	}); !errors.Is(err, verticalslice.ErrNotFound) {
		t.Fatalf("ticker without open position must be hidden as not found: %v", err)
	}
	if _, err := h.service.UpsertManualValuation(h.ctx, uuid.NewString(), verticalslice.ManualValuationRequest{
		PortfolioID: h.portfolioID,
		Ticker:      "SBER",
		Price:       verticalslice.Money{Amount: decimal.Must("100.00000000"), Currency: verticalslice.RUB},
		AsOfDate:    "2026-09-08",
	}); !errors.Is(err, postgres.ErrNotFound) {
		t.Fatalf("foreign subject must not enumerate portfolio valuation: %v", err)
	}
}
