package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/importer"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

const importCSV = "transaction_type,ticker,quantity,unit_price,gross_amount,commission,tax,trade_date,settlement_date,currency,broker_operation_id,note\n" +
	"BUY,SBER,2.00000000,100.00000000,200.00000000,1.00000000,0.00000000,2026-01-10,2026-01-13,RUB,broker-row-1,Imported buy\n"

func TestNewRejectsMissingOrShortImportReviewTokenSecret(t *testing.T) {
	for _, secret := range [][]byte{nil, []byte("too-short")} {
		app, err := New(nil, nil, secret)
		if err == nil {
			t.Fatalf("expected import review token secret validation error for %q", secret)
		}
		if app != nil {
			t.Fatalf("expected no app when import review token secret is invalid")
		}
	}
}

func TestImportReviewTokenRejectsContextTampering(t *testing.T) {
	secret, err := normalizedImportReviewSecret([]byte("test-import-review-token-secret-32-bytes"))
	if err != nil {
		t.Fatalf("normalize import review token secret: %v", err)
	}
	now := time.Date(2026, 8, 23, 8, 0, 0, 0, time.UTC)
	api := &API{importReviewSecret: secret, now: func() time.Time { return now }}
	rowHash := strings.Repeat("a", 64)
	review := importer.Review{
		PortfolioID:        "portfolio-a",
		SourceKind:         importer.SourceKindUserUploadedFile,
		SourceAccountLabel: "Broker A",
		FileHash:           strings.Repeat("b", 64),
		Rows:               []importer.RowReview{{RowNumber: 2, RowHash: rowHash, Status: importer.ReviewStatusAppendable}},
	}
	review.Summary = importer.Summary{TotalRows: 1, AppendableRows: 1}
	token, err := api.signImportReviewToken("subject-a", review, review)
	if err != nil {
		t.Fatalf("sign import review token: %v", err)
	}
	decision := []importer.Decision{{RowNumber: 2, RowHash: rowHash, Action: importer.DecisionApprove}}

	for _, testCase := range []struct {
		name      string
		subjectID string
		portfolio string
		label     string
		fileHash  string
		decisions []importer.Decision
	}{
		{name: "subject", subjectID: "subject-b", portfolio: "portfolio-a", label: "Broker A", fileHash: review.FileHash, decisions: decision},
		{name: "portfolio", subjectID: "subject-a", portfolio: "portfolio-b", label: "Broker A", fileHash: review.FileHash, decisions: decision},
		{name: "source label", subjectID: "subject-a", portfolio: "portfolio-a", label: "Broker B", fileHash: review.FileHash, decisions: decision},
		{name: "file hash", subjectID: "subject-a", portfolio: "portfolio-a", label: "Broker A", fileHash: strings.Repeat("c", 64), decisions: decision},
		{name: "row identity", subjectID: "subject-a", portfolio: "portfolio-a", label: "Broker A", fileHash: review.FileHash, decisions: []importer.Decision{{RowNumber: 2, RowHash: strings.Repeat("d", 64), Action: importer.DecisionApprove}}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			err := api.verifyImportReviewToken(token, testCase.subjectID, testCase.portfolio, importer.SourceKindUserUploadedFile, testCase.label, testCase.fileHash, review, testCase.decisions)
			if !errors.Is(err, importer.ErrUnsafeAppend) {
				t.Fatalf("expected unsafe append error, got %v", err)
			}
		})
	}

	err = api.verifyImportReviewToken(token+"x", "subject-a", "portfolio-a", importer.SourceKindUserUploadedFile, "Broker A", review.FileHash, review, decision)
	if !errors.Is(err, importer.ErrUnsafeAppend) {
		t.Fatalf("expected unsafe append error for a tampered signature, got %v", err)
	}
}

func TestAssetSearchReturnsCatalogSummariesWithoutPriceOrSource(t *testing.T) {
	app := NewDevelopment(verticalslice.NewService(&importAPITestStore{}, fixedHTTPClock{}))
	request := httptest.NewRequest(http.MethodGet, "/api/v1/assets/search?query=sber&limit=20", nil)

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request asset search endpoint: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.StatusCode)
	}
	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	text := string(encoded)
	if !strings.Contains(text, `"ticker":"SBER"`) || !strings.Contains(text, `"lastPrice":null`) {
		t.Fatalf("expected SBER search result with null lastPrice, got %s", text)
	}
	if strings.Contains(text, "source") || strings.Contains(text, "EXAMPLE_") {
		t.Fatalf("runtime asset search must not emit source provenance or example identifiers: %s", text)
	}
}

