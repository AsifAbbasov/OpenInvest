package httpapi

import (
	"github.com/gofiber/fiber/v3"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

type portfolioPositionsDTO struct {
	Items                 []portfolioPositionProjectionDTO `json:"items"`
	TotalAcquisitionBasis moneyDTO                         `json:"totalAcquisitionBasis"`
	Calculation           portfolioPositionsCalculationDTO `json:"calculation"`
}

type portfolioPositionProjectionDTO struct {
	Ticker                 string                        `json:"ticker"`
	AssetType              string                        `json:"assetType"`
	Quantity               string                        `json:"quantity"`
	WeightedAverageCost    moneyDTO                      `json:"weightedAverageCost"`
	AcquisitionBasis       moneyDTO                      `json:"acquisitionBasis"`
	AcquisitionBasisWeight *string                       `json:"acquisitionBasisWeight"`
	MarketValuation        marketValuationUnavailableDTO `json:"marketValuation"`
}

type marketValuationUnavailableDTO struct {
	Status         string    `json:"status"`
	Reason         string    `json:"reason"`
	MarketPrice    *moneyDTO `json:"marketPrice"`
	MarketValue    *moneyDTO `json:"marketValue"`
	UnrealizedGain *moneyDTO `json:"unrealizedGain"`
	MarketWeight   *string   `json:"marketWeight"`
	Provider       *string   `json:"provider"`
	AsOf           *string   `json:"asOf"`
}

type portfolioPositionsCalculationDTO struct {
	MethodologyVersion string  `json:"methodologyVersion"`
	InputsAsOf         *string `json:"inputsAsOf"`
}

func (api *API) getPortfolioPositions(c fiber.Ctx) error {
	subjectID, err := api.subjectID(c)
	if err != nil {
		return writeMappedError(c, err)
	}
	asOfDate, err := optionalQueryValue(c, "asOfDate")
	if err != nil {
		return writeMappedError(c, err)
	}
	projection, err := api.service.GetPortfolioPositions(c.Context(), subjectID, c.Params("portfolioId"), asOfDate)
	if err != nil {
		return writeMappedError(c, err)
	}
	return writeOK(c, mapPortfolioPositions(projection))
}

func mapPortfolioPositions(projection verticalslice.PortfolioPositionsProjection) portfolioPositionsDTO {
	items := make([]portfolioPositionProjectionDTO, 0, len(projection.Items))
	for _, item := range projection.Items {
		var acquisitionBasisWeight *string
		if item.AcquisitionBasisWeight != nil {
			value := item.AcquisitionBasisWeight.String()
			acquisitionBasisWeight = &value
		}
		items = append(items, portfolioPositionProjectionDTO{
			Ticker:                 item.Ticker,
			AssetType:              item.AssetType,
			Quantity:               item.Quantity.String(),
			WeightedAverageCost:    mapMoney(item.WeightedAverageCost),
			AcquisitionBasis:       mapMoney(item.AcquisitionBasis),
			AcquisitionBasisWeight: acquisitionBasisWeight,
			MarketValuation: marketValuationUnavailableDTO{
				Status: item.MarketValuation.Status,
				Reason: item.MarketValuation.Reason,
			},
		})
	}
	return portfolioPositionsDTO{
		Items:                 items,
		TotalAcquisitionBasis: mapMoney(projection.TotalAcquisitionBasis),
		Calculation: portfolioPositionsCalculationDTO{
			MethodologyVersion: projection.MethodologyVersion,
			InputsAsOf:         projection.InputsAsOf,
		},
	}
}
