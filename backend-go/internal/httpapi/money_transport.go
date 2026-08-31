package httpapi

import (
	"fmt"
	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

type moneyDTO struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

func parseOptionalDecimal(value *string) (*decimal.Decimal, error) {
	if value == nil {
		return nil, nil
	}
	parsed, err := decimal.FromString(*value)
	if err != nil {
		return nil, fmt.Errorf("%w: decimal value is invalid", verticalslice.ErrInvalidInput)
	}
	return &parsed, nil
}

func parseOptionalMoney(value *moneyDTO) (*verticalslice.Money, error) {
	if value == nil {
		return nil, nil
	}
	parsed, err := parseMoney(*value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseMoney(value moneyDTO) (verticalslice.Money, error) {
	amount, err := decimal.FromString(value.Amount)
	if err != nil {
		return verticalslice.Money{}, fmt.Errorf("%w: money amount is invalid", verticalslice.ErrInvalidInput)
	}
	return verticalslice.Money{Amount: amount, Currency: value.Currency}, nil
}

func mapOptionalMoney(item *verticalslice.Money) *moneyDTO {
	if item == nil {
		return nil
	}
	value := mapMoney(*item)
	return &value
}

func mapMoney(item verticalslice.Money) moneyDTO {
	return moneyDTO{Amount: item.Amount.String(), Currency: item.Currency}
}
