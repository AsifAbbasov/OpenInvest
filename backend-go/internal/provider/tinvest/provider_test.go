package tinvest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

type fixedClock struct{ now time.Time }

func (clock fixedClock) Now() time.Time { return clock.now }

type nilableClock struct{}

func (*nilableClock) Now() time.Time { return time.Time{} }

type countingClock struct {
	now   time.Time
	calls atomic.Int32
}

func (clock *countingClock) Now() time.Time {
	clock.calls.Add(1)
	return clock.now
}

const testToken = "test-readonly-token"

var testNow = time.Date(2026, 9, 9, 10, 0, 0, 0, time.UTC)

func newTestProvider(t *testing.T, handler http.Handler) (*Provider, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	provider, err := newCorporateActionProvider(server.Client(), fixedClock{now: testNow}, testToken, server.URL+"/rest")
	if err != nil {
		server.Close()
		t.Fatalf("new provider: %v", err)
	}
	return provider, server
}

func testQuery(instrumentIDs ...string) verticalslice.CorporateActionQuery {
	return verticalslice.CorporateActionQuery{
		InstrumentIDs: instrumentIDs,
		From:          "2026-09-01",
		To:            "2026-09-30",
	}
}

func TestCorporateActionsRoutesSharesOnlyToGetDividends(t *testing.T) {
	var requests atomic.Int32
	provider, server := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		assertProviderRequest(t, r, getDividendsMethod, "SBER_TQBR")
		_, _ = io.WriteString(w, `{"dividends":[{"dividendNet":{"currency":"rub","units":"12","nano":340000000},"paymentDate":"2026-09-20T00:00:00Z","recordDate":"2026-09-10T00:00:00Z","dividendType":"Regular Cash"}]}`)
	}))
	defer server.Close()

	events, err := provider.CorporateActions(context.Background(), testQuery("SBER"))
	if err != nil {
		t.Fatalf("corporate actions: %v", err)
	}
	if requests.Load() != 1 {
		t.Fatalf("requests: %d", requests.Load())
	}
	if len(events) != 1 {
		t.Fatalf("events: %d", len(events))
	}
	event := events[0]
	if event.InstrumentID != "SBER" || event.Kind != verticalslice.CorporateActionDividend {
		t.Fatalf("event identity: %+v", event)
	}
	if event.Status != verticalslice.CorporateActionAnnounced {
		t.Fatalf("status: %s", event.Status)
	}
	if event.RecordDate == nil || *event.RecordDate != "2026-09-10" {
		t.Fatalf("recordDate: %v", event.RecordDate)
	}
	if event.PaymentDate == nil || *event.PaymentDate != "2026-09-20" {
		t.Fatalf("paymentDate: %v", event.PaymentDate)
	}
	if event.AmountPerUnit == nil || event.AmountPerUnit.Amount.String() != "12.34000000" || event.AmountPerUnit.Currency != "RUB" {
		t.Fatalf("amount: %+v", event.AmountPerUnit)
	}
	if event.Provenance.Provider != providerID || !strings.HasPrefix(event.Provenance.SourceEventID, "sha256:") {
		t.Fatalf("provenance: %+v", event.Provenance)
	}
	if event.AsOf != testNow || event.RetrievedAt != testNow {
		t.Fatalf("timestamps: asOf=%s retrievedAt=%s", event.AsOf, event.RetrievedAt)
	}
}

