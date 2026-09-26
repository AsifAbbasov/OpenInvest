package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestImportAdmissionBoundsFreshWorkAndReclaimsExpiredSubjects(t *testing.T) {
	admission := newDefaultImportAdmission()
	now := time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)

	if admission.executionRate.perSubject != 12 || admission.executionRate.globalLimit != 120 || admission.executionRate.maxSubjects != 2048 ||
		admission.freshRate.perSubject != 6 || admission.freshRate.globalLimit != 60 || admission.freshRate.maxSubjects != 2048 || cap(admission.capacity) != 2 {
		t.Fatalf("default import admission policy = execution=%+v fresh=%+v capacity=%d, want 12/120/2048 and 6/60/2048/2", admission.executionRate, admission.freshRate, cap(admission.capacity))
	}
	for index := 0; index < defaultImportFreshPerSubjectLimit; index++ {
		if !admission.allowFresh("subject-a", now) {
			t.Fatalf("subject quota rejected attempt %d before its limit", index+1)
		}
	}
	if admission.allowFresh("subject-a", now) {
		t.Fatal("per-subject quota admitted a seventh fresh command")
	}
	if !admission.allowFresh("subject-b", now) {
		t.Fatal("subject A exhaustion incorrectly exhausted subject B")
	}
	if !admission.allowFresh("subject-a", now.Add(time.Minute)) {
		t.Fatal("expired per-subject window did not restore admission")
	}

	global := newImportAdmission(6, 60, 2048, time.Minute, 2)
	for index := 0; index < defaultImportFreshGlobalLimit; index++ {
		if !global.allowFresh(fmt.Sprintf("global-subject-%03d", index), now) {
			t.Fatalf("global quota rejected attempt %d before its limit", index+1)
		}
	}
	if global.allowFresh("global-subject-061", now) {
		t.Fatal("global quota admitted a sixty-first fresh command")
	}

	limitedSubjects := newImportAdmission(6, 4096, 2048, time.Minute, 2)
	for index := 0; index < 2048; index++ {
		if !limitedSubjects.allowFresh(fmt.Sprintf("subject-%04d", index), now) {
			t.Fatalf("subject map rejected tracked subject %d", index)
		}
	}
	if limitedSubjects.allowFresh("subject-overflow", now) {
		t.Fatal("bounded subject map admitted a 2049th subject")
	}
	if !limitedSubjects.allowFresh("subject-after-expiry", now.Add(time.Minute)) {
		t.Fatal("expired subject buckets were not reclaimed")
	}
	if got := len(limitedSubjects.freshRate.attempts); got != 1 {
		t.Fatalf("expired subject buckets were retained: got %d want 1", got)
	}

	release, err := admission.acquire()
	if err != nil {
		t.Fatalf("acquire first capacity slot: %v", err)
	}
	releaseSecond, err := admission.acquire()
	if err != nil {
		t.Fatalf("acquire second capacity slot: %v", err)
	}
	if _, err := admission.acquire(); !errors.Is(err, errImportCapacityExhausted) {
		t.Fatalf("third capacity acquire error = %v, want immediate exhaustion without queuing", err)
	}
	release()
	releaseSecond()
	release()
	if release, err := admission.acquire(); err != nil {
		t.Fatalf("capacity did not recover after release: %v", err)
	} else {
		release()
	}
}

