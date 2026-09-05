package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

type corporateActionFixtureProvider struct {
	events    []verticalslice.CorporateActionEvent
	err       error
	lastQuery verticalslice.CorporateActionQuery
}

func (provider *corporateActionFixtureProvider) CorporateActions(_ context.Context, query verticalslice.CorporateActionQuery) ([]verticalslice.CorporateActionEvent, error) {
	provider.lastQuery = query
	if provider.err != nil {
		return nil, provider.err
	}
	return append([]verticalslice.CorporateActionEvent(nil), provider.events...), nil
}

func TestCorporateActionProjectionReturns503WhenSourceIsUnavailable(t *testing.T) {
	response := performCorporateActionRequest(t, nil, "/api/v1/corporate-actions/projection?instrumentId=SBER&from=2026-01-01&to=2026-12-31")
	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusServiceUnavailable)
	}
	if got := response.Header.Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", got)
	}
	body := decodeCorporateActionResponse(t, response)
	if got := body.Error.Code; got != "CORPORATE_ACTIONS_SOURCE_UNAVAILABLE" {
		t.Fatalf("error code = %q", got)
	}
}

func TestCorporateActionProjectionRejectsInvalidQueryBeforeProvider(t *testing.T) {
	provider := &corporateActionFixtureProvider{}
	response := performCorporateActionRequest(t, provider, "/api/v1/corporate-actions/projection?instrumentId=sber&from=2026-01-01&to=2026-12-31")
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusBadRequest)
	}
	if provider.lastQuery.InstrumentIDs != nil {
		t.Fatalf("provider invoked for invalid query: %#v", provider.lastQuery)
	}
	body := decodeCorporateActionResponse(t, response)
	if body.Error.Code != "VALIDATION_ERROR" {
		t.Fatalf("error code = %q", body.Error.Code)
	}
}

func TestCorporateActionProjectionRejectsMoreThanFiftyInstrumentsBeforeProvider(t *testing.T) {
	provider := &corporateActionFixtureProvider{}
	instruments := make([]string, 51)
	for i := range instruments {
		instruments[i] = fmt.Sprintf("I%02d", i)
	}
	target := "/api/v1/corporate-actions/projection?instrumentId=" + strings.Join(instruments, ",") + "&from=2026-01-01&to=2026-12-31"
	response := performCorporateActionRequest(t, provider, target)
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusBadRequest)
	}
	if provider.lastQuery.InstrumentIDs != nil {
		t.Fatalf("provider invoked for oversized query: %#v", provider.lastQuery)
	}
}

func TestCorporateActionProjectionReturnsDeterministicCalendarAndHeatmap(t *testing.T) {
	record := "2026-10-10"
	payment := "2026-11-15"
	amount := verticalslice.Money{Amount: decimal.Must("12.34000000"), Currency: verticalslice.RUB}
	provider := &corporateActionFixtureProvider{events: []verticalslice.CorporateActionEvent{
		corporateActionHTTPEvent("evt-cpn", "RU000A", verticalslice.CorporateActionCoupon, verticalslice.CorporateActionPaid, nil, &payment, nil),
		corporateActionHTTPEvent("evt-div", "SBER", verticalslice.CorporateActionDividend, verticalslice.CorporateActionConfirmed, &record, nil, &amount),
	}}

	response := performCorporateActionRequest(t, provider, "/api/v1/corporate-actions/projection?instrumentId=SBER,RU000A&from=2026-01-01&to=2026-12-31")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusOK)
	}
	if got, want := provider.lastQuery.InstrumentIDs, []string{"RU000A", "SBER"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("provider instrumentIDs = %#v, want %#v", got, want)
	}

	body := decodeCorporateActionResponse(t, response)
	if got, want := len(body.Data.Calendar), 2; got != want {
		t.Fatalf("calendar len = %d, want %d", got, want)
	}
	if body.Data.Calendar[0].EffectiveDate != record || body.Data.Calendar[0].Event.EventID != "evt-div" {
		t.Fatalf("calendar[0] = %#v", body.Data.Calendar[0])
	}
	if body.Data.Calendar[0].Event.AmountPerUnit == nil || body.Data.Calendar[0].Event.AmountPerUnit.Amount != "12.34000000" {
		t.Fatalf("amount = %#v", body.Data.Calendar[0].Event.AmountPerUnit)
	}
	if got, want := len(body.Data.Heatmap), 2; got != want {
		t.Fatalf("heatmap len = %d, want %d", got, want)
	}
	if got, want := body.Data.Coverage.InstrumentIDs, []string{"RU000A", "SBER"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("coverage instruments = %#v, want %#v", got, want)
	}
	if body.Data.Coverage.InputMode != "PROVIDER" {
		t.Fatalf("input mode = %q", body.Data.Coverage.InputMode)
	}
}

