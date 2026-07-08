package httpapi

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/importer"
	"github.com/openinvest/openinvest/backend-go/internal/importflow"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

const devSubjectID = "00000000-0000-4000-8000-000000000001"
const maxHTTPImportPayloadBytes = 2 * 1024 * 1024
const maxHTTPImportRows = 100

type API struct {
	service *verticalslice.Service
}

func New(service *verticalslice.Service) *fiber.App {
	api := &API{service: service}
	app := fiber.New(fiber.Config{AppName: "OpenInvest API"})

	app.Use(localDevelopmentCORS)

	app.Get("/api/v1/health", api.health)
	app.Get("/api/v1/ready", api.ready)
	app.Get("/api/v1/portfolios", api.listPortfolios)
	app.Post("/api/v1/portfolios", api.createPortfolio)
	app.Get("/api/v1/portfolios/:portfolioId", api.getPortfolio)
	app.Get("/api/v1/portfolios/:portfolioId/summary", api.getPortfolioSummary)
	app.Get("/api/v1/portfolios/:portfolioId/transactions", api.listTransactions)
	app.Post("/api/v1/portfolios/:portfolioId/transactions", api.appendTransaction)
	app.Post("/api/v1/portfolios/:portfolioId/imports/review", api.reviewImport)
	app.Post("/api/v1/portfolios/:portfolioId/imports/append", api.appendImport)

	return app
}

func localDevelopmentCORS(c fiber.Ctx) error {
	origin := strings.TrimSpace(c.Get("Origin"))
	if origin != "" && allowedWebOrigin(origin) {
		c.Set("Access-Control-Allow-Origin", origin)
		c.Set("Vary", "Origin")
		c.Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Accept, Content-Type, Idempotency-Key, X-Request-ID, traceparent")
		c.Set("Access-Control-Expose-Headers", "X-Request-ID, X-Trace-ID")
	}
	if c.Method() == http.MethodOptions {
		return c.SendStatus(http.StatusNoContent)
	}
	return c.Next()
}

func allowedWebOrigin(origin string) bool {
	configured := strings.TrimSpace(os.Getenv("OPENINVEST_ALLOWED_WEB_ORIGINS"))
	if configured == "" {
		configured = "http://localhost:3000,http://127.0.0.1:3000"
	}
	for _, allowed := range strings.Split(configured, ",") {
		if strings.TrimSpace(allowed) == origin {
			return true
		}
	}
	return false
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
	meta := requestMeta(c)
	var request createPortfolioRequestDTO
	if err := c.Bind().Body(&request); err != nil {
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
	}
	portfolio, err := api.service.CreatePortfolio(c.Context(), meta.toApp(), subjectID(), c.Get("Idempotency-Key"), c.Path(), verticalslice.CreatePortfolioRequest{
		Name:         request.Name,
		BaseCurrency: request.BaseCurrency,
	})
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	return writeStatusWithMeta(c, meta, http.StatusCreated, mapPortfolio(portfolio))
}

func (api *API) getPortfolio(c fiber.Ctx) error {
	portfolio, err := api.service.GetPortfolio(c.Context(), subjectID(), c.Params("portfolioId"))
	if err != nil {
		return writeMappedError(c, err)
	}
	return writeOK(c, mapPortfolio(portfolio))
}

func (api *API) listTransactions(c fiber.Ctx) error {
	if strings.TrimSpace(c.Query("cursor")) != "" {
		return writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "cursor pagination is outside Stage 3.2 scope")
	}
	filter := verticalslice.TransactionFilter{
		TransactionType: c.Query("transactionType"),
		FromDate:        c.Query("fromDate"),
		ToDate:          c.Query("toDate"),
		Limit:           queryLimit(c, 50),
	}
	items, err := api.service.ListTransactions(c.Context(), subjectID(), c.Params("portfolioId"), filter)
	if err != nil {
		return writeMappedError(c, err)
	}
	return writeOK(c, listData[transactionDTO]{Items: mapTransactions(items), Pagination: paginationDTO{Limit: queryLimit(c, 50)}})
}