func TestImportAdmissionConcurrentFreshRateAccounting(t *testing.T) {
	admission := newImportAdmission(6, 60, 2048, time.Minute, 2)
	now := time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)
	var group sync.WaitGroup
	results := make(chan bool, 64)
	for index := 0; index < cap(results); index++ {
		group.Add(1)
		go func() {
			defer group.Done()
			results <- admission.allowFresh("concurrent-subject", now)
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
	if allowed != 6 {
		t.Fatalf("concurrent per-subject admissions = %d, want 6", allowed)
	}
}

func TestImportReviewCapacityDenialPrecedesJSONAndDoesNotChargeFreshRate(t *testing.T) {
	_, _, api, app, _ := newStage336ReplayApp(t)
	api.importAdmission = newImportAdmission(6, 60, 2048, time.Minute, 2)
	releaseFirst, err := api.acquireImportCapacity()
	if err != nil {
		t.Fatalf("acquire first capacity slot: %v", err)
	}
	releaseSecond, err := api.acquireImportCapacity()
	if err != nil {
		t.Fatalf("acquire second capacity slot: %v", err)
	}
	defer releaseFirst()
	defer releaseSecond()

	for _, request := range []*http.Request{
		importReviewRequest(`{"not valid":`),
		importAppendRequest([]byte(`{"not valid":`), "oi-if-001-capacity-before-json-key"),
	} {
		response, err := app.Test(request)
		if err != nil {
			t.Fatalf("request under capacity exhaustion: %v", err)
		}
		if response.StatusCode != http.StatusServiceUnavailable {
			response.Body.Close()
			t.Fatalf("capacity denial status = %d, want %d", response.StatusCode, http.StatusServiceUnavailable)
		}
		if response.Header.Get("Retry-After") != importCapacityRetryAfterSeconds {
			response.Body.Close()
			t.Fatalf("capacity retry-after = %q, want %q", response.Header.Get("Retry-After"), importCapacityRetryAfterSeconds)
		}
		assertImportAdmissionErrorCode(t, response, "IMPORT_CAPACITY_EXHAUSTED")
		response.Body.Close()
	}
	if got := len(api.importAdmission.executionRate.globalAttempts); got != 0 {
		t.Fatalf("capacity-denied malformed request charged execution budget: %d attempts", got)
	}
	if got := len(api.importAdmission.freshRate.globalAttempts); got != 0 {
		t.Fatalf("capacity-denied malformed request charged fresh budget: %d attempts", got)
	}
}

func TestImportReviewAndAppendShareFreshRateBudgets(t *testing.T) {
	store, api, app := newImportAdmissionLookupApp(t, nil, verticalslice.CommandReplayArtifact{}, false)
	api.importAdmission = newImportAdmission(6, 60, 2048, time.Minute, 2)
	body := validImportAppendBody(t, app, importCSV)
	for index := 0; index < 5; index++ {
		response := stage336Append(t, app, body, fmt.Sprintf("oi-if-001-shared-subject-key-%02d", index))
		if response.StatusCode != http.StatusCreated {
			response.Body.Close()
			t.Fatalf("fresh append %d status = %d, want %d", index+1, response.StatusCode, http.StatusCreated)
		}
		response.Body.Close()
	}
	denied := stage336Append(t, app, body, "oi-if-001-shared-subject-key-06")
	defer denied.Body.Close()
	if denied.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("shared review/append subject budget status = %d, want %d", denied.StatusCode, http.StatusTooManyRequests)
	}
	if denied.Header.Get("Retry-After") != importFreshRateRetryAfterSeconds {
		t.Fatalf("shared subject rate retry-after = %q, want %q", denied.Header.Get("Retry-After"), importFreshRateRetryAfterSeconds)
	}
	assertImportAdmissionErrorCode(t, denied, "RATE_LIMITED")
	if store.appendImportedCalls != 5 {
		t.Fatalf("rate-denied append entered writable path: append calls=%d, want 5", store.appendImportedCalls)
	}

	api.importAdmission = newImportAdmission(6, 60, 2048, time.Minute, 2)
	for index := 0; index < 60; index++ {
		if err := api.admitFreshImport(fmt.Sprintf("global-budget-%03d", index)); err != nil {
			t.Fatalf("global budget setup admission %d: %v", index+1, err)
		}
	}
	reviewBody := []byte(`{"sourceAccountLabel":"Manual CSV","csvPayload":"ignored"}`)
	review := httptest.NewRequest(http.MethodPost, "/api/v1/portfolios/00000000-0000-4000-8000-000000000002/imports/review", bytes.NewReader(reviewBody))
	review.Header.Set("Content-Type", "application/json")
	reviewResponse, err := app.Test(review)
	if err != nil {
		t.Fatalf("global budget review request: %v", err)
	}
	defer reviewResponse.Body.Close()
	if reviewResponse.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("shared global review budget status = %d, want %d", reviewResponse.StatusCode, http.StatusTooManyRequests)
	}
	assertImportAdmissionErrorCode(t, reviewResponse, "RATE_LIMITED")
}