func TestPortfolioAndTransactionListsExposeReachableNextPages(t *testing.T) {
	now := time.Date(2026, 8, 8, 12, 0, 0, 0, time.UTC)
	zero := verticalslice.ZeroMoney()
	store := &importAPITestStore{
		portfolios: []verticalslice.Portfolio{
			{ID: "00000000-0000-4000-8000-000000000013", Name: "Three", BaseCurrency: verticalslice.RUB, Version: 1, CreatedAt: now, UpdatedAt: now},
			{ID: "00000000-0000-4000-8000-000000000012", Name: "Two", BaseCurrency: verticalslice.RUB, Version: 1, CreatedAt: now, UpdatedAt: now},
			{ID: "00000000-0000-4000-8000-000000000011", Name: "One", BaseCurrency: verticalslice.RUB, Version: 1, CreatedAt: now, UpdatedAt: now},
		},
		existingTransactions: []verticalslice.Transaction{
			{ID: "00000000-0000-4000-8000-000000000021", EntryID: "00000000-0000-4000-8000-000000000031", PortfolioID: "00000000-0000-4000-8000-000000000002", TransactionType: "DEPOSIT", Status: "ACTIVE", GrossAmount: verticalslice.Money{Amount: decimal.Must("100.00000000"), Currency: verticalslice.RUB}, Commission: zero, Tax: zero, TradeDate: "2026-08-03", Revision: 1, CreatedAt: now, UpdatedAt: now},
			{ID: "00000000-0000-4000-8000-000000000022", EntryID: "00000000-0000-4000-8000-000000000032", PortfolioID: "00000000-0000-4000-8000-000000000002", TransactionType: "DEPOSIT", Status: "ACTIVE", GrossAmount: verticalslice.Money{Amount: decimal.Must("200.00000000"), Currency: verticalslice.RUB}, Commission: zero, Tax: zero, TradeDate: "2026-08-02", Revision: 1, CreatedAt: now, UpdatedAt: now},
			{ID: "00000000-0000-4000-8000-000000000023", EntryID: "00000000-0000-4000-8000-000000000033", PortfolioID: "00000000-0000-4000-8000-000000000002", TransactionType: "DEPOSIT", Status: "ACTIVE", GrossAmount: verticalslice.Money{Amount: decimal.Must("300.00000000"), Currency: verticalslice.RUB}, Commission: zero, Tax: zero, TradeDate: "2026-08-01", Revision: 1, CreatedAt: now, UpdatedAt: now},
		},
	}
	app := NewDevelopment(verticalslice.NewService(store, fixedHTTPClock{}))

	portfolioRequest := httptest.NewRequest(http.MethodGet, "/api/v1/portfolios?limit=2", nil)
	portfolioResponse, err := app.Test(portfolioRequest)
	if err != nil {
		t.Fatalf("request portfolio page: %v", err)
	}
	defer portfolioResponse.Body.Close()
	var portfolios struct {
		Data listData[portfolioDTO] `json:"data"`
	}
	if err := json.NewDecoder(portfolioResponse.Body).Decode(&portfolios); err != nil {
		t.Fatalf("decode portfolio page: %v", err)
	}
	if len(portfolios.Data.Items) != 2 || !portfolios.Data.Pagination.HasMore || portfolios.Data.Pagination.NextCursor == nil {
		t.Fatalf("expected first portfolio page and a next cursor, got %+v", portfolios.Data)
	}
	if strings.Contains(*portfolios.Data.Pagination.NextCursor, "offset") {
		t.Fatalf("expected an opaque portfolio cursor, got %q", *portfolios.Data.Pagination.NextCursor)
	}
	secondPortfolioRequest := httptest.NewRequest(http.MethodGet, "/api/v1/portfolios?limit=2&cursor="+*portfolios.Data.Pagination.NextCursor, nil)
	secondPortfolioResponse, err := app.Test(secondPortfolioRequest)
	if err != nil {
		t.Fatalf("request second portfolio page: %v", err)
	}
	defer secondPortfolioResponse.Body.Close()
	var secondPortfolios struct {
		Data listData[portfolioDTO] `json:"data"`
	}
	if err := json.NewDecoder(secondPortfolioResponse.Body).Decode(&secondPortfolios); err != nil {
		t.Fatalf("decode second portfolio page: %v", err)
	}
	if len(secondPortfolios.Data.Items) != 1 || secondPortfolios.Data.Items[0].Name != "One" || secondPortfolios.Data.Pagination.HasMore {
		t.Fatalf("expected remaining portfolio page, got %+v", secondPortfolios.Data)
	}

	transactionRequest := httptest.NewRequest(http.MethodGet, "/api/v1/portfolios/00000000-0000-4000-8000-000000000002/transactions?limit=2", nil)
	transactionResponse, err := app.Test(transactionRequest)
	if err != nil {
		t.Fatalf("request transaction page: %v", err)
	}
	defer transactionResponse.Body.Close()
	var transactions struct {
		Data listData[transactionDTO] `json:"data"`
	}
	if err := json.NewDecoder(transactionResponse.Body).Decode(&transactions); err != nil {
		t.Fatalf("decode transaction page: %v", err)
	}
	if len(transactions.Data.Items) != 2 || !transactions.Data.Pagination.HasMore || transactions.Data.Pagination.NextCursor == nil {
		t.Fatalf("expected first transaction page and a next cursor, got %+v", transactions.Data)
	}
	if strings.Contains(*transactions.Data.Pagination.NextCursor, "offset") {
		t.Fatalf("expected an opaque transaction cursor, got %q", *transactions.Data.Pagination.NextCursor)
	}
	secondTransactionRequest := httptest.NewRequest(http.MethodGet, "/api/v1/portfolios/00000000-0000-4000-8000-000000000002/transactions?limit=2&cursor="+*transactions.Data.Pagination.NextCursor, nil)
	secondTransactionResponse, err := app.Test(secondTransactionRequest)
	if err != nil {
		t.Fatalf("request second transaction page: %v", err)
	}
	defer secondTransactionResponse.Body.Close()
	var secondTransactions struct {
		Data listData[transactionDTO] `json:"data"`
	}
	if err := json.NewDecoder(secondTransactionResponse.Body).Decode(&secondTransactions); err != nil {
		t.Fatalf("decode second transaction page: %v", err)
	}
	if len(secondTransactions.Data.Items) != 1 || secondTransactions.Data.Items[0].ID != "00000000-0000-4000-8000-000000000023" || secondTransactions.Data.Pagination.HasMore {
		t.Fatalf("expected remaining transaction page, got %+v", secondTransactions.Data)
	}
}

