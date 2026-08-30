package httpapi

import (
	"fmt"
	"github.com/gofiber/fiber/v3"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
	"net/http"
	"strings"
	"time"
)

func (api *API) listTransactions(c fiber.Ctx) error {
	subjectID, err := api.subjectID(c)
	if err != nil {
		return writeMappedError(c, err)
	}
	limit, err := queryLimitStrict(c, 50)
	if err != nil {
		return writeMappedError(c, err)
	}
	portfolioID := c.Params("portfolioId")
	filter := verticalslice.TransactionFilter{
		TransactionType: strings.TrimSpace(c.Query("transactionType")),
		FromDate:        strings.TrimSpace(c.Query("fromDate")),
		ToDate:          strings.TrimSpace(c.Query("toDate")),
	}
	cursor, err := optionalQueryValue(c, "cursor")
	if err != nil {
		return writeMappedError(c, err)
	}
	if err := api.applyTransactionCursor(cursor, subjectID, portfolioID, &filter); err != nil {
		return writeMappedError(c, err)
	}
	filter.Limit = limit + 1
	items, err := api.service.ListTransactions(c.Context(), subjectID, portfolioID, filter)
	if err != nil {
		return writeMappedError(c, err)
	}
	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}
	var nextCursor *string
	if hasMore {
		value, err := api.encodeTransactionCursor(subjectID, portfolioID, filter, items[len(items)-1])
		if err != nil {
			return writeMappedError(c, err)
		}
		nextCursor = &value
	}
	return writeOK(c, listData[transactionDTO]{
		Items:      mapTransactions(items),
		Pagination: paginationDTO{NextCursor: nextCursor, HasMore: hasMore, Limit: limit},
	})
}

func (api *API) appendTransaction(c fiber.Ctx) error {
	meta := requestMeta(c)
	subjectID, err := api.subjectID(c)
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	if !jsonFieldPresent(c.Request().Body(), "settlementDate") {
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "settlementDate is required")
	}
	var request appendTransactionRequestDTO
	if err := decodeStrictJSON(c.Request().Body(), &request); err != nil {
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
	}
	appRequest, err := request.toApp(c.Params("portfolioId"))
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	transaction, err := api.service.AppendTransaction(c.Context(), meta.toApp(), subjectID, c.Get("Idempotency-Key"), c.Path(), appRequest)
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	return writeStatusWithMeta(c, meta, http.StatusCreated, mapTransaction(transaction))
}

type appendTransactionRequestDTO struct {
	TransactionType string    `json:"transactionType"`
	Ticker          *string   `json:"ticker"`
	Quantity        *string   `json:"quantity"`
	UnitPrice       *moneyDTO `json:"unitPrice"`
	GrossAmount     *moneyDTO `json:"grossAmount"`
	Commission      moneyDTO  `json:"commission"`
	Tax             moneyDTO  `json:"tax"`
	TradeDate       string    `json:"tradeDate"`
	SettlementDate  *string   `json:"settlementDate"`
	Note            *string   `json:"note"`
}

func (request appendTransactionRequestDTO) toApp(portfolioID string) (verticalslice.AppendTransactionRequest, error) {
	quantity, err := parseOptionalDecimal(request.Quantity)
	if err != nil {
		return verticalslice.AppendTransactionRequest{}, err
	}
	unitPrice, err := parseOptionalMoney(request.UnitPrice)
	if err != nil {
		return verticalslice.AppendTransactionRequest{}, err
	}
	grossAmount, err := parseOptionalMoney(request.GrossAmount)
	if err != nil {
		return verticalslice.AppendTransactionRequest{}, err
	}
	commission, err := parseMoney(request.Commission)
	if err != nil {
		return verticalslice.AppendTransactionRequest{}, err
	}
	tax, err := parseMoney(request.Tax)
	if err != nil {
		return verticalslice.AppendTransactionRequest{}, err
	}
	return verticalslice.AppendTransactionRequest{
		PortfolioID:     portfolioID,
		TransactionType: request.TransactionType,
		Ticker:          request.Ticker,
		Quantity:        quantity,
		UnitPrice:       unitPrice,
		GrossAmount:     grossAmount,
		Commission:      commission,
		Tax:             tax,
		TradeDate:       request.TradeDate,
		SettlementDate:  request.SettlementDate,
		Note:            request.Note,
	}, nil
}

