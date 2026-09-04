package verticalslice

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrMarketQuoteNotFound            = errors.New("market quote not found")
	ErrMarketQuoteProviderUnavailable = errors.New("market quote provider unavailable")
	ErrInvalidMarketQuote             = errors.New("invalid market quote")
)

type QuoteProvider interface {
	Quote(ctx context.Context, ticker string) (MarketQuote, error)
}

type MarketDataProvenance struct {
	Provider string
}

type MarketQuote struct {
	Ticker      string
	Price       Money
	AsOf        time.Time
	RetrievedAt time.Time
	Provenance  MarketDataProvenance
}

type FreshnessStatus string

const (
	FreshnessFresh   FreshnessStatus = "FRESH"
	FreshnessStale   FreshnessStatus = "STALE"
	FreshnessUnknown FreshnessStatus = "UNKNOWN"
)

type FreshnessPolicy struct {
	MaxRetrievedAge time.Duration
	MaxMarketAge    time.Duration
}

func ClassifyFreshness(now time.Time, quote MarketQuote, policy FreshnessPolicy) FreshnessStatus {
	if now.IsZero() || policy.MaxRetrievedAge <= 0 || policy.MaxMarketAge <= 0 {
		return FreshnessUnknown
	}
	if quote.AsOf.IsZero() || quote.RetrievedAt.IsZero() {
		return FreshnessUnknown
	}

	now = now.UTC()
	if quote.AsOf.After(now) || quote.RetrievedAt.After(now) {
		return FreshnessUnknown
	}
	if now.Sub(quote.RetrievedAt) > policy.MaxRetrievedAge {
		return FreshnessStale
	}
	if now.Sub(quote.AsOf) > policy.MaxMarketAge {
		return FreshnessStale
	}
	return FreshnessFresh
}

func (s *Service) MarketQuote(ctx context.Context, ticker string) (MarketQuote, error) {
	if !tickerPattern.MatchString(ticker) {
		return MarketQuote{}, fmt.Errorf("%w: ticker is invalid", ErrInvalidInput)
	}
	if s == nil || s.quoteProvider == nil {
		return MarketQuote{}, ErrMarketQuoteProviderUnavailable
	}

	quote, err := s.quoteProvider.Quote(ctx, ticker)
	if err != nil {
		return MarketQuote{}, err
	}
	if err := validateMarketQuote(ticker, quote); err != nil {
		return MarketQuote{}, err
	}
	return quote, nil
}

func validateMarketQuote(requestedTicker string, quote MarketQuote) error {
	if quote.Ticker != requestedTicker {
		return fmt.Errorf("%w: returned ticker does not match requested ticker", ErrInvalidMarketQuote)
	}
	if quote.Price.Currency != RUB {
		return fmt.Errorf("%w: price currency must be RUB", ErrInvalidMarketQuote)
	}
	if !quote.Price.Amount.FitsStorage() || quote.Price.Amount.IsNegative() {
		return fmt.Errorf("%w: price amount must be a non-negative canonical decimal", ErrInvalidMarketQuote)
	}
	if quote.AsOf.IsZero() || quote.RetrievedAt.IsZero() {
		return fmt.Errorf("%w: asOf and retrievedAt are required", ErrInvalidMarketQuote)
	}
	if quote.AsOf.Location() != time.UTC || quote.RetrievedAt.Location() != time.UTC {
		return fmt.Errorf("%w: asOf and retrievedAt must be normalized to UTC", ErrInvalidMarketQuote)
	}
	provider := strings.TrimSpace(quote.Provenance.Provider)
	if provider == "" || provider != quote.Provenance.Provider {
		return fmt.Errorf("%w: provider must be a non-empty canonical identifier", ErrInvalidMarketQuote)
	}
	return nil
}