func TestCorporateActionsMapsExplicitCancelledDividendWithoutInferringOtherLifecycle(t *testing.T) {
	provider, server := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"dividends":[{"dividendNet":{"currency":"rub","units":"12","nano":0},"paymentDate":"2026-09-20T00:00:00Z","recordDate":"2026-09-10T00:00:00Z","dividendType":"Cancelled"}]}`)
	}))
	defer server.Close()

	events, err := provider.CorporateActions(context.Background(), testQuery("SBER"))
	if err != nil {
		t.Fatalf("corporate actions: %v", err)
	}
	if len(events) != 1 || events[0].Status != verticalslice.CorporateActionCancelled {
		t.Fatalf("cancelled dividend mapping: %+v", events)
	}
}

func TestCorporateActionsRejectsDividendTypesOutsideCanonicalCashDividendScope(t *testing.T) {
	for _, dividendType := range []string{"Return of Capital", "Daily Accrual", "Unknown Provider Type"} {
		t.Run(dividendType, func(t *testing.T) {
			provider, server := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = io.WriteString(w, fmt.Sprintf(`{"dividends":[{"dividendNet":{"currency":"rub","units":"12","nano":0},"paymentDate":"2026-09-20T00:00:00Z","recordDate":"2026-09-10T00:00:00Z","dividendType":%q}]}`, dividendType))
			}))
			defer server.Close()

			_, err := provider.CorporateActions(context.Background(), testQuery("SBER"))
			if !errors.Is(err, verticalslice.ErrCorporateActionsProviderData) {
				t.Fatalf("expected provider data error, got %v", err)
			}
		})
	}
}

func TestCorporateActionsStampsRetrievalOnlyAfterProviderResponse(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	clock := &countingClock{now: testNow}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(started)
		<-release
		_, _ = io.WriteString(w, `{"dividends":[{"dividendNet":{"currency":"rub","units":"12","nano":0},"paymentDate":"2026-09-20T00:00:00Z","recordDate":"2026-09-10T00:00:00Z","dividendType":"Regular Cash"}]}`)
	}))
	defer server.Close()
	provider, err := newCorporateActionProvider(server.Client(), clock, testToken, server.URL+"/rest")
	if err != nil {
		t.Fatal(err)
	}

	result := make(chan struct {
		events []verticalslice.CorporateActionEvent
		err    error
	}, 1)
	go func() {
		events, err := provider.CorporateActions(context.Background(), testQuery("SBER"))
		result <- struct {
			events []verticalslice.CorporateActionEvent
			err    error
		}{events: events, err: err}
	}()

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("provider request did not start")
	}
	if got := clock.calls.Load(); got != 0 {
		t.Fatalf("retrieval clock sampled before provider response: %d", got)
	}
	close(release)
	got := <-result
	if got.err != nil {
		t.Fatal(got.err)
	}
	if clock.calls.Load() != 1 || len(got.events) != 1 || got.events[0].RetrievedAt != testNow {
		t.Fatalf("retrieval timestamp evidence: calls=%d events=%+v", clock.calls.Load(), got.events)
	}
}

func TestCorporateActionsRoutesBondsOnlyToGetBondCouponsAndPreservesMissingFixDate(t *testing.T) {
	provider, server := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertProviderRequest(t, r, getBondCouponsMethod, "SU26238RMFS4_TQOB")
		_, _ = io.WriteString(w, `{"events":[{"couponDate":"2026-09-21T00:00:00Z","couponNumber":"11","payOneBond":{"currency":"rub","units":"35","nano":400000000}}]}`)
	}))
	defer server.Close()

	events, err := provider.CorporateActions(context.Background(), testQuery("SU26238RMFS4"))
	if err != nil {
		t.Fatalf("corporate actions: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("events: %d", len(events))
	}
	event := events[0]
	if event.Kind != verticalslice.CorporateActionCoupon || event.Status != verticalslice.CorporateActionAnnounced {
		t.Fatalf("kind/status: %s/%s", event.Kind, event.Status)
	}
	if event.RecordDate != nil {
		t.Fatalf("recordDate must stay nil, got %v", *event.RecordDate)
	}
	if event.PaymentDate == nil || *event.PaymentDate != "2026-09-21" {
		t.Fatalf("paymentDate: %v", event.PaymentDate)
	}
	if event.AmountPerUnit == nil || event.AmountPerUnit.Amount.String() != "35.40000000" {
		t.Fatalf("amount: %+v", event.AmountPerUnit)
	}
}

func TestCorporateActionsMixedQueryMakesExactlyOneApprovedCallPerInstrument(t *testing.T) {
	var mu sync.Mutex
	counts := map[string]int{}
	provider, server := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request corporateActionRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		key := r.URL.Path + "|" + request.InstrumentID
		mu.Lock()
		counts[key]++
		mu.Unlock()
		switch {
		case strings.HasSuffix(r.URL.Path, "/"+getDividendsMethod):
			_, _ = io.WriteString(w, `{"dividends":[]}`)
		case strings.HasSuffix(r.URL.Path, "/"+getBondCouponsMethod):
			_, _ = io.WriteString(w, `{"events":[]}`)
		default:
			t.Fatalf("unauthorized path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	_, err := provider.CorporateActions(context.Background(), testQuery("SBER", "GAZP", "SU26238RMFS4"))
	if err != nil {
		t.Fatalf("corporate actions: %v", err)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(counts) != 3 {
		t.Fatalf("unique calls: %#v", counts)
	}
	for _, want := range []string{
		"/rest/" + instrumentsService + "/" + getDividendsMethod + "|SBER_TQBR",
		"/rest/" + instrumentsService + "/" + getDividendsMethod + "|GAZP_TQBR",
		"/rest/" + instrumentsService + "/" + getBondCouponsMethod + "|SU26238RMFS4_TQOB",
	} {
		if counts[want] != 1 {
			t.Fatalf("call %q count=%d; all=%#v", want, counts[want], counts)
		}
	}
}

func TestCorporateActionsUnknownInstrumentFailsBeforeNetwork(t *testing.T) {
	var requests atomic.Int32
	provider, server := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	_, err := provider.CorporateActions(context.Background(), testQuery("LKOH"))
	if !errors.Is(err, verticalslice.ErrCorporateActionsProviderUnavailable) {
		t.Fatalf("expected unavailable, got %v", err)
	}
	if requests.Load() != 0 {
		t.Fatalf("unexpected network calls: %d", requests.Load())
	}
}

func TestCorporateActionsRejectsMoneyThatWouldRequireRounding(t *testing.T) {
	provider, server := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"dividends":[{"dividendNet":{"currency":"rub","units":"1","nano":1},"paymentDate":"2026-09-20T00:00:00Z","recordDate":"2026-09-10T00:00:00Z","dividendType":"Regular Cash"}]}`)
	}))
	defer server.Close()

	_, err := provider.CorporateActions(context.Background(), testQuery("SBER"))
	if !errors.Is(err, verticalslice.ErrCorporateActionsProviderData) {
		t.Fatalf("expected provider data error, got %v", err)
	}
}

