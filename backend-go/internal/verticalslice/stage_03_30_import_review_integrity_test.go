package verticalslice

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestStage330ImportReviewHistoryFilterIsValidatedAndNormalized(t *testing.T) {
	store := &recordingStore{}
	service := NewService(store, fixedClock{})
	hashA := strings.Repeat("a", 64)
	hashB := strings.Repeat("b", 64)

	_, err := service.ListImportReviewTransactions(context.Background(), "subject-a", "portfolio-a", ImportReviewHistoryFilter{
		SourceAccountLabel:  " Broker A ",
		TradeDates:          []string{"2026-08-23", "2026-08-23"},
		BrokerOperationKeys: []string{hashA, hashA},
		SourceFingerprints:  []string{hashB},
	})
	if err != nil {
		t.Fatalf("list import review transactions: %v", err)
	}
	if store.importHistoryFilter.SourceAccountLabel != "Broker A" ||
		len(store.importHistoryFilter.TradeDates) != 1 ||
		len(store.importHistoryFilter.BrokerOperationKeys) != 1 ||
		len(store.importHistoryFilter.SourceFingerprints) != 1 {
		t.Fatalf("expected normalized bounded filter, got %+v", store.importHistoryFilter)
	}
}

func TestStage330ImportReviewHistoryFilterRejectsInvalidKeys(t *testing.T) {
	service := NewService(&recordingStore{}, fixedClock{})

	_, err := service.ListImportReviewTransactions(context.Background(), "subject-a", "portfolio-a", ImportReviewHistoryFilter{
		TradeDates: []string{"not-a-date"},
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid trade date rejection, got %v", err)
	}

	_, err = service.ListImportReviewTransactions(context.Background(), "subject-a", "portfolio-a", ImportReviewHistoryFilter{
		SourceFingerprints: []string{"ABC"},
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid fingerprint rejection, got %v", err)
	}
}
