package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestDividendReplaySubjectIDIsDeterministicPerKeyAndSeparatedFromDevelopmentSubject(t *testing.T) {
	key := "stage-03-68-idempotency-key-000001"
	first := dividendReplaySubjectID(key)
	second := dividendReplaySubjectID(key)
	if first != second {
		t.Fatalf("same key produced different technical scopes: %s != %s", first, second)
	}
	if first == devSubjectID {
		t.Fatalf("calculator replay scope must not collide with development subject %s", devSubjectID)
	}
	if len(first) != 36 || first[14] != '8' || !strings.ContainsRune("89ab", rune(first[19])) {
		t.Fatalf("calculator replay scope must be an RFC-variant UUIDv8, got %s", first)
	}
	other := dividendReplaySubjectID("stage-03-68-idempotency-key-000002")
	if first == other {
		t.Fatalf("different idempotency keys produced the same technical scope: %s", first)
	}
}

func TestDividendRequestDTORejectsInvalidDecimalWithoutFloatParsing(t *testing.T) {
	request := dividendCalculationRequestDTO{
		Ticker:   "SBER",
		Quantity: "1e3",
		DividendPerUnit: moneyDTO{
			Amount:   "34.84000000",
			Currency: verticalslice.RUB,
		},
	}
	if _, err := request.toApp(); err == nil {
		t.Fatal("scientific notation must fail closed")
	}
}

func TestMapDividendCalculationKeepsNullableYieldAndGrossOnlySemantics(t *testing.T) {
	calculation := verticalslice.DividendCalculation{
		Ticker:             "SBER",
		Quantity:           decimal.Must("1000.00000000"),
		DividendPerUnit:    verticalslice.Money{Amount: decimal.Must("34.84000000"), Currency: verticalslice.RUB},
		GrossDividend:      verticalslice.Money{Amount: decimal.Must("34840.00000000"), Currency: verticalslice.RUB},
		TaxIncluded:        false,
		MethodologyVersion: verticalslice.DividendCalculatorMethodologyVersion,
	}
	mapped := mapDividendCalculation(calculation)
	if mapped.GrossYield != nil || mapped.PositionCost != nil {
		t.Fatalf("omitted position cost must map to null cost/yield: %#v", mapped)
	}
	if mapped.TaxIncluded {
		t.Fatal("taxIncluded must remain false")
	}
	encoded, err := json.Marshal(mapped)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var object map[string]any
	if err := json.Unmarshal(encoded, &object); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := object["grossYield"]; !ok || object["grossYield"] != nil {
		t.Fatalf("grossYield must be present as JSON null: %s", encoded)
	}
}

func TestDividendCalculatorLimiterBoundsUniqueKeyAdmission(t *testing.T) {
	limiter := newBoundedAuthRateLimiter(2, 2, 2, defaultDividendCalculatorWindow)
	now := testTime()
	if !limiter.allow("key-000000000001", now) || !limiter.allow("key-000000000002", now) {
		t.Fatal("initial bounded admissions must pass")
	}
	if limiter.allow("key-000000000003", now) {
		t.Fatal("global unique-key amplification must be bounded")
	}
}

type dividendHTTPTestStore struct {
	command        verticalslice.CommandContext
	replayArtifact *verticalslice.CommandReplayArtifact
	replayOnLookup int
	lookupCalls    int
	writeCalls     int
}

