package verticalslice

import (
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
)

func FuzzCorporateActionProjectionIntegrity(f *testing.F) {
	f.Add("evt-1", "evt-2", "SBER", "2026-01-01", "2026-01-10", false)
	f.Add("evt-1", "evt-1", "SBER", "bad-date", "2026-01-10", true)

	f.Fuzz(func(t *testing.T, firstID, secondID, instrument, recordDate, paymentDate string, link bool) {
		if len(firstID) > 256 || len(secondID) > 256 || len(instrument) > 64 ||
			len(recordDate) > 64 || len(paymentDate) > 64 {
			t.Skip()
		}
		now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		amount := Money{Amount: decimal.Must("1.00000000"), Currency: RUB}
		first := CorporateActionEvent{
			EventID: firstID, InstrumentID: instrument, Kind: CorporateActionDividend,
			Status: CorporateActionConfirmed, RecordDate: &recordDate, PaymentDate: &paymentDate,
			AmountPerUnit: &amount, AsOf: now, RetrievedAt: now,
			Provenance: CorporateActionProvenance{Provider: "FUZZ", SourceEventID: "source-1"},
		}
		second := first
		second.EventID = secondID
		second.Provenance.SourceEventID = "source-2"
		if link {
			second.SupersedesEventID = &firstID
		}

		events := []CorporateActionEvent{first, second}
		firstCalendar, err := ProjectCorporateActionCalendar(events)
		if err != nil {
			return
		}
		secondCalendar, err := ProjectCorporateActionCalendar(events)
		if err != nil {
			t.Fatalf("corporate-action projection became nondeterministic: %v", err)
		}
		if len(firstCalendar) != len(secondCalendar) {
			t.Fatalf("corporate-action projection length nondeterminism")
		}
		for i := range firstCalendar {
			if firstCalendar[i].EffectiveDate != secondCalendar[i].EffectiveDate ||
				firstCalendar[i].Event.EventID != secondCalendar[i].Event.EventID {
				t.Fatalf("corporate-action projection ordering nondeterminism")
			}
		}

		heatmap, err := ProjectCorporateActionHeatmap(events)
		if err != nil {
			t.Fatalf("calendar accepted events but heatmap rejected them: %v", err)
		}
		total := 0
		for _, bucket := range heatmap {
			if bucket.TotalCount != bucket.DividendCount+bucket.CouponCount {
				t.Fatalf("heatmap kind accounting drift: %+v", bucket)
			}
			if bucket.TotalCount != bucket.AnnouncedCount+bucket.ConfirmedCount+bucket.PaidCount+bucket.CancelledCount {
				t.Fatalf("heatmap status accounting drift: %+v", bucket)
			}
			total += bucket.TotalCount
		}
		if total != len(firstCalendar) {
			t.Fatalf("calendar/heatmap cardinality drift: calendar=%d heatmap=%d", len(firstCalendar), total)
		}
	})
}