func TestPaginationCursorRejectsTamperingAndScopeChanges(t *testing.T) {
	secret, err := normalizedImportReviewSecret([]byte("test-pagination-cursor-secret-32-bytes"))
	if err != nil {
		t.Fatalf("normalize cursor secret: %v", err)
	}
	api := &API{paginationCursorSecret: derivePaginationCursorSecret(secret)}
	portfolio := verticalslice.Portfolio{
		ID:        "00000000-0000-4000-8000-000000000010",
		UpdatedAt: time.Date(2026, 8, 8, 12, 0, 0, 0, time.UTC),
	}
	cursor, err := api.encodePortfolioCursor("subject-a", portfolio)
	if err != nil {
		t.Fatalf("encode cursor: %v", err)
	}
	if _, err := api.decodePortfolioCursor(cursor+"x", "subject-a"); !errors.Is(err, verticalslice.ErrInvalidInput) {
		t.Fatalf("expected tampered cursor rejection, got %v", err)
	}
	if _, err := api.decodePortfolioCursor(cursor, "subject-b"); !errors.Is(err, verticalslice.ErrInvalidInput) {
		t.Fatalf("expected cross-subject cursor rejection, got %v", err)
	}

	transaction := verticalslice.Transaction{
		ID:        "00000000-0000-4000-8000-000000000020",
		EntryID:   "00000000-0000-4000-8000-000000000021",
		TradeDate: "2026-08-08",
	}
	filter := verticalslice.TransactionFilter{TransactionType: "DEPOSIT", FromDate: "2026-08-01"}
	transactionCursor, err := api.encodeTransactionCursor("subject-a", "portfolio-a", filter, transaction)
	if err != nil {
		t.Fatalf("encode transaction cursor: %v", err)
	}
	if err := api.applyTransactionCursor(transactionCursor, "subject-a", "portfolio-a", &verticalslice.TransactionFilter{TransactionType: "BUY", FromDate: "2026-08-01"}); !errors.Is(err, verticalslice.ErrInvalidInput) {
		t.Fatalf("expected changed filter rejection, got %v", err)
	}
	if err := api.applyTransactionCursor(transactionCursor, "subject-a", "portfolio-b", &verticalslice.TransactionFilter{TransactionType: "DEPOSIT", FromDate: "2026-08-01"}); !errors.Is(err, verticalslice.ErrInvalidInput) {
		t.Fatalf("expected changed portfolio rejection, got %v", err)
	}
	matchedFilter := verticalslice.TransactionFilter{TransactionType: "DEPOSIT", FromDate: "2026-08-01"}
	if err := api.applyTransactionCursor(transactionCursor, "subject-a", "portfolio-a", &matchedFilter); err != nil {
		t.Fatalf("expected matching transaction cursor to be accepted: %v", err)
	}
	if matchedFilter.BeforeEntryID != transaction.EntryID {
		t.Fatalf("expected transaction cursor to preserve internal entry ID %q, got %q", transaction.EntryID, matchedFilter.BeforeEntryID)
	}
}

func TestPrivatePaginationRejectsSuppliedEmptyCursors(t *testing.T) {
	app := NewDevelopment(verticalslice.NewService(&importAPITestStore{}, fixedHTTPClock{}))
	for _, path := range []string{
		"/api/v1/portfolios?cursor=",
		"/api/v1/portfolios?cursor=%20%20",
		"/api/v1/portfolios/00000000-0000-4000-8000-000000000002/transactions?cursor=",
		"/api/v1/portfolios/00000000-0000-4000-8000-000000000002/transactions?cursor=%20%20",
	} {
		t.Run(path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, path, nil)
			response, err := app.Test(request)
			if err != nil {
				t.Fatalf("request private listing: %v", err)
			}
			defer response.Body.Close()
			if response.StatusCode != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.StatusCode)
			}
		})
	}
}

