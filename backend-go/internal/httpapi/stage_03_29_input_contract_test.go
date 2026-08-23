package httpapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

type stage329Store struct {
	importAPITestStore
	createPortfolioCalls   int
	appendTransactionCalls int
}

func (store *stage329Store) CreatePortfolio(_ context.Context, _ verticalslice.CommandContext, _ verticalslice.CreatePortfolioRequest) (verticalslice.Portfolio, error) {
	store.createPortfolioCalls++
	return verticalslice.Portfolio{}, nil
}

func (store *stage329Store) AppendTransaction(_ context.Context, _ verticalslice.CommandContext, _ verticalslice.AppendTransactionRequest) (verticalslice.Transaction, error) {
	store.appendTransactionCalls++
	return verticalslice.Transaction{}, nil
}

func TestStage329StrictJSONRejectsUnknownPortfolioField(t *testing.T) {
	store := &stage329Store{}
	app := NewDevelopment(verticalslice.NewService(store, fixedHTTPClock{}))
	request := httptest.NewRequest(http.MethodPost, "/api/v1/portfolios",
		bytes.NewBufferString(`{"name":"Primary","baseCurrency":"RUB","unexpected":true}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "stage329-test-key-0001")

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request portfolio create: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for unknown portfolio field, got %d", response.StatusCode)
	}
	if store.createPortfolioCalls != 0 {
		t.Fatalf("expected strict JSON rejection before store, got %d create calls", store.createPortfolioCalls)
	}
}

func TestStage329StrictJSONRejectsUnknownTransactionField(t *testing.T) {
	store := &stage329Store{}
	app := NewDevelopment(verticalslice.NewService(store, fixedHTTPClock{}))
	body := `{"transactionType":"DEPOSIT","ticker":null,"quantity":null,"unitPrice":null,` +
		`"grossAmount":{"amount":"100.00000000","currency":"RUB"},` +
		`"commission":{"amount":"0.00000000","currency":"RUB"},` +
		`"tax":{"amount":"0.00000000","currency":"RUB"},` +
		`"tradeDate":"2026-01-10","settlementDate":null,"note":null,"unexpected":true}`
	request := httptest.NewRequest(http.MethodPost,
		"/api/v1/portfolios/00000000-0000-4000-8000-000000000002/transactions",
		bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "stage329-test-key-0002")

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request transaction append: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for unknown transaction field, got %d", response.StatusCode)
	}
	if store.appendTransactionCalls != 0 {
		t.Fatalf("expected strict JSON rejection before store, got %d append calls", store.appendTransactionCalls)
	}
}

func TestStage329InvalidAndOversizedDecimalsReturn400BeforeStore(t *testing.T) {
	for _, testCase := range []struct {
		name     string
		quantity string
	}{
		{name: "malformed", quantity: "abc"},
		{name: "outside NUMERIC(28,8)", quantity: "100000000000000000000.00000000"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			store := &stage329Store{}
			app := NewDevelopment(verticalslice.NewService(store, fixedHTTPClock{}))
			body := `{"transactionType":"BUY","ticker":"SBER","quantity":"` + testCase.quantity + `",` +
				`"unitPrice":{"amount":"100.00000000","currency":"RUB"},` +
				`"grossAmount":{"amount":"200.00000000","currency":"RUB"},` +
				`"commission":{"amount":"0.00000000","currency":"RUB"},` +
				`"tax":{"amount":"0.00000000","currency":"RUB"},` +
				`"tradeDate":"2026-01-10","settlementDate":null,"note":null}`
			request := httptest.NewRequest(http.MethodPost,
				"/api/v1/portfolios/00000000-0000-4000-8000-000000000002/transactions",
				bytes.NewBufferString(body))
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Idempotency-Key", "stage329-test-key-0003")

			response, err := app.Test(request)
			if err != nil {
				t.Fatalf("request transaction append: %v", err)
			}
			defer response.Body.Close()
			if response.StatusCode != http.StatusBadRequest {
				t.Fatalf("expected 400 for %s decimal, got %d", testCase.name, response.StatusCode)
			}
			if store.appendTransactionCalls != 0 {
				t.Fatalf("expected decimal rejection before store, got %d append calls", store.appendTransactionCalls)
			}
		})
	}
}

func TestStage329OverlongNoteReturns400BeforeStore(t *testing.T) {
	store := &stage329Store{}
	app := NewDevelopment(verticalslice.NewService(store, fixedHTTPClock{}))
	body := `{"transactionType":"DEPOSIT","ticker":null,"quantity":null,"unitPrice":null,` +
		`"grossAmount":{"amount":"100.00000000","currency":"RUB"},` +
		`"commission":{"amount":"0.00000000","currency":"RUB"},` +
		`"tax":{"amount":"0.00000000","currency":"RUB"},` +
		`"tradeDate":"2026-01-10","settlementDate":null,"note":` + quote(strings.Repeat("Ж", 501)) + `}`
	request := httptest.NewRequest(http.MethodPost,
		"/api/v1/portfolios/00000000-0000-4000-8000-000000000002/transactions",
		bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "stage329-test-key-0004")

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request transaction append: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for overlong note, got %d", response.StatusCode)
	}
	if store.appendTransactionCalls != 0 {
		t.Fatalf("expected note rejection before store, got %d append calls", store.appendTransactionCalls)
	}
}

func TestStage329DuplicateCSVHeaderReturns400(t *testing.T) {
	store := &stage329Store{}
	app := NewDevelopment(verticalslice.NewService(store, fixedHTTPClock{}))
	payload := "transaction_type,ticker, TICKER ,quantity,unit_price,gross_amount,commission,tax,trade_date,settlement_date,currency,broker_operation_id,note\n" +
		"BUY,SBER,SBER,2.00000000,100.00000000,200.00000000,0.00000000,0.00000000,2026-01-10,2026-01-13,RUB,broker-1,note\n"
	body := `{"sourceAccountLabel":"Manual CSV","csvPayload":` + quote(payload) + `}`
	request := httptest.NewRequest(http.MethodPost,
		"/api/v1/portfolios/00000000-0000-4000-8000-000000000002/imports/review",
		bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request import review: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for duplicate CSV header, got %d", response.StatusCode)
	}
	if store.listTransactionsCalls != 1 {
		t.Fatalf("expected one existing-transaction read before CSV validation, got %d", store.listTransactionsCalls)
	}
}
