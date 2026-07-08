package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

const importCSV = "transaction_type,ticker,quantity,unit_price,gross_amount,commission,tax,trade_date,settlement_date,currency,broker_operation_id,note\n" +
	"BUY,SBER,2.00000000,100.00000000,200.00000000,1.00000000,0.00000000,2026-01-10,2026-01-13,RUB,broker-row-1,Imported buy\n"

func TestImportReviewReturnsTransientRowsWithoutBrokerOperationID(t *testing.T) {
	app := New(verticalslice.NewService(&importAPITestStore{}, fixedHTTPClock{}))
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
	app := New(verticalslice.NewService(&importAPITestStore{}, fixedHTTPClock{}))
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

func TestImportAppendRequiresIdempotencyKey(t *testing.T) {
	store := &importAPITestStore{}
	app := New(verticalslice.NewService(store, fixedHTTPClock{}))
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
	app := New(verticalslice.NewService(store, fixedHTTPClock{}))
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
	app := New(verticalslice.NewService(store, fixedHTTPClock{}))
	body := []byte(`{"sourceAccountLabel":"Manual CSV","csvPayload":` + quote(importCSV) + `,"decisions":[{"rowNumber":2,"action":"APPROVE"}]}`)
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

func TestImportAppendRevalidatesAgainstCurrentLedger(t *testing.T) {
	quantity := decimal.Must("2.00000000")
	unitPrice := verticalslice.Money{Amount: decimal.Must("100.00000000"), Currency: verticalslice.RUB}
	ticker := "SBER"
	settlementDate := "2026-01-13"
	store := &importAPITestStore{
		existingTransactions: []verticalslice.Transaction{{
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
		}},
	}
	app := New(verticalslice.NewService(store, fixedHTTPClock{}))
	body := []byte(`{"sourceAccountLabel":"Manual CSV","csvPayload":` + quote(importCSV) + `,"decisions":[{"rowNumber":2,"action":"APPROVE"}]}`)
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
	if store.appendImportedCalls != 0 {
		t.Fatalf("expected stale duplicate to be rejected before append, got append calls=%d", store.appendImportedCalls)
	}
}

func quote(value string) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return string(encoded)
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
	listTransactionsCalls int
	appendImportedCalls   int
}

func (store *importAPITestStore) Ping(context.Context) error {
	return nil
}

func (store *importAPITestStore) ListPortfolios(context.Context, string, int) ([]verticalslice.Portfolio, error) {
	return []verticalslice.Portfolio{}, nil
}

func (store *importAPITestStore) CreatePortfolio(context.Context, verticalslice.CommandContext, verticalslice.CreatePortfolioRequest) (verticalslice.Portfolio, error) {
	return verticalslice.Portfolio{}, nil
}

func (store *importAPITestStore) GetPortfolio(context.Context, string, string) (verticalslice.Portfolio, error) {
	return verticalslice.Portfolio{}, nil
}

func (store *importAPITestStore) ListTransactions(context.Context, string, string, verticalslice.TransactionFilter) ([]verticalslice.Transaction, error) {
	store.listTransactionsCalls++
	return store.existingTransactions, nil
}

func (store *importAPITestStore) AppendTransaction(context.Context, verticalslice.CommandContext, verticalslice.AppendTransactionRequest) (verticalslice.Transaction, error) {
	return verticalslice.Transaction{}, nil
}

func (store *importAPITestStore) AppendImportedTransactions(_ context.Context, _ verticalslice.CommandContext, request verticalslice.AppendImportBatchRequest) ([]verticalslice.Transaction, error) {
	store.appendImportedCalls++
	store.appendedBatch = &request
	quantity, _ := decimal.FromString("2.00000000")
	unitPriceAmount, _ := decimal.FromString("100.00000000")
	unitPrice := verticalslice.Money{Amount: unitPriceAmount, Currency: verticalslice.RUB}
	transaction := verticalslice.Transaction{
		ID:              "00000000-0000-4000-8000-000000000101",
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
	return []verticalslice.Transaction{transaction}, nil
}

func (store *importAPITestStore) GetPortfolioSummary(context.Context, string, string, string) (verticalslice.PortfolioSummary, error) {
	return verticalslice.PortfolioSummary{}, nil
}