func (store *dividendHTTPTestStore) Ping(context.Context) error { return nil }
func (store *dividendHTTPTestStore) SearchAssets(context.Context, verticalslice.AssetSearchFilter) ([]verticalslice.AssetSummary, error) {
	return nil, nil
}
func (store *dividendHTTPTestStore) ListPortfolios(context.Context, string, verticalslice.PortfolioFilter) ([]verticalslice.Portfolio, error) {
	return nil, nil
}
func (store *dividendHTTPTestStore) CreatePortfolio(context.Context, verticalslice.CommandContext, verticalslice.CreatePortfolioRequest) (verticalslice.Portfolio, error) {
	return verticalslice.Portfolio{}, nil
}
func (store *dividendHTTPTestStore) GetPortfolio(context.Context, string, string) (verticalslice.Portfolio, error) {
	return verticalslice.Portfolio{}, nil
}
func (store *dividendHTTPTestStore) ListTransactions(context.Context, string, string, verticalslice.TransactionFilter) ([]verticalslice.Transaction, error) {
	return nil, nil
}
func (store *dividendHTTPTestStore) ListImportReviewTransactions(context.Context, string, string, verticalslice.ImportReviewHistoryFilter) ([]verticalslice.Transaction, error) {
	return nil, nil
}
func (store *dividendHTTPTestStore) AppendTransaction(context.Context, verticalslice.CommandContext, verticalslice.AppendTransactionRequest) (verticalslice.Transaction, error) {
	return verticalslice.Transaction{}, nil
}
func (store *dividendHTTPTestStore) AppendImportedTransactions(context.Context, verticalslice.CommandContext, verticalslice.AppendImportBatchRequest) ([]verticalslice.Transaction, error) {
	return nil, nil
}
func (store *dividendHTTPTestStore) GetPortfolioSummary(context.Context, string, string, string) (verticalslice.PortfolioSummary, error) {
	return verticalslice.PortfolioSummary{}, nil
}
func (store *dividendHTTPTestStore) LookupReplayArtifact(
	_ context.Context,
	command verticalslice.CommandContext,
	_ string,
) (verticalslice.CommandReplayArtifact, bool, error) {
	store.command = command
	store.lookupCalls++
	if store.replayArtifact != nil && (store.replayOnLookup == 0 || store.lookupCalls >= store.replayOnLookup) {
		return *store.replayArtifact, true, nil
	}
	return verticalslice.CommandReplayArtifact{}, false, nil
}

func (store *dividendHTTPTestStore) CalculateDividendWithReplay(
	_ context.Context,
	command verticalslice.CommandContext,
	calculation verticalslice.DividendCalculation,
	build verticalslice.DividendReplayBuilder,
) (verticalslice.DividendCalculation, verticalslice.CommandReplayArtifact, error) {
	store.command = command
	store.writeCalls++
	artifact, err := build(calculation)
	return calculation, artifact, err
}

func TestCalculateDividendRouteMatchesCanonicalPublicContract(t *testing.T) {
	store := &dividendHTTPTestStore{}
	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	app := newApp(&API{service: service, dividendLimiter: newDividendCalculatorRateLimiter()})
	body := []byte(`{"ticker":"SBER","quantity":"1000.00000000","dividendPerUnit":{"amount":"34.84000000","currency":"RUB"},"positionCost":{"amount":"280000.00000000","currency":"RUB"}}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/dividends/calculate", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "stage-03-68-route-key-000001")

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", response.StatusCode)
	}
	var payload struct {
		Data dividendCalculationDTO `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.GrossDividend.Amount != "34840.00000000" {
		t.Fatalf("grossDividend = %s", payload.Data.GrossDividend.Amount)
	}
	if payload.Data.GrossYield == nil || *payload.Data.GrossYield != "0.12442857" {
		t.Fatalf("grossYield = %v", payload.Data.GrossYield)
	}
	if store.command.SubjectID != dividendReplaySubjectID("stage-03-68-route-key-000001") {
		t.Fatalf("unexpected technical replay scope %s", store.command.SubjectID)
	}
	if store.command.RequestPath != "/api/v1/dividends/calculate" {
		t.Fatalf("request path = %s", store.command.RequestPath)
	}
}

func TestCalculateDividendExactReplayBypassesFreshAdmissionLimiter(t *testing.T) {
	artifact := verticalslice.CommandReplayArtifact{
		StatusCode: http.StatusOK,
		Body:       []byte(`{"data":{"ticker":"SBER"},"meta":{"requestId":"10000000-0000-4000-8000-000000000001","traceId":"10000000000000000000000000000001","generatedAt":"2026-09-05T00:00:00Z"}}`),
		RequestID:  "10000000-0000-4000-8000-000000000001",
		TraceID:    "10000000000000000000000000000001",
	}
	store := &dividendHTTPTestStore{replayArtifact: &artifact}
	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	app := newApp(&API{
		service:         service,
		dividendLimiter: newBoundedAuthRateLimiter(0, 0, 0, defaultDividendCalculatorWindow),
	})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/dividends/calculate", bytes.NewBufferString(`{"ticker":"SBER","quantity":"1.00000000","dividendPerUnit":{"amount":"1.00000000","currency":"RUB"}}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "stage-03-68-replay-key-000001")

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("replay status = %d, want 200", response.StatusCode)
	}
	if store.writeCalls != 0 {
		t.Fatalf("exact replay must bypass writable reservation: calls = %d", store.writeCalls)
	}
}

