package moexiss

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"time"
	_ "time/tzdata"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

const (
	productionBaseURL          = "https://iss.moex.com"
	providerID                 = "MOEX_ISS"
	boardID                    = "TQBR"
	requestTimeout             = 5 * time.Second
	maxResponseBodyBytes int64 = 64 * 1024
)

type Provider struct {
	client  *http.Client
	clock   verticalslice.Clock
	baseURL *url.URL
	moscow  *time.Location
}

var _ verticalslice.QuoteProvider = (*Provider)(nil)

func NewQuoteProvider(client *http.Client, clock verticalslice.Clock) (*Provider, error) {
	return newQuoteProvider(client, clock, productionBaseURL)
}

func newQuoteProvider(client *http.Client, clock verticalslice.Clock, baseURL string) (*Provider, error) {
	if client == nil {
		return nil, errors.New("moexiss: http client is required")
	}
	if isNilDependency(clock) {
		return nil, errors.New("moexiss: clock is required")
	}

	parsedBaseURL, err := url.Parse(baseURL)
	if err != nil || parsedBaseURL.Scheme == "" || parsedBaseURL.Host == "" || parsedBaseURL.User != nil || parsedBaseURL.RawQuery != "" || parsedBaseURL.Fragment != "" {
		return nil, errors.New("moexiss: base URL is invalid")
	}
	if parsedBaseURL.Scheme != "http" && parsedBaseURL.Scheme != "https" {
		return nil, errors.New("moexiss: base URL scheme is invalid")
	}

	moscow, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return nil, fmt.Errorf("moexiss: load Europe/Moscow: %w", err)
	}

	clientCopy := *client
	clientCopy.Timeout = requestTimeout
	clientCopy.Jar = nil

	return &Provider{
		client:  &clientCopy,
		clock:   clock,
		baseURL: parsedBaseURL,
		moscow:  moscow,
	}, nil
}

func (provider *Provider) Quote(ctx context.Context, ticker string) (verticalslice.MarketQuote, error) {
	requestURL := provider.baseURL.JoinPath("iss", "engines", "stock", "markets", "shares", "boards", boardID, "securities", ticker+".json")
	query := requestURL.Query()
	query.Set("iss.meta", "off")
	query.Set("iss.only", "marketdata,dataversion")
	query.Set("marketdata.columns", "SECID,BOARDID,LAST,TIME")
	query.Set("dataversion.columns", "trade_date")
	requestURL.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return verticalslice.MarketQuote{}, providerDataError("build request")
	}
	request.Header.Set("Accept", "application/json")

	response, err := provider.client.Do(request)
	if err != nil {
		if contextErr := ctx.Err(); contextErr != nil {
			return verticalslice.MarketQuote{}, contextErr
		}
		return verticalslice.MarketQuote{}, fmt.Errorf("%w: request failed", verticalslice.ErrMarketQuoteProviderUnavailable)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusTooManyRequests || response.StatusCode >= http.StatusInternalServerError {
		return verticalslice.MarketQuote{}, fmt.Errorf("%w: provider returned HTTP %d", verticalslice.ErrMarketQuoteProviderUnavailable, response.StatusCode)
	}
	if response.StatusCode != http.StatusOK {
		return verticalslice.MarketQuote{}, providerDataError(fmt.Sprintf("unexpected HTTP %d", response.StatusCode))
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBodyBytes+1))
	if err != nil {
		if contextErr := ctx.Err(); contextErr != nil {
			return verticalslice.MarketQuote{}, contextErr
		}
		return verticalslice.MarketQuote{}, fmt.Errorf("%w: response read failed", verticalslice.ErrMarketQuoteProviderUnavailable)
	}
	if int64(len(body)) > maxResponseBodyBytes {
		return verticalslice.MarketQuote{}, providerDataError("response body exceeds 64 KiB")
	}

	parsed, err := parseResponse(body, ticker, provider.moscow)
	if err != nil {
		return verticalslice.MarketQuote{}, err
	}

	return verticalslice.MarketQuote{
		Ticker: ticker,
		Price: verticalslice.Money{
			Amount:   parsed.price,
			Currency: verticalslice.RUB,
		},
		AsOf:        parsed.asOf,
		RetrievedAt: provider.clock.Now().UTC(),
		Provenance:  verticalslice.MarketDataProvenance{Provider: providerID},
	}, nil
}

func isNilDependency(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
