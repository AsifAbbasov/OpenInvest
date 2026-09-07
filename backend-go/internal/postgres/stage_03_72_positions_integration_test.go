package postgres_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestStage372ProjectionCanonicalMixedPortfolioAndStableOrdering(t *testing.T) {
	h := newStage371Harness(t, "Stage 3.72 canonical projection")

	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "100.00000000", "250.00000000", "2026-09-01"))
	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "100.00000000", "300.00000000", "2026-09-02"))
	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "SELL", "SBER", "50.00000000", "999.00000000", "2026-09-03"))
	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "GAZP", "10.00000000", "100.00000000", "2026-09-02"))
	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SU26238RMFS4", "2.00000000", "1000.00000000", "2026-09-02"))

	var snapshotPositionsBefore int64
	if err := h.db.QueryRowContext(h.ctx, `
		SELECT COUNT(*)
		FROM analytics.snapshot_positions sp
		JOIN analytics.portfolio_snapshots ps ON ps.id = sp.snapshot_id
		WHERE ps.portfolio_id = $1
	`, h.portfolioID).Scan(&snapshotPositionsBefore); err != nil {
		t.Fatalf("count snapshot positions before Stage 3.72 read: %v", err)
	}

	first, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "")
	if err != nil {
		t.Fatalf("get Stage 3.72 projection: %v", err)
	}
	second, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "")
	if err != nil {
		t.Fatalf("rebuild Stage 3.72 projection: %v", err)
	}
	if len(first.Items) != 3 || len(second.Items) != 3 {
		t.Fatalf("expected three open positions, got first=%d second=%d", len(first.Items), len(second.Items))
	}
	wantTickers := []string{"GAZP", "SBER", "SU26238RMFS4"}
	for index, want := range wantTickers {
		if first.Items[index].Ticker != want || second.Items[index].Ticker != want {
			t.Fatalf("ticker order is not stable ASC at %d: first=%s second=%s want=%s", index, first.Items[index].Ticker, second.Items[index].Ticker, want)
		}
	}

	sber := first.Items[1]
	if sber.Quantity.String() != "150.00000000" ||
		sber.WeightedAverageCost.Amount.String() != "275.00000000" ||
		sber.AcquisitionBasis.Amount.String() != "41250.00000000" {
		t.Fatalf("canonical SBER vector drifted: %+v", sber)
	}
	if first.TotalAcquisitionBasis.Amount.String() != "44250.00000000" {
		t.Fatalf("total acquisition basis got %s want 44250.00000000", first.TotalAcquisitionBasis.Amount.String())
	}
	wantWeight, err := decimal.Must("41250.00000000").Div(decimal.Must("44250.00000000"))
	if err != nil {
		t.Fatalf("build expected weight: %v", err)
	}
	if sber.AcquisitionBasisWeight == nil || sber.AcquisitionBasisWeight.String() != wantWeight.String() {
		t.Fatalf("SBER acquisition-basis weight got %v want %s", sber.AcquisitionBasisWeight, wantWeight.String())
	}
	if first.InputsAsOf == nil || *first.InputsAsOf != "2026-09-03" {
		t.Fatalf("omitted asOfDate must resolve max included tradeDate, got %v", first.InputsAsOf)
	}
	for _, item := range first.Items {
		if item.MarketValuation.Status != verticalslice.MarketValuationUnavailableStatus ||
			item.MarketValuation.Reason != verticalslice.MarketValuationUnavailableReason {
			t.Fatalf("market state must remain unavailable: %+v", item.MarketValuation)
		}
	}

	stockValue, bondValue := stage371SnapshotValues(t, h, "2026-09-03")
	if stockValue != "42250.00000000" || bondValue != "2000.00000000" {
		t.Fatalf("Stage 3.71 snapshot totals changed after shared replay extraction: stock=%s bond=%s", stockValue, bondValue)
	}
	summary, err := h.service.GetPortfolioSummary(h.ctx, h.subjectID, h.portfolioID, "")
	if err != nil {
		t.Fatalf("read Stage 3.71 summary after Stage 3.72 projection: %v", err)
	}
	if len(summary.Positions) != 0 {
		t.Fatalf("Stage 3.72 must not activate legacy market-valued PortfolioSummary.positions: %+v", summary.Positions)
	}
	var snapshotPositionsAfter int64
	if err := h.db.QueryRowContext(h.ctx, `
		SELECT COUNT(*)
		FROM analytics.snapshot_positions sp
		JOIN analytics.portfolio_snapshots ps ON ps.id = sp.snapshot_id
		WHERE ps.portfolio_id = $1
	`, h.portfolioID).Scan(&snapshotPositionsAfter); err != nil {
		t.Fatalf("count snapshot positions after Stage 3.72 read: %v", err)
	}
	if snapshotPositionsAfter != snapshotPositionsBefore {
		t.Fatalf("Stage 3.72 read mutated analytics.snapshot_positions: before=%d after=%d", snapshotPositionsBefore, snapshotPositionsAfter)
	}
}

