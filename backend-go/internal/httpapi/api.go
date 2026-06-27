package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

const devSubjectID = "00000000-0000-4000-8000-000000000001"

type API struct {
	service *verticalslice.Service
}

func New(service *verticalslice.Service) *fiber.App {
	api := &API{service: service}
	app := fiber.New(fiber.Config{AppName: "OpenInvest API"})

	app.Get("/api/v1/health", api.health)
	app.Get("/api/v1/ready", api.ready)
	app.Get("/api/v1/portfolios", api.listPortfolios)
	app.Post("/api/v1/portfolios", api.createPortfolio)
	app.Get("/api/v1/portfolios/:portfolioId", api.getPortfolio)
	app.Get("/api/v1/portfolios/:portfolioId/summary", api.getPortfolioSummary)
	app.Get("/api/v1/portfolios/:portfolioId/transactions", api.listTransactions)
	app.Post("/api/v1/portfolios/:portfolioId/transactions", api.appendTransaction)

	return app
}

func (api *API) health(c fiber.Ctx) error {
	return writeOK(c, map[string]string{"status": "ok"})
}

func (api *API) ready(c fiber.Ctx) error {
	if err := api.service.Ready(c.Context()); err != nil {
		return writeError(c, http.StatusServiceUnavailable, "SERVICE_NOT_READY", "Service is not ready")
	}
	return writeOK(c, map[string]string{"status": "ready"})
}

func (api *API) listPortfolios(c fiber.Ctx) error {
	items, err := api.service.ListPortfolios(c.Context(), subjectID(), queryLimit(c, 20))
	if err != nil {
		return writeMappedError(c, err)
	}
	return writeOK(c, listData[portfolioDTO]{Items: mapPortfolios(items), Pagination: paginationDTO{Limit: queryLimit(c, 20)}})
}

func (api *API) createPortfolio(c fiber.Ctx) error {
	var request createPortfolioRequestDTO
	if err := c.Bind().Body(&request); err != nil {
		return writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
	}
	portfolio, err := api.service.CreatePortfolio(c.Context(), subjectID(), c.Get("Idempotency-Key"), c.Path(), verticalslice.CreatePortfolioRequest{
		Name:         request.Name,
		BaseCurrency: request.BaseCurrency,
	})
	if err != nil {
		return writeMappedError(c, err)
	}
	return writeStatus(c, http.StatusCreated, mapPortfolio(portfolio))
}

func (api *API) getPortfolio(c fiber.Ctx) error {
	portfolio, err := api.service.GetPortfolio(c.Context(), subjectID(), c.Params("portfolioId"))
	if err != nil {
		return writeMappedError(c, err)
	}
	return writeOK(c, mapPortfolio(portfolio))
}

func (api *API) listTransactions(c fiber.Ctx) error {
	items, err := api.service.ListTransactions(c.Context(), subjectID(), c.Params("portfolioId"), queryLimit(c, 50))
	if err != nil {
		return writeMappedError(c, err)
	}
	return writeOK(c, listData[transactionDTO]{Items: mapTransactions(items), Pagination: paginationDTO{Limit: queryLimit(c, 50)}})
}

func (api *API) appendTransaction(c fiber.Ctx) error {
	var request appendTransactionRequestDTO
	if err := c.Bind().Body(&request); err != nil {
		return writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
	}
	appRequest, err := request.toApp(c.Params("portfolioId"))
	if err != nil {
		return writeMappedError(c, err)
	}
	transaction, err := api.service.AppendTransaction(c.Context(), subjectID(), c.Get("Idempotency-Key"), c.Path(), appRequest)
	if err != nil {
		return writeMappedError(c, err)
	}
	return writeStatus(c, http.StatusCreated, mapTransaction(transaction))
}

func (api *API) getPortfolioSummary(c fiber.Ctx) error {
	summary, err := api.service.GetPortfolioSummary(c.Context(), subjectID(), c.Params("portfolioId"), c.Query("asOfDate"))
	if err != nil {
		return writeMappedError(c, err)
	}
	return writeOK(c, mapSummary(summary))
}

func subjectID() string {
	if configured := strings.TrimSpace(os.Getenv("OPENINVEST_DEV_SUBJECT_ID")); configured != "" {
		return configured
	}
	return devSubjectID
}

func queryLimit(c fiber.Ctx, fallback int) int {
	value, err := strconv.Atoi(c.Query("limit", strconv.Itoa(fallback)))
	if err != nil || value <= 0 {
		return fallback
	}
	if value > 100 {
		return 100
	}
	return value
}

func writeOK(c fiber.Ctx, data any) error {
	return writeStatus(c, http.StatusOK, data)
}

func writeStatus(c fiber.Ctx, status int, data any) error {
	meta := responseMeta()
	c.Set("X-Request-ID", meta.RequestID)
	c.Set("X-Trace-ID", meta.TraceID)
	return c.Status(status).JSON(baseResponse{Data: data, Meta: meta})
}

