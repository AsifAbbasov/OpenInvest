package verticalslice

import (
	"context"
	"fmt"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
)

const DividendCalculatorMethodologyVersion = "dividend-calculator-v1"

type DividendCalculationRequest struct {
	Ticker          string          `json:"ticker"`
	Quantity        decimal.Decimal `json:"quantity"`
	DividendPerUnit Money           `json:"dividendPerUnit"`
	PositionCost    *Money          `json:"positionCost,omitempty"`
}

type DividendCalculation struct {
	Ticker             string
	Quantity           decimal.Decimal
	DividendPerUnit    Money
	GrossDividend      Money
	PositionCost       *Money
	GrossYield         *decimal.Decimal
	TaxIncluded        bool
	MethodologyVersion string
}

type DividendReplayBuilder func(DividendCalculation) (CommandReplayArtifact, error)
type DividendReplayAdmission func() error

type DividendReplayStore interface {
	CalculateDividendWithReplay(
		ctx context.Context,
		command CommandContext,
		calculation DividendCalculation,
		build DividendReplayBuilder,
	) (DividendCalculation, CommandReplayArtifact, error)
}

func CalculateDividend(request DividendCalculationRequest) (DividendCalculation, error) {
	if !tickerPattern.MatchString(request.Ticker) {
		return DividendCalculation{}, fmt.Errorf("%w: ticker must match ^[A-Z0-9]{1,32}$", ErrInvalidInput)
	}
	if !request.Quantity.FitsStorage() {
		return DividendCalculation{}, fmt.Errorf("%w: quantity exceeds NUMERIC(28,8) precision", ErrInvalidInput)
	}
	if !request.Quantity.IsPositive() {
		return DividendCalculation{}, fmt.Errorf("%w: quantity must be greater than zero", ErrInvalidInput)
	}
	if request.DividendPerUnit.Currency != RUB {
		return DividendCalculation{}, fmt.Errorf("%w: dividendPerUnit must be non-negative RUB", ErrInvalidInput)
	}
	if !request.DividendPerUnit.Amount.FitsStorage() {
		return DividendCalculation{}, fmt.Errorf("%w: dividendPerUnit exceeds NUMERIC(28,8) precision", ErrInvalidInput)
	}
	if request.DividendPerUnit.Amount.IsNegative() {
		return DividendCalculation{}, fmt.Errorf("%w: dividendPerUnit must be non-negative RUB", ErrInvalidInput)
	}
	if request.PositionCost != nil {
		if request.PositionCost.Currency != RUB {
			return DividendCalculation{}, fmt.Errorf("%w: positionCost must be positive RUB when supplied", ErrInvalidInput)
		}
		if !request.PositionCost.Amount.FitsStorage() {
			return DividendCalculation{}, fmt.Errorf("%w: positionCost exceeds NUMERIC(28,8) precision", ErrInvalidInput)
		}
		if !request.PositionCost.Amount.IsPositive() {
			return DividendCalculation{}, fmt.Errorf("%w: positionCost must be positive RUB when supplied", ErrInvalidInput)
		}
	}

	gross := request.Quantity.Mul(request.DividendPerUnit.Amount)
	if !gross.FitsStorage() {
		return DividendCalculation{}, fmt.Errorf("%w: derived grossDividend exceeds NUMERIC(28,8) precision", ErrInvalidInput)
	}

	var grossYield *decimal.Decimal
	if request.PositionCost != nil {
		value, err := gross.Div(request.PositionCost.Amount)
		if err != nil {
			return DividendCalculation{}, fmt.Errorf("%w: grossYield cannot be calculated", ErrInvalidInput)
		}
		if value.IsNegative() || !value.FitsStorage() {
			return DividendCalculation{}, fmt.Errorf("%w: derived grossYield exceeds NUMERIC(28,8) precision", ErrInvalidInput)
		}
		grossYield = &value
	}

	return DividendCalculation{
		Ticker:             request.Ticker,
		Quantity:           request.Quantity,
		DividendPerUnit:    request.DividendPerUnit,
		GrossDividend:      Money{Amount: gross, Currency: RUB},
		PositionCost:       request.PositionCost,
		GrossYield:         grossYield,
		TaxIncluded:        false,
		MethodologyVersion: DividendCalculatorMethodologyVersion,
	}, nil
}

func (s *Service) CalculateDividendWithReplay(
	ctx context.Context,
	requestContext RequestContext,
	subjectID string,
	idempotencyKey string,
	requestPath string,
	request DividendCalculationRequest,
	admit DividendReplayAdmission,
	build DividendReplayBuilder,
) (DividendCalculation, CommandReplayArtifact, error) {
	if err := ValidateIdempotencyKey(idempotencyKey); err != nil {
		return DividendCalculation{}, CommandReplayArtifact{}, err
	}
	if admit == nil {
		return DividendCalculation{}, CommandReplayArtifact{}, fmt.Errorf("%w: dividend replay admission is required", ErrReplayUnavailable)
	}
	if build == nil {
		return DividendCalculation{}, CommandReplayArtifact{}, fmt.Errorf("%w: dividend replay builder is required", ErrReplayUnavailable)
	}

	calculation, err := CalculateDividend(request)
	if err != nil {
		return DividendCalculation{}, CommandReplayArtifact{}, err
	}

	canonicalRequest := DividendCalculationRequest{
		Ticker:          calculation.Ticker,
		Quantity:        calculation.Quantity,
		DividendPerUnit: calculation.DividendPerUnit,
		PositionCost:    calculation.PositionCost,
	}
	command, err := s.command(requestContext, subjectID, idempotencyKey, requestPath, canonicalRequest)
	if err != nil {
		return DividendCalculation{}, CommandReplayArtifact{}, err
	}
	lookupStore, ok := s.store.(ReplayLookupStore)
	if !ok {
		return DividendCalculation{}, CommandReplayArtifact{}, ErrReplayUnavailable
	}
	replayStore, ok := s.store.(DividendReplayStore)
	if !ok {
		return DividendCalculation{}, CommandReplayArtifact{}, ErrReplayUnavailable
	}
	artifact, found, err := lookupStore.LookupReplayArtifact(ctx, command, "POST")
	if err != nil {
		return DividendCalculation{}, CommandReplayArtifact{}, err
	}
	if found {
		return DividendCalculation{}, artifact, nil
	}

	if err := admit(); err != nil {
		// Close the race between the initial read-only lookup and admission. A concurrent
		// identical command may have become replayable or in-flight while this request was
		// checking the fresh-command budget. Never replace that stronger idempotency state
		// with a rate-limit response.
		artifact, found, lookupErr := lookupStore.LookupReplayArtifact(ctx, command, "POST")
		if lookupErr != nil {
			return DividendCalculation{}, CommandReplayArtifact{}, lookupErr
		}
		if found {
			return DividendCalculation{}, artifact, nil
		}
		return DividendCalculation{}, CommandReplayArtifact{}, err
	}

	return replayStore.CalculateDividendWithReplay(ctx, command, calculation, build)
}
