package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
		TotalAcquisitionBasis: verticalslice.Money{Amount: decimal.Must("120578.77813505"), Currency: verticalslice.RUB},
		InputsAsOf:            &inputsAsOf,
		MethodologyVersion:    verticalslice.PortfolioPositionProjectionMethodologyVersion,
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
	for _, field := range []string{"marketPrice", "marketValue", "unrealizedGain", "marketWeight", "provider", "asOf"} {
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
		MethodologyVersion:    verticalslice.PortfolioPositionProjectionMethodologyVersion,
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
