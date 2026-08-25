package httpapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/openinvest/openinvest/backend-go/internal/importer"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

const stage336ImportPath = "/api/v1/portfolios/00000000-0000-4000-8000-000000000002/imports/append"
const stage336ImportKey = "stage-03-36-legacy-decimal-key-01"

func TestStage0336CompletedPriorParserImportReplaysExactly(t *testing.T) {
	for _, testCase := range []struct {
		name string
		csv  string
	}{
		{
			name: "previously permissive leading zero",
			csv: csvHeaderForHTTP +
				"BUY,SBER,2.00000000,001.25,2.50000000,0.00000000,0.00000000,2026-01-10,2026-01-13,RUB,legacy-leading-zero,legacy\n",
		},
		{
			name: "contract conforming decimal",
			csv: csvHeaderForHTTP +
				"BUY,SBER,2.00000000,1.25,2.50000000,0.00000000,0.00000000,2026-01-10,2026-01-13,RUB,legacy-conforming,legacy\n",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			store, service, api, app, advance := newStage336ReplayApp(t)
			body, request := stage336HistoricalAppendRequest(t, api, testCase.csv)
			artifact := seedStage336CompletedImport(t, service, request)
			advance(20 * time.Minute)

			response := stage336Append(t, app, body, stage336ImportKey)
			defer response.Body.Close()
			bodyAfterDeploy, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatalf("read replay body: %v", err)
			}
			if response.StatusCode != artifact.StatusCode || !bytes.Equal(bodyAfterDeploy, artifact.Body) {
				t.Fatalf("completed replay changed response: status=%d body=%s", response.StatusCode, bodyAfterDeploy)
			}
			if response.Header.Get("X-Request-ID") != artifact.RequestID || response.Header.Get("X-Trace-ID") != artifact.TraceID {
				t.Fatalf("completed replay changed technical identity: request=%q trace=%q", response.Header.Get("X-Request-ID"), response.Header.Get("X-Trace-ID"))
			}
			if store.appendReplayCalls != 1 || store.lookupCalls != 1 {
				t.Fatalf("completed replay must be read-only: append=%d lookup=%d", store.appendReplayCalls, store.lookupCalls)
			}
		})
	}
}

func TestStage0336PriorParserTokenCannotAuthorizeFreshOrTamperedReplay(t *testing.T) {
	store, service, api, app, _ := newStage336ReplayApp(t)
	body, request := stage336HistoricalAppendRequest(t, api, csvHeaderForHTTP+
		"BUY,SBER,2.00000000,001.25,2.50000000,0.00000000,0.00000000,2026-01-10,2026-01-13,RUB,legacy-guard,legacy\n")
	seedStage336CompletedImport(t, service, request)

	for _, testCase := range []struct {
		name       string
		key        string
		lookupCall int
		body       func([]byte) []byte
	}{
		{
			name:       "new idempotency key",
			key:        "stage-03-36-legacy-decimal-key-02",
			lookupCall: 1,
			body:       func(body []byte) []byte { return body },
		},
		{
			name: "tampered token HMAC",
			key:  stage336ImportKey,
			body: func(body []byte) []byte {
				request := decodeJSONObject(t, body)
				request["reviewToken"] = request["reviewToken"].(string) + "x"
				return encodeJSONObject(t, request)
			},
		},
		{
			name: "changed source context",
			key:  stage336ImportKey,
			body: func(body []byte) []byte {
				request := decodeJSONObject(t, body)
				request["sourceAccountLabel"] = "Other account"
				return encodeJSONObject(t, request)
			},
		},
		{
			name: "changed raw CSV payload",
			key:  stage336ImportKey,
			body: func(body []byte) []byte {
				request := decodeJSONObject(t, body)
				changedPayload := strings.Replace(request["csvPayload"].(string), ",legacy\n", ",changed\n", 1)
				request["csvPayload"] = changedPayload
				request["sourceFileHash"] = importPayloadHash(changedPayload)
				return encodeJSONObject(t, request)
			},
		},
		{
			name: "changed row identity",
			key:  stage336ImportKey,
			body: func(body []byte) []byte {
				request := decodeJSONObject(t, body)
				decisions := request["decisions"].([]any)
				decisions[0].(map[string]any)["rowHash"] = "0000000000000000000000000000000000000000000000000000000000000000"
				return encodeJSONObject(t, request)
			},
		},
		{
			name: "changed decision",
			key:  stage336ImportKey,
			body: func(body []byte) []byte {
				request := decodeJSONObject(t, body)
				decisions := request["decisions"].([]any)
				decisions[0].(map[string]any)["action"] = importer.DecisionIgnore
				return encodeJSONObject(t, request)
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			lookupsBefore := store.lookupCalls
			response := stage336Append(t, app, testCase.body(body), testCase.key)
			defer response.Body.Close()
			if response.StatusCode != http.StatusBadRequest {
				t.Fatalf("old parser token must not recover %s: status=%d", testCase.name, response.StatusCode)
			}
			if store.appendReplayCalls != 1 {
				t.Fatalf("old parser token executed a fresh append for %s: %d", testCase.name, store.appendReplayCalls)
			}
			if got := store.lookupCalls - lookupsBefore; got != testCase.lookupCall {
				t.Fatalf("unexpected replay lookup count for %s: got %d want %d", testCase.name, got, testCase.lookupCall)
			}
		})
	}

	if _, found := api.recoverCompletedImportReplay(
		context.Background(), verticalslice.RequestContext{}, "another-subject", request.portfolioID,
		stage336ImportKey, stage336ImportPath, request.sourceAccountLabel, request.fileHash,
		request.reviewToken, request.csvPayload, request.decisions,
	); found {
		t.Fatal("prior parser token recovered an artifact for another principal")
	}
}

