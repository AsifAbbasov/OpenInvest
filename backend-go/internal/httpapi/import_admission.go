package httpapi

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
)

var (
	errImportAdmissionUnavailable = errors.New("import admission is unavailable")
	errImportCapacityExhausted    = errors.New("import processing capacity exhausted")
	errImportAdmissionExhausted   = errors.New("import execution admission exhausted")
	errImportRateLimited          = errors.New("import fresh-command rate limited")
)

// importAdmission applies three independent bounds to both import routes: active capacity, every
// admitted execution attempt, and commands proven fresh after replay resolution.
type importAdmission struct {
	executionRate importRate
	freshRate     importRate
	capacity      chan struct{}
}

type importRatePolicy struct {
	perSubject  int
	globalLimit int
	maxSubjects int
	window      time.Duration
}

type importRate struct {
	mu             sync.Mutex
	perSubject     int
	globalLimit    int
	maxSubjects    int
	window         time.Duration
	attempts       map[string][]time.Time
	globalAttempts []time.Time
	lastSweep      time.Time
}

func newDefaultImportAdmission() *importAdmission {
	return newImportAdmissionWithPolicies(defaultImportExecutionRatePolicy(), defaultImportFreshRatePolicy(), defaultImportActiveCapacity)
}

// newImportAdmission retains a concise fresh-policy constructor for focused legacy tests while
// production construction always provides both explicit policies through the default factory.
func newImportAdmission(perSubject int, globalLimit int, maxSubjects int, window time.Duration, capacity int) *importAdmission {
	return newImportAdmissionWithPolicies(defaultImportExecutionRatePolicy(), importRatePolicy{
		perSubject: perSubject, globalLimit: globalLimit, maxSubjects: maxSubjects, window: window,
	}, capacity)
}

func newImportAdmissionWithPolicies(execution importRatePolicy, fresh importRatePolicy, capacity int) *importAdmission {
	return &importAdmission{
		executionRate: newImportRate(execution),
		freshRate:     newImportRate(fresh),
		capacity:      make(chan struct{}, nonNegativeInt(capacity)),
	}
}

func defaultImportExecutionRatePolicy() importRatePolicy {
	return importRatePolicy{
		perSubject:  defaultImportExecutionPerSubjectLimit,
		globalLimit: defaultImportExecutionGlobalLimit,
		maxSubjects: defaultImportExecutionMaxSubjects,
		window:      defaultImportExecutionWindow,
	}
}

func defaultImportFreshRatePolicy() importRatePolicy {
	return importRatePolicy{
		perSubject:  defaultImportFreshPerSubjectLimit,
		globalLimit: defaultImportFreshGlobalLimit,
		maxSubjects: defaultImportFreshMaxSubjects,
		window:      defaultImportFreshWindow,
	}
}

func newImportRate(policy importRatePolicy) importRate {
	return importRate{
		perSubject:  nonNegativeInt(policy.perSubject),
		globalLimit: nonNegativeInt(policy.globalLimit),
		maxSubjects: nonNegativeInt(policy.maxSubjects),
		window:      policy.window,
		attempts:    map[string][]time.Time{},
	}
}

func (admission *importAdmission) acquire() (func(), error) {
	if admission == nil || cap(admission.capacity) == 0 {
		return nil, errImportAdmissionUnavailable
	}

	select {
	case admission.capacity <- struct{}{}:
		var once sync.Once
		return func() {
			once.Do(func() { <-admission.capacity })
		}, nil
	default:
		return nil, errImportCapacityExhausted
	}
}

func (admission *importAdmission) allowExecution(subjectID string, now time.Time) bool {
	if admission == nil {
		return false
	}
	return admission.executionRate.allow(subjectID, now)
}

func (admission *importAdmission) allowFresh(subjectID string, now time.Time) bool {
	if admission == nil {
		return false
	}
	return admission.freshRate.allow(subjectID, now)
}

