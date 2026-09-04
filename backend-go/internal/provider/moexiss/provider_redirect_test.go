package moexiss

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestQuoteDoesNotFollowRedirectOutsideProviderHost(t *testing.T) {
	var targetHits atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		targetHits.Add(1)
	}))
	defer target.Close()

	provider, source := newTestProvider(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusFound)
	}), fixedClock{now: deterministicNow})
	defer source.Close()

	_, err := provider.Quote(context.Background(), "SBER")
	if !errors.Is(err, verticalslice.ErrMarketQuoteProviderData) {
		t.Fatalf("expected provider data error, got %v", err)
	}
	if got := targetHits.Load(); got != 0 {
		t.Fatalf("redirect target was contacted %d times", got)
	}
}