func TestImportCapacityReleasesAfterEverySynchronousHandlerExit(t *testing.T) {
	for _, testCase := range []struct {
		name      string
		newApp    func(*testing.T) (*API, *fiber.App)
		request   func(*testing.T, *fiber.App) *http.Request
		wantCode  int
		wantError string
	}{
		{
			name: "successful review",
			newApp: func(t *testing.T) (*API, *fiber.App) {
				_, api, app := newImportAdmissionLookupApp(t, nil, verticalslice.CommandReplayArtifact{}, false)
				return api, app
			},
			request: func(_ *testing.T, _ *fiber.App) *http.Request {
				return importReviewJSONRequest(importCSV)
			},
			wantCode: http.StatusOK,
		},
		{
			name: "successful append",
			newApp: func(t *testing.T) (*API, *fiber.App) {
				_, api, app := newImportAdmissionLookupApp(t, nil, verticalslice.CommandReplayArtifact{}, false)
				return api, app
			},
			request: func(t *testing.T, app *fiber.App) *http.Request {
				return importAppendRequest(validImportAppendBody(t, app, importCSV), "oi-if-001-release-success-key")
			},
			wantCode: http.StatusCreated,
		},
		{
			name: "strict validation error",
			newApp: func(t *testing.T) (*API, *fiber.App) {
				_, api, app := newImportAdmissionLookupApp(t, nil, verticalslice.CommandReplayArtifact{}, false)
				return api, app
			},
			request: func(_ *testing.T, _ *fiber.App) *http.Request {
				return importReviewRequest("{not-json")
			},
			wantCode:  http.StatusBadRequest,
			wantError: "VALIDATION_ERROR",
		},
		{
			name: "history lookup error",
			newApp: func(t *testing.T) (*API, *fiber.App) {
				store := &importAdmissionHistoryErrorStore{}
				return newImportAdmissionAppForStore(t, store)
			},
			request: func(_ *testing.T, _ *fiber.App) *http.Request {
				return importReviewJSONRequest(importCSV)
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "INTERNAL_ERROR",
		},
		{
			name: "replay lookup error",
			newApp: func(t *testing.T) (*API, *fiber.App) {
				_, api, app := newImportAdmissionLookupApp(t, postgres.ErrIdempotencyInFlight, verticalslice.CommandReplayArtifact{}, false)
				return api, app
			},
			request: func(t *testing.T, app *fiber.App) *http.Request {
				return importAppendRequest(validImportAppendBody(t, app, importCSV), "oi-if-001-release-replay-key")
			},
			wantCode:  http.StatusConflict,
			wantError: "IDEMPOTENCY_IN_FLIGHT",
		},
		{
			name: "mutation error",
			newApp: func(t *testing.T) (*API, *fiber.App) {
				store := &importAPITestStore{appendImportedError: errors.New("store unavailable")}
				return newImportAdmissionAppForStore(t, store)
			},
			request: func(t *testing.T, app *fiber.App) *http.Request {
				return importAppendRequest(validImportAppendBody(t, app, importCSV), "oi-if-001-release-mutation-key")
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "INTERNAL_ERROR",
		},
		{
			name: "canceled downstream operation",
			newApp: func(t *testing.T) (*API, *fiber.App) {
				store := &importAPITestStore{appendImportedError: context.Canceled}
				return newImportAdmissionAppForStore(t, store)
			},
			request: func(t *testing.T, app *fiber.App) *http.Request {
				return importAppendRequest(validImportAppendBody(t, app, importCSV), "oi-if-001-release-canceled-key")
			},
			wantCode:  http.StatusInternalServerError,
			wantError: "INTERNAL_ERROR",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			api, app := testCase.newApp(t)
			response, err := app.Test(testCase.request(t, app))
			if err != nil {
				t.Fatalf("request: %v", err)
			}
			defer response.Body.Close()
			if response.StatusCode != testCase.wantCode {
				t.Fatalf("status = %d, want %d", response.StatusCode, testCase.wantCode)
			}
			if testCase.wantError != "" {
				assertImportAdmissionErrorCode(t, response, testCase.wantError)
			}
			assertImportCapacityFullyReleased(t, api)
		})
	}
}

func TestImportAdmissionFailsClosedAndDoesNotLeakAfterRepeatedFailures(t *testing.T) {
	_, api, app := newImportAdmissionLookupApp(t, nil, verticalslice.CommandReplayArtifact{}, false)
	api.importAdmission = nil
	for _, request := range []*http.Request{
		importReviewRequest("{not-json"),
		importAppendRequest([]byte("{not-json"), "oi-if-001-nil-admission-key"),
	} {
		response, err := app.Test(request)
		if err != nil {
			t.Fatalf("nil-admission request: %v", err)
		}
		if response.StatusCode != http.StatusServiceUnavailable || response.Header.Get("Retry-After") != importCapacityRetryAfterSeconds {
			response.Body.Close()
			t.Fatalf("nil-admission status=%d retry-after=%q", response.StatusCode, response.Header.Get("Retry-After"))
		}
		assertImportAdmissionErrorCode(t, response, "IMPORT_CAPACITY_EXHAUSTED")
		response.Body.Close()
	}

	api.importAdmission = newImportAdmissionWithPolicies(
		importRatePolicy{perSubject: 100, globalLimit: 100, maxSubjects: 2048, window: time.Minute},
		importRatePolicy{perSubject: 100, globalLimit: 100, maxSubjects: 2048, window: time.Minute},
		2,
	)
	for index := 0; index < 20; index++ {
		response, err := app.Test(importReviewRequest("{not-json"))
		if err != nil {
			t.Fatalf("repeated invalid review %d: %v", index+1, err)
		}
		if response.StatusCode != http.StatusBadRequest {
			response.Body.Close()
			t.Fatalf("repeated invalid review %d status=%d", index+1, response.StatusCode)
		}
		response.Body.Close()
	}
	assertImportCapacityFullyReleased(t, api)
}

func TestImportReplayAndConflictBypassFreshRateBudget(t *testing.T) {
	store, _, api, app, _ := newStage336ReplayApp(t)
	body := validImportAppendBody(t, app, importCSV)
	first := stage336Append(t, app, body, "oi-if-001-current-replay-key-01")
	firstBody, err := io.ReadAll(first.Body)
	if err != nil {
		t.Fatalf("read first append: %v", err)
	}
	first.Body.Close()
	if first.StatusCode != http.StatusCreated || store.appendReplayCalls != 1 {
		t.Fatalf("fresh append setup failed: status=%d appends=%d", first.StatusCode, store.appendReplayCalls)
	}

	api.importAdmission = newImportAdmission(0, 0, 0, time.Minute, 2)
	replay := stage336Append(t, app, body, "oi-if-001-current-replay-key-01")
	replayBody, err := io.ReadAll(replay.Body)
	if err != nil {
		t.Fatalf("read completed replay: %v", err)
	}
	replay.Body.Close()
	if replay.StatusCode != http.StatusCreated || !bytes.Equal(replayBody, firstBody) {
		t.Fatalf("completed replay did not bypass fresh admission: status=%d body=%s", replay.StatusCode, replayBody)
	}
	if store.appendReplayCalls != 1 {
		t.Fatalf("completed replay executed another financial append: %d", store.appendReplayCalls)
	}

	for _, lookupErr := range []error{postgres.ErrIdempotencyConflict, postgres.ErrIdempotencyInFlight} {
		t.Run(lookupErr.Error(), func(t *testing.T) {
			fixture, api, app := newImportAdmissionLookupApp(t, lookupErr, verticalslice.CommandReplayArtifact{}, false)
			body := validImportAppendBody(t, app, importCSV)
			api.importAdmission = newImportAdmission(0, 0, 0, time.Minute, 2)
			response := stage336Append(t, app, body, "oi-if-001-conflict-key-01")
			defer response.Body.Close()
			if response.StatusCode != http.StatusConflict {
				t.Fatalf("lookup outcome %v status = %d, want %d", lookupErr, response.StatusCode, http.StatusConflict)
			}
			if fixture.appendImportedCalls != 0 {
				t.Fatalf("lookup outcome %v executed a fresh append", lookupErr)
			}
			if fixture.lookupCalls != 1 {
				t.Fatalf("lookup outcome %v made %d lookups, want 1 before admission", lookupErr, fixture.lookupCalls)
			}
		})
	}
}

func TestImportRateDenialRechecksReplayOnce(t *testing.T) {
	artifact := verticalslice.CommandReplayArtifact{
		StatusCode: http.StatusCreated,
		Body:       []byte(`{"data":{"replayed":true}}`),
		RequestID:  "11111111-1111-4111-8111-111111111111",
		TraceID:    "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	store, api, app := newImportAdmissionLookupApp(t, nil, artifact, true)
	body := validImportAppendBody(t, app, importCSV)
	api.importAdmission = newImportAdmission(0, 0, 0, time.Minute, 2)

	response := stage336Append(t, app, body, "oi-if-001-rate-recheck-key-01")
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read race-recheck response: %v", err)
	}
	if response.StatusCode != artifact.StatusCode || !bytes.Equal(responseBody, artifact.Body) {
		t.Fatalf("rate denial did not return concurrently completed artifact: status=%d body=%s", response.StatusCode, responseBody)
	}
	if store.lookupCalls != 2 {
		t.Fatalf("rate denial lookup count = %d, want exactly one initial lookup plus one recheck", store.lookupCalls)
	}
	if store.appendImportedCalls != 0 {
		t.Fatalf("race-recheck path executed a fresh append: %d", store.appendImportedCalls)
	}
}

func TestHistoricCompletedReplayBypassesFreshRateBudget(t *testing.T) {
	for _, testCase := range []struct {
		parserVersion int
		payload       string
	}{
		{parserVersion: 1, payload: csvHeaderForHTTP + "BUY,SBER,2.00000000,001.25,2.50000000,0.00000000,0.00000000,2026-01-10,2026-01-13,RUB,oi-if-001-historic,historic\n"},
		{parserVersion: 2, payload: stage340ExcessiveHeaderCSV()},
	} {
		t.Run(fmt.Sprintf("v%d", testCase.parserVersion), func(t *testing.T) {
			store, service, api, app, advance := newStage336ReplayApp(t)
			body, request := stage336HistoricalAppendRequestForParserVersion(t, api, testCase.payload, testCase.parserVersion)
			artifact := seedStage336CompletedImport(t, service, request)
			advance(20 * time.Minute)
			api.importAdmission = newImportAdmission(0, 0, 0, time.Minute, 2)
			releaseOne, err := api.acquireImportCapacity()
			if err != nil {
				t.Fatalf("hold first historic replay capacity slot: %v", err)
			}
			releaseTwo, err := api.acquireImportCapacity()
			if err != nil {
				releaseOne()
				t.Fatalf("hold second historic replay capacity slot: %v", err)
			}
			blocked := stage336Append(t, app, body, stage336ImportKey)
			if blocked.StatusCode != http.StatusServiceUnavailable {
				blocked.Body.Close()
				releaseOne()
				releaseTwo()
				t.Fatalf("historic v%d replay bypassed shared capacity: status=%d", testCase.parserVersion, blocked.StatusCode)
			}
			assertImportAdmissionErrorCode(t, blocked, "IMPORT_CAPACITY_EXHAUSTED")
			blocked.Body.Close()
			releaseOne()
			releaseTwo()

			response := stage336Append(t, app, body, stage336ImportKey)
			defer response.Body.Close()
			replayedBody, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatalf("read historic replay response: %v", err)
			}
			if response.StatusCode != artifact.StatusCode || !bytes.Equal(replayedBody, artifact.Body) {
				t.Fatalf("historic v%d replay did not bypass fresh admission: status=%d body=%s", testCase.parserVersion, response.StatusCode, replayedBody)
			}
			if store.appendReplayCalls != 1 || store.lookupCalls != 1 {
				t.Fatalf("historic v%d replay executed fresh work: appends=%d lookups=%d", testCase.parserVersion, store.appendReplayCalls, store.lookupCalls)
			}
		})
	}
}

