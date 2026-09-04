package verticalslice

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
)

type marketDataFixedClock struct {
	now time.Time
}

func (clock marketDataFixedClock) Now() time.Time {
	return clock.now
}

type fakeQuoteProvider struct {
	quotes map[string]MarketQuote
	err    error
	calls  int
}

func (provider *fakeQuoteProvider) Quote(_ context.Context, ticker string) (MarketQuote, error) {
	provider.calls++
	if provider.err != nil {
		return MarketQuote{}, provider.err
	}
	quote, ok := provider.quotes[ticker]
	if !ok {
		return MarketQuote{}, ErrMarketQuoteNotFound
	}
	return quote, nil
}

func validMarketQuote(ticker string) MarketQuote {
	return MarketQuote{
		Ticker: ticker,
		Price: Money{
			Amount:   decimal.Must("321.12345678"),
			Currency: RUB,
		},
		AsOf:        time.Date(2026, 9, 4, 10, 0, 0, 0, time.UTC),
		RetrievedAt: time.Date(2026, 9, 4, 10, 0, 5, 0, time.UTC),
		Provenance: MarketDataProvenance{
			Provider: "TEST_FAKE",
		},
	}
}

func serviceWithQuoteProvider(provider QuoteProvider, clock Clock) *Service {
	service := NewService(nil, clock)
	service.quoteProvider = provider
	return service
}

func TestMarketQuotePreservesCanonicalMoneyAndTimestamps(t *testing.T) {
	quote := validMarketQuote("SBER")
	provider := &fakeQuoteProvider{quotes: map[string]MarketQuote{"SBER": quote}}
	service := serviceWithQuoteProvider(provider, marketDataFixedClock{now: quote.RetrievedAt})

	got, err := service.MarketQuote(context.Background(), "SBER")
	if err != nil {
		t.Fatalf("market quote: %v", err)
	}
	if got.Price.Amount.String() != "321.12345678" {
		t.Fatalf("expected exact decimal amount, got %s", got.Price.Amount.String())
	}
	if got.Price.Currency != RUB {
		t.Fatalf("expected RUB, got %s", got.Price.Currency)
	}
	if got.AsOf.Equal(got.RetrievedAt) {
		t.Fatal("expected asOf and retrievedAt to remain semantically distinct")
	}
	if got.AsOf.Location() != time.UTC || got.RetrievedAt.Location() != time.UTC {
		t.Fatalf("expected UTC timestamps, got asOf=%v retrievedAt=%v", got.AsOf.Location(), got.RetrievedAt.Location())
	}
	if got.Provenance.Provider != "TEST_FAKE" {
		t.Fatalf("expected provider identity to be preserved, got %q", got.Provenance.Provider)
	}
}

func TestMarketQuoteAllowsAsOfAfterRetrievedAt(t *testing.T) {
	quote := validMarketQuote("SBER")
	quote.RetrievedAt = time.Date(2026, 9, 4, 10, 0, 5, 0, time.UTC)
	quote.AsOf = time.Date(2026, 9, 4, 10, 0, 6, 0, time.UTC)
	provider := &fakeQuoteProvider{quotes: map[string]MarketQuote{"SBER": quote}}
	service := serviceWithQuoteProvider(provider, marketDataFixedClock{now: time.Date(2026, 9, 4, 10, 1, 0, 0, time.UTC)})

	got, err := service.MarketQuote(context.Background(), "SBER")
	if err != nil {
		t.Fatalf("market quote: %v", err)
	}
	if !got.AsOf.After(got.RetrievedAt) {
		t.Fatalf("expected asOf after retrievedAt to remain valid, got asOf=%s retrievedAt=%s", got.AsOf, got.RetrievedAt)
	}
}

