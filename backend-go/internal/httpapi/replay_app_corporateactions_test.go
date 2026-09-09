package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

type replayCorporateActionProvider struct{}

func (replayCorporateActionProvider) CorporateActions(_ context.Context, query verticalslice.CorporateActionQuery) ([]verticalslice.CorporateActionEvent, error) {
	recordDate := query.From
	paymentDate := query.To
	return []verticalslice.CorporateActionEvent{{
		EventID:      "replay-route-proof",
		InstrumentID: query.InstrumentIDs[0],
		Kind:         verticalslice.CorporateActionDividend,
		Status:       verticalslice.CorporateActionAnnounced,
		RecordDate:   &recordDate,
		PaymentDate:  &paymentDate,
		AsOf:         time.Date(2026, 9, 9, 10, 0, 0, 0, time.UTC),
		RetrievedAt:  time.Date(2026, 9, 9, 10, 0, 0, 0, time.UTC),
		Provenance: verticalslice.CorporateActionProvenance{
			Provider:      "FIXTURE",
			SourceEventID: "fixture-replay-route-proof",
		},
	}}, nil
}

func TestReplayProductionConstructorRegistersCorporateActionsRouteWhenProviderIsDisabled(t *testing.T) {
	app, err := NewReplay(nil, nil, []byte("openinvest-test-import-review-token-secret-32bytes"))
	if err != nil {
		t.Fatalf("NewReplay: %v", err)
	}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/corporate-actions/projection?instrumentId=SBER&from=2026-09-01&to=2026-09-30", nil)
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status=%d, want %d; a 404 means the shipped route regressed", response.StatusCode, http.StatusServiceUnavailable)
	}
}

func TestReplayProductionConstructorInjectsCorporateActionsProvider(t *testing.T) {
	app, err := NewReplayWithCorporateActionProvider(
		nil,
		nil,
		[]byte("openinvest-test-import-review-token-secret-32bytes"),
		replayCorporateActionProvider{},
	)
	if err != nil {
		t.Fatalf("NewReplayWithCorporateActionProvider: %v", err)
	}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/corporate-actions/projection?instrumentId=SBER&from=2026-09-01&to=2026-09-30", nil)
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status=%d, want %d", response.StatusCode, http.StatusOK)
	}
}
