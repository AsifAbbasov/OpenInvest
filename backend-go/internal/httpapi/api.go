package httpapi

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/auth"
	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/importer"
	"github.com/openinvest/openinvest/backend-go/internal/importflow"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

const devSubjectID = "00000000-0000-4000-8000-000000000001"
const maxHTTPImportPayloadBytes = 2 * 1024 * 1024
const maxHTTPImportRows = 100
const authRateLimitRetryAfterSeconds = "60"
const defaultAuthRateLimiterMaxKeys = 2048
const defaultAuthRateLimiterGlobalLimit = 2000
const minImportReviewTokenSecretBytes = 32
const maxPaginationCursorBytes = 512
const maxImportReviewTokenBytes = 16384
const importReviewTokenVersion = 1
const importReviewTokenTTL = 15 * time.Minute

var errAuthRateLimited = errors.New("auth rate limited")

type API struct {
	service                 *verticalslice.Service
	auth                    *auth.Service
	allowDevelopmentSubject bool
	authLimiter             *authRateLimiter
	importReviewSecret      []byte
	paginationCursorSecret  []byte
	now                     func() time.Time
}

func New(service *verticalslice.Service, authService *auth.Service, importReviewTokenSecret []byte) (*fiber.App, error) {
	secret, err := normalizedImportReviewSecret(importReviewTokenSecret)
	if err != nil {
		return nil, err
	}
	return newApp(&API{
		service:                service,
		auth:                   authService,
		authLimiter:            newAuthRateLimiter(20, time.Minute),
		importReviewSecret:     secret,
		paginationCursorSecret: derivePaginationCursorSecret(secret),
	}), nil
}

func NewDevelopment(service *verticalslice.Service) *fiber.App {
	secret, err := normalizedImportReviewSecret([]byte("openinvest-development-import-review-token-secret"))
	if err != nil {
		panic(err)
	}
	return newApp(&API{
		service:                 service,
		allowDevelopmentSubject: true,
		authLimiter:             newAuthRateLimiter(20, time.Minute),
		importReviewSecret:      secret,
		paginationCursorSecret:  derivePaginationCursorSecret(secret),
	})
}