func TestMarketQuoteRejectsNonCanonicalTickerWithoutCallingProvider(t *testing.T) {
	for _, ticker := range []string{"sber", " SBER", "SBER ", "SBER!"} {
		t.Run(ticker, func(t *testing.T) {
			provider := &fakeQuoteProvider{quotes: map[string]MarketQuote{}}
			service := serviceWithQuoteProvider(provider, marketDataFixedClock{now: time.Date(2026, 9, 4, 10, 1, 0, 0, time.UTC)})

			_, err := service.MarketQuote(context.Background(), ticker)
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("expected invalid input, got %v", err)
			}
			if provider.calls != 0 {
				t.Fatalf("expected provider not to be called, got %d calls", provider.calls)
			}
		})
	}
}

func TestMarketQuoteFailsClosedWithoutProvider(t *testing.T) {
	service := NewService(nil, marketDataFixedClock{now: time.Date(2026, 9, 4, 10, 1, 0, 0, time.UTC)})

	_, err := service.MarketQuote(context.Background(), "SBER")
	if !errors.Is(err, ErrMarketQuoteProviderUnavailable) {
		t.Fatalf("expected provider unavailable, got %v", err)
	}
}

func TestMarketQuotePreservesNotFound(t *testing.T) {
	provider := &fakeQuoteProvider{quotes: map[string]MarketQuote{}}
	service := serviceWithQuoteProvider(provider, marketDataFixedClock{now: time.Date(2026, 9, 4, 10, 1, 0, 0, time.UTC)})

	_, err := service.MarketQuote(context.Background(), "GAZP")
	if !errors.Is(err, ErrMarketQuoteNotFound) {
		t.Fatalf("expected quote not found, got %v", err)
	}
}

func TestMarketQuotePropagatesProviderError(t *testing.T) {
	providerErr := errors.New("provider failed")
	provider := &fakeQuoteProvider{err: providerErr}
	service := serviceWithQuoteProvider(provider, marketDataFixedClock{now: time.Date(2026, 9, 4, 10, 1, 0, 0, time.UTC)})

	_, err := service.MarketQuote(context.Background(), "SBER")
	if !errors.Is(err, providerErr) {
		t.Fatalf("expected provider error, got %v", err)
	}
}

func TestMarketQuoteRejectsMalformedProviderOutput(t *testing.T) {
	nonUTC := time.FixedZone("UTC+3", 3*60*60)
	tests := []struct {
		name   string
		mutate func(*MarketQuote)
	}{
		{
			name: "ticker mismatch",
			mutate: func(quote *MarketQuote) {
				quote.Ticker = "GAZP"
			},
		},
		{
			name: "negative price",
			mutate: func(quote *MarketQuote) {
				quote.Price.Amount = decimal.Must("-0.00000001")
			},
		},
		{
			name: "zero value decimal",
			mutate: func(quote *MarketQuote) {
				quote.Price.Amount = decimal.Decimal{}
			},
		},
		{
			name: "non rub currency",
			mutate: func(quote *MarketQuote) {
				quote.Price.Currency = "USD"
			},
		},
		{
			name: "zero as of",
			mutate: func(quote *MarketQuote) {
				quote.AsOf = time.Time{}
			},
		},
		{
			name: "zero retrieved at",
			mutate: func(quote *MarketQuote) {
				quote.RetrievedAt = time.Time{}
			},
		},
		{
			name: "as of not normalized to utc",
			mutate: func(quote *MarketQuote) {
				quote.AsOf = quote.AsOf.In(nonUTC)
			},
		},
		{
			name: "retrieved at not normalized to utc",
			mutate: func(quote *MarketQuote) {
				quote.RetrievedAt = quote.RetrievedAt.In(nonUTC)
			},
		},
		{
			name: "missing provider",
			mutate: func(quote *MarketQuote) {
				quote.Provenance.Provider = "   "
			},
		},
		{
			name: "provider identifier has surrounding whitespace",
			mutate: func(quote *MarketQuote) {
				quote.Provenance.Provider = " TEST_FAKE "
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			quote := validMarketQuote("SBER")
			test.mutate(&quote)
			provider := &fakeQuoteProvider{quotes: map[string]MarketQuote{"SBER": quote}}
			service := serviceWithQuoteProvider(provider, marketDataFixedClock{now: time.Date(2026, 9, 4, 10, 1, 0, 0, time.UTC)})

			_, err := service.MarketQuote(context.Background(), "SBER")
			if !errors.Is(err, ErrInvalidMarketQuote) {
				t.Fatalf("expected invalid market quote, got %v", err)
			}
		})
	}
}

