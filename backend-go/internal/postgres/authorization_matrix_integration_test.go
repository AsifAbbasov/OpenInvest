package postgres_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestAuthorizationMatrixRejectsCrossSubjectPortfolioAccess(t *testing.T) {
	h := newStage371Harness(t, "authorization owner portfolio")
	owner := h.subjectID
	attacker := uuid.NewString()

	appendStage371Trade(t, h, stage371Trade(
		h.portfolioID, "BUY", "SBER", "10.00000000", "100.00000000", "2026-01-01",
	))

	assertNotFound := func(name string, err error) {
		t.Helper()
		if !errors.Is(err, postgres.ErrNotFound) && !errors.Is(err, verticalslice.ErrNotFound) {
			t.Fatalf("%s: cross-subject access must fail closed as not found, got %v", name, err)
		}
	}

	if _, err := h.service.GetPortfolio(h.ctx, attacker, h.portfolioID); err == nil {
		t.Fatal("cross-subject GetPortfolio unexpectedly succeeded")
	} else {
		assertNotFound("GetPortfolio", err)
	}

	if _, err := h.service.GetPortfolioSummary(h.ctx, attacker, h.portfolioID, ""); err == nil {
		t.Fatal("cross-subject GetPortfolioSummary unexpectedly succeeded")
	} else {
		assertNotFound("GetPortfolioSummary", err)
	}

	if _, err := h.service.GetPortfolioPositions(h.ctx, attacker, h.portfolioID, ""); err == nil {
		t.Fatal("cross-subject GetPortfolioPositions unexpectedly succeeded")
	} else {
		assertNotFound("GetPortfolioPositions", err)
	}

	if _, err := h.service.ListTransactions(
		h.ctx, attacker, h.portfolioID, verticalslice.TransactionFilter{Limit: 100},
	); err == nil {
		t.Fatal("cross-subject ListTransactions unexpectedly succeeded")
	} else {
		assertNotFound("ListTransactions", err)
	}

	_, err := h.service.AppendTransaction(
		h.ctx,
		verticalslice.RequestContext{RequestID: uuid.NewString()},
		attacker,
		uuid.NewString(),
		"/api/v1/portfolios/"+h.portfolioID+"/transactions",
		stage371Trade(h.portfolioID, "BUY", "SBER", "1.00000000", "101.00000000", "2026-01-02"),
	)
	if err == nil {
		t.Fatal("cross-subject AppendTransaction unexpectedly succeeded")
	}
	assertNotFound("AppendTransaction", err)

	items, err := h.service.ListTransactions(
		h.ctx, owner, h.portfolioID, verticalslice.TransactionFilter{Limit: 100},
	)
	if err != nil {
		t.Fatalf("owner ListTransactions after cross-subject attempts: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("cross-subject attempts mutated owner ledger: got %d rows want 1", len(items))
	}
}

func TestAuthorizationMatrixPortfolioIDsDoNotGrantAuthority(t *testing.T) {
	h := newStage371Harness(t, "authorization opaque id")
	attacker := uuid.NewString()

	for _, asOf := range []string{"", "2026-01-01"} {
		_, err := h.service.GetPortfolioPositions(h.ctx, attacker, h.portfolioID, asOf)
		if err == nil {
			t.Fatalf("attacker with exact portfolio UUID read positions for asOf=%q", asOf)
		}
		if !errors.Is(err, postgres.ErrNotFound) && !errors.Is(err, verticalslice.ErrNotFound) {
			t.Fatalf("asOf=%q returned non-fail-closed error: %v", asOf, err)
		}
	}
}