func TestCorporateActionsRejectsMoneyWithInconsistentUnitsAndNanoSigns(t *testing.T) {
	for _, payload := range []string{
		`{"currency":"rub","units":"1","nano":-100000000}`,
		`{"currency":"rub","units":"-1","nano":100000000}`,
	} {
		t.Run(payload, func(t *testing.T) {
			provider, server := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = io.WriteString(w, `{"dividends":[{"dividendNet":`+payload+`,"paymentDate":"2026-09-20T00:00:00Z","recordDate":"2026-09-10T00:00:00Z","dividendType":"Regular Cash"}]}`)
			}))
			defer server.Close()

			_, err := provider.CorporateActions(context.Background(), testQuery("SBER"))
			if !errors.Is(err, verticalslice.ErrCorporateActionsProviderData) {
				t.Fatalf("expected provider data error, got %v", err)
			}
		})
	}
}

func TestCorporateActionsPreservesUnknownDividendFieldsAsNil(t *testing.T) {
	provider, server := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"dividends":[{"dividendType":"Regular Cash"}]}`)
	}))
	defer server.Close()

	events, err := provider.CorporateActions(context.Background(), testQuery("SBER"))
	if err != nil {
		t.Fatalf("corporate actions: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("events: %d", len(events))
	}
	if events[0].RecordDate != nil || events[0].PaymentDate != nil || events[0].AmountPerUnit != nil {
		t.Fatalf("unknown provider fields were fabricated: %+v", events[0])
	}
}

func TestCorporateActionsPreservesUnknownCouponFieldsAsNil(t *testing.T) {
	provider, server := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"events":[{"couponNumber":"11"}]}`)
	}))
	defer server.Close()

	events, err := provider.CorporateActions(context.Background(), testQuery("SU26238RMFS4"))
	if err != nil {
		t.Fatalf("corporate actions: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("events: %d", len(events))
	}
	if events[0].RecordDate != nil || events[0].PaymentDate != nil || events[0].AmountPerUnit != nil {
		t.Fatalf("unknown provider fields were fabricated: %+v", events[0])
	}
}