func TestImportAppendValidationMatchesOpenAPIBounds(t *testing.T) {
	validHash := strings.Repeat("a", 64)
	validToken := strings.Repeat("a", 32)
	validDecision := []importDecisionDTO{{RowNumber: 2, RowHash: validHash, Action: "APPROVE"}}
	valid := importAppendRequestDTO{CSVPayload: importCSV, SourceFileHash: validHash, ReviewToken: validToken, Decisions: validDecision}
	if err := valid.validate(); err != nil {
		t.Fatalf("expected valid OpenAPI-shaped append request: %v", err)
	}
	for _, testCase := range []importAppendRequestDTO{
		{CSVPayload: importCSV, SourceFileHash: strings.ToUpper(validHash), ReviewToken: validToken, Decisions: validDecision},
		{CSVPayload: importCSV, SourceFileHash: validHash, ReviewToken: strings.Repeat("a", 31), Decisions: validDecision},
		{CSVPayload: importCSV, SourceFileHash: validHash, ReviewToken: strings.Repeat("a", maxImportReviewTokenBytes+1), Decisions: validDecision},
		{CSVPayload: importCSV, SourceFileHash: validHash, ReviewToken: validToken, Decisions: []importDecisionDTO{{RowNumber: 2, RowHash: strings.ToUpper(validHash), Action: "APPROVE"}}},
	} {
		if err := testCase.validate(); !errors.Is(err, verticalslice.ErrInvalidInput) {
			t.Fatalf("expected invalid OpenAPI boundary request, got %v", err)
		}
	}
}

func TestAssetSearchReturnsSignedOpaqueCursorForNextPage(t *testing.T) {
	app := NewDevelopment(verticalslice.NewService(&importAPITestStore{}, fixedHTTPClock{}))
	request := httptest.NewRequest(http.MethodGet, "/api/v1/assets/search?query=S&limit=1", nil)

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request asset search endpoint: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.StatusCode)
	}
	var payload struct {
		Data struct {
			Items      []assetSummaryDTO `json:"items"`
			Pagination paginationDTO     `json:"pagination"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Data.Items) != 1 || payload.Data.Items[0].Ticker != "SBER" {
		t.Fatalf("unexpected first page: %+v", payload.Data.Items)
	}
	if !payload.Data.Pagination.HasMore || payload.Data.Pagination.NextCursor == nil {
		t.Fatalf("expected next cursor, got %+v", payload.Data.Pagination)
	}
	if len(*payload.Data.Pagination.NextCursor) > maxPaginationCursorBytes || !strings.Contains(*payload.Data.Pagination.NextCursor, ".") {
		t.Fatalf("expected a bounded signed cursor, got %q", *payload.Data.Pagination.NextCursor)
	}

	secondRequest := httptest.NewRequest(http.MethodGet, "/api/v1/assets/search?query=S&limit=1&cursor="+*payload.Data.Pagination.NextCursor, nil)
	secondResponse, err := app.Test(secondRequest)
	if err != nil {
		t.Fatalf("request second asset search page: %v", err)
	}
	defer secondResponse.Body.Close()
	if secondResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected second status %d, got %d", http.StatusOK, secondResponse.StatusCode)
	}
	var secondPayload struct {
		Data struct {
			Items      []assetSummaryDTO `json:"items"`
			Pagination paginationDTO     `json:"pagination"`
		} `json:"data"`
	}
	if err := json.NewDecoder(secondResponse.Body).Decode(&secondPayload); err != nil {
		t.Fatalf("decode second response: %v", err)
	}
	if len(secondPayload.Data.Items) != 1 || secondPayload.Data.Items[0].Ticker != "SU26238RMFS4" {
		t.Fatalf("unexpected second page: %+v", secondPayload.Data.Items)
	}
	if secondPayload.Data.Pagination.HasMore || secondPayload.Data.Pagination.NextCursor != nil {
		t.Fatalf("expected final page, got %+v", secondPayload.Data.Pagination)
	}

	for _, path := range []string{
		"/api/v1/assets/search?query=S&limit=1&cursor=" + *payload.Data.Pagination.NextCursor + "x",
		"/api/v1/assets/search?query=SU&limit=1&cursor=" + *payload.Data.Pagination.NextCursor,
		"/api/v1/assets/search?query=S&assetType=STOCK&limit=1&cursor=" + *payload.Data.Pagination.NextCursor,
	} {
		response, err := app.Test(httptest.NewRequest(http.MethodGet, path, nil))
		if err != nil {
			t.Fatalf("request invalid asset cursor %s: %v", path, err)
		}
		response.Body.Close()
		if response.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected invalid asset cursor request to return %d, got %d", http.StatusBadRequest, response.StatusCode)
		}
	}
}

func TestAssetSearchRejectsMalformedAndOutOfRangeLimit(t *testing.T) {
	app := NewDevelopment(verticalslice.NewService(&importAPITestStore{}, fixedHTTPClock{}))
	for _, path := range []string{
		"/api/v1/assets/search?query=SBER&limit=abc",
		"/api/v1/assets/search?query=SBER&limit=",
		"/api/v1/assets/search?query=SBER&limit=%20",
		"/api/v1/assets/search?query=SBER&limit=0",
		"/api/v1/assets/search?query=SBER&limit=101",
	} {
		response, err := app.Test(httptest.NewRequest(http.MethodGet, path, nil))
		if err != nil {
			t.Fatalf("request asset search endpoint %s: %v", path, err)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected %s to return %d, got %d", path, http.StatusBadRequest, response.StatusCode)
		}
	}
}

func TestAssetSearchRejectsEmptyOptionalQueryParameters(t *testing.T) {
	app := NewDevelopment(verticalslice.NewService(&importAPITestStore{}, fixedHTTPClock{}))
	for _, path := range []string{
		"/api/v1/assets/search?query=SBER&assetType=",
		"/api/v1/assets/search?query=SBER&assetType=%20",
		"/api/v1/assets/search?query=SBER&cursor=",
		"/api/v1/assets/search?query=SBER&cursor=%20",
	} {
		response, err := app.Test(httptest.NewRequest(http.MethodGet, path, nil))
		if err != nil {
			t.Fatalf("request asset search endpoint %s: %v", path, err)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected %s to return %d, got %d", path, http.StatusBadRequest, response.StatusCode)
		}
	}
}

func TestAssetDetailIsDeferredUntilMandatorySourceAndDetailsExist(t *testing.T) {
	app := NewDevelopment(verticalslice.NewService(&importAPITestStore{}, fixedHTTPClock{}))
	request := httptest.NewRequest(http.MethodGet, "/api/v1/assets/SBER", nil)

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request asset detail endpoint: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, response.StatusCode)
	}
}

func TestAssetDetailInvalidTickerStillUsesFrozenNotFoundContract(t *testing.T) {
	app := NewDevelopment(verticalslice.NewService(&importAPITestStore{}, fixedHTTPClock{}))
	request := httptest.NewRequest(http.MethodGet, "/api/v1/assets/sber", nil)

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request asset detail endpoint: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, response.StatusCode)
	}
}

func TestImportReviewReturnsTransientRowsWithoutBrokerOperationID(t *testing.T) {
	app := NewDevelopment(verticalslice.NewService(&importAPITestStore{}, fixedHTTPClock{}))
	body := []byte(`{"sourceAccountLabel":"Manual CSV","csvPayload":` + quote(importCSV) + `}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/portfolios/00000000-0000-4000-8000-000000000002/imports/review", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request import review endpoint: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.StatusCode)
	}
	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	text := string(encoded)
	if !strings.Contains(text, `"retentionPolicy":"TRANSIENT_NOT_STORED"`) {
		t.Fatalf("expected transient retention policy, got %s", text)
	}
	if strings.Contains(text, "brokerOperationId") || strings.Contains(text, "broker-row-1") {
		t.Fatalf("broker operation identifier leaked in review response: %s", text)
	}
}

