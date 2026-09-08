package verticalslice

import "context"

func (adapter *stage371StoreAdapter) GetPortfolioReturns(
	ctx context.Context,
	subjectID string,
	portfolioID string,
	asOfDate string,
) (PortfolioReturnProjection, error) {
	store, ok := adapter.Store.(PortfolioReturnStore)
	if !ok {
		return PortfolioReturnProjection{}, ErrPortfolioReturnProjectionUnavailable
	}
	return store.GetPortfolioReturns(ctx, subjectID, portfolioID, asOfDate)
}
