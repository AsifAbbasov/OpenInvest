package postgres_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func stage377AppendCash(t *testing.T, h stage371Harness, transactionType, amount, tradeDate string) verticalslice.Transaction {
	t.Helper()
	return appendStage371Trade(t, h, stage375Request(
		h.portfolioID,
		transactionType,
		"",
		amount,
		"0.00000000",
		"0.00000000",
		tradeDate,
	))
}

func stage377CorrectCash(
	t *testing.T,
	h stage371Harness,
	transaction verticalslice.Transaction,
	amount string,
	tradeDate string,
) verticalslice.Transaction {
	t.Helper()
	requestContext := verticalslice.RequestContext{RequestID: uuid.NewString(), TraceID: "stage-03-77-correction"}
	correctedRequest := stage375Request(
		h.portfolioID,
		transaction.TransactionType,
		"",
		amount,
		"0.00000000",
		"0.00000000",
		tradeDate,
	)
	corrected, _, err := h.service.CorrectTransactionWithReplay(
		h.ctx,
		requestContext,
		h.subjectID,
		uuid.NewString(),
		"/api/v1/portfolios/"+h.portfolioID+"/transactions/"+transaction.ID,
		verticalslice.CorrectTransactionRequest{
			PortfolioID:      h.portfolioID,
			TransactionID:    transaction.ID,
			ExpectedRevision: transaction.Revision,
			Reason:           "Correct external cash flow",
			Corrected:        correctedRequest,
		},
		func(result verticalslice.Transaction) (verticalslice.CommandReplayArtifact, error) {
			return stage374Artifact(requestContext, result)
		},
	)
	if err != nil {
		t.Fatalf("correct Stage 3.77 cash flow: %v", err)
	}
	return corrected
}

func TestStage377CashOnlyReturnExcludesInternalIncomeAndMirrorsSummary(t *testing.T) {
	h := newStage371Harness(t, "Stage 3.77 cash-only return")
	stage377AppendCash(t, h, "DEPOSIT", "100.00000000", "2025-01-01")
	appendStage371Trade(t, h, stage375Request(
		h.portfolioID,
		"DIVIDEND",
		"SBER",
		"10.00000000",
		"0.00000000",
		"0.00000000",
		"2026-01-01",
	))

	projection, err := h.service.GetPortfolioReturns(h.ctx, h.subjectID, h.portfolioID, "2026-01-01")
	if err != nil {
		t.Fatalf("get Stage 3.77 returns: %v", err)
	}
	if projection.Status != verticalslice.PortfolioReturnAvailableStatus || projection.XIRR == nil ||
		projection.XIRR.String() != "0.10000000" || projection.TerminalPortfolioValue == nil ||
		projection.TerminalPortfolioValue.Amount.String() != "110.00000000" {
		t.Fatalf("cash-only XIRR mismatch: %+v", projection)
	}
	if len(projection.ExternalCashFlows) != 1 || projection.ExternalCashFlows[0].Date != "2025-01-01" ||
		projection.ExternalCashFlows[0].Amount.String() != "-100.00000000" {
		t.Fatalf("DIVIDEND must not become an external XIRR flow: %+v", projection.ExternalCashFlows)
	}

	summary, err := h.service.GetPortfolioSummary(h.ctx, h.subjectID, h.portfolioID, "2026-01-01")
	if err != nil {
		t.Fatalf("get Stage 3.77 summary mirror: %v", err)
	}
	if summary.XIRR == nil || summary.XIRR.String() != projection.XIRR.String() {
		t.Fatalf("summary must mirror canonical XIRR: summary=%v returns=%v", summary.XIRR, projection.XIRR)
	}

	_, err = h.service.GetPortfolioReturns(h.ctx, uuid.NewString(), h.portfolioID, "2026-01-01")
	if !errors.Is(err, postgres.ErrNotFound) {
		t.Fatalf("foreign subject must not read portfolio returns, err=%v", err)
	}
}

