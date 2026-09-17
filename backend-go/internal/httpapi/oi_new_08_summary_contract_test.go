package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

type oiNew08SummaryHTTPStore struct {
	importAPITestStore
	selectedAsOf string
	forwarded    string
	calls        int
}

func (store *oiNew08SummaryHTTPStore) GetPortfolioSummary(
	_ context.Context,
	_ string,
	portfolioID string,
	asOfDate string,
) (verticalslice.PortfolioSummary, error) {
	store.calls++
	store.forwarded = asOfDate
	selected := store.selectedAsOf
	if selected == "" {
		selected = "2026-09-01"
	}
	zero := verticalslice.ZeroMoney()
	return verticalslice.PortfolioSummary{
		PortfolioID:       portfolioID,
		AsOfDate:          selected,
		TotalValue:        zero,
		CashValue:         zero,
		StockValue:        zero,
		BondValue:         zero,
		InvestedCapital:   zero,
		DividendsReceived: zero,
		CouponsReceived:   zero,
		NominalReturnRate: decimal.Zero(),
		PurchasingPower: verticalslice.PurchasingPower{
			PortfolioValue: zero,
			AsOfDate:       selected,
		},
		MethodologyVersion: "oi-new-08-http-contract-test",
		CalculatedAt:       time.Date(2026, 9, 17, 0, 0, 0, 0, time.UTC),
	}, nil
}

func TestOINew08SummaryHTTPOmittedAsOfDatePreservesOmittedMode(t *testing.T) {
	store := &oiNew08SummaryHTTPStore{selectedAsOf: "2099-01-01"}
	app := NewDevelopment(verticalslice.NewService(store, fixedHTTPClock{}))

	response, err := app.Test(httptest.NewRequest(
		http.MethodGet,
		"/api/v1/portfolios/00000000-0000-4000-8000-000000000002/summary",
		nil,
	))
	if err != nil {
		t.Fatalf("request omitted summary asOfDate: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("omitted asOfDate: got status %d want %d", response.StatusCode, http.StatusOK)
	}
	if store.calls != 1 || store.forwarded != "" {
		t.Fatalf("omitted asOfDate must forward empty mode exactly: calls=%d forwarded=%q", store.calls, store.forwarded)
	}

	var payload struct {
		Data struct {
			AsOfDate    string `json:"asOfDate"`
			Calculation struct {
				InputsAsOf string `json:"inputsAsOf"`
			} `json:"calculation"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode omitted summary response: %v", err)
	}
	if payload.Data.AsOfDate != "2099-01-01" || payload.Data.Calculation.InputsAsOf != "2099-01-01" {
		t.Fatalf("selected future snapshot date must remain authoritative: %+v", payload.Data)
	}
}

func TestOINew08SummaryHTTPExplicitDateForwardsRequestButReturnsSelectedSnapshotDate(t *testing.T) {
	store := &oiNew08SummaryHTTPStore{selectedAsOf: "2026-09-01"}
	app := NewDevelopment(verticalslice.NewService(store, fixedHTTPClock{}))

	response, err := app.Test(httptest.NewRequest(
		http.MethodGet,
		"/api/v1/portfolios/00000000-0000-4000-8000-000000000002/summary?asOfDate=2026-09-07",
		nil,
	))
	if err != nil {
		t.Fatalf("request explicit summary asOfDate: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("explicit asOfDate: got status %d want %d", response.StatusCode, http.StatusOK)
	}
	if store.calls != 1 || store.forwarded != "2026-09-07" {
		t.Fatalf("explicit asOfDate must be forwarded unchanged: calls=%d forwarded=%q", store.calls, store.forwarded)
	}

	var payload struct {
		Data struct {
			AsOfDate    string `json:"asOfDate"`
			Calculation struct {
				InputsAsOf string `json:"inputsAsOf"`
			} `json:"calculation"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode explicit summary response: %v", err)
	}
	if payload.Data.AsOfDate != "2026-09-01" || payload.Data.Calculation.InputsAsOf != "2026-09-01" {
		t.Fatalf("response must describe actual selected snapshot, not raw requested date: %+v", payload.Data)
	}
}

func TestOINew08SummaryHTTPRejectsEmptyAndInvalidAsOfDate(t *testing.T) {
	cases := []string{
		"/api/v1/portfolios/00000000-0000-4000-8000-000000000002/summary?asOfDate=",
		"/api/v1/portfolios/00000000-0000-4000-8000-000000000002/summary?asOfDate=not-a-date",
	}
	for _, path := range cases {
		t.Run(path, func(t *testing.T) {
			store := &oiNew08SummaryHTTPStore{}
			app := NewDevelopment(verticalslice.NewService(store, fixedHTTPClock{}))
			response, err := app.Test(httptest.NewRequest(http.MethodGet, path, nil))
			if err != nil {
				t.Fatalf("request invalid summary asOfDate: %v", err)
			}
			defer response.Body.Close()
			if response.StatusCode != http.StatusBadRequest {
				t.Fatalf("invalid asOfDate %q: got status %d want %d", path, response.StatusCode, http.StatusBadRequest)
			}
			var payload struct {
				Error struct {
					Code string `json:"code"`
				} `json:"error"`
			}
			if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
				t.Fatalf("decode invalid summary response: %v", err)
			}
			if payload.Error.Code != "VALIDATION_ERROR" {
				t.Fatalf("invalid asOfDate must return VALIDATION_ERROR, got %q", payload.Error.Code)
			}
			if store.calls != 0 {
				t.Fatalf("invalid asOfDate must fail before store work, calls=%d", store.calls)
			}
		})
	}
}
