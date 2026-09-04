package verticalslice

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"
)

var (
	ErrCorporateActionsProviderUnavailable = errors.New("corporate actions provider unavailable")
	ErrCorporateActionsProviderData        = errors.New("corporate actions provider data invalid")
	ErrInvalidCorporateAction              = errors.New("invalid corporate action")
	ErrInvalidCorporateActionQuery         = errors.New("invalid corporate action query")
)

var (
	corporateActionEventIDPattern  = regexp.MustCompile(`^[A-Za-z0-9._:-]{1,128}$`)
	corporateActionProviderPattern = regexp.MustCompile(`^[A-Z0-9_]{1,64}$`)
	currencyPattern                = regexp.MustCompile(`^[A-Z]{3}$`)
)

type CorporateActionKind string

const (
	CorporateActionDividend CorporateActionKind = "DIVIDEND"
	CorporateActionCoupon   CorporateActionKind = "COUPON"
)

type CorporateActionStatus string

const (
	CorporateActionAnnounced CorporateActionStatus = "ANNOUNCED"
	CorporateActionConfirmed CorporateActionStatus = "CONFIRMED"
	CorporateActionPaid      CorporateActionStatus = "PAID"
	CorporateActionCancelled CorporateActionStatus = "CANCELLED"
)

type CorporateActionProvenance struct {
	Provider      string
	SourceEventID string
}

type CorporateActionEvent struct {
	EventID           string
	InstrumentID      string
	Kind              CorporateActionKind
	Status            CorporateActionStatus
	RecordDate        *string
	PaymentDate       *string
	AmountPerUnit     *Money
	SupersedesEventID *string
	AsOf              time.Time
	RetrievedAt       time.Time
	Provenance        CorporateActionProvenance
}

type CorporateActionQuery struct {
	InstrumentIDs []string
	From          string
	To            string
}

type CorporateActionProvider interface {
	CorporateActions(ctx context.Context, query CorporateActionQuery) ([]CorporateActionEvent, error)
}

func FetchCorporateActions(ctx context.Context, provider CorporateActionProvider, query CorporateActionQuery) ([]CorporateActionEvent, error) {
	if err := ValidateCorporateActionQuery(query); err != nil {
		return nil, err
	}
	if isNilCorporateActionProvider(provider) {
		return nil, ErrCorporateActionsProviderUnavailable
	}
	providerQuery := query
	providerQuery.InstrumentIDs = append([]string(nil), query.InstrumentIDs...)
	events, err := provider.CorporateActions(ctx, providerQuery)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			return nil, err
		case errors.Is(err, ErrCorporateActionsProviderUnavailable), errors.Is(err, ErrCorporateActionsProviderData):
			return nil, err
		default:
			return nil, ErrCorporateActionsProviderUnavailable
		}
	}
	if err := ValidateCorporateActionEvents(events); err != nil {
		return nil, err
	}
	requested := make(map[string]struct{}, len(query.InstrumentIDs))
	for _, instrumentID := range query.InstrumentIDs {
		requested[instrumentID] = struct{}{}
	}
	for i, event := range events {
		if _, ok := requested[event.InstrumentID]; !ok {
			return nil, fmt.Errorf("event[%d]: %w: returned instrument is outside requested set", i, ErrInvalidCorporateAction)
		}
	}
	return events, nil
}

func ValidateCorporateActionQuery(query CorporateActionQuery) error {
	if len(query.InstrumentIDs) == 0 {
		return fmt.Errorf("%w: at least one instrument is required", ErrInvalidCorporateActionQuery)
	}
	seen := make(map[string]struct{}, len(query.InstrumentIDs))
	for _, instrumentID := range query.InstrumentIDs {
		if !tickerPattern.MatchString(instrumentID) {
			return fmt.Errorf("%w: instrumentID is invalid", ErrInvalidCorporateActionQuery)
		}
		if _, ok := seen[instrumentID]; ok {
			return fmt.Errorf("%w: duplicate instrumentID", ErrInvalidCorporateActionQuery)
		}
		seen[instrumentID] = struct{}{}
	}
	from, err := parseCorporateActionDate(query.From, "from")
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidCorporateActionQuery, err)
	}
	to, err := parseCorporateActionDate(query.To, "to")
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidCorporateActionQuery, err)
	}
	if to.Before(from) {
		return fmt.Errorf("%w: to must not be before from", ErrInvalidCorporateActionQuery)
	}
	return nil
}

