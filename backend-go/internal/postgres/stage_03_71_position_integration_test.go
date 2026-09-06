package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

const stage371SnapshotMethodologyTest = "stage-03-71-position-cost-snapshot-v1"

type stage371Harness struct {
	ctx         context.Context
	db          *sql.DB
	service     *verticalslice.Service
	subjectID   string
	portfolioID string
}

func newStage371Harness(t *testing.T, name string) stage371Harness {
	t.Helper()
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}

	store, err := postgres.Open(databaseURL)
	if err != nil {
		t.Fatalf("open Stage 3.71 postgres store: %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Fatalf("close Stage 3.71 postgres store: %v", err)
		}
	})

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open Stage 3.71 verification db: %v", err)
	}
	closeDBOnCleanup(t, db, "Stage 3.71 verification")

	ctx := context.Background()
	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	subjectID := uuid.NewString()
	portfolio, err := service.CreatePortfolio(
		ctx,
		verticalslice.RequestContext{},
		subjectID,
		uuid.NewString(),
		"/api/v1/portfolios",
		verticalslice.CreatePortfolioRequest{Name: name, BaseCurrency: verticalslice.RUB},
	)
	if err != nil {
		t.Fatalf("create Stage 3.71 test portfolio: %v", err)
	}
	t.Cleanup(func() {
		cleanupPortfolioRows(t, ctx, db, portfolio.ID)
	})

	return stage371Harness{
		ctx:         ctx,
		db:          db,
		service:     service,
		subjectID:   subjectID,
		portfolioID: portfolio.ID,
	}
}

func stage371Trade(portfolioID string, transactionType string, ticker string, quantity string, unitPrice string, tradeDate string) verticalslice.AppendTransactionRequest {
	parsedQuantity := decimal.Must(quantity)
	price := verticalslice.Money{Amount: decimal.Must(unitPrice), Currency: verticalslice.RUB}
	return verticalslice.AppendTransactionRequest{
		PortfolioID:     portfolioID,
		TransactionType: transactionType,
		Ticker:          &ticker,
		Quantity:        &parsedQuantity,
		UnitPrice:       &price,
		Commission:      verticalslice.ZeroMoney(),
		Tax:             verticalslice.ZeroMoney(),
		TradeDate:       tradeDate,
	}
}

func appendStage371Trade(t *testing.T, h stage371Harness, request verticalslice.AppendTransactionRequest) verticalslice.Transaction {
	t.Helper()
	transaction, err := h.service.AppendTransaction(
		h.ctx,
		verticalslice.RequestContext{},
		h.subjectID,
		uuid.NewString(),
		"/api/v1/portfolios/"+h.portfolioID+"/transactions",
		request,
	)
	if err != nil {
		t.Fatalf("append Stage 3.71 %s: %v", request.TransactionType, err)
	}
	return transaction
}

func stage371SnapshotValues(t *testing.T, h stage371Harness, snapshotDate string) (string, string) {
	t.Helper()
	var stockValue string
	var bondValue string
	if err := h.db.QueryRowContext(h.ctx, `
		SELECT stock_value_amount::text, bond_value_amount::text
		FROM analytics.portfolio_snapshots
		WHERE portfolio_id = $1
			AND snapshot_date = $2::date
			AND methodology_version = $3
		ORDER BY snapshot_version DESC
		LIMIT 1
	`, h.portfolioID, snapshotDate, stage371SnapshotMethodologyTest).Scan(&stockValue, &bondValue); err != nil {
		t.Fatalf("read Stage 3.71 snapshot %s: %v", snapshotDate, err)
	}
	return stockValue, bondValue
}

