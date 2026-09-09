package tinvest

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

const (
	productionBaseURL = "https://invest-public-api.tbank.ru/rest"
	providerID        = "T_INVEST_API"

	getDividendsMethod   = "GetDividends"
	getBondCouponsMethod = "GetBondCoupons"
	instrumentsService   = "tinkoff.public.invest.api.contract.v1.InstrumentsService"

	regularCashDividendType = "Regular Cash"
	cancelledDividendType   = "Cancelled"

	requestTimeout               = 5 * time.Second
	maxResponseBodyBytes   int64 = 256 * 1024
	maxConcurrency               = 4
	maxRequestsPerMinute         = 60
	maxInstrumentsPerQuery       = 50
)

type instrumentKind uint8

const (
	instrumentKindShare instrumentKind = iota + 1
	instrumentKindBond
)

type instrumentMapping struct {
	providerInstrumentID string
	kind                 instrumentKind
}

var canonicalInstrumentMappings = map[string]instrumentMapping{
	"SBER": {
		providerInstrumentID: "SBER_TQBR",
		kind:                 instrumentKindShare,
	},
	"GAZP": {
		providerInstrumentID: "GAZP_TQBR",
		kind:                 instrumentKindShare,
	},
	"SU26238RMFS4": {
		providerInstrumentID: "SU26238RMFS4_TQOB",
		kind:                 instrumentKindBond,
	},
}

type Provider struct {
	client      *http.Client
	clock       verticalslice.Clock
	baseURL     *url.URL
	token       string
	semaphore   chan struct{}
	requestGate *requestBudget
}

var _ verticalslice.CorporateActionProvider = (*Provider)(nil)

func NewCorporateActionProvider(client *http.Client, clock verticalslice.Clock, readOnlyToken string) (*Provider, error) {
	return newCorporateActionProvider(client, clock, readOnlyToken, productionBaseURL)
}

func newCorporateActionProvider(client *http.Client, clock verticalslice.Clock, readOnlyToken string, baseURL string) (*Provider, error) {
	if client == nil {
		return nil, errors.New("tinvest: http client is required")
	}
	if isNilDependency(clock) {
		return nil, errors.New("tinvest: clock is required")
	}
	if err := validateToken(readOnlyToken); err != nil {
		return nil, err
	}

	parsedBaseURL, err := url.Parse(baseURL)
	if err != nil || parsedBaseURL.Scheme == "" || parsedBaseURL.Host == "" || parsedBaseURL.User != nil || parsedBaseURL.RawQuery != "" || parsedBaseURL.Fragment != "" {
		return nil, errors.New("tinvest: base URL is invalid")
	}
	if parsedBaseURL.Scheme != "http" && parsedBaseURL.Scheme != "https" {
		return nil, errors.New("tinvest: base URL scheme is invalid")
	}

	clientCopy := *client
	clientCopy.Timeout = requestTimeout
	clientCopy.Jar = nil
	clientCopy.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &Provider{
		client:      &clientCopy,
		clock:       clock,
		baseURL:     parsedBaseURL,
		token:       readOnlyToken,
		semaphore:   make(chan struct{}, maxConcurrency),
		requestGate: newRequestBudget(maxRequestsPerMinute, time.Minute, time.Now),
	}, nil
}

