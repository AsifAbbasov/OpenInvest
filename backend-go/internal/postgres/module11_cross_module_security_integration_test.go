package postgres_test

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/auth"
	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func m11Artifact(requestContext verticalslice.RequestContext, transaction verticalslice.Transaction) (verticalslice.CommandReplayArtifact, error) {
	return verticalslice.CommandReplayArtifact{
		StatusCode: 201,
		Body:       []byte(transaction.ID),
		RequestID:  requestContext.RequestID,
		TraceID:    requestContext.TraceID,
	}, nil
}

func TestM11ForeignSubjectReadMatrix(t *testing.T) {
	h := newStage371Harness(t, "M11 foreign subject read matrix")
	stage377AppendCash(t, h, "DEPOSIT", "1000.00000000", "2025-01-01")
	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "2.00000000", "100.00000000", "2025-01-02"))
	foreign := uuid.NewString()

	cases := []struct {
		name string
		run  func() error
	}{
		{"portfolio", func() error { _, err := h.service.GetPortfolio(h.ctx, foreign, h.portfolioID); return err }},
		{"transactions", func() error { _, err := h.service.ListTransactions(h.ctx, foreign, h.portfolioID, verticalslice.TransactionFilter{Limit: 10}); return err }},
		{"positions", func() error { _, err := h.service.GetPortfolioPositions(h.ctx, foreign, h.portfolioID, ""); return err }},
		{"summary", func() error { _, err := h.service.GetPortfolioSummary(h.ctx, foreign, h.portfolioID, "2025-01-02"); return err }},
		{"cash-flow", func() error { _, err := h.service.GetPortfolioCashFlow(h.ctx, foreign, h.portfolioID, "", ""); return err }},
		{"returns", func() error { _, err := h.service.GetPortfolioReturns(h.ctx, foreign, h.portfolioID, "2026-01-01"); return err }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.run(); !errors.Is(err, postgres.ErrNotFound) {
				t.Fatalf("foreign subject %s error=%v want postgres.ErrNotFound", tc.name, err)
			}
		})
	}
	t.Log("M11_CM01_FOREIGN_SUBJECT_READ_MATRIX=PASS")
}