func assertContiguousLedgerSequence(t *testing.T, h stage371Harness, expectedRows int) {
	t.Helper()
	rows, err := h.db.QueryContext(h.ctx, `
		SELECT ledger_sequence
		FROM investment.transaction_entries
		WHERE portfolio_id = $1
		ORDER BY ledger_sequence ASC
	`, h.portfolioID)
	if err != nil {
		t.Fatalf("query Stage 3.71 ledger sequence: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
		var sequence int64
		if err := rows.Scan(&sequence); err != nil {
			t.Fatalf("scan Stage 3.71 ledger sequence: %v", err)
		}
		if sequence != int64(count) {
			t.Fatalf("ledger sequence gap/order mismatch at row %d: got %d", count, sequence)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate Stage 3.71 ledger sequence: %v", err)
	}
	if count != expectedRows {
		t.Fatalf("expected %d committed ledger rows, got %d", expectedRows, count)
	}
}

func TestStage371PositionLifecycleUsesRemainingAcquisitionBasis(t *testing.T) {
	h := newStage371Harness(t, "Stage 3.71 lifecycle")

	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "10.00000000", "100.00000000", "2026-09-01"))
	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "10.00000000", "200.00000000", "2026-09-01"))
	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SU26238RMFS4", "2.00000000", "1000.00000000", "2026-09-01"))
	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "SELL", "SBER", "5.00000000", "999.00000000", "2026-09-01"))

	stockValue, bondValue := stage371SnapshotValues(t, h, "2026-09-01")
	if stockValue != "2250.00000000" {
		t.Fatalf("partial SELL must preserve authoritative WAC: stock basis got %s want 2250.00000000", stockValue)
	}
	if bondValue != "2000.00000000" {
		t.Fatalf("bond acquisition basis got %s want 2000.00000000", bondValue)
	}

	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "SELL", "SBER", "15.00000000", "1.00000000", "2026-09-02"))
	stockValue, bondValue = stage371SnapshotValues(t, h, "2026-09-02")
	if stockValue != "0.00000000" || bondValue != "2000.00000000" {
		t.Fatalf("full close snapshot mismatch: stock=%s bond=%s", stockValue, bondValue)
	}

	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "2.00000000", "333.33333333", "2026-09-03"))
	stockValue, bondValue = stage371SnapshotValues(t, h, "2026-09-03")
	if stockValue != "666.66666666" || bondValue != "2000.00000000" {
		t.Fatalf("reopen must start fresh WAC state: stock=%s bond=%s", stockValue, bondValue)
	}

	assertContiguousLedgerSequence(t, h, 6)
}

func TestStage371BackdatedOversellRollsBackLedgerSnapshotsAndReplayReservation(t *testing.T) {
	h := newStage371Harness(t, "Stage 3.71 backdated oversell")
	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "10.00000000", "100.00000000", "2026-09-10"))
	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "SELL", "SBER", "8.00000000", "120.00000000", "2026-09-20"))

	var snapshotsBefore int64
	if err := h.db.QueryRowContext(h.ctx, `SELECT COUNT(*) FROM analytics.portfolio_snapshots WHERE portfolio_id = $1`, h.portfolioID).Scan(&snapshotsBefore); err != nil {
		t.Fatalf("count snapshots before backdated oversell: %v", err)
	}

	oversell := stage371Trade(h.portfolioID, "SELL", "SBER", "5.00000000", "110.00000000", "2026-09-15")
	idempotencyKey := uuid.NewString()
	requestPath := "/api/v1/portfolios/" + h.portfolioID + "/transactions"
	builderCalls := 0
	builder := func(verticalslice.Transaction) (verticalslice.CommandReplayArtifact, error) {
		builderCalls++
		return verticalslice.CommandReplayArtifact{
			StatusCode: 201,
			Body:       []byte(`{"data":{"status":"unexpected"}}`),
			RequestID:  uuid.NewString(),
			TraceID:    "10000000000000000000000000000001",
		}, nil
	}

	for attempt := 1; attempt <= 2; attempt++ {
		_, _, err := h.service.AppendTransactionWithReplay(
			h.ctx,
			verticalslice.RequestContext{},
			h.subjectID,
			idempotencyKey,
			requestPath,
			oversell,
			builder,
		)
		if !errors.Is(err, verticalslice.ErrInsufficientPositionQuantity) {
			t.Fatalf("attempt %d expected backdated oversell rejection, got %v", attempt, err)
		}
	}
	if builderCalls != 0 {
		t.Fatalf("replay artifact builder must not run for rejected oversell; calls=%d", builderCalls)
	}

	var sellRows int64
	if err := h.db.QueryRowContext(h.ctx, `
		SELECT COUNT(*)
		FROM investment.transaction_entries
		WHERE portfolio_id = $1 AND transaction_type = 'SELL'
	`, h.portfolioID).Scan(&sellRows); err != nil {
		t.Fatalf("count committed SELL rows after oversell: %v", err)
	}
	if sellRows != 1 {
		t.Fatalf("rejected backdated SELL leaked into immutable ledger: committed SELL rows=%d", sellRows)
	}

	var snapshotsAfter int64
	if err := h.db.QueryRowContext(h.ctx, `SELECT COUNT(*) FROM analytics.portfolio_snapshots WHERE portfolio_id = $1`, h.portfolioID).Scan(&snapshotsAfter); err != nil {
		t.Fatalf("count snapshots after backdated oversell: %v", err)
	}
	if snapshotsAfter != snapshotsBefore {
		t.Fatalf("rejected oversell mutated snapshots: before=%d after=%d", snapshotsBefore, snapshotsAfter)
	}

	var replayRows int64
	if err := h.db.QueryRowContext(h.ctx, `
		SELECT COUNT(*)
		FROM investment.command_deduplication
		WHERE principal_id = $1
			AND method = 'POST'
			AND canonical_path = $2
			AND idempotency_key = $3
	`, h.subjectID, requestPath, idempotencyKey).Scan(&replayRows); err != nil {
		t.Fatalf("count rejected oversell replay reservation: %v", err)
	}
	if replayRows != 0 {
		t.Fatalf("rejected oversell retained idempotency reservation rows=%d", replayRows)
	}

	assertContiguousLedgerSequence(t, h, 2)
}