func writeMappedError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, verticalslice.ErrInvalidInput):
		return writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, verticalslice.ErrMissingIdempotency):
		return writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Idempotency-Key header is required")
	case errors.Is(err, postgres.ErrNotFound):
		return writeError(c, http.StatusNotFound, "NOT_FOUND", "Resource not found")
	case errors.Is(err, postgres.ErrIdempotencyConflict):
		return writeError(c, http.StatusConflict, "IDEMPOTENCY_CONFLICT", "Idempotency-Key is already bound to another request")
	case errors.Is(err, postgres.ErrIdempotencyInFlight):
		return writeError(c, http.StatusConflict, "IDEMPOTENCY_IN_FLIGHT", "Idempotency-Key is currently being processed")
	default:
		return writeError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
	}
}

func writeError(c fiber.Ctx, status int, code string, message string) error {
	meta := responseMeta()
	c.Set("X-Request-ID", meta.RequestID)
	c.Set("X-Trace-ID", meta.TraceID)
	return c.Status(status).JSON(errorResponse{
		Error: errorBody{Code: code, Message: message, Details: []errorDetailDTO{}},
		Meta:  meta,
	})
}

func responseMeta() metaDTO {
	return metaDTO{
		RequestID:   uuid.NewString(),
		TraceID:     traceID(),
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func traceID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "00000000000000000000000000000001"
	}
	return hex.EncodeToString(bytes)
}

type baseResponse struct {
	Data any     `json:"data"`
	Meta metaDTO `json:"meta"`
}

type errorResponse struct {
	Error errorBody `json:"error"`
	Meta  metaDTO   `json:"meta"`
}

type errorBody struct {
	Code    string           `json:"code"`
	Message string           `json:"message"`
	Details []errorDetailDTO `json:"details"`
}

type errorDetailDTO struct {
	Field   *string `json:"field"`
	Code    string  `json:"code"`
	Message string  `json:"message"`
}

type metaDTO struct {
	RequestID   string `json:"requestId"`
	TraceID     string `json:"traceId"`
	GeneratedAt string `json:"generatedAt"`
}

type paginationDTO struct {
	NextCursor *string `json:"nextCursor"`
	HasMore    bool    `json:"hasMore"`
	Limit      int     `json:"limit"`
}

type listData[T any] struct {
	Items      []T           `json:"items"`
	Pagination paginationDTO `json:"pagination"`
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

type moneyDTO struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
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

func parseOptionalDecimal(value *string) (*decimal.Decimal, error) {
	if value == nil {
		return nil, nil
	}
	parsed, err := decimal.FromString(*value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseOptionalMoney(value *moneyDTO) (*verticalslice.Money, error) {
	if value == nil {
		return nil, nil
	}
	parsed, err := parseMoney(*value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseMoney(value moneyDTO) (verticalslice.Money, error) {
	amount, err := decimal.FromString(value.Amount)
	if err != nil {
		return verticalslice.Money{}, err
	}
	return verticalslice.Money{Amount: amount, Currency: value.Currency}, nil
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

type summaryDTO struct {
	PortfolioID       string        `json:"portfolioId"`
	AsOfDate          string        `json:"asOfDate"`
	TotalValue        moneyDTO      `json:"totalValue"`
	CashValue         moneyDTO      `json:"cashValue"`
	StockValue        moneyDTO      `json:"stockValue"`
	BondValue         moneyDTO      `json:"bondValue"`
	InvestedCapital   moneyDTO      `json:"investedCapital"`
	DividendsReceived moneyDTO      `json:"dividendsReceived"`
	CouponsReceived   moneyDTO      `json:"couponsReceived"`
	NominalReturnRate string        `json:"nominalReturnRate"`
	XIRR              *string       `json:"xirr"`
	RealReturn        realReturnDTO `json:"realReturn"`
	PurchasingPower   powerDTO      `json:"purchasingPower"`
	Positions         []any         `json:"positions"`
	Calculation       calcDTO       `json:"calculation"`
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
		NominalReturnRate: item.NominalReturnRate.String(),
		XIRR:              nil,
		RealReturn: realReturnDTO{
			NominalReturnRate: item.RealReturn.NominalReturnRate.String(),
			InflationRate:     item.RealReturn.InflationRate.String(),
			RealReturnRate:    item.RealReturn.RealReturnRate.String(),
			NominalGain:       mapMoney(item.RealReturn.NominalGain),
			RealGain:          mapMoney(item.RealReturn.RealGain),
			FromDate:          item.RealReturn.FromDate,
			ToDate:            item.RealReturn.ToDate,
			Methodology:       item.RealReturn.Methodology,
		},
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

func mapOptionalMoney(item *verticalslice.Money) *moneyDTO {
	if item == nil {
		return nil
	}
	value := mapMoney(*item)
	return &value
}

func mapMoney(item verticalslice.Money) moneyDTO {
	return moneyDTO{Amount: item.Amount.String(), Currency: item.Currency}
}