func TestCorporateActionsRejectsNonPositiveCouponNumber(t *testing.T) {
	for _, couponNumber := range []string{"0", "-1"} {
		t.Run(couponNumber, func(t *testing.T) {
			provider, server := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = io.WriteString(w, `{"events":[{"couponNumber":"`+couponNumber+`"}]}`)
			}))
			defer server.Close()

			_, err := provider.CorporateActions(context.Background(), testQuery("SU26238RMFS4"))
			if !errors.Is(err, verticalslice.ErrCorporateActionsProviderData) {
				t.Fatalf("expected provider data error, got %v", err)
			}
		})
	}
}

func TestCorporateActionsDoesNotInferPaidFromPastPaymentDate(t *testing.T) {
	provider, server := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"dividends":[{"dividendNet":{"currency":"rub","units":"5","nano":0},"paymentDate":"2020-01-02T00:00:00Z","recordDate":"2020-01-01T00:00:00Z","dividendType":"Regular Cash"}]}`)
	}))
	defer server.Close()

	events, err := provider.CorporateActions(context.Background(), verticalslice.CorporateActionQuery{InstrumentIDs: []string{"SBER"}, From: "2020-01-01", To: "2020-01-31"})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Status != verticalslice.CorporateActionAnnounced {
		t.Fatalf("historical schedule must remain ANNOUNCED: %+v", events)
	}
}

func TestCorporateActionsDeduplicatesExactProviderRows(t *testing.T) {
	row := `{"dividendNet":{"currency":"rub","units":"5","nano":0},"paymentDate":"2026-09-20T00:00:00Z","recordDate":"2026-09-10T00:00:00Z","dividendType":"Regular Cash"}`
	provider, server := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"dividends":[`+row+`,`+row+`]}`)
	}))
	defer server.Close()

	events, err := provider.CorporateActions(context.Background(), testQuery("SBER"))
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("deduplicated events: %d", len(events))
	}
}

func TestCorporateActionsEventIDIsDeterministic(t *testing.T) {
	provider, server := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"dividends":[{"dividendNet":{"currency":"rub","units":"5","nano":0},"paymentDate":"2026-09-20T00:00:00Z","recordDate":"2026-09-10T00:00:00Z","dividendType":"Regular Cash"}]}`)
	}))
	defer server.Close()

	first, err := provider.CorporateActions(context.Background(), testQuery("SBER"))
	if err != nil {
		t.Fatal(err)
	}
	second, err := provider.CorporateActions(context.Background(), testQuery("SBER"))
	if err != nil {
		t.Fatal(err)
	}
	if first[0].EventID != second[0].EventID || first[0].Provenance.SourceEventID != second[0].Provenance.SourceEventID {
		t.Fatalf("identity drift: first=%+v second=%+v", first[0].Provenance, second[0].Provenance)
	}
}

func TestCorporateActionsErrorClassIsDeterministicAcrossConcurrentCompletionOrder(t *testing.T) {
	for _, dataDelay := range []time.Duration{0, 40 * time.Millisecond} {
		t.Run(dataDelay.String(), func(t *testing.T) {
			var requests atomic.Int32
			provider, server := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				var request corporateActionRequest
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Fatalf("decode request: %v", err)
				}
				switch request.InstrumentID {
				case "SBER_TQBR":
					time.Sleep(dataDelay)
					w.WriteHeader(http.StatusBadRequest)
				case "GAZP_TQBR":
					time.Sleep(40*time.Millisecond - dataDelay)
					w.WriteHeader(http.StatusInternalServerError)
				default:
					t.Fatalf("unexpected instrument: %s", request.InstrumentID)
				}
			}))
			defer server.Close()

			_, err := provider.CorporateActions(context.Background(), testQuery("SBER", "GAZP"))
			if !errors.Is(err, verticalslice.ErrCorporateActionsProviderData) {
				t.Fatalf("completion order changed error class: %v", err)
			}
			if requests.Load() != 2 {
				t.Fatalf("bounded request set was not completed: %d", requests.Load())
			}
		})
	}
}

func TestCorporateActionsHTTPClassificationAndNoRetry(t *testing.T) {
	for _, tc := range []struct {
		status int
		want   error
	}{
		{http.StatusUnauthorized, verticalslice.ErrCorporateActionsProviderUnavailable},
		{http.StatusForbidden, verticalslice.ErrCorporateActionsProviderUnavailable},
		{http.StatusRequestTimeout, verticalslice.ErrCorporateActionsProviderUnavailable},
		{http.StatusTooManyRequests, verticalslice.ErrCorporateActionsProviderUnavailable},
		{http.StatusInternalServerError, verticalslice.ErrCorporateActionsProviderUnavailable},
		{http.StatusBadRequest, verticalslice.ErrCorporateActionsProviderData},
		{http.StatusNotFound, verticalslice.ErrCorporateActionsProviderData},
	} {
		t.Run(fmt.Sprint(tc.status), func(t *testing.T) {
			var requests atomic.Int32
			provider, server := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				requests.Add(1)
				w.WriteHeader(tc.status)
			}))
			defer server.Close()
			_, err := provider.CorporateActions(context.Background(), testQuery("SBER"))
			if !errors.Is(err, tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, err)
			}
			if requests.Load() != 1 {
				t.Fatalf("automatic retry detected: %d requests", requests.Load())
			}
		})
	}
}

func TestCorporateActionsAcceptsOmittedOrNullRepeatedFieldsAsLegitimateEmptyProtoJSON(t *testing.T) {
	for _, tc := range []struct {
		name       string
		instrument string
		body       string
	}{
		{"dividends omitted", "SBER", `{}`},
		{"dividends null", "SBER", `{"dividends":null}`},
		{"coupons omitted", "SU26238RMFS4", `{}`},
		{"coupons null", "SU26238RMFS4", `{"events":null}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			provider, server := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = io.WriteString(w, tc.body)
			}))
			defer server.Close()

			events, err := provider.CorporateActions(context.Background(), testQuery(tc.instrument))
			if err != nil {
				t.Fatalf("legitimate empty protobuf response rejected: %v", err)
			}
			if len(events) != 0 {
				t.Fatalf("events=%d, want 0", len(events))
			}
		})
	}
}