func TestStage377ExactDateManualValuationRequiredForOpenPositions(t *testing.T) {
	h := newStage371Harness(t, "Stage 3.77 exact valuation")
	cleanupStage376Valuations(t, h)
	stage377AppendCash(t, h, "DEPOSIT", "1000.00000000", "2025-01-01")
	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "10.00000000", "50.00000000", "2025-01-01"))

	incomplete, err := h.service.GetPortfolioReturns(h.ctx, h.subjectID, h.portfolioID, "2026-01-01")
	if err != nil {
		t.Fatalf("read incomplete Stage 3.77 valuation: %v", err)
	}
	if incomplete.Status != verticalslice.PortfolioReturnUnavailableStatus ||
		incomplete.Reason != verticalslice.PortfolioReturnReasonIncompleteValuation || incomplete.XIRR != nil {
		t.Fatalf("missing manual valuation must fail closed: %+v", incomplete)
	}

	upsertStage376(t, h, "SBER", "60.00000000", "2026-01-01")
	available, err := h.service.GetPortfolioReturns(h.ctx, h.subjectID, h.portfolioID, "2026-01-01")
	if err != nil {
		t.Fatalf("read exact-date Stage 3.77 valuation: %v", err)
	}
	if available.Status != verticalslice.PortfolioReturnAvailableStatus || available.XIRR == nil ||
		available.XIRR.String() != "0.10000000" || available.TerminalPortfolioValue == nil ||
		available.TerminalPortfolioValue.Amount.String() != "1100.00000000" {
		t.Fatalf("exact-date valuation XIRR mismatch: %+v", available)
	}

	mismatch, err := h.service.GetPortfolioReturns(h.ctx, h.subjectID, h.portfolioID, "2025-12-31")
	if err != nil {
		t.Fatalf("read mismatched-date Stage 3.77 valuation: %v", err)
	}
	if mismatch.Reason != verticalslice.PortfolioReturnReasonIncompleteValuation || mismatch.XIRR != nil {
		t.Fatalf("manual valuation must not carry across BusinessDate: %+v", mismatch)
	}
}

func TestStage377EffectiveLedgerCorrectionAndReversal(t *testing.T) {
	h := newStage371Harness(t, "Stage 3.77 effective ledger")
	deposit := stage377AppendCash(t, h, "DEPOSIT", "100.00000000", "2025-01-01")
	appendStage371Trade(t, h, stage375Request(
		h.portfolioID,
		"DIVIDEND",
		"SBER",
		"10.00000000",
		"0.00000000",
		"0.00000000",
		"2026-01-01",
	))

	corrected := stage377CorrectCash(t, h, deposit, "200.00000000", "2025-01-01")
	afterCorrection, err := h.service.GetPortfolioReturns(h.ctx, h.subjectID, h.portfolioID, "2026-01-01")
	if err != nil {
		t.Fatalf("returns after correction: %v", err)
	}
	if afterCorrection.XIRR == nil || afterCorrection.XIRR.String() != "0.05000000" ||
		len(afterCorrection.ExternalCashFlows) != 1 || afterCorrection.ExternalCashFlows[0].Amount.String() != "-200.00000000" {
		t.Fatalf("corrected effective ledger not reflected in XIRR: %+v", afterCorrection)
	}

	if _, err := reverseStage374(t, h, corrected, "2026-01-02", uuid.NewString()); err != nil {
		t.Fatalf("reverse corrected deposit: %v", err)
	}
	beforeReversal, err := h.service.GetPortfolioReturns(h.ctx, h.subjectID, h.portfolioID, "2026-01-01")
	if err != nil || beforeReversal.XIRR == nil || beforeReversal.XIRR.String() != "0.05000000" {
		t.Fatalf("historical result before reversal effective date drifted: projection=%+v err=%v", beforeReversal, err)
	}
	afterReversal, err := h.service.GetPortfolioReturns(h.ctx, h.subjectID, h.portfolioID, "2026-01-02")
	if err != nil {
		t.Fatalf("returns on reversal effective date: %v", err)
	}
	if afterReversal.Status != verticalslice.PortfolioReturnUnavailableStatus ||
		afterReversal.Reason != verticalslice.PortfolioReturnReasonNoExternalContributions || afterReversal.XIRR != nil {
		t.Fatalf("reversal must remove contribution on effective date: %+v", afterReversal)
	}
}

func TestStage377FutureExternalFlowAndInternalLedgerTypesStayOutOfXIRREvidence(t *testing.T) {
	h := newStage371Harness(t, "Stage 3.77 external-flow boundary")
	stage377AppendCash(t, h, "DEPOSIT", "1000.00000000", "2025-01-01")
	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "1.00000000", "100.00000000", "2025-02-01"))
	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "SELL", "SBER", "1.00000000", "110.00000000", "2025-03-01"))
	appendStage371Trade(t, h, stage375Request(h.portfolioID, "DIVIDEND", "SBER", "20.00000000", "0.00000000", "0.00000000", "2025-04-01"))
	appendStage371Trade(t, h, stage375Request(h.portfolioID, "COUPON", "SU26238RMFS4", "30.00000000", "0.00000000", "0.00000000", "2025-05-01"))
	appendStage371Trade(t, h, stage375Request(h.portfolioID, "FEE", "", "5.00000000", "0.00000000", "0.00000000", "2025-06-01"))
	appendStage371Trade(t, h, stage375Request(h.portfolioID, "TAX", "", "3.00000000", "0.00000000", "0.00000000", "2025-07-01"))
	stage377AppendCash(t, h, "WITHDRAWAL", "500.00000000", "2026-01-02")

	projection, err := h.service.GetPortfolioReturns(h.ctx, h.subjectID, h.portfolioID, "2026-01-01")
	if err != nil {
		t.Fatalf("get Stage 3.77 external-flow boundary: %v", err)
	}
	if len(projection.ExternalCashFlows) != 1 || projection.ExternalCashFlows[0].Date != "2025-01-01" ||
		projection.ExternalCashFlows[0].Amount.String() != "-1000.00000000" {
		t.Fatalf("BUY/SELL/DIVIDEND/COUPON/FEE/TAX and future WITHDRAWAL must stay out of XIRR evidence: %+v", projection.ExternalCashFlows)
	}
	if projection.Status != verticalslice.PortfolioReturnAvailableStatus || projection.XIRR == nil {
		t.Fatalf("cash-only terminal value should remain calculable before future withdrawal: %+v", projection)
	}
}

