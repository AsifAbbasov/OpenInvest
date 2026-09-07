package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

type stage374HTTPStore struct {
	importAPITestStore
	correctionCalls   int
	reversalCalls     int
	correctionCommand verticalslice.CommandContext
	reversalCommand   verticalslice.CommandContext
	correctionRequest verticalslice.CorrectTransactionRequest
	reversalRequest   verticalslice.ReverseTransactionRequest
	correctionErr     error
	reversalErr       error
}

func (store *stage374HTTPStore) CorrectTransactionWithReplayStage374(
	_ context.Context,
	command verticalslice.CommandContext,
	request verticalslice.CorrectTransactionRequest,
	build verticalslice.TransactionReplayBuilder,
) (verticalslice.Transaction, verticalslice.CommandReplayArtifact, error) {
	store.correctionCalls++
	store.correctionCommand = command
	store.correctionRequest = request
	if store.correctionErr != nil {
		return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, store.correctionErr
	}
	gross, err := verticalslice.GrossFor(request.Corrected)
	if err != nil {
		return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err
	}
	transaction := verticalslice.Transaction{
		ID:              request.TransactionID,
		EntryID:         "00000000-0000-4000-8000-000000000074",
		PortfolioID:     request.PortfolioID,
		TransactionType: request.Corrected.TransactionType,
		Status:          "CORRECTED",
		Ticker:          request.Corrected.Ticker,
		Quantity:        request.Corrected.Quantity,
		UnitPrice:       request.Corrected.UnitPrice,
		GrossAmount:     gross,
		Commission:      request.Corrected.Commission,
		Tax:             request.Corrected.Tax,
		TradeDate:       request.Corrected.TradeDate,
		SettlementDate:  request.Corrected.SettlementDate,
		Note:            request.Corrected.Note,
		Revision:        request.ExpectedRevision + 1,
		CreatedAt:       command.Now,
		UpdatedAt:       command.Now,
	}
	artifact, err := build(transaction)
	if err != nil {
		return verticalslice.Transaction{}, verticalslice.CommandReplayArtifact{}, err
	}
	return transaction, artifact, nil
}

func (store *stage374HTTPStore) ReverseTransactionWithReplayStage374(
	_ context.Context,
	command verticalslice.CommandContext,
	request verticalslice.ReverseTransactionRequest,
	build verticalslice.TransactionReversalReplayBuilder,
) (verticalslice.TransactionReversal, verticalslice.CommandReplayArtifact, error) {
	store.reversalCalls++
	store.reversalCommand = command
	store.reversalRequest = request
	if store.reversalErr != nil {
		return verticalslice.TransactionReversal{}, verticalslice.CommandReplayArtifact{}, store.reversalErr
	}
	result := verticalslice.TransactionReversal{
		TransactionID:         request.TransactionID,
		ReversalTransactionID: "00000000-0000-4000-8000-000000000075",
		Status:                "REVERSED",
		EffectiveDate:         request.EffectiveDate,
	}
	artifact, err := build(result)
	if err != nil {
		return verticalslice.TransactionReversal{}, verticalslice.CommandReplayArtifact{}, err
	}
	return result, artifact, nil
}

func TestStage374CorrectionHTTPUsesFrozenPATCHContract(t *testing.T) {
	store := &stage374HTTPStore{}
	app := NewDevelopment(verticalslice.NewService(store, fixedHTTPClock{}))
	body := []byte(`{
		"expectedRevision":1,
		"reason":"Incorrect broker price",
		"corrected":{
			"transactionType":"BUY",
			"ticker":"SBER",
			"quantity":"10.00000000",
			"unitPrice":{"amount":"125.00000000","currency":"RUB"},
			"grossAmount":null,
			"commission":{"amount":"1.00000000","currency":"RUB"},
			"tax":{"amount":"0.00000000","currency":"RUB"},
			"tradeDate":"2026-03-01",
			"settlementDate":null,
			"note":"corrected"
		}
	}`)
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/portfolios/00000000-0000-4000-8000-000000000002/transactions/00000000-0000-4000-8000-000000000071", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "stage-3-74-correction-http")
	request.Header.Set("X-Request-ID", "00000000-0000-4000-8000-000000000076")

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request correction endpoint: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("correction status=%d want=%d", response.StatusCode, http.StatusOK)
	}
	if store.correctionCalls != 1 || store.reversalCalls != 0 {
		t.Fatalf("unexpected repair store calls: correction=%d reversal=%d", store.correctionCalls, store.reversalCalls)
	}
	if store.correctionRequest.ExpectedRevision != 1 || store.correctionRequest.Reason != "Incorrect broker price" {
		t.Fatalf("correction metadata drifted: %+v", store.correctionRequest)
	}
	if store.correctionRequest.Corrected.Quantity == nil || store.correctionRequest.Corrected.Quantity.String() != "10.00000000" || store.correctionRequest.Corrected.UnitPrice == nil || store.correctionRequest.Corrected.UnitPrice.Amount.String() != "125.00000000" {
		t.Fatalf("correction financial payload drifted: %+v", store.correctionRequest.Corrected)
	}
	if store.correctionCommand.IdempotencyKey != "stage-3-74-correction-http" || store.correctionCommand.RequestPath != "/api/v1/portfolios/00000000-0000-4000-8000-000000000002/transactions/00000000-0000-4000-8000-000000000071" {
		t.Fatalf("correction command scope drifted: %+v", store.correctionCommand)
	}
	if response.Header.Get("X-Request-ID") != "00000000-0000-4000-8000-000000000076" {
		t.Fatalf("correction replay response did not preserve request id")
	}
	var payload struct {
		Data transactionDTO `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode correction response: %v", err)
	}
	if payload.Data.ID != "00000000-0000-4000-8000-000000000071" || payload.Data.Status != "CORRECTED" || payload.Data.Revision != 2 {
		t.Fatalf("correction response drifted: %+v", payload.Data)
	}
}

func TestStage374ReversalHTTPUsesFrozenDELETECommandContract(t *testing.T) {
	store := &stage374HTTPStore{}
	app := NewDevelopment(verticalslice.NewService(store, fixedHTTPClock{}))
	body := []byte(`{"expectedRevision":2,"reason":"Broker cancelled operation","effectiveDate":"2026-04-15"}`)
	request := httptest.NewRequest(http.MethodDelete, "/api/v1/portfolios/00000000-0000-4000-8000-000000000002/transactions/00000000-0000-4000-8000-000000000071", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "stage-3-74-reversal-http")

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request reversal endpoint: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("reversal status=%d want=%d", response.StatusCode, http.StatusOK)
	}
	if store.reversalCalls != 1 || store.correctionCalls != 0 {
		t.Fatalf("unexpected repair store calls: correction=%d reversal=%d", store.correctionCalls, store.reversalCalls)
	}
	if store.reversalRequest.ExpectedRevision != 2 || store.reversalRequest.Reason != "Broker cancelled operation" || store.reversalRequest.EffectiveDate != "2026-04-15" {
		t.Fatalf("reversal request drifted: %+v", store.reversalRequest)
	}
	if store.reversalCommand.IdempotencyKey != "stage-3-74-reversal-http" || store.reversalCommand.RequestPath == "" {
		t.Fatalf("reversal command scope drifted: %+v", store.reversalCommand)
	}
	var payload struct {
		Data transactionReversalDTO `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode reversal response: %v", err)
	}
	if payload.Data.TransactionID != "00000000-0000-4000-8000-000000000071" || payload.Data.Status != "REVERSED" || payload.Data.EffectiveDate != "2026-04-15" {
		t.Fatalf("reversal response drifted: %+v", payload.Data)
	}
}