func TestCorporateActionsRejectsMalformedTrailingAndOversizedResponses(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
	}{
		{"malformed", `{"dividends":`},
		{"trailing", `{"dividends":[]} {}`},
		{"top-level null", `null`},
		{"top-level array", `[]`},
		{"oversized", strings.Repeat("x", int(maxResponseBodyBytes)+1)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			provider, server := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = io.WriteString(w, tc.body)
			}))
			defer server.Close()
			_, err := provider.CorporateActions(context.Background(), testQuery("SBER"))
			if !errors.Is(err, verticalslice.ErrCorporateActionsProviderData) {
				t.Fatalf("expected provider data error, got %v", err)
			}
		})
	}
}

func TestCorporateActionsNeverForwardsAuthorizationAcrossRedirect(t *testing.T) {
	var redirectedRequests atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirectedRequests.Add(1)
		if got := r.Header.Get("Authorization"); got != "" {
			t.Fatalf("authorization leaked to redirect target: %q", got)
		}
		_, _ = io.WriteString(w, `{"dividends":[]}`)
	}))
	defer target.Close()

	provider, server := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusFound)
	}))
	defer server.Close()

	_, err := provider.CorporateActions(context.Background(), testQuery("SBER"))
	if !errors.Is(err, verticalslice.ErrCorporateActionsProviderData) {
		t.Fatalf("expected redirect rejection as provider data error, got %v", err)
	}
	if redirectedRequests.Load() != 0 {
		t.Fatalf("redirect target was contacted %d times", redirectedRequests.Load())
	}
}

func TestCorporateActionsErrorsNeverContainTokenOrRawResponseBody(t *testing.T) {
	secretBody := `{"error":"` + testToken + `"}`
	provider, server := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, secretBody)
	}))
	defer server.Close()

	_, err := provider.CorporateActions(context.Background(), testQuery("SBER"))
	if err == nil {
		t.Fatal("expected provider error")
	}
	if strings.Contains(err.Error(), testToken) || strings.Contains(err.Error(), secretBody) {
		t.Fatalf("provider error leaked secret material: %q", err.Error())
	}
}

