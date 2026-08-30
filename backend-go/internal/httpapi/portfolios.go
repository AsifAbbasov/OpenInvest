package httpapi

import (
	"github.com/gofiber/fiber/v3"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
	"net/http"
	"time"
)

func (api *API) listPortfolios(c fiber.Ctx) error {
	subjectID, err := api.subjectID(c)
	if err != nil {
		return writeMappedError(c, err)
	}
	limit, err := queryLimitStrict(c, 20)
	if err != nil {
		return writeMappedError(c, err)
	}
	cursor, err := optionalQueryValue(c, "cursor")
	if err != nil {
		return writeMappedError(c, err)
	}
	filter, err := api.decodePortfolioCursor(cursor, subjectID)
	if err != nil {
		return writeMappedError(c, err)
	}
	filter.Limit = limit + 1
	items, err := api.service.ListPortfolios(c.Context(), subjectID, filter)
	if err != nil {
		return writeMappedError(c, err)
	}
	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}
	var nextCursor *string
	if hasMore {
		value, err := api.encodePortfolioCursor(subjectID, items[len(items)-1])
		if err != nil {
			return writeMappedError(c, err)
		}
		nextCursor = &value
	}
	return writeOK(c, listData[portfolioDTO]{
		Items:      mapPortfolios(items),
		Pagination: paginationDTO{NextCursor: nextCursor, HasMore: hasMore, Limit: limit},
	})
}

func (api *API) createPortfolio(c fiber.Ctx) error {
	meta := requestMeta(c)
	subjectID, err := api.subjectID(c)
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	var request createPortfolioRequestDTO
	if err := decodeStrictJSON(c.Request().Body(), &request); err != nil {
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
	}
	portfolio, err := api.service.CreatePortfolio(c.Context(), meta.toApp(), subjectID, c.Get("Idempotency-Key"), c.Path(), verticalslice.CreatePortfolioRequest{
		Name:         request.Name,
		BaseCurrency: request.BaseCurrency,
	})
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	return writeStatusWithMeta(c, meta, http.StatusCreated, mapPortfolio(portfolio))
}

func (api *API) getPortfolio(c fiber.Ctx) error {
	subjectID, err := api.subjectID(c)
	if err != nil {
		return writeMappedError(c, err)
	}
	portfolio, err := api.service.GetPortfolio(c.Context(), subjectID, c.Params("portfolioId"))
	if err != nil {
		return writeMappedError(c, err)
	}
	return writeOK(c, mapPortfolio(portfolio))
}

func (api *API) getPortfolioSummary(c fiber.Ctx) error {
	subjectID, err := api.subjectID(c)
	if err != nil {
		return writeMappedError(c, err)
	}
	summary, err := api.service.GetPortfolioSummary(c.Context(), subjectID, c.Params("portfolioId"), c.Query("asOfDate"))
	if err != nil {
		return writeMappedError(c, err)
	}
	return writeOK(c, mapSummary(summary))
}

type createPortfolioRequestDTO struct {
	Name         string `json:"name"`
	BaseCurrency string `json:"baseCurrency"`
}

