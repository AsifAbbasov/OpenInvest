package httpapi

import (
	"errors"
	"strings"
	"sync"
)

var (
	errExpensiveReadSubjectLimited    = errors.New("expensive read subject capacity exhausted")
	errExpensiveReadCapacityExhausted = errors.New("expensive read global capacity exhausted")
)

// expensiveReadAdmission bounds only the full-history read projections that can materialize an
// entire effective ledger. It is intentionally simple: no queue, no timer, and no long-lived
// attacker-controlled key state. Subject entries exist only while a request owns global capacity.
type expensiveReadAdmission struct {
	mu               sync.Mutex
	perSubjectActive map[string]int
	perSubjectLimit  int
	capacity         chan struct{}
}

func newExpensiveReadAdmission(perSubjectLimit int, globalCapacity int) *expensiveReadAdmission {
	return &expensiveReadAdmission{
		perSubjectActive: map[string]int{},
		perSubjectLimit:  nonNegativeInt(perSubjectLimit),
		capacity:         make(chan struct{}, nonNegativeInt(globalCapacity)),
	}
}

func newDefaultExpensiveReadAdmission() *expensiveReadAdmission {
	return newExpensiveReadAdmission(defaultExpensiveReadPerSubjectCapacity, defaultExpensiveReadGlobalCapacity)
}

func (admission *expensiveReadAdmission) acquire(subjectID string) (func(), error) {
	if admission == nil || admission.perSubjectLimit == 0 || cap(admission.capacity) == 0 {
		return nil, errExpensiveReadCapacityExhausted
	}
	subjectID = strings.TrimSpace(subjectID)
	if subjectID == "" {
		return nil, errExpensiveReadSubjectLimited
	}

	select {
	case admission.capacity <- struct{}{}:
	default:
		return nil, errExpensiveReadCapacityExhausted
	}

	admission.mu.Lock()
	if admission.perSubjectActive[subjectID] >= admission.perSubjectLimit {
		admission.mu.Unlock()
		<-admission.capacity
		return nil, errExpensiveReadSubjectLimited
	}
	admission.perSubjectActive[subjectID]++
	admission.mu.Unlock()

	var once sync.Once
	return func() {
		once.Do(func() {
			admission.mu.Lock()
			remaining := admission.perSubjectActive[subjectID] - 1
			if remaining <= 0 {
				delete(admission.perSubjectActive, subjectID)
			} else {
				admission.perSubjectActive[subjectID] = remaining
			}
			admission.mu.Unlock()
			<-admission.capacity
		})
	}, nil
}

func (api *API) acquireExpensiveRead(subjectID string) (func(), error) {
	if api == nil || api.expensiveReads == nil {
		return func() {}, nil
	}
	return api.expensiveReads.acquire(subjectID)
}
