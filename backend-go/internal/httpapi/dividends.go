package httpapi

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

const dividendCalculatorReplayScopeDomain = "openinvest:dividend-calculator:anonymous-replay:v1:"

var errDividendCalculatorRateLimited = errors.New("dividend calculator rate limited")

type dividendCalculationRequestDTO struct {
	Ticker          string    `json:"ticker"`
	Quantity        string    `json:"quantity"`
	DividendPerUnit moneyDTO  `json:"dividendPerUnit"`
	PositionCost    *moneyDTO `json:"positionCost,omitempty"`
}

type dividendCalculationDTO struct {
	Ticker             string    `json:"ticker"`
	Quantity           string    `json:"quantity"`
	DividendPerUnit    moneyDTO  `json:"dividendPerUnit"`
	GrossDividend      moneyDTO  `json:"grossDividend"`
	PositionCost       *moneyDTO `json:"positionCost"`
	GrossYield         *string   `json:"grossYield"`
	TaxIncluded        bool      `json:"taxIncluded"`
	MethodologyVersion string    `json:"methodologyVersion"`
}

func newDividendCalculatorRateLimiter() *authRateLimiter {
	return newBoundedAuthRateLimiter(
		defaultDividendCalculatorPerKeyLimit,
		defaultDividendCalculatorGlobalLimit,
		defaultDividendCalculatorMaxKeys,
		defaultDividendCalculatorWindow,
	)
}

func (api *API) calculateDividend(c fiber.Ctx) error {
	meta := requestMeta(c)
	idempotencyKey := c.Get("Idempotency-Key")
	if err := verticalslice.ValidateIdempotencyKey(idempotencyKey); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	var request dividendCalculationRequestDTO
	if err := decodeStrictJSON(c.Request().Body(), &request); err != nil {
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
	}
	appRequest, err := request.toApp()
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}

	_, artifact, err := api.service.CalculateDividendWithReplay(
		c.Context(),
		meta.toApp(),
		dividendReplaySubjectID(idempotencyKey),
		idempotencyKey,
		c.Path(),
		appRequest,
		func() error {
			if api.dividendLimiter != nil && !api.dividendLimiter.allow(idempotencyKey, api.nowUTC()) {
				return errDividendCalculatorRateLimited
			}
			return nil
		},
		func(calculation verticalslice.DividendCalculation) (verticalslice.CommandReplayArtifact, error) {
			return buildCommandReplayArtifact(meta, http.StatusOK, mapDividendCalculation(calculation))
		},
	)
	if err != nil {
		if errors.Is(err, errDividendCalculatorRateLimited) {
			c.Set("Retry-After", dividendCalculatorRateLimitRetryAfterSeconds)
			return writeErrorWithMeta(c, meta, http.StatusTooManyRequests, "RATE_LIMITED", "Too many dividend calculator requests")
		}
		return writeReplayAwareError(c, meta, err)
	}
	return writeCommandReplayArtifact(c, artifact)
}

func (request dividendCalculationRequestDTO) toApp() (verticalslice.DividendCalculationRequest, error) {
	quantity, err := decimal.FromString(request.Quantity)
	if err != nil {
		return verticalslice.DividendCalculationRequest{}, fmt.Errorf("%w: quantity is invalid", verticalslice.ErrInvalidInput)
	}
	dividendPerUnit, err := parseMoney(request.DividendPerUnit)
	if err != nil {
		return verticalslice.DividendCalculationRequest{}, err
	}
	positionCost, err := parseOptionalMoney(request.PositionCost)
	if err != nil {
		return verticalslice.DividendCalculationRequest{}, err
	}
	return verticalslice.DividendCalculationRequest{
		Ticker:          request.Ticker,
		Quantity:        quantity,
		DividendPerUnit: dividendPerUnit,
		PositionCost:    positionCost,
	}, nil
}

func mapDividendCalculation(calculation verticalslice.DividendCalculation) dividendCalculationDTO {
	var grossYield *string
	if calculation.GrossYield != nil {
		value := calculation.GrossYield.String()
		grossYield = &value
	}
	return dividendCalculationDTO{
		Ticker:             calculation.Ticker,
		Quantity:           calculation.Quantity.String(),
		DividendPerUnit:    mapMoney(calculation.DividendPerUnit),
		GrossDividend:      mapMoney(calculation.GrossDividend),
		PositionCost:       mapOptionalMoney(calculation.PositionCost),
		GrossYield:         grossYield,
		TaxIncluded:        calculation.TaxIncluded,
		MethodologyVersion: calculation.MethodologyVersion,
	}
}

func dividendReplaySubjectID(idempotencyKey string) string {
	sum := sha256.Sum256([]byte(dividendCalculatorReplayScopeDomain + strings.TrimSpace(idempotencyKey)))
	var value uuid.UUID
	copy(value[:], sum[:16])
	// RFC 9562 UUIDv8 marks this as an application-defined technical identifier.
	value[6] = (value[6] & 0x0f) | 0x80
	value[8] = (value[8] & 0x3f) | 0x80
	return value.String()
}
