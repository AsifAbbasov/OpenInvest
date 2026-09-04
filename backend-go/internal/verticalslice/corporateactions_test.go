package verticalslice

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
)

type corporateActionFixtureProvider struct {
	events []CorporateActionEvent
	err    error
	calls  int
	query  CorporateActionQuery
}

func (p *corporateActionFixtureProvider) CorporateActions(_ context.Context, query CorporateActionQuery) ([]CorporateActionEvent, error) {
	p.calls++
	p.query = query
	if p.err != nil {
		return nil, p.err
	}
	return append([]CorporateActionEvent(nil), p.events...), nil
}

type panicCorporateActionProvider struct{}

func (*panicCorporateActionProvider) CorporateActions(context.Context, CorporateActionQuery) ([]CorporateActionEvent, error) {
	panic("typed nil provider must never be called")
}

func TestValidateCorporateActionEventAcceptsCanonicalLifecycleStates(t *testing.T) {
	statuses := []CorporateActionStatus{
		CorporateActionAnnounced,
		CorporateActionConfirmed,
		CorporateActionPaid,
		CorporateActionCancelled,
	}
	for _, status := range statuses {
		t.Run(string(status), func(t *testing.T) {
			event := validCorporateActionEvent()
			event.Status = status
			if err := ValidateCorporateActionEvent(event); err != nil {
				t.Fatalf("ValidateCorporateActionEvent() error = %v", err)
			}
		})
	}
}

func TestValidateCorporateActionEventAcceptsUnknownAmountAndDates(t *testing.T) {
	event := validCorporateActionEvent()
	event.AmountPerUnit = nil
	event.RecordDate = nil
	event.PaymentDate = nil
	event.Status = CorporateActionAnnounced

	if err := ValidateCorporateActionEvent(event); err != nil {
		t.Fatalf("ValidateCorporateActionEvent() error = %v", err)
	}
}

func TestValidateCorporateActionEventAcceptsDividendAndCoupon(t *testing.T) {
	for _, kind := range []CorporateActionKind{CorporateActionDividend, CorporateActionCoupon} {
		event := validCorporateActionEvent()
		event.Kind = kind
		if err := ValidateCorporateActionEvent(event); err != nil {
			t.Fatalf("kind %q: %v", kind, err)
		}
	}
}

func TestValidateCorporateActionEventRejectsInvalidFields(t *testing.T) {
	nonUTC := time.FixedZone("UTC+3", 3*60*60)
	badDate := "2026-02-30"
	spacedDate := " 2026-10-01"
	negative := Money{Amount: decimal.Must("-1.00000000"), Currency: RUB}
	nilDecimal := Money{Currency: RUB}

	tests := []struct {
		name   string
		mutate func(*CorporateActionEvent)
	}{
		{"event id", func(e *CorporateActionEvent) { e.EventID = "bad id" }},
		{"instrument", func(e *CorporateActionEvent) { e.InstrumentID = "sber" }},
		{"kind", func(e *CorporateActionEvent) { e.Kind = "SPLIT" }},
		{"status", func(e *CorporateActionEvent) { e.Status = "EXPECTED" }},
		{"record date", func(e *CorporateActionEvent) { e.RecordDate = &badDate }},
		{"payment date", func(e *CorporateActionEvent) { e.PaymentDate = &spacedDate }},
		{"negative amount", func(e *CorporateActionEvent) { e.AmountPerUnit = &negative }},
		{"nil decimal", func(e *CorporateActionEvent) { e.AmountPerUnit = &nilDecimal }},
		{"currency", func(e *CorporateActionEvent) { e.AmountPerUnit.Currency = "rub" }},
		{"zero asof", func(e *CorporateActionEvent) { e.AsOf = time.Time{} }},
		{"zero retrieved", func(e *CorporateActionEvent) { e.RetrievedAt = time.Time{} }},
		{"non utc asof", func(e *CorporateActionEvent) { e.AsOf = time.Date(2026, 9, 4, 12, 0, 0, 0, nonUTC) }},
		{"non utc retrieved", func(e *CorporateActionEvent) { e.RetrievedAt = time.Date(2026, 9, 4, 12, 0, 0, 0, nonUTC) }},
		{"provider empty", func(e *CorporateActionEvent) { e.Provenance.Provider = "" }},
		{"provider noncanonical", func(e *CorporateActionEvent) { e.Provenance.Provider = "nsd" }},
		{"source event empty", func(e *CorporateActionEvent) { e.Provenance.SourceEventID = "" }},
		{"source event padded", func(e *CorporateActionEvent) { e.Provenance.SourceEventID = " source-1 " }},
		{"source event control", func(e *CorporateActionEvent) { e.Provenance.SourceEventID = "source\n1" }},
		{"source event c1 control", func(e *CorporateActionEvent) { e.Provenance.SourceEventID = "source\u00851" }},
		{"self supersedes", func(e *CorporateActionEvent) { value := e.EventID; e.SupersedesEventID = &value }},
		{"bad supersedes", func(e *CorporateActionEvent) { value := "bad id"; e.SupersedesEventID = &value }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := validCorporateActionEvent()
			tt.mutate(&event)
			if err := ValidateCorporateActionEvent(event); !errors.Is(err, ErrInvalidCorporateAction) {
				t.Fatalf("error = %v, want ErrInvalidCorporateAction", err)
			}
		})
	}
}