func TestCalculateDividendFreshAdmissionCanRateLimitWithoutBuildingResponse(t *testing.T) {
	store := &dividendHTTPTestStore{}
	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	app := newApp(&API{
		service:         service,
		dividendLimiter: newBoundedAuthRateLimiter(0, 0, 0, defaultDividendCalculatorWindow),
	})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/dividends/calculate", bytes.NewBufferString(`{"ticker":"SBER","quantity":"1.00000000","dividendPerUnit":{"amount":"1.00000000","currency":"RUB"}}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "stage-03-68-rate-key-000001")

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("fresh admission status = %d, want 429", response.StatusCode)
	}
	if response.Header.Get("Retry-After") != dividendCalculatorRateLimitRetryAfterSeconds {
		t.Fatalf("Retry-After = %q, want %q", response.Header.Get("Retry-After"), dividendCalculatorRateLimitRetryAfterSeconds)
	}
	if store.writeCalls != 0 {
		t.Fatalf("rate-limited fresh command must not reach writable reservation: calls = %d", store.writeCalls)
	}
}

func TestCalculateDividendRateLimitRaceRechecksReplayBeforeReturning429(t *testing.T) {
	artifact := verticalslice.CommandReplayArtifact{
		StatusCode: http.StatusOK,
		Body:       []byte(`{"data":{"ticker":"SBER"},"meta":{"requestId":"10000000-0000-4000-8000-000000000001","traceId":"10000000000000000000000000000001","generatedAt":"2026-09-05T00:00:00Z"}}`),
		RequestID:  "10000000-0000-4000-8000-000000000001",
		TraceID:    "10000000000000000000000000000001",
	}
	store := &dividendHTTPTestStore{replayArtifact: &artifact, replayOnLookup: 2}
	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	app := newApp(&API{
		service:         service,
		dividendLimiter: newBoundedAuthRateLimiter(0, 0, 0, defaultDividendCalculatorWindow),
	})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/dividends/calculate", bytes.NewBufferString(`{"ticker":"SBER","quantity":"1.00000000","dividendPerUnit":{"amount":"1.00000000","currency":"RUB"}}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "stage-03-68-race-key-000001")

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("race replay status = %d, want 200", response.StatusCode)
	}
	if store.lookupCalls != 2 {
		t.Fatalf("rate-limit denial must perform one replay race recheck: lookups = %d, want 2", store.lookupCalls)
	}
	if store.writeCalls != 0 {
		t.Fatalf("race replay must not enter writable reservation: calls = %d", store.writeCalls)
	}
}

func TestCalculateDividendRouteRejectsMissingIdempotencyBeforeReplay(t *testing.T) {
	store := &dividendHTTPTestStore{}
	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	app := newApp(&API{service: service, dividendLimiter: newDividendCalculatorRateLimiter()})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/dividends/calculate", bytes.NewBufferString(`{"ticker":"SBER"}`))
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d", response.StatusCode)
	}
	if store.command.SubjectID != "" {
		t.Fatal("missing idempotency key must not reach replay persistence")
	}
}

func testTime() time.Time {
	return time.Unix(1788610000, 0).UTC()
}