func TestImportReviewRejectsMoreThanOneHundredRows(t *testing.T) {
	app := NewDevelopment(verticalslice.NewService(&importAPITestStore{}, fixedHTTPClock{}))
	body := []byte(`{"csvPayload":` + quote(importCSVWithRows(101)) + `}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/portfolios/00000000-0000-4000-8000-000000000002/imports/review", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request import review endpoint: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.StatusCode)
	}
}

func TestImportReviewRejectsUnknownJSONFields(t *testing.T) {
	store := &importAPITestStore{}
	app := NewDevelopment(verticalslice.NewService(store, fixedHTTPClock{}))
	body := []byte(`{"csvPayload":` + quote(importCSV) + `,"unexpected":true}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/portfolios/00000000-0000-4000-8000-000000000002/imports/review", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request import review endpoint: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.StatusCode)
	}
	if store.listTransactionsCalls != 0 {
		t.Fatalf("expected unknown field to be rejected before store work, got %d calls", store.listTransactionsCalls)
	}
}

func TestImportAppendRequiresIdempotencyKey(t *testing.T) {
	store := &importAPITestStore{}
	app := NewDevelopment(verticalslice.NewService(store, fixedHTTPClock{}))
	body := []byte(`{"sourceAccountLabel":"Manual CSV","csvPayload":` + quote(importCSV) + `,"decisions":[{"rowNumber":2,"action":"APPROVE"}]}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/portfolios/00000000-0000-4000-8000-000000000002/imports/append", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request import append endpoint: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.StatusCode)
	}
	if store.listTransactionsCalls != 0 || store.appendImportedCalls != 0 {
		t.Fatalf("expected missing idempotency key to short-circuit before store work, got list=%d append=%d", store.listTransactionsCalls, store.appendImportedCalls)
	}
}

func TestImportAppendRejectsInvalidIdempotencyKeyBeforeStoreWork(t *testing.T) {
	store := &importAPITestStore{}
	app := NewDevelopment(verticalslice.NewService(store, fixedHTTPClock{}))
	body := []byte(`{"sourceAccountLabel":"Manual CSV","csvPayload":` + quote(importCSV) + `,"decisions":[{"rowNumber":2,"action":"APPROVE"}]}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/portfolios/00000000-0000-4000-8000-000000000002/imports/append", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "too-short")

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request import append endpoint: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.StatusCode)
	}
	if store.listTransactionsCalls != 0 || store.appendImportedCalls != 0 {
		t.Fatalf("expected invalid idempotency key to short-circuit before store work, got list=%d append=%d", store.listTransactionsCalls, store.appendImportedCalls)
	}
}