func TestCorporateActionsGlobalConcurrencyNeverExceedsFourAndDoesNotQueueCallers(t *testing.T) {
	var active atomic.Int32
	var maximum atomic.Int32
	started := make(chan struct{}, maxConcurrency)
	release := make(chan struct{})

	provider, server := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		current := active.Add(1)
		for {
			seen := maximum.Load()
			if current <= seen || maximum.CompareAndSwap(seen, current) {
				break
			}
		}
		started <- struct{}{}
		<-release
		active.Add(-1)
		_, _ = io.WriteString(w, `{"dividends":[]}`)
	}))
	defer server.Close()

	primaryErrs := make(chan error, maxConcurrency)
	for i := 0; i < maxConcurrency; i++ {
		go func() {
			_, err := provider.CorporateActions(context.Background(), testQuery("SBER"))
			primaryErrs <- err
		}()
	}
	for i := 0; i < maxConcurrency; i++ {
		select {
		case <-started:
		case <-time.After(2 * time.Second):
			t.Fatal("did not reach expected concurrency")
		}
	}

	const extraCallers = 8
	extraErrs := make(chan error, extraCallers)
	for i := 0; i < extraCallers; i++ {
		go func() {
			_, err := provider.CorporateActions(context.Background(), testQuery("SBER"))
			extraErrs <- err
		}()
	}
	for i := 0; i < extraCallers; i++ {
		select {
		case err := <-extraErrs:
			if !errors.Is(err, verticalslice.ErrCorporateActionsProviderUnavailable) {
				t.Fatalf("extra caller did not fail closed: %v", err)
			}
		case <-time.After(500 * time.Millisecond):
			t.Fatal("extra caller queued behind provider semaphore")
		}
	}
	if got := maximum.Load(); got > maxConcurrency {
		t.Fatalf("max concurrency: %d", got)
	}

	close(release)
	for i := 0; i < maxConcurrency; i++ {
		if err := <-primaryErrs; err != nil {
			t.Fatalf("primary caller error: %v", err)
		}
	}
}

func TestCorporateActionsGlobalBudgetStopsSixtyFirstRequest(t *testing.T) {
	var requests atomic.Int32
	provider, server := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		_, _ = io.WriteString(w, `{"dividends":[]}`)
	}))
	defer server.Close()

	for i := 0; i < maxRequestsPerMinute; i++ {
		if _, err := provider.CorporateActions(context.Background(), testQuery("SBER")); err != nil {
			t.Fatalf("request %d: %v", i+1, err)
		}
	}
	_, err := provider.CorporateActions(context.Background(), testQuery("SBER"))
	if !errors.Is(err, verticalslice.ErrCorporateActionsProviderUnavailable) {
		t.Fatalf("expected internal throttling, got %v", err)
	}
	if requests.Load() != maxRequestsPerMinute {
		t.Fatalf("network requests: %d", requests.Load())
	}
}

func TestCorporateActionsHonorsProviderRateLimitHeadersWithoutRetry(t *testing.T) {
	var requests atomic.Int32
	provider, server := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		w.Header().Set("X-RateLimit-Limit", "200")
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Reset", "60")
		_, _ = io.WriteString(w, `{"dividends":[]}`)
	}))
	defer server.Close()

	if _, err := provider.CorporateActions(context.Background(), testQuery("SBER")); err != nil {
		t.Fatalf("first request: %v", err)
	}
	_, err := provider.CorporateActions(context.Background(), testQuery("SBER"))
	if !errors.Is(err, verticalslice.ErrCorporateActionsProviderUnavailable) {
		t.Fatalf("expected provider-header throttling, got %v", err)
	}
	if requests.Load() != 1 {
		t.Fatalf("provider rate-limit header was ignored: %d network requests", requests.Load())
	}
}

func TestRequestBudgetProviderHeadersCannotIncreaseKnownAllowance(t *testing.T) {
	now := testNow
	budget := newRequestBudget(60, time.Minute, func() time.Time { return now })
	firstHeaders := make(http.Header)
	firstHeaders.Set("X-RateLimit-Limit", "200")
	firstHeaders.Set("X-RateLimit-Remaining", "2")
	firstHeaders.Set("X-RateLimit-Reset", "60")
	budget.ObserveProviderHeaders(firstHeaders)
	if !budget.Allow() {
		t.Fatal("expected first remotely permitted request")
	}
	// A stale/out-of-order response must not increase the already known remaining allowance.
	staleHeaders := make(http.Header)
	staleHeaders.Set("X-RateLimit-Limit", "200")
	staleHeaders.Set("X-RateLimit-Remaining", "9")
	staleHeaders.Set("X-RateLimit-Reset", "60")
	budget.ObserveProviderHeaders(staleHeaders)
	if !budget.Allow() {
		t.Fatal("expected second remotely permitted request")
	}
	if budget.Allow() {
		t.Fatal("stale provider header increased remote allowance")
	}
	now = now.Add(61 * time.Second)
	if !budget.Allow() {
		t.Fatal("provider-header throttle did not expire after reset window")
	}
}

