package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/importer"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func stage330TokenReview(t *testing.T) importer.Review {
	t.Helper()
	review, err := importer.ReviewCSV(importer.ReviewRequest{
		SubjectID:          "subject-a",
		PortfolioID:        "portfolio-a",
		SourceKind:         importer.SourceKindUserUploadedFile,
		SourceAccountLabel: "Broker A",
		FileHash:           strings.Repeat("b", 64),
		Reader: strings.NewReader("transaction_type,ticker,quantity,unit_price,gross_amount,commission,tax,trade_date,settlement_date,currency,broker_operation_id,note\n" +
			"DEPOSIT,,,,100.00000000,0.00000000,0.00000000,2026-08-23,,RUB,broker-1,note\n"),
	})
	if err != nil {
		t.Fatalf("build token review: %v", err)
	}
	return review
}

func TestStage330ReviewTokenExpiresAndRejectsNormalizedSemanticDrift(t *testing.T) {
	secret, err := normalizedImportReviewSecret([]byte("stage-3.30-review-token-secret-32-bytes"))
	if err != nil {
		t.Fatalf("normalize secret: %v", err)
	}
	now := time.Date(2026, 8, 23, 8, 0, 0, 0, time.UTC)
	api := &API{importReviewSecret: secret, now: func() time.Time { return now }}
	parserReview := stage330TokenReview(t)
	finalReview := parserReview
	token, err := api.signImportReviewToken("subject-a", parserReview, finalReview)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	decision := []importer.Decision{{
		RowNumber: parserReview.Rows[0].RowNumber,
		RowHash:   parserReview.Rows[0].RowHash,
		Action:    importer.DecisionApprove,
	}}

	if err := api.verifyImportReviewToken(token, "subject-a", "portfolio-a", importer.SourceKindUserUploadedFile, "Broker A", parserReview.FileHash, parserReview, decision); err != nil {
		t.Fatalf("expected fresh token to verify: %v", err)
	}

	api.now = func() time.Time { return now.Add(importReviewTokenTTL) }
	if err := api.verifyImportReviewToken(token, "subject-a", "portfolio-a", importer.SourceKindUserUploadedFile, "Broker A", parserReview.FileHash, parserReview, decision); !errors.Is(err, importer.ErrUnsafeAppend) {
		t.Fatalf("expected expired token rejection, got %v", err)
	}

	api.now = func() time.Time { return now }
	mutated := parserReview
	mutated.Rows = append([]importer.RowReview(nil), parserReview.Rows...)
	row := mutated.Rows[0]
	candidate := *row.Candidate
	note := "changed normalized note"
	candidate.SafeNote = &note
	row.Candidate = &candidate
	mutated.Rows[0] = row
	if err := api.verifyImportReviewToken(token, "subject-a", "portfolio-a", importer.SourceKindUserUploadedFile, "Broker A", parserReview.FileHash, mutated, decision); !errors.Is(err, importer.ErrUnsafeAppend) {
		t.Fatalf("expected normalized semantic drift rejection, got %v", err)
	}
}

func TestStage330ReviewTokenPreventsApprovingPreviouslyNonAppendableRow(t *testing.T) {
	secret, err := normalizedImportReviewSecret([]byte("stage-3.30-review-token-secret-32-bytes"))
	if err != nil {
		t.Fatalf("normalize secret: %v", err)
	}
	now := time.Date(2026, 8, 23, 8, 0, 0, 0, time.UTC)
	api := &API{importReviewSecret: secret, now: func() time.Time { return now }}
	parserReview := stage330TokenReview(t)
	finalReview := parserReview
	finalReview.Rows = append([]importer.RowReview(nil), parserReview.Rows...)
	finalReview.Rows[0] = parserReview.Rows[0]
	finalReview.Rows[0].Status = importer.ReviewStatusDuplicate
	finalReview.Summary = importer.Summary{TotalRows: 1, DuplicateRows: 1}

	token, err := api.signImportReviewToken("subject-a", parserReview, finalReview)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	decision := []importer.Decision{{
		RowNumber: parserReview.Rows[0].RowNumber,
		RowHash:   parserReview.Rows[0].RowHash,
		Action:    importer.DecisionApprove,
	}}
	if err := api.verifyImportReviewToken(token, "subject-a", "portfolio-a", importer.SourceKindUserUploadedFile, "Broker A", parserReview.FileHash, parserReview, decision); !errors.Is(err, importer.ErrUnsafeAppend) {
		t.Fatalf("expected signed non-appendable status to block approval, got %v", err)
	}
}

