package main

func init() {
	requiredOperations["PUT /api/v1/portfolios/{portfolioId}/valuations/{ticker}"] = "upsertManualValuation"
	requiredOperations["DELETE /api/v1/portfolios/{portfolioId}/valuations/{ticker}"] = "clearManualValuation"
	requiredSchemas = append(requiredSchemas,
		"ManualValuationRequest",
		"MarketValuationAvailable",
		"MarketValuation",
		"PortfolioValuationSummary",
	)
}
