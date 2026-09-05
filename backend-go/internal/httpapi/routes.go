package httpapi

import (
	"github.com/gofiber/fiber/v3"
)

func registerRoutes(app *fiber.App, api *API) {
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
	app.Post("/api/v1/portfolios", api.createPortfolio)
	app.Get("/api/v1/portfolios/:portfolioId", api.getPortfolio)
	app.Get("/api/v1/portfolios/:portfolioId/summary", api.getPortfolioSummary)
	app.Get("/api/v1/portfolios/:portfolioId/transactions", api.listTransactions)
	app.Post("/api/v1/portfolios/:portfolioId/transactions", api.appendTransaction)
	app.Post("/api/v1/portfolios/:portfolioId/imports/review", api.reviewImport)
	app.Post("/api/v1/portfolios/:portfolioId/imports/append", api.appendImport)
}