type transactionDTO struct {
	ID              string    `json:"id"`
	PortfolioID     string    `json:"portfolioId"`
	TransactionType string    `json:"transactionType"`
	Status          string    `json:"status"`
	Ticker          *string   `json:"ticker"`
	Quantity        *string   `json:"quantity"`
	UnitPrice       *moneyDTO `json:"unitPrice"`
	GrossAmount     moneyDTO  `json:"grossAmount"`
	Commission      moneyDTO  `json:"commission"`
	Tax             moneyDTO  `json:"tax"`
	TradeDate       string    `json:"tradeDate"`
	SettlementDate  *string   `json:"settlementDate"`
	Note            *string   `json:"note,omitempty"`
	Revision        int       `json:"revision"`
	CreatedAt       string    `json:"createdAt"`
	UpdatedAt       string    `json:"updatedAt"`
}

func mapTransactions(items []verticalslice.Transaction) []transactionDTO {
	mapped := make([]transactionDTO, 0, len(items))
	for _, item := range items {
		mapped = append(mapped, mapTransaction(item))
	}
	return mapped
}

func mapTransaction(item verticalslice.Transaction) transactionDTO {
	var quantity *string
	if item.Quantity != nil {
		value := item.Quantity.String()
		quantity = &value
	}
	return transactionDTO{
		ID:              item.ID,
		PortfolioID:     item.PortfolioID,
		TransactionType: item.TransactionType,
		Status:          item.Status,
		Ticker:          item.Ticker,
		Quantity:        quantity,
		UnitPrice:       mapOptionalMoney(item.UnitPrice),
		GrossAmount:     mapMoney(item.GrossAmount),
		Commission:      mapMoney(item.Commission),
		Tax:             mapMoney(item.Tax),
		TradeDate:       item.TradeDate,
		SettlementDate:  item.SettlementDate,
		Note:            item.Note,
		Revision:        item.Revision,
		CreatedAt:       item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:       item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func (api *API) encodeTransactionCursor(subjectID string, portfolioID string, filter verticalslice.TransactionFilter, item verticalslice.Transaction) (string, error) {
	if strings.TrimSpace(item.EntryID) == "" {
		return "", fmt.Errorf("transaction cursor anchor is missing")
	}
	return api.signPaginationCursor(paginationCursorPayload{
		Version:         1,
		Resource:        "transactions",
		SubjectID:       subjectID,
		PortfolioID:     portfolioID,
		TransactionType: filter.TransactionType,
		FromDate:        filter.FromDate,
		ToDate:          filter.ToDate,
		TradeDate:       item.TradeDate,
		EntryID:         item.EntryID,
	})
}

func (api *API) applyTransactionCursor(raw string, subjectID string, portfolioID string, filter *verticalslice.TransactionFilter) error {
	if raw == "" {
		return nil
	}
	payload, err := api.verifyPaginationCursor(raw)
	if err != nil {
		return err
	}
	if payload.Version != 1 ||
		payload.Resource != "transactions" ||
		payload.SubjectID != subjectID ||
		payload.PortfolioID != portfolioID ||
		payload.TransactionType != filter.TransactionType ||
		payload.FromDate != filter.FromDate ||
		payload.ToDate != filter.ToDate ||
		payload.TradeDate == "" ||
		payload.EntryID == "" {
		return invalidPaginationCursor()
	}
	if _, err := time.Parse("2006-01-02", payload.TradeDate); err != nil {
		return invalidPaginationCursor()
	}
	filter.BeforeTradeDate = payload.TradeDate
	filter.BeforeEntryID = payload.EntryID
	return nil
}
