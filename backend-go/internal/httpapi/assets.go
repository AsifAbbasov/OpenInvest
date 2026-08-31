package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/gofiber/fiber/v3"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
	"strings"
	"unicode/utf8"
)

func (api *API) searchAssets(c fiber.Ctx) error {
	limit, err := queryLimitStrict(c, 20)
	if err != nil {
		return writeMappedError(c, err)
	}
	assetType, err := optionalQueryValue(c, "assetType")
	if err != nil {
		return writeMappedError(c, err)
	}
	cursor, err := optionalQueryValue(c, "cursor")
	if err != nil {
		return writeMappedError(c, err)
	}
	rawQuery := c.Query("query")
	if !utf8.ValidString(rawQuery) {
		return writeMappedError(c, fmt.Errorf("%w: query must be valid UTF-8", verticalslice.ErrInvalidInput))
	}
	if utf8.RuneCountInString(rawQuery) > 100 {
		return writeMappedError(c, fmt.Errorf("%w: query must be at most 100 characters", verticalslice.ErrInvalidInput))
	}
	query := strings.TrimSpace(rawQuery)
	filter := verticalslice.AssetSearchFilter{Query: query, AssetType: assetType, Limit: limit}
	if err := api.applyAssetCursor(cursor, query, assetType, &filter); err != nil {
		return writeMappedError(c, err)
	}
	result, err := api.service.SearchAssets(c.Context(), filter)
	if err != nil {
		return writeMappedError(c, err)
	}
	var nextCursor *string
	if result.NextTicker != nil {
		value, err := api.encodeAssetCursor(query, assetType, *result.NextTicker)
		if err != nil {
			return writeMappedError(c, err)
		}
		nextCursor = &value
	}
	return writeOK(c, listData[assetSummaryDTO]{
		Items: mapAssetSummaries(result.Items),
		Pagination: paginationDTO{
			NextCursor: nextCursor,
			HasMore:    result.HasMore,
			Limit:      result.Limit,
		},
	})
}

func (api *API) getAsset(c fiber.Ctx) error {
	if _, err := api.service.GetAsset(c.Context(), c.Params("ticker")); err != nil {
		return writeMappedError(c, err)
	}
	return writeMappedError(c, verticalslice.ErrNotFound)
}

type assetSummaryDTO struct {
	Ticker    string    `json:"ticker"`
	Name      string    `json:"name"`
	AssetType string    `json:"assetType"`
	Currency  string    `json:"currency"`
	LotSize   string    `json:"lotSize"`
	LastPrice *moneyDTO `json:"lastPrice"`
}

func mapAssetSummaries(items []verticalslice.AssetSummary) []assetSummaryDTO {
	mapped := make([]assetSummaryDTO, 0, len(items))
	for _, item := range items {
		mapped = append(mapped, assetSummaryDTO{
			Ticker:    item.Ticker,
			Name:      item.Name,
			AssetType: item.AssetType,
			Currency:  item.Currency,
			LotSize:   item.LotSize.String(),
			LastPrice: mapOptionalMoney(item.LastPrice),
		})
	}
	return mapped
}

func assetSearchQueryHash(query string) string {
	sum := sha256.Sum256([]byte(query))
	return hex.EncodeToString(sum[:])
}

func (api *API) encodeAssetCursor(query string, assetType string, ticker string) (string, error) {
	if strings.TrimSpace(ticker) == "" {
		return "", fmt.Errorf("asset cursor anchor is missing")
	}
	return api.signPaginationCursor(paginationCursorPayload{
		Version:   1,
		Resource:  "assets",
		AssetType: assetType,
		QueryHash: assetSearchQueryHash(query),
		Ticker:    ticker,
	})
}

func (api *API) applyAssetCursor(raw string, query string, assetType string, filter *verticalslice.AssetSearchFilter) error {
	if raw == "" {
		return nil
	}
	payload, err := api.verifyPaginationCursor(raw)
	if err != nil {
		return err
	}
	if payload.Version != 1 ||
		payload.Resource != "assets" ||
		payload.SubjectID != "" ||
		payload.PortfolioID != "" ||
		payload.AssetType != assetType ||
		payload.QueryHash != assetSearchQueryHash(query) ||
		payload.Ticker == "" ||
		payload.TransactionType != "" ||
		payload.FromDate != "" ||
		payload.ToDate != "" ||
		payload.UpdatedAt != "" ||
		payload.TradeDate != "" ||
		payload.EntryID != "" {
		return invalidPaginationCursor()
	}
	filter.AfterTicker = payload.Ticker
	return nil
}
