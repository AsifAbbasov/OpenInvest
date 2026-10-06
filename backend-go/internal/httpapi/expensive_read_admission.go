package httpapi

import (
	"errors"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
)

var (
	errExpensiveReadRateLimited       = errors.New("expensive read subject rate limited")
	errExpensiveReadGlobalExhausted   = errors.New("expensive read process admission exhausted")
	errExpensiveReadCapacityExhausted = errors.New("expensive read capacity exhausted")
	errExpensiveReadUnavailable       = errors.New("expensive read admission unavailable")
)

// expensiveReadAdmission is intentionally process-local safety admission.
// Deployment-global abuse control is a separate ownership boundary.
type expensiveReadAdmission struct {
	rate     importRate
	capacity chan struct{}
}

func newDefaultExpensiveReadAdmission() *expensiveReadAdmission {
	return &expensiveReadAdmission{
		rate: newImportRate(importRatePolicy{
			perSubject:  defaultExpensiveReadPerSubjectLimit,
			globalLimit: defaultExpensiveReadGlobalLimit,
			maxSubjects: defaultExpensiveReadMaxSubjects,
			window:      defaultExpensiveReadWindow,
		}),
		capacity: make(chan struct{}, defaultExpensiveReadActiveCapacity),
	}
}

func (admission *expensiveReadAdmission) acquire(subjectID string, now time.Time) (func(), error) {
	if admission == nil || cap(admission.capacity) == 0 {
		return nil, errExpensiveReadUnavailable
	}

	admission.rate.mu.Lock()
	if admission.rate.window <= 0 || admission.rate.perSubject == 0 || admission.rate.globalLimit == 0 || admission.rate.maxSubjects == 0 {
		admission.rate.mu.Unlock()
		return nil, errExpensiveReadUnavailable
	}

	cutoff := now.Add(-admission.rate.window)
	admission.rate.globalAttempts = retainImportAttempts(admission.rate.globalAttempts, cutoff)
	attempts, exists := admission.rate.attempts[subjectID]
	if exists {
		attempts = retainImportAttempts(attempts, cutoff)
		if len(attempts) == 0 {
			delete(admission.rate.attempts, subjectID)
			exists = false
		} else {
			admission.rate.attempts[subjectID] = attempts
		}
	}
	if !exists && (len(admission.rate.attempts) >= admission.rate.maxSubjects || admission.rate.shouldSweep(now)) {
		admission.rate.sweepExpired(cutoff, now)
		attempts = admission.rate.attempts[subjectID]
		exists = len(attempts) > 0
	}
	switch {
	case len(attempts) >= admission.rate.perSubject:
		admission.rate.mu.Unlock()
		return nil, errExpensiveReadRateLimited
	case len(admission.rate.globalAttempts) >= admission.rate.globalLimit || (!exists && len(admission.rate.attempts) >= admission.rate.maxSubjects):
		admission.rate.mu.Unlock()
		return nil, errExpensiveReadGlobalExhausted
	}

	select {
	case admission.capacity <- struct{}{}:
		admission.rate.attempts[subjectID] = append(attempts, now)
		admission.rate.globalAttempts = append(admission.rate.globalAttempts, now)
		admission.rate.mu.Unlock()
		var once sync.Once
		return func() {
			once.Do(func() { <-admission.capacity })
		}, nil
	default:
		admission.rate.mu.Unlock()
		return nil, errExpensiveReadCapacityExhausted
	}
}

func (api *API) acquireExpensiveRead(subjectID string) (func(), error) {
	if api == nil || api.expensiveReadAdmission == nil {
		return nil, errExpensiveReadUnavailable
	}
	return api.expensiveReadAdmission.acquire(subjectID, api.nowUTC())
}

func writeExpensiveReadAdmissionError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, errExpensiveReadRateLimited):
		c.Set("Retry-After", expensiveReadRateRetryAfterSeconds)
		return writeError(c, 429, "RATE_LIMITED", "Too many expensive portfolio reads")
	case errors.Is(err, errExpensiveReadGlobalExhausted):
		c.Set("Retry-After", expensiveReadRateRetryAfterSeconds)
		return writeError(c, 503, "READ_ADMISSION_EXHAUSTED", "Expensive portfolio read admission is temporarily exhausted")
	case errors.Is(err, errExpensiveReadCapacityExhausted), errors.Is(err, errExpensiveReadUnavailable):
		c.Set("Retry-After", expensiveReadCapacityRetryAfterSeconds)
		return writeError(c, 503, "READ_CAPACITY_EXHAUSTED", "Expensive portfolio read capacity is temporarily exhausted")
	default:
		return writeMappedError(c, err)
	}
}
