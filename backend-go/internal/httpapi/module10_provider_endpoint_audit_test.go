package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

type module10CountingCorporateProvider struct {
	calls atomic.Int64
}

func (p *module10CountingCorporateProvider) CorporateActions(_ context.Context, _ verticalslice.CorporateActionQuery) ([]verticalslice.CorporateActionEvent, error) {
	p.calls.Add(1)
	return []verticalslice.CorporateActionEvent{}, nil
}

func TestModule10ProviderEndpointBudgetSurface(t *testing.T) {
	provider := &module10CountingCorporateProvider{}
	app := fiber.New()
	api := &API{corporateActionProvider: provider}
	app.Get("/api/v1/corporate-actions/projection", api.getCorporateActionProjection)

	validRequests := 30
	for i := 0; i < validRequests; i++ {
		instrument := "SBER"
		if i%2 == 1 {
			instrument = "GAZP"
		}
		target := fmt.Sprintf(
			"/api/v1/corporate-actions/projection?instrumentId=%s&from=2026-01-01&to=2026-12-31",
			instrument,
		)
		req := httptest.NewRequest(http.MethodGet, target, nil)
		req.Header.Set("X-Request-ID", fmt.Sprintf("00000000-0000-4000-8000-%012d", i))
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("valid request %d: %v", i, err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("valid request %d status=%d", i, resp.StatusCode)
		}
		_ = resp.Body.Close()
	}
	if got := provider.calls.Load(); got != int64(validRequests) {
		t.Fatalf("provider calls=%d want=%d", got, validRequests)
	}

	beforeInvalid := provider.calls.Load()
	invalidTargets := []string{
		"/api/v1/corporate-actions/projection?instrumentId=sber&from=2026-01-01&to=2026-12-31",
		"/api/v1/corporate-actions/projection?instrumentId=SBER&from=bad&to=2026-12-31",
		"/api/v1/corporate-actions/projection?instrumentId=SBER,SBER&from=2026-01-01&to=2026-12-31",
		"/api/v1/corporate-actions/projection?instrumentId=SBER&from=2026-12-31&to=2026-01-01",
	}
	for _, target := range invalidTargets {
		resp, err := app.Test(httptest.NewRequest(http.MethodGet, target, nil))
		if err != nil {
			t.Fatalf("invalid request %q: %v", target, err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("invalid request %q status=%d want=400", target, resp.StatusCode)
		}
		_ = resp.Body.Close()
	}
	if got := provider.calls.Load(); got != beforeInvalid {
		t.Fatalf("invalid requests reached provider: before=%d after=%d", beforeInvalid, got)
	}

	t.Logf(
		"MODULE10_ENDPOINT valid_requests=%d provider_calls=%d invalid_rejected_before_provider=%d auth_required=false endpoint_cache=false endpoint_dedup=false",
		validRequests,
		provider.calls.Load(),
		len(invalidTargets),
	)
}
