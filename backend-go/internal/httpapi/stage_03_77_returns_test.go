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

type stage377HTTPStore struct {
	importAPITestStore
	projection verticalslice.PortfolioReturnProjection
	asOfDate   string
	calls      int
}

func (store *stage377HTTPStore) GetPortfolioReturns(
	_ context.Context,
	_ string,
	_ string,
	asOfDate string,
) (verticalslice.PortfolioReturnProjection, error) {
	store.calls++
	store.asOfDate = asOfDate
	return store.projection, nil
}

func TestStage377ReturnsHTTPPreservesBackendProjection(t *testing.T) {
	xirr := decimal.Must("0.12048717")
	terminal := verticalslice.Money{Amount: decimal.Must("1284.22996102"), Currency: verticalslice.RUB}
	store := &stage377HTTPStore{projection: verticalslice.PortfolioReturnProjection{
		PortfolioID: "00000000-0000-4000-8000-000000000002",
		AsOfDate:    "2026-02-17",
		Status:      verticalslice.PortfolioReturnAvailableStatus,
		XIRR:        &xirr,
		ExternalCashFlows: []verticalslice.PortfolioReturnCashFlow{
			{Date: "2025-01-05", Amount: decimal.Must("-1000.00000000")},
			{Date: "2025-04-20", Amount: decimal.Must("-250.00000000")},
			{Date: "2025-09-10", Amount: decimal.Must("120.00000000")},
		},
		TerminalPortfolioValue: &terminal,
		MethodologyVersion:     verticalslice.PortfolioReturnMethodologyVersion,
		DayCountConvention:     verticalslice.PortfolioReturnDayCountConvention,
	}}
	app := NewDevelopment(verticalslice.NewService(store, fixedHTTPClock{}))
	response, err := app.Test(httptest.NewRequest(
		http.MethodGet,
		"/api/v1/portfolios/00000000-0000-4000-8000-000000000002/returns?asOfDate=2026-02-17",
		nil,
	))
	if err != nil {
		t.Fatalf("request Stage 3.77 returns: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("returns status got %d want %d", response.StatusCode, http.StatusOK)
	}
	if store.calls != 1 || store.asOfDate != "2026-02-17" {
		t.Fatalf("explicit asOfDate was not preserved: calls=%d asOf=%q", store.calls, store.asOfDate)
	}
	var payload struct {
		Data portfolioReturnDTO `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode returns response: %v", err)
	}
	if payload.Data.Status != verticalslice.PortfolioReturnAvailableStatus || payload.Data.XIRR == nil || *payload.Data.XIRR != "0.12048717" ||
		payload.Data.Reason != nil || payload.Data.TerminalPortfolioValue == nil || payload.Data.TerminalPortfolioValue.Amount != "1284.22996102" ||
		payload.Data.Calculation.MethodologyVersion != verticalslice.PortfolioReturnMethodologyVersion ||
		payload.Data.Calculation.DayCountConvention != verticalslice.PortfolioReturnDayCountConvention || len(payload.Data.ExternalCashFlows) != 3 {
		t.Fatalf("HTTP mapper changed backend XIRR truth: %+v", payload.Data)
	}
}

func TestStage377ReturnsHTTPRequiresExplicitValidAsOfDate(t *testing.T) {
	store := &stage377HTTPStore{}
	app := NewDevelopment(verticalslice.NewService(store, fixedHTTPClock{}))
	for _, path := range []string{
		"/api/v1/portfolios/00000000-0000-4000-8000-000000000002/returns",
		"/api/v1/portfolios/00000000-0000-4000-8000-000000000002/returns?asOfDate=",
		"/api/v1/portfolios/00000000-0000-4000-8000-000000000002/returns?asOfDate=2026-02-30",
		"/api/v1/portfolios/00000000-0000-4000-8000-000000000002/returns?asOfDate=%202026-02-17",
	} {
		response, err := app.Test(httptest.NewRequest(http.MethodGet, path, nil))
		if err != nil {
			t.Fatalf("request invalid Stage 3.77 returns: %v", err)
		}
		response.Body.Close()
		if response.StatusCode != http.StatusBadRequest {
			t.Fatalf("invalid returns query %q got %d want %d", path, response.StatusCode, http.StatusBadRequest)
		}
	}
	if store.calls != 0 {
		t.Fatalf("invalid asOfDate must fail before store work, calls=%d", store.calls)
	}
}

func TestStage377SummaryMapperActivatesOnlyXIRR(t *testing.T) {
	xirr := decimal.Must("0.10000000")
	zeroMoney := verticalslice.Money{Amount: decimal.Zero(), Currency: verticalslice.RUB}
	summary := verticalslice.PortfolioSummary{
		TotalValue:        zeroMoney,
		CashValue:         zeroMoney,
		StockValue:        zeroMoney,
		BondValue:         zeroMoney,
		InvestedCapital:   zeroMoney,
		DividendsReceived: zeroMoney,
		CouponsReceived:   zeroMoney,
		XIRR:              &xirr,
		PurchasingPower: verticalslice.PurchasingPower{
			PortfolioValue: zeroMoney,
		},
	}
	mapped := mapSummary(summary)
	if mapped.XIRR == nil || *mapped.XIRR != "0.10000000" {
		t.Fatalf("summary XIRR mirror mismatch: %+v", mapped.XIRR)
	}
	if mapped.NominalReturnRate != nil || mapped.RealReturn != nil {
		t.Fatalf("Stage 3.77 must not activate nominal/real returns: %+v", mapped)
	}
}

func TestStage377NormalAndReplayConstructorsExposeValuationAndReturnsRoutes(t *testing.T) {
	store := &stage377HTTPStore{}
	service := verticalslice.NewService(store, fixedHTTPClock{})
	normal := NewDevelopment(service)
	replay := NewDevelopmentReplay(service)
	apps := map[string]func(*http.Request) (*http.Response, error){
		"normal": func(request *http.Request) (*http.Response, error) { return normal.Test(request) },
		"replay": func(request *http.Request) (*http.Response, error) { return replay.Test(request) },
	}
	for name, testRequest := range apps {
		t.Run(name, func(t *testing.T) {
			cases := []*http.Request{
				httptest.NewRequest(http.MethodGet, "/api/v1/portfolios/p/returns", nil),
				httptest.NewRequest(http.MethodPut, "/api/v1/portfolios/p/valuations/SBER", strings.NewReader(`{}`)),
				httptest.NewRequest(http.MethodDelete, "/api/v1/portfolios/p/valuations/SBER", nil),
			}
			for _, request := range cases {
				response, err := testRequest(request)
				if err != nil {
					t.Fatalf("route probe %s %s: %v", request.Method, request.URL.Path, err)
				}
				response.Body.Close()
				if response.StatusCode == http.StatusNotFound {
					t.Fatalf("constructor %s omitted route %s %s", name, request.Method, request.URL.Path)
				}
			}
		})
	}
}