func ValidateCorporateActionEvent(event CorporateActionEvent) error {
	if !corporateActionEventIDPattern.MatchString(event.EventID) {
		return fmt.Errorf("%w: eventID is invalid", ErrInvalidCorporateAction)
	}
	if !tickerPattern.MatchString(event.InstrumentID) {
		return fmt.Errorf("%w: instrumentID is invalid", ErrInvalidCorporateAction)
	}
	switch event.Kind {
	case CorporateActionDividend, CorporateActionCoupon:
	default:
		return fmt.Errorf("%w: kind is invalid", ErrInvalidCorporateAction)
	}
	switch event.Status {
	case CorporateActionAnnounced, CorporateActionConfirmed, CorporateActionPaid, CorporateActionCancelled:
	default:
		return fmt.Errorf("%w: status is invalid", ErrInvalidCorporateAction)
	}
	if err := validateOptionalCorporateActionDate(event.RecordDate, "recordDate"); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidCorporateAction, err)
	}
	if err := validateOptionalCorporateActionDate(event.PaymentDate, "paymentDate"); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidCorporateAction, err)
	}
	if event.AmountPerUnit != nil {
		if !event.AmountPerUnit.Amount.FitsStorage() || event.AmountPerUnit.Amount.IsNegative() {
			return fmt.Errorf("%w: amountPerUnit must be a non-negative canonical decimal", ErrInvalidCorporateAction)
		}
		if !currencyPattern.MatchString(event.AmountPerUnit.Currency) {
			return fmt.Errorf("%w: amountPerUnit currency must be a canonical three-letter code", ErrInvalidCorporateAction)
		}
	}
	if event.SupersedesEventID != nil {
		if !corporateActionEventIDPattern.MatchString(*event.SupersedesEventID) || *event.SupersedesEventID == event.EventID {
			return fmt.Errorf("%w: supersedesEventID is invalid", ErrInvalidCorporateAction)
		}
	}
	if event.AsOf.IsZero() || event.RetrievedAt.IsZero() {
		return fmt.Errorf("%w: asOf and retrievedAt are required", ErrInvalidCorporateAction)
	}
	if event.AsOf.Location() != time.UTC || event.RetrievedAt.Location() != time.UTC {
		return fmt.Errorf("%w: asOf and retrievedAt must be normalized to UTC", ErrInvalidCorporateAction)
	}
	if !corporateActionProviderPattern.MatchString(event.Provenance.Provider) {
		return fmt.Errorf("%w: provenance provider is invalid", ErrInvalidCorporateAction)
	}
	if err := validateOpaqueCorporateActionID(event.Provenance.SourceEventID, 256, "provenance sourceEventID"); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidCorporateAction, err)
	}
	return nil
}

func ValidateCorporateActionEvents(events []CorporateActionEvent) error {
	seen := make(map[string]struct{}, len(events))
	for i, event := range events {
		if err := ValidateCorporateActionEvent(event); err != nil {
			return fmt.Errorf("event[%d]: %w", i, err)
		}
		if _, ok := seen[event.EventID]; ok {
			return fmt.Errorf("event[%d]: %w: duplicate eventID", i, ErrInvalidCorporateAction)
		}
		seen[event.EventID] = struct{}{}
	}
	return nil
}

func parseCorporateActionDate(value string, field string) (time.Time, error) {
	if value == "" || strings.TrimSpace(value) != value {
		return time.Time{}, fmt.Errorf("%s must be canonical YYYY-MM-DD", field)
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil || parsed.Format("2006-01-02") != value {
		return time.Time{}, fmt.Errorf("%s must be canonical YYYY-MM-DD", field)
	}
	return parsed, nil
}

func validateOptionalCorporateActionDate(value *string, field string) error {
	if value == nil {
		return nil
	}
	_, err := parseCorporateActionDate(*value, field)
	return err
}

func isNilCorporateActionProvider(provider CorporateActionProvider) bool {
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

func validateOpaqueCorporateActionID(value string, maxBytes int, field string) error {
	if value == "" || strings.TrimSpace(value) != value || len(value) > maxBytes {
		return fmt.Errorf("%s is invalid", field)
	}
	for _, r := range value {
		if r < 0x20 || r == 0x7f {
			return fmt.Errorf("%s is invalid", field)
		}
	}
	return nil
}
