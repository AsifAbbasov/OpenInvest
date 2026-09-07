package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

type stage376HTTPStore struct {
	importAPITestStore
	projection  verticalslice.PortfolioPositionsProjection
	upsert      verticalslice.ManualValuationRequest
	upsertCalls int
	clearTicker string
	clearCalls  int
}

func (store *stage376HTTPStore) GetPortfolioPositions(
	_ context.Context,
	_ string,
	_ string,
	_ string,
) (verticalslice.PortfolioPositionsProjection, error) {
	return store.projection, nil
}

func (store *stage376HTTPStore) UpsertManualValuation(
	_ context.Context,
	_ string,
	request verticalslice.ManualValuationRequest,
) error {
	store.upsertCalls++
	store.upsert = request
	return nil
}

func (store *stage376HTTPStore) ClearManualValuation(
	_ context.Context,
	_ string,
	_ string,
	ticker string,
) error {
	store.clearCalls++
	store.clearTicker = ticker
	return nil
}

func stage376HTTPProjection(available bool) verticalslice.PortfolioPositionsProjection {
	market := verticalslice.MarketValuationUnavailable{
		Status: verticalslice.MarketValuationUnavailableStatus,
		Reason: verticalslice.MarketValuationUnavailableReason,
	}
	valuedPositions := 0
	coverage := verticalslice.ValuationCoveragePartialStatus
	var currentValue *verticalslice.Money
	if available {
		price := verticalslice.Money{Amount: decimal.Must("318.50000000"), Currency: verticalslice.RUB}
		value := verticalslice.Money{Amount: decimal.Must("47775.00000000"), Currency: verticalslice.RUB}
		gain := verticalslice.Money{Amount: decimal.Must("6525.00000000"), Currency: verticalslice.RUB}
		ret := decimal.Must("0.15818182")
		weight := decimal.Must("1.00000000")
		asOf := "2026-09-08"
		market = verticalslice.MarketValuationUnavailable{
			Status:           verticalslice.MarketValuationAvailableStatus,
			Source:           verticalslice.ManualValuationSourceUserSupplied,
			MarketPrice:      &price,
			MarketValue:      &value,
			UnrealizedGain:   &gain,
			UnrealizedReturn: &ret,
			MarketWeight:     &weight,
			AsOf:             &asOf,
		}
		valuedPositions = 1
		coverage = verticalslice.ValuationCoverageCompleteStatus
		portfolioValue := verticalslice.Money{Amount: decimal.Must("129775.00000000"), Currency: verticalslice.RUB}
		currentValue = &portfolioValue
	}
	return verticalslice.PortfolioPositionsProjection{
		Items: []verticalslice.PortfolioPositionProjection{{
			Ticker:              "SBER",
			AssetType:           "STOCK",
			Quantity:            decimal.Must("150.00000000"),
			WeightedAverageCost: verticalslice.Money{Amount: decimal.Must("275.00000000"), Currency: verticalslice.RUB},
			AcquisitionBasis:    verticalslice.Money{Amount: decimal.Must("41250.00000000"), Currency: verticalslice.RUB},
			AcquisitionBasisWeight: func() *decimal.Decimal {
				value := decimal.Must("1.00000000")
				return &value
			}(),
			MarketValuation: market,
		}},
		TotalAcquisitionBasis: verticalslice.Money{Amount: decimal.Must("41250.00000000"), Currency: verticalslice.RUB},
		ValuationSummary: verticalslice.PortfolioValuationSummary{
			Status:                          coverage,
			ValuedPositions:                 valuedPositions,
			TotalOpenPositions:              1,
			ValuedPositionsMarketValue:      verticalslice.Money{Amount: func() decimal.Decimal { if available { return decimal.Must("47775.00000000") }; return decimal.Zero() }(), Currency: verticalslice.RUB},
			ValuedPositionsAcquisitionBasis: verticalslice.Money{Amount: func() decimal.Decimal { if available { return decimal.Must("41250.00000000") }; return decimal.Zero() }(), Currency: verticalslice.RUB},
			TotalAcquisitionBasis:           verticalslice.Money{Amount: decimal.Must("41250.00000000"), Currency: verticalslice.RUB},
			UnrealizedGain:                  verticalslice.Money{Amount: func() decimal.Decimal { if available { return decimal.Must("6525.00000000") }; return decimal.Zero() }(), Currency: verticalslice.RUB},
			CashValue:                       verticalslice.Money{Amount: decimal.Must("82000.00000000"), Currency: verticalslice.RUB},
			CurrentPortfolioValue:           currentValue,
		},
		MethodologyVersion: verticalslice.PortfolioPositionProjectionMethodologyVersion,
	}
}

