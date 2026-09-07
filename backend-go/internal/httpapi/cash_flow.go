package httpapi

import (
	"github.com/gofiber/fiber/v3"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

type portfolioCashFlowDTO struct {
	PortfolioID string                       `json:"portfolioId"`
	FromDate    *string                      `json:"fromDate"`
	ToDate      *string                      `json:"toDate"`
	Totals      portfolioCashFlowTotalsDTO    `json:"totals"`
	Periods     []portfolioCashFlowPeriodDTO  `json:"periods"`
	Calculation portfolioCashFlowCalculationDTO `json:"calculation"`
}

type portfolioCashFlowTotalsDTO struct {
	Deposits            moneyDTO `json:"deposits"`
	Withdrawals         moneyDTO `json:"withdrawals"`
	BuyOutflows         moneyDTO `json:"buyOutflows"`
	SellInflows         moneyDTO `json:"sellInflows"`
	DividendsGross      moneyDTO `json:"dividendsGross"`
	CouponsGross        moneyDTO `json:"couponsGross"`
	Fees                moneyDTO `json:"fees"`
	Taxes               moneyDTO `json:"taxes"`
	NetExternalFlow     moneyDTO `json:"netExternalFlow"`
	NetInvestmentIncome moneyDTO `json:"netInvestmentIncome"`
	NetCashFlow         moneyDTO `json:"netCashFlow"`
}

type portfolioCashFlowPeriodDTO struct {
	Month  string                    `json:"month"`
	Totals portfolioCashFlowTotalsDTO `json:"totals"`
}

type portfolioCashFlowCalculationDTO struct {
	MethodologyVersion string  `json:"methodologyVersion"`
	InputsAsOf         *string `json:"inputsAsOf"`
}

func (api *API) getPortfolioCashFlow(c fiber.Ctx) error {
	subjectID, err := api.subjectID(c)
	if err != nil {
		return writeMappedError(c, err)
	}
	fromDate, err := optionalQueryValue(c, "fromDate")
	if err != nil {
		return writeMappedError(c, err)
	}
	toDate, err := optionalQueryValue(c, "toDate")
	if err != nil {
		return writeMappedError(c, err)
	}
	projection, err := api.service.GetPortfolioCashFlow(c.Context(), subjectID, c.Params("portfolioId"), fromDate, toDate)
	if err != nil {
		return writeMappedError(c, err)
	}
	return writeOK(c, mapPortfolioCashFlow(projection))
}

func mapPortfolioCashFlow(projection verticalslice.PortfolioCashFlowProjection) portfolioCashFlowDTO {
	periods := make([]portfolioCashFlowPeriodDTO, 0, len(projection.Periods))
	for _, period := range projection.Periods {
		periods = append(periods, portfolioCashFlowPeriodDTO{Month: period.Month, Totals: mapPortfolioCashFlowTotals(period.Totals)})
	}
	return portfolioCashFlowDTO{
		PortfolioID: projection.PortfolioID,
		FromDate:    projection.FromDate,
		ToDate:      projection.ToDate,
		Totals:      mapPortfolioCashFlowTotals(projection.Totals),
		Periods:     periods,
		Calculation: portfolioCashFlowCalculationDTO{
			MethodologyVersion: projection.MethodologyVersion,
			InputsAsOf:         projection.InputsAsOf,
		},
	}
}

func mapPortfolioCashFlowTotals(totals verticalslice.PortfolioCashFlowTotals) portfolioCashFlowTotalsDTO {
	return portfolioCashFlowTotalsDTO{
		Deposits:            mapMoney(totals.Deposits),
		Withdrawals:         mapMoney(totals.Withdrawals),
		BuyOutflows:         mapMoney(totals.BuyOutflows),
		SellInflows:         mapMoney(totals.SellInflows),
		DividendsGross:      mapMoney(totals.DividendsGross),
		CouponsGross:        mapMoney(totals.CouponsGross),
		Fees:                mapMoney(totals.Fees),
		Taxes:               mapMoney(totals.Taxes),
		NetExternalFlow:     mapMoney(totals.NetExternalFlow),
		NetInvestmentIncome: mapMoney(totals.NetInvestmentIncome),
		NetCashFlow:         mapMoney(totals.NetCashFlow),
	}
}
