package postgres_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func stage375Request(portfolioID, transactionType, ticker, gross, commission, tax, tradeDate string) verticalslice.AppendTransactionRequest {
	request := verticalslice.AppendTransactionRequest{
		PortfolioID:     portfolioID,
		TransactionType: transactionType,
		GrossAmount:     moneyPtr375(gross),
		Commission:      verticalslice.Money{Amount: decimal.Must(commission), Currency: verticalslice.RUB},
		Tax:             verticalslice.Money{Amount: decimal.Must(tax), Currency: verticalslice.RUB},
		TradeDate:       tradeDate,
	}
	if ticker != "" {
		request.Ticker = &ticker
	}
	return request
}

func moneyPtr375(amount string) *verticalslice.Money {
	value := verticalslice.Money{Amount: decimal.Must(amount), Currency: verticalslice.RUB}
	return &value
}

func assertMoney375(t *testing.T, got verticalslice.Money, want string, label string) {
	t.Helper()
	if got.Currency != verticalslice.RUB || got.Amount.String() != want {
		t.Fatalf("%s got %s %s want %s RUB", label, got.Amount.String(), got.Currency, want)
	}
}

func TestStage375MixedLedgerCashFlowMonthlySummaryAndRange(t *testing.T) {
	h := newStage371Harness(t, "Stage 3.75 mixed cash-flow truth")

	appendStage371Trade(t, h, stage375Request(h.portfolioID, "DEPOSIT", "", "100000.00000000", "0.00000000", "0.00000000", "2026-01-05"))
	buy := stage371Trade(h.portfolioID, "BUY", "SBER", "10.00000000", "8000.00000000", "2026-01-10")
	buy.Commission = verticalslice.Money{Amount: decimal.Must("100.00000000"), Currency: verticalslice.RUB}
	appendStage371Trade(t, h, buy)
	dividend := appendStage371Trade(t, h, stage375Request(h.portfolioID, "DIVIDEND", "SBER", "3000.00000000", "10.00000000", "390.00000000", "2026-02-15"))
	appendStage371Trade(t, h, stage375Request(h.portfolioID, "COUPON", "SU26238RMFS4", "1200.00000000", "0.00000000", "156.00000000", "2026-02-20"))
	appendStage371Trade(t, h, stage375Request(h.portfolioID, "FEE", "", "100.00000000", "0.00000000", "0.00000000", "2026-03-01"))
	appendStage371Trade(t, h, stage375Request(h.portfolioID, "TAX", "", "50.00000000", "0.00000000", "0.00000000", "2026-03-02"))
	sell := stage371Trade(h.portfolioID, "SELL", "SBER", "1.00000000", "9000.00000000", "2026-03-10")
	sell.Commission = verticalslice.Money{Amount: decimal.Must("20.00000000"), Currency: verticalslice.RUB}
	appendStage371Trade(t, h, sell)
	appendStage371Trade(t, h, stage375Request(h.portfolioID, "WITHDRAWAL", "", "5000.00000000", "0.00000000", "0.00000000", "2026-03-15"))

	projection, err := h.service.GetPortfolioCashFlow(h.ctx, h.subjectID, h.portfolioID, "", "")
	if err != nil {
		t.Fatalf("get mixed cash-flow projection: %v", err)
	}
	if projection.MethodologyVersion != verticalslice.PortfolioCashFlowMethodologyVersion || len(projection.Periods) != 3 {
		t.Fatalf("cash-flow metadata/periods mismatch: %+v", projection)
	}
	assertMoney375(t, projection.Totals.Deposits, "100000.00000000", "deposits")
	assertMoney375(t, projection.Totals.Withdrawals, "5000.00000000", "withdrawals")
	assertMoney375(t, projection.Totals.BuyOutflows, "80000.00000000", "buy outflows")
	assertMoney375(t, projection.Totals.SellInflows, "9000.00000000", "sell inflows")
	assertMoney375(t, projection.Totals.DividendsGross, "3000.00000000", "dividends")
	assertMoney375(t, projection.Totals.CouponsGross, "1200.00000000", "coupons")
	assertMoney375(t, projection.Totals.Fees, "230.00000000", "fees")
	assertMoney375(t, projection.Totals.Taxes, "596.00000000", "taxes")
	assertMoney375(t, projection.Totals.NetExternalFlow, "95000.00000000", "net external flow")
	assertMoney375(t, projection.Totals.NetInvestmentIncome, "3374.00000000", "net investment income")
	assertMoney375(t, projection.Totals.NetCashFlow, "27374.00000000", "net cash flow")
	if projection.Periods[0].Month != "2026-01" || projection.Periods[1].Month != "2026-02" || projection.Periods[2].Month != "2026-03" {
		t.Fatalf("monthly ordering/grouping drifted: %+v", projection.Periods)
	}

	february, err := h.service.GetPortfolioCashFlow(h.ctx, h.subjectID, h.portfolioID, "2026-02-01", "2026-02-28")
	if err != nil {
		t.Fatalf("get February range: %v", err)
	}
	assertMoney375(t, february.Totals.DividendsGross, "3000.00000000", "February dividends")
	assertMoney375(t, february.Totals.CouponsGross, "1200.00000000", "February coupons")
	assertMoney375(t, february.Totals.Fees, "10.00000000", "February fees")
	assertMoney375(t, february.Totals.Taxes, "546.00000000", "February taxes")
	assertMoney375(t, february.Totals.NetCashFlow, "3644.00000000", "February net cash")
	if len(february.Periods) != 1 || february.Periods[0].Month != "2026-02" || february.InputsAsOf == nil || *february.InputsAsOf != "2026-02-28" {
		t.Fatalf("explicit range metadata mismatch: %+v", february)
	}

	summary, err := h.service.GetPortfolioSummary(h.ctx, h.subjectID, h.portfolioID, "")
	if err != nil {
		t.Fatalf("get Stage 3.75 summary truth: %v", err)
	}
	assertMoney375(t, summary.DividendsReceived, "3000.00000000", "summary dividends")
	assertMoney375(t, summary.CouponsReceived, "1200.00000000", "summary coupons")
	assertMoney375(t, summary.CashValue, "27374.00000000", "summary cash")

	correctedRequest := stage375Request(h.portfolioID, "DIVIDEND", "SBER", "2500.00000000", "10.00000000", "390.00000000", "2026-02-15")
	requestContext := verticalslice.RequestContext{RequestID: uuid.NewString(), TraceID: "stage-03-75-correction"}
	corrected, _, err := h.service.CorrectTransactionWithReplay(
		h.ctx,
		requestContext,
		h.subjectID,
		uuid.NewString(),
		"/api/v1/portfolios/"+h.portfolioID+"/transactions/"+dividend.ID,
		verticalslice.CorrectTransactionRequest{
			PortfolioID:      h.portfolioID,
			TransactionID:    dividend.ID,
			ExpectedRevision: dividend.Revision,
			Reason:           "Correct recorded dividend",
			Corrected:        correctedRequest,
		},
		func(result verticalslice.Transaction) (verticalslice.CommandReplayArtifact, error) {
			return stage374Artifact(requestContext, result)
		},
	)
	if err != nil || corrected.Revision != 2 || corrected.Status != "CORRECTED" {
		t.Fatalf("correct dividend: result=%+v err=%v", corrected, err)
	}
	afterCorrection, err := h.service.GetPortfolioCashFlow(h.ctx, h.subjectID, h.portfolioID, "", "")
	if err != nil {
		t.Fatalf("cash flow after correction: %v", err)
	}
	assertMoney375(t, afterCorrection.Totals.DividendsGross, "2500.00000000", "corrected dividends")
	assertMoney375(t, afterCorrection.Totals.NetCashFlow, "26874.00000000", "cash after dividend correction")
}

