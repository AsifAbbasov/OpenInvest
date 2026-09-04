package moexiss

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

type fixedClock struct{ now time.Time }

func (clock fixedClock) Now() time.Time { return clock.now }

var deterministicNow = time.Date(2026, 9, 4, 12, 30, 0, 0, time.UTC)

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

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return fn(request) }

func validPayload(last string) string {
	return fmt.Sprintf(`{
      "marketdata":{"columns":["SECID","BOARDID","LAST","TIME"],"data":[["SBER","TQBR",%s,"15:04:05"]]},
      "dataversion":{"columns":["trade_date"],"data":[["2026-09-04"]]}
    }`, last)
}

func newTestProvider(t *testing.T, handler http.Handler, clock verticalslice.Clock) (*Provider, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	provider, err := newQuoteProvider(server.Client(), clock, server.URL)
	if err != nil {
		server.Close()
		t.Fatalf("new provider: %v", err)
	}
	return provider, server
}

func TestQuoteValidContract(t *testing.T) {
	retrievedAt := time.Date(2026, 9, 4, 12, 30, 0, 123, time.FixedZone("UTC+4", 4*60*60))
	provider, server := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method: %s", r.Method)
		}
		if r.URL.Path != "/iss/engines/stock/markets/shares/boards/TQBR/securities/SBER.json" {
			t.Fatalf("path: %s", r.URL.Path)
		}
		query := r.URL.Query()
		if query.Get("iss.meta") != "off" {
			t.Fatalf("iss.meta: %q", query.Get("iss.meta"))
		}
		if query.Get("iss.only") != "marketdata,dataversion" {
			t.Fatalf("iss.only: %q", query.Get("iss.only"))
		}
		if query.Get("marketdata.columns") != "SECID,BOARDID,LAST,TIME" {
			t.Fatalf("marketdata.columns: %q", query.Get("marketdata.columns"))
		}
		if query.Get("dataversion.columns") != "trade_date" {
			t.Fatalf("dataversion.columns: %q", query.Get("dataversion.columns"))
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Fatalf("accept: %q", r.Header.Get("Accept"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, validPayload("321.12345678"))
	}), fixedClock{now: retrievedAt})
	defer server.Close()

	quote, err := provider.Quote(context.Background(), "SBER")
	if err != nil {
		t.Fatalf("quote: %v", err)
	}
	if quote.Ticker != "SBER" {
		t.Fatalf("ticker: %q", quote.Ticker)
	}
	if got := quote.Price.Amount.String(); got != "321.12345678" {
		t.Fatalf("price: %s", got)
	}
	if quote.Price.Currency != verticalslice.RUB {
		t.Fatalf("currency: %s", quote.Price.Currency)
	}
	wantAsOf := time.Date(2026, 9, 4, 12, 4, 5, 0, time.UTC)
	if !quote.AsOf.Equal(wantAsOf) || quote.AsOf.Location() != time.UTC {
		t.Fatalf("asOf: %s (%v)", quote.AsOf, quote.AsOf.Location())
	}
	if !quote.RetrievedAt.Equal(retrievedAt.UTC()) || quote.RetrievedAt.Location() != time.UTC {
		t.Fatalf("retrievedAt: %s (%v)", quote.RetrievedAt, quote.RetrievedAt.Location())
	}
	if quote.Provenance.Provider != providerID {
		t.Fatalf("provider: %q", quote.Provenance.Provider)
	}
}

func TestQuoteStampsRetrievedAtExactlyOnce(t *testing.T) {
	clock := &countingClock{now: time.Date(2026, 9, 4, 12, 30, 0, 0, time.UTC)}
	provider, server := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, validPayload("1.23"))
	}), clock)
	defer server.Close()

	quote, err := provider.Quote(context.Background(), "SBER")
	if err != nil {
		t.Fatalf("quote: %v", err)
	}
	if got := clock.calls.Load(); got != 1 {
		t.Fatalf("expected one clock read, got %d", got)
	}
	if !quote.RetrievedAt.Equal(clock.now) {
		t.Fatalf("retrievedAt: %s", quote.RetrievedAt)
	}
}

