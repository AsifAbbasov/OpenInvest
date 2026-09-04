package verticalslice

import (
	"errors"
	"fmt"
	"sort"
)

var ErrInvalidCorporateActionProjection = errors.New("invalid corporate action projection")

type CorporateActionCalendarEntry struct {
	EffectiveDate string
	Event         CorporateActionEvent
}

type CorporateActionHeatmapBucket struct {
	Date           string
	TotalCount     int
	DividendCount  int
	CouponCount    int
	AnnouncedCount int
	ConfirmedCount int
	PaidCount      int
	CancelledCount int
}

func ProjectCorporateActionCalendar(events []CorporateActionEvent) ([]CorporateActionCalendarEntry, error) {
	if err := ValidateCorporateActionEvents(events); err != nil {
		return nil, err
	}
	if err := validateCorporateActionSupersession(events); err != nil {
		return nil, err
	}

	superseded := make(map[string]struct{}, len(events))
	for _, event := range events {
		if event.SupersedesEventID != nil {
			superseded[*event.SupersedesEventID] = struct{}{}
		}
	}

	entries := make([]CorporateActionCalendarEntry, 0, len(events))
	for _, event := range events {
		if _, ok := superseded[event.EventID]; ok {
			continue
		}
		effectiveDate, ok := corporateActionEffectiveDate(event)
		if !ok {
			continue
		}
		entries = append(entries, CorporateActionCalendarEntry{
			EffectiveDate: effectiveDate,
			Event:         cloneCorporateActionEvent(event),
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		left, right := entries[i], entries[j]
		if left.EffectiveDate != right.EffectiveDate {
			return left.EffectiveDate < right.EffectiveDate
		}
		if left.Event.InstrumentID != right.Event.InstrumentID {
			return left.Event.InstrumentID < right.Event.InstrumentID
		}
		if left.Event.Kind != right.Event.Kind {
			return left.Event.Kind < right.Event.Kind
		}
		return left.Event.EventID < right.Event.EventID
	})
	return entries, nil
}

func ProjectCorporateActionHeatmap(events []CorporateActionEvent) ([]CorporateActionHeatmapBucket, error) {
	entries, err := ProjectCorporateActionCalendar(events)
	if err != nil {
		return nil, err
	}

	byDate := make(map[string]*CorporateActionHeatmapBucket, len(entries))
	for _, entry := range entries {
		bucket := byDate[entry.EffectiveDate]
		if bucket == nil {
			bucket = &CorporateActionHeatmapBucket{Date: entry.EffectiveDate}
			byDate[entry.EffectiveDate] = bucket
		}
		bucket.TotalCount++
		switch entry.Event.Kind {
		case CorporateActionDividend:
			bucket.DividendCount++
		case CorporateActionCoupon:
			bucket.CouponCount++
		}
		switch entry.Event.Status {
		case CorporateActionAnnounced:
			bucket.AnnouncedCount++
		case CorporateActionConfirmed:
			bucket.ConfirmedCount++
		case CorporateActionPaid:
			bucket.PaidCount++
		case CorporateActionCancelled:
			bucket.CancelledCount++
		}
	}

	buckets := make([]CorporateActionHeatmapBucket, 0, len(byDate))
	for _, bucket := range byDate {
		buckets = append(buckets, *bucket)
	}
	sort.Slice(buckets, func(i, j int) bool { return buckets[i].Date < buckets[j].Date })
	return buckets, nil
}

func corporateActionEffectiveDate(event CorporateActionEvent) (string, bool) {
	switch event.Kind {
	case CorporateActionDividend:
		if event.RecordDate != nil {
			return *event.RecordDate, true
		}
		if event.PaymentDate != nil {
			return *event.PaymentDate, true
		}
	case CorporateActionCoupon:
		if event.PaymentDate != nil {
			return *event.PaymentDate, true
		}
	}
	return "", false
}

func validateCorporateActionSupersession(events []CorporateActionEvent) error {
	byID := make(map[string]CorporateActionEvent, len(events))
	for _, event := range events {
		byID[event.EventID] = event
	}

	supersederByTarget := make(map[string]string, len(events))
	for _, event := range events {
		if event.SupersedesEventID == nil {
			continue
		}
		target := *event.SupersedesEventID
		if prior, ok := supersederByTarget[target]; ok && prior != event.EventID {
			return fmt.Errorf("%w: event %q is superseded by both %q and %q", ErrInvalidCorporateActionProjection, target, prior, event.EventID)
		}
		supersederByTarget[target] = event.EventID
	}

	for _, event := range events {
		seen := make(map[string]struct{}, len(events))
		current := event
		for current.SupersedesEventID != nil {
			if _, ok := seen[current.EventID]; ok {
				return fmt.Errorf("%w: supersession cycle detected at event %q", ErrInvalidCorporateActionProjection, current.EventID)
			}
			seen[current.EventID] = struct{}{}
			predecessor, ok := byID[*current.SupersedesEventID]
			if !ok {
				break
			}
			if predecessor.InstrumentID != current.InstrumentID {
				return fmt.Errorf("%w: event %q supersedes a different instrument", ErrInvalidCorporateActionProjection, current.EventID)
			}
			if predecessor.Kind != current.Kind {
				return fmt.Errorf("%w: event %q supersedes a different corporate-action kind", ErrInvalidCorporateActionProjection, current.EventID)
			}
			current = predecessor
		}
	}
	return nil
}

func cloneCorporateActionEvent(event CorporateActionEvent) CorporateActionEvent {
	cloned := event
	if event.RecordDate != nil {
		value := *event.RecordDate
		cloned.RecordDate = &value
	}
	if event.PaymentDate != nil {
		value := *event.PaymentDate
		cloned.PaymentDate = &value
	}
	if event.AmountPerUnit != nil {
		value := *event.AmountPerUnit
		cloned.AmountPerUnit = &value
	}
	if event.SupersedesEventID != nil {
		value := *event.SupersedesEventID
		cloned.SupersedesEventID = &value
	}
	return cloned
}
