package httpapi

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func r3ImportAdmission(executionPerSubject, executionGlobal, freshPerSubject, freshGlobal int) *importAdmission {
	return newImportAdmissionWithPolicies(
		importRatePolicy{perSubject: executionPerSubject, globalLimit: executionGlobal, maxSubjects: 2048, window: time.Minute},
		importRatePolicy{perSubject: freshPerSubject, globalLimit: freshGlobal, maxSubjects: 2048, window: time.Minute},
		2,
	)
}

func assertR3ImportResponse(t *testing.T, response *http.Response, wantStatus int, wantCode string) {
	t.Helper()
	defer response.Body.Close()
	if response.StatusCode != wantStatus {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("status = %d, want %d: %s", response.StatusCode, wantStatus, body)
	}
	if wantCode == "IMPORT_ADMISSION_EXHAUSTED" && response.Header.Get("Retry-After") != importExecutionRateRetryAfterSeconds {
		t.Fatalf("execution rejection Retry-After = %q, want %q", response.Header.Get("Retry-After"), importExecutionRateRetryAfterSeconds)
	}
	if wantCode != "" {
		assertImportAdmissionErrorCode(t, response, wantCode)
	}
}

func TestOIIF001R3ExecutionBudgetBoundsAndReclaimsState(t *testing.T) {
	now := time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)
	admission := newDefaultImportAdmission()

	for attempt := 0; attempt < defaultImportExecutionPerSubjectLimit; attempt++ {
		if !admission.allowExecution("same-subject", now) {
			t.Fatalf("execution budget rejected same-subject attempt %d before its limit", attempt+1)
		}
	}
	if admission.allowExecution("same-subject", now) {
		t.Fatal("execution budget admitted a thirteenth same-subject attempt")
	}
	if !admission.allowExecution("different-subject", now) {
		t.Fatal("same-subject execution exhaustion starved a different subject")
	}
	if len(admission.freshRate.globalAttempts) != 0 {
		t.Fatal("execution admission charged the independent fresh-command budget")
	}

	global := r3ImportAdmission(12, 120, 6, 60)
	for attempt := 0; attempt < defaultImportExecutionGlobalLimit; attempt++ {
		if !global.allowExecution(fmt.Sprintf("global-subject-%03d", attempt), now) {
			t.Fatalf("execution global budget rejected attempt %d before its limit", attempt+1)
		}
	}
	if global.allowExecution("global-subject-120", now) {
		t.Fatal("execution global budget admitted a 121st attempt")
	}

	bounded := r3ImportAdmission(12, 4096, 6, 4096)
	for subject := 0; subject < defaultImportExecutionMaxSubjects; subject++ {
		if !bounded.allowExecution(fmt.Sprintf("subject-%04d", subject), now) {
			t.Fatalf("execution subject map rejected subject %d", subject)
		}
	}
	if bounded.allowExecution("subject-overflow", now) {
		t.Fatal("execution subject map admitted a 2049th subject")
	}
	if !bounded.allowExecution("subject-after-expiry", now.Add(time.Minute)) {
		t.Fatal("expired execution subject buckets were not reclaimed")
	}
	if got := len(bounded.executionRate.attempts); got != 1 {
		t.Fatalf("expired execution subject buckets retained: got %d, want 1", got)
	}
}

func TestOIIF001R3ConcurrentExecutionAndFreshAccounting(t *testing.T) {
	admission := newDefaultImportAdmission()
	now := time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)

	for _, testCase := range []struct {
		name  string
		allow func(string, time.Time) bool
		want  int
	}{
		{name: "execution", allow: admission.allowExecution, want: defaultImportExecutionPerSubjectLimit},
		{name: "fresh", allow: admission.allowFresh, want: defaultImportFreshPerSubjectLimit},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var group sync.WaitGroup
			results := make(chan bool, 64)
			for attempt := 0; attempt < cap(results); attempt++ {
				group.Add(1)
				go func() {
					defer group.Done()
					results <- testCase.allow("concurrent-subject-"+testCase.name, now)
				}()
			}
			group.Wait()
			close(results)

			allowed := 0
			for result := range results {
				if result {
					allowed++
				}
			}
			if allowed != testCase.want {
				t.Fatalf("concurrent %s admissions = %d, want %d", testCase.name, allowed, testCase.want)
			}
		})
	}
}

