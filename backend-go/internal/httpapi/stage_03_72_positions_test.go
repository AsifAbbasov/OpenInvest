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

type stage372HTTPStore struct {
	importAPITestStore
	projection verticalslice.PortfolioPositionsProjection
	asOfDate   string
	calls      int
}

func (store *stage372HTTPStore) GetPortfolioPositions(
	_ context.Context,
	_ string,
	_ string,
	asOfDate string,
) (verticalslice.PortfolioPositionsProjection, error) {
	store.calls++
	store.asOfDate = asOfDate
	return store.projection, nil
}

func TestStage372PositionsHTTPEmitsExplicitUnavailableMarketNulls(t *testing.T) {
	weight := decimal.Must("0.34210000")
	inputsAsOf := "2026-09-03"
	totalBasis := verticalslice.Money{Amount: decimal.Must("120578.77813505"), Currency: verticalslice.RUB}
	store := &stage372HTTPStore{projection: verticalslice.PortfolioPositionsProjection{
		Items: []verticalslice.PortfolioPositionProjection{
			{
				Ticker:    "SBER",
				AssetType: "STOCK",
				Quantity:  decimal.Must("150.00000000"),
				WeightedAverageCost: verticalslice.Money{
					Amount: decimal.Must("275.00000000"), Currency: verticalslice.RUB,
				},
				AcquisitionBasis: verticalslice.Money{
					Amount: decimal.Must("41250.00000000"), Currency: verticalslice.RUB,
				},
				AcquisitionBasisWeight: &weight,
				MarketValuation: verticalslice.MarketValuationUnavailable{
					Status: verticalslice.MarketValuationUnavailableStatus,
					Reason: verticalslice.MarketValuationUnavailableReason,
				},
			},
			{
				Ticker:    "SU26238RMFS4",
				AssetType: "BOND",
				Quantity:  decimal.Must("1.00000000"),
				WeightedAverageCost: verticalslice.Money{
					Amount: decimal.Must("79328.77813505"), Currency: verticalslice.RUB,
				},
				AcquisitionBasis: verticalslice.Money{
					Amount: decimal.Must("79328.77813505"), Currency: verticalslice.RUB,
				},
				AcquisitionBasisWeight: decimalPtrHTTP(decimal.Must("0.65790000")),
				MarketValuation: verticalslice.MarketValuationUnavailable{
					Status: verticalslice.MarketValuationUnavailableStatus,
					Reason: verticalslice.MarketValuationUnavailableReason,
				},
			},
		},
		TotalAcquisitionBasis: totalBasis,
		ValuationSummary: verticalslice.PortfolioValuationSummary{
			Status:                          verticalslice.ValuationCoveragePartialStatus,
			ValuedPositions:                 0,
			TotalOpenPositions:              2,
			ValuedPositionsMarketValue:      verticalslice.ZeroMoney(),
			ValuedPositionsAcquisitionBasis: verticalslice.ZeroMoney(),
			TotalAcquisitionBasis:           totalBasis,
			UnrealizedGain:                  verticalslice.ZeroMoney(),
			CashValue:                       verticalslice.ZeroMoney(),
		},
		InputsAsOf:         &inputsAsOf,
		MethodologyVersion: verticalslice.PortfolioPositionProjectionMethodologyVersion,
	}}
	app := NewDevelopment(verticalslice.NewService(store, fixedHTTPClock{}))
	request := httptest.NewRequest(http.MethodGet, "/api/v1/portfolios/00000000-0000-4000-8000-000000000002/positions?asOfDate=2026-09-03", nil)

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request Stage 3.72 positions endpoint: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.StatusCode)
	}
	if store.calls != 1 || store.asOfDate != "2026-09-03" {
		t.Fatalf("positions handler did not preserve explicit asOfDate: calls=%d asOf=%q", store.calls, store.asOfDate)
	}

	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode Stage 3.72 response: %v", err)
	}
	data := payload["data"].(map[string]any)
	items := data["items"].([]any)
	item := items[0].(map[string]any)
	market := item["marketValuation"].(map[string]any)
	if market["status"] != verticalslice.MarketValuationUnavailableStatus || market["reason"] != verticalslice.MarketValuationUnavailableReason {
		t.Fatalf("unexpected market-unavailable state: %+v", market)
	}
	for _, field := range []string{"source", "marketPrice", "marketValue", "unrealizedGain", "unrealizedReturn", "marketWeight", "provider", "asOf"} {
		value, exists := market[field]
		if !exists {
			t.Fatalf("market field %s must be explicitly present", field)
		}
		if value != nil {
			t.Fatalf("market field %s must be null, got %#v", field, value)
		}
	}
	if item["quantity"] != "150.00000000" || item["acquisitionBasisWeight"] != "0.34210000" {
		t.Fatalf("financial DTO mismatch: %+v", item)
	}
	if data["totalAcquisitionBasis"].(map[string]any)["amount"] != "120578.77813505" {
		t.Fatalf("total acquisition basis must equal the server projection denominator: %+v", data["totalAcquisitionBasis"])
	}
	calculation := data["calculation"].(map[string]any)
	if calculation["inputsAsOf"] != "2026-09-03" || calculation["methodologyVersion"] != verticalslice.PortfolioPositionProjectionMethodologyVersion {
		t.Fatalf("calculation metadata mismatch: %+v", calculation)
	}
}