func TestStage371ConcurrentSellSerializesAndCommitsOnlyLegalQuantity(t *testing.T) {
	h := newStage371Harness(t, "Stage 3.71 concurrent SELL")
	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "10.00000000", "100.00000000", "2026-09-01"))

	requests := []verticalslice.AppendTransactionRequest{
		stage371Trade(h.portfolioID, "SELL", "SBER", "7.00000000", "110.00000000", "2026-09-02"),
		stage371Trade(h.portfolioID, "SELL", "SBER", "7.00000000", "120.00000000", "2026-09-02"),
	}
	start := make(chan struct{})
	errs := make([]error, len(requests))
	var wg sync.WaitGroup
	for index := range requests {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			<-start
			_, errs[index] = h.service.AppendTransaction(
				h.ctx,
				verticalslice.RequestContext{},
				h.subjectID,
				uuid.NewString(),
				"/api/v1/portfolios/"+h.portfolioID+"/transactions",
				requests[index],
			)
		}(index)
	}
	close(start)
	wg.Wait()

	successes := 0
	oversells := 0
	for _, err := range errs {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, verticalslice.ErrInsufficientPositionQuantity):
			oversells++
		default:
			t.Fatalf("unexpected concurrent SELL result: %v", err)
		}
	}
	if successes != 1 || oversells != 1 {
		t.Fatalf("expected one committed SELL and one oversell rejection, successes=%d oversells=%d errors=%v", successes, oversells, errs)
	}

	stockValue, _ := stage371SnapshotValues(t, h, "2026-09-02")
	if stockValue != "300.00000000" {
		t.Fatalf("concurrent SELL final acquisition basis got %s want 300.00000000", stockValue)
	}
	assertContiguousLedgerSequence(t, h, 2)
}

func TestStage371ConcurrentBuySellPreservesSerializablePositionInvariant(t *testing.T) {
	h := newStage371Harness(t, "Stage 3.71 concurrent BUY SELL")
	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "5.00000000", "100.00000000", "2026-09-01"))

	buy := stage371Trade(h.portfolioID, "BUY", "SBER", "5.00000000", "200.00000000", "2026-09-02")
	sell := stage371Trade(h.portfolioID, "SELL", "SBER", "7.00000000", "150.00000000", "2026-09-02")
	start := make(chan struct{})
	var buyErr error
	var sellErr error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		_, buyErr = h.service.AppendTransaction(h.ctx, verticalslice.RequestContext{}, h.subjectID, uuid.NewString(), "/api/v1/portfolios/"+h.portfolioID+"/transactions", buy)
	}()
	go func() {
		defer wg.Done()
		<-start
		_, sellErr = h.service.AppendTransaction(h.ctx, verticalslice.RequestContext{}, h.subjectID, uuid.NewString(), "/api/v1/portfolios/"+h.portfolioID+"/transactions", sell)
	}()
	close(start)
	wg.Wait()

	if buyErr != nil {
		t.Fatalf("concurrent BUY must commit in either legal serial order: %v", buyErr)
	}
	stockValue, _ := stage371SnapshotValues(t, h, "2026-09-02")
	switch {
	case sellErr == nil:
		if stockValue != "450.00000000" {
			t.Fatalf("BUY-before-SELL serial outcome basis got %s want 450.00000000", stockValue)
		}
		assertContiguousLedgerSequence(t, h, 3)
	case errors.Is(sellErr, verticalslice.ErrInsufficientPositionQuantity):
		if stockValue != "1500.00000000" {
			t.Fatalf("SELL-before-BUY rejection outcome basis got %s want 1500.00000000", stockValue)
		}
		assertContiguousLedgerSequence(t, h, 2)
	default:
		t.Fatalf("unexpected concurrent SELL result: %v", sellErr)
	}
}