func TestCompletedReplayRemainsInsideSharedCapacity(t *testing.T) {
	store, _, api, app, _ := newStage336ReplayApp(t)
	body := validImportAppendBody(t, app, importCSV)
	first := stage336Append(t, app, body, "oi-if-001-capacity-replay-key-01")
	if first.StatusCode != http.StatusCreated {
		first.Body.Close()
		t.Fatalf("fresh append setup status=%d", first.StatusCode)
	}
	first.Body.Close()
	releaseOne, err := api.acquireImportCapacity()
	if err != nil {
		t.Fatalf("hold first capacity slot: %v", err)
	}
	releaseTwo, err := api.acquireImportCapacity()
	if err != nil {
		releaseOne()
		t.Fatalf("hold second capacity slot: %v", err)
	}
	defer releaseOne()
	defer releaseTwo()

	replay := stage336Append(t, app, body, "oi-if-001-capacity-replay-key-01")
	defer replay.Body.Close()
	if replay.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("completed replay bypassed shared capacity: status=%d", replay.StatusCode)
	}
	assertImportAdmissionErrorCode(t, replay, "IMPORT_CAPACITY_EXHAUSTED")
	if store.appendReplayCalls != 1 {
		t.Fatalf("capacity-denied completed replay executed a financial append: %d", store.appendReplayCalls)
	}
}