func (rate *importRate) allow(subjectID string, now time.Time) bool {
	if strings.TrimSpace(subjectID) == "" {
		return false
	}

	rate.mu.Lock()
	defer rate.mu.Unlock()

	if rate.window <= 0 || rate.perSubject == 0 || rate.globalLimit == 0 || rate.maxSubjects == 0 {
		return false
	}

	cutoff := now.Add(-rate.window)
	rate.globalAttempts = retainImportAttempts(rate.globalAttempts, cutoff)

	attempts, exists := rate.attempts[subjectID]
	if exists {
		attempts = retainImportAttempts(attempts, cutoff)
		if len(attempts) == 0 {
			delete(rate.attempts, subjectID)
			exists = false
		} else {
			rate.attempts[subjectID] = attempts
		}
	}

	if !exists && (len(rate.attempts) >= rate.maxSubjects || rate.shouldSweep(now)) {
		rate.sweepExpired(cutoff, now)
		attempts = rate.attempts[subjectID]
		exists = len(attempts) > 0
	}

	if len(rate.globalAttempts) >= rate.globalLimit || (!exists && len(rate.attempts) >= rate.maxSubjects) || len(attempts) >= rate.perSubject {
		return false
	}

	rate.attempts[subjectID] = append(attempts, now)
	rate.globalAttempts = append(rate.globalAttempts, now)
	return true
}

func (rate *importRate) shouldSweep(now time.Time) bool {
	return rate.lastSweep.IsZero() || !now.Before(rate.lastSweep.Add(rate.window))
}

func (rate *importRate) sweepExpired(cutoff time.Time, now time.Time) {
	for subjectID, attempts := range rate.attempts {
		retained := retainImportAttempts(attempts, cutoff)
		if len(retained) == 0 {
			delete(rate.attempts, subjectID)
			continue
		}
		rate.attempts[subjectID] = retained
	}
	rate.lastSweep = now
}

func retainImportAttempts(attempts []time.Time, cutoff time.Time) []time.Time {
	retained := attempts[:0]
	for _, attempt := range attempts {
		if attempt.After(cutoff) {
			retained = append(retained, attempt)
		}
	}
	return retained
}

func (api *API) acquireImportCapacity() (func(), error) {
	if api.importAdmission == nil {
		return nil, errImportAdmissionUnavailable
	}
	return api.importAdmission.acquire()
}

func (api *API) admitFreshImport(subjectID string) error {
	if api.importAdmission == nil {
		return errImportAdmissionUnavailable
	}
	if !api.importAdmission.allowFresh(subjectID, api.nowUTC()) {
		return errImportRateLimited
	}
	return nil
}

func (api *API) admitImportExecution(subjectID string) error {
	if api.importAdmission == nil {
		return errImportAdmissionUnavailable
	}
	if !api.importAdmission.allowExecution(subjectID, api.nowUTC()) {
		return errImportAdmissionExhausted
	}
	return nil
}

func writeImportAdmissionError(c fiber.Ctx, meta metaDTO, err error) error {
	switch {
	case errors.Is(err, errImportRateLimited):
		c.Set("Retry-After", importFreshRateRetryAfterSeconds)
		return writeErrorWithMeta(c, meta, 429, "RATE_LIMITED", "Too many fresh import requests")
	case errors.Is(err, errImportAdmissionExhausted):
		c.Set("Retry-After", importExecutionRateRetryAfterSeconds)
		return writeErrorWithMeta(c, meta, 503, "IMPORT_ADMISSION_EXHAUSTED", "Import admission is temporarily exhausted")
	case errors.Is(err, errImportCapacityExhausted), errors.Is(err, errImportAdmissionUnavailable):
		c.Set("Retry-After", importCapacityRetryAfterSeconds)
		return writeErrorWithMeta(c, meta, 503, "IMPORT_CAPACITY_EXHAUSTED", "Import processing capacity is temporarily exhausted")
	default:
		return writeMappedErrorWithMeta(c, meta, err)
	}
}