func (api *API) appendTransaction(c fiber.Ctx) error {
	meta := requestMeta(c)
	if !jsonFieldPresent(c.Request().Body(), "settlementDate") {
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "settlementDate is required")
	}
	var request appendTransactionRequestDTO
	if err := c.Bind().Body(&request); err != nil {
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
	}
	appRequest, err := request.toApp(c.Params("portfolioId"))
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	transaction, err := api.service.AppendTransaction(c.Context(), meta.toApp(), subjectID(), c.Get("Idempotency-Key"), c.Path(), appRequest)
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	return writeStatusWithMeta(c, meta, http.StatusCreated, mapTransaction(transaction))
}

func (api *API) reviewImport(c fiber.Ctx) error {
	meta := requestMeta(c)
	var request importReviewRequestDTO
	if err := c.Bind().Body(&request); err != nil {
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
	}
	if err := request.validate(); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	portfolioID := c.Params("portfolioId")
	existing, err := api.service.ListTransactions(c.Context(), subjectID(), portfolioID, verticalslice.TransactionFilter{Limit: 100})
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	fileHash := importPayloadHash(request.CSVPayload)
	review, err := importer.ReviewCSV(importer.ReviewRequest{
		SubjectID:          subjectID(),
		PortfolioID:        portfolioID,
		SourceKind:         importer.SourceKindUserUploadedFile,
		SourceAccountLabel: request.SourceAccountLabel,
		FileHash:           fileHash,
		Existing:           existing,
		Reader:             strings.NewReader(request.CSVPayload),
	})
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	if err := validateImportRowCount(review.Summary.TotalRows); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	return writeStatusWithMeta(c, meta, http.StatusOK, mapImportReview(review))
}

func (api *API) appendImport(c fiber.Ctx) error {
	meta := requestMeta(c)
	if err := verticalslice.ValidateIdempotencyKey(c.Get("Idempotency-Key")); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	var request importAppendRequestDTO
	if err := c.Bind().Body(&request); err != nil {
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
	}
	if err := request.validate(); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	portfolioID := c.Params("portfolioId")
	fileHash := importPayloadHash(request.CSVPayload)
	preflightReview, err := importer.ReviewCSV(importer.ReviewRequest{
		SubjectID:          subjectID(),
		PortfolioID:        portfolioID,
		SourceKind:         importer.SourceKindUserUploadedFile,
		SourceAccountLabel: request.SourceAccountLabel,
		FileHash:           fileHash,
		Reader:             strings.NewReader(request.CSVPayload),
	})
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	if err := validateImportRowCount(preflightReview.Summary.TotalRows); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	existing, err := api.service.ListTransactions(c.Context(), subjectID(), portfolioID, verticalslice.TransactionFilter{Limit: 100})
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	result, err := importflow.ReviewAndAppend(c.Context(), api.service, importflow.Request{
		RequestContext:     meta.toApp(),
		SubjectID:          subjectID(),
		PortfolioID:        portfolioID,
		IdempotencyKey:     c.Get("Idempotency-Key"),
		RequestPath:        c.Path(),
		SourceKind:         importer.SourceKindUserUploadedFile,
		SourceAccountLabel: request.SourceAccountLabel,
		SourceFileHash:     fileHash,
		Existing:           existing,
		Reader:             strings.NewReader(request.CSVPayload),
		Decisions:          request.toAppDecisions(),
	})
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	return writeStatusWithMeta(c, meta, http.StatusCreated, mapImportAppendResult(portfolioID, importer.SourceKindUserUploadedFile, fileHash, result))
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
	return writeStatusWithMeta(c, requestMeta(c), status, data)
}

func writeStatusWithMeta(c fiber.Ctx, meta metaDTO, status int, data any) error {
	c.Set("X-Request-ID", meta.RequestID)
	c.Set("X-Trace-ID", meta.TraceID)
	return c.Status(status).JSON(baseResponse{Data: data, Meta: meta})
}