func newApp(api *API) *fiber.App {
	app := fiber.New(fiber.Config{AppName: "OpenInvest API"})

	app.Use(localDevelopmentCORS)

	app.Get("/api/v1/health", api.health)
	app.Get("/api/v1/ready", api.ready)
	app.Post("/api/v1/auth/register", api.register)
	app.Post("/api/v1/auth/login", api.login)
	app.Post("/api/v1/auth/refresh", api.refresh)
	app.Post("/api/v1/auth/logout", api.logout)
	app.Get("/api/v1/assets/search", api.searchAssets)
	app.Get("/api/v1/assets/:ticker", api.getAsset)
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
		c.Set("Access-Control-Allow-Credentials", "true")
		c.Set("Vary", "Origin")
		c.Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, Idempotency-Key, X-CSRF-Token, X-Request-ID, traceparent")
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

type authRateLimiter struct {
	mu             sync.Mutex
	limit          int
	globalLimit    int
	maxKeys        int
	window         time.Duration
	attempts       map[string][]time.Time
	globalAttempts []time.Time
	lastSweep      time.Time
}

func newAuthRateLimiter(limit int, window time.Duration) *authRateLimiter {
	return newBoundedAuthRateLimiter(limit, defaultAuthRateLimiterGlobalLimit, defaultAuthRateLimiterMaxKeys, window)
}

func newBoundedAuthRateLimiter(limit int, globalLimit int, maxKeys int, window time.Duration) *authRateLimiter {
	return &authRateLimiter{
		limit:       nonNegativeInt(limit),
		globalLimit: nonNegativeInt(globalLimit),
		maxKeys:     nonNegativeInt(maxKeys),
		window:      window,
		attempts:    map[string][]time.Time{},
	}
}

func (limiter *authRateLimiter) allow(key string, now time.Time) bool {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	if limiter.window <= 0 || limiter.limit == 0 || limiter.globalLimit == 0 || limiter.maxKeys == 0 {
		return false
	}

	cutoff := now.Add(-limiter.window)
	limiter.globalAttempts = retainAuthAttempts(limiter.globalAttempts, cutoff)

	attempts, exists := limiter.attempts[key]
	if exists {
		attempts = retainAuthAttempts(attempts, cutoff)
		if len(attempts) == 0 {
			delete(limiter.attempts, key)
			exists = false
		} else {
			limiter.attempts[key] = attempts
		}
	}

	if !exists && (len(limiter.attempts) >= limiter.maxKeys || limiter.shouldSweep(now)) {
		limiter.sweepExpired(cutoff, now)
		attempts = limiter.attempts[key]
		exists = len(attempts) > 0
	}

	if len(limiter.globalAttempts) >= limiter.globalLimit {
		return false
	}
	if !exists && len(limiter.attempts) >= limiter.maxKeys {
		return false
	}
	if len(attempts) >= limiter.limit {
		return false
	}

	attempts = append(attempts, now)
	limiter.attempts[key] = attempts
	limiter.globalAttempts = append(limiter.globalAttempts, now)
	return true
}

func (limiter *authRateLimiter) shouldSweep(now time.Time) bool {
	return limiter.lastSweep.IsZero() || !now.Before(limiter.lastSweep.Add(limiter.window))
}

func (limiter *authRateLimiter) sweepExpired(cutoff time.Time, now time.Time) {
	for key, attempts := range limiter.attempts {
		retained := retainAuthAttempts(attempts, cutoff)
		if len(retained) == 0 {
			delete(limiter.attempts, key)
			continue
		}
		limiter.attempts[key] = retained
	}
	limiter.lastSweep = now
}

func retainAuthAttempts(attempts []time.Time, cutoff time.Time) []time.Time {
	retained := attempts[:0]
	for _, attempt := range attempts {
		if attempt.After(cutoff) {
			retained = append(retained, attempt)
		}
	}
	return retained
}

func nonNegativeInt(value int) int {
	if value < 0 {
		return 0
	}
	return value
}

func (api *API) checkAuthRateLimit(c fiber.Ctx) error {
	if api.authLimiter == nil {
		return nil
	}
	key := c.Path() + "|" + c.IP()
	if !api.authLimiter.allow(key, api.nowUTC()) {
		return errAuthRateLimited
	}
	return nil
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

func (api *API) register(c fiber.Ctx) error {
	meta := requestMeta(c)
	if err := api.checkAuthRateLimit(c); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	if api.auth == nil {
		return writeErrorWithMeta(c, meta, http.StatusServiceUnavailable, "SERVICE_NOT_READY", "Authentication service is not ready")
	}
	var request registerRequestDTO
	if err := decodeStrictJSON(c.Request().Body(), &request); err != nil {
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
	}
	result, err := api.auth.Register(c.Context(), auth.RegistrationRequest{
		Email:    request.Email,
		Password: request.Password,
		Language: request.Language,
		Theme:    request.Theme,
		Timezone: request.Timezone,
	})
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	api.setRefreshCookie(c, result.RefreshToken)
	return writeStatusWithMeta(c, meta, http.StatusCreated, mapAuthData(result))
}

func (api *API) login(c fiber.Ctx) error {
	meta := requestMeta(c)
	if err := api.checkAuthRateLimit(c); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	if api.auth == nil {
		return writeErrorWithMeta(c, meta, http.StatusServiceUnavailable, "SERVICE_NOT_READY", "Authentication service is not ready")
	}
	var request loginRequestDTO
	if err := decodeStrictJSON(c.Request().Body(), &request); err != nil {
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
	}
	result, err := api.auth.Login(c.Context(), auth.LoginRequest{Email: request.Email, Password: request.Password})
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	api.setRefreshCookie(c, result.RefreshToken)
	return writeStatusWithMeta(c, meta, http.StatusOK, mapAuthData(result))
}

func (api *API) refresh(c fiber.Ctx) error {
	meta := requestMeta(c)
	if err := api.checkAuthRateLimit(c); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	if api.auth == nil {
		return writeErrorWithMeta(c, meta, http.StatusServiceUnavailable, "SERVICE_NOT_READY", "Authentication service is not ready")
	}
	result, err := api.auth.Refresh(c.Context(), c.Cookies(auth.RefreshCookieName), c.Get("X-CSRF-Token"))
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	api.setRefreshCookie(c, result.RefreshToken)
	return writeStatusWithMeta(c, meta, http.StatusOK, mapAuthSession(result.Session))
}

func (api *API) logout(c fiber.Ctx) error {
	meta := requestMeta(c)
	if err := api.checkAuthRateLimit(c); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	if api.auth == nil {
		return writeErrorWithMeta(c, meta, http.StatusServiceUnavailable, "SERVICE_NOT_READY", "Authentication service is not ready")
	}
	var request logoutRequestDTO
	if len(bytes.TrimSpace(c.Request().Body())) == 0 {
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "allSessions is required")
	}
	if err := decodeStrictJSON(c.Request().Body(), &request); err != nil {
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
	}
	if request.AllSessions == nil {
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "allSessions is required")
	}
	revoked, err := api.auth.Logout(c.Context(), c.Cookies(auth.RefreshCookieName), c.Get("X-CSRF-Token"), *request.AllSessions)
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	api.clearRefreshCookie(c)
	return writeStatusWithMeta(c, meta, http.StatusOK, logoutResultDTO{Revoked: revoked})
}

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
	query := strings.TrimSpace(c.Query("query"))
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