func TestM11IdempotencyOwnershipReplayIsolation(t *testing.T) {
	h := newStage371Harness(t, "M11 idempotency ownership")
	key := "m11-shared-key-ownership-0001"
	path := "/api/v1/portfolios/" + h.portfolioID + "/transactions"
	request := stage371Trade(h.portfolioID, "BUY", "SBER", "1.00000000", "100.00000000", "2026-01-02")
	rc := verticalslice.RequestContext{RequestID: uuid.NewString(), TraceID: "m11-idempotency-owner"}

	first, firstArtifact, err := h.service.AppendTransactionWithReplay(h.ctx, rc, h.subjectID, key, path, request, func(tx verticalslice.Transaction) (verticalslice.CommandReplayArtifact, error) {
		return m11Artifact(rc, tx)
	})
	if err != nil {
		t.Fatalf("owner append: %v", err)
	}
	_, replayArtifact, err := h.service.AppendTransactionWithReplay(h.ctx, rc, h.subjectID, key, path, request, func(tx verticalslice.Transaction) (verticalslice.CommandReplayArtifact, error) {
		return m11Artifact(rc, tx)
	})
	if err != nil {
		t.Fatalf("owner replay: %v", err)
	}
	if string(replayArtifact.Body) != string(firstArtifact.Body) {
		t.Fatalf("owner replay artifact drifted: first=%q replay=%q", string(firstArtifact.Body), string(replayArtifact.Body))
	}
	t.Log("M11_COMPLETED_REPLAY_TYPED_RESULT=ARTIFACT_ONLY_AS_DESIGNED")

	foreignSubject := uuid.NewString()
	if _, _, err := h.service.AppendTransactionWithReplay(h.ctx, rc, foreignSubject, key, path, request, func(tx verticalslice.Transaction) (verticalslice.CommandReplayArtifact, error) {
		return m11Artifact(rc, tx)
	}); !errors.Is(err, postgres.ErrNotFound) {
		t.Fatalf("foreign principal with same idempotency key error=%v want not found", err)
	}

	changed := stage371Trade(h.portfolioID, "BUY", "SBER", "2.00000000", "100.00000000", "2026-01-02")
	if _, _, err := h.service.AppendTransactionWithReplay(h.ctx, rc, h.subjectID, key, path, changed, func(tx verticalslice.Transaction) (verticalslice.CommandReplayArtifact, error) {
		return m11Artifact(rc, tx)
	}); !errors.Is(err, postgres.ErrIdempotencyConflict) {
		t.Fatalf("changed body with same scoped key error=%v want idempotency conflict", err)
	}

	secondPortfolio, err := h.service.CreatePortfolio(
		h.ctx,
		verticalslice.RequestContext{},
		h.subjectID,
		"m11-second-portfolio-key-0001",
		"/api/v1/portfolios",
		verticalslice.CreatePortfolioRequest{Name: "M11 second portfolio", BaseCurrency: verticalslice.RUB},
	)
	if err != nil {
		t.Fatalf("create second portfolio: %v", err)
	}
	t.Cleanup(func() { cleanupPortfolioRows(t, h.ctx, h.db, secondPortfolio.ID) })
	secondPath := "/api/v1/portfolios/" + secondPortfolio.ID + "/transactions"
	secondRequest := stage371Trade(secondPortfolio.ID, "BUY", "SBER", "1.00000000", "101.00000000", "2026-01-03")
	if _, _, err := h.service.AppendTransactionWithReplay(h.ctx, rc, h.subjectID, key, secondPath, secondRequest, func(tx verticalslice.Transaction) (verticalslice.CommandReplayArtifact, error) {
		return m11Artifact(rc, tx)
	}); err != nil {
		t.Fatalf("same idempotency key on distinct canonical portfolio path must be independently scoped: %v", err)
	}

	var firstRows int
	if err := h.db.QueryRowContext(h.ctx, "SELECT count(*) FROM investment.transaction_entries WHERE portfolio_id=$1", h.portfolioID).Scan(&firstRows); err != nil {
		t.Fatal(err)
	}
	if firstRows != 1 {
		t.Fatalf("first portfolio business effects=%d want=1", firstRows)
	}
	var foreignReservations int
	if err := h.db.QueryRowContext(h.ctx, "SELECT count(*) FROM investment.command_deduplication WHERE principal_id=$1 AND canonical_path=$2 AND idempotency_key=$3", foreignSubject, path, key).Scan(&foreignReservations); err != nil {
		t.Fatal(err)
	}
	if foreignReservations != 0 {
		t.Fatalf("foreign ownership failure left replay reservation rows=%d", foreignReservations)
	}
	t.Log("M11_CM02_NO_CROSS_PRINCIPAL_MUTATION=PASS")
	t.Log("M11_CM03_IDEMPOTENCY_SCOPE_AND_EXACT_REPLAY=PASS")
}