func TestValidateCorporateActionEventAcceptsOpaqueSourceEventID(t *testing.T) {
	event := validCorporateActionEvent()
	event.Provenance.SourceEventID = "issuer/2026:дивиденд:42"
	if err := ValidateCorporateActionEvent(event); err != nil {
		t.Fatalf("ValidateCorporateActionEvent() error = %v", err)
	}
}

func TestValidateCorporateActionEventRepresentsCorrectionWithoutDeletingPriorEvidence(t *testing.T) {
	previous := validCorporateActionEvent()
	corrected := validCorporateActionEvent()
	corrected.EventID = "ca:SBER:2026:2"
	corrected.Provenance.SourceEventID = "source-2"
	corrected.SupersedesEventID = &previous.EventID
	corrected.Status = CorporateActionCancelled

	if err := ValidateCorporateActionEvents([]CorporateActionEvent{previous, corrected}); err != nil {
		t.Fatalf("ValidateCorporateActionEvents() error = %v", err)
	}
}

func TestValidateCorporateActionEventsRejectsDuplicateApplicationIdentity(t *testing.T) {
	first := validCorporateActionEvent()
	second := validCorporateActionEvent()
	second.Provenance.SourceEventID = "source-2"

	err := ValidateCorporateActionEvents([]CorporateActionEvent{first, second})
	if !errors.Is(err, ErrInvalidCorporateAction) {
		t.Fatalf("error = %v, want ErrInvalidCorporateAction", err)
	}
}

func TestValidateCorporateActionQuery(t *testing.T) {
	valid := CorporateActionQuery{InstrumentIDs: []string{"SBER", "RU000A123456"}, From: "2026-01-01", To: "2026-12-31"}
	if err := ValidateCorporateActionQuery(valid); err != nil {
		t.Fatalf("valid query error = %v", err)
	}

	tests := []CorporateActionQuery{
		{},
		{InstrumentIDs: []string{"sber"}, From: "2026-01-01", To: "2026-12-31"},
		{InstrumentIDs: []string{"SBER", "SBER"}, From: "2026-01-01", To: "2026-12-31"},
		{InstrumentIDs: []string{"SBER"}, From: "", To: "2026-12-31"},
		{InstrumentIDs: []string{"SBER"}, From: "2026-01-01", To: "2026-02-30"},
		{InstrumentIDs: []string{"SBER"}, From: "2027-01-01", To: "2026-12-31"},
	}
	for i, query := range tests {
		if err := ValidateCorporateActionQuery(query); !errors.Is(err, ErrInvalidCorporateActionQuery) {
			t.Fatalf("case %d error = %v, want ErrInvalidCorporateActionQuery", i, err)
		}
	}
}

func TestFetchCorporateActionsValidatesBoundaryAndCopiesFixtureSlice(t *testing.T) {
	event := validCorporateActionEvent()
	provider := &corporateActionFixtureProvider{events: []CorporateActionEvent{event}}
	query := CorporateActionQuery{InstrumentIDs: []string{"SBER"}, From: "2026-01-01", To: "2026-12-31"}

	events, err := FetchCorporateActions(context.Background(), provider, query)
	if err != nil {
		t.Fatalf("FetchCorporateActions() error = %v", err)
	}
	if provider.calls != 1 || provider.query.From != query.From || len(events) != 1 || events[0].EventID != event.EventID {
		t.Fatalf("unexpected boundary result: calls=%d query=%+v events=%+v", provider.calls, provider.query, events)
	}

	events[0].EventID = "changed"
	if provider.events[0].EventID != event.EventID {
		t.Fatal("fixture provider leaked slice mutation")
	}
}

func TestFetchCorporateActionsIsolatesQuerySliceFromProviderMutation(t *testing.T) {
	query := CorporateActionQuery{InstrumentIDs: []string{"SBER"}, From: "2026-01-01", To: "2026-12-31"}
	provider := corporateActionMutatingProvider{}

	_, err := FetchCorporateActions(context.Background(), provider, query)
	if err != nil {
		t.Fatalf("FetchCorporateActions() error = %v", err)
	}
	if query.InstrumentIDs[0] != "SBER" {
		t.Fatalf("caller query mutated: %+v", query)
	}
}

type corporateActionMutatingProvider struct{}

func (corporateActionMutatingProvider) CorporateActions(_ context.Context, query CorporateActionQuery) ([]CorporateActionEvent, error) {
	query.InstrumentIDs[0] = "GAZP"
	return nil, nil
}