func TestStage330ReviewUsesTargetedFullHistoryInsteadOfLatestPage(t *testing.T) {
	ticker := "SBER"
	quantity := decimal.Must("2.00000000")
	unitPrice := verticalslice.Money{Amount: decimal.Must("100.00000000"), Currency: verticalslice.RUB}
	settlementDate := "2026-01-13"
	store := &importAPITestStore{
		existingTransactions: []verticalslice.Transaction{{
			ID:              "00000000-0000-4000-8000-000000000100",
			EntryID:         "00000000-0000-4000-8000-000000000101",
			PortfolioID:     "00000000-0000-4000-8000-000000000002",
			TransactionType: "BUY",
			Ticker:          &ticker,
			Quantity:        &quantity,
			UnitPrice:       &unitPrice,
			GrossAmount:     verticalslice.Money{Amount: decimal.Must("200.00000000"), Currency: verticalslice.RUB},
			Commission:      verticalslice.Money{Amount: decimal.Must("1.00000000"), Currency: verticalslice.RUB},
			Tax:             verticalslice.ZeroMoney(),
			TradeDate:       "2026-01-10",
			SettlementDate:  &settlementDate,
		}},
	}
	app := NewDevelopment(verticalslice.NewService(store, fixedHTTPClock{}))
	request := httptest.NewRequest(http.MethodPost,
		"/api/v1/portfolios/00000000-0000-4000-8000-000000000002/imports/review",
		bytes.NewBufferString(`{"sourceAccountLabel":"Manual CSV","csvPayload":`+quote(importCSV)+`}`))
	request.Header.Set("Content-Type", "application/json")

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("review import: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.StatusCode)
	}
	var payload struct {
		Data struct {
			Rows []struct {
				Status string `json:"status"`
			} `json:"rows"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode review: %v", err)
	}
	if len(payload.Data.Rows) != 1 || payload.Data.Rows[0].Status != importer.ReviewStatusDuplicate {
		t.Fatalf("expected relevant historical duplicate to be surfaced, got %+v", payload.Data.Rows)
	}
	if store.listTransactionsCalls != 0 || store.listImportReviewCalls != 1 {
		t.Fatalf("expected targeted history lookup only, got list=%d targeted=%d", store.listTransactionsCalls, store.listImportReviewCalls)
	}
	if len(store.importReviewFilter.TradeDates) != 1 || store.importReviewFilter.TradeDates[0] != "2026-01-10" {
		t.Fatalf("expected parser-derived target date, got %+v", store.importReviewFilter)
	}
}

func TestStage330OverLimitReviewFailsBeforeHistoryLookup(t *testing.T) {
	store := &importAPITestStore{}
	app := NewDevelopment(verticalslice.NewService(store, fixedHTTPClock{}))
	payload := importCSVWithRows(importer.MaxReviewRows + 1)
	request := httptest.NewRequest(http.MethodPost,
		"/api/v1/portfolios/00000000-0000-4000-8000-000000000002/imports/review",
		bytes.NewBufferString(`{"sourceAccountLabel":"Manual CSV","csvPayload":`+quote(payload)+`}`))
	request.Header.Set("Content-Type", "application/json")

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("review over-limit import: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.StatusCode)
	}
	if store.listImportReviewCalls != 0 || store.listTransactionsCalls != 0 {
		t.Fatalf("expected row bound before any ledger history lookup, got list=%d targeted=%d", store.listTransactionsCalls, store.listImportReviewCalls)
	}
}