func TestM11CancellationReplayAtomicity(t *testing.T) {
	h := newStage371Harness(t, "M11 cancellation replay atomicity")
	blocker, err := h.db.BeginTx(h.ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer blocker.Rollback()
	var locked string
	if err := blocker.QueryRowContext(h.ctx, "SELECT id::text FROM investment.portfolios WHERE id=$1 FOR UPDATE", h.portfolioID).Scan(&locked); err != nil {
		t.Fatalf("lock portfolio: %v", err)
	}

	key := "m11-cancel-replay-key-0001"
	path := "/api/v1/portfolios/" + h.portfolioID + "/transactions"
	request := stage371Trade(h.portfolioID, "BUY", "SBER", "1.00000000", "100.00000000", "2026-02-01")
	rc := verticalslice.RequestContext{RequestID: uuid.NewString(), TraceID: "m11-cancel"}

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	_, _, err = h.service.AppendTransactionWithReplay(ctx, rc, h.subjectID, key, path, request, func(tx verticalslice.Transaction) (verticalslice.CommandReplayArtifact, error) {
		return m11Artifact(rc, tx)
	})
	if err == nil || ctx.Err() == nil {
		t.Fatalf("blocked mutation was not canceled: err=%v ctx=%v", err, ctx.Err())
	}
	if err := blocker.Rollback(); err != nil {
		t.Fatalf("rollback blocker: %v", err)
	}

	var ledgerRows int
	if err := h.db.QueryRowContext(h.ctx, "SELECT count(*) FROM investment.transaction_entries WHERE portfolio_id=$1", h.portfolioID).Scan(&ledgerRows); err != nil {
		t.Fatal(err)
	}
	if ledgerRows != 0 {
		t.Fatalf("canceled command persisted ledger rows=%d", ledgerRows)
	}
	var reservations int
	if err := h.db.QueryRowContext(h.ctx, "SELECT count(*) FROM investment.command_deduplication WHERE principal_id=$1 AND canonical_path=$2 AND idempotency_key=$3", h.subjectID, path, key).Scan(&reservations); err != nil {
		t.Fatal(err)
	}
	if reservations != 0 {
		t.Fatalf("canceled transaction left idempotency reservation rows=%d", reservations)
	}

	if _, _, err := h.service.AppendTransactionWithReplay(h.ctx, rc, h.subjectID, key, path, request, func(tx verticalslice.Transaction) (verticalslice.CommandReplayArtifact, error) {
		return m11Artifact(rc, tx)
	}); err != nil {
		t.Fatalf("retry after canceled transaction did not recover: %v", err)
	}
	t.Log("M11_CM07_CANCELLATION_BEFORE_MUTATION_COMMIT=PASS")
	t.Log("M11_RETRY_AFTER_CANCELLATION=PASS")
}

func TestM11ConcurrentImportVsSellSerializes(t *testing.T) {
	h := newStage371Harness(t, "M11 import vs sell")
	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "10.00000000", "100.00000000", "2026-03-01"))

	imported := stage371Trade(h.portfolioID, "BUY", "SBER", "5.00000000", "120.00000000", "2026-03-02")
	batch := verticalslice.AppendImportBatchRequest{
		PortfolioID:        h.portfolioID,
		Transactions:       []verticalslice.AppendTransactionRequest{imported},
		SourceKind:         "USER_UPLOADED_FILE",
		SourceAccountLabel: "m11-account",
		SourceFileHash:     testImportFileHashC,
	}
	sell := stage371Trade(h.portfolioID, "SELL", "SBER", "8.00000000", "130.00000000", "2026-03-03")
	start := make(chan struct{})
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		_, err := h.service.AppendImportedTransactions(h.ctx, verticalslice.RequestContext{}, h.subjectID, "m11-import-concurrent-key-0001", "/internal/imports/append", batch)
		errs <- err
	}()
	go func() {
		defer wg.Done()
		<-start
		_, err := h.service.AppendTransaction(h.ctx, verticalslice.RequestContext{}, h.subjectID, "m11-sell-concurrent-key-0001", "/api/v1/portfolios/"+h.portfolioID+"/transactions", sell)
		errs <- err
	}()
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent import/sell error=%v", err)
		}
	}
	positions, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(positions.Items) != 1 || positions.Items[0].Quantity.String() != "7.00000000" {
		t.Fatalf("import/sell final position drift: %+v", positions.Items)
	}
	assertContiguousLedgerSequence(t, h, 3)
	t.Log("M11_IMPORT_VS_SELL_SERIALIZATION=PASS")
}

func TestM11ConcurrentSellVsSellNoOversell(t *testing.T) {
	h := newStage371Harness(t, "M11 sell vs sell")
	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "10.00000000", "100.00000000", "2026-04-01"))
	start := make(chan struct{})
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for index, price := range []string{"120.00000000", "121.00000000"} {
		wg.Add(1)
		go func(index int, price string) {
			defer wg.Done()
			<-start
			_, err := h.service.AppendTransaction(
				h.ctx,
				verticalslice.RequestContext{},
				h.subjectID,
				[]string{"m11-sell-race-key-0001", "m11-sell-race-key-0002"}[index],
				"/api/v1/portfolios/"+h.portfolioID+"/transactions",
				stage371Trade(h.portfolioID, "SELL", "SBER", "7.00000000", price, "2026-04-02"),
			)
			errs <- err
		}(index, price)
	}
	close(start)
	wg.Wait()
	close(errs)
	success, insufficient := 0, 0
	for err := range errs {
		switch {
		case err == nil:
			success++
		case errors.Is(err, verticalslice.ErrInsufficientPositionQuantity):
			insufficient++
		default:
			t.Fatalf("unexpected sell race error=%v", err)
		}
	}
	if success != 1 || insufficient != 1 {
		t.Fatalf("sell race success=%d insufficient=%d", success, insufficient)
	}
	positions, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(positions.Items) != 1 || positions.Items[0].Quantity.String() != "3.00000000" {
		t.Fatalf("sell race final positions=%+v", positions.Items)
	}
	t.Log("M11_SELL_VS_SELL_SERIALIZATION=PASS")
}

