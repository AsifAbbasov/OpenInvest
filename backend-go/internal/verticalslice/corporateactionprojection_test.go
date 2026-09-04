package verticalslice

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
)

func TestProjectCorporateActionCalendarUsesCanonicalEffectiveDateRules(t *testing.T) {
	record := "2026-10-10"
	payment := "2026-10-20"
	couponRecord := "2026-11-01"
	couponPayment := "2026-11-15"
	dividendFallbackPayment := "2026-12-05"

	dividend := projectionEvent("evt-div", "SBER", CorporateActionDividend, CorporateActionConfirmed)
	dividend.RecordDate = &record
	dividend.PaymentDate = &payment

	coupon := projectionEvent("evt-cpn", "RU000A", CorporateActionCoupon, CorporateActionConfirmed)
	coupon.RecordDate = &couponRecord
	coupon.PaymentDate = &couponPayment

	dividendFallback := projectionEvent("evt-div-fallback", "GAZP", CorporateActionDividend, CorporateActionAnnounced)
	dividendFallback.PaymentDate = &dividendFallbackPayment

	undatedCoupon := projectionEvent("evt-undated", "LKOH", CorporateActionCoupon, CorporateActionAnnounced)

	entries, err := ProjectCorporateActionCalendar([]CorporateActionEvent{coupon, undatedCoupon, dividendFallback, dividend})
	if err != nil {
		t.Fatalf("ProjectCorporateActionCalendar() error = %v", err)
	}
	if got, want := len(entries), 3; got != want {
		t.Fatalf("len(entries) = %d, want %d", got, want)
	}
	assertCalendarEntry(t, entries[0], "2026-10-10", "evt-div")
	assertCalendarEntry(t, entries[1], "2026-11-15", "evt-cpn")
	assertCalendarEntry(t, entries[2], "2026-12-05", "evt-div-fallback")
}

func TestProjectCorporateActionCalendarSortsDeterministically(t *testing.T) {
	date := "2026-10-10"
	first := projectionEvent("evt-b", "SBER", CorporateActionDividend, CorporateActionConfirmed)
	first.RecordDate = &date
	second := projectionEvent("evt-a", "GAZP", CorporateActionCoupon, CorporateActionPaid)
	second.PaymentDate = &date
	third := projectionEvent("evt-c", "SBER", CorporateActionCoupon, CorporateActionPaid)
	third.PaymentDate = &date

	entries, err := ProjectCorporateActionCalendar([]CorporateActionEvent{first, third, second})
	if err != nil {
		t.Fatalf("ProjectCorporateActionCalendar() error = %v", err)
	}
	got := []string{entries[0].Event.EventID + "/" + entries[0].Event.InstrumentID, entries[1].Event.EventID + "/" + entries[1].Event.InstrumentID, entries[2].Event.EventID + "/" + entries[2].Event.InstrumentID}
	want := []string{"evt-a/GAZP", "evt-c/SBER", "evt-b/SBER"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("order = %#v, want %#v", got, want)
	}
}

func TestProjectCorporateActionCalendarUsesEventIDAsFinalTieBreak(t *testing.T) {
	date := "2026-10-10"
	one := projectionEvent("evt-b", "SBER", CorporateActionDividend, CorporateActionConfirmed)
	one.RecordDate = &date
	two := projectionEvent("evt-a", "SBER", CorporateActionDividend, CorporateActionConfirmed)
	two.RecordDate = &date

	entries, err := ProjectCorporateActionCalendar([]CorporateActionEvent{one, two})
	if err != nil {
		t.Fatalf("ProjectCorporateActionCalendar() error = %v", err)
	}
	if got, want := []string{entries[0].Event.EventID, entries[1].Event.EventID}, []string{"evt-a", "evt-b"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("event order = %#v, want %#v", got, want)
	}
}

