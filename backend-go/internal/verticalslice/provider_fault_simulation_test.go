package verticalslice

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
)

type adversarialCorporateActionProvider struct {
	events []CorporateActionEvent
	err    error
}

func (p adversarialCorporateActionProvider) CorporateActions(_ context.Context, _ CorporateActionQuery) ([]CorporateActionEvent, error) {
	return p.events, p.err
}

func validAdversarialCorporateAction() CorporateActionEvent {
	record := "2026-01-10"
	payment := "2026-01-20"
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	amount := Money{Amount: decimal.Must("1.00000000"), Currency: RUB}
	return CorporateActionEvent{
		EventID: "evt-1", InstrumentID: "SBER", Kind: CorporateActionDividend,
		Status: CorporateActionConfirmed, RecordDate: &record, PaymentDate: &payment,
		AmountPerUnit: &amount, AsOf: now, RetrievedAt: now,
		Provenance: CorporateActionProvenance{Provider: "TEST", SourceEventID: "source-1"},
	}
}

func TestProviderFaultSimulationMapsUnknownProviderFailureToUnavailable(t *testing.T) {
	_, err := FetchCorporateActions(context.Background(), adversarialCorporateActionProvider{
		err: errors.New("upstream connection reset"),
	}, CorporateActionQuery{InstrumentIDs: []string{"SBER"}, From: "2026-01-01", To: "2026-02-01"})
	if !errors.Is(err, ErrCorporateActionsProviderUnavailable) {
		t.Fatalf("unknown provider failure must fail closed as unavailable, got %v", err)
	}
}

func TestProviderFaultSimulationPreservesCancellation(t *testing.T) {
	for _, providerErr := range []error{context.Canceled, context.DeadlineExceeded} {
		_, err := FetchCorporateActions(context.Background(), adversarialCorporateActionProvider{
			err: providerErr,
		}, CorporateActionQuery{InstrumentIDs: []string{"SBER"}, From: "2026-01-01", To: "2026-02-01"})
		if !errors.Is(err, providerErr) {
			t.Fatalf("provider cancellation/deadline was masked: input=%v got=%v", providerErr, err)
		}
	}
}

func TestProviderFaultSimulationRejectsSemanticPoisoning(t *testing.T) {
	base := validAdversarialCorporateAction()

	cases := []struct {
		name  string
		event CorporateActionEvent
	}{
		{
			name: "outside requested instrument",
			event: func() CorporateActionEvent { e := base; e.InstrumentID = "GAZP"; return e }(),
		},
		{
			name: "negative amount",
			event: func() CorporateActionEvent {
				e := base
				m := *e.AmountPerUnit
				m.Amount = decimal.Must("-1.00000000")
				e.AmountPerUnit = &m
				return e
			}(),
		},
		{
			name: "non UTC timestamps",
			event: func() CorporateActionEvent {
				e := base
				e.AsOf = e.AsOf.In(time.FixedZone("bad", 3600))
				return e
			}(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := FetchCorporateActions(context.Background(), adversarialCorporateActionProvider{
				events: []CorporateActionEvent{tc.event},
			}, CorporateActionQuery{InstrumentIDs: []string{"SBER"}, From: "2026-01-01", To: "2026-02-01"})
			if err == nil {
				t.Fatal("semantically poisoned provider response was accepted")
			}
		})
	}
}