func writeMappedError(c fiber.Ctx, err error) error {
	return writeMappedErrorWithMeta(c, requestMeta(c), err)
}

func writeMappedErrorWithMeta(c fiber.Ctx, meta metaDTO, err error) error {
	switch {
	case errors.Is(err, verticalslice.ErrInvalidInput):
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, verticalslice.ErrMissingIdempotency):
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "Idempotency-Key header is required")
	case errors.Is(err, importer.ErrInvalidImport):
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, importer.ErrUnsafeAppend):
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, importflow.ErrInvalidFlowInput):
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, importflow.ErrNoApprovedRows):
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "at least one appendable import row must be approved")
	case errors.Is(err, postgres.ErrNotFound):
		return writeErrorWithMeta(c, meta, http.StatusNotFound, "NOT_FOUND", "Resource not found")
	case errors.Is(err, postgres.ErrIdempotencyConflict):
		return writeErrorWithMeta(c, meta, http.StatusConflict, "IDEMPOTENCY_CONFLICT", "Idempotency-Key is already bound to another request")
	case errors.Is(err, postgres.ErrIdempotencyInFlight):
		return writeErrorWithMeta(c, meta, http.StatusConflict, "IDEMPOTENCY_IN_FLIGHT", "Idempotency-Key is currently being processed")
	default:
		return writeErrorWithMeta(c, meta, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
	}
}

func writeError(c fiber.Ctx, status int, code string, message string) error {
	return writeErrorWithMeta(c, requestMeta(c), status, code, message)
}

func writeErrorWithMeta(c fiber.Ctx, meta metaDTO, status int, code string, message string) error {
	c.Set("X-Request-ID", meta.RequestID)
	c.Set("X-Trace-ID", meta.TraceID)
	return c.Status(status).JSON(errorResponse{
		Error: errorBody{Code: code, Message: message, Details: []errorDetailDTO{}},
		Meta:  meta,
	})
}

func requestMeta(c fiber.Ctx) metaDTO {
	requestID := strings.TrimSpace(c.Get("X-Request-ID"))
	if _, err := uuid.Parse(requestID); err != nil {
		requestID = uuid.NewString()
	}
	traceIDValue := traceID()
	if traceparent := strings.TrimSpace(c.Get("traceparent")); validTraceparent(traceparent) {
		traceIDValue = strings.Split(traceparent, "-")[1]
	}
	return metaDTO{
		RequestID:   requestID,
		TraceID:     traceIDValue,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func (meta metaDTO) toApp() verticalslice.RequestContext {
	return verticalslice.RequestContext{RequestID: meta.RequestID, TraceID: meta.TraceID}
}

func traceID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "00000000000000000000000000000001"
	}
	return hex.EncodeToString(bytes)
}

func validTraceparent(value string) bool {
	parts := strings.Split(value, "-")
	if len(parts) != 4 {
		return false
	}
	version, traceIDPart, parentIDPart, flags := parts[0], parts[1], parts[2], parts[3]
	return len(version) == 2 &&
		len(traceIDPart) == 32 &&
		len(parentIDPart) == 16 &&
		len(flags) == 2 &&
		version != "ff" &&
		traceIDPart != strings.Repeat("0", 32) &&
		parentIDPart != strings.Repeat("0", 16) &&
		isLowerHex(version) &&
		isLowerHex(traceIDPart) &&
		isLowerHex(parentIDPart) &&
		isLowerHex(flags)
}

func isLowerHex(value string) bool {
	for _, char := range value {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
			return false
		}
	}
	return true
}

func jsonFieldPresent(body []byte, field string) bool {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return false
	}
	_, ok := raw[field]
	return ok
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

type importReviewRequestDTO struct {
	SourceAccountLabel string `json:"sourceAccountLabel"`
	CSVPayload         string `json:"csvPayload"`
}

