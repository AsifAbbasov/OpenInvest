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

type stage375HTTPStore struct {
	importAPITestStore
	projection verticalslice.PortfolioCashFlowProjection
	fromDate   string
	toDate     string
	calls      int
}

func (store *stage375HTTPStore) GetPortfolioCashFlow(
	_ context.Context,
	_ string,
	_ string,
	fromDate string,
	toDate string,
) (verticalslice.PortfolioCashFlowProjection, error) {
	store.calls++
	store.fromDate = fromDate
	store.toDate = toDate
	return store.projection, nil
}

func cashMoney375(amount string) verticalslice.Money {
	return verticalslice.Money{Amount: decimal.Must(amount), Currency: verticalslice.RUB}
}

func zeroCashFlowTotals375HTTP() verticalslice.PortfolioCashFlowTotals {
	zero := cashMoney375("0.00000000")
	return verticalslice.PortfolioCashFlowTotals{
		Deposits:            zero,
		Withdrawals:         zero,
		BuyOutflows:         zero,
		SellInflows:         zero,
		DividendsGross:      zero,
		CouponsGross:        zero,
		Fees:                zero,
		Taxes:               zero,
		NetExternalFlow:     zero,
		NetInvestmentIncome: zero,
		NetCashFlow:         zero,
	}
}

func TestStage375CashFlowHTTPPreservesRangeAndBackendTotals(t *testing.T) {
	inputsAsOf := "2026-03-31"
	periodTotals := zeroCashFlowTotals375HTTP()
	periodTotals.NetCashFlow = cashMoney375("3830.00000000")
	store := &stage375HTTPStore{projection: verticalslice.PortfolioCashFlowProjection{
		PortfolioID: "00000000-0000-4000-8000-000000000002",
		FromDate:    stringPtr375HTTP("2026-01-01"),
		ToDate:      stringPtr375HTTP("2026-03-31"),
		Totals: verticalslice.PortfolioCashFlowTotals{
			Deposits: cashMoney375("100000.00000000"), Withdrawals: cashMoney375("5000.00000000"),
			BuyOutflows: cashMoney375("80000.00000000"), SellInflows: cashMoney375("9000.00000000"),
			DividendsGross: cashMoney375("3000.00000000"), CouponsGross: cashMoney375("1200.00000000"),
			Fees: cashMoney375("230.00000000"), Taxes: cashMoney375("596.00000000"),
			NetExternalFlow: cashMoney375("95000.00000000"), NetInvestmentIncome: cashMoney375("3644.00000000"),
			NetCashFlow: cashMoney375("27374.00000000"),
		},
		Periods: []verticalslice.PortfolioCashFlowPeriod{{
			Month:  "2026-03",
			Totals: periodTotals,
		}},
		InputsAsOf:         &inputsAsOf,
		MethodologyVersion: verticalslice.PortfolioCashFlowMethodologyVersion,
	}}
	app := NewDevelopment(verticalslice.NewService(store, fixedHTTPClock{}))
	request := httptest.NewRequest(http.MethodGet, "/api/v1/portfolios/00000000-0000-4000-8000-000000000002/cash-flow?fromDate=2026-01-01&toDate=2026-03-31", nil)
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request cash-flow endpoint: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("cash-flow status got %d want %d", response.StatusCode, http.StatusOK)
	}
	if store.calls != 1 || store.fromDate != "2026-01-01" || store.toDate != "2026-03-31" {
		t.Fatalf("range was not preserved: calls=%d from=%q to=%q", store.calls, store.fromDate, store.toDate)
	}
	var payload struct {
		Data struct {
			Totals struct {
				DividendsGross      moneyDTO `json:"dividendsGross"`
				Fees                moneyDTO `json:"fees"`
				Taxes               moneyDTO `json:"taxes"`
				NetInvestmentIncome moneyDTO `json:"netInvestmentIncome"`
				NetCashFlow         moneyDTO `json:"netCashFlow"`
			} `json:"totals"`
			Periods []struct {
				Month string `json:"month"`
			} `json:"periods"`
			Calculation struct {
				MethodologyVersion string  `json:"methodologyVersion"`
				InputsAsOf         *string `json:"inputsAsOf"`
			} `json:"calculation"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode cash-flow response: %v", err)
	}
	if payload.Data.Totals.DividendsGross.Amount != "3000.00000000" ||
		payload.Data.Totals.Fees.Amount != "230.00000000" ||
		payload.Data.Totals.Taxes.Amount != "596.00000000" ||
		payload.Data.Totals.NetInvestmentIncome.Amount != "3644.00000000" ||
		payload.Data.Totals.NetCashFlow.Amount != "27374.00000000" {
		t.Fatalf("HTTP mapper changed backend financial truth: %+v", payload.Data.Totals)
	}
	if len(payload.Data.Periods) != 1 || payload.Data.Periods[0].Month != "2026-03" ||
		payload.Data.Calculation.MethodologyVersion != verticalslice.PortfolioCashFlowMethodologyVersion ||
		payload.Data.Calculation.InputsAsOf == nil || *payload.Data.Calculation.InputsAsOf != "2026-03-31" {
		t.Fatalf("cash-flow metadata mismatch: %+v", payload.Data)
	}
}

func TestStage375CashFlowHTTPRejectsInvalidRangesBeforeStore(t *testing.T) {
	store := &stage375HTTPStore{}
	app := NewDevelopment(verticalslice.NewService(store, fixedHTTPClock{}))
	paths := []string{
		"/api/v1/portfolios/00000000-0000-4000-8000-000000000002/cash-flow?fromDate=2026-02-30",
		"/api/v1/portfolios/00000000-0000-4000-8000-000000000002/cash-flow?toDate=",
		"/api/v1/portfolios/00000000-0000-4000-8000-000000000002/cash-flow?fromDate=2026-04-01&toDate=2026-03-01",
	}
	for _, path := range paths {
		response, err := app.Test(httptest.NewRequest(http.MethodGet, path, nil))
		if err != nil {
			t.Fatalf("request invalid cash-flow range: %v", err)
		}
		response.Body.Close()
		if response.StatusCode != http.StatusBadRequest {
			t.Fatalf("invalid range %q got status %d want %d", path, response.StatusCode, http.StatusBadRequest)
		}
	}
	if store.calls != 0 {
		t.Fatalf("invalid range must fail before store work, calls=%d", store.calls)
	}
}

func stringPtr375HTTP(value string) *string {
	copy := value
	return &copy
}