func TestImportAppendUsesAtomicImportBatch(t *testing.T) {
	store := &importAPITestStore{}
	app := NewDevelopment(verticalslice.NewService(store, fixedHTTPClock{}))
	body := validImportAppendBody(t, app, importCSV)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/portfolios/00000000-0000-4000-8000-000000000002/imports/append", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "import-api-key-0001")

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request import append endpoint: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, response.StatusCode)
	}
	if store.appendedBatch == nil || len(store.appendedBatch.Transactions) != 1 {
		t.Fatalf("expected one approved imported transaction, got %#v", store.appendedBatch)
	}
	if got := store.appendedBatch.SourceKind; got != "USER_UPLOADED_FILE" {
		t.Fatalf("expected import source kind, got %q", got)
	}
}

func TestImportAppendRejectsUnknownJSONFields(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		mutate func(t *testing.T, body []byte) []byte
	}{
		{
			name: "top level",
			mutate: func(t *testing.T, body []byte) []byte {
				request := decodeJSONObject(t, body)
				request["unexpected"] = true
				return encodeJSONObject(t, request)
			},
		},
		{
			name: "decision",
			mutate: func(t *testing.T, body []byte) []byte {
				request := decodeJSONObject(t, body)
				decisions, ok := request["decisions"].([]any)
				if !ok || len(decisions) != 1 {
					t.Fatalf("expected one decision, got %#v", request["decisions"])
				}
				decision, ok := decisions[0].(map[string]any)
				if !ok {
					t.Fatalf("expected object decision, got %#v", decisions[0])
				}
				decision["unexpected"] = true
				return encodeJSONObject(t, request)
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			store := &importAPITestStore{}
			app := NewDevelopment(verticalslice.NewService(store, fixedHTTPClock{}))
			body := testCase.mutate(t, validImportAppendBody(t, app, importCSV))
			listCallsAfterReview := store.listTransactionsCalls
			request := httptest.NewRequest(http.MethodPost, "/api/v1/portfolios/00000000-0000-4000-8000-000000000002/imports/append", bytes.NewReader(body))
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Idempotency-Key", "import-api-key-unknown-field")

			response, err := app.Test(request)
			if err != nil {
				t.Fatalf("request import append endpoint: %v", err)
			}
			defer response.Body.Close()
			if response.StatusCode != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.StatusCode)
			}
			if store.listTransactionsCalls != listCallsAfterReview || store.appendImportedCalls != 0 {
				t.Fatalf("expected unknown field to be rejected before append work, got list=%d append=%d", store.listTransactionsCalls, store.appendImportedCalls)
			}
		})
	}
}