func TestM11ConcurrentCorrectionVsReversalSerializes(t *testing.T) {
	h := newStage371Harness(t, "M11 correction vs reversal")
	original := appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "10.00000000", "100.00000000", "2026-05-01"))
	start := make(chan struct{})
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		_, err := correctStage374(t, h, original, "10.00000000", "110.00000000", "2026-05-01", "m11-correct-race-key-0001")
		errs <- err
	}()
	go func() {
		defer wg.Done()
		<-start
		_, err := reverseStage374(t, h, original, "2026-05-02", "m11-reverse-race-key-0001")
		errs <- err
	}()
	close(start)
	wg.Wait()
	close(errs)
	success, conflicts := 0, 0
	for err := range errs {
		switch {
		case err == nil:
			success++
		case errors.Is(err, verticalslice.ErrTransactionConflict):
			conflicts++
		default:
			t.Fatalf("unexpected correction/reversal race error=%v", err)
		}
	}
	if success != 1 || conflicts != 1 {
		t.Fatalf("correction/reversal race success=%d conflicts=%d", success, conflicts)
	}
	assertContiguousLedgerSequence(t, h, 2)
	t.Log("M11_CORRECTION_VS_REVERSAL_SERIALIZATION=PASS")
}

func m11MoneyTotalsEqual(a, b verticalslice.PortfolioCashFlowTotals) bool {
	return a.Deposits.Amount.Equal(b.Deposits.Amount) &&
		a.Withdrawals.Amount.Equal(b.Withdrawals.Amount) &&
		a.BuyOutflows.Amount.Equal(b.BuyOutflows.Amount) &&
		a.SellInflows.Amount.Equal(b.SellInflows.Amount) &&
		a.DividendsGross.Amount.Equal(b.DividendsGross.Amount) &&
		a.CouponsGross.Amount.Equal(b.CouponsGross.Amount) &&
		a.Fees.Amount.Equal(b.Fees.Amount) &&
		a.Taxes.Amount.Equal(b.Taxes.Amount) &&
		a.NetExternalFlow.Amount.Equal(b.NetExternalFlow.Amount) &&
		a.NetInvestmentIncome.Amount.Equal(b.NetInvestmentIncome.Amount) &&
		a.NetCashFlow.Amount.Equal(b.NetCashFlow.Amount)
}

