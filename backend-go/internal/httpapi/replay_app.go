package httpapi

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/openinvest/openinvest/backend-go/internal/auth"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

// NewReplay is the canonical Stage 3.32 production constructor. It preserves the existing HTTP
// surface while routing every currently implemented idempotent write through atomic exact replay.
func NewReplay(service *verticalslice.Service, authService *auth.Service, importReviewTokenSecret []byte) (*fiber.App, error) {
	return NewReplayWithCorporateActionProvider(service, authService, importReviewTokenSecret, nil)
}

// NewReplayWithCorporateActionProvider preserves the replay-safe production surface while
// explicitly injecting the provider-neutral Corporate Actions dependency. A nil provider keeps
// the route present but fail-closed with CORPORATE_ACTIONS_SOURCE_UNAVAILABLE.
func NewReplayWithCorporateActionProvider(
	service *verticalslice.Service,
	authService *auth.Service,
	importReviewTokenSecret []byte,
	corporateActionProvider verticalslice.CorporateActionProvider,
) (*fiber.App, error) {
	secret, err := normalizedImportReviewSecret(importReviewTokenSecret)
	if err != nil {
		return nil, err
	}
	return newReplayApp(&API{
		service:                 service,
		auth:                    authService,
		corporateActionProvider: corporateActionProvider,
		authLimiter:             newAuthRateLimiter(20, time.Minute),
		dividendLimiter:         newDividendCalculatorRateLimiter(),
		importReviewSecret:      secret,
		paginationCursorSecret:  derivePaginationCursorSecret(secret),
	}), nil
}

// NewDevelopmentReplay provides the same Stage 3.32 route wiring with the existing explicit
// development subject bypass. It is never selected by production safety checks.
func NewDevelopmentReplay(service *verticalslice.Service) *fiber.App {
	return NewDevelopmentReplayWithCorporateActionProvider(service, nil)
}

// NewDevelopmentReplayWithCorporateActionProvider keeps development composition explicit: the
// real provider is wired only when the same runtime activation gate used by production supplied it.
func NewDevelopmentReplayWithCorporateActionProvider(
	service *verticalslice.Service,
	corporateActionProvider verticalslice.CorporateActionProvider,
) *fiber.App {
	secret, err := normalizedImportReviewSecret([]byte("openinvest-development-import-review-token-secret"))
	if err != nil {
		panic(err)
	}
	return newReplayApp(&API{
		service:                 service,
		corporateActionProvider: corporateActionProvider,
		allowDevelopmentSubject: true,
		authLimiter:             newAuthRateLimiter(20, time.Minute),
		dividendLimiter:         newDividendCalculatorRateLimiter(),
		importReviewSecret:      secret,
		paginationCursorSecret:  derivePaginationCursorSecret(secret),
	})
}

func newReplayApp(api *API) *fiber.App {
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
	app.Get("/api/v1/corporate-actions/projection", api.getCorporateActionProjection)
	app.Post("/api/v1/dividends/calculate", api.calculateDividend)
	app.Get("/api/v1/portfolios", api.listPortfolios)
	app.Post("/api/v1/portfolios", api.createPortfolioReplay)
	app.Get("/api/v1/portfolios/:portfolioId", api.getPortfolio)
	app.Get("/api/v1/portfolios/:portfolioId/summary", api.getPortfolioSummary)
	app.Get("/api/v1/portfolios/:portfolioId/positions", api.getPortfolioPositions)
	app.Put("/api/v1/portfolios/:portfolioId/valuations/:ticker", api.upsertManualValuation)
	app.Delete("/api/v1/portfolios/:portfolioId/valuations/:ticker", api.clearManualValuation)
	app.Get("/api/v1/portfolios/:portfolioId/returns", api.getPortfolioReturns)
	app.Get("/api/v1/portfolios/:portfolioId/cash-flow", api.getPortfolioCashFlow)
	app.Get("/api/v1/portfolios/:portfolioId/transactions", api.listTransactions)
	app.Post("/api/v1/portfolios/:portfolioId/transactions", api.appendTransactionReplay)
	app.Patch("/api/v1/portfolios/:portfolioId/transactions/:transactionId", api.correctTransactionReplay)
	app.Delete("/api/v1/portfolios/:portfolioId/transactions/:transactionId", api.reverseTransactionReplay)
	app.Post("/api/v1/portfolios/:portfolioId/imports/review", api.reviewImport)
	app.Post("/api/v1/portfolios/:portfolioId/imports/append", api.appendImportReplaySafe)

	return app
}
