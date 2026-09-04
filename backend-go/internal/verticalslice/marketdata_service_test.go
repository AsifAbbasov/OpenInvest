package verticalslice

import (
	"context"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
)

type constructorQuoteProvider struct {
	quote MarketQuote
	calls int
}

func (provider *constructorQuoteProvider) Quote(_ context.Context, _ string) (MarketQuote, error) {
	provider.calls++
	return provider.quote, nil
}

func TestNewServiceWithQuoteProviderInjectsImmutableDependency(t *testing.T) {
	quote := MarketQuote{
		Ticker:      "SBER",
		Price:       Money{Amount: decimal.Must("1.23"), Currency: RUB},
		AsOf:        time.Date(2026, 9, 4, 10, 0, 0, 0, time.UTC),
		RetrievedAt: time.Date(2026, 9, 4, 10, 0, 5, 0, time.UTC),
		Provenance:  MarketDataProvenance{Provider: "TEST_FAKE"},
	}
	provider := &constructorQuoteProvider{quote: quote}
	service := NewServiceWithQuoteProvider(nil, nil, provider)

	got, err := service.MarketQuote(context.Background(), "SBER")
	if err != nil {
		t.Fatalf("market quote: %v", err)
	}
	if provider.calls != 1 {
		t.Fatalf("expected one provider call, got %d", provider.calls)
	}
	if got.Ticker != quote.Ticker || !got.Price.Amount.Equal(quote.Price.Amount) {
		t.Fatalf("unexpected quote: %+v", got)
	}
}