type importAdmissionLookupStore struct {
	importAPITestStore
	lookupErr     error
	completed     verticalslice.CommandReplayArtifact
	completeOnTwo bool
	lookupCalls   int
}

type importAdmissionHistoryErrorStore struct {
	importAPITestStore
}

func (store *importAdmissionHistoryErrorStore) ListImportReviewTransactions(
	context.Context,
	string,
	string,
	verticalslice.ImportReviewHistoryFilter,
) ([]verticalslice.Transaction, error) {
	return nil, errors.New("history lookup unavailable")
}

func (store *importAdmissionLookupStore) LookupReplayArtifact(
	context.Context,
	verticalslice.CommandContext,
	string,
) (verticalslice.CommandReplayArtifact, bool, error) {
	store.lookupCalls++
	if store.lookupErr != nil {
		return verticalslice.CommandReplayArtifact{}, false, store.lookupErr
	}
	if store.completeOnTwo && store.lookupCalls == 2 {
		return store.completed, true, nil
	}
	return verticalslice.CommandReplayArtifact{}, false, nil
}

func newImportAdmissionLookupApp(
	t *testing.T,
	lookupErr error,
	completed verticalslice.CommandReplayArtifact,
	completeOnTwo bool,
) (*importAdmissionLookupStore, *API, *fiber.App) {
	t.Helper()
	store := &importAdmissionLookupStore{lookupErr: lookupErr, completed: completed, completeOnTwo: completeOnTwo}
	api, app := newImportAdmissionAppForStore(t, store)
	return store, api, app
}