func TestProjectCorporateActionHeatmapPropagatesProjectionError(t *testing.T) {
	bad := projectionEvent("evt-bad", "sber", CorporateActionDividend, CorporateActionConfirmed)
	_, err := ProjectCorporateActionHeatmap([]CorporateActionEvent{bad})
	if !errors.Is(err, ErrInvalidCorporateAction) {
		t.Fatalf("error = %v, want ErrInvalidCorporateAction", err)
	}
}

func TestProjectCorporateActionCalendarUsesOnlyCurrentUnsupersededEvidence(t *testing.T) {
	oldDate := "2026-10-10"
	newDate := "2026-10-12"
	old := projectionEvent("evt-old", "SBER", CorporateActionDividend, CorporateActionAnnounced)
	old.RecordDate = &oldDate
	corrected := projectionEvent("evt-corrected", "SBER", CorporateActionDividend, CorporateActionConfirmed)
	corrected.RecordDate = &newDate
	corrected.SupersedesEventID = stringPointer("evt-old")

	entries, err := ProjectCorporateActionCalendar([]CorporateActionEvent{old, corrected})
	if err != nil {
		t.Fatalf("ProjectCorporateActionCalendar() error = %v", err)
	}
	if got, want := len(entries), 1; got != want {
		t.Fatalf("len(entries) = %d, want %d", got, want)
	}
	assertCalendarEntry(t, entries[0], newDate, "evt-corrected")
}