func (request importReviewRequestDTO) validate() error {
	if strings.TrimSpace(request.CSVPayload) == "" {
		return fmt.Errorf("%w: csvPayload is required", verticalslice.ErrInvalidInput)
	}
	if len([]byte(request.CSVPayload)) > maxHTTPImportPayloadBytes {
		return fmt.Errorf("%w: csvPayload exceeds 2097152 bytes", verticalslice.ErrInvalidInput)
	}
	if len(request.SourceAccountLabel) > 120 {
		return fmt.Errorf("%w: sourceAccountLabel must be at most 120 characters", verticalslice.ErrInvalidInput)
	}
	return nil
}

type importAppendRequestDTO struct {
	SourceAccountLabel string              `json:"sourceAccountLabel"`
	CSVPayload         string              `json:"csvPayload"`
	Decisions          []importDecisionDTO `json:"decisions"`
}

func (request importAppendRequestDTO) validate() error {
	if err := (importReviewRequestDTO{
		SourceAccountLabel: request.SourceAccountLabel,
		CSVPayload:         request.CSVPayload,
	}).validate(); err != nil {
		return err
	}
	if len(request.Decisions) == 0 {
		return fmt.Errorf("%w: decisions are required", verticalslice.ErrInvalidInput)
	}
	if len(request.Decisions) > 100 {
		return fmt.Errorf("%w: decisions must contain at most 100 rows", verticalslice.ErrInvalidInput)
	}
	return nil
}

func (request importAppendRequestDTO) toAppDecisions() []importer.Decision {
	decisions := make([]importer.Decision, 0, len(request.Decisions))
	for _, decision := range request.Decisions {
		decisions = append(decisions, importer.Decision{RowNumber: decision.RowNumber, Action: decision.Action})
	}
	return decisions
}