func TestM11EffectiveLedgerProjectionEquivalence(t *testing.T) {
	mutated := newStage371Harness(t, "M11 mutated effective ledger")
	reference := newStage371Harness(t, "M11 clean reference ledger")
	t.Cleanup(func() { cleanupStage376Valuations(t, mutated) })
	t.Cleanup(func() { cleanupStage376Valuations(t, reference) })

	deposit := stage377AppendCash(t, mutated, "DEPOSIT", "1000.00000000", "2026-01-10")
	stage377CorrectCash(t, mutated, deposit, "1200.00000000", "2026-01-01")
	withdrawal := stage377AppendCash(t, mutated, "WITHDRAWAL", "100.00000000", "2026-04-01")
	stage377CorrectCash(t, mutated, withdrawal, "150.00000000", "2026-04-15")
	dividend := appendStage371Trade(t, mutated, stage375Request(mutated.portfolioID, "DIVIDEND", "SBER", "50.00000000", "0.00000000", "0.00000000", "2026-02-10"))
	if _, err := reverseStage374(t, mutated, dividend, "2026-02-15", "m11-dividend-reversal-key-0001"); err != nil {
		t.Fatalf("reverse dividend: %v", err)
	}
	buy := appendStage371Trade(t, mutated, stage371Trade(mutated.portfolioID, "BUY", "SBER", "10.00000000", "100.00000000", "2026-03-01"))
	correctedBuy, err := correctStage374(t, mutated, buy, "12.00000000", "90.00000000", "2026-02-20", "m11-buy-correction-key-0001")
	if err != nil {
		t.Fatalf("correct buy: %v", err)
	}
	if correctedBuy.TradeDate != "2026-02-20" {
		t.Fatalf("corrected buy date=%s", correctedBuy.TradeDate)
	}
	appendStage371Trade(t, mutated, stage371Trade(mutated.portfolioID, "SELL", "SBER", "2.00000000", "150.00000000", "2026-05-01"))
	ephemeral := appendStage371Trade(t, mutated, stage371Trade(mutated.portfolioID, "BUY", "SBER", "1.00000000", "200.00000000", "2026-06-01"))
	if _, err := reverseStage374(t, mutated, ephemeral, "2026-06-02", "m11-buy-reversal-key-0001"); err != nil {
		t.Fatalf("reverse ephemeral buy: %v", err)
	}

	stage377AppendCash(t, reference, "DEPOSIT", "1200.00000000", "2026-01-01")
	appendStage371Trade(t, reference, stage371Trade(reference.portfolioID, "BUY", "SBER", "12.00000000", "90.00000000", "2026-02-20"))
	stage377AppendCash(t, reference, "WITHDRAWAL", "150.00000000", "2026-04-15")
	appendStage371Trade(t, reference, stage371Trade(reference.portfolioID, "SELL", "SBER", "2.00000000", "150.00000000", "2026-05-01"))

	upsertStage376(t, mutated, "SBER", "110.00000000", "2026-12-31")
	upsertStage376(t, reference, "SBER", "110.00000000", "2026-12-31")

	mutPos, err := mutated.service.GetPortfolioPositions(mutated.ctx, mutated.subjectID, mutated.portfolioID, "")
	if err != nil {
		t.Fatal(err)
	}
	refPos, err := reference.service.GetPortfolioPositions(reference.ctx, reference.subjectID, reference.portfolioID, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(mutPos.Items) != 1 || len(refPos.Items) != 1 ||
		mutPos.Items[0].Quantity.String() != refPos.Items[0].Quantity.String() ||
		mutPos.Items[0].WeightedAverageCost.Amount.String() != refPos.Items[0].WeightedAverageCost.Amount.String() ||
		mutPos.Items[0].AcquisitionBasis.Amount.String() != refPos.Items[0].AcquisitionBasis.Amount.String() {
		t.Fatalf("position/WAC equivalence failed mutated=%+v reference=%+v", mutPos.Items, refPos.Items)
	}

	mutCash, err := mutated.service.GetPortfolioCashFlow(mutated.ctx, mutated.subjectID, mutated.portfolioID, "", "2026-12-31")
	if err != nil {
		t.Fatal(err)
	}
	refCash, err := reference.service.GetPortfolioCashFlow(reference.ctx, reference.subjectID, reference.portfolioID, "", "2026-12-31")
	if err != nil {
		t.Fatal(err)
	}
	if !m11MoneyTotalsEqual(mutCash.Totals, refCash.Totals) {
		t.Fatalf("cash-flow totals mismatch mutated=%+v reference=%+v", mutCash.Totals, refCash.Totals)
	}

	mutReturns, err := mutated.service.GetPortfolioReturns(mutated.ctx, mutated.subjectID, mutated.portfolioID, "2026-12-31")
	if err != nil {
		t.Fatal(err)
	}
	refReturns, err := reference.service.GetPortfolioReturns(reference.ctx, reference.subjectID, reference.portfolioID, "2026-12-31")
	if err != nil {
		t.Fatal(err)
	}
	if mutReturns.Status != refReturns.Status || mutReturns.Reason != refReturns.Reason ||
		(mutReturns.XIRR == nil) != (refReturns.XIRR == nil) ||
		(mutReturns.XIRR != nil && mutReturns.XIRR.String() != refReturns.XIRR.String()) ||
		len(mutReturns.ExternalCashFlows) != len(refReturns.ExternalCashFlows) {
		t.Fatalf("XIRR projection mismatch mutated=%+v reference=%+v", mutReturns, refReturns)
	}
	for i := range mutReturns.ExternalCashFlows {
		if mutReturns.ExternalCashFlows[i].Date != refReturns.ExternalCashFlows[i].Date ||
			mutReturns.ExternalCashFlows[i].Amount.String() != refReturns.ExternalCashFlows[i].Amount.String() {
			t.Fatalf("XIRR flow[%d] mismatch mutated=%+v reference=%+v", i, mutReturns.ExternalCashFlows[i], refReturns.ExternalCashFlows[i])
		}
	}

	mutSummary, err := mutated.service.GetPortfolioSummary(mutated.ctx, mutated.subjectID, mutated.portfolioID, "2026-12-31")
	if err != nil {
		t.Fatal(err)
	}
	refSummary, err := reference.service.GetPortfolioSummary(reference.ctx, reference.subjectID, reference.portfolioID, "2026-12-31")
	if err != nil {
		t.Fatal(err)
	}
	if mutSummary.TotalValue.Amount.String() != refSummary.TotalValue.Amount.String() ||
		mutSummary.CashValue.Amount.String() != refSummary.CashValue.Amount.String() ||
		mutSummary.StockValue.Amount.String() != refSummary.StockValue.Amount.String() ||
		(mutSummary.XIRR == nil) != (refSummary.XIRR == nil) ||
		(mutSummary.XIRR != nil && mutSummary.XIRR.String() != refSummary.XIRR.String()) {
		t.Fatalf("summary equivalence failed mutated=%+v reference=%+v", mutSummary, refSummary)
	}
	t.Log("M11_CM04_EFFECTIVE_LEDGER_SINGLE_SOURCE=PASS")
	t.Log("M11_CM05_POSITION_WAC_REBUILD_EQUIVALENCE=PASS")
	t.Log("M11_CM06_CASH_FLOW_XIRR_REBUILD_EQUIVALENCE=PASS")
}

func TestM11XIRRReadDuringCorrectionCommittedStatesOnly(t *testing.T) {
	h := newStage371Harness(t, "M11 XIRR read vs mutation")
	deposit := stage377AppendCash(t, h, "DEPOSIT", "100.00000000", "2025-01-01")
	appendStage371Trade(t, h, stage375Request(h.portfolioID, "DIVIDEND", "SBER", "10.00000000", "0.00000000", "0.00000000", "2026-01-01"))

	start := make(chan struct{})
	results := make(chan string, 40)
	errs := make(chan error, 41)
	var wg sync.WaitGroup
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			projection, err := h.service.GetPortfolioReturns(h.ctx, h.subjectID, h.portfolioID, "2026-01-01")
			if err != nil {
				errs <- err
				return
			}
			if projection.XIRR == nil {
				errs <- errors.New("XIRR unexpectedly unavailable during concurrent correction")
				return
			}
			results <- projection.XIRR.String()
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		rc := verticalslice.RequestContext{RequestID: uuid.NewString(), TraceID: "m11-xirr-correction"}
		corrected := stage375Request(h.portfolioID, "DEPOSIT", "", "200.00000000", "0.00000000", "0.00000000", "2025-01-01")
		_, _, err := h.service.CorrectTransactionWithReplay(
			h.ctx, rc, h.subjectID, "m11-xirr-correction-key-0001",
			"/api/v1/portfolios/"+h.portfolioID+"/transactions/"+deposit.ID,
			verticalslice.CorrectTransactionRequest{
				PortfolioID: h.portfolioID, TransactionID: deposit.ID, ExpectedRevision: deposit.Revision,
				Reason: "M11 concurrent XIRR correction", Corrected: corrected,
			},
			func(tx verticalslice.Transaction) (verticalslice.CommandReplayArtifact, error) { return m11Artifact(rc, tx) },
		)
		errs <- err
	}()
	close(start)
	wg.Wait()
	close(results)
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent XIRR/correction error=%v", err)
		}
	}
	for got := range results {
		if got != "0.10000000" && got != "0.05000000" {
			t.Fatalf("reader observed non-committed XIRR state %s", got)
		}
	}
	final, err := h.service.GetPortfolioReturns(h.ctx, h.subjectID, h.portfolioID, "2026-01-01")
	if err != nil || final.XIRR == nil || final.XIRR.String() != "0.05000000" {
		t.Fatalf("final corrected XIRR=%+v err=%v", final, err)
	}
	t.Log("M11_XIRR_READ_VS_MUTATION_COMMITTED_STATES_ONLY=PASS")
}

