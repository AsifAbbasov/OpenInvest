package verticalslice

import (
	"errors"
	"strings"
	"testing"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
)

func stage329Deposit(note *string) AppendTransactionRequest {
	gross := Money{Amount: decimal.Must("100.00000000"), Currency: RUB}
	return AppendTransactionRequest{
		PortfolioID:     "00000000-0000-4000-8000-000000000002",
		TransactionType: "DEPOSIT",
		GrossAmount:     &gross,
		Commission:      ZeroMoney(),
		Tax:             ZeroMoney(),
		TradeDate:       "2026-01-10",
		Note:            note,
	}
}

func TestStage329ValidateAppendTransactionEnforcesNoteCharacterLimit(t *testing.T) {
	accepted := strings.Repeat("Ж", 500)
	if err := validateAppendTransaction(stage329Deposit(&accepted)); err != nil {
		t.Fatalf("expected 500-character note to be accepted: %v", err)
	}

	rejected := strings.Repeat("Ж", 501)
	if err := validateAppendTransaction(stage329Deposit(&rejected)); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected 501-character note to be invalid input, got %v", err)
	}
}

func TestStage329ValidateAppendTransactionRejectsDerivedGrossOverflow(t *testing.T) {
	ticker := "SBER"
	quantity := decimal.Must("99999999999999999999.00000000")
	unitPrice := Money{Amount: decimal.Must("2.00000000"), Currency: RUB}
	request := AppendTransactionRequest{
		PortfolioID:     "00000000-0000-4000-8000-000000000002",
		TransactionType: "BUY",
		Ticker:          &ticker,
		Quantity:        &quantity,
		UnitPrice:       &unitPrice,
		Commission:      ZeroMoney(),
		Tax:             ZeroMoney(),
		TradeDate:       "2026-01-10",
	}

	if err := validateAppendTransaction(request); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected derived gross overflow to be invalid input, got %v", err)
	}
}