func (provider *Provider) CorporateActions(ctx context.Context, query verticalslice.CorporateActionQuery) ([]verticalslice.CorporateActionEvent, error) {
	if err := verticalslice.ValidateCorporateActionQuery(query); err != nil {
		return nil, err
	}
	if len(query.InstrumentIDs) > maxInstrumentsPerQuery {
		return nil, fmt.Errorf("%w: instrument count exceeds provider limit", verticalslice.ErrInvalidCorporateActionQuery)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	tasks := make([]providerTask, 0, len(query.InstrumentIDs))
	for _, instrumentID := range query.InstrumentIDs {
		mapping, ok := canonicalInstrumentMappings[instrumentID]
		if !ok {
			// Resolve the complete mapping set before performing any network I/O. An
			// unsupported canonical asset must fail closed instead of falling back to
			// an unauthorized discovery endpoint or returning a partial result.
			return nil, providerUnavailableError("canonical instrument mapping is unavailable")
		}
		tasks = append(tasks, providerTask{instrumentID: instrumentID, mapping: mapping})
	}

	from, to := providerWindow(query.From, query.To)
	workerCount := len(tasks)
	if workerCount > maxConcurrency {
		workerCount = maxConcurrency
	}

	jobs := make(chan providerTask)
	results := make(chan providerResult, len(tasks))

	var workers sync.WaitGroup
	workers.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func() {
			defer workers.Done()
			for task := range jobs {
				events, fetchErr := provider.fetchInstrument(ctx, task, from, to)
				results <- providerResult{events: events, err: fetchErr}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for _, task := range tasks {
			select {
			case jobs <- task:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() {
		workers.Wait()
		close(results)
	}()

	var events []verticalslice.CorporateActionEvent
	var providerDataErr bool
	var providerUnavailableErr bool
	var firstContextErr error
	for result := range results {
		if result.err != nil {
			switch {
			case errors.Is(result.err, context.Canceled), errors.Is(result.err, context.DeadlineExceeded):
				if firstContextErr == nil {
					firstContextErr = result.err
				}
			case errors.Is(result.err, verticalslice.ErrCorporateActionsProviderData):
				providerDataErr = true
			default:
				providerUnavailableErr = true
			}
			continue
		}
		events = append(events, result.events...)
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// Fixed class precedence makes the public 502/503 mapping independent of goroutine completion order.
	// The query is already bounded to 50 instruments, so finishing its approved one-call-per-instrument
	// work does not create unbounded fan-out or retries.
	if providerDataErr {
		return nil, providerDataError("one or more provider responses are invalid")
	}
	if providerUnavailableErr {
		return nil, providerUnavailableError("one or more provider requests are unavailable")
	}
	if firstContextErr != nil {
		return nil, firstContextErr
	}

	events = deduplicateEvents(events)
	sort.Slice(events, func(i, j int) bool {
		left, right := events[i], events[j]
		if left.InstrumentID != right.InstrumentID {
			return left.InstrumentID < right.InstrumentID
		}
		if left.Kind != right.Kind {
			return left.Kind < right.Kind
		}
		leftDate, rightDate := effectiveDate(left), effectiveDate(right)
		if leftDate != rightDate {
			return leftDate < rightDate
		}
		return left.EventID < right.EventID
	})
	return events, nil
}

type providerTask struct {
	instrumentID string
	mapping      instrumentMapping
}

type providerResult struct {
	events []verticalslice.CorporateActionEvent
	err    error
}

func (provider *Provider) fetchInstrument(ctx context.Context, task providerTask, from, to string) ([]verticalslice.CorporateActionEvent, error) {
	switch task.mapping.kind {
	case instrumentKindShare:
		return provider.fetchDividends(ctx, task.instrumentID, task.mapping.providerInstrumentID, from, to)
	case instrumentKindBond:
		return provider.fetchBondCoupons(ctx, task.instrumentID, task.mapping.providerInstrumentID, from, to)
	default:
		return nil, providerUnavailableError("canonical instrument kind is unsupported")
	}
}

func (provider *Provider) fetchDividends(ctx context.Context, instrumentID, providerInstrumentID, from, to string) ([]verticalslice.CorporateActionEvent, error) {
	body, err := provider.post(ctx, getDividendsMethod, corporateActionRequest{
		InstrumentID: providerInstrumentID,
		From:         from,
		To:           to,
	})
	if err != nil {
		return nil, err
	}

	var response dividendsResponse
	if err := decodeSingleJSON(body, &response); err != nil {
		return nil, providerDataError("malformed GetDividends response")
	}
	retrievedAt, err := provider.retrievalTimestamp(len(response.Dividends))
	if err != nil {
		return nil, err
	}
	events := make([]verticalslice.CorporateActionEvent, 0, len(response.Dividends))
	for _, item := range response.Dividends {
		status, err := dividendStatus(item.DividendType)
		if err != nil {
			return nil, err
		}
		recordDate, err := optionalBusinessDate(item.RecordDate, "recordDate")
		if err != nil {
			return nil, err
		}
		paymentDate, err := optionalBusinessDate(item.PaymentDate, "paymentDate")
		if err != nil {
			return nil, err
		}
		amount, err := optionalCanonicalMoney(item.DividendNet)
		if err != nil {
			return nil, err
		}

		digest := eventDigest(
			"v2", "DIVIDEND", instrumentID, providerInstrumentID, item.DividendType,
			optionalStringIdentity(recordDate), optionalStringIdentity(paymentDate), optionalMoneyIdentity(amount),
		)
		events = append(events, verticalslice.CorporateActionEvent{
			EventID:       "tinvest:dividend:" + digest,
			InstrumentID:  instrumentID,
			Kind:          verticalslice.CorporateActionDividend,
			Status:        status,
			RecordDate:    cloneStringPointer(recordDate),
			PaymentDate:   cloneStringPointer(paymentDate),
			AmountPerUnit: cloneMoneyPointer(amount),
			AsOf:          retrievedAt,
			RetrievedAt:   retrievedAt,
			Provenance: verticalslice.CorporateActionProvenance{
				Provider:      providerID,
				SourceEventID: "sha256:" + digest,
			},
		})
	}
	return events, nil
}

func (provider *Provider) fetchBondCoupons(ctx context.Context, instrumentID, providerInstrumentID, from, to string) ([]verticalslice.CorporateActionEvent, error) {
	body, err := provider.post(ctx, getBondCouponsMethod, corporateActionRequest{
		InstrumentID: providerInstrumentID,
		From:         from,
		To:           to,
	})
	if err != nil {
		return nil, err
	}

	var response bondCouponsResponse
	if err := decodeSingleJSON(body, &response); err != nil {
		return nil, providerDataError("malformed GetBondCoupons response")
	}
	retrievedAt, err := provider.retrievalTimestamp(len(response.Events))
	if err != nil {
		return nil, err
	}
	events := make([]verticalslice.CorporateActionEvent, 0, len(response.Events))
	for _, item := range response.Events {
		if item.CouponNumber <= 0 {
			return nil, providerDataError("couponNumber must be positive")
		}
		paymentDate, err := optionalBusinessDate(item.CouponDate, "couponDate")
		if err != nil {
			return nil, err
		}
		recordDate, err := optionalBusinessDate(item.FixDate, "fixDate")
		if err != nil {
			return nil, err
		}
		amount, err := optionalCanonicalMoney(item.PayOneBond)
		if err != nil {
			return nil, err
		}

		digest := eventDigest(
			"v2", "COUPON", instrumentID, providerInstrumentID,
			fmt.Sprintf("%d", int64(item.CouponNumber)), optionalStringIdentity(recordDate),
			optionalStringIdentity(paymentDate), optionalMoneyIdentity(amount),
		)
		events = append(events, verticalslice.CorporateActionEvent{
			EventID:       "tinvest:coupon:" + digest,
			InstrumentID:  instrumentID,
			Kind:          verticalslice.CorporateActionCoupon,
			Status:        verticalslice.CorporateActionAnnounced,
			RecordDate:    cloneStringPointer(recordDate),
			PaymentDate:   cloneStringPointer(paymentDate),
			AmountPerUnit: cloneMoneyPointer(amount),
			AsOf:          retrievedAt,
			RetrievedAt:   retrievedAt,
			Provenance: verticalslice.CorporateActionProvenance{
				Provider:      providerID,
				SourceEventID: "sha256:" + digest,
			},
		})
	}
	return events, nil
}

func (provider *Provider) post(ctx context.Context, method string, payload corporateActionRequest) ([]byte, error) {
	if err := provider.acquire(ctx); err != nil {
		return nil, err
	}
	defer func() { <-provider.semaphore }()

	if !provider.requestGate.Allow() {
		return nil, providerUnavailableError("internal provider request budget exhausted")
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, providerDataError("request encoding failed")
	}
	requestURL := provider.baseURL.JoinPath(instrumentsService, method)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL.String(), bytes.NewReader(encoded))
	if err != nil {
		return nil, providerDataError("request construction failed")
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+provider.token)

	response, err := provider.client.Do(request)
	if err != nil {
		if contextErr := ctx.Err(); contextErr != nil {
			return nil, contextErr
		}
		return nil, providerUnavailableError("provider request failed")
	}
	defer response.Body.Close()
	provider.requestGate.ObserveProviderHeaders(response.Header)

	switch {
	case response.StatusCode == http.StatusUnauthorized,
		response.StatusCode == http.StatusForbidden,
		response.StatusCode == http.StatusRequestTimeout,
		response.StatusCode == http.StatusTooManyRequests,
		response.StatusCode >= http.StatusInternalServerError:
		return nil, providerUnavailableError(fmt.Sprintf("provider returned HTTP %d", response.StatusCode))
	case response.StatusCode != http.StatusOK:
		return nil, providerDataError(fmt.Sprintf("provider returned HTTP %d", response.StatusCode))
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBodyBytes+1))
	if err != nil {
		if contextErr := ctx.Err(); contextErr != nil {
			return nil, contextErr
		}
		return nil, providerUnavailableError("provider response read failed")
	}
	if int64(len(body)) > maxResponseBodyBytes {
		return nil, providerDataError("provider response body exceeds 256 KiB")
	}
	return body, nil
}

func (provider *Provider) acquire(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	select {
	case provider.semaphore <- struct{}{}:
		return nil
	default:
		return providerUnavailableError("provider concurrency limit exhausted")
	}
}

func providerWindow(from, to string) (string, string) {
	// ValidateCorporateActionQuery already proved both values are canonical YYYY-MM-DD and from <= to.
	return from + "T00:00:00Z", to + "T23:59:59.999999999Z"
}

func optionalBusinessDate(raw, field string) (*string, error) {
	if raw == "" {
		return nil, nil
	}
	value, err := parseProviderTimestamp(raw, field)
	if err != nil {
		return nil, err
	}
	return stringPointer(value), nil
}

func parseProviderTimestamp(raw, field string) (string, error) {
	if raw == "" {
		return "", nil
	}
	if strings.TrimSpace(raw) != raw {
		return "", providerDataError(field + " is invalid")
	}
	parsed, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		return "", providerDataError(field + " is invalid")
	}
	_, offset := parsed.Zone()
	if offset != 0 {
		return "", providerDataError(field + " is not UTC")
	}
	return parsed.UTC().Format("2006-01-02"), nil
}

func (provider *Provider) retrievalTimestamp(eventCount int) (time.Time, error) {
	if eventCount == 0 {
		return time.Time{}, nil
	}
	retrievedAt := provider.clock.Now().UTC()
	if retrievedAt.IsZero() {
		return time.Time{}, providerDataError("clock returned a zero timestamp")
	}
	return retrievedAt, nil
}

func dividendStatus(dividendType string) (verticalslice.CorporateActionStatus, error) {
	if dividendType == "" || strings.TrimSpace(dividendType) != dividendType {
		return "", providerDataError("dividend type is invalid")
	}
	switch dividendType {
	case regularCashDividendType:
		return verticalslice.CorporateActionAnnounced, nil
	case cancelledDividendType:
		return verticalslice.CorporateActionCancelled, nil
	default:
		return "", providerDataError("dividend type is outside the canonical DIVIDEND scope")
	}
}

func optionalCanonicalMoney(value *moneyValue) (*verticalslice.Money, error) {
	if value == nil {
		return nil, nil
	}
	money, err := canonicalMoney(value)
	if err != nil {
		return nil, err
	}
	return &money, nil
}

func canonicalMoney(value *moneyValue) (verticalslice.Money, error) {
	if value == nil {
		return verticalslice.Money{}, providerDataError("money value is missing")
	}
	currency := strings.ToUpper(value.Currency)
	if value.Currency == "" || strings.TrimSpace(value.Currency) != value.Currency || len(currency) != 3 {
		return verticalslice.Money{}, providerDataError("money currency is invalid")
	}
	for _, r := range currency {
		if r < 'A' || r > 'Z' {
			return verticalslice.Money{}, providerDataError("money currency is invalid")
		}
	}
	if value.Nano < -999999999 || value.Nano > 999999999 {
		return verticalslice.Money{}, providerDataError("money nano is outside protobuf bounds")
	}
	if (value.Units > 0 && value.Nano < 0) || (value.Units < 0 && value.Nano > 0) {
		return verticalslice.Money{}, providerDataError("money units and nano have inconsistent signs")
	}

	totalNano := new(big.Int).Mul(big.NewInt(int64(value.Units)), big.NewInt(1_000_000_000))
	totalNano.Add(totalNano, big.NewInt(int64(value.Nano)))
	if totalNano.Sign() < 0 {
		return verticalslice.Money{}, providerDataError("money amount is negative")
	}

	quotient, remainder := new(big.Int).QuoRem(totalNano, big.NewInt(10), new(big.Int))
	if remainder.Sign() != 0 {
		// OpenInvest's canonical Decimal scale is 8 while MoneyValue nano is scale 9.
		// Reject values that would require rounding rather than silently changing
		// financial truth.
		return verticalslice.Money{}, providerDataError("money amount exceeds canonical decimal scale")
	}
	canonicalText := decimalTextFromScale8(quotient)
	amount, err := decimal.FromString(canonicalText)
	if err != nil || !amount.FitsStorage() {
		return verticalslice.Money{}, providerDataError("money amount does not fit canonical decimal storage")
	}
	return verticalslice.Money{Amount: amount, Currency: currency}, nil
}

func decimalTextFromScale8(unscaled *big.Int) string {
	text := new(big.Int).Set(unscaled).Text(10)
	if len(text) <= decimal.Scale {
		text = strings.Repeat("0", decimal.Scale-len(text)+1) + text
	}
	whole := text[:len(text)-decimal.Scale]
	fraction := text[len(text)-decimal.Scale:]
	return whole + "." + fraction
}

func decodeSingleJSON(body []byte, destination any) error {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return errors.New("response must be a JSON object")
	}
	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("trailing JSON value")
		}
		return err
	}
	return nil
}