func TestQuoteAcceptsReorderedAndExtraColumns(t *testing.T) {
	payload := `{
      "marketdata":{"columns":["TIME","EXTRA","LAST","BOARDID","SECID"],"data":[["15:04:05","ignored",321.12345678,"TQBR","SBER"]]},
      "dataversion":{"columns":["other","trade_date"],"data":[["ignored","2026-09-04"]]}
    }`
	provider, server := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, payload) }), fixedClock{now: time.Date(2026, 9, 4, 12, 30, 0, 0, time.UTC)})
	defer server.Close()
	quote, err := provider.Quote(context.Background(), "SBER")
	if err != nil {
		t.Fatalf("quote: %v", err)
	}
	if quote.Price.Amount.String() != "321.12345678" {
		t.Fatalf("price: %s", quote.Price.Amount.String())
	}
}

func TestQuoteAbsence(t *testing.T) {
	tests := []struct{ name, payload string }{
		{"empty marketdata", `{"marketdata":{"columns":["SECID","BOARDID","LAST","TIME"],"data":[]},"dataversion":{"columns":["trade_date"],"data":[["2026-09-04"]]}}`},
		{"null last", `{"marketdata":{"columns":["SECID","BOARDID","LAST","TIME"],"data":[["SBER","TQBR",null,"15:04:05"]]},"dataversion":{"columns":["trade_date"],"data":[["2026-09-04"]]}}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			provider, server := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, tc.payload) }), fixedClock{now: deterministicNow})
			defer server.Close()
			_, err := provider.Quote(context.Background(), "SBER")
			if !errors.Is(err, verticalslice.ErrMarketQuoteNotFound) {
				t.Fatalf("expected not found, got %v", err)
			}
		})
	}
}

func TestQuoteAbsenceRequiresValidDataVersionEvidence(t *testing.T) {
	payload := `{"marketdata":{"columns":["SECID","BOARDID","LAST","TIME"],"data":[]},"dataversion":{"columns":["trade_date"],"data":[]}}`
	provider, server := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, payload)
	}), fixedClock{now: deterministicNow})
	defer server.Close()

	_, err := provider.Quote(context.Background(), "SBER")
	if !errors.Is(err, verticalslice.ErrMarketQuoteProviderData) {
		t.Fatalf("expected provider data error, got %v", err)
	}
}

func TestQuoteHTTPClassification(t *testing.T) {
	for _, tc := range []struct {
		status int
		want   error
	}{
		{http.StatusTooManyRequests, verticalslice.ErrMarketQuoteProviderUnavailable},
		{http.StatusInternalServerError, verticalslice.ErrMarketQuoteProviderUnavailable},
		{http.StatusServiceUnavailable, verticalslice.ErrMarketQuoteProviderUnavailable},
		{http.StatusBadRequest, verticalslice.ErrMarketQuoteProviderData},
		{http.StatusNotFound, verticalslice.ErrMarketQuoteProviderData},
		{http.StatusNoContent, verticalslice.ErrMarketQuoteProviderData},
	} {
		t.Run(fmt.Sprint(tc.status), func(t *testing.T) {
			provider, server := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(tc.status) }), fixedClock{now: deterministicNow})
			defer server.Close()
			_, err := provider.Quote(context.Background(), "SBER")
			if !errors.Is(err, tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, err)
			}
		})
	}
}

func TestQuoteTransportFailureAndContextSemantics(t *testing.T) {
	clock := fixedClock{now: time.Date(2026, 9, 4, 12, 30, 0, 0, time.UTC)}

	provider, err := newQuoteProvider(&http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("network down")
	})}, clock, "https://example.invalid")
	if err != nil {
		t.Fatal(err)
	}
	_, err = provider.Quote(context.Background(), "SBER")
	if !errors.Is(err, verticalslice.ErrMarketQuoteProviderUnavailable) {
		t.Fatalf("network: %v", err)
	}

	provider, err = newQuoteProvider(&http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, context.DeadlineExceeded
	})}, clock, "https://example.invalid")
	if err != nil {
		t.Fatal(err)
	}
	_, err = provider.Quote(context.Background(), "SBER")
	if !errors.Is(err, verticalslice.ErrMarketQuoteProviderUnavailable) {
		t.Fatalf("client timeout: %v", err)
	}

	provider, err = newQuoteProvider(&http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		<-request.Context().Done()
		return nil, request.Context().Err()
	})}, clock, "https://example.invalid")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = provider.Quote(ctx, "SBER")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("caller cancellation: %v", err)
	}

	deadlineCtx, deadlineCancel := context.WithDeadline(context.Background(), deterministicNow.Add(-time.Second))
	defer deadlineCancel()
	_, err = provider.Quote(deadlineCtx, "SBER")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("caller deadline: %v", err)
	}
}

func TestConstructorForcesFiveSecondTimeoutAndDropsCookieJarWithoutMutatingCallerClient(t *testing.T) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Timeout: 17 * time.Second, Jar: jar}
	provider, err := newQuoteProvider(client, fixedClock{now: deterministicNow}, "https://example.com")
	if err != nil {
		t.Fatal(err)
	}
	if provider.client.Timeout != requestTimeout {
		t.Fatalf("provider timeout: %s", provider.client.Timeout)
	}
	if provider.client.Jar != nil {
		t.Fatal("expected provider client cookie jar to be disabled")
	}
	if client.Timeout != 17*time.Second || client.Jar != jar {
		t.Fatalf("caller client mutated: timeout=%s jar_preserved=%v", client.Timeout, client.Jar == jar)
	}
}

func TestConstructorRejectsMissingDependenciesAndInvalidBaseURL(t *testing.T) {
	clock := fixedClock{now: deterministicNow}
	if _, err := newQuoteProvider(nil, clock, "https://example.com"); err == nil {
		t.Fatal("expected nil client error")
	}
	if _, err := newQuoteProvider(&http.Client{}, nil, "https://example.com"); err == nil {
		t.Fatal("expected nil clock error")
	}
	var typedNilClock *nilableClock
	if _, err := newQuoteProvider(&http.Client{}, typedNilClock, "https://example.com"); err == nil {
		t.Fatal("expected typed-nil clock error")
	}
	for _, raw := range []string{"", "://bad", "ftp://example.com", "https://user@example.com", "https://example.com?q=x", "https://example.com#x"} {
		if _, err := newQuoteProvider(&http.Client{}, clock, raw); err == nil {
			t.Fatalf("expected invalid base URL error for %q", raw)
		}
	}
}

func TestQuoteMalformedProviderData(t *testing.T) {
	valid := validPayload("321.12345678")
	cases := []struct{ name, payload string }{
		{"malformed json", `{"marketdata":`},
		{"trailing json", valid + `{}`},
		{"missing marketdata", `{"dataversion":{"columns":["trade_date"],"data":[["2026-09-04"]]}}`},
		{"missing dataversion", `{"marketdata":{"columns":["SECID","BOARDID","LAST","TIME"],"data":[["SBER","TQBR",1,"15:04:05"]]}}`},
		{"missing market columns", `{"marketdata":{"data":[["SBER","TQBR",1,"15:04:05"]]},"dataversion":{"columns":["trade_date"],"data":[["2026-09-04"]]}}`},
		{"missing market data", `{"marketdata":{"columns":["SECID","BOARDID","LAST","TIME"]},"dataversion":{"columns":["trade_date"],"data":[["2026-09-04"]]}}`},
		{"missing required market column", `{"marketdata":{"columns":["SECID","BOARDID","LAST"],"data":[["SBER","TQBR",1]]},"dataversion":{"columns":["trade_date"],"data":[["2026-09-04"]]}}`},
		{"duplicate market column", `{"marketdata":{"columns":["SECID","BOARDID","LAST","TIME","LAST"],"data":[["SBER","TQBR",1,"15:04:05",1]]},"dataversion":{"columns":["trade_date"],"data":[["2026-09-04"]]}}`},
		{"long market row", `{"marketdata":{"columns":["SECID","BOARDID","LAST","TIME"],"data":[["SBER","TQBR",1,"15:04:05","extra"]]},"dataversion":{"columns":["trade_date"],"data":[["2026-09-04"]]}}`},
		{"short market row", `{"marketdata":{"columns":["SECID","BOARDID","LAST","TIME"],"data":[["SBER","TQBR",1]]},"dataversion":{"columns":["trade_date"],"data":[["2026-09-04"]]}}`},
		{"multiple market rows", `{"marketdata":{"columns":["SECID","BOARDID","LAST","TIME"],"data":[["SBER","TQBR",1,"15:04:05"],["SBER","TQBR",2,"15:04:06"]]},"dataversion":{"columns":["trade_date"],"data":[["2026-09-04"]]}}`},
		{"wrong secid", `{"marketdata":{"columns":["SECID","BOARDID","LAST","TIME"],"data":[["GAZP","TQBR",1,"15:04:05"]]},"dataversion":{"columns":["trade_date"],"data":[["2026-09-04"]]}}`},
		{"wrong board", `{"marketdata":{"columns":["SECID","BOARDID","LAST","TIME"],"data":[["SBER","TQTF",1,"15:04:05"]]},"dataversion":{"columns":["trade_date"],"data":[["2026-09-04"]]}}`},
		{"string last", `{"marketdata":{"columns":["SECID","BOARDID","LAST","TIME"],"data":[["SBER","TQBR","1.23","15:04:05"]]},"dataversion":{"columns":["trade_date"],"data":[["2026-09-04"]]}}`},
		{"exponent last", `{"marketdata":{"columns":["SECID","BOARDID","LAST","TIME"],"data":[["SBER","TQBR",1e2,"15:04:05"]]},"dataversion":{"columns":["trade_date"],"data":[["2026-09-04"]]}}`},
		{"negative last", `{"marketdata":{"columns":["SECID","BOARDID","LAST","TIME"],"data":[["SBER","TQBR",-0.01,"15:04:05"]]},"dataversion":{"columns":["trade_date"],"data":[["2026-09-04"]]}}`},
		{"too many decimals", `{"marketdata":{"columns":["SECID","BOARDID","LAST","TIME"],"data":[["SBER","TQBR",1.123456789,"15:04:05"]]},"dataversion":{"columns":["trade_date"],"data":[["2026-09-04"]]}}`},
		{"too many integer digits", `{"marketdata":{"columns":["SECID","BOARDID","LAST","TIME"],"data":[["SBER","TQBR",100000000000000000000,"15:04:05"]]},"dataversion":{"columns":["trade_date"],"data":[["2026-09-04"]]}}`},
		{"invalid time", `{"marketdata":{"columns":["SECID","BOARDID","LAST","TIME"],"data":[["SBER","TQBR",1,"99:99:99"]]},"dataversion":{"columns":["trade_date"],"data":[["2026-09-04"]]}}`},
		{"null time", `{"marketdata":{"columns":["SECID","BOARDID","LAST","TIME"],"data":[["SBER","TQBR",1,null]]},"dataversion":{"columns":["trade_date"],"data":[["2026-09-04"]]}}`},
		{"missing version columns", `{"marketdata":{"columns":["SECID","BOARDID","LAST","TIME"],"data":[["SBER","TQBR",1,"15:04:05"]]},"dataversion":{"data":[["2026-09-04"]]}}`},
		{"missing version data", `{"marketdata":{"columns":["SECID","BOARDID","LAST","TIME"],"data":[["SBER","TQBR",1,"15:04:05"]]},"dataversion":{"columns":["trade_date"]}}`},
		{"empty version data", `{"marketdata":{"columns":["SECID","BOARDID","LAST","TIME"],"data":[["SBER","TQBR",1,"15:04:05"]]},"dataversion":{"columns":["trade_date"],"data":[]}}`},
		{"multiple version rows", `{"marketdata":{"columns":["SECID","BOARDID","LAST","TIME"],"data":[["SBER","TQBR",1,"15:04:05"]]},"dataversion":{"columns":["trade_date"],"data":[["2026-09-04"],["2026-09-05"]]}}`},
		{"duplicate version column", `{"marketdata":{"columns":["SECID","BOARDID","LAST","TIME"],"data":[["SBER","TQBR",1,"15:04:05"]]},"dataversion":{"columns":["trade_date","trade_date"],"data":[["2026-09-04","2026-09-04"]]}}`},
		{"invalid trade date", `{"marketdata":{"columns":["SECID","BOARDID","LAST","TIME"],"data":[["SBER","TQBR",1,"15:04:05"]]},"dataversion":{"columns":["trade_date"],"data":[["2026-99-99"]]}}`},
		{"null trade date", `{"marketdata":{"columns":["SECID","BOARDID","LAST","TIME"],"data":[["SBER","TQBR",1,"15:04:05"]]},"dataversion":{"columns":["trade_date"],"data":[[null]]}}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			provider, server := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, tc.payload) }), fixedClock{now: deterministicNow})
			defer server.Close()
			_, err := provider.Quote(context.Background(), "SBER")
			if !errors.Is(err, verticalslice.ErrMarketQuoteProviderData) {
				t.Fatalf("expected provider data error, got %v", err)
			}
		})
	}
}