func TestOIIF001R3ExecutionAdmissionBoundsMalformedFloodsBeforeExpensiveWork(t *testing.T) {
	t.Run("invalid JSON append", func(t *testing.T) {
		store, api, app := newImportAdmissionLookupApp(t, nil, verticalslice.CommandReplayArtifact{}, false)
		api.importAdmission = r3ImportAdmission(12, 120, 100, 100)

		for attempt := 0; attempt < 12; attempt++ {
			response, err := app.Test(importAppendRequest([]byte(`{"not valid":`), fmt.Sprintf("r3-invalid-json-%02d", attempt)))
			if err != nil {
				t.Fatalf("invalid JSON attempt %d: %v", attempt+1, err)
			}
			assertR3ImportResponse(t, response, http.StatusBadRequest, "VALIDATION_ERROR")
		}
		lookupCalls := store.lookupCalls
		appendCalls := store.appendImportedCalls
		response, err := app.Test(importAppendRequest([]byte(`{"not valid":`), "r3-invalid-json-over-limit"))
		if err != nil {
			t.Fatalf("over-limit invalid JSON: %v", err)
		}
		if response.Header.Get("Retry-After") != importExecutionRateRetryAfterSeconds {
			response.Body.Close()
			t.Fatalf("execution rejection Retry-After = %q", response.Header.Get("Retry-After"))
		}
		assertR3ImportResponse(t, response, http.StatusServiceUnavailable, "IMPORT_ADMISSION_EXHAUSTED")
		if store.lookupCalls != lookupCalls || store.appendImportedCalls != appendCalls {
			t.Fatalf("execution-denied malformed append reached expensive work: lookups=%d/%d appends=%d/%d", store.lookupCalls, lookupCalls, store.appendImportedCalls, appendCalls)
		}
		if got := len(api.importAdmission.freshRate.globalAttempts); got != 0 {
			t.Fatalf("malformed append flood charged fresh budget: %d", got)
		}
		assertImportCapacityFullyReleased(t, api)
	})

	t.Run("invalid CSV review", func(t *testing.T) {
		_, api, app := newImportAdmissionLookupApp(t, nil, verticalslice.CommandReplayArtifact{}, false)
		api.importAdmission = r3ImportAdmission(12, 120, 100, 100)

		for attempt := 0; attempt < 12; attempt++ {
			response, err := app.Test(importReviewJSONRequest("not,a,valid,import"))
			if err != nil {
				t.Fatalf("invalid CSV attempt %d: %v", attempt+1, err)
			}
			assertR3ImportResponse(t, response, http.StatusBadRequest, "VALIDATION_ERROR")
		}
		response, err := app.Test(importReviewJSONRequest("not,a,valid,import"))
		if err != nil {
			t.Fatalf("over-limit invalid CSV: %v", err)
		}
		assertR3ImportResponse(t, response, http.StatusServiceUnavailable, "IMPORT_ADMISSION_EXHAUSTED")
		if got := len(api.importAdmission.freshRate.globalAttempts); got != 12 {
			t.Fatalf("invalid review flood fresh charges = %d, want 12", got)
		}
		assertImportCapacityFullyReleased(t, api)
	})

	t.Run("invalid signed token", func(t *testing.T) {
		store, api, app := newImportAdmissionLookupApp(t, nil, verticalslice.CommandReplayArtifact{}, false)
		body := validImportAppendBody(t, app, importCSV)
		request := decodeJSONObject(t, body)
		request["reviewToken"] = "not-a-signed-review-token"
		invalidTokenBody := encodeJSONObject(t, request)
		api.importAdmission = r3ImportAdmission(12, 120, 100, 100)

		for attempt := 0; attempt < 12; attempt++ {
			response := stage336Append(t, app, invalidTokenBody, fmt.Sprintf("r3-invalid-token-%02d", attempt))
			assertR3ImportResponse(t, response, http.StatusBadRequest, "VALIDATION_ERROR")
		}
		lookupCalls := store.lookupCalls
		response := stage336Append(t, app, invalidTokenBody, "r3-invalid-token-over-limit")
		assertR3ImportResponse(t, response, http.StatusServiceUnavailable, "IMPORT_ADMISSION_EXHAUSTED")
		if store.lookupCalls != lookupCalls || store.appendImportedCalls != 0 {
			t.Fatalf("execution-denied invalid-token request reached replay or append work: lookups=%d/%d appends=%d", store.lookupCalls, lookupCalls, store.appendImportedCalls)
		}
		if got := len(api.importAdmission.freshRate.globalAttempts); got != 0 {
			t.Fatalf("invalid token flood charged fresh budget: %d", got)
		}
		assertImportCapacityFullyReleased(t, api)
	})
}