func TestStage372PositionsHTTPPreservesNullableWeightAndInputsAsOf(t *testing.T) {
	store := &stage372HTTPStore{projection: verticalslice.PortfolioPositionsProjection{
		Items: []verticalslice.PortfolioPositionProjection{{
			Ticker:    "ZERO",
			AssetType: "STOCK",
			Quantity:  decimal.Must("0.00000001"),
			WeightedAverageCost: verticalslice.Money{
				Amount: decimal.Must("0.00000001"), Currency: verticalslice.RUB,
			},
			AcquisitionBasis: verticalslice.Money{Amount: decimal.Zero(), Currency: verticalslice.RUB},
			MarketValuation: verticalslice.MarketValuationUnavailable{
				Status: verticalslice.MarketValuationUnavailableStatus,
				Reason: verticalslice.MarketValuationUnavailableReason,
			},
		}},
		TotalAcquisitionBasis: verticalslice.ZeroMoney(),
		ValuationSummary: verticalslice.PortfolioValuationSummary{
			Status:                          verticalslice.ValuationCoveragePartialStatus,
			ValuedPositions:                 0,
			TotalOpenPositions:              1,
			ValuedPositionsMarketValue:      verticalslice.ZeroMoney(),
			ValuedPositionsAcquisitionBasis: verticalslice.ZeroMoney(),
			TotalAcquisitionBasis:           verticalslice.ZeroMoney(),
			UnrealizedGain:                  verticalslice.ZeroMoney(),
			CashValue:                       verticalslice.ZeroMoney(),
		},
		MethodologyVersion: verticalslice.PortfolioPositionProjectionMethodologyVersion,
	}}
	app := NewDevelopment(verticalslice.NewService(store, fixedHTTPClock{}))
	request := httptest.NewRequest(http.MethodGet, "/api/v1/portfolios/00000000-0000-4000-8000-000000000002/positions", nil)

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request zero-basis positions endpoint: %v", err)
	}
	defer response.Body.Close()
	var payload struct {
		Data struct {
			Items []struct {
				AcquisitionBasisWeight *string `json:"acquisitionBasisWeight"`
			} `json:"items"`
			Calculation struct {
				InputsAsOf *string `json:"inputsAsOf"`
			} `json:"calculation"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode zero-basis response: %v", err)
	}
	if len(payload.Data.Items) != 1 || payload.Data.Items[0].AcquisitionBasisWeight != nil || payload.Data.Calculation.InputsAsOf != nil {
		t.Fatalf("nullable Stage 3.72 fields drifted: %+v", payload.Data)
	}
}

func TestStage372PositionsHTTPRejectsInvalidOrSuppliedEmptyAsOfDate(t *testing.T) {
	store := &stage372HTTPStore{}
	app := NewDevelopment(verticalslice.NewService(store, fixedHTTPClock{}))
	for _, path := range []string{
		"/api/v1/portfolios/00000000-0000-4000-8000-000000000002/positions?asOfDate=2026-02-30",
		"/api/v1/portfolios/00000000-0000-4000-8000-000000000002/positions?asOfDate=",
	} {
		response, err := app.Test(httptest.NewRequest(http.MethodGet, path, nil))
		if err != nil {
			t.Fatalf("request invalid Stage 3.72 asOfDate: %v", err)
		}
		response.Body.Close()
		if response.StatusCode != http.StatusBadRequest {
			t.Fatalf("invalid asOfDate %q: got status %d want %d", path, response.StatusCode, http.StatusBadRequest)
		}
	}
	if store.calls != 0 {
		t.Fatalf("invalid asOfDate must fail before store work, calls=%d", store.calls)
	}
}

func decimalPtrHTTP(value decimal.Decimal) *decimal.Decimal {
	copy := value
	return &copy
}

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
