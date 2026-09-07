package httpapi

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

type portfolioPositionsDTO struct {
	Items                 []portfolioPositionProjectionDTO `json:"items"`
	TotalAcquisitionBasis moneyDTO                         `json:"totalAcquisitionBasis"`
	ValuationSummary      portfolioValuationSummaryDTO    `json:"valuationSummary"`
	Calculation           portfolioPositionsCalculationDTO `json:"calculation"`
}

type portfolioPositionProjectionDTO struct {
	Ticker                 string             `json:"ticker"`
	AssetType              string             `json:"assetType"`
	Quantity               string             `json:"quantity"`
	WeightedAverageCost    moneyDTO           `json:"weightedAverageCost"`
	AcquisitionBasis       moneyDTO           `json:"acquisitionBasis"`
	AcquisitionBasisWeight *string            `json:"acquisitionBasisWeight"`
	MarketValuation        marketValuationDTO `json:"marketValuation"`
}

type marketValuationDTO struct {
	Status           string    `json:"status"`
	Reason           *string   `json:"reason"`
	Source           *string   `json:"source"`
	MarketPrice      *moneyDTO `json:"marketPrice"`
	MarketValue      *moneyDTO `json:"marketValue"`
	UnrealizedGain   *moneyDTO `json:"unrealizedGain"`
	UnrealizedReturn *string   `json:"unrealizedReturn"`
	MarketWeight     *string   `json:"marketWeight"`
	Provider         *string   `json:"provider"`
	AsOf             *string   `json:"asOf"`
}

type portfolioValuationSummaryDTO struct {
	Status                          string    `json:"status"`
	ValuedPositions                 int       `json:"valuedPositions"`
	TotalOpenPositions              int       `json:"totalOpenPositions"`
	ValuedPositionsMarketValue      moneyDTO  `json:"valuedPositionsMarketValue"`
	ValuedPositionsAcquisitionBasis moneyDTO  `json:"valuedPositionsAcquisitionBasis"`
	TotalAcquisitionBasis           moneyDTO  `json:"totalAcquisitionBasis"`
	UnrealizedGain                  moneyDTO  `json:"unrealizedGain"`
	CashValue                       moneyDTO  `json:"cashValue"`
	CurrentPortfolioValue           *moneyDTO `json:"currentPortfolioValue"`
}

type portfolioPositionsCalculationDTO struct {
	MethodologyVersion string  `json:"methodologyVersion"`
	InputsAsOf         *string `json:"inputsAsOf"`
}

type manualValuationRequestDTO struct {
	MarketPrice moneyDTO `json:"marketPrice"`
	AsOfDate    string   `json:"asOfDate"`
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

func (api *API) upsertManualValuation(c fiber.Ctx) error {
	meta := requestMeta(c)
	subjectID, err := api.subjectID(c)
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	var request manualValuationRequestDTO
	if err := decodeStrictJSON(c.Request().Body(), &request); err != nil {
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
	}
	price, err := parseMoney(request.MarketPrice)
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	projection, err := api.service.UpsertManualValuation(c.Context(), subjectID, verticalslice.ManualValuationRequest{
		PortfolioID: c.Params("portfolioId"),
		Ticker:      c.Params("ticker"),
		Price:       price,
		AsOfDate:    request.AsOfDate,
	})
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	return writeOK(c, mapPortfolioPositions(projection))
}

func (api *API) clearManualValuation(c fiber.Ctx) error {
	meta := requestMeta(c)
	subjectID, err := api.subjectID(c)
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	projection, err := api.service.ClearManualValuation(
		c.Context(),
		subjectID,
		c.Params("portfolioId"),
		c.Params("ticker"),
	)
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
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
			MarketValuation:        mapMarketValuation(item.MarketValuation),
		})
	}
	return portfolioPositionsDTO{
		Items:                 items,
		TotalAcquisitionBasis: mapMoney(projection.TotalAcquisitionBasis),
		ValuationSummary: portfolioValuationSummaryDTO{
			Status:                          projection.ValuationSummary.Status,
			ValuedPositions:                 projection.ValuationSummary.ValuedPositions,
			TotalOpenPositions:              projection.ValuationSummary.TotalOpenPositions,
			ValuedPositionsMarketValue:      mapMoney(projection.ValuationSummary.ValuedPositionsMarketValue),
			ValuedPositionsAcquisitionBasis: mapMoney(projection.ValuationSummary.ValuedPositionsAcquisitionBasis),
			TotalAcquisitionBasis:           mapMoney(projection.ValuationSummary.TotalAcquisitionBasis),
			UnrealizedGain:                  mapMoney(projection.ValuationSummary.UnrealizedGain),
			CashValue:                       mapMoney(projection.ValuationSummary.CashValue),
			CurrentPortfolioValue:           mapOptionalMoney(projection.ValuationSummary.CurrentPortfolioValue),
		},
		Calculation: portfolioPositionsCalculationDTO{
			MethodologyVersion: projection.MethodologyVersion,
			InputsAsOf:         projection.InputsAsOf,
		},
	}
}

func mapMarketValuation(item verticalslice.MarketValuationUnavailable) marketValuationDTO {
	result := marketValuationDTO{
		Status:         item.Status,
		MarketPrice:    mapOptionalMoney(item.MarketPrice),
		MarketValue:    mapOptionalMoney(item.MarketValue),
		UnrealizedGain: mapOptionalMoney(item.UnrealizedGain),
		AsOf:           item.AsOf,
	}
	if item.Reason != "" {
		value := item.Reason
		result.Reason = &value
	}
	if item.Source != "" {
		value := item.Source
		result.Source = &value
	}
	if item.UnrealizedReturn != nil {
		value := item.UnrealizedReturn.String()
		result.UnrealizedReturn = &value
	}
	if item.MarketWeight != nil {
		value := item.MarketWeight.String()
		result.MarketWeight = &value
	}
	return result
}