type stage336HistoricalRequest struct {
	portfolioID        string
	sourceAccountLabel string
	fileHash           string
	reviewToken        string
	csvPayload         string
	decisions          []importer.Decision
	appendRequest      verticalslice.AppendImportBatchRequest
}

func newStage336ReplayApp(t *testing.T) (*stage32ImportReplayStore, *verticalslice.Service, *API, *fiber.App, func(time.Duration)) {
	t.Helper()
	store := &stage32ImportReplayStore{}
	service := verticalslice.NewService(store, fixedHTTPClock{})
	secret, err := normalizedImportReviewSecret([]byte("stage-03-36-import-replay-test-secret-32-bytes"))
	if err != nil {
		t.Fatalf("normalize import review secret: %v", err)
	}
	now := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)
	api := &API{
		service:                 service,
		allowDevelopmentSubject: true,
		authLimiter:             newAuthRateLimiter(20, time.Minute),
		importReviewSecret:      secret,
		paginationCursorSecret:  derivePaginationCursorSecret(secret),
		now:                     func() time.Time { return now },
	}
	return store, service, api, newReplayApp(api), func(elapsed time.Duration) { now = now.Add(elapsed) }
}

func stage336HistoricalAppendRequest(t *testing.T, api *API, csvPayload string) ([]byte, stage336HistoricalRequest) {
	t.Helper()
	portfolioID := "00000000-0000-4000-8000-000000000002"
	sourceAccountLabel := "Manual CSV"
	fileHash := importPayloadHash(csvPayload)
	review, err := importer.ReviewCSVForParserVersion(importer.ReviewRequest{
		SubjectID:          devSubjectID,
		PortfolioID:        portfolioID,
		SourceKind:         importer.SourceKindUserUploadedFile,
		SourceAccountLabel: sourceAccountLabel,
		FileHash:           fileHash,
		Reader:             bytes.NewBufferString(csvPayload),
	}, 1)
	if err != nil || len(review.Rows) != 1 || review.Rows[0].Status != importer.ReviewStatusAppendable {
		t.Fatalf("reconstruct legacy review: review=%+v err=%v", review, err)
	}
	decisions := []importer.Decision{{RowNumber: review.Rows[0].RowNumber, RowHash: review.Rows[0].RowHash, Action: importer.DecisionApprove}}
	appendRequests, err := importer.BuildAppendRequests(review, decisions)
	if err != nil || len(appendRequests) != 1 {
		t.Fatalf("build legacy append request: requests=%+v err=%v", appendRequests, err)
	}
	token := signStage336ParserV1Token(t, api, devSubjectID, review)
	encoded, err := json.Marshal(map[string]any{
		"sourceAccountLabel": sourceAccountLabel,
		"sourceFileHash":     fileHash,
		"reviewToken":        token,
		"csvPayload":         csvPayload,
		"decisions": []map[string]any{{
			"rowNumber": review.Rows[0].RowNumber,
			"rowHash":   review.Rows[0].RowHash,
			"action":    importer.DecisionApprove,
		}},
	})
	if err != nil {
		t.Fatalf("encode historic append body: %v", err)
	}
	return encoded, stage336HistoricalRequest{
		portfolioID:        portfolioID,
		sourceAccountLabel: sourceAccountLabel,
		fileHash:           fileHash,
		reviewToken:        token,
		csvPayload:         csvPayload,
		decisions:          decisions,
		appendRequest: verticalslice.AppendImportBatchRequest{
			PortfolioID: portfolioID, Transactions: appendRequests, SourceKind: review.SourceKind,
			SourceAccountLabel: review.SourceAccountLabel, SourceFileHash: review.FileHash,
			Decisions: stage0332ImportDecisions(decisions),
		},
	}
}