func (api *API) reviewImport(c fiber.Ctx) error {
	meta := requestMeta(c)
	subjectID, err := api.subjectID(c)
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	var request importReviewRequestDTO
	if err := decodeStrictJSON(c.Request().Body(), &request); err != nil {
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
	}
	if err := request.validate(); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	portfolioID := c.Params("portfolioId")
	fileHash := importPayloadHash(request.CSVPayload)

	parserReview, err := importer.ReviewCSV(importer.ReviewRequest{
		SubjectID:          subjectID,
		PortfolioID:        portfolioID,
		SourceKind:         importer.SourceKindUserUploadedFile,
		SourceAccountLabel: request.SourceAccountLabel,
		FileHash:           fileHash,
		Reader:             strings.NewReader(request.CSVPayload),
	})
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	if err := validateImportRowCount(parserReview.Summary.TotalRows); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}

	existing, err := api.service.ListImportReviewTransactions(
		c.Context(),
		subjectID,
		portfolioID,
		importer.ReviewHistoryFilter(parserReview),
	)
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}

	review, err := importer.ReviewCSV(importer.ReviewRequest{
		SubjectID:          subjectID,
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

	reviewToken, err := api.signImportReviewToken(subjectID, parserReview, review)
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	return writeStatusWithMeta(c, meta, http.StatusOK, mapImportReview(review, reviewToken))
}

func (api *API) appendImport(c fiber.Ctx) error {
	meta := requestMeta(c)
	subjectID, err := api.subjectID(c)
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	if err := verticalslice.ValidateIdempotencyKey(c.Get("Idempotency-Key")); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	var request importAppendRequestDTO
	if err := decodeStrictJSON(c.Request().Body(), &request); err != nil {
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
	}
	if err := request.validate(); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	portfolioID := c.Params("portfolioId")
	fileHash := importPayloadHash(request.CSVPayload)
	if request.SourceFileHash != fileHash {
		return writeMappedErrorWithMeta(c, meta, fmt.Errorf("%w: sourceFileHash does not match import payload", importer.ErrUnsafeAppend))
	}
	preflightReview, err := importer.ReviewCSV(importer.ReviewRequest{
		SubjectID:          subjectID,
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
	if err := api.verifyImportReviewToken(
		request.ReviewToken,
		subjectID,
		portfolioID,
		importer.SourceKindUserUploadedFile,
		request.SourceAccountLabel,
		fileHash,
		preflightReview,
		request.toAppDecisions(),
	); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	if err := importer.VerifyDecisionIdentities(preflightReview, request.toAppDecisionIdentities()); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	result, err := importflow.ReviewAndAppend(c.Context(), api.service, importflow.Request{
		RequestContext:     meta.toApp(),
		SubjectID:          subjectID,
		PortfolioID:        portfolioID,
		IdempotencyKey:     c.Get("Idempotency-Key"),
		RequestPath:        c.Path(),
		SourceKind:         importer.SourceKindUserUploadedFile,
		SourceAccountLabel: request.SourceAccountLabel,
		SourceFileHash:     fileHash,
		// The store performs the atomic duplicate check after reserving the command. Passing the
		// mutable current ledger here would reject an idempotent retry before that reservation.
		Existing:  nil,
		Reader:    strings.NewReader(request.CSVPayload),
		Decisions: request.toAppDecisions(),
	})
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	return writeStatusWithMeta(c, meta, http.StatusCreated, mapImportAppendResult(portfolioID, importer.SourceKindUserUploadedFile, fileHash, result))
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

func (api *API) subjectID(c fiber.Ctx) (string, error) {
	if api.auth != nil {
		if token := bearerToken(c.Get("Authorization")); token != "" {
			return api.auth.AuthenticateAccessToken(token)
		}
		if api.auth.AllowsDevelopmentBypass() {
			return developmentSubjectID(), nil
		}
		return "", auth.ErrInvalidSession
	}
	if api.allowDevelopmentSubject {
		return developmentSubjectID(), nil
	}
	return "", auth.ErrInvalidSession
}

func developmentSubjectID() string {
	if configured := strings.TrimSpace(os.Getenv("OPENINVEST_DEV_SUBJECT_ID")); configured != "" {
		return configured
	}
	return devSubjectID
}

func bearerToken(header string) string {
	prefix := "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}

func (api *API) setRefreshCookie(c fiber.Ctx, value string) {
	c.Cookie(&fiber.Cookie{
		Name:     auth.RefreshCookieName,
		Value:    value,
		Path:     "/api/v1/auth",
		MaxAge:   api.auth.RefreshCookieMaxAgeSeconds(),
		SameSite: "Strict",
		Secure:   api.auth.RefreshCookieSecure(),
		HTTPOnly: true,
	})
}

func (api *API) clearRefreshCookie(c fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     auth.RefreshCookieName,
		Value:    "",
		Path:     "/api/v1/auth",
		MaxAge:   -1,
		SameSite: "Strict",
		Secure:   api.auth.RefreshCookieSecure(),
		HTTPOnly: true,
	})
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

func queryLimitStrict(c fiber.Ctx, fallback int) (int, error) {
	raw, present := queryValue(c, "limit")
	if !present {
		return fallback, nil
	}
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0, fmt.Errorf("%w: limit must be an integer between 1 and 100", verticalslice.ErrInvalidInput)
	}
	value, err := strconv.Atoi(trimmed)
	if err != nil || value < 1 || value > 100 {
		return 0, fmt.Errorf("%w: limit must be an integer between 1 and 100", verticalslice.ErrInvalidInput)
	}
	return value, nil
}

func optionalQueryValue(c fiber.Ctx, name string) (string, error) {
	raw, present := queryValue(c, name)
	if !present {
		return "", nil
	}
	if strings.TrimSpace(raw) == "" {
		return "", fmt.Errorf("%w: %s must not be empty when supplied", verticalslice.ErrInvalidInput, name)
	}
	return raw, nil
}

func queryValue(c fiber.Ctx, name string) (string, bool) {
	args := c.Request().URI().QueryArgs()
	if !args.Has(name) {
		return "", false
	}
	return string(args.Peek(name)), true
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
	case errors.Is(err, errAuthRateLimited):
		c.Set("Retry-After", authRateLimitRetryAfterSeconds)
		return writeErrorWithMeta(c, meta, http.StatusTooManyRequests, "RATE_LIMITED", "Too many authentication attempts")
	case errors.Is(err, auth.ErrInvalidInput):
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, auth.ErrEmailAlreadyExists):
		return writeErrorWithMeta(c, meta, http.StatusConflict, "CONFLICT", "Email is already registered")
	case errors.Is(err, auth.ErrInvalidCredentials):
		return writeErrorWithMeta(c, meta, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid email or password")
	case errors.Is(err, auth.ErrInvalidSession):
		return writeErrorWithMeta(c, meta, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired session")
	case errors.Is(err, auth.ErrInvalidCSRF):
		return writeErrorWithMeta(c, meta, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid CSRF token")
	case errors.Is(err, verticalslice.ErrInvalidInput):
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, verticalslice.ErrMissingIdempotency):
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "Idempotency-Key header is required")
	case errors.Is(err, verticalslice.ErrNotFound):
		return writeErrorWithMeta(c, meta, http.StatusNotFound, "NOT_FOUND", "Resource not found")
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

func decodeStrictJSON(body []byte, target any) error {
	if len(bytes.TrimSpace(body)) == 0 {
		return io.EOF
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("unexpected trailing JSON")
	}
	return nil
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

type registerRequestDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Language string `json:"language"`
	Theme    string `json:"theme"`
	Timezone string `json:"timezone"`
}

type loginRequestDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type logoutRequestDTO struct {
	AllSessions *bool `json:"allSessions"`
}

type userDTO struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	Language    string `json:"language"`
	Theme       string `json:"theme"`
	Timezone    string `json:"timezone"`
	PrivacyMode bool   `json:"privacyMode"`
	CreatedAt   string `json:"createdAt"`
}

type authSessionDTO struct {
	AccessToken          string `json:"accessToken"`
	AccessTokenExpiresAt string `json:"accessTokenExpiresAt"`
	CSRFToken            string `json:"csrfToken"`
}

type authDataDTO struct {
	User    userDTO        `json:"user"`
	Session authSessionDTO `json:"session"`
}

type logoutResultDTO struct {
	Revoked bool `json:"revoked"`
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

type assetSummaryDTO struct {
	Ticker    string    `json:"ticker"`
	Name      string    `json:"name"`
	AssetType string    `json:"assetType"`
	Currency  string    `json:"currency"`
	LotSize   string    `json:"lotSize"`
	LastPrice *moneyDTO `json:"lastPrice"`
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
	SourceFileHash     string              `json:"sourceFileHash"`
	ReviewToken        string              `json:"reviewToken"`
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
	if !validSHA256Hex(request.SourceFileHash) {
		return fmt.Errorf("%w: sourceFileHash must be a SHA-256 hex digest from the review response", verticalslice.ErrInvalidInput)
	}
	if len(request.ReviewToken) < 32 || len(request.ReviewToken) > maxImportReviewTokenBytes {
		return fmt.Errorf("%w: reviewToken must be between 32 and %d bytes", verticalslice.ErrInvalidInput, maxImportReviewTokenBytes)
	}
	for _, decision := range request.Decisions {
		if !validSHA256Hex(decision.RowHash) {
			return fmt.Errorf("%w: rowHash must be a SHA-256 hex digest from the review response", verticalslice.ErrInvalidInput)
		}
	}
	return nil
}

func (request importAppendRequestDTO) toAppDecisions() []importer.Decision {
	decisions := make([]importer.Decision, 0, len(request.Decisions))
	for _, decision := range request.Decisions {
		decisions = append(decisions, importer.Decision{RowNumber: decision.RowNumber, RowHash: decision.RowHash, Action: decision.Action})
	}
	return decisions
}

func (request importAppendRequestDTO) toAppDecisionIdentities() []importer.DecisionIdentity {
	identities := make([]importer.DecisionIdentity, 0, len(request.Decisions))
	for _, decision := range request.Decisions {
		identities = append(identities, importer.DecisionIdentity{RowNumber: decision.RowNumber, RowHash: decision.RowHash})
	}
	return identities
}

type importDecisionDTO struct {
	RowNumber int    `json:"rowNumber"`
	RowHash   string `json:"rowHash"`
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
		return nil, fmt.Errorf("%w: decimal value is invalid", verticalslice.ErrInvalidInput)
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
		return verticalslice.Money{}, fmt.Errorf("%w: money amount is invalid", verticalslice.ErrInvalidInput)
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

type importReviewDTO struct {
	PortfolioID        string               `json:"portfolioId"`
	SourceKind         string               `json:"sourceKind"`
	SourceAccountLabel string               `json:"sourceAccountLabel"`
	SourceFileHash     string               `json:"sourceFileHash"`
	ReviewToken        string               `json:"reviewToken"`
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

func mapAuthData(result auth.AuthResult) authDataDTO {
	return authDataDTO{User: mapUser(result.User), Session: mapAuthSession(result.Session)}
}

func mapUser(user auth.StoredUser) userDTO {
	return userDTO{
		ID:          user.ID,
		Email:       user.Email,
		Language:    user.Language,
		Theme:       user.Theme,
		Timezone:    user.Timezone,
		PrivacyMode: user.PrivacyMode,
		CreatedAt:   user.CreatedAt.UTC().Format(time.RFC3339),
	}
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

func mapAuthSession(session auth.ClientSession) authSessionDTO {
	return authSessionDTO{
		AccessToken:          session.AccessToken,
		AccessTokenExpiresAt: session.AccessTokenExpiresAt.UTC().Format(time.RFC3339),
		CSRFToken:            session.CSRFToken,
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

func mapImportReview(review importer.Review, reviewToken string) importReviewDTO {
	rows := make([]importRowReviewDTO, 0, len(review.Rows))
	for _, row := range review.Rows {
		rows = append(rows, mapImportRowReview(row))
	}
	return importReviewDTO{
		PortfolioID:        review.PortfolioID,
		SourceKind:         review.SourceKind,
		SourceAccountLabel: review.SourceAccountLabel,
		SourceFileHash:     review.FileHash,
		ReviewToken:        reviewToken,
		RetentionPolicy:    "TRANSIENT_NOT_STORED",
		ReviewGuarantee:    "SIGNED_REVIEW_TOKEN_APPEND_RERUNS_REVIEW_AND_STORE_CHECKS",
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

type importReviewTokenRowIdentity struct {
	RowNumber int    `json:"n"`
	RowHash   string `json:"h"`
}

type importReviewTokenPayload struct {
	Version            int                            `json:"v"`
	ParserVersion      int                            `json:"pv"`
	IssuedAt           int64                          `json:"iat"`
	ExpiresAt          int64                          `json:"exp"`
	SubjectID          string                         `json:"s"`
	PortfolioID        string                         `json:"p"`
	SourceKind         string                         `json:"k"`
	SourceAccountLabel string                         `json:"l"`
	SourceFileHash     string                         `json:"f"`
	ParserReviewDigest string                         `json:"pd"`
	FinalReviewDigest  string                         `json:"fd"`
	AppendableRows     []int                          `json:"a"`
	Rows               []importReviewTokenRowIdentity `json:"r"`
}

func normalizedImportReviewSecret(configured []byte) ([]byte, error) {
	secret := bytes.TrimSpace(configured)
	if len(secret) < minImportReviewTokenSecretBytes {
		return nil, fmt.Errorf("import review token secret must be at least %d bytes", minImportReviewTokenSecretBytes)
	}
	hash := sha256.Sum256(secret)
	return hash[:], nil
}

func (api *API) nowUTC() time.Time {
	if api.now != nil {
		return api.now().UTC()
	}
	return time.Now().UTC()
}

func (api *API) signImportReviewToken(subjectID string, parserReview importer.Review, finalReview importer.Review) (string, error) {
	if parserReview.PortfolioID != finalReview.PortfolioID ||
		parserReview.SourceKind != finalReview.SourceKind ||
		parserReview.SourceAccountLabel != finalReview.SourceAccountLabel ||
		!strings.EqualFold(parserReview.FileHash, finalReview.FileHash) ||
		len(parserReview.Rows) != len(finalReview.Rows) {
		return "", fmt.Errorf("import review phases do not share the same source context")
	}

	parserDigest, err := importer.ReviewSemanticDigest(parserReview)
	if err != nil {
		return "", err
	}
	finalDigest, err := importer.ReviewSemanticDigest(finalReview)
	if err != nil {
		return "", err
	}

	rows := make([]importReviewTokenRowIdentity, 0, len(parserReview.Rows))
	appendableRows := []int{}
	for index, parserRow := range parserReview.Rows {
		finalRow := finalReview.Rows[index]
		if parserRow.RowNumber != finalRow.RowNumber || parserRow.RowHash != finalRow.RowHash {
			return "", fmt.Errorf("import review phases do not share row identity")
		}
		rows = append(rows, importReviewTokenRowIdentity{
			RowNumber: parserRow.RowNumber,
			RowHash:   parserRow.RowHash,
		})
		if finalRow.Status == importer.ReviewStatusAppendable {
			appendableRows = append(appendableRows, finalRow.RowNumber)
		}
	}

	issuedAt := api.nowUTC()
	payload := importReviewTokenPayload{
		Version:            importReviewTokenVersion,
		ParserVersion:      importer.ReviewParserVersion,
		IssuedAt:           issuedAt.Unix(),
		ExpiresAt:          issuedAt.Add(importReviewTokenTTL).Unix(),
		SubjectID:          subjectID,
		PortfolioID:        finalReview.PortfolioID,
		SourceKind:         finalReview.SourceKind,
		SourceAccountLabel: finalReview.SourceAccountLabel,
		SourceFileHash:     strings.ToLower(finalReview.FileHash),
		ParserReviewDigest: parserDigest,
		FinalReviewDigest:  finalDigest,
		AppendableRows:     appendableRows,
		Rows:               rows,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	bodyPart := base64.RawURLEncoding.EncodeToString(body)
	signature := api.signImportReviewTokenPart(bodyPart)
	token := bodyPart + "." + signature
	if len(token) > maxImportReviewTokenBytes {
		return "", fmt.Errorf("import review token exceeds %d bytes", maxImportReviewTokenBytes)
	}
	return token, nil
}

func (api *API) verifyImportReviewToken(
	token string,
	subjectID string,
	portfolioID string,
	sourceKind string,
	sourceAccountLabel string,
	sourceFileHash string,
	parserReview importer.Review,
	decisions []importer.Decision,
) error {
	token = strings.TrimSpace(token)
	if len(token) == 0 || len(token) > maxImportReviewTokenBytes {
		return fmt.Errorf("%w: reviewToken is invalid", verticalslice.ErrInvalidInput)
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("%w: reviewToken is invalid", verticalslice.ErrInvalidInput)
	}
	expectedSignature := api.signImportReviewTokenPart(parts[0])
	if !hmac.Equal([]byte(parts[1]), []byte(expectedSignature)) {
		return fmt.Errorf("%w: reviewToken signature is invalid", importer.ErrUnsafeAppend)
	}
	body, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return fmt.Errorf("%w: reviewToken payload is invalid", verticalslice.ErrInvalidInput)
	}
	var payload importReviewTokenPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("%w: reviewToken payload is invalid", verticalslice.ErrInvalidInput)
	}

	if payload.Version != importReviewTokenVersion ||
		payload.ParserVersion != importer.ReviewParserVersion ||
		payload.ExpiresAt-payload.IssuedAt != int64(importReviewTokenTTL/time.Second) ||
		payload.IssuedAt <= 0 {
		return fmt.Errorf("%w: reviewToken version or lifetime is invalid", importer.ErrUnsafeAppend)
	}
	now := api.nowUTC()
	issuedAt := time.Unix(payload.IssuedAt, 0).UTC()
	expiresAt := time.Unix(payload.ExpiresAt, 0).UTC()
	if now.Before(issuedAt.Add(-time.Minute)) || !now.Before(expiresAt) {
		return fmt.Errorf("%w: reviewToken is expired or not yet valid", importer.ErrUnsafeAppend)
	}

	if payload.SubjectID != subjectID ||
		payload.PortfolioID != portfolioID ||
		payload.SourceKind != sourceKind ||
		payload.SourceAccountLabel != strings.TrimSpace(sourceAccountLabel) ||
		payload.SourceFileHash != sourceFileHash {
		return fmt.Errorf("%w: reviewToken does not match append context", importer.ErrUnsafeAppend)
	}
	if !validSHA256Hex(payload.ParserReviewDigest) || !validSHA256Hex(payload.FinalReviewDigest) {
		return fmt.Errorf("%w: reviewToken semantic digest is invalid", importer.ErrUnsafeAppend)
	}

	parserDigest, err := importer.ReviewSemanticDigest(parserReview)
	if err != nil {
		return err
	}
	if parserDigest != payload.ParserReviewDigest {
		return fmt.Errorf("%w: normalized import semantics changed; review again", importer.ErrUnsafeAppend)
	}
	if len(payload.Rows) != len(parserReview.Rows) {
		return fmt.Errorf("%w: reviewToken row set does not match parser review", importer.ErrUnsafeAppend)
	}

	tokenRows := map[int]string{}
	for _, row := range payload.Rows {
		if row.RowNumber < 2 || !validSHA256Hex(row.RowHash) {
			return fmt.Errorf("%w: reviewToken row identity is invalid", importer.ErrUnsafeAppend)
		}
		if _, exists := tokenRows[row.RowNumber]; exists {
			return fmt.Errorf("%w: reviewToken contains duplicate row identity", importer.ErrUnsafeAppend)
		}
		tokenRows[row.RowNumber] = row.RowHash
	}
	for _, row := range parserReview.Rows {
		if tokenRows[row.RowNumber] != row.RowHash {
			return fmt.Errorf("%w: reviewToken row set does not match parser review", importer.ErrUnsafeAppend)
		}
	}

	appendableRows := map[int]struct{}{}
	for _, rowNumber := range payload.AppendableRows {
		if _, exists := appendableRows[rowNumber]; exists {
			return fmt.Errorf("%w: reviewToken contains duplicate appendable row", importer.ErrUnsafeAppend)
		}
		if _, exists := tokenRows[rowNumber]; !exists {
			return fmt.Errorf("%w: reviewToken appendable row is not in reviewed rows", importer.ErrUnsafeAppend)
		}
		appendableRows[rowNumber] = struct{}{}
	}

	for _, decision := range decisions {
		if tokenRows[decision.RowNumber] != decision.RowHash {
			return fmt.Errorf("%w: reviewToken row identity does not match decision", importer.ErrUnsafeAppend)
		}
		if decision.Action == importer.DecisionApprove {
			if _, ok := appendableRows[decision.RowNumber]; !ok {
				return fmt.Errorf("%w: row %d was not appendable in the approved review", importer.ErrUnsafeAppend, decision.RowNumber)
			}
		}
	}
	return nil
}

func (api *API) signImportReviewTokenPart(bodyPart string) string {
	mac := hmac.New(sha256.New, api.importReviewSecret)
	_, _ = mac.Write([]byte(bodyPart))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func validSHA256Hex(value string) bool {
	if len(value) != sha256.Size*2 || value != strings.ToLower(value) {
		return false
	}
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == sha256.Size
}

type paginationCursorPayload struct {
	Version         int    `json:"v"`
	Resource        string `json:"resource"`
	SubjectID       string `json:"subjectId"`
	PortfolioID     string `json:"portfolioId,omitempty"`
	AssetType       string `json:"assetType,omitempty"`
	QueryHash       string `json:"queryHash,omitempty"`
	Ticker          string `json:"ticker,omitempty"`
	TransactionType string `json:"transactionType,omitempty"`
	FromDate        string `json:"fromDate,omitempty"`
	ToDate          string `json:"toDate,omitempty"`
	UpdatedAt       string `json:"updatedAt,omitempty"`
	TradeDate       string `json:"tradeDate,omitempty"`
	EntryID         string `json:"entryId"`
}

func derivePaginationCursorSecret(importReviewSecret []byte) []byte {
	mac := hmac.New(sha256.New, importReviewSecret)
	_, _ = mac.Write([]byte("openinvest/pagination-cursor/v1"))
	return mac.Sum(nil)
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

func (api *API) signPaginationCursor(payload paginationCursorPayload) (string, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	bodyPart := base64.RawURLEncoding.EncodeToString(body)
	mac := hmac.New(sha256.New, api.paginationCursorSecret)
	_, _ = mac.Write([]byte(bodyPart))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return bodyPart + "." + signature, nil
}

func (api *API) verifyPaginationCursor(raw string) (paginationCursorPayload, error) {
	if len(raw) == 0 || len(raw) > maxPaginationCursorBytes {
		return paginationCursorPayload{}, invalidPaginationCursor()
	}
	parts := strings.Split(raw, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return paginationCursorPayload{}, invalidPaginationCursor()
	}
	mac := hmac.New(sha256.New, api.paginationCursorSecret)
	_, _ = mac.Write([]byte(parts[0]))
	expectedSignature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(parts[1]), []byte(expectedSignature)) {
		return paginationCursorPayload{}, invalidPaginationCursor()
	}
	body, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return paginationCursorPayload{}, invalidPaginationCursor()
	}
	var payload paginationCursorPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return paginationCursorPayload{}, invalidPaginationCursor()
	}
	return payload, nil
}

func invalidPaginationCursor() error {
	return fmt.Errorf("%w: cursor is invalid", verticalslice.ErrInvalidInput)
}

func validateImportRowCount(totalRows int) error {
	if totalRows > maxHTTPImportRows {
		return fmt.Errorf("%w: import CSV must contain at most %d data rows", verticalslice.ErrInvalidInput, maxHTTPImportRows)
	}
	return nil
}