func TestQuoteRejectsOversizedResponse(t *testing.T) {
	provider, server := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, strings.Repeat("x", int(maxResponseBodyBytes)+1))
	}), fixedClock{now: deterministicNow})
	defer server.Close()
	_, err := provider.Quote(context.Background(), "SBER")
	if !errors.Is(err, verticalslice.ErrMarketQuoteProviderData) {
		t.Fatalf("expected provider data error, got %v", err)
	}
}

func TestQuoteRejectsEachMissingRequiredColumn(t *testing.T) {
	marketColumns := []string{"SECID", "BOARDID", "LAST", "TIME"}
	for missingIndex, missing := range marketColumns {
		t.Run("marketdata_"+missing, func(t *testing.T) {
			columns := append([]string(nil), marketColumns[:missingIndex]...)
			columns = append(columns, marketColumns[missingIndex+1:]...)
			payload := fmt.Sprintf(`{"marketdata":{"columns":["%s"],"data":[[]]},"dataversion":{"columns":["trade_date"],"data":[["2026-09-04"]]}}`, strings.Join(columns, `","`))
			provider, server := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, payload) }), fixedClock{now: deterministicNow})
			defer server.Close()
			_, err := provider.Quote(context.Background(), "SBER")
			if !errors.Is(err, verticalslice.ErrMarketQuoteProviderData) {
				t.Fatalf("expected provider data error, got %v", err)
			}
		})
	}

	t.Run("dataversion_trade_date", func(t *testing.T) {
		payload := `{"marketdata":{"columns":["SECID","BOARDID","LAST","TIME"],"data":[["SBER","TQBR",1,"15:04:05"]]},"dataversion":{"columns":["other"],"data":[["x"]]}}`
		provider, server := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, payload) }), fixedClock{now: deterministicNow})
		defer server.Close()
		_, err := provider.Quote(context.Background(), "SBER")
		if !errors.Is(err, verticalslice.ErrMarketQuoteProviderData) {
			t.Fatalf("expected provider data error, got %v", err)
		}
	})
}

