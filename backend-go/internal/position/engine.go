package position

import (
	"errors"
	"fmt"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
)

var (
	ErrInsufficientQuantity = errors.New("insufficient position quantity")
	ErrDerivedOverflow      = errors.New("position derived value exceeds NUMERIC(28,8)")
	ErrInvalidTrade         = errors.New("invalid position trade")
)

type Trade struct {
	Type      string
	Quantity  decimal.Decimal
	UnitPrice decimal.Decimal
}

type State struct {
	Quantity            decimal.Decimal
	WeightedAverageCost decimal.Decimal
	AcquisitionBasis    decimal.Decimal
	Open                bool
}

func Empty() State {
	return State{
		Quantity:            decimal.Zero(),
		WeightedAverageCost: decimal.Zero(),
		AcquisitionBasis:    decimal.Zero(),
	}
}

func Rebuild(trades []Trade) (State, error) {
	state := Empty()
	for index, trade := range trades {
		next, err := Apply(state, trade)
		if err != nil {
			return Empty(), fmt.Errorf("trade %d: %w", index+1, err)
		}
		state = next
	}
	return state, nil
}

func Apply(state State, trade Trade) (State, error) {
	if !trade.Quantity.IsPositive() || !trade.Quantity.FitsStorage() ||
		!trade.UnitPrice.IsPositive() || !trade.UnitPrice.FitsStorage() {
		return Empty(), ErrInvalidTrade
	}

	switch trade.Type {
	case "BUY":
		buyBasis := trade.Quantity.Mul(trade.UnitPrice)
		if !buyBasis.FitsStorage() {
			return Empty(), ErrDerivedOverflow
		}
		if !state.Open {
			return checkedState(trade.Quantity, trade.UnitPrice)
		}

		oldBasis := state.Quantity.Mul(state.WeightedAverageCost)
		if !oldBasis.FitsStorage() {
			return Empty(), ErrDerivedOverflow
		}
		numerator := oldBasis.Add(buyBasis)
		newQuantity := state.Quantity.Add(trade.Quantity)
		if !numerator.FitsStorage() || !newQuantity.FitsStorage() {
			return Empty(), ErrDerivedOverflow
		}
		newWAC, err := numerator.Div(newQuantity)
		if err != nil || !newWAC.FitsStorage() {
			return Empty(), ErrDerivedOverflow
		}
		return checkedState(newQuantity, newWAC)

	case "SELL":
		if !state.Open {
			return Empty(), ErrInsufficientQuantity
		}
		newQuantity := state.Quantity.Sub(trade.Quantity)
		if newQuantity.IsNegative() {
			return Empty(), ErrInsufficientQuantity
		}
		if newQuantity.IsZero() {
			return Empty(), nil
		}
		if !newQuantity.FitsStorage() {
			return Empty(), ErrDerivedOverflow
		}
		return checkedState(newQuantity, state.WeightedAverageCost)

	default:
		return Empty(), ErrInvalidTrade
	}
}

func checkedState(quantity decimal.Decimal, weightedAverageCost decimal.Decimal) (State, error) {
	basis := quantity.Mul(weightedAverageCost)
	if !quantity.FitsStorage() || !weightedAverageCost.FitsStorage() || !basis.FitsStorage() {
		return Empty(), ErrDerivedOverflow
	}
	return State{
		Quantity:            quantity,
		WeightedAverageCost: weightedAverageCost,
		AcquisitionBasis:    basis,
		Open:                true,
	}, nil
}