func TestProjectCorporateActionCalendarCancellationCanRemoveDatedProjection(t *testing.T) {
	date := "2026-10-10"
	old := projectionEvent("evt-old", "SBER", CorporateActionDividend, CorporateActionConfirmed)
	old.RecordDate = &date
	cancelled := projectionEvent("evt-cancel", "SBER", CorporateActionDividend, CorporateActionCancelled)
	cancelled.SupersedesEventID = stringPointer("evt-old")

	entries, err := ProjectCorporateActionCalendar([]CorporateActionEvent{old, cancelled})
	if err != nil {
		t.Fatalf("ProjectCorporateActionCalendar() error = %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("entries = %#v, want empty dated projection", entries)
	}
}

func TestProjectCorporateActionCalendarAllowsMissingHistoricalPredecessor(t *testing.T) {
	date := "2026-10-12"
	current := projectionEvent("evt-current", "SBER", CorporateActionDividend, CorporateActionConfirmed)
	current.RecordDate = &date
	current.SupersedesEventID = stringPointer("evt-outside-window")

	entries, err := ProjectCorporateActionCalendar([]CorporateActionEvent{current})
	if err != nil {
		t.Fatalf("ProjectCorporateActionCalendar() error = %v", err)
	}
	if got, want := len(entries), 1; got != want {
		t.Fatalf("len(entries) = %d, want %d", got, want)
	}
	assertCalendarEntry(t, entries[0], date, "evt-current")
}

func TestProjectCorporateActionCalendarRejectsAmbiguousSupersessionFork(t *testing.T) {
	date := "2026-10-10"
	base := projectionEvent("evt-base", "SBER", CorporateActionDividend, CorporateActionAnnounced)
	base.RecordDate = &date
	one := projectionEvent("evt-one", "SBER", CorporateActionDividend, CorporateActionConfirmed)
	one.RecordDate = &date
	one.SupersedesEventID = stringPointer("evt-base")
	two := projectionEvent("evt-two", "SBER", CorporateActionDividend, CorporateActionCancelled)
	two.RecordDate = &date
	two.SupersedesEventID = stringPointer("evt-base")

	_, err := ProjectCorporateActionCalendar([]CorporateActionEvent{base, one, two})
	if !errors.Is(err, ErrInvalidCorporateActionProjection) {
		t.Fatalf("error = %v, want ErrInvalidCorporateActionProjection", err)
	}
}

func TestProjectCorporateActionCalendarRejectsCrossInstrumentSupersession(t *testing.T) {
	date := "2026-10-10"
	base := projectionEvent("evt-base", "SBER", CorporateActionDividend, CorporateActionConfirmed)
	base.RecordDate = &date
	wrong := projectionEvent("evt-wrong", "GAZP", CorporateActionDividend, CorporateActionCancelled)
	wrong.RecordDate = &date
	wrong.SupersedesEventID = stringPointer("evt-base")

	_, err := ProjectCorporateActionCalendar([]CorporateActionEvent{base, wrong})
	if !errors.Is(err, ErrInvalidCorporateActionProjection) {
		t.Fatalf("error = %v, want ErrInvalidCorporateActionProjection", err)
	}
}

func TestProjectCorporateActionCalendarRejectsCrossKindSupersession(t *testing.T) {
	date := "2026-10-10"
	base := projectionEvent("evt-base", "SBER", CorporateActionDividend, CorporateActionConfirmed)
	base.RecordDate = &date
	wrong := projectionEvent("evt-wrong", "SBER", CorporateActionCoupon, CorporateActionCancelled)
	wrong.PaymentDate = &date
	wrong.SupersedesEventID = stringPointer("evt-base")

	_, err := ProjectCorporateActionCalendar([]CorporateActionEvent{base, wrong})
	if !errors.Is(err, ErrInvalidCorporateActionProjection) {
		t.Fatalf("error = %v, want ErrInvalidCorporateActionProjection", err)
	}
}

func TestProjectCorporateActionCalendarRejectsSupersessionCycle(t *testing.T) {
	date := "2026-10-10"
	one := projectionEvent("evt-one", "SBER", CorporateActionDividend, CorporateActionConfirmed)
	one.RecordDate = &date
	one.SupersedesEventID = stringPointer("evt-two")
	two := projectionEvent("evt-two", "SBER", CorporateActionDividend, CorporateActionCancelled)
	two.RecordDate = &date
	two.SupersedesEventID = stringPointer("evt-one")

	_, err := ProjectCorporateActionCalendar([]CorporateActionEvent{one, two})
	if !errors.Is(err, ErrInvalidCorporateActionProjection) {
		t.Fatalf("error = %v, want ErrInvalidCorporateActionProjection", err)
	}
}

func TestProjectCorporateActionCalendarPreservesCanonicalValidation(t *testing.T) {
	bad := projectionEvent("evt-bad", "sber", CorporateActionDividend, CorporateActionConfirmed)
	_, err := ProjectCorporateActionCalendar([]CorporateActionEvent{bad})
	if !errors.Is(err, ErrInvalidCorporateAction) {
		t.Fatalf("error = %v, want ErrInvalidCorporateAction", err)
	}
}

func TestProjectCorporateActionCalendarDoesNotAliasPointerFields(t *testing.T) {
	date := "2026-10-10"
	amount := Money{Amount: decimal.Must("12.34000000"), Currency: RUB}
	event := projectionEvent("evt-one", "SBER", CorporateActionDividend, CorporateActionConfirmed)
	event.RecordDate = &date
	event.AmountPerUnit = &amount

	entries, err := ProjectCorporateActionCalendar([]CorporateActionEvent{event})
	if err != nil {
		t.Fatalf("ProjectCorporateActionCalendar() error = %v", err)
	}
	*entries[0].Event.RecordDate = "2026-12-31"
	entries[0].Event.AmountPerUnit.Currency = "USD"
	if *event.RecordDate != date {
		t.Fatalf("input RecordDate mutated to %q", *event.RecordDate)
	}
	if event.AmountPerUnit.Currency != RUB {
		t.Fatalf("input AmountPerUnit currency mutated to %q", event.AmountPerUnit.Currency)
	}
}

func TestProjectCorporateActionHeatmapCountsByKindAndStatusWithoutMoneyAggregation(t *testing.T) {
	date1 := "2026-10-10"
	date2 := "2026-10-11"

	announcedDividend := projectionEvent("evt-1", "SBER", CorporateActionDividend, CorporateActionAnnounced)
	announcedDividend.RecordDate = &date1
	confirmedCoupon := projectionEvent("evt-2", "RU000A", CorporateActionCoupon, CorporateActionConfirmed)
	confirmedCoupon.PaymentDate = &date1
	paidDividend := projectionEvent("evt-3", "GAZP", CorporateActionDividend, CorporateActionPaid)
	paidDividend.RecordDate = &date2
	cancelledCoupon := projectionEvent("evt-4", "LKOH", CorporateActionCoupon, CorporateActionCancelled)
	cancelledCoupon.PaymentDate = &date2

	buckets, err := ProjectCorporateActionHeatmap([]CorporateActionEvent{cancelledCoupon, paidDividend, confirmedCoupon, announcedDividend})
	if err != nil {
		t.Fatalf("ProjectCorporateActionHeatmap() error = %v", err)
	}
	want := []CorporateActionHeatmapBucket{
		{Date: date1, TotalCount: 2, DividendCount: 1, CouponCount: 1, AnnouncedCount: 1, ConfirmedCount: 1},
		{Date: date2, TotalCount: 2, DividendCount: 1, CouponCount: 1, PaidCount: 1, CancelledCount: 1},
	}
	if !reflect.DeepEqual(buckets, want) {
		t.Fatalf("buckets = %#v, want %#v", buckets, want)
	}
}

func TestProjectCorporateActionHeatmapUsesCurrentProjectionAfterSupersession(t *testing.T) {
	oldDate := "2026-10-10"
	newDate := "2026-10-12"
	old := projectionEvent("evt-old", "SBER", CorporateActionDividend, CorporateActionAnnounced)
	old.RecordDate = &oldDate
	current := projectionEvent("evt-current", "SBER", CorporateActionDividend, CorporateActionConfirmed)
	current.RecordDate = &newDate
	current.SupersedesEventID = stringPointer("evt-old")

	buckets, err := ProjectCorporateActionHeatmap([]CorporateActionEvent{old, current})
	if err != nil {
		t.Fatalf("ProjectCorporateActionHeatmap() error = %v", err)
	}
	want := []CorporateActionHeatmapBucket{{Date: newDate, TotalCount: 1, DividendCount: 1, ConfirmedCount: 1}}
	if !reflect.DeepEqual(buckets, want) {
		t.Fatalf("buckets = %#v, want %#v", buckets, want)
	}
}

func TestCorporateActionProjectionAcceptsEmptyInput(t *testing.T) {
	calendar, err := ProjectCorporateActionCalendar(nil)
	if err != nil {
		t.Fatalf("ProjectCorporateActionCalendar(nil) error = %v", err)
	}
	if len(calendar) != 0 {
		t.Fatalf("calendar = %#v, want empty", calendar)
	}
	heatmap, err := ProjectCorporateActionHeatmap(nil)
	if err != nil {
		t.Fatalf("ProjectCorporateActionHeatmap(nil) error = %v", err)
	}
	if len(heatmap) != 0 {
		t.Fatalf("heatmap = %#v, want empty", heatmap)
	}
}

func projectionEvent(id string, instrument string, kind CorporateActionKind, status CorporateActionStatus) CorporateActionEvent {
	return CorporateActionEvent{
		EventID:      id,
		InstrumentID: instrument,
		Kind:         kind,
		Status:       status,
		AsOf:         time.Date(2026, 9, 4, 12, 0, 0, 0, time.UTC),
		RetrievedAt:  time.Date(2026, 9, 4, 12, 1, 0, 0, time.UTC),
		Provenance: CorporateActionProvenance{
			Provider:      "FIXTURE",
			SourceEventID: "source-" + id,
		},
	}
}

func stringPointer(value string) *string { return &value }

func assertCalendarEntry(t *testing.T, entry CorporateActionCalendarEntry, date string, eventID string) {
	t.Helper()
	if entry.EffectiveDate != date || entry.Event.EventID != eventID {
		t.Fatalf("entry = %#v, want date=%q eventID=%q", entry, date, eventID)
	}
}