type failingBody struct{}

func (failingBody) Read([]byte) (int, error) { return 0, errors.New("read failed") }
func (failingBody) Close() error             { return nil }

func TestQuoteClassifiesResponseReadFailureAsUnavailable(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: failingBody{}, Header: make(http.Header)}, nil
	})}
	provider, err := newQuoteProvider(client, fixedClock{now: deterministicNow}, "https://example.com")
	if err != nil {
		t.Fatal(err)
	}
	_, err = provider.Quote(context.Background(), "SBER")
	if !errors.Is(err, verticalslice.ErrMarketQuoteProviderUnavailable) {
		t.Fatalf("expected unavailable, got %v", err)
	}
}

func TestServiceConstructorInjectsProvider(t *testing.T) {
	clock := fixedClock{now: time.Date(2026, 9, 4, 12, 30, 0, 0, time.UTC)}
	provider, server := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, validPayload("1.23")) }), clock)
	defer server.Close()
	service := verticalslice.NewServiceWithQuoteProvider(nil, clock, provider)
	quote, err := service.MarketQuote(context.Background(), "SBER")
	if err != nil {
		t.Fatalf("market quote: %v", err)
	}
	if quote.Provenance.Provider != providerID {
		t.Fatalf("provider: %q", quote.Provenance.Provider)
	}
}

type trackingBody struct {
	io.Reader
	closed atomic.Bool
}

func (body *trackingBody) Close() error { body.closed.Store(true); return nil }

func TestQuoteClosesResponseBody(t *testing.T) {
	body := &trackingBody{Reader: strings.NewReader(validPayload("1.23"))}
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: body, Header: make(http.Header)}, nil
	})}
	provider, err := newQuoteProvider(client, fixedClock{now: deterministicNow}, "https://example.com")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Quote(context.Background(), "SBER"); err != nil {
		t.Fatalf("quote: %v", err)
	}
	if !body.closed.Load() {
		t.Fatal("expected response body to close")
	}
}

func TestServiceBoundaryRejectsZeroRetrievedAt(t *testing.T) {
	provider, server := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, validPayload("1.23")) }), fixedClock{})
	defer server.Close()
	service := verticalslice.NewServiceWithQuoteProvider(nil, fixedClock{}, provider)
	_, err := service.MarketQuote(context.Background(), "SBER")
	if !errors.Is(err, verticalslice.ErrInvalidMarketQuote) {
		t.Fatalf("expected invalid market quote, got %v", err)
	}
}