func TestStage376ManualValuationHTTPPutEmitsBackendDerivedProjection(t *testing.T) {
	store := &stage376HTTPStore{projection: stage376HTTPProjection(true)}
	app := NewDevelopment(verticalslice.NewService(store, fixedHTTPClock{}))
	request := httptest.NewRequest(
		http.MethodPut,
		"/api/v1/portfolios/00000000-0000-4000-8000-000000000002/valuations/SBER",
		strings.NewReader(`{"marketPrice":{"amount":"318.50000000","currency":"RUB"},"asOfDate":"2026-09-08"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("PUT Stage 3.76 valuation: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK || store.upsertCalls != 1 {
		t.Fatalf("PUT status/calls mismatch: status=%d calls=%d", response.StatusCode, store.upsertCalls)
	}
	if store.upsert.Ticker != "SBER" || store.upsert.Price.Amount.String() != "318.50000000" ||
		store.upsert.Price.Currency != verticalslice.RUB || store.upsert.AsOfDate != "2026-09-08" {
		t.Fatalf("PUT transport drifted: %+v", store.upsert)
	}
	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode Stage 3.76 PUT response: %v", err)
	}
	data := payload["data"].(map[string]any)
	item := data["items"].([]any)[0].(map[string]any)
	market := item["marketValuation"].(map[string]any)
	if market["status"] != "AVAILABLE" || market["source"] != "USER_SUPPLIED" || market["reason"] != nil ||
		market["provider"] != nil || market["unrealizedReturn"] != "0.15818182" {
		t.Fatalf("available market DTO mismatch: %+v", market)
	}
	if market["marketPrice"].(map[string]any)["amount"] != "318.50000000" ||
		market["marketValue"].(map[string]any)["amount"] != "47775.00000000" ||
		market["unrealizedGain"].(map[string]any)["amount"] != "6525.00000000" {
		t.Fatalf("backend-derived valuation DTO mismatch: %+v", market)
	}
	summary := data["valuationSummary"].(map[string]any)
	if summary["status"] != "COMPLETE" || summary["currentPortfolioValue"].(map[string]any)["amount"] != "129775.00000000" {
		t.Fatalf("valuation summary DTO mismatch: %+v", summary)
	}
}

func TestStage376ManualValuationHTTPRejectsInvalidPayloadBeforeStore(t *testing.T) {
	for _, body := range []string{
		`{"marketPrice":{"amount":"0.00000000","currency":"RUB"},"asOfDate":"2026-09-08"}`,
		`{"marketPrice":{"amount":"318.50000000","currency":"USD"},"asOfDate":"2026-09-08"}`,
		`{"marketPrice":{"amount":"318.50000000","currency":"RUB"},"asOfDate":"2026-02-30"}`,
		`{"marketPrice":{"amount":"318.50000000","currency":"RUB"},"asOfDate":"2026-09-08","unexpected":true}`,
	} {
		store := &stage376HTTPStore{projection: stage376HTTPProjection(false)}
		app := NewDevelopment(verticalslice.NewService(store, fixedHTTPClock{}))
		request := httptest.NewRequest(http.MethodPut, "/api/v1/portfolios/00000000-0000-4000-8000-000000000002/valuations/SBER", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		response, err := app.Test(request)
		if err != nil {
			t.Fatalf("request invalid Stage 3.76 valuation: %v", err)
		}
		response.Body.Close()
		if response.StatusCode != http.StatusBadRequest || store.upsertCalls != 0 {
			t.Fatalf("invalid PUT must fail before store: body=%s status=%d calls=%d", body, response.StatusCode, store.upsertCalls)
		}
	}
}

func TestStage376ManualValuationHTTPDeleteIsRetrySafe(t *testing.T) {
	store := &stage376HTTPStore{projection: stage376HTTPProjection(false)}
	app := NewDevelopment(verticalslice.NewService(store, fixedHTTPClock{}))
	for attempt := 1; attempt <= 2; attempt++ {
		request := httptest.NewRequest(http.MethodDelete, "/api/v1/portfolios/00000000-0000-4000-8000-000000000002/valuations/SBER", nil)
		response, err := app.Test(request)
		if err != nil {
			t.Fatalf("DELETE Stage 3.76 valuation attempt %d: %v", attempt, err)
		}
		response.Body.Close()
		if response.StatusCode != http.StatusOK {
			t.Fatalf("DELETE attempt %d status=%d", attempt, response.StatusCode)
		}
	}
	if store.clearCalls != 2 || store.clearTicker != "SBER" {
		t.Fatalf("DELETE retry-safe transport mismatch: calls=%d ticker=%q", store.clearCalls, store.clearTicker)
	}
}
