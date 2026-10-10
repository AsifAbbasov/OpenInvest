package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

const maxCorporateActionInstruments = 50

func (api *API) getCorporateActionProjection(c fiber.Ctx) error {
	c.Set("Cache-Control", "no-store")
	instrumentIDs, err := corporateActionInstrumentIDs(c)
	if err != nil {
		return writeCorporateActionProjectionError(c, err)
	}
	query := verticalslice.CorporateActionQuery{
		InstrumentIDs: instrumentIDs,
		From:          c.Query("from"),
		To:            c.Query("to"),
	}
	sort.Strings(query.InstrumentIDs)
	if err := verticalslice.ValidateCorporateActionQuery(query); err != nil {
		return writeCorporateActionProjectionError(c, err)
	}
	if err := api.admitCorporateActionProjection(c); err != nil {
		return writeCorporateActionProjectionError(c, err)
	}

	projection, err := api.loadCorporateActionProjection(c.Context(), query)
	if err != nil {
		return writeCorporateActionProjectionError(c, err)
	}
	return writeOK(c, projection)
}

func corporateActionInstrumentIDs(c fiber.Ctx) ([]string, error) {
	raw := c.Query("instrumentId")
	if raw == "" {
		return nil, nil
	}
	instrumentIDs := strings.Split(raw, ",")
	if len(instrumentIDs) > maxCorporateActionInstruments {
		return nil, fmt.Errorf("%w: instrumentId supports at most %d instruments", verticalslice.ErrInvalidCorporateActionQuery, maxCorporateActionInstruments)
	}
	return instrumentIDs, nil
}

func writeCorporateActionProjectionError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, errCorporateActionProjectionRateLimited):
		c.Set("Retry-After", corporateActionProjectionRateLimitRetryAfterSeconds)
		return writeError(c, http.StatusTooManyRequests, "RATE_LIMITED", "Too many corporate actions requests")
	case errors.Is(err, errCorporateActionProjectionAdmissionUnavailable):
		return writeError(c, http.StatusServiceUnavailable, "CORPORATE_ACTIONS_SOURCE_UNAVAILABLE", "Corporate actions source is unavailable")
	case errors.Is(err, verticalslice.ErrInvalidCorporateActionQuery):
		return writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, verticalslice.ErrCorporateActionsProviderUnavailable):
		return writeError(c, http.StatusServiceUnavailable, "CORPORATE_ACTIONS_SOURCE_UNAVAILABLE", "Corporate actions source is unavailable")
	case errors.Is(err, verticalslice.ErrCorporateActionsProviderData),
		errors.Is(err, verticalslice.ErrInvalidCorporateAction),
		errors.Is(err, verticalslice.ErrInvalidCorporateActionProjection):
		return writeError(c, http.StatusBadGateway, "CORPORATE_ACTIONS_SOURCE_INVALID", "Corporate actions source returned unverifiable data")
	default:
		return writeMappedError(c, err)
	}
}

type corporateActionProjectionDTO struct {
	Calendar []corporateActionCalendarEntryDTO `json:"calendar"`
	Heatmap  []corporateActionHeatmapBucketDTO `json:"heatmap"`
	Coverage corporateActionCoverageDTO        `json:"coverage"`
}

type corporateActionCoverageDTO struct {
	InputMode     string   `json:"inputMode"`
	InstrumentIDs []string `json:"instrumentIds"`
	From          string   `json:"from"`
	To            string   `json:"to"`
}

type corporateActionCalendarEntryDTO struct {
	EffectiveDate string                  `json:"effectiveDate"`
	Event         corporateActionEventDTO `json:"event"`
}

type corporateActionEventDTO struct {
	EventID           string                       `json:"eventId"`
	InstrumentID      string                       `json:"instrumentId"`
	Kind              string                       `json:"kind"`
	Status            string                       `json:"status"`
	RecordDate        *string                      `json:"recordDate"`
	PaymentDate       *string                      `json:"paymentDate"`
	AmountPerUnit     *moneyDTO                    `json:"amountPerUnit"`
	SupersedesEventID *string                      `json:"supersedesEventId"`
	AsOf              string                       `json:"asOf"`
	RetrievedAt       string                       `json:"retrievedAt"`
	Provenance        corporateActionProvenanceDTO `json:"provenance"`
}

type corporateActionProvenanceDTO struct {
	Provider string `json:"provider"`
}

type corporateActionHeatmapBucketDTO struct {
	Date           string `json:"date"`
	TotalCount     int    `json:"totalCount"`
	DividendCount  int    `json:"dividendCount"`
	CouponCount    int    `json:"couponCount"`
	AnnouncedCount int    `json:"announcedCount"`
	ConfirmedCount int    `json:"confirmedCount"`
	PaidCount      int    `json:"paidCount"`
	CancelledCount int    `json:"cancelledCount"`
}

func mapCorporateActionCalendar(entries []verticalslice.CorporateActionCalendarEntry) []corporateActionCalendarEntryDTO {
	mapped := make([]corporateActionCalendarEntryDTO, 0, len(entries))
	for _, entry := range entries {
		mapped = append(mapped, corporateActionCalendarEntryDTO{
			EffectiveDate: entry.EffectiveDate,
			Event:         mapCorporateActionEvent(entry.Event),
		})
	}
	return mapped
}

func mapCorporateActionEvent(event verticalslice.CorporateActionEvent) corporateActionEventDTO {
	return corporateActionEventDTO{
		EventID:           event.EventID,
		InstrumentID:      event.InstrumentID,
		Kind:              string(event.Kind),
		Status:            string(event.Status),
		RecordDate:        cloneOptionalString(event.RecordDate),
		PaymentDate:       cloneOptionalString(event.PaymentDate),
		AmountPerUnit:     mapOptionalMoney(event.AmountPerUnit),
		SupersedesEventID: cloneOptionalString(event.SupersedesEventID),
		AsOf:              event.AsOf.UTC().Format(time.RFC3339),
		RetrievedAt:       event.RetrievedAt.UTC().Format(time.RFC3339),
		Provenance: corporateActionProvenanceDTO{
			Provider: event.Provenance.Provider,
		},
	}
}

func mapCorporateActionHeatmap(buckets []verticalslice.CorporateActionHeatmapBucket) []corporateActionHeatmapBucketDTO {
	mapped := make([]corporateActionHeatmapBucketDTO, 0, len(buckets))
	for _, bucket := range buckets {
		mapped = append(mapped, corporateActionHeatmapBucketDTO{
			Date:           bucket.Date,
			TotalCount:     bucket.TotalCount,
			DividendCount:  bucket.DividendCount,
			CouponCount:    bucket.CouponCount,
			AnnouncedCount: bucket.AnnouncedCount,
			ConfirmedCount: bucket.ConfirmedCount,
			PaidCount:      bucket.PaidCount,
			CancelledCount: bucket.CancelledCount,
		})
	}
	return mapped
}

func filterCorporateActionCalendarWindow(entries []verticalslice.CorporateActionCalendarEntry, from, to string) []verticalslice.CorporateActionCalendarEntry {
	filtered := make([]verticalslice.CorporateActionCalendarEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.EffectiveDate >= from && entry.EffectiveDate <= to {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func filterCorporateActionHeatmapWindow(buckets []verticalslice.CorporateActionHeatmapBucket, from, to string) []verticalslice.CorporateActionHeatmapBucket {
	filtered := make([]verticalslice.CorporateActionHeatmapBucket, 0, len(buckets))
	for _, bucket := range buckets {
		if bucket.Date >= from && bucket.Date <= to {
			filtered = append(filtered, bucket)
		}
	}
	return filtered
}

func cloneOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
