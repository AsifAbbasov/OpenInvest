package verticalslice

// NewServiceWithQuoteProvider provides the immutable construction path for a
// market-data-capable service without changing existing NewService call sites.
func NewServiceWithQuoteProvider(store Store, clock Clock, provider QuoteProvider) *Service {
	service := NewService(store, clock)
	service.quoteProvider = provider
	return service
}
