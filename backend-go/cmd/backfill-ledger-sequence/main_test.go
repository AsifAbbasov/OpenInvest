package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

type backfillFixture struct {
	db *sql.DB
}

func newBackfillFixture(t *testing.T, withIndex bool) backfillFixture {
	t.Helper()
	adminURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if adminURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}

	parsed, err := url.Parse(adminURL)
	if err != nil {
		t.Fatalf("parse test database URL: %v", err)
	}
	databaseName := "openinvest_backfill_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	adminDB, err := sql.Open("pgx", adminURL)
	if err != nil {
		t.Fatalf("open admin test database: %v", err)
	}
	if _, err := adminDB.Exec(`CREATE DATABASE ` + databaseName); err != nil {
		_ = adminDB.Close()
		t.Fatalf("create isolated backfill database: %v", err)
	}

	isolatedURL := *parsed
	isolatedURL.Path = "/" + databaseName
	db, err := sql.Open("pgx", isolatedURL.String())
	if err != nil {
		_, _ = adminDB.Exec(`DROP DATABASE ` + databaseName + ` WITH (FORCE)`)
		_ = adminDB.Close()
		t.Fatalf("open isolated backfill database: %v", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		_, _ = adminDB.Exec(`DROP DATABASE ` + databaseName + ` WITH (FORCE)`)
		_ = adminDB.Close()
		t.Fatalf("ping isolated backfill database: %v", err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close isolated backfill database: %v", err)
		}
		if _, err := adminDB.Exec(`DROP DATABASE ` + databaseName + ` WITH (FORCE)`); err != nil {
			t.Errorf("drop isolated backfill database: %v", err)
		}
		if err := adminDB.Close(); err != nil {
			t.Errorf("close admin test database: %v", err)
		}
	})

	if _, err := db.Exec(`
		CREATE SCHEMA investment;
		CREATE TABLE investment.transaction_entries (
			entry_id UUID PRIMARY KEY,
			transaction_id UUID NOT NULL,
			portfolio_id UUID NOT NULL,
			revision INTEGER NOT NULL,
			transaction_type TEXT NOT NULL,
			trade_date DATE NOT NULL,
			created_at TIMESTAMPTZ NOT NULL,
			prior_entry_id UUID,
			reverses_transaction_id UUID,
			ledger_sequence BIGINT
		);
	`); err != nil {
		t.Fatalf("create isolated backfill schema: %v", err)
	}
	if withIndex {
		if _, err := db.Exec(`
			CREATE UNIQUE INDEX transaction_entries_portfolio_ledger_sequence_uidx
			ON investment.transaction_entries (portfolio_id, ledger_sequence)
		`); err != nil {
			t.Fatalf("create Stage 3.71 unique index fixture: %v", err)
		}
	}

	return backfillFixture{db: db}
}

func insertBackfillRow(
	t *testing.T,
	fixture backfillFixture,
	entryID string,
	transactionID string,
	portfolioID string,
	transactionType string,
	tradeDate string,
	createdAt string,
	revision int,
	ledgerSequence any,
) {
	t.Helper()
	if _, err := fixture.db.Exec(`
		INSERT INTO investment.transaction_entries (
			entry_id, transaction_id, portfolio_id, revision, transaction_type,
			trade_date, created_at, ledger_sequence
		)
		VALUES ($1, $2, $3, $4, $5, $6::date, $7::timestamptz, $8)
	`, entryID, transactionID, portfolioID, revision, transactionType, tradeDate, createdAt, ledgerSequence); err != nil {
		t.Fatalf("insert backfill fixture row %s: %v", entryID, err)
	}
}

func TestBackfillRequiresExactReadyUniqueIndexBeforeMutation(t *testing.T) {
	fixture := newBackfillFixture(t, false)
	portfolioID := "10000000-0000-4000-8000-000000000001"
	entryID := "10000000-0000-4000-8000-000000000011"
	insertBackfillRow(
		t, fixture, entryID, "10000000-0000-4000-8000-000000000021", portfolioID,
		"BUY", "2026-01-01", "2026-01-01T10:00:00Z", 1, nil,
	)

	err := backfill(context.Background(), fixture.db)
	if err == nil || !strings.Contains(err.Error(), "apply migration 000009 before backfill") {
		t.Fatalf("expected missing exact unique index to fail closed, got %v", err)
	}

	var sequence sql.NullInt64
	if err := fixture.db.QueryRow(`SELECT ledger_sequence FROM investment.transaction_entries WHERE entry_id = $1`, entryID).Scan(&sequence); err != nil {
		t.Fatalf("read row after rejected backfill: %v", err)
	}
	if sequence.Valid {
		t.Fatalf("backfill mutated ledger_sequence before index gate: got %d", sequence.Int64)
	}
}