func TestM11AuthAccessTokenLifecycleMutationObservation(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}
	h := newStage371Harness(t, "M11 access token lifecycle")
	authStore, err := postgres.Open(databaseURL)
	if err != nil {
		t.Fatalf("open auth store: %v", err)
	}
	defer authStore.Close()
	authService, err := auth.NewService(authStore, nil, auth.Config{
		AccessTokenSecret:      []byte("m11-access-token-secret-32-bytes-minimum"),
		AccessTokenTTL:         15 * time.Minute,
		RefreshTokenTTL:        time.Hour,
		RefreshCookieSecure:    true,
		AllowDevelopmentBypass: false,
	})
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}
	registered, err := authService.Register(h.ctx, auth.RegistrationRequest{
		Email: "m11-" + uuid.NewString() + "@example.com", Password: "correct horse battery staple",
		Language: auth.LanguageEN, Theme: auth.ThemeSystem, Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("register audit user: %v", err)
	}
	subject := registered.User.InvestmentSubjectID
	portfolio, err := h.service.CreatePortfolio(
		h.ctx, verticalslice.RequestContext{}, subject, "m11-auth-portfolio-key-0001", "/api/v1/portfolios",
		verticalslice.CreatePortfolioRequest{Name: "M11 bearer lifecycle", BaseCurrency: verticalslice.RUB},
	)
	if err != nil {
		t.Fatalf("create auth-owned portfolio: %v", err)
	}
	t.Cleanup(func() { cleanupPortfolioRows(t, h.ctx, h.db, portfolio.ID) })

	rotated, err := authService.Refresh(h.ctx, registered.RefreshToken, registered.Session.CSRFToken)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	oldSubject, oldErr := authService.AuthenticateAccessToken(registered.Session.AccessToken)
	if oldErr == nil {
		gross := verticalslice.Money{Amount: decimal.Must("10.00000000"), Currency: verticalslice.RUB}
		_, mutationErr := h.service.AppendTransaction(
			h.ctx, verticalslice.RequestContext{}, oldSubject, "m11-stale-access-key-0001",
			"/api/v1/portfolios/"+portfolio.ID+"/transactions",
			verticalslice.AppendTransactionRequest{
				PortfolioID: portfolio.ID, TransactionType: "DEPOSIT", GrossAmount: &gross,
				Commission: verticalslice.ZeroMoney(), Tax: verticalslice.ZeroMoney(), TradeDate: "2026-07-01",
			},
		)
		if mutationErr != nil {
			t.Fatalf("authenticated stale bearer could not execute expected bearer-authorized mutation: %v", mutationErr)
		}
		t.Log("M11_STALE_ACCESS_AFTER_REFRESH_ROTATION=VALID_UNTIL_TTL")
	} else {
		t.Log("M11_STALE_ACCESS_AFTER_REFRESH_ROTATION=REVOKED")
	}

	revoked, err := authService.Logout(h.ctx, rotated.RefreshToken, rotated.Session.CSRFToken, false)
	if err != nil || !revoked {
		t.Fatalf("logout rotated session revoked=%v err=%v", revoked, err)
	}
	loggedOutSubject, loggedOutErr := authService.AuthenticateAccessToken(rotated.Session.AccessToken)
	if loggedOutErr == nil {
		gross := verticalslice.Money{Amount: decimal.Must("20.00000000"), Currency: verticalslice.RUB}
		_, mutationErr := h.service.AppendTransaction(
			h.ctx, verticalslice.RequestContext{}, loggedOutSubject, "m11-post-logout-access-key-0001",
			"/api/v1/portfolios/"+portfolio.ID+"/transactions",
			verticalslice.AppendTransactionRequest{
				PortfolioID: portfolio.ID, TransactionType: "DEPOSIT", GrossAmount: &gross,
				Commission: verticalslice.ZeroMoney(), Tax: verticalslice.ZeroMoney(), TradeDate: "2026-07-02",
			},
		)
		if mutationErr != nil {
			t.Fatalf("still-valid short-lived bearer could not execute bearer-authorized mutation: %v", mutationErr)
		}
		t.Log("M11_ACCESS_AFTER_LOGOUT=VALID_UNTIL_TTL")
	} else {
		t.Log("M11_ACCESS_AFTER_LOGOUT=REVOKED")
	}
	t.Log("M11_ACCESS_TOKEN_CONTRACT=SHORT_LIVED_STATELESS_BEARER_15_MINUTES")
}
