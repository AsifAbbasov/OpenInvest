package verticalslice

import (
	"context"
	"errors"
	"testing"
)

type typedNilQuoteProvider struct{}

func (*typedNilQuoteProvider) Quote(context.Context, string) (MarketQuote, error) {
	panic("typed-nil provider must not be called")
}

func TestNewServiceWithQuoteProviderNormalizesTypedNilProvider(t *testing.T) {
	var provider *typedNilQuoteProvider
	service := NewServiceWithQuoteProvider(nil, nil, provider)

	_, err := service.MarketQuote(context.Background(), "SBER")
	if !errors.Is(err, ErrMarketQuoteProviderUnavailable) {
		t.Fatalf("expected provider unavailable, got %v", err)
	}
}