func signStage336ParserV1Token(t *testing.T, api *API, subjectID string, review importer.Review) string {
	t.Helper()
	digest, err := importer.ReviewSemanticDigestForParserVersion(review, 1)
	if err != nil {
		t.Fatalf("legacy review digest: %v", err)
	}
	rows := make([]importReviewTokenRowIdentity, 0, len(review.Rows))
	appendableRows := make([]int, 0, len(review.Rows))
	for _, row := range review.Rows {
		rows = append(rows, importReviewTokenRowIdentity{RowNumber: row.RowNumber, RowHash: row.RowHash})
		if row.Status == importer.ReviewStatusAppendable {
			appendableRows = append(appendableRows, row.RowNumber)
		}
	}
	issuedAt := api.nowUTC()
	payload := importReviewTokenPayload{
		Version: importReviewTokenVersion, ParserVersion: 1, IssuedAt: issuedAt.Unix(), ExpiresAt: issuedAt.Add(importReviewTokenTTL).Unix(),
		SubjectID: subjectID, PortfolioID: review.PortfolioID, SourceKind: review.SourceKind, SourceAccountLabel: review.SourceAccountLabel,
		SourceFileHash: review.FileHash, ParserReviewDigest: digest, FinalReviewDigest: digest, AppendableRows: appendableRows, Rows: rows,
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("encode legacy token: %v", err)
	}
	bodyPart := base64.RawURLEncoding.EncodeToString(encoded)
	return bodyPart + "." + api.signImportReviewTokenPart(bodyPart)
}

func seedStage336CompletedImport(t *testing.T, service *verticalslice.Service, request stage336HistoricalRequest) verticalslice.CommandReplayArtifact {
	t.Helper()
	artifact := verticalslice.CommandReplayArtifact{
		StatusCode: http.StatusCreated,
		Body:       []byte(`{"data":{"historicalReplay":true}}`),
		RequestID:  "11111111-1111-4111-8111-111111111111",
		TraceID:    "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	_, stored, err := service.AppendImportedTransactionsWithReplay(
		context.Background(), verticalslice.RequestContext{RequestID: artifact.RequestID, TraceID: artifact.TraceID}, devSubjectID,
		stage336ImportKey, stage336ImportPath, request.appendRequest,
		func([]verticalslice.Transaction) (verticalslice.CommandReplayArtifact, error) { return artifact, nil },
	)
	if err != nil {
		t.Fatalf("seed completed historic import: %v", err)
	}
	return stored
}

func stage336Append(t *testing.T, app *fiber.App, body []byte, key string) *http.Response {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, stage336ImportPath, bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", key)
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("append request: %v", err)
	}
	return response
}

const csvHeaderForHTTP = "transaction_type,ticker,quantity,unit_price,gross_amount,commission,tax,trade_date,settlement_date,currency,broker_operation_id,note\n"
