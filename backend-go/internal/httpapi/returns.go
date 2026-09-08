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
	rawAsOfDate, present := queryValue(c, "asOfDate")
	if !present || strings.TrimSpace(rawAsOfDate) == "" {
		return writeMappedError(c, fmt.Errorf("%w: asOfDate is required", verticalslice.ErrInvalidInput))
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