func TestStage372ProjectionAsOfCutoffBackdatedAndOmittedFutureSemantics(t *testing.T) {
	h := newStage371Harness(t, "Stage 3.72 as-of semantics")

	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "10.00000000", "100.00000000", "2026-09-10"))
	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "SELL", "SBER", "2.00000000", "999.00000000", "2026-09-20"))
	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "10.00000000", "200.00000000", "2026-09-05"))
	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "2.00000000", "300.00000000", "2030-01-01"))

	explicit, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "2026-09-10")
	if err != nil {
		t.Fatalf("get explicit as-of projection: %v", err)
	}
	if len(explicit.Items) != 1 {
		t.Fatalf("expected one explicit as-of item, got %d", len(explicit.Items))
	}
	if explicit.Items[0].Quantity.String() != "20.00000000" ||
		explicit.Items[0].WeightedAverageCost.Amount.String() != "150.00000000" ||
		explicit.Items[0].AcquisitionBasis.Amount.String() != "3000.00000000" {
		t.Fatalf("explicit cutoff/backdated replay mismatch: %+v", explicit.Items[0])
	}
	if explicit.InputsAsOf == nil || *explicit.InputsAsOf != "2026-09-10" {
		t.Fatalf("explicit inputsAsOf must echo requested BusinessDate, got %v", explicit.InputsAsOf)
	}

	omitted, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "")
	if err != nil {
		t.Fatalf("get omitted as-of projection: %v", err)
	}
	if omitted.Items[0].Quantity.String() != "20.00000000" || omitted.Items[0].WeightedAverageCost.Amount.String() != "165.00000000" {
		t.Fatalf("omitted asOfDate must replay all accepted rows including future-dated rows: %+v", omitted.Items[0])
	}
	if omitted.InputsAsOf == nil || *omitted.InputsAsOf != "2030-01-01" {
		t.Fatalf("omitted inputsAsOf must be maximum included tradeDate, got %v", omitted.InputsAsOf)
	}

	beforeAllRows, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "2020-01-01")
	if err != nil {
		t.Fatalf("get empty historical cutoff: %v", err)
	}
	if len(beforeAllRows.Items) != 0 || beforeAllRows.TotalAcquisitionBasis.Amount.String() != "0.00000000" ||
		beforeAllRows.InputsAsOf == nil || *beforeAllRows.InputsAsOf != "2020-01-01" {
		t.Fatalf("explicit empty historical projection mismatch: %+v", beforeAllRows)
	}
}