func TestClassifyFreshnessDeterministicPolicy(t *testing.T) {
	clock := marketDataFixedClock{now: time.Date(2026, 9, 4, 12, 0, 0, 0, time.UTC)}
	policy := FreshnessPolicy{
		MaxRetrievedAge: 2 * time.Minute,
		MaxMarketAge:    5 * time.Minute,
	}
	base := validMarketQuote("SBER")
	base.RetrievedAt = clock.Now().Add(-time.Minute)
	base.AsOf = clock.Now().Add(-4 * time.Minute)

	tests := []struct {
		name   string
		now    time.Time
		policy FreshnessPolicy
		mutate func(*MarketQuote)
		want   FreshnessStatus
	}{
		{name: "fresh", now: clock.Now(), policy: policy, mutate: func(*MarketQuote) {}, want: FreshnessFresh},
		{name: "fresh with non utc now", now: clock.Now().In(time.FixedZone("UTC+4", 4*60*60)), policy: policy, mutate: func(*MarketQuote) {}, want: FreshnessFresh},
		{name: "stale retrieval age", now: clock.Now(), policy: policy, mutate: func(quote *MarketQuote) { quote.RetrievedAt = clock.Now().Add(-2*time.Minute - time.Nanosecond) }, want: FreshnessStale},
		{name: "stale market age", now: clock.Now(), policy: policy, mutate: func(quote *MarketQuote) { quote.AsOf = clock.Now().Add(-5*time.Minute - time.Nanosecond) }, want: FreshnessStale},
		{name: "zero now", now: time.Time{}, policy: policy, mutate: func(*MarketQuote) {}, want: FreshnessUnknown},
		{name: "zero retrieved threshold", now: clock.Now(), policy: FreshnessPolicy{MaxRetrievedAge: 0, MaxMarketAge: 5 * time.Minute}, mutate: func(*MarketQuote) {}, want: FreshnessUnknown},
		{name: "zero market threshold", now: clock.Now(), policy: FreshnessPolicy{MaxRetrievedAge: 2 * time.Minute, MaxMarketAge: 0}, mutate: func(*MarketQuote) {}, want: FreshnessUnknown},
		{name: "negative retrieval threshold", now: clock.Now(), policy: FreshnessPolicy{MaxRetrievedAge: -time.Second, MaxMarketAge: 5 * time.Minute}, mutate: func(*MarketQuote) {}, want: FreshnessUnknown},
		{name: "missing as of", now: clock.Now(), policy: policy, mutate: func(quote *MarketQuote) { quote.AsOf = time.Time{} }, want: FreshnessUnknown},
		{name: "missing retrieved at", now: clock.Now(), policy: policy, mutate: func(quote *MarketQuote) { quote.RetrievedAt = time.Time{} }, want: FreshnessUnknown},
		{name: "retrieval exactly threshold", now: clock.Now(), policy: policy, mutate: func(quote *MarketQuote) { quote.RetrievedAt = clock.Now().Add(-2 * time.Minute) }, want: FreshnessFresh},
		{name: "market exactly threshold", now: clock.Now(), policy: policy, mutate: func(quote *MarketQuote) { quote.AsOf = clock.Now().Add(-5 * time.Minute) }, want: FreshnessFresh},
		{name: "future as of", now: clock.Now(), policy: policy, mutate: func(quote *MarketQuote) { quote.AsOf = clock.Now().Add(time.Nanosecond) }, want: FreshnessUnknown},
		{name: "future retrieved at", now: clock.Now(), policy: policy, mutate: func(quote *MarketQuote) { quote.RetrievedAt = clock.Now().Add(time.Nanosecond) }, want: FreshnessUnknown},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			quote := base
			test.mutate(&quote)
			if got := ClassifyFreshness(test.now, quote, test.policy); got != test.want {
				t.Fatalf("expected %s, got %s", test.want, got)
			}
		})
	}
}