func TestStage374HTTPMapsRevisionConflictAndRejectsMissingCorrectionSettlementField(t *testing.T) {
	conflictStore := &stage374HTTPStore{correctionErr: verticalslice.ErrTransactionConflict}
	conflictApp := NewDevelopment(verticalslice.NewService(conflictStore, fixedHTTPClock{}))
	validBody := []byte(`{
		"expectedRevision":1,"reason":"conflict witness","corrected":{
			"transactionType":"BUY","ticker":"SBER","quantity":"1.00000000",
			"unitPrice":{"amount":"100.00000000","currency":"RUB"},"grossAmount":null,
			"commission":{"amount":"0.00000000","currency":"RUB"},"tax":{"amount":"0.00000000","currency":"RUB"},
			"tradeDate":"2026-03-01","settlementDate":null,"note":null
		}
	}`)
	conflictRequest := httptest.NewRequest(http.MethodPatch, "/api/v1/portfolios/00000000-0000-4000-8000-000000000002/transactions/00000000-0000-4000-8000-000000000071", bytes.NewReader(validBody))
	conflictRequest.Header.Set("Content-Type", "application/json")
	conflictRequest.Header.Set("Idempotency-Key", "stage-3-74-conflict-http")
	conflictResponse, err := conflictApp.Test(conflictRequest)
	if err != nil {
		t.Fatalf("request conflict witness: %v", err)
	}
	defer conflictResponse.Body.Close()
	if conflictResponse.StatusCode != http.StatusConflict {
		t.Fatalf("revision conflict status=%d want=%d", conflictResponse.StatusCode, http.StatusConflict)
	}
	var conflictPayload errorResponse
	if err := json.NewDecoder(conflictResponse.Body).Decode(&conflictPayload); err != nil {
		t.Fatalf("decode conflict response: %v", err)
	}
	if conflictPayload.Error.Code != "CONFLICT" {
		t.Fatalf("revision conflict error code=%q", conflictPayload.Error.Code)
	}

	validationStore := &stage374HTTPStore{}
	validationApp := NewDevelopment(verticalslice.NewService(validationStore, fixedHTTPClock{}))
	missingSettlement := []byte(`{
		"expectedRevision":1,"reason":"missing field witness","corrected":{
			"transactionType":"BUY","ticker":"SBER","quantity":"1.00000000",
			"unitPrice":{"amount":"100.00000000","currency":"RUB"},"grossAmount":null,
			"commission":{"amount":"0.00000000","currency":"RUB"},"tax":{"amount":"0.00000000","currency":"RUB"},
			"tradeDate":"2026-03-01","note":null
		}
	}`)
	validationRequest := httptest.NewRequest(http.MethodPatch, "/api/v1/portfolios/00000000-0000-4000-8000-000000000002/transactions/00000000-0000-4000-8000-000000000071", bytes.NewReader(missingSettlement))
	validationRequest.Header.Set("Content-Type", "application/json")
	validationRequest.Header.Set("Idempotency-Key", "stage-3-74-validation-http")
	validationResponse, err := validationApp.Test(validationRequest)
	if err != nil {
		t.Fatalf("request validation witness: %v", err)
	}
	defer validationResponse.Body.Close()
	if validationResponse.StatusCode != http.StatusBadRequest || validationStore.correctionCalls != 0 {
		t.Fatalf("missing corrected.settlementDate must fail before store: status=%d calls=%d", validationResponse.StatusCode, validationStore.correctionCalls)
	}
}