func TestBackfillUsesFrozenTupleAndRerunIsVerifyOnly(t *testing.T) {
	fixture := newBackfillFixture(t, true)
	portfolioA := "20000000-0000-4000-8000-000000000001"
	portfolioB := "20000000-0000-4000-8000-000000000002"

	rows := []struct {
		entryID       string
		transactionID string
		portfolioID   string
		tradeDate     string
		createdAt     string
	}{
		{"20000000-0000-4000-8000-000000000014", "20000000-0000-4000-8000-000000000024", portfolioA, "2026-01-02", "2026-01-02T10:00:00Z"},
		{"20000000-0000-4000-8000-000000000012", "20000000-0000-4000-8000-000000000022", portfolioA, "2026-01-02", "2026-01-02T09:00:00Z"},
		{"20000000-0000-4000-8000-000000000011", "20000000-0000-4000-8000-000000000021", portfolioA, "2026-01-01", "2026-01-03T12:00:00Z"},
		{"20000000-0000-4000-8000-000000000013", "20000000-0000-4000-8000-000000000023", portfolioA, "2026-01-02", "2026-01-02T09:00:00Z"},
		{"20000000-0000-4000-8000-000000000015", "20000000-0000-4000-8000-000000000025", portfolioB, "2026-01-05", "2026-01-05T08:00:00Z"},
	}
	for _, row := range rows {
		insertBackfillRow(t, fixture, row.entryID, row.transactionID, row.portfolioID, "BUY", row.tradeDate, row.createdAt, 1, nil)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := backfill(ctx, fixture.db); err != nil {
		t.Fatalf("backfill frozen tuple fixture: %v", err)
	}

	want := map[string]int64{
		"20000000-0000-4000-8000-000000000011": 1,
		"20000000-0000-4000-8000-000000000012": 2,
		"20000000-0000-4000-8000-000000000013": 3,
		"20000000-0000-4000-8000-000000000014": 4,
		"20000000-0000-4000-8000-000000000015": 1,
	}
	for entryID, expectedSequence := range want {
		var sequence int64
		if err := fixture.db.QueryRow(`SELECT ledger_sequence FROM investment.transaction_entries WHERE entry_id = $1`, entryID).Scan(&sequence); err != nil {
			t.Fatalf("read populated sequence for %s: %v", entryID, err)
		}
		if sequence != expectedSequence {
			t.Fatalf("frozen backfill ordering mismatch for %s: got %d want %d", entryID, sequence, expectedSequence)
		}
	}

	laterRuntimeEntry := "20000000-0000-4000-8000-000000000016"
	insertBackfillRow(
		t, fixture, laterRuntimeEntry, "20000000-0000-4000-8000-000000000026", portfolioA,
		"BUY", "2025-12-01", "2026-02-01T00:00:00Z", 1, int64(5),
	)
	if err := backfill(context.Background(), fixture.db); err != nil {
		t.Fatalf("verify-only rerun: %v", err)
	}
	var retainedSequence int64
	if err := fixture.db.QueryRow(`SELECT ledger_sequence FROM investment.transaction_entries WHERE entry_id = $1`, laterRuntimeEntry).Scan(&retainedSequence); err != nil {
		t.Fatalf("read verify-only runtime sequence: %v", err)
	}
	if retainedSequence != 5 {
		t.Fatalf("verify-only rerun reordered a later runtime row: got %d want 5", retainedSequence)
	}
}

func TestBackfillRejectsPartialPopulationWithoutMutation(t *testing.T) {
	fixture := newBackfillFixture(t, true)
	portfolioID := "30000000-0000-4000-8000-000000000001"
	first := "30000000-0000-4000-8000-000000000011"
	second := "30000000-0000-4000-8000-000000000012"
	insertBackfillRow(t, fixture, first, "30000000-0000-4000-8000-000000000021", portfolioID, "BUY", "2026-01-01", "2026-01-01T00:00:00Z", 1, int64(1))
	insertBackfillRow(t, fixture, second, "30000000-0000-4000-8000-000000000022", portfolioID, "BUY", "2026-01-02", "2026-01-02T00:00:00Z", 1, nil)

	err := backfill(context.Background(), fixture.db)
	if err == nil || !strings.Contains(err.Error(), "refusing partial ledger_sequence population") {
		t.Fatalf("expected partial population rejection, got %v", err)
	}
	var sequence sql.NullInt64
	if err := fixture.db.QueryRow(`SELECT ledger_sequence FROM investment.transaction_entries WHERE entry_id = $1`, second).Scan(&sequence); err != nil {
		t.Fatalf("read NULL row after partial rejection: %v", err)
	}
	if sequence.Valid {
		t.Fatalf("partial-population rejection mutated NULL row to %d", sequence.Int64)
	}
}

func TestBackfillRejectsHistoricalSellAndUnsupportedRevision(t *testing.T) {
	cases := []struct {
		name            string
		transactionType string
		revision        int
		wantError       string
	}{
		{name: "historical sell", transactionType: "SELL", revision: 1, wantError: "historical SELL premise failed"},
		{name: "unsupported revision", transactionType: "BUY", revision: 2, wantError: "correction/reversal premise failed"},
	}
	for index, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newBackfillFixture(t, true)
			portfolioID := fmt.Sprintf("40000000-0000-4000-8000-%012d", index+1)
			entryID := fmt.Sprintf("40000000-0000-4000-8001-%012d", index+1)
			transactionID := fmt.Sprintf("40000000-0000-4000-8002-%012d", index+1)
			insertBackfillRow(t, fixture, entryID, transactionID, portfolioID, testCase.transactionType, "2026-01-01", "2026-01-01T00:00:00Z", testCase.revision, nil)

			err := backfill(context.Background(), fixture.db)
			if err == nil || !strings.Contains(err.Error(), testCase.wantError) {
				t.Fatalf("expected %q, got %v", testCase.wantError, err)
			}
			var sequence sql.NullInt64
			if err := fixture.db.QueryRow(`SELECT ledger_sequence FROM investment.transaction_entries WHERE entry_id = $1`, entryID).Scan(&sequence); err != nil {
				t.Fatalf("read row after premise rejection: %v", err)
			}
			if sequence.Valid {
				t.Fatalf("premise rejection mutated ledger_sequence to %d", sequence.Int64)
			}
		})
	}
}
