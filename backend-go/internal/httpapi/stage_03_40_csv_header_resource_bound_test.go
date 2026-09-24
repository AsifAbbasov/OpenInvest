package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/importer"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestStage340ReviewRejectsExcessiveStructureBeforeHistoryLookup(t *testing.T) {
	for _, testCase := range []struct {
		name    string
		payload string
	}{
		{name: "header", payload: stage340ExcessiveHeaderCSV()},
		{name: "data row", payload: stage340ExcessiveDataRowCSV()},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			store := &importAPITestStore{}
			app := NewDevelopment(verticalslice.NewService(store, fixedHTTPClock{}))
			request := httptest.NewRequest(http.MethodPost,
				"/api/v1/portfolios/00000000-0000-4000-8000-000000000002/imports/review",
				bytes.NewBufferString(`{"sourceAccountLabel":"Manual CSV","csvPayload":`+quote(testCase.payload)+`}`))
			request.Header.Set("Content-Type", "application/json")

			response, err := app.Test(request)
			if err != nil {
				t.Fatalf("review excessive %s: %v", testCase.name, err)
			}
			defer response.Body.Close()
			assertStage340ValidationError(t, response)
			if store.listTransactionsCalls != 0 || store.listImportReviewCalls != 0 {
				t.Fatalf("excessive %s reached history lookup: list=%d reviewHistory=%d", testCase.name, store.listTransactionsCalls, store.listImportReviewCalls)
			}
		})
	}
}

func TestStage340AppendRejectsExcessiveStructureBeforeMaterialStoreWork(t *testing.T) {
	for _, testCase := range []struct {
		name    string
		payload string
		key     string
	}{
		{name: "header", payload: stage340ExcessiveHeaderCSV(), key: "stage-03-40-header-bound-key-0001"},
		{name: "data row", payload: stage340ExcessiveDataRowCSV(), key: "stage-03-40-row-bound-key-000001"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			store := &importAPITestStore{}
			app := NewDevelopment(verticalslice.NewService(store, fixedHTTPClock{}))
			body := `{"sourceAccountLabel":"Manual CSV","sourceFileHash":"` + importPayloadHash(testCase.payload) + `",` +
				`"reviewToken":"` + strings.Repeat("a", 32) + `","csvPayload":` + quote(testCase.payload) +
				`,"decisions":[{"rowNumber":2,"rowHash":"` + strings.Repeat("b", 64) + `","action":"APPROVE"}]}`
			request := httptest.NewRequest(http.MethodPost,
				"/api/v1/portfolios/00000000-0000-4000-8000-000000000002/imports/append",
				bytes.NewBufferString(body))
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Idempotency-Key", testCase.key)

			response, err := app.Test(request)
			if err != nil {
				t.Fatalf("append excessive %s: %v", testCase.name, err)
			}
			defer response.Body.Close()
			assertStage340ValidationError(t, response)
			if store.listTransactionsCalls != 0 || store.listImportReviewCalls != 0 || store.appendImportedCalls != 0 {
				t.Fatalf("excessive %s reached material store work: list=%d reviewHistory=%d append=%d", testCase.name, store.listTransactionsCalls, store.listImportReviewCalls, store.appendImportedCalls)
			}
		})
	}
}

func TestStage340V2StructuralTokenCannotAuthorizeFreshWriteButCompletedReplayRemainsExact(t *testing.T) {
	for _, testCase := range []struct {
		name    string
		payload string
		key     string
	}{
		{name: "header", payload: stage340ExcessiveHeaderCSV(), key: "stage-03-40-v2-header-new-key-01"},
		{name: "data row", payload: stage340ExcessiveDataRowCSV(), key: "stage-03-40-v2-row-new-key-000001"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			store, _, api, app, _ := newStage336ReplayApp(t)
			body, _ := stage336HistoricalAppendRequestForParserVersion(t, api, testCase.payload, 2)
			response := stage336Append(t, app, body, testCase.key)
			defer response.Body.Close()
			assertStage340ValidationError(t, response)
			if store.appendReplayCalls != 0 || store.lookupCalls != 1 {
				t.Fatalf("v2 token must not authorize a fresh write: append=%d lookup=%d", store.appendReplayCalls, store.lookupCalls)
			}

			store, service, api, app, advance := newStage336ReplayApp(t)
			body, request := stage336HistoricalAppendRequestForParserVersion(t, api, testCase.payload, 2)
			artifact := seedStage336CompletedImport(t, service, request)
			advance(20 * time.Minute)
			response = stage336Append(t, app, body, stage336ImportKey)
			defer response.Body.Close()
			replayedBody, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatalf("read historic v2 replay: %v", err)
			}
			if response.StatusCode != artifact.StatusCode || !bytes.Equal(replayedBody, artifact.Body) {
				t.Fatalf("historic v2 replay changed: status=%d body=%s", response.StatusCode, replayedBody)
			}
			if response.Header.Get("X-Request-ID") != artifact.RequestID || response.Header.Get("X-Trace-ID") != artifact.TraceID {
				t.Fatalf("historic v2 replay changed technical identity")
			}
			if store.appendReplayCalls != 1 || store.lookupCalls != 1 {
				t.Fatalf("historic v2 replay must remain read-only: append=%d lookup=%d", store.appendReplayCalls, store.lookupCalls)
			}
		})
	}
}

func assertStage340ValidationError(t *testing.T, response *http.Response) {
	t.Helper()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected validation status %d, got %d", http.StatusBadRequest, response.StatusCode)
	}
	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode validation response: %v", err)
	}
	if body.Error.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR, got %q", body.Error.Code)
	}
}

func stage340ExcessiveHeaderCSV() string {
	payload := "transaction_type,ticker,quantity,unit_price,gross_amount,commission,tax,trade_date,settlement_date,currency,broker_operation_id,note,unexpected_extension\n" +
		"BUY,SBER,2.00000000,100.00000000,200.00000000,1.00000000,0.00000000,2026-01-10,2026-01-13,RUB,stage340-header-bound,Imported buy,ignored\n"
	if len(payload) >= 2*1024*1024 {
		panic("stage340 fixture must remain below the import payload bound")
	}
	if strings.Count(strings.Split(payload, "\n")[0], ",")+1 != importer.MaxCSVHeaderColumns+1 {
		panic("stage340 fixture must exceed the CSV header column bound by exactly one")
	}
	return payload
}

func stage340ExcessiveDataRowCSV() string {
	extras := make([]string, 1024)
	for index := range extras {
		extras[index] = "row_extra_" + strconv.Itoa(index)
	}
	payload := csvHeaderForHTTP +
		"BUY,SBER,2.00000000,100.00000000,200.00000000,1.00000000,0.00000000,2026-01-10,2026-01-13,RUB,stage340-row-bound,Imported buy," + strings.Join(extras, ",") + "\n"
	if len(payload) >= 2*1024*1024 {
		panic("stage340 row fixture must remain below the import payload bound")
	}
	lines := strings.Split(payload, "\n")
	if strings.Count(lines[0], ",")+1 != importer.MaxCSVHeaderColumns {
		panic("stage340 row fixture must retain the canonical header width")
	}
	if strings.Count(lines[1], ",")+1 <= importer.ExpectedCSVDataRowFields {
		panic("stage340 row fixture must exceed the data-row field bound")
	}
	return payload
}
