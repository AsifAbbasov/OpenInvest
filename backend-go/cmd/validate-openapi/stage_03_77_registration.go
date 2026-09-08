package main

func init() {
	requiredOperations["GET /api/v1/portfolios/{portfolioId}/returns"] = "getPortfolioReturns"
	requiredSchemas = append(requiredSchemas,
		"PortfolioReturnCashFlow",
		"PortfolioReturnCalculation",
		"PortfolioReturnAvailable",
		"PortfolioReturnUnavailableReason",
		"PortfolioReturnUnavailable",
		"PortfolioReturnProjection",
		"PortfolioReturnResponse",
	)
}