func newImportAdmissionAppForStore(t *testing.T, store verticalslice.Store) (*API, *fiber.App) {
	t.Helper()
	secret, err := normalizedImportReviewSecret([]byte("oi-if-001-import-admission-secret-32-bytes"))
	if err != nil {
		t.Fatalf("normalize import review secret: %v", err)
	}
	api := &API{
		service:                 verticalslice.NewService(store, fixedHTTPClock{}),
		allowDevelopmentSubject: true,
		authLimiter:             newAuthRateLimiter(20, time.Minute),
		importAdmission:         newDefaultImportAdmission(),
		importReviewSecret:      secret,
		paginationCursorSecret:  derivePaginationCursorSecret(secret),
	}
	return api, newReplayApp(api)
}

func importReviewRequest(body string) *http.Request {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/portfolios/00000000-0000-4000-8000-000000000002/imports/review", bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	return request
}

func importReviewJSONRequest(csvPayload string) *http.Request {
	return importReviewRequest(`{"sourceAccountLabel":"Manual CSV","csvPayload":` + quote(csvPayload) + `}`)
}

func importAppendRequest(body []byte, key string) *http.Request {
	request := httptest.NewRequest(http.MethodPost, stage336ImportPath, bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", key)
	return request
}

func assertImportCapacityFullyReleased(t *testing.T, api *API) {
	t.Helper()
	first, err := api.acquireImportCapacity()
	if err != nil {
		t.Fatalf("first capacity slot leaked: %v", err)
	}
	second, err := api.acquireImportCapacity()
	if err != nil {
		first()
		t.Fatalf("second capacity slot leaked: %v", err)
	}
	if _, err := api.acquireImportCapacity(); !errors.Is(err, errImportCapacityExhausted) {
		first()
		second()
		t.Fatalf("capacity pool did not remain bounded after handler exit: %v", err)
	}
	first()
	second()
}

func TestImportAdmissionConstructorParity(t *testing.T) {
	apiNode, err := parser.ParseFile(token.NewFileSet(), "api.go", nil, 0)
	if err != nil {
		t.Fatalf("parse api.go: %v", err)
	}
	for _, name := range []string{"New", "NewDevelopment"} {
		constructor := findReplayConstructor(t, apiNode, name)
		if got := countReplayConstructorCalls(constructor.Body, "newDefaultImportAdmission"); got != 1 {
			t.Fatalf("%s newDefaultImportAdmission calls=%d, want 1", name, got)
		}
	}

	replayNode, err := parser.ParseFile(token.NewFileSet(), "replay_app.go", nil, 0)
	if err != nil {
		t.Fatalf("parse replay_app.go: %v", err)
	}
	for _, name := range []string{
		"NewReplay", "NewReplayWithCorporateActionProvider", "NewReplayWithCorporateActionProviderAndHTTPNetworkConfig", "NewReplayRuntime",
		"NewDevelopmentReplay", "NewDevelopmentReplayWithCorporateActionProvider", "NewDevelopmentReplayWithCorporateActionProviderAndHTTPNetworkConfig", "NewDevelopmentReplayRuntime",
	} {
		constructor := findReplayConstructor(t, replayNode, name)
		if countReplayConstructorCalls(constructor.Body, "newReplayWithCorporateActionProviderAndHTTPNetworkConfig")+
			countReplayConstructorCalls(constructor.Body, "newDevelopmentReplayWithCorporateActionProviderAndHTTPNetworkConfig")+
			countReplayConstructorCalls(constructor.Body, "NewReplayWithCorporateActionProvider")+
			countReplayConstructorCalls(constructor.Body, "NewDevelopmentReplayWithCorporateActionProvider") != 1 {
			t.Fatalf("%s no longer delegates to one import-admission-wired replay constructor", name)
		}
	}
	for _, name := range []string{
		"newReplayWithCorporateActionProviderAndHTTPNetworkConfig",
		"newDevelopmentReplayWithCorporateActionProviderAndHTTPNetworkConfig",
	} {
		constructor := findReplayConstructor(t, replayNode, name)
		if got := countReplayConstructorCalls(constructor.Body, "newDefaultImportAdmission"); got != 1 {
			t.Fatalf("%s newDefaultImportAdmission calls=%d, want 1", name, got)
		}
	}

	composition, err := os.ReadFile("../../cmd/api/main.go")
	if err != nil {
		t.Fatalf("read cmd/api production composition: %v", err)
	}
	if !strings.Contains(string(composition), "httpapi.NewReplayRuntime(") || !strings.Contains(string(composition), "httpapi.NewDevelopmentReplayRuntime(") {
		t.Fatal("cmd/api no longer composes production and development API through admission-wired replay runtimes")
	}
}

func TestImportAdmissionOpenAPIContract(t *testing.T) {
	specification, err := os.ReadFile("../../../openapi/openapi.yaml")
	if err != nil {
		t.Fatalf("read OpenAPI: %v", err)
	}
	responses, err := os.ReadFile("../../../openapi/components/responses.yaml")
	if err != nil {
		t.Fatalf("read response components: %v", err)
	}
	paths := string(specification)
	appendStart := strings.Index(paths, "/api/v1/portfolios/{portfolioId}/imports/append:")
	appendEnd := strings.Index(paths[appendStart:], "\n  /api/v1/dividends/calendar:")
	if appendStart < 0 || appendEnd < 0 {
		t.Fatal("cannot isolate import append OpenAPI operation")
	}
	appendOperation := paths[appendStart : appendStart+appendEnd]
	if strings.Count(appendOperation, "\"503\":") != 1 || !strings.Contains(appendOperation, "ImportAppendUnavailable") {
		t.Fatalf("append operation must expose exactly one combined 503 response: %s", appendOperation)
	}
	reviewStart := strings.Index(paths, "/api/v1/portfolios/{portfolioId}/imports/review:")
	reviewEnd := strings.Index(paths[reviewStart:], "\n  /api/v1/portfolios/{portfolioId}/imports/append:")
	if reviewStart < 0 || reviewEnd < 0 {
		t.Fatal("cannot isolate import review OpenAPI operation")
	}
	reviewOperation := paths[reviewStart : reviewStart+reviewEnd]
	if !strings.Contains(reviewOperation, "ImportAdmissionUnavailable") || !strings.Contains(reviewOperation, "RateLimited") {
		t.Fatalf("review contract does not preserve 429 and add combined admission 503: %s", reviewOperation)
	}
	components := string(responses)
	if !strings.Contains(components, "REPLAY_STATE_STALE") || !strings.Contains(components, "REPLAY_EPOCH_MISSING") ||
		!strings.Contains(components, "importCapacityExhausted") || !strings.Contains(components, "importAdmissionExhausted") {
		t.Fatal("combined append 503 response does not preserve replay-state, capacity, and execution-admission examples")
	}
}

func assertImportAdmissionErrorCode(t *testing.T, response *http.Response, want string) {
	t.Helper()
	var payload struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
		Meta struct {
			RequestID string `json:"requestId"`
			TraceID   string `json:"traceId"`
		} `json:"meta"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode admission error: %v", err)
	}
	if payload.Error.Code != want {
		t.Fatalf("admission error code = %q, want %q", payload.Error.Code, want)
	}
	if payload.Meta.RequestID == "" || payload.Meta.TraceID == "" {
		t.Fatalf("admission error omitted request/trace metadata: %+v", payload.Meta)
	}
}
