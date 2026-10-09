package postgres_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestSecurityAuthorizationMatrixRejectsCrossPrincipalPortfolioAccess(t *testing.T) {
	h := newStage371Harness(t, "security authorization matrix")
	original := appendStage371Trade(t, h, stage371Trade(
		h.portfolioID,
		"BUY",
		"SBER",
		"10.00000000",
		"100.00000000",
		"2026-01-02",
	))

	attackerSubject := uuid.NewString()

	type readCase struct {
		name string
		call func() error
	}
	reads := []readCase{
		{
			name: "get portfolio",
			call: func() error {
				_, err := h.service.GetPortfolio(h.ctx, attackerSubject, h.portfolioID)
				return err
			},
		},
		{
			name: "summary",
			call: func() error {
				_, err := h.service.GetPortfolioSummary(h.ctx, attackerSubject, h.portfolioID, "")
				return err
			},
		},
		{
			name: "positions",
			call: func() error {
				_, err := h.service.GetPortfolioPositions(h.ctx, attackerSubject, h.portfolioID, "")
				return err
			},
		},
		{
			name: "transactions",
			call: func() error {
				_, err := h.service.ListTransactions(
					h.ctx,
					attackerSubject,
					h.portfolioID,
					verticalslice.TransactionFilter{Limit: 10},
				)
				return err
			},
		},
		{
			name: "cash flow",
			call: func() error {
				_, err := h.service.GetPortfolioCashFlow(h.ctx, attackerSubject, h.portfolioID, "", "")
				return err
			},
		},
		{
			name: "returns",
			call: func() error {
				_, err := h.service.GetPortfolioReturns(h.ctx, attackerSubject, h.portfolioID, "2026-01-02")
				return err
			},
		},
	}

	for _, tc := range reads {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.call(); err == nil {
				t.Fatalf("cross-principal read unexpectedly succeeded: %s", tc.name)
			}
		})
	}

	var ledgerBefore int
	if err := h.db.QueryRowContext(
		h.ctx,
		"SELECT COUNT(*) FROM investment.transaction_entries WHERE portfolio_id=$1",
		h.portfolioID,
	).Scan(&ledgerBefore); err != nil {
		t.Fatal(err)
	}

	_, err := h.service.AppendTransaction(
		h.ctx,
		verticalslice.RequestContext{RequestID: uuid.NewString()},
		attackerSubject,
		uuid.NewString(),
		"/api/v1/portfolios/"+h.portfolioID+"/transactions",
		stage371Trade(
			h.portfolioID,
			"BUY",
			"SBER",
			"1.00000000",
			"1.00000000",
			"2026-01-03",
		),
	)
	if err == nil {
		t.Fatal("cross-principal transaction mutation unexpectedly succeeded")
	}

	_, err = h.service.UpsertManualValuation(
		h.ctx,
		attackerSubject,
		verticalslice.ManualValuationRequest{
			PortfolioID: h.portfolioID,
			Ticker:      "SBER",
			Price: verticalslice.Money{
				Amount:   decimal.Must("123.00000000"),
				Currency: verticalslice.RUB,
			},
			AsOfDate: "2026-01-02",
		},
	)
	if err == nil {
		t.Fatal("cross-principal manual valuation mutation unexpectedly succeeded")
	}

	_, err = h.service.ClearManualValuation(h.ctx, attackerSubject, h.portfolioID, "SBER")
	if err == nil {
		t.Fatal("cross-principal manual valuation clear unexpectedly succeeded")
	}

	corrected := stage371Trade(h.portfolioID, "BUY", "SBER", "10.00000000", "101.00000000", "2026-01-02")
	_, _, err = h.service.CorrectTransactionWithReplay(
		h.ctx,
		verticalslice.RequestContext{RequestID: uuid.NewString()},
		attackerSubject,
		uuid.NewString(),
		"/api/v1/portfolios/"+h.portfolioID+"/transactions/"+original.ID,
		verticalslice.CorrectTransactionRequest{
			PortfolioID:      h.portfolioID,
			TransactionID:    original.ID,
			ExpectedRevision: original.Revision,
			Reason:           "authorization matrix",
			Corrected:        corrected,
		},
		func(transaction verticalslice.Transaction) (verticalslice.CommandReplayArtifact, error) {
			return verticalslice.CommandReplayArtifact{
				StatusCode: 200,
				Body:       []byte(transaction.ID),
				RequestID:  uuid.NewString(),
				TraceID:    "authorization-matrix",
			}, nil
		},
	)
	if err == nil {
		t.Fatal("cross-principal correction unexpectedly succeeded")
	}

	_, _, err = h.service.ReverseTransactionWithReplay(
		h.ctx,
		verticalslice.RequestContext{RequestID: uuid.NewString()},
		attackerSubject,
		uuid.NewString(),
		"/api/v1/portfolios/"+h.portfolioID+"/transactions/"+original.ID,
		verticalslice.ReverseTransactionRequest{
			PortfolioID:      h.portfolioID,
			TransactionID:    original.ID,
			ExpectedRevision: original.Revision,
			Reason:           "authorization matrix",
			EffectiveDate:    "2026-01-04",
		},
		func(result verticalslice.TransactionReversal) (verticalslice.CommandReplayArtifact, error) {
			return verticalslice.CommandReplayArtifact{
				StatusCode: 200,
				Body:       []byte(result.ReversalTransactionID),
				RequestID:  uuid.NewString(),
				TraceID:    "authorization-matrix",
			}, nil
		},
	)
	if err == nil {
		t.Fatal("cross-principal reversal unexpectedly succeeded")
	}

	var ledgerAfter int
	if err := h.db.QueryRowContext(
		h.ctx,
		"SELECT COUNT(*) FROM investment.transaction_entries WHERE portfolio_id=$1",
		h.portfolioID,
	).Scan(&ledgerAfter); err != nil {
		t.Fatal(err)
	}
	if ledgerAfter != ledgerBefore {
		t.Fatalf("rejected cross-principal mutations changed immutable ledger: before=%d after=%d", ledgerBefore, ledgerAfter)
	}

	ownerProjection, err := h.service.GetPortfolioPositions(h.ctx, h.subjectID, h.portfolioID, "")
	if err != nil {
		t.Fatalf("owner lost access after rejected attacker operations: %v", err)
	}
	if len(ownerProjection.Items) != 1 || ownerProjection.Items[0].Quantity.String() != "10.00000000" {
		t.Fatalf("owner state changed after rejected attacker operations: %+v", ownerProjection.Items)
	}
}