func TestStage372ProjectionSameDateLedgerOrderingAndBackdatedSell(t *testing.T) {
	h := newStage371Harness(t, "Stage 3.72 same-date and backdated SELL")

	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "10.00000000", "100.00000000", "2026-09-10"))
	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "SELL", "SBER", "5.00000000", "999.00000000", "2026-09-10"))
	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "5.00000000", "200.00000000", "2026-09-10"))

	sameDate, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "2026-09-10")
	if err != nil {
		t.Fatalf("get same-date projection: %v", err)
	}
	if len(sameDate.Items) != 1 || sameDate.Items[0].Quantity.String() != "10.00000000" ||
		sameDate.Items[0].WeightedAverageCost.Amount.String() != "150.00000000" ||
		sameDate.Items[0].AcquisitionBasis.Amount.String() != "1500.00000000" {
		t.Fatalf("same BusinessDate replay must follow ledger_sequence: %+v", sameDate.Items)
	}

	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "GAZP", "10.00000000", "100.00000000", "2026-09-20"))
	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "GAZP", "10.00000000", "200.00000000", "2026-09-30"))
	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "SELL", "GAZP", "5.00000000", "999.00000000", "2026-09-25"))

	backdated, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "")
	if err != nil {
		t.Fatalf("get backdated-SELL projection: %v", err)
	}
	var gazp *verticalslice.PortfolioPositionProjection
	for index := range backdated.Items {
		if backdated.Items[index].Ticker == "GAZP" {
			gazp = &backdated.Items[index]
			break
		}
	}
	if gazp == nil || gazp.Quantity.String() != "15.00000000" ||
		gazp.WeightedAverageCost.Amount.String() != "166.66666667" ||
		gazp.AcquisitionBasis.Amount.String() != "2500.00000005" {
		t.Fatalf("backdated SELL must replay by tradeDate then ledgerSequence: %+v", gazp)
	}
}

func TestStage372ProjectionEmptyCloseReopenInactiveAssetAndSubjectIsolation(t *testing.T) {
	h := newStage371Harness(t, "Stage 3.72 lifecycle and ownership")

	empty, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "")
	if err != nil {
		t.Fatalf("get empty projection: %v", err)
	}
	if len(empty.Items) != 0 || empty.TotalAcquisitionBasis.Amount.String() != "0.00000000" || empty.InputsAsOf != nil {
		t.Fatalf("empty portfolio projection mismatch: %+v", empty)
	}

	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "5.00000000", "100.00000000", "2026-09-01"))
	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "SELL", "SBER", "5.00000000", "1.00000000", "2026-09-02"))
	closed, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "")
	if err != nil {
		t.Fatalf("get fully closed projection: %v", err)
	}
	if len(closed.Items) != 0 {
		t.Fatalf("fully closed position must be omitted: %+v", closed.Items)
	}

	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "2.00000000", "333.33333333", "2026-09-03"))
	if _, err := h.db.ExecContext(h.ctx, `UPDATE investment.assets SET lifecycle_status = 'inactive' WHERE ticker = 'SBER'`); err != nil {
		t.Fatalf("mark owned catalog asset inactive: %v", err)
	}
	t.Cleanup(func() {
		_, _ = h.db.ExecContext(h.ctx, `UPDATE investment.assets SET lifecycle_status = 'active' WHERE ticker = 'SBER'`)
	})

	reopened, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "")
	if err != nil {
		t.Fatalf("get reopened inactive-catalog projection: %v", err)
	}
	if len(reopened.Items) != 1 || reopened.Items[0].Quantity.String() != "2.00000000" ||
		reopened.Items[0].WeightedAverageCost.Amount.String() != "333.33333333" ||
		reopened.Items[0].AcquisitionBasis.Amount.String() != "666.66666666" {
		t.Fatalf("close/reopen or inactive-catalog semantics drifted: %+v", reopened.Items)
	}

	_, err = h.service.GetPortfolioPositions(h.ctx, uuid.NewString(), h.portfolioID, "")
	if !errors.Is(err, postgres.ErrNotFound) {
		t.Fatalf("foreign subject must receive existing anti-enumeration not-found behavior, got %v", err)
	}
}

