package verticalslice

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
)

func TestAppendImportedTransactionsRejectsMixedIdentityStrengthInBothOrders(t *testing.T) {
	gross := Money{Amount: decimal.Must("1000.00000000"), Currency: RUB}
	base := AppendTransactionRequest{
		PortfolioID:     "portfolio-id",
		TransactionType: "DEPOSIT",
		GrossAmount:     &gross,
		Commission:      ZeroMoney(),
		Tax:             ZeroMoney(),
		TradeDate:       "2026-06-26",
	}
	fingerprint, err := NormalizedTransactionFingerprint(base)
	if err != nil {
		t.Fatalf("fingerprint transaction: %v", err)
	}
	fallback := base
	fallback.ImportProvenance = &ImportProvenance{
		IdentityVersion:   ImportIdentityVersion,
		SourceFingerprint: fingerprint,
	}
	strong := base
	strong.ImportProvenance = &ImportProvenance{
		IdentityVersion:    ImportIdentityVersion,
		BrokerOperationKey: BrokerOperationKey("broker-operation-123"),
		SourceFingerprint:  fingerprint,
	}

	tests := []struct {
		name         string
		transactions []AppendTransactionRequest
	}{
		{name: "fallback then strong", transactions: []AppendTransactionRequest{fallback, strong}},
		{name: "strong then fallback", transactions: []AppendTransactionRequest{strong, fallback}},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			store := &recordingStore{}
			service := NewService(store, fixedClock{})
			_, err := service.AppendImportedTransactions(context.Background(), RequestContext{}, "subject", "mixed-strength-key-01", "/internal/imports/append", AppendImportBatchRequest{
				PortfolioID:        "portfolio-id",
				Transactions:       testCase.transactions,
				SourceKind:         "USER_UPLOADED_FILE",
				SourceAccountLabel: "Broker account A",
				SourceFileHash:     strings.Repeat("a", 64),
			})
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("expected mixed identity strength to fail closed, got %v", err)
			}
			if len(store.importRequest.Transactions) != 0 {
				t.Fatalf("expected validation to stop before persistence, got %+v", store.importRequest)
			}
		})
	}
}

func TestAppendImportedTransactionsAllowsDistinctStrongIdentitiesWithSameFingerprint(t *testing.T) {
	store := &recordingStore{}
	service := NewService(store, fixedClock{})
	gross := Money{Amount: decimal.Must("1000.00000000"), Currency: RUB}
	base := AppendTransactionRequest{
		PortfolioID:     "portfolio-id",
		TransactionType: "DEPOSIT",
		GrossAmount:     &gross,
		Commission:      ZeroMoney(),
		Tax:             ZeroMoney(),
		TradeDate:       "2026-06-26",
	}
	fingerprint, err := NormalizedTransactionFingerprint(base)
	if err != nil {
		t.Fatalf("fingerprint transaction: %v", err)
	}
	first := base
	first.ImportProvenance = &ImportProvenance{
		IdentityVersion:    ImportIdentityVersion,
		BrokerOperationKey: BrokerOperationKey("broker-operation-a"),
		SourceFingerprint:  fingerprint,
	}
	second := base
	second.ImportProvenance = &ImportProvenance{
		IdentityVersion:    ImportIdentityVersion,
		BrokerOperationKey: BrokerOperationKey("broker-operation-b"),
		SourceFingerprint:  fingerprint,
	}

	_, err = service.AppendImportedTransactions(context.Background(), RequestContext{}, "subject", "strong-pair-key-01", "/internal/imports/append", AppendImportBatchRequest{
		PortfolioID:        "portfolio-id",
		Transactions:       []AppendTransactionRequest{first, second},
		SourceKind:         "USER_UPLOADED_FILE",
		SourceAccountLabel: "Broker account A",
		SourceFileHash:     strings.Repeat("b", 64),
	})
	if err != nil {
		t.Fatalf("expected distinct strong identities to remain valid, got %v", err)
	}
	if len(store.importRequest.Transactions) != 2 {
		t.Fatalf("expected both strong identities to reach persistence, got %+v", store.importRequest)
	}
}