func TestStage371ImportSequenceFollowsAppendPlanAndCraftedSellRemainsRejected(t *testing.T) {
	h := newStage371Harness(t, "Stage 3.71 import sequence")
	first := stage371Trade(h.portfolioID, "BUY", "SBER", "1.00000000", "100.00000000", "2026-10-10")
	second := stage371Trade(h.portfolioID, "BUY", "SBER", "2.00000000", "101.00000000", "2026-10-01")
	fileHash := strings.Repeat("a", 64)

	transactions, err := h.service.AppendImportedTransactions(
		h.ctx,
		verticalslice.RequestContext{},
		h.subjectID,
		uuid.NewString(),
		"/api/v1/portfolios/"+h.portfolioID+"/imports/append",
		verticalslice.AppendImportBatchRequest{
			PortfolioID:        h.portfolioID,
			Transactions:       []verticalslice.AppendTransactionRequest{first, second},
			SourceKind:         "USER_UPLOADED_FILE",
			SourceAccountLabel: "stage-03-71-import",
			SourceFileHash:     fileHash,
		},
	)
	if err != nil {
		t.Fatalf("append Stage 3.71 import batch: %v", err)
	}
	if len(transactions) != 2 {
		t.Fatalf("expected 2 imported transactions, got %d", len(transactions))
	}

	rows, err := h.db.QueryContext(h.ctx, `
		SELECT ledger_sequence, trade_date::text
		FROM investment.transaction_entries
		WHERE portfolio_id = $1 AND source_file_hash = $2
		ORDER BY ledger_sequence ASC
	`, h.portfolioID, fileHash)
	if err != nil {
		t.Fatalf("query imported ledger sequence range: %v", err)
	}
	defer rows.Close()

	expectedDates := []string{"2026-10-10", "2026-10-01"}
	index := 0
	for rows.Next() {
		if index >= len(expectedDates) {
			t.Fatalf("import produced extra ledger rows")
		}
		var sequence int64
		var tradeDate string
		if err := rows.Scan(&sequence, &tradeDate); err != nil {
			t.Fatalf("scan imported ledger sequence: %v", err)
		}
		if sequence != int64(index+1) || tradeDate != expectedDates[index] {
			t.Fatalf("import append-plan ordering mismatch at %d: sequence=%d tradeDate=%s", index, sequence, tradeDate)
		}
		index++
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate imported ledger sequence: %v", err)
	}
	if index != len(expectedDates) {
		t.Fatalf("expected %d imported sequence rows, got %d", len(expectedDates), index)
	}

	craftedSell := stage371Trade(h.portfolioID, "SELL", "SBER", "1.00000000", "105.00000000", "2026-10-11")
	_, err = h.service.AppendImportedTransactions(
		h.ctx,
		verticalslice.RequestContext{},
		h.subjectID,
		uuid.NewString(),
		"/api/v1/portfolios/"+h.portfolioID+"/imports/append",
		verticalslice.AppendImportBatchRequest{
			PortfolioID:        h.portfolioID,
			Transactions:       []verticalslice.AppendTransactionRequest{craftedSell},
			SourceKind:         "USER_UPLOADED_FILE",
			SourceAccountLabel: "stage-03-71-import",
			SourceFileHash:     strings.Repeat("b", 64),
		},
	)
	if !errors.Is(err, verticalslice.ErrInvalidInput) {
		t.Fatalf("crafted imported SELL must remain rejected, got %v", err)
	}

	var importedSellRows int64
	if err := h.db.QueryRowContext(h.ctx, `
		SELECT COUNT(*)
		FROM investment.transaction_entries
		WHERE portfolio_id = $1
			AND source_kind = 'USER_UPLOADED_FILE'
			AND transaction_type = 'SELL'
	`, h.portfolioID).Scan(&importedSellRows); err != nil {
		t.Fatalf("count crafted imported SELL rows: %v", err)
	}
	if importedSellRows != 0 {
		t.Fatalf("crafted imported SELL leaked into ledger: rows=%d", importedSellRows)
	}
	assertContiguousLedgerSequence(t, h, 2)
}
