package verticalslice

import (
	"context"
	"errors"
	"testing"
)

type oiNew08SummaryServiceStore struct {
	recordingStore
	summary   PortfolioSummary
	forwarded string
	calls     int
}

func (store *oiNew08SummaryServiceStore) GetPortfolioSummary(
	_ context.Context,
	_ string,
	_ string,
	asOfDate string,
) (PortfolioSummary, error) {
	store.calls++
	store.forwarded = asOfDate
	return store.summary, nil
}

func TestOINew08SummaryServicePreservesExplicitBusinessDateAndSelectedDate(t *testing.T) {
	store := &oiNew08SummaryServiceStore{summary: PortfolioSummary{AsOfDate: "2026-09-01"}}
	service := NewService(store, SystemClock{})

	summary, err := service.GetPortfolioSummary(
		context.Background(),
		"subject",
		"portfolio",
		"2026-09-07",
	)
	if err != nil {
		t.Fatalf("explicit summary date: %v", err)
	}
	if store.calls != 1 || store.forwarded != "2026-09-07" {
		t.Fatalf("service must forward explicit valid BusinessDate unchanged: calls=%d forwarded=%q", store.calls, store.forwarded)
	}
	if summary.AsOfDate != "2026-09-01" {
		t.Fatalf("selected snapshot date must remain store result, got %q", summary.AsOfDate)
	}
}

func TestOINew08SummaryServicePreservesOmittedMode(t *testing.T) {
	store := &oiNew08SummaryServiceStore{summary: PortfolioSummary{AsOfDate: "2099-01-01"}}
	service := NewService(store, SystemClock{})

	summary, err := service.GetPortfolioSummary(context.Background(), "subject", "portfolio", "")
	if err != nil {
		t.Fatalf("omitted summary date: %v", err)
	}
	if store.calls != 1 || store.forwarded != "" {
		t.Fatalf("omitted summary date must remain empty: calls=%d forwarded=%q", store.calls, store.forwarded)
	}
	if summary.AsOfDate != "2099-01-01" {
		t.Fatalf("future selected snapshot must not be clock-clamped, got %q", summary.AsOfDate)
	}
}

func TestOINew08SummaryServiceRejectsInvalidBusinessDateBeforeStore(t *testing.T) {
	store := &oiNew08SummaryServiceStore{}
	service := NewService(store, SystemClock{})

	_, err := service.GetPortfolioSummary(context.Background(), "subject", "portfolio", "not-a-date")
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("invalid summary date: got %v want ErrInvalidInput", err)
	}
	if store.calls != 0 {
		t.Fatalf("invalid BusinessDate must fail before store, calls=%d", store.calls)
	}
}