type portfolioDTO struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	BaseCurrency string `json:"baseCurrency"`
	Version      int64  `json:"version"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

type summaryDTO struct {
	PortfolioID       string         `json:"portfolioId"`
	AsOfDate          string         `json:"asOfDate"`
	TotalValue        moneyDTO       `json:"totalValue"`
	CashValue         moneyDTO       `json:"cashValue"`
	StockValue        moneyDTO       `json:"stockValue"`
	BondValue         moneyDTO       `json:"bondValue"`
	InvestedCapital   moneyDTO       `json:"investedCapital"`
	DividendsReceived moneyDTO       `json:"dividendsReceived"`
	CouponsReceived   moneyDTO       `json:"couponsReceived"`
	NominalReturnRate *string        `json:"nominalReturnRate"`
	XIRR              *string        `json:"xirr"`
	RealReturn        *realReturnDTO `json:"realReturn"`
	PurchasingPower   powerDTO       `json:"purchasingPower"`
	Positions         []any          `json:"positions"`
	Calculation       calcDTO        `json:"calculation"`
}

type realReturnDTO struct {
	NominalReturnRate string   `json:"nominalReturnRate"`
	InflationRate     string   `json:"inflationRate"`
	RealReturnRate    string   `json:"realReturnRate"`
	NominalGain       moneyDTO `json:"nominalGain"`
	RealGain          moneyDTO `json:"realGain"`
	FromDate          string   `json:"fromDate"`
	ToDate            string   `json:"toDate"`
	Methodology       string   `json:"methodologyVersion"`
}

type powerDTO struct {
	PortfolioValue moneyDTO `json:"portfolioValue"`
	AsOfDate       string   `json:"asOfDate"`
	Equivalents    []any    `json:"equivalents"`
}

type calcDTO struct {
	MethodologyVersion string `json:"methodologyVersion"`
	CalculatedAt       string `json:"calculatedAt"`
	InputsAsOf         string `json:"inputsAsOf"`
}

func mapPortfolios(items []verticalslice.Portfolio) []portfolioDTO {
	mapped := make([]portfolioDTO, 0, len(items))
	for _, item := range items {
		mapped = append(mapped, mapPortfolio(item))
	}
	return mapped
}

func mapPortfolio(item verticalslice.Portfolio) portfolioDTO {
	return portfolioDTO{
		ID:           item.ID,
		Name:         item.Name,
		BaseCurrency: item.BaseCurrency,
		Version:      item.Version,
		CreatedAt:    item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func mapSummary(item verticalslice.PortfolioSummary) summaryDTO {
	return summaryDTO{
		PortfolioID:       item.PortfolioID,
		AsOfDate:          item.AsOfDate,
		TotalValue:        mapMoney(item.TotalValue),
		CashValue:         mapMoney(item.CashValue),
		StockValue:        mapMoney(item.StockValue),
		BondValue:         mapMoney(item.BondValue),
		InvestedCapital:   mapMoney(item.InvestedCapital),
		DividendsReceived: mapMoney(item.DividendsReceived),
		CouponsReceived:   mapMoney(item.CouponsReceived),
		NominalReturnRate: nil,
		XIRR:              nil,
		RealReturn:        nil,
		PurchasingPower: powerDTO{
			PortfolioValue: mapMoney(item.PurchasingPower.PortfolioValue),
			AsOfDate:       item.PurchasingPower.AsOfDate,
			Equivalents:    []any{},
		},
		Positions: []any{},
		Calculation: calcDTO{
			MethodologyVersion: item.MethodologyVersion,
			CalculatedAt:       item.CalculatedAt.UTC().Format(time.RFC3339),
			InputsAsOf:         item.AsOfDate,
		},
	}
}

func (api *API) encodePortfolioCursor(subjectID string, item verticalslice.Portfolio) (string, error) {
	return api.signPaginationCursor(paginationCursorPayload{
		Version:   1,
		Resource:  "portfolios",
		SubjectID: subjectID,
		UpdatedAt: item.UpdatedAt.UTC().Format(time.RFC3339Nano),
		EntryID:   item.ID,
	})
}

func (api *API) decodePortfolioCursor(raw string, subjectID string) (verticalslice.PortfolioFilter, error) {
	if raw == "" {
		return verticalslice.PortfolioFilter{}, nil
	}
	payload, err := api.verifyPaginationCursor(raw)
	if err != nil {
		return verticalslice.PortfolioFilter{}, err
	}
	if payload.Version != 1 || payload.Resource != "portfolios" || payload.SubjectID != subjectID || payload.PortfolioID != "" || payload.EntryID == "" || payload.UpdatedAt == "" {
		return verticalslice.PortfolioFilter{}, invalidPaginationCursor()
	}
	updatedAt, err := time.Parse(time.RFC3339Nano, payload.UpdatedAt)
	if err != nil {
		return verticalslice.PortfolioFilter{}, invalidPaginationCursor()
	}
	return verticalslice.PortfolioFilter{BeforeUpdatedAt: &updatedAt, BeforeID: payload.EntryID}, nil
}