func deduplicateEvents(events []verticalslice.CorporateActionEvent) []verticalslice.CorporateActionEvent {
	seen := make(map[string]struct{}, len(events))
	deduplicated := make([]verticalslice.CorporateActionEvent, 0, len(events))
	for _, event := range events {
		if _, ok := seen[event.EventID]; ok {
			continue
		}
		seen[event.EventID] = struct{}{}
		deduplicated = append(deduplicated, event)
	}
	return deduplicated
}

func optionalStringIdentity(value *string) string {
	if value == nil {
		return "<nil>"
	}
	return *value
}

func optionalMoneyIdentity(value *verticalslice.Money) string {
	if value == nil {
		return "<nil>"
	}
	return value.Amount.String() + "|" + value.Currency
}

func cloneMoneyPointer(value *verticalslice.Money) *verticalslice.Money {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func effectiveDate(event verticalslice.CorporateActionEvent) string {
	if event.RecordDate != nil {
		return *event.RecordDate
	}
	if event.PaymentDate != nil {
		return *event.PaymentDate
	}
	return ""
}

func eventDigest(parts ...string) string {
	hash := sha256.New()
	for _, part := range parts {
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write([]byte(part))
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func validateToken(token string) error {
	if token == "" || strings.TrimSpace(token) != token {
		return errors.New("tinvest: read-only token is required")
	}
	for _, r := range token {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return errors.New("tinvest: read-only token is invalid")
		}
	}
	return nil
}

func providerUnavailableError(reason string) error {
	return fmt.Errorf("%w: %s", verticalslice.ErrCorporateActionsProviderUnavailable, reason)
}

func providerDataError(reason string) error {
	return fmt.Errorf("%w: %s", verticalslice.ErrCorporateActionsProviderData, reason)
}

func stringPointer(value string) *string {
	cloned := value
	return &cloned
}

func cloneStringPointer(value *string) *string {
	if value == nil {
		return nil
	}
	return stringPointer(*value)
}

func isNilDependency(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

type requestBudget struct {
	mu              sync.Mutex
	limit           int
	window          time.Duration
	now             func() time.Time
	used            []time.Time
	remoteRemaining int
	remoteResetAt   time.Time
}

func newRequestBudget(limit int, window time.Duration, now func() time.Time) *requestBudget {
	return &requestBudget{limit: limit, window: window, now: now}
}

func (budget *requestBudget) Allow() bool {
	budget.mu.Lock()
	defer budget.mu.Unlock()

	now := budget.now()
	budget.expireRemoteLimit(now)
	cutoff := now.Add(-budget.window)
	firstLive := 0
	for firstLive < len(budget.used) && !budget.used[firstLive].After(cutoff) {
		firstLive++
	}
	if firstLive > 0 {
		copy(budget.used, budget.used[firstLive:])
		budget.used = budget.used[:len(budget.used)-firstLive]
	}
	if len(budget.used) >= budget.limit {
		return false
	}
	if !budget.remoteResetAt.IsZero() && budget.remoteRemaining <= 0 {
		return false
	}
	budget.used = append(budget.used, now)
	if !budget.remoteResetAt.IsZero() {
		budget.remoteRemaining--
	}
	return true
}

func (budget *requestBudget) ObserveProviderHeaders(header http.Header) {
	remainingRaw := strings.TrimSpace(header.Get("X-RateLimit-Remaining"))
	resetRaw := strings.TrimSpace(header.Get("X-RateLimit-Reset"))
	if remainingRaw == "" || resetRaw == "" {
		return
	}
	remaining, err := strconv.Atoi(remainingRaw)
	if err != nil || remaining < 0 {
		return
	}
	resetSeconds, err := strconv.Atoi(resetRaw)
	if err != nil || resetSeconds < 0 || resetSeconds > 3600 {
		return
	}
	budget.mu.Lock()
	defer budget.mu.Unlock()
	now := budget.now()
	budget.expireRemoteLimit(now)
	if resetSeconds == 0 {
		return
	}
	resetAt := now.Add(time.Duration(resetSeconds) * time.Second)
	if budget.remoteResetAt.IsZero() {
		budget.remoteRemaining = remaining
		budget.remoteResetAt = resetAt
		return
	}
	if remaining < budget.remoteRemaining {
		budget.remoteRemaining = remaining
	}
	if resetAt.After(budget.remoteResetAt) {
		budget.remoteResetAt = resetAt
	}
}

func (budget *requestBudget) expireRemoteLimit(now time.Time) {
	if !budget.remoteResetAt.IsZero() && !now.Before(budget.remoteResetAt) {
		budget.remoteRemaining = 0
		budget.remoteResetAt = time.Time{}
	}
}