func TestStage375ReversalDateBackdatedRangeAndIsolation(t *testing.T) {
	h := newStage371Harness(t, "Stage 3.75 reversal")
	coupon := appendStage371Trade(t, h, stage375Request(h.portfolioID, "COUPON", "SU26238RMFS4", "1000.00000000", "0.00000000", "130.00000000", "2026-02-10"))
	if _, err := reverseStage374(t, h, coupon, "2026-04-01", uuid.NewString()); err != nil {
		t.Fatalf("reverse coupon: %v", err)
	}

	before, err := h.service.GetPortfolioCashFlow(h.ctx, h.subjectID, h.portfolioID, "2026-01-01", "2026-03-31")
	if err != nil {
		t.Fatalf("cash flow before reversal effective date: %v", err)
	}
	assertMoney375(t, before.Totals.CouponsGross, "1000.00000000", "historical coupon before reversal")
	assertMoney375(t, before.Totals.NetCashFlow, "870.00000000", "historical coupon net cash")

	after, err := h.service.GetPortfolioCashFlow(h.ctx, h.subjectID, h.portfolioID, "2026-01-01", "2026-04-01")
	if err != nil {
		t.Fatalf("cash flow on reversal effective date: %v", err)
	}
	assertMoney375(t, after.Totals.CouponsGross, "0.00000000", "coupon after reversal")
	if len(after.Periods) != 0 {
		t.Fatalf("reversed economic effect must disappear from effective periods: %+v", after.Periods)
	}

	backdated := appendStage371Trade(t, h, stage375Request(h.portfolioID, "DIVIDEND", "SBER", "333.33333333", "0.00000000", "0.00000000", "2026-03-15"))
	_ = backdated
	march, err := h.service.GetPortfolioCashFlow(h.ctx, h.subjectID, h.portfolioID, "2026-03-01", "2026-03-31")
	if err != nil {
		t.Fatalf("get backdated March income: %v", err)
	}
	if len(march.Periods) != 1 || march.Periods[0].Month != "2026-03" {
		t.Fatalf("backdated event must group by tradeDate, got %+v", march.Periods)
	}
	assertMoney375(t, march.Totals.DividendsGross, "333.33333333", "backdated March dividend")

	_, err = h.service.GetPortfolioCashFlow(h.ctx, uuid.NewString(), h.portfolioID, "", "")
	if !errors.Is(err, postgres.ErrNotFound) {
		t.Fatalf("foreign subject must preserve anti-enumeration not-found, got %v", err)
	}

	empty, err := h.service.GetPortfolioCashFlow(h.ctx, h.subjectID, h.portfolioID, "2035-01-01", "2035-12-31")
	if err != nil {
		t.Fatalf("get empty range: %v", err)
	}
	if len(empty.Periods) != 0 || empty.ToDate == nil || *empty.ToDate != "2035-12-31" {
		t.Fatalf("empty range mismatch: %+v", empty)
	}
	assertMoney375(t, empty.Totals.NetCashFlow, "0.00000000", "empty range net cash")
}