func TestOIIF001R3SourceFileHashMismatchPrecedesEveryReplayPath(t *testing.T) {
	store, _, api, app, _ := newStage336ReplayApp(t)
	body := validImportAppendBody(t, app, importCSV)
	first := stage336Append(t, app, body, "r3-source-hash-completed-key")
	assertR3ImportResponse(t, first, http.StatusCreated, "")
	if store.appendReplayCalls != 1 || store.lookupCalls != 1 {
		t.Fatalf("completed replay setup = appends=%d lookups=%d, want 1/1", store.appendReplayCalls, store.lookupCalls)
	}

	request := decodeJSONObject(t, body)
	request["sourceFileHash"] = strings.Repeat("0", 64)
	tampered := encodeJSONObject(t, request)
	api.importAdmission = r3ImportAdmission(100, 100, 100, 100)
	lookupCalls := store.lookupCalls

	for _, key := range []string{"r3-source-hash-completed-key", "r3-source-hash-fresh-key"} {
		response := stage336Append(t, app, tampered, key)
		assertR3ImportResponse(t, response, http.StatusBadRequest, "VALIDATION_ERROR")
	}
	if store.lookupCalls != lookupCalls || store.appendReplayCalls != 1 {
		t.Fatalf("tampered source hash reached current or historic replay work: lookups=%d/%d appends=%d", store.lookupCalls, lookupCalls, store.appendReplayCalls)
	}
	if got := len(api.importAdmission.freshRate.globalAttempts); got != 0 {
		t.Fatalf("tampered source hash reached fresh-command admission: %d", got)
	}
	assertImportCapacityFullyReleased(t, api)
}

func TestOIIF001R3CompletedReplaysRemainExecutionBoundedButBypassFreshBudget(t *testing.T) {
	t.Run("current parser replay", func(t *testing.T) {
		store, _, api, app, advance := newStage336ReplayApp(t)
		body := validImportAppendBody(t, app, importCSV)
		first := stage336Append(t, app, body, "r3-current-replay-key")
		assertR3ImportResponse(t, first, http.StatusCreated, "")

		api.importAdmission = r3ImportAdmission(12, 120, 0, 0)
		for attempt := 0; attempt < 12; attempt++ {
			response := stage336Append(t, app, body, "r3-current-replay-key")
			assertR3ImportResponse(t, response, http.StatusCreated, "")
		}
		blocked := stage336Append(t, app, body, "r3-current-replay-key")
		assertR3ImportResponse(t, blocked, http.StatusServiceUnavailable, "IMPORT_ADMISSION_EXHAUSTED")
		if store.appendReplayCalls != 1 || len(api.importAdmission.freshRate.globalAttempts) != 0 {
			t.Fatalf("current replay changed financial state or consumed fresh budget: appends=%d fresh=%d", store.appendReplayCalls, len(api.importAdmission.freshRate.globalAttempts))
		}
		advance(time.Minute)
		recovered := stage336Append(t, app, body, "r3-current-replay-key")
		assertR3ImportResponse(t, recovered, http.StatusCreated, "")
		assertImportCapacityFullyReleased(t, api)
	})

	for _, testCase := range []struct {
		name          string
		parserVersion int
		payload       string
	}{
		{name: "historic v1", parserVersion: 1, payload: csvHeaderForHTTP + "BUY,SBER,2.00000000,001.25,2.50000000,0.00000000,0.00000000,2026-01-10,2026-01-13,RUB,r3-v1,historic\\n"},
		{name: "historic v2", parserVersion: 2, payload: stage340ExcessiveHeaderCSV()},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			store, service, api, app, advance := newStage336ReplayApp(t)
			body, request := stage336HistoricalAppendRequestForParserVersion(t, api, testCase.payload, testCase.parserVersion)
			seedStage336CompletedImport(t, service, request)
			advance(20 * time.Minute)
			api.importAdmission = r3ImportAdmission(12, 120, 0, 0)

			for attempt := 0; attempt < 12; attempt++ {
				response := stage336Append(t, app, body, stage336ImportKey)
				assertR3ImportResponse(t, response, http.StatusCreated, "")
			}
			blocked := stage336Append(t, app, body, stage336ImportKey)
			assertR3ImportResponse(t, blocked, http.StatusServiceUnavailable, "IMPORT_ADMISSION_EXHAUSTED")
			if store.appendReplayCalls != 1 || len(api.importAdmission.freshRate.globalAttempts) != 0 {
				t.Fatalf("%s changed financial state or consumed fresh budget: appends=%d fresh=%d", testCase.name, store.appendReplayCalls, len(api.importAdmission.freshRate.globalAttempts))
			}
			advance(time.Minute)
			recovered := stage336Append(t, app, body, stage336ImportKey)
			assertR3ImportResponse(t, recovered, http.StatusCreated, "")
			assertImportCapacityFullyReleased(t, api)
		})
	}
}