func TestImportAppendRetryReachesIdempotentAppenderAfterLedgerChanges(t *testing.T) {
	store := &importAPITestStore{}
	app := NewDevelopment(verticalslice.NewService(store, fixedHTTPClock{}))
	body := validImportAppendBody(t, app, importCSV)
	for attempt := 0; attempt < 2; attempt++ {
		request := httptest.NewRequest(http.MethodPost, "/api/v1/portfolios/00000000-0000-4000-8000-000000000002/imports/append", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Idempotency-Key", "import-api-key-replay-0001")

		response, err := app.Test(request)
		if err != nil {
			t.Fatalf("request import append attempt %d: %v", attempt+1, err)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusCreated {
			t.Fatalf("expected retry attempt %d to reach the idempotent appender, got %d", attempt+1, response.StatusCode)
		}
	}
	if store.appendImportedCalls != 2 {
		t.Fatalf("expected both attempts to reach the appender for idempotent replay, got %d", store.appendImportedCalls)
	}
}

func TestImportAppendRevalidatesAgainstCurrentLedger(t *testing.T) {
	quantity := decimal.Must("2.00000000")
	unitPrice := verticalslice.Money{Amount: decimal.Must("100.00000000"), Currency: verticalslice.RUB}
	ticker := "SBER"
	settlementDate := "2026-01-13"

	// Review first against a clean ledger so the row is legitimately APPENDABLE
	// in the signed review. Then simulate a concurrent writer creating the same
	// financial identity before append reaches the locked store.
	store := &importAPITestStore{}
	app := NewDevelopment(verticalslice.NewService(store, fixedHTTPClock{}))
	body := validImportAppendBody(t, app, importCSV)

	store.appendImportedError = verticalslice.ErrInvalidInput
	store.existingTransactions = []verticalslice.Transaction{{
		ID:              "00000000-0000-4000-8000-000000000201",
		PortfolioID:     "00000000-0000-4000-8000-000000000002",
		TransactionType: "BUY",
		Status:          "ACTIVE",
		Ticker:          &ticker,
		Quantity:        &quantity,
		UnitPrice:       &unitPrice,
		GrossAmount:     verticalslice.Money{Amount: decimal.Must("200.00000000"), Currency: verticalslice.RUB},
		Commission:      verticalslice.Money{Amount: decimal.Must("1.00000000"), Currency: verticalslice.RUB},
		Tax:             verticalslice.ZeroMoney(),
		TradeDate:       "2026-01-10",
		SettlementDate:  &settlementDate,
		Revision:        1,
	}}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/portfolios/00000000-0000-4000-8000-000000000002/imports/append", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "import-api-key-0002")

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request import append endpoint: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected stale duplicate to be rejected with status %d, got %d", http.StatusBadRequest, response.StatusCode)
	}
	if store.appendImportedCalls != 1 {
		t.Fatalf("expected stale duplicate to be rejected by the atomic appender, got append calls=%d", store.appendImportedCalls)
	}
}

func TestImportAppendRejectsPayloadChangedAfterReview(t *testing.T) {
	store := &importAPITestStore{}
	app := NewDevelopment(verticalslice.NewService(store, fixedHTTPClock{}))
	review := mustReviewImportCSV(t, app, importCSV)
	listCallsAfterReview := store.listTransactionsCalls
	reviewHistoryCallsAfterReview := store.listImportReviewCalls
	changedCSV := "transaction_type,ticker,quantity,unit_price,gross_amount,commission,tax,trade_date,settlement_date,currency,broker_operation_id,note\n" +
		"BUY,SBER,2.00000000,999.00000000,1998.00000000,1.00000000,0.00000000,2026-01-10,2026-01-13,RUB,broker-row-1,Changed buy\n"
	body := []byte(`{"sourceAccountLabel":"Manual CSV","sourceFileHash":"` + review.SourceFileHash + `","reviewToken":"` + review.ReviewToken + `","csvPayload":` + quote(changedCSV) + `,"decisions":[{"rowNumber":2,"rowHash":"` + review.Rows[0].RowHash + `","action":"APPROVE"}]}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/portfolios/00000000-0000-4000-8000-000000000002/imports/append", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "import-api-key-0003")

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request import append endpoint: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected changed payload to be rejected with status %d, got %d", http.StatusBadRequest, response.StatusCode)
	}
	if store.listTransactionsCalls != listCallsAfterReview ||
		store.listImportReviewCalls != reviewHistoryCallsAfterReview ||
		store.appendImportedCalls != 0 {
		t.Fatalf(
			"expected changed payload to be rejected before store work, got list=%d reviewHistory=%d append=%d",
			store.listTransactionsCalls,
			store.listImportReviewCalls,
			store.appendImportedCalls,
		)
	}
}

func quote(value string) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return string(encoded)
}

func decodeJSONObject(t *testing.T, body []byte) map[string]any {
	t.Helper()
	var request map[string]any
	if err := json.Unmarshal(body, &request); err != nil {
		t.Fatalf("decode JSON object: %v", err)
	}
	return request
}

func encodeJSONObject(t *testing.T, request map[string]any) []byte {
	t.Helper()
	body, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("encode JSON object: %v", err)
	}
	return body
}

func validImportAppendBody(t *testing.T, app *fiber.App, payload string) []byte {
	t.Helper()
	review := mustReviewImportCSV(t, app, payload)
	return []byte(`{"sourceAccountLabel":"Manual CSV","sourceFileHash":"` + review.SourceFileHash + `","reviewToken":"` + review.ReviewToken + `","csvPayload":` + quote(payload) + `,"decisions":[{"rowNumber":2,"rowHash":"` + review.Rows[0].RowHash + `","action":"APPROVE"}]}`)
}

type importReviewTestResponse struct {
	Data importReviewTestData `json:"data"`
}

type importReviewTestData struct {
	SourceFileHash string                `json:"sourceFileHash"`
	ReviewToken    string                `json:"reviewToken"`
	Rows           []importReviewTestRow `json:"rows"`
}

type importReviewTestRow struct {
	RowNumber int    `json:"rowNumber"`
	RowHash   string `json:"rowHash"`
}

func mustReviewImportCSV(t *testing.T, app *fiber.App, payload string) importReviewTestData {
	t.Helper()
	body := []byte(`{"sourceAccountLabel":"Manual CSV","csvPayload":` + quote(payload) + `}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/portfolios/00000000-0000-4000-8000-000000000002/imports/review", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request import review endpoint: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected review status %d, got %d", http.StatusOK, response.StatusCode)
	}
	var decoded importReviewTestResponse
	if err := json.NewDecoder(response.Body).Decode(&decoded); err != nil {
		t.Fatalf("decode review response: %v", err)
	}
	if decoded.Data.SourceFileHash == "" || decoded.Data.ReviewToken == "" || len(decoded.Data.Rows) == 0 {
		t.Fatalf("review response missing import identity: %+v", decoded.Data)
	}
	return decoded.Data
}

func importCSVWithRows(rows int) string {
	var builder strings.Builder
	builder.WriteString("transaction_type,ticker,quantity,unit_price,gross_amount,commission,tax,trade_date,settlement_date,currency,broker_operation_id,note\n")
	for row := 0; row < rows; row++ {
		builder.WriteString("BUY,SBER,2.00000000,100.00000000,200.00000000,1.00000000,0.00000000,2026-01-10,2026-01-13,RUB,broker-row-")
		builder.WriteString(string(rune('a' + row%26)))
		builder.WriteString(",Imported buy\n")
	}
	return builder.String()
}

type fixedHTTPClock struct{}

func (fixedHTTPClock) Now() time.Time {
	return time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC)
}

type importAPITestStore struct {
	appendedBatch         *verticalslice.AppendImportBatchRequest
	existingTransactions  []verticalslice.Transaction
	portfolios            []verticalslice.Portfolio
	listTransactionsCalls int
	listImportReviewCalls int
	importReviewFilter    verticalslice.ImportReviewHistoryFilter
	appendImportedCalls   int
	appendImportedError   error
}