func TestStage377PartialCoverageAndReopenedGenerationFailClosed(t *testing.T) {
	t.Run("partial coverage", func(t *testing.T) {
		h := newStage371Harness(t, "Stage 3.77 partial coverage")
		cleanupStage376Valuations(t, h)
		stage377AppendCash(t, h, "DEPOSIT", "2000.00000000", "2025-01-01")
		appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "1.00000000", "100.00000000", "2025-01-02"))
		appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "GAZP", "1.00000000", "100.00000000", "2025-01-02"))
		upsertStage376(t, h, "SBER", "120.00000000", "2026-01-01")

		projection, err := h.service.GetPortfolioReturns(h.ctx, h.subjectID, h.portfolioID, "2026-01-01")
		if err != nil {
			t.Fatalf("get Stage 3.77 partial valuation: %v", err)
		}
		if projection.Status != verticalslice.PortfolioReturnUnavailableStatus ||
			projection.Reason != verticalslice.PortfolioReturnReasonIncompleteValuation || projection.XIRR != nil {
			t.Fatalf("partial manual valuation coverage must fail closed: %+v", projection)
		}
	})

	t.Run("stale generation", func(t *testing.T) {
		h := newStage371Harness(t, "Stage 3.77 stale valuation generation")
		cleanupStage376Valuations(t, h)
		stage377AppendCash(t, h, "DEPOSIT", "2000.00000000", "2026-09-01")
		appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "10.00000000", "100.00000000", "2026-09-01"))
		// Stage 3.76 deliberately permits a future BusinessDate. Store it while the old
		// generation is open, then close/reopen so date equality alone cannot reactivate it.
		upsertStage376(t, h, "SBER", "120.00000000", "2026-09-12")
		appendStage371Trade(t, h, stage371Trade(h.portfolioID, "SELL", "SBER", "10.00000000", "125.00000000", "2026-09-11"))
		appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "2.00000000", "200.00000000", "2026-09-12"))

		projection, err := h.service.GetPortfolioReturns(h.ctx, h.subjectID, h.portfolioID, "2026-09-12")
		if err != nil {
			t.Fatalf("get Stage 3.77 reopened-generation valuation: %v", err)
		}
		if projection.Status != verticalslice.PortfolioReturnUnavailableStatus ||
			projection.Reason != verticalslice.PortfolioReturnReasonIncompleteValuation || projection.XIRR != nil {
			t.Fatalf("stale pre-close valuation generation must not value reopened position: %+v", projection)
		}
	})
}

func TestStage377PostgresMultipleRootVectorFailsClosed(t *testing.T) {
	h := newStage371Harness(t, "Stage 3.77 multiple roots")
	stage377AppendCash(t, h, "DEPOSIT", "100.00000000", "2025-01-01")
	stage377AppendCash(t, h, "WITHDRAWAL", "230.00000000", "2026-01-01")
	stage377AppendCash(t, h, "DEPOSIT", "200.00000000", "2027-01-01")
	appendStage371Trade(t, h, stage375Request(
		h.portfolioID,
		"FEE",
		"",
		"2.00000000",
		"0.00000000",
		"0.00000000",
		"2027-01-01",
	))

	projection, err := h.service.GetPortfolioReturns(h.ctx, h.subjectID, h.portfolioID, "2027-01-01")
	if err != nil {
		t.Fatalf("get Stage 3.77 multiple-root projection: %v", err)
	}
	if projection.Status != verticalslice.PortfolioReturnUnavailableStatus ||
		projection.Reason != verticalslice.PortfolioReturnReasonAmbiguousMultipleRoots || projection.XIRR != nil ||
		projection.TerminalPortfolioValue == nil || projection.TerminalPortfolioValue.Amount.String() != "68.00000000" {
		t.Fatalf("multiple-root vector must fail closed: %+v", projection)
	}
	if len(projection.ExternalCashFlows) != 3 || projection.ExternalCashFlows[0].Amount.String() != "-100.00000000" ||
		projection.ExternalCashFlows[1].Amount.String() != "230.00000000" || projection.ExternalCashFlows[2].Amount.String() != "-200.00000000" {
		t.Fatalf("FEE must affect terminal cash but not external-flow evidence: %+v", projection.ExternalCashFlows)
	}
}