func TestCorporateActionProjectionFiltersProviderOverReturnToRequestedEffectiveDateWindow(t *testing.T) {
	inside := "2026-10-10"
	outside := "2027-01-10"
	provider := &corporateActionFixtureProvider{events: []verticalslice.CorporateActionEvent{
		corporateActionHTTPEvent("evt-inside", "SBER", verticalslice.CorporateActionDividend, verticalslice.CorporateActionConfirmed, &inside, nil, nil),
		corporateActionHTTPEvent("evt-outside", "SBER", verticalslice.CorporateActionDividend, verticalslice.CorporateActionConfirmed, &outside, nil, nil),
	}}

	response := performCorporateActionRequest(t, provider, "/api/v1/corporate-actions/projection?instrumentId=SBER&from=2026-01-01&to=2026-12-31")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusOK)
	}
	body := decodeCorporateActionResponse(t, response)
	if got, want := len(body.Data.Calendar), 1; got != want {
		t.Fatalf("calendar len = %d, want %d", got, want)
	}
	if got := body.Data.Calendar[0].Event.EventID; got != "evt-inside" {
		t.Fatalf("calendar event = %q, want evt-inside", got)
	}
	if got, want := len(body.Data.Heatmap), 1; got != want {
		t.Fatalf("heatmap len = %d, want %d", got, want)
	}
	if got := body.Data.Heatmap[0].Date; got != inside {
		t.Fatalf("heatmap date = %q, want %q", got, inside)
	}
}

func TestCorporateActionProjectionReturns200ForLegitimateEmptyResult(t *testing.T) {
	provider := &corporateActionFixtureProvider{events: []verticalslice.CorporateActionEvent{}}
	response := performCorporateActionRequest(t, provider, "/api/v1/corporate-actions/projection?instrumentId=SBER&from=2026-01-01&to=2026-12-31")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusOK)
	}
	body := decodeCorporateActionResponse(t, response)
	if len(body.Data.Calendar) != 0 || len(body.Data.Heatmap) != 0 {
		t.Fatalf("expected legitimate empty projection, got %#v", body.Data)
	}
}

func TestCorporateActionProjectionMapsInvalidProviderEvidenceTo502(t *testing.T) {
	date := "2026-10-10"
	base := corporateActionHTTPEvent("evt-base", "SBER", verticalslice.CorporateActionDividend, verticalslice.CorporateActionConfirmed, &date, nil, nil)
	wrong := corporateActionHTTPEvent("evt-wrong", "SBER", verticalslice.CorporateActionCoupon, verticalslice.CorporateActionCancelled, nil, &date, nil)
	wrong.SupersedesEventID = stringPointerHTTP("evt-base")
	provider := &corporateActionFixtureProvider{events: []verticalslice.CorporateActionEvent{base, wrong}}

	response := performCorporateActionRequest(t, provider, "/api/v1/corporate-actions/projection?instrumentId=SBER&from=2026-01-01&to=2026-12-31")
	if response.StatusCode != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusBadGateway)
	}
	body := decodeCorporateActionResponse(t, response)
	if body.Error.Code != "CORPORATE_ACTIONS_SOURCE_INVALID" {
		t.Fatalf("error code = %q", body.Error.Code)
	}
}

func TestCorporateActionProjectionPreservesProviderNeutralDataErrorAs502(t *testing.T) {
	provider := &corporateActionFixtureProvider{err: verticalslice.ErrCorporateActionsProviderData}
	response := performCorporateActionRequest(t, provider, "/api/v1/corporate-actions/projection?instrumentId=SBER&from=2026-01-01&to=2026-12-31")
	if response.StatusCode != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusBadGateway)
	}
}

func TestCorporateActionProjectionDoesNotMisclassifyUnknownInternalError(t *testing.T) {
	provider := &corporateActionFixtureProvider{err: errors.New("provider detail")}
	response := performCorporateActionRequest(t, provider, "/api/v1/corporate-actions/projection?instrumentId=SBER&from=2026-01-01&to=2026-12-31")
	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusServiceUnavailable)
	}
}

func performCorporateActionRequest(t *testing.T, provider verticalslice.CorporateActionProvider, target string) *http.Response {
	t.Helper()
	app := fiber.New()
	api := &API{corporateActionProvider: provider}
	app.Get("/api/v1/corporate-actions/projection", api.getCorporateActionProjection)
	response, err := app.Test(httptest.NewRequest(http.MethodGet, target, nil))
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	return response
}

type corporateActionHTTPResponse struct {
	Data struct {
		Calendar []corporateActionCalendarEntryDTO `json:"calendar"`
		Heatmap  []corporateActionHeatmapBucketDTO `json:"heatmap"`
		Coverage corporateActionCoverageDTO        `json:"coverage"`
	} `json:"data"`
	Error struct {
		Code string `json:"code"`
	} `json:"error"`
}

func decodeCorporateActionResponse(t *testing.T, response *http.Response) corporateActionHTTPResponse {
	t.Helper()
	defer response.Body.Close()
	payload, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	if strings.Contains(string(payload), `"sourceEventId"`) {
		t.Fatalf("public response leaked provider-owned sourceEventId: %s", payload)
	}
	var body corporateActionHTTPResponse
	if err := json.Unmarshal(payload, &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return body
}

func corporateActionHTTPEvent(id, instrument string, kind verticalslice.CorporateActionKind, status verticalslice.CorporateActionStatus, recordDate, paymentDate *string, amount *verticalslice.Money) verticalslice.CorporateActionEvent {
	return verticalslice.CorporateActionEvent{
		EventID:       id,
		InstrumentID:  instrument,
		Kind:          kind,
		Status:        status,
		RecordDate:    recordDate,
		PaymentDate:   paymentDate,
		AmountPerUnit: amount,
		AsOf:          time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC),
		RetrievedAt:   time.Date(2026, 9, 5, 0, 1, 0, 0, time.UTC),
		Provenance: verticalslice.CorporateActionProvenance{
			Provider:      "FIXTURE",
			SourceEventID: "source-" + id,
		},
	}
}

func stringPointerHTTP(value string) *string { return &value }