type importDecisionDTO struct {
	RowNumber int    `json:"rowNumber"`
	Action    string `json:"action"`
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

type importReviewDTO struct {
	PortfolioID        string               `json:"portfolioId"`
	SourceKind         string               `json:"sourceKind"`
	SourceAccountLabel string               `json:"sourceAccountLabel"`
	SourceFileHash     string               `json:"sourceFileHash"`
	RetentionPolicy    string               `json:"retentionPolicy"`
	ReviewGuarantee    string               `json:"reviewGuarantee"`
	Summary            importSummaryDTO     `json:"summary"`
	Rows               []importRowReviewDTO `json:"rows"`
}

type importSummaryDTO struct {
	TotalRows      int `json:"totalRows"`
	AppendableRows int `json:"appendableRows"`
	DuplicateRows  int `json:"duplicateRows"`
	ConflictRows   int `json:"conflictRows"`
	InvalidRows    int `json:"invalidRows"`
}

type importRowReviewDTO struct {
	RowNumber   int                 `json:"rowNumber"`
	RowHash     string              `json:"rowHash"`
	Status      string              `json:"status"`
	ReasonCodes []string            `json:"reasonCodes"`
	Fingerprint string              `json:"fingerprint,omitempty"`
	Candidate   *importCandidateDTO `json:"candidate,omitempty"`
}

type importCandidateDTO struct {
	TransactionType string    `json:"transactionType"`
	Ticker          *string   `json:"ticker,omitempty"`
	Quantity        *string   `json:"quantity,omitempty"`
	UnitPrice       *moneyDTO `json:"unitPrice,omitempty"`
	GrossAmount     moneyDTO  `json:"grossAmount"`
	Commission      moneyDTO  `json:"commission"`
	Tax             moneyDTO  `json:"tax"`
	TradeDate       string    `json:"tradeDate"`
	SettlementDate  *string   `json:"settlementDate,omitempty"`
	SafeNote        *string   `json:"safeNote,omitempty"`
}

type importAppendResultDTO struct {
	PortfolioID             string   `json:"portfolioId"`
	SourceKind              string   `json:"sourceKind"`
	SourceFileHash          string   `json:"sourceFileHash"`
	ParsedRowCount          int      `json:"parsedRowCount"`
	AcceptedRowCount        int      `json:"acceptedRowCount"`
	NonAppendedRowCount     int      `json:"nonAppendedRowCount"`
	AppendedTransactionIDs  []string `json:"appendedTransactionIds"`
	SnapshotDatesRebuilt    []string `json:"snapshotDatesRebuilt"`
	AuditActionCode         string   `json:"auditActionCode"`
	NonSensitiveWarnings    []string `json:"nonSensitiveWarnings"`
	AppendValidationPolicy  string   `json:"appendValidationPolicy"`
	RawPayloadRetentionRule string   `json:"rawPayloadRetentionRule"`
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

func mapImportReview(review importer.Review) importReviewDTO {
	rows := make([]importRowReviewDTO, 0, len(review.Rows))
	for _, row := range review.Rows {
		rows = append(rows, mapImportRowReview(row))
	}
	return importReviewDTO{
		PortfolioID:        review.PortfolioID,
		SourceKind:         review.SourceKind,
		SourceAccountLabel: review.SourceAccountLabel,
		SourceFileHash:     review.FileHash,
		RetentionPolicy:    "TRANSIENT_NOT_STORED",
		ReviewGuarantee:    "PREFLIGHT_ONLY_APPEND_RERUNS_REVIEW_AND_STORE_CHECKS",
		Summary: importSummaryDTO{
			TotalRows:      review.Summary.TotalRows,
			AppendableRows: review.Summary.AppendableRows,
			DuplicateRows:  review.Summary.DuplicateRows,
			ConflictRows:   review.Summary.ConflictRows,
			InvalidRows:    review.Summary.InvalidRows,
		},
		Rows: rows,
	}
}

func mapImportRowReview(row importer.RowReview) importRowReviewDTO {
	return importRowReviewDTO{
		RowNumber:   row.RowNumber,
		RowHash:     row.RowHash,
		Status:      row.Status,
		ReasonCodes: row.ReasonCodes,
		Fingerprint: row.Fingerprint,
		Candidate:   mapImportCandidate(row.Candidate),
	}
}

func mapImportCandidate(candidate *importer.Candidate) *importCandidateDTO {
	if candidate == nil {
		return nil
	}
	var quantity *string
	if candidate.Quantity != nil {
		value := candidate.Quantity.String()
		quantity = &value
	}
	return &importCandidateDTO{
		TransactionType: candidate.TransactionType,
		Ticker:          candidate.Ticker,
		Quantity:        quantity,
		UnitPrice:       mapOptionalMoney(candidate.UnitPrice),
		GrossAmount:     mapMoney(candidate.GrossAmount),
		Commission:      mapMoney(candidate.Commission),
		Tax:             mapMoney(candidate.Tax),
		TradeDate:       candidate.TradeDate,
		SettlementDate:  candidate.SettlementDate,
		SafeNote:        candidate.SafeNote,
	}
}

func mapImportAppendResult(portfolioID string, sourceKind string, fileHash string, result importflow.Result) importAppendResultDTO {
	return importAppendResultDTO{
		PortfolioID:             portfolioID,
		SourceKind:              sourceKind,
		SourceFileHash:          fileHash,
		ParsedRowCount:          result.ParsedRowCount,
		AcceptedRowCount:        result.AcceptedRowCount,
		NonAppendedRowCount:     result.NonAppendedRowCount,
		AppendedTransactionIDs:  result.AppendedTransactionIDs,
		SnapshotDatesRebuilt:    result.SnapshotDatesRebuilt,
		AuditActionCode:         result.AuditActionCode,
		NonSensitiveWarnings:    result.NonSensitiveWarnings,
		AppendValidationPolicy:  "REVIEW_RERUN_AND_ATOMIC_STORE_REVALIDATION",
		RawPayloadRetentionRule: "RAW_CSV_NOT_STORED",
	}
}

func importPayloadHash(payload string) string {
	hash := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(hash[:])
}

func validateImportRowCount(totalRows int) error {
	if totalRows > maxHTTPImportRows {
		return fmt.Errorf("%w: import CSV must contain at most %d data rows", verticalslice.ErrInvalidInput, maxHTTPImportRows)
	}
	return nil
}