func TestCorporateActionsPropagatesCallerCancellation(t *testing.T) {
	provider, err := newCorporateActionProvider(&http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		<-request.Context().Done()
		return nil, request.Context().Err()
	})}, fixedClock{now: testNow}, testToken, "https://example.invalid/rest")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = provider.CorporateActions(ctx, testQuery("SBER"))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected caller cancellation, got %v", err)
	}
}

func TestConstructorHardensHTTPClientWithoutMutatingCaller(t *testing.T) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Timeout: 17 * time.Second, Jar: jar}
	provider, err := newCorporateActionProvider(client, fixedClock{now: testNow}, testToken, "https://example.com/rest")
	if err != nil {
		t.Fatal(err)
	}
	if provider.client.Timeout != requestTimeout || provider.client.Jar != nil || provider.client.CheckRedirect == nil {
		t.Fatalf("provider client not hardened: timeout=%s jar=%v redirect=%v", provider.client.Timeout, provider.client.Jar, provider.client.CheckRedirect != nil)
	}
	if client.Timeout != 17*time.Second || client.Jar != jar || client.CheckRedirect != nil {
		t.Fatal("caller-owned client was mutated")
	}
}

func TestConstructorRejectsMissingDependenciesTokenAndInvalidURL(t *testing.T) {
	clock := fixedClock{now: testNow}
	if _, err := newCorporateActionProvider(nil, clock, testToken, "https://example.com/rest"); err == nil {
		t.Fatal("expected nil client error")
	}
	if _, err := newCorporateActionProvider(&http.Client{}, nil, testToken, "https://example.com/rest"); err == nil {
		t.Fatal("expected nil clock error")
	}
	var typedNilClock *nilableClock
	if _, err := newCorporateActionProvider(&http.Client{}, typedNilClock, testToken, "https://example.com/rest"); err == nil {
		t.Fatal("expected typed nil clock error")
	}
	for _, token := range []string{"", " token", "token ", "to ken", "token\n"} {
		if _, err := newCorporateActionProvider(&http.Client{}, clock, token, "https://example.com/rest"); err == nil {
			t.Fatalf("expected token rejection for %q", token)
		}
	}
	for _, rawURL := range []string{"", "://bad", "ftp://example.com", "https://user@example.com/rest", "https://example.com/rest?q=x", "https://example.com/rest#x"} {
		if _, err := newCorporateActionProvider(&http.Client{}, clock, testToken, rawURL); err == nil {
			t.Fatalf("expected invalid URL rejection for %q", rawURL)
		}
	}
}

func assertProviderRequest(t *testing.T, r *http.Request, method, instrumentID string) {
	t.Helper()
	if r.Method != http.MethodPost {
		t.Fatalf("method: %s", r.Method)
	}
	if r.URL.Path != "/rest/"+instrumentsService+"/"+method {
		t.Fatalf("path: %s", r.URL.Path)
	}
	if got := r.Header.Get("Authorization"); got != "Bearer "+testToken {
		t.Fatalf("authorization header: %q", got)
	}
	if got := r.Header.Get("Accept"); got != "application/json" {
		t.Fatalf("accept: %q", got)
	}
	if got := r.Header.Get("Content-Type"); got != "application/json" {
		t.Fatalf("content-type: %q", got)
	}
	var request corporateActionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		t.Fatalf("decode request: %v", err)
	}
	if request.InstrumentID != instrumentID {
		t.Fatalf("instrumentId: %q", request.InstrumentID)
	}
	if request.From != "2026-09-01T00:00:00Z" {
		t.Fatalf("from: %q", request.From)
	}
	if request.To != "2026-09-30T23:59:59.999999999Z" {
		t.Fatalf("to: %q", request.To)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return fn(request) }