func TestFetchCorporateActionsRejectsInvalidQueryBeforeProviderCall(t *testing.T) {
	provider := &corporateActionFixtureProvider{}
	_, err := FetchCorporateActions(context.Background(), provider, CorporateActionQuery{})
	if !errors.Is(err, ErrInvalidCorporateActionQuery) {
		t.Fatalf("error = %v, want ErrInvalidCorporateActionQuery", err)
	}
	if provider.calls != 0 {
		t.Fatalf("provider calls = %d, want 0", provider.calls)
	}
}

func TestFetchCorporateActionsRejectsNilAndTypedNilProvider(t *testing.T) {
	query := CorporateActionQuery{InstrumentIDs: []string{"SBER"}, From: "2026-01-01", To: "2026-12-31"}
	var typedNil *panicCorporateActionProvider
	for _, provider := range []CorporateActionProvider{nil, typedNil} {
		_, err := FetchCorporateActions(context.Background(), provider, query)
		if !errors.Is(err, ErrCorporateActionsProviderUnavailable) {
			t.Fatalf("error = %v, want ErrCorporateActionsProviderUnavailable", err)
		}
	}
}

func TestFetchCorporateActionsPreservesProviderNeutralErrors(t *testing.T) {
	query := CorporateActionQuery{InstrumentIDs: []string{"SBER"}, From: "2026-01-01", To: "2026-12-31"}
	for _, sentinel := range []error{ErrCorporateActionsProviderUnavailable, ErrCorporateActionsProviderData, context.Canceled, context.DeadlineExceeded} {
		provider := &corporateActionFixtureProvider{err: sentinel}
		_, err := FetchCorporateActions(context.Background(), provider, query)
		if !errors.Is(err, sentinel) {
			t.Fatalf("provider error %v normalized to %v", sentinel, err)
		}
	}
}

func TestFetchCorporateActionsNormalizesUnclassifiedProviderError(t *testing.T) {
	provider := &corporateActionFixtureProvider{err: errors.New("provider-specific HTTP 599 body secret")}
	query := CorporateActionQuery{InstrumentIDs: []string{"SBER"}, From: "2026-01-01", To: "2026-12-31"}

	_, err := FetchCorporateActions(context.Background(), provider, query)
	if !errors.Is(err, ErrCorporateActionsProviderUnavailable) {
		t.Fatalf("error = %v, want ErrCorporateActionsProviderUnavailable", err)
	}
	if err.Error() != ErrCorporateActionsProviderUnavailable.Error() {
		t.Fatalf("provider-specific error leaked: %q", err.Error())
	}
}

func TestFetchCorporateActionsRejectsMalformedProviderOutput(t *testing.T) {
	event := validCorporateActionEvent()
	event.Provenance.Provider = ""
	provider := &corporateActionFixtureProvider{events: []CorporateActionEvent{event}}
	query := CorporateActionQuery{InstrumentIDs: []string{"SBER"}, From: "2026-01-01", To: "2026-12-31"}

	_, err := FetchCorporateActions(context.Background(), provider, query)
	if !errors.Is(err, ErrInvalidCorporateAction) {
		t.Fatalf("error = %v, want ErrInvalidCorporateAction", err)
	}
}

func TestFetchCorporateActionsRejectsReturnedInstrumentOutsideRequestedSet(t *testing.T) {
	event := validCorporateActionEvent()
	event.InstrumentID = "GAZP"
	provider := &corporateActionFixtureProvider{events: []CorporateActionEvent{event}}
	query := CorporateActionQuery{InstrumentIDs: []string{"SBER"}, From: "2026-01-01", To: "2026-12-31"}

	_, err := FetchCorporateActions(context.Background(), provider, query)
	if !errors.Is(err, ErrInvalidCorporateAction) {
		t.Fatalf("error = %v, want ErrInvalidCorporateAction", err)
	}
}

func TestFetchCorporateActionsAllowsEmptyResultWithoutFabrication(t *testing.T) {
	provider := &corporateActionFixtureProvider{events: nil}
	query := CorporateActionQuery{InstrumentIDs: []string{"SBER"}, From: "2026-01-01", To: "2026-12-31"}

	events, err := FetchCorporateActions(context.Background(), provider, query)
	if err != nil {
		t.Fatalf("FetchCorporateActions() error = %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("events = %v, want empty", events)
	}
}

func validCorporateActionEvent() CorporateActionEvent {
	recordDate := "2026-10-01"
	paymentDate := "2026-10-15"
	amount := Money{Amount: decimal.Must("25.12345678"), Currency: RUB}
	return CorporateActionEvent{
		EventID:       "ca:SBER:2026:1",
		InstrumentID:  "SBER",
		Kind:          CorporateActionDividend,
		Status:        CorporateActionConfirmed,
		RecordDate:    &recordDate,
		PaymentDate:   &paymentDate,
		AmountPerUnit: &amount,
		AsOf:          time.Date(2026, 9, 4, 10, 0, 0, 0, time.UTC),
		RetrievedAt:   time.Date(2026, 9, 4, 10, 5, 0, 0, time.UTC),
		Provenance: CorporateActionProvenance{
			Provider:      "FIXTURE_CORPORATE_ACTIONS",
			SourceEventID: "source-1",
		},
	}
}
