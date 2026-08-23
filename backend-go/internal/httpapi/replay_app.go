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
	secret, err := normalizedImportReviewSecret(importReviewTokenSecret)
	if err != nil {
		return nil, err
	}
	return newReplayApp(&API{
		service:                service,
		auth:                   authService,
		authLimiter:            newAuthRateLimiter(20, time.Minute),
		importReviewSecret:     secret,
		paginationCursorSecret: derivePaginationCursorSecret(secret),
	}), nil
}

// NewDevelopmentReplay provides the same Stage 3.32 route wiring with the existing explicit
// development subject bypass. It is never selected by production safety checks.
func NewDevelopmentReplay(service *verticalslice.Service) *fiber.App {
	secret, err := normalizedImportReviewSecret([]byte("openinvest-development-import-review-token-secret"))
	if err != nil {
		panic(err)
	}
	return newReplayApp(&API{
		service:                 service,
		allowDevelopmentSubject: true,
		authLimiter:             newAuthRateLimiter(20, time.Minute),
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
	app.Get("/api/v1/portfolios", api.listPortfolios)
	app.Post("/api/v1/portfolios", api.createPortfolioReplay)
	app.Get("/api/v1/portfolios/:portfolioId", api.getPortfolio)
	app.Get("/api/v1/portfolios/:portfolioId/summary", api.getPortfolioSummary)
	app.Get("/api/v1/portfolios/:portfolioId/transactions", api.listTransactions)
	app.Post("/api/v1/portfolios/:portfolioId/transactions", api.appendTransactionReplay)
	app.Post("/api/v1/portfolios/:portfolioId/imports/review", api.reviewImport)
	app.Post("/api/v1/portfolios/:portfolioId/imports/append", api.appendImportReplay)

	return app
}
