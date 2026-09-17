package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

type oiNew08SummaryFixture struct {
	ctx         context.Context
	db          *sql.DB
	service     *verticalslice.Service
	subjectID   string
	portfolioID string
}

func newOINew08SummaryFixture(t *testing.T, name string) oiNew08SummaryFixture {
	t.Helper()
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}

	store, err := postgres.Open(databaseURL)
	if err != nil {
		t.Fatalf("open OI-NEW-08 postgres store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open OI-NEW-08 verification database: %v", err)
	}
	closeDBOnCleanup(t, db, "OI-NEW-08 summary semantics")

	ctx := context.Background()
	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	subjectID := uuid.NewString()
	portfolio, err := service.CreatePortfolio(
		ctx,
		verticalslice.RequestContext{},
		subjectID,
		"oi-new-08-portfolio-"+uuid.NewString(),
		"/api/v1/portfolios",
		verticalslice.CreatePortfolioRequest{Name: name, BaseCurrency: verticalslice.RUB},
	)
	if err != nil {
		t.Fatalf("create OI-NEW-08 portfolio: %v", err)
	}
	t.Cleanup(func() { cleanupPortfolioRows(t, ctx, db, portfolio.ID) })

	return oiNew08SummaryFixture{
		ctx:         ctx,
		db:          db,
		service:     service,
		subjectID:   subjectID,
		portfolioID: portfolio.ID,
	}
}

func (fixture oiNew08SummaryFixture) appendDeposit(t *testing.T, tradeDate string, amount string) {
	t.Helper()
	_, err := fixture.service.AppendTransaction(
		fixture.ctx,
		verticalslice.RequestContext{},
		fixture.subjectID,
		"oi-new-08-deposit-"+uuid.NewString(),
		"/api/v1/portfolios/"+fixture.portfolioID+"/transactions",
		verticalslice.AppendTransactionRequest{
			PortfolioID:     fixture.portfolioID,
			TransactionType: "DEPOSIT",
			GrossAmount: &verticalslice.Money{
				Amount:   decimal.Must(amount),
				Currency: verticalslice.RUB,
			},
			Commission: verticalslice.ZeroMoney(),
			Tax:        verticalslice.ZeroMoney(),
			TradeDate:  tradeDate,
		},
	)
	if err != nil {
		t.Fatalf("append OI-NEW-08 deposit %s: %v", tradeDate, err)
	}
}

func TestOINew08SummaryExplicitDateSelectsLatestCalculatedSnapshotAtOrBeforeRequest(t *testing.T) {
	fixture := newOINew08SummaryFixture(t, "OI-NEW-08 explicit cutoff")
	fixture.appendDeposit(t, "2026-09-01", "100.00000000")
	fixture.appendDeposit(t, "2026-09-10", "200.00000000")

	summary, err := fixture.service.GetPortfolioSummary(
		fixture.ctx,
		fixture.subjectID,
		fixture.portfolioID,
		"2026-09-07",
	)
	if err != nil {
		t.Fatalf("get explicit-date OI-NEW-08 summary: %v", err)
	}
	if summary.AsOfDate != "2026-09-01" {
		t.Fatalf("explicit cutoff selected %q want 2026-09-01", summary.AsOfDate)
	}
	if got := summary.TotalValue.Amount.String(); got != "100.00000000" {
		t.Fatalf("explicit cutoff total=%s want 100.00000000", got)
	}
}

func TestOINew08SummaryOmissionSelectsAcceptedFutureDatedLatestSnapshot(t *testing.T) {
	fixture := newOINew08SummaryFixture(t, "OI-NEW-08 omitted cutoff")
	fixture.appendDeposit(t, "2026-09-01", "100.00000000")
	fixture.appendDeposit(t, "2099-01-01", "200.00000000")

	summary, err := fixture.service.GetPortfolioSummary(
		fixture.ctx,
		fixture.subjectID,
		fixture.portfolioID,
		"",
	)
	if err != nil {
		t.Fatalf("get omitted-date OI-NEW-08 summary: %v", err)
	}
	if summary.AsOfDate != "2099-01-01" {
		t.Fatalf("omitted cutoff selected %q want accepted future snapshot 2099-01-01", summary.AsOfDate)
	}
	if got := summary.TotalValue.Amount.String(); got != "300.00000000" {
		t.Fatalf("omitted cutoff total=%s want 300.00000000", got)
	}
}

func TestOINew08SummaryBeforeEarliestCalculatedSnapshotReturnsNotFound(t *testing.T) {
	fixture := newOINew08SummaryFixture(t, "OI-NEW-08 before earliest")
	fixture.appendDeposit(t, "2026-09-10", "100.00000000")

	_, err := fixture.service.GetPortfolioSummary(
		fixture.ctx,
		fixture.subjectID,
		fixture.portfolioID,
		"2026-09-01",
	)
	if !errors.Is(err, postgres.ErrNotFound) {
		t.Fatalf("before-earliest summary: got %v want postgres.ErrNotFound", err)
	}
}