func (store *importAPITestStore) Ping(context.Context) error {
	return nil
}

func (store *importAPITestStore) SearchAssets(_ context.Context, filter verticalslice.AssetSearchFilter) ([]verticalslice.AssetSummary, error) {
	assets := []verticalslice.AssetSummary{
		{
			Ticker:    "SBER",
			Name:      "Sberbank ordinary shares",
			AssetType: "STOCK",
			Currency:  verticalslice.RUB,
			LotSize:   decimal.Must("10.00000000"),
			LastPrice: nil,
		},
		{
			Ticker:    "SU26238RMFS4",
			Name:      "OFZ 26238",
			AssetType: "BOND",
			Currency:  verticalslice.RUB,
			LotSize:   decimal.Must("1.00000000"),
			LastPrice: nil,
		},
	}
	start := 0
	if filter.AfterTicker != "" {
		for index, asset := range assets {
			if asset.Ticker > filter.AfterTicker {
				start = index
				break
			}
			start = index + 1
		}
	}
	end := start + filter.Limit
	if end > len(assets) {
		end = len(assets)
	}
	return assets[start:end], nil
}

func (store *importAPITestStore) ListPortfolios(_ context.Context, _ string, filter verticalslice.PortfolioFilter) ([]verticalslice.Portfolio, error) {
	items := store.portfolios
	if filter.BeforeUpdatedAt != nil {
		start := len(items)
		for index, item := range items {
			if item.UpdatedAt.Before(*filter.BeforeUpdatedAt) || (item.UpdatedAt.Equal(*filter.BeforeUpdatedAt) && item.ID < filter.BeforeID) {
				start = index
				break
			}
		}
		items = items[start:]
	}
	if filter.Limit < len(items) {
		items = items[:filter.Limit]
	}
	return items, nil
}

func (store *importAPITestStore) CreatePortfolio(context.Context, verticalslice.CommandContext, verticalslice.CreatePortfolioRequest) (verticalslice.Portfolio, error) {
	return verticalslice.Portfolio{}, nil
}

func (store *importAPITestStore) GetPortfolio(context.Context, string, string) (verticalslice.Portfolio, error) {
	return verticalslice.Portfolio{}, nil
}

func (store *importAPITestStore) ListTransactions(_ context.Context, _ string, _ string, filter verticalslice.TransactionFilter) ([]verticalslice.Transaction, error) {
	store.listTransactionsCalls++
	items := store.existingTransactions
	if filter.BeforeTradeDate != "" {
		start := len(items)
		for index, item := range items {
			if item.TradeDate < filter.BeforeTradeDate || (item.TradeDate == filter.BeforeTradeDate && item.EntryID < filter.BeforeEntryID) {
				start = index
				break
			}
		}
		items = items[start:]
	}
	if filter.Limit < len(items) {
		items = items[:filter.Limit]
	}
	return items, nil
}

func (store *importAPITestStore) ListImportReviewTransactions(_ context.Context, _ string, _ string, filter verticalslice.ImportReviewHistoryFilter) ([]verticalslice.Transaction, error) {
	store.listImportReviewCalls++
	store.importReviewFilter = filter
	return store.existingTransactions, nil
}

func (store *importAPITestStore) AppendTransaction(context.Context, verticalslice.CommandContext, verticalslice.AppendTransactionRequest) (verticalslice.Transaction, error) {
	return verticalslice.Transaction{}, nil
}

func (store *importAPITestStore) AppendImportedTransactions(_ context.Context, _ verticalslice.CommandContext, request verticalslice.AppendImportBatchRequest) ([]verticalslice.Transaction, error) {
	store.appendImportedCalls++
	store.appendedBatch = &request
	if store.appendImportedError != nil {
		return nil, store.appendImportedError
	}
	quantity, _ := decimal.FromString("2.00000000")
	unitPriceAmount, _ := decimal.FromString("100.00000000")
	unitPrice := verticalslice.Money{Amount: unitPriceAmount, Currency: verticalslice.RUB}
	transaction := verticalslice.Transaction{
		ID:              "00000000-0000-4000-8000-000000000101",
		EntryID:         "00000000-0000-4000-8000-000000000102",
		PortfolioID:     request.PortfolioID,
		TransactionType: "BUY",
		Status:          "ACTIVE",
		Ticker:          request.Transactions[0].Ticker,
		Quantity:        &quantity,
		UnitPrice:       &unitPrice,
		GrossAmount:     *request.Transactions[0].GrossAmount,
		Commission:      request.Transactions[0].Commission,
		Tax:             request.Transactions[0].Tax,
		TradeDate:       request.Transactions[0].TradeDate,
		SettlementDate:  request.Transactions[0].SettlementDate,
		Revision:        1,
		CreatedAt:       time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC),
		UpdatedAt:       time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC),
	}
	store.existingTransactions = append(store.existingTransactions, transaction)
	return []verticalslice.Transaction{transaction}, nil
}

func (store *importAPITestStore) GetPortfolioSummary(context.Context, string, string, string) (verticalslice.PortfolioSummary, error) {
	return verticalslice.PortfolioSummary{}, nil
}
