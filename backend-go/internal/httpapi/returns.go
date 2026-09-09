package httpapi

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

type portfolioReturnDTO struct {
	PortfolioID            string                        `json:"portfolioId"`
	AsOfDate               string                        `json:"asOfDate"`
	Status                 string                        `json:"status"`
	Reason                 *string                       `json:"reason"`
	XIRR                   *string                       `json:"xirr"`
	ExternalCashFlows      []portfolioReturnCashFlowDTO  `json:"externalCashFlows"`
	TerminalPortfolioValue *moneyDTO                     `json:"terminalPortfolioValue"`
	Calculation            portfolioReturnCalculationDTO `json:"calculation"`
}

type portfolioReturnCashFlowDTO struct {
	Date   string `json:"date"`
	Amount string `json:"amount"`
}

type portfolioReturnCalculationDTO struct {
	MethodologyVersion string `json:"methodologyVersion"`
	DayCountConvention string `json:"dayCountConvention"`
}

func (api *API) getPortfolioReturns(c fiber.Ctx) error {
	subjectID, err := api.subjectID(c)
	if err != nil {
		return writeMappedError(c, err)
	}
	rawAsOfDate, err := requiredSingleQueryValue(c, "asOfDate")
	if err != nil {
		return writeMappedError(c, err)
	}
	projection, err := api.service.GetPortfolioReturns(
		c.Context(),
		subjectID,
		c.Params("portfolioId"),
		rawAsOfDate,
	)
	if err != nil {
		return writeMappedError(c, err)
	}
	return writeOK(c, mapPortfolioReturn(projection))
}

func requiredSingleQueryValue(c fiber.Ctx, name string) (string, error) {
	args := c.Request().URI().QueryArgs()
	count := 0
	value := ""
	args.VisitAll(func(key, rawValue []byte) {
		if string(key) != name {
			return
		}
		count++
		if count == 1 {
			value = string(rawValue)
		}
	})
	if count == 0 || strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("%w: %s is required", verticalslice.ErrInvalidInput, name)
	}
	if count != 1 {
		return "", fmt.Errorf("%w: %s must be supplied exactly once", verticalslice.ErrInvalidInput, name)
	}
	return value, nil
}

func mapPortfolioReturn(projection verticalslice.PortfolioReturnProjection) portfolioReturnDTO {
	cashFlows := make([]portfolioReturnCashFlowDTO, 0, len(projection.ExternalCashFlows))
	for _, flow := range projection.ExternalCashFlows {
		cashFlows = append(cashFlows, portfolioReturnCashFlowDTO{
			Date:   flow.Date,
			Amount: flow.Amount.String(),
		})
	}
	result := portfolioReturnDTO{
		PortfolioID:            projection.PortfolioID,
		AsOfDate:               projection.AsOfDate,
		Status:                 projection.Status,
		ExternalCashFlows:      cashFlows,
		TerminalPortfolioValue: mapOptionalMoney(projection.TerminalPortfolioValue),
		Calculation: portfolioReturnCalculationDTO{
			MethodologyVersion: projection.MethodologyVersion,
			DayCountConvention: projection.DayCountConvention,
		},
	}
	if projection.Reason != "" {
		reason := projection.Reason
		result.Reason = &reason
	}
	if projection.XIRR != nil {
		xirr := projection.XIRR.String()
		result.XIRR = &xirr
	}
	return result
}
