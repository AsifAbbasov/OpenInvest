package httpapi

import (
	"errors"
	"testing"
	"time"
)

func TestExpensiveReadAdmissionSeparatesSubjectGlobalAndCapacityBounds(t *testing.T) {
	now := time.Date(2026, 10, 6, 8, 0, 0, 0, time.UTC)

	subjectAdmission := &expensiveReadAdmission{
		rate: newImportRate(importRatePolicy{perSubject: 2, globalLimit: 10, maxSubjects: 10, window: time.Minute}),
		capacity: make(chan struct{}, 2),
	}
	for i := 0; i < 2; i++ {
		release, err := subjectAdmission.acquire("subject-a", now)
		if err != nil {
			t.Fatalf("subject admission %d: %v", i, err)
		}
		release()
	}
	if _, err := subjectAdmission.acquire("subject-a", now); !errors.Is(err, errExpensiveReadRateLimited) {
		t.Fatalf("subject exhaustion error=%v", err)
	}

	globalAdmission := &expensiveReadAdmission{
		rate: newImportRate(importRatePolicy{perSubject: 10, globalLimit: 2, maxSubjects: 10, window: time.Minute}),
		capacity: make(chan struct{}, 2),
	}
	for _, subjectID := range []string{"subject-a", "subject-b"} {
		release, err := globalAdmission.acquire(subjectID, now)
		if err != nil {
			t.Fatalf("global admission for %s: %v", subjectID, err)
		}
		release()
	}
	if _, err := globalAdmission.acquire("subject-c", now); !errors.Is(err, errExpensiveReadGlobalExhausted) {
		t.Fatalf("global exhaustion error=%v", err)
	}

	capacityAdmission := &expensiveReadAdmission{
		rate: newImportRate(importRatePolicy{perSubject: 10, globalLimit: 10, maxSubjects: 10, window: time.Minute}),
		capacity: make(chan struct{}, 1),
	}
	release, err := capacityAdmission.acquire("subject-a", now)
	if err != nil {
		t.Fatalf("capacity first admission: %v", err)
	}
	defer release()
	if _, err := capacityAdmission.acquire("subject-b", now); !errors.Is(err, errExpensiveReadCapacityExhausted) {
		t.Fatalf("capacity exhaustion error=%v", err)
	}
}

func TestExpensiveReadAdmissionDoesNotChargeCapacityRejection(t *testing.T) {
	now := time.Date(2026, 10, 6, 8, 0, 0, 0, time.UTC)
	admission := &expensiveReadAdmission{
		rate: newImportRate(importRatePolicy{perSubject: 1, globalLimit: 10, maxSubjects: 10, window: time.Minute}),
		capacity: make(chan struct{}, 1),
	}
	releaseA, err := admission.acquire("subject-a", now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := admission.acquire("subject-b", now); !errors.Is(err, errExpensiveReadCapacityExhausted) {
		t.Fatalf("capacity rejection=%v", err)
	}
	releaseA()

	releaseB, err := admission.acquire("subject-b", now)
	if err != nil {
		t.Fatalf("capacity rejection consumed subject budget: %v", err)
	}
	releaseB()
	if _, err := admission.acquire("subject-b", now); !errors.Is(err, errExpensiveReadRateLimited) {
		t.Fatalf("subject budget after real admission=%v", err)
	}
}
