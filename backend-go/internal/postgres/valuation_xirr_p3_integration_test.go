package postgres_test

import (
	"testing"

	"github.com/google/uuid"
)

type valuationXIRRSnapshot struct {
	total       string
	cash        string
	stock       string
	bond        string
	invested    string
	nominal     string
	real        string
	legacyState string
	methodology string
}

func latestValuationXIRRSnapshot(t *testing.T, h stage371Harness, asOfDate string) valuationXIRRSnapshot {
	t.Helper()
	var got valuationXIRRSnapshot
	err := h.db.QueryRowContext(h.ctx, `
		SELECT total_value_amount::text, cash_value_amount::text, stock_value_amount::text,
			bond_value_amount::text, invested_capital_amount::text, nominal_return_rate::text,
			real_return_rate::text, legacy_return_status, methodology_version
		FROM analytics.portfolio_snapshots
		WHERE portfolio_id = $1::uuid
			AND snapshot_date = $2::date
		ORDER BY
			CASE methodology_version
				WHEN 'stage-03-71-position-cost-snapshot-v2' THEN 2
				WHEN 'stage-03-71-position-cost-snapshot-v1' THEN 1
				ELSE 0
			END DESC,
			snapshot_version DESC, calculated_at DESC, id DESC
		LIMIT 1
	`, h.portfolioID, asOfDate).Scan(
		&got.total, &got.cash, &got.stock, &got.bond, &got.invested, &got.nominal,
		&got.real, &got.legacyState, &got.methodology,
	)
	if err != nil {
		t.Fatalf("read valuation/XIRR snapshot: %v", err)
	}
	return got
}

func assertValuationXIRRUnavailable(t *testing.T, got valuationXIRRSnapshot) {
	t.Helper()
	if got.legacyState != "UNAVAILABLE" || got.nominal != "0.00000000" || got.real != "0.00000000" {
		t.Fatalf("legacy ROI must be an unavailable non-economic placeholder: %+v", got)
	}
	if got.methodology != stage371SnapshotMethodologyTest {
		t.Fatalf("new snapshot methodology got %q want %q", got.methodology, stage371SnapshotMethodologyTest)
	}
}

func assertValuationXIRRAmounts(t *testing.T, got valuationXIRRSnapshot, total, cash, stock, invested string) {
	t.Helper()
	if got.total != total || got.cash != cash || got.stock != stock || got.invested != invested {
		t.Fatalf("snapshot amount mismatch: got=%+v want total=%s cash=%s stock=%s invested=%s", got, total, cash, stock, invested)
	}
	assertValuationXIRRUnavailable(t, got)
}