func TestStage372ProjectionFailsClosedOnLedgerMetadata(t *testing.T) {
	t.Run("missing ledger sequence", func(t *testing.T) {
		h := newStage371Harness(t, "Stage 3.72 invalid sequence")
		transaction := appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "1.00000000", "100.00000000", "2026-09-01"))
		if _, err := h.db.ExecContext(h.ctx, `UPDATE investment.transaction_entries SET ledger_sequence = NULL WHERE entry_id = $1`, transaction.EntryID); err != nil {
			t.Fatalf("corrupt test ledger sequence: %v", err)
		}
		_, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "")
		if !errors.Is(err, postgres.ErrLedgerSequenceUnavailable) {
			t.Fatalf("missing ledger sequence must fail closed, got %v", err)
		}
	})

	t.Run("unsupported correction revision", func(t *testing.T) {
		h := newStage371Harness(t, "Stage 3.72 unsupported correction")
		transaction := appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "1.00000000", "100.00000000", "2026-09-01"))
		if _, err := h.db.ExecContext(h.ctx, `UPDATE investment.transaction_entries SET revision = 2 WHERE entry_id = $1`, transaction.EntryID); err != nil {
			t.Fatalf("corrupt test revision: %v", err)
		}
		_, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "")
		if !errors.Is(err, postgres.ErrUnsupportedPositionLedger) {
			t.Fatalf("unsupported correction/revision state must fail closed, got %v", err)
		}
	})

	t.Run("derived basis overflow", func(t *testing.T) {
		h := newStage371Harness(t, "Stage 3.72 derived overflow")
		transaction := appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "1.00000000", "100.00000000", "2026-09-01"))
		if _, err := h.db.ExecContext(h.ctx, `
			UPDATE investment.transaction_entries
			SET quantity = 99999999999999999999.99999999, unit_price_amount = 2.00000000
			WHERE entry_id = $1
		`, transaction.EntryID); err != nil {
			t.Fatalf("corrupt test trade into derived overflow: %v", err)
		}
		_, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "")
		if !errors.Is(err, verticalslice.ErrInvalidInput) {
			t.Fatalf("derived acquisition-basis overflow must fail closed, got %v", err)
		}
	})
}

func TestStage372ProjectionConcurrentReadReturnsOnlyCoherentStates(t *testing.T) {
	h := newStage371Harness(t, "Stage 3.72 coherent read")
	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "10.00000000", "100.00000000", "2026-09-01"))

	start := make(chan struct{})
	writeDone := make(chan error, 1)
	go func() {
		<-start
		_, err := h.service.AppendTransaction(
			h.ctx,
			verticalslice.RequestContext{},
			h.subjectID,
			uuid.NewString(),
			"/api/v1/portfolios/"+h.portfolioID+"/transactions",
			stage371Trade(h.portfolioID, "BUY", "GAZP", "2.00000000", "500.00000000", "2026-09-02"),
		)
		writeDone <- err
	}()
	close(start)

	var wg sync.WaitGroup
	readErrors := make(chan error, 12)
	for index := 0; index < 12; index++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			projection, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "")
			if err != nil {
				readErrors <- err
				return
			}
			switch len(projection.Items) {
			case 1:
				if projection.Items[0].Ticker != "SBER" || projection.TotalAcquisitionBasis.Amount.String() != "1000.00000000" {
					readErrors <- errors.New("incoherent pre-write projection")
				}
			case 2:
				if projection.Items[0].Ticker != "GAZP" || projection.Items[1].Ticker != "SBER" ||
					projection.TotalAcquisitionBasis.Amount.String() != "2000.00000000" {
					readErrors <- errors.New("incoherent post-write projection")
				}
			default:
				readErrors <- errors.New("projection observed partial ledger state")
			}
		}()
	}
	wg.Wait()
	close(readErrors)
	for err := range readErrors {
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := <-writeDone; err != nil {
		t.Fatalf("concurrent accepted write failed: %v", err)
	}
}
