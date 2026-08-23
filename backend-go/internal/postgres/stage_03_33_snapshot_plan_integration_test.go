package postgres_test

import (
	"context"
	"database/sql"
	"os"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestStage0333ImportRebuildsExactAffectedSnapshotDatesOnce(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}

	store, err := postgres.Open(databaseURL)
	if err != nil {
		t.Fatalf("open postgres store: %v", err)
	}
	defer store.Close()

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open verification db: %v", err)
	}
	closeDBOnCleanup(t, db, "stage 3.33 snapshot verification")

	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	ctx := context.Background()
	subjectID := uuid.NewString()

	portfolio, err := service.CreatePortfolio(
		ctx,
		verticalslice.RequestContext{},
		subjectID,
		"stage-03-33-portfolio-key-0001",
		"/api/v1/portfolios",
		verticalslice.CreatePortfolioRequest{Name: "Stage 3.33 snapshots", BaseCurrency: verticalslice.RUB},
	)
	if err != nil {
		t.Fatalf("create portfolio: %v", err)
	}
	t.Cleanup(func() { cleanupPortfolioRows(t, ctx, db, portfolio.ID) })

	appendDeposit := func(key string, tradeDate string, amount string) {
		t.Helper()
		_, err := service.AppendTransaction(
			ctx,
			verticalslice.RequestContext{},
			subjectID,
			key,
			"/api/v1/portfolios/"+portfolio.ID+"/transactions",
			verticalslice.AppendTransactionRequest{
				PortfolioID:     portfolio.ID,
				TransactionType: "DEPOSIT",
				GrossAmount: &verticalslice.Money{
					Amount:   decimal.Must(amount),
					Currency: verticalslice.RUB,
				},
				Commission: verticalslice.Money{Amount: decimal.Zero(), Currency: verticalslice.RUB},
				Tax:        verticalslice.Money{Amount: decimal.Zero(), Currency: verticalslice.RUB},
				TradeDate:  tradeDate,
			},
		)
		if err != nil {
			t.Fatalf("append baseline deposit %s: %v", tradeDate, err)
		}
	}

	appendDeposit("stage-03-33-deposit-key-0010", "2026-06-10", "1000.00000000")
	appendDeposit("stage-03-33-deposit-key-0020", "2026-06-20", "200.00000000")
	appendDeposit("stage-03-33-deposit-key-0030", "2026-06-30", "300.00000000")

	request := verticalslice.AppendImportBatchRequest{
		PortfolioID:        portfolio.ID,
		SourceKind:         "USER_UPLOADED_FILE",
		SourceAccountLabel: "stage-03-33-account",
		SourceFileHash:     "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Transactions: []verticalslice.AppendTransactionRequest{
			{
				PortfolioID:     portfolio.ID,
				TransactionType: "DEPOSIT",
				GrossAmount: &verticalslice.Money{
					Amount:   decimal.Must("150.00000000"),
					Currency: verticalslice.RUB,
				},
				Commission: verticalslice.Money{Amount: decimal.Zero(), Currency: verticalslice.RUB},
				Tax:        verticalslice.Money{Amount: decimal.Zero(), Currency: verticalslice.RUB},
				TradeDate:  "2026-06-15",
			},
			{
				PortfolioID:     portfolio.ID,
				TransactionType: "DEPOSIT",
				GrossAmount: &verticalslice.Money{
					Amount:   decimal.Must("250.00000000"),
					Currency: verticalslice.RUB,
				},
				Commission: verticalslice.Money{Amount: decimal.Zero(), Currency: verticalslice.RUB},
				Tax:        verticalslice.Money{Amount: decimal.Zero(), Currency: verticalslice.RUB},
				TradeDate:  "2026-06-25",
			},
		},
	}
	for index := range request.Transactions {
		fingerprint, err := verticalslice.NormalizedTransactionFingerprint(request.Transactions[index])
		if err != nil {
			t.Fatalf("compute normalized import fingerprint for row %d: %v", index+1, err)
		}
		request.Transactions[index].ImportProvenance = &verticalslice.ImportProvenance{
			IdentityVersion:   verticalslice.ImportIdentityVersion,
			SourceFingerprint: fingerprint,
		}
	}

	var builderOutcome verticalslice.ImportAppendOutcome
	transactions, _, err := service.AppendImportedTransactionsWithOutcomeReplay(
		ctx,
		verticalslice.RequestContext{
			RequestID: "33333333-3333-4333-8333-333333333333",
			TraceID:   "stage0333snapshottrace",
		},
		subjectID,
		"stage-03-33-import-key-0001",
		"/api/v1/portfolios/"+portfolio.ID+"/imports/append",
		request,
		func(outcome verticalslice.ImportAppendOutcome) (verticalslice.CommandReplayArtifact, error) {
			builderOutcome = verticalslice.ImportAppendOutcome{
				Transactions:         append([]verticalslice.Transaction(nil), outcome.Transactions...),
				SnapshotDatesRebuilt: append([]string(nil), outcome.SnapshotDatesRebuilt...),
			}
			return verticalslice.CommandReplayArtifact{
				StatusCode: 201,
				Body:       []byte(`{"data":{"stage":"3.33"}}`),
				RequestID:  "33333333-3333-4333-8333-333333333333",
				TraceID:    "stage0333snapshottrace",
			}, nil
		},
	)
	if err != nil {
		t.Fatalf("append import with exact outcome replay: %v", err)
	}
	if len(transactions) != 2 || len(builderOutcome.Transactions) != 2 {
		t.Fatalf("expected two imported transactions, got returned=%d builder=%d", len(transactions), len(builderOutcome.Transactions))
	}

	wantDates := []string{"2026-06-15", "2026-06-20", "2026-06-25", "2026-06-30"}
	if !reflect.DeepEqual(builderOutcome.SnapshotDatesRebuilt, wantDates) {
		t.Fatalf("exact snapshot dates mismatch: got %v want %v", builderOutcome.SnapshotDatesRebuilt, wantDates)
	}

	rows, err := db.QueryContext(ctx, `
		SELECT snapshot_date::text, MAX(snapshot_version)
		FROM analytics.portfolio_snapshots
		WHERE portfolio_id = $1
			AND methodology_version = 'stage-03-02-local-cost-snapshot-v1'
		GROUP BY snapshot_date
		ORDER BY snapshot_date
	`, portfolio.ID)
	if err != nil {
		t.Fatalf("query snapshot versions: %v", err)
	}
	defer rows.Close()

	versions := map[string]int{}
	for rows.Next() {
		var date string
		var version int
		if err := rows.Scan(&date, &version); err != nil {
			t.Fatalf("scan snapshot version: %v", err)
		}
		versions[date] = version
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("read snapshot versions: %v", err)
	}

	wantVersions := map[string]int{
		"2026-06-10": 1,
		"2026-06-15": 1,
		"2026-06-20": 2,
		"2026-06-25": 1,
		"2026-06-30": 2,
	}
	if !reflect.DeepEqual(versions, wantVersions) {
		t.Fatalf("snapshot rebuild version counts prove duplicate work: got %v want %v", versions, wantVersions)
	}
}