func TestValuationXIRRP3LegacyReturnPersistenceIsUnavailable(t *testing.T) {
	tests := []struct {
		name  string
		build func(t *testing.T, h stage371Harness)
		date  string
		total string
		cash  string
		stock string
		inv   string
	}{
		{
			name: "cash only",
			build: func(t *testing.T, h stage371Harness) {
				stage377AppendCash(t, h, "DEPOSIT", "1000.00000000", "2026-01-01")
			},
			date: "2026-01-01", total: "1000.00000000", cash: "1000.00000000", stock: "0.00000000", inv: "0.00000000",
		},
		{
			name: "idle cash",
			build: func(t *testing.T, h stage371Harness) {
				stage377AppendCash(t, h, "DEPOSIT", "1000000.00000000", "2026-01-01")
				appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "1.00000000", "1.00000000", "2026-01-01"))
			},
			date: "2026-01-01", total: "1000000.00000000", cash: "999999.00000000", stock: "1.00000000", inv: "1.00000000",
		},
		{
			name: "fully invested neutral",
			build: func(t *testing.T, h stage371Harness) {
				stage377AppendCash(t, h, "DEPOSIT", "1000.00000000", "2026-01-01")
				appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "1.00000000", "1000.00000000", "2026-01-01"))
			},
			date: "2026-01-01", total: "1000.00000000", cash: "0.00000000", stock: "1000.00000000", inv: "1000.00000000",
		},
		{
			name: "partially invested neutral",
			build: func(t *testing.T, h stage371Harness) {
				stage377AppendCash(t, h, "DEPOSIT", "1000.00000000", "2026-01-01")
				appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "1.00000000", "500.00000000", "2026-01-01"))
			},
			date: "2026-01-01", total: "1000.00000000", cash: "500.00000000", stock: "500.00000000", inv: "500.00000000",
		},
		{
			name: "economic profit",
			build: func(t *testing.T, h stage371Harness) {
				stage377AppendCash(t, h, "DEPOSIT", "1000.00000000", "2026-01-01")
				appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "1.00000000", "500.00000000", "2026-01-01"))
				appendStage371Trade(t, h, stage371Trade(h.portfolioID, "SELL", "SBER", "1.00000000", "600.00000000", "2026-01-02"))
			},
			date: "2026-01-02", total: "1100.00000000", cash: "1100.00000000", stock: "0.00000000", inv: "500.00000000",
		},
		{
			name: "economic loss",
			build: func(t *testing.T, h stage371Harness) {
				stage377AppendCash(t, h, "DEPOSIT", "1000.00000000", "2026-01-01")
				appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "1.00000000", "500.00000000", "2026-01-01"))
				appendStage371Trade(t, h, stage371Trade(h.portfolioID, "SELL", "SBER", "1.00000000", "400.00000000", "2026-01-02"))
			},
			date: "2026-01-02", total: "900.00000000", cash: "900.00000000", stock: "0.00000000", inv: "500.00000000",
		},
		{
			name: "withdrawal after deposit",
			build: func(t *testing.T, h stage371Harness) {
				stage377AppendCash(t, h, "DEPOSIT", "1000.00000000", "2026-01-01")
				stage377AppendCash(t, h, "WITHDRAWAL", "250.00000000", "2026-01-02")
			},
			date: "2026-01-02", total: "750.00000000", cash: "750.00000000", stock: "0.00000000", inv: "0.00000000",
		},
		{
			name: "deposit withdrawal partial investment",
			build: func(t *testing.T, h stage371Harness) {
				stage377AppendCash(t, h, "DEPOSIT", "1000.00000000", "2026-01-01")
				stage377AppendCash(t, h, "WITHDRAWAL", "100.00000000", "2026-01-02")
				appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "1.00000000", "500.00000000", "2026-01-03"))
			},
			date: "2026-01-03", total: "900.00000000", cash: "400.00000000", stock: "500.00000000", inv: "500.00000000",
		},
		{
			name: "deposit correction",
			build: func(t *testing.T, h stage371Harness) {
				deposit := stage377AppendCash(t, h, "DEPOSIT", "1000.00000000", "2026-01-01")
				stage377CorrectCash(t, h, deposit, "800.00000000", "2026-01-01")
			},
			date: "2026-01-01", total: "800.00000000", cash: "800.00000000", stock: "0.00000000", inv: "0.00000000",
		},
		{
			name: "buy correction",
			build: func(t *testing.T, h stage371Harness) {
				stage377AppendCash(t, h, "DEPOSIT", "1000.00000000", "2026-01-01")
				buy := appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "1.00000000", "100.00000000", "2026-01-01"))
				if _, err := correctStage374(t, h, buy, "1.00000000", "200.00000000", "2026-01-01", uuid.NewString()); err != nil {
					t.Fatalf("correct BUY: %v", err)
				}
			},
			date: "2026-01-01", total: "1000.00000000", cash: "800.00000000", stock: "200.00000000", inv: "200.00000000",
		},
		{
			name: "deposit reversal",
			build: func(t *testing.T, h stage371Harness) {
				deposit := stage377AppendCash(t, h, "DEPOSIT", "1000.00000000", "2026-01-01")
				if _, err := reverseStage374(t, h, deposit, "2026-01-02", uuid.NewString()); err != nil {
					t.Fatalf("reverse deposit: %v", err)
				}
			},
			date: "2026-01-02", total: "0.00000000", cash: "0.00000000", stock: "0.00000000", inv: "0.00000000",
		},
		{
			name: "buy reversal",
			build: func(t *testing.T, h stage371Harness) {
				stage377AppendCash(t, h, "DEPOSIT", "1000.00000000", "2026-01-01")
				buy := appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "1.00000000", "1000.00000000", "2026-01-01"))
				if _, err := reverseStage374(t, h, buy, "2026-01-02", uuid.NewString()); err != nil {
					t.Fatalf("reverse BUY: %v", err)
				}
			},
			date: "2026-01-02", total: "1000.00000000", cash: "1000.00000000", stock: "0.00000000", inv: "0.00000000",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			h := newStage371Harness(t, "valuation xirr p3 "+test.name)
			test.build(t, h)
			assertValuationXIRRAmounts(t, latestValuationXIRRSnapshot(t, h, test.date), test.total, test.cash, test.stock, test.inv)
		})
	}
}

func TestValuationXIRRP3SummaryAndReturnsRemainSeparate(t *testing.T) {
	h := newStage371Harness(t, "valuation xirr p3 basis witness")
	stage377AppendCash(t, h, "DEPOSIT", "100.00000000", "2025-01-01")
	appendStage371Trade(t, h, stage371Trade(h.portfolioID, "BUY", "SBER", "1.00000000", "100.00000000", "2025-01-01"))
	cleanupStage376Valuations(t, h)
	upsertStage376(t, h, "SBER", "125.00000000", "2026-01-01")

	summary, err := h.service.GetPortfolioSummary(h.ctx, h.subjectID, h.portfolioID, "2026-01-01")
	if err != nil {
		t.Fatalf("get local-basis summary: %v", err)
	}
	if summary.TotalValue.Amount.String() != "100.00000000" {
		t.Fatalf("summary must remain local acquisition-cost basis, got %s", summary.TotalValue.Amount.String())
	}
	returns, err := h.service.GetPortfolioReturns(h.ctx, h.subjectID, h.portfolioID, "2026-01-01")
	if err != nil {
		t.Fatalf("get terminal-market returns: %v", err)
	}
	if returns.TerminalPortfolioValue == nil || returns.TerminalPortfolioValue.Amount.String() != "125.00000000" {
		t.Fatalf("XIRR terminal value must remain exact-date market value, got %+v", returns)
	}
	assertValuationXIRRUnavailable(t, latestValuationXIRRSnapshot(t, h, "2025-01-01"))
}