func TestOIIF001R3ConflictAndInflightRemainExecutionBounded(t *testing.T) {
	for _, lookupErr := range []error{postgres.ErrIdempotencyConflict, postgres.ErrIdempotencyInFlight} {
		t.Run(lookupErr.Error(), func(t *testing.T) {
			store, api, app := newImportAdmissionLookupApp(t, lookupErr, verticalslice.CommandReplayArtifact{}, false)
			body := validImportAppendBody(t, app, importCSV)
			api.importAdmission = r3ImportAdmission(12, 120, 0, 0)

			for attempt := 0; attempt < 12; attempt++ {
				response := stage336Append(t, app, body, "r3-lookup-outcome-key")
				assertR3ImportResponse(t, response, http.StatusConflict, "")
			}
			lookups := store.lookupCalls
			blocked := stage336Append(t, app, body, "r3-lookup-outcome-key")
			assertR3ImportResponse(t, blocked, http.StatusServiceUnavailable, "IMPORT_ADMISSION_EXHAUSTED")
			if store.lookupCalls != lookups || store.appendImportedCalls != 0 {
				t.Fatalf("execution-denied %v reached replay or append: lookups=%d/%d appends=%d", lookupErr, store.lookupCalls, lookups, store.appendImportedCalls)
			}
			if len(api.importAdmission.freshRate.globalAttempts) != 0 {
				t.Fatalf("%v charged fresh-command admission", lookupErr)
			}
			assertImportCapacityFullyReleased(t, api)
		})
	}
}

func TestOIIF001R3ReviewAndAppendShareExecutionBudget(t *testing.T) {
	_, api, app := newImportAdmissionLookupApp(t, nil, verticalslice.CommandReplayArtifact{}, false)
	api.importAdmission = r3ImportAdmission(12, 120, 100, 100)

	for attempt := 0; attempt < 6; attempt++ {
		response, err := app.Test(importReviewRequest(`{"not valid":`))
		if err != nil {
			t.Fatalf("invalid review %d: %v", attempt+1, err)
		}
		assertR3ImportResponse(t, response, http.StatusBadRequest, "VALIDATION_ERROR")
	}
	for attempt := 0; attempt < 6; attempt++ {
		response, err := app.Test(importAppendRequest([]byte(`{"not valid":`), fmt.Sprintf("r3-shared-execution-%02d", attempt)))
		if err != nil {
			t.Fatalf("invalid append %d: %v", attempt+1, err)
		}
		assertR3ImportResponse(t, response, http.StatusBadRequest, "VALIDATION_ERROR")
	}
	response, err := app.Test(importReviewRequest(`{"not valid":`))
	if err != nil {
		t.Fatalf("over-limit mixed route request: %v", err)
	}
	assertR3ImportResponse(t, response, http.StatusServiceUnavailable, "IMPORT_ADMISSION_EXHAUSTED")
	if got := len(api.importAdmission.freshRate.globalAttempts); got != 6 {
		t.Fatalf("only review route should charge fresh admission: got %d, want 6", got)
	}
	assertImportCapacityFullyReleased(t, api)
}
