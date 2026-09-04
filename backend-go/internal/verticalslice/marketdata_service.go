package verticalslice

import "reflect"

// NewServiceWithQuoteProvider provides the immutable construction path for a
// market-data-capable service without changing existing NewService call sites.
func NewServiceWithQuoteProvider(store Store, clock Clock, provider QuoteProvider) *Service {
	service := NewService(store, clock)
	if isNilQuoteProvider(provider) {
		return service
	}
	service.quoteProvider = provider
	return service
}

func isNilQuoteProvider(provider QuoteProvider) bool {
	if provider == nil {
		return true
	}
	value := reflect.ValueOf(provider)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
