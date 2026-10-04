package main

import (
	"bytes"
	"context"
	"database/sql"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type recoveryFixture struct {
	db *sql.DB
}

func newRecoveryFixture(t *testing.T) recoveryFixture {
	t.Helper()
	adminURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if adminURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}
	parsed, err := url.Parse(adminURL)
	if err != nil {
		t.Fatalf("parse test database URL: %v", err)
	}
	databaseName := "openinvest_stage371_recovery_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	adminDB, err := sql.Open("pgx", adminURL)
	if err != nil {
		t.Fatalf("open recovery admin database: %v", err)
	}
	if _, err := adminDB.Exec(`CREATE DATABASE ` + databaseName); err != nil {
		_ = adminDB.Close()
		t.Fatalf("create isolated recovery database: %v", err)
	}

	isolateURL := *parsed
	isolateURL.Path = "/" + databaseName
	db, err := sql.Open("pgx", isolateURL.String())
	if err != nil {
		_, _ = adminDB.Exec(`DROP DATABASE ` + databaseName + ` WITH (FORCE)`)
		_ = adminDB.Close()
		t.Fatalf("open isolated recovery database: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
		if _, err := adminDB.Exec(`DROP DATABASE ` + databaseName + ` WITH (FORCE)`); err != nil {
			t.Errorf("drop isolated recovery database: %v", err)
		}
		_ = adminDB.Close()
	})
	if _, err := db.Exec(`
		CREATE SCHEMA investment;
		CREATE TABLE investment.transaction_entries (
			entry_id UUID PRIMARY KEY,
			portfolio_id UUID NOT NULL,
			ledger_sequence BIGINT
		);
	`); err != nil {
		t.Fatalf("create isolated recovery schema: %v", err)
	}
	return recoveryFixture{db: db}
}

func TestRecoverStage371LedgerSequenceIndexCreatesMissingExactIndex(t *testing.T) {
	fixture := newRecoveryFixture(t)
	if err := recoverStage371LedgerSequenceIndex(context.Background(), fixture.db); err != nil {
		t.Fatalf("recover missing Stage 3.71 index: %v", err)
	}
	state, err := inspectStage371LedgerSequenceIndex(context.Background(), fixture.db)
	if err != nil {
		t.Fatalf("inspect recovered index: %v", err)
	}
	if !state.present || !state.exact || !state.valid || !state.ready {
		t.Fatalf("recovered index state = %+v, want exact valid ready", state)
	}
}

func TestRecoverStage371LedgerSequenceIndexFailsClosedThenRetriesAfterCorrection(t *testing.T) {
	fixture := newRecoveryFixture(t)
	if _, err := fixture.db.Exec(`
		INSERT INTO investment.transaction_entries (entry_id, portfolio_id, ledger_sequence)
		VALUES
			('10000000-0000-4000-8000-000000000001', '20000000-0000-4000-8000-000000000001', 1),
			('10000000-0000-4000-8000-000000000002', '20000000-0000-4000-8000-000000000001', 1)
	`); err != nil {
		t.Fatalf("insert duplicate ledger sequence fixture: %v", err)
	}
	if _, err := fixture.db.Exec(`CREATE UNIQUE INDEX CONCURRENTLY transaction_entries_portfolio_ledger_sequence_uidx ON investment.transaction_entries (portfolio_id, ledger_sequence)`); err == nil {
		t.Fatal("interrupted concurrent unique index fixture unexpectedly succeeded")
	}

	invalid, err := inspectStage371LedgerSequenceIndex(context.Background(), fixture.db)
	if err != nil {
		t.Fatalf("inspect interrupted concurrent index: %v", err)
	}
	if !invalid.present || !invalid.exact || invalid.valid || invalid.ready {
		t.Fatalf("interrupted concurrent index state = %+v, want exact invalid not-ready", invalid)
	}

	err = recoverStage371LedgerSequenceIndex(context.Background(), fixture.db)
	if err == nil || !strings.Contains(err.Error(), "duplicate (portfolio_id, ledger_sequence)") {
		t.Fatalf("duplicate preflight error = %v, want clear duplicate rejection", err)
	}
	stillInvalid, err := inspectStage371LedgerSequenceIndex(context.Background(), fixture.db)
	if err != nil {
		t.Fatalf("inspect index after failed preflight: %v", err)
	}
	if !stillInvalid.present || !stillInvalid.exact || stillInvalid.valid || stillInvalid.ready {
		t.Fatalf("preflight changed invalid index state to %+v", stillInvalid)
	}

	if _, err := fixture.db.Exec(`UPDATE investment.transaction_entries SET ledger_sequence = NULL`); err != nil {
		t.Fatalf("apply operator data correction: %v", err)
	}
	if err := recoverStage371LedgerSequenceIndex(context.Background(), fixture.db); err != nil {
		t.Fatalf("recover after operator correction: %v", err)
	}
	final, err := inspectStage371LedgerSequenceIndex(context.Background(), fixture.db)
	if err != nil {
		t.Fatalf("inspect final recovery index: %v", err)
	}
	if !final.present || !final.exact || !final.valid || !final.ready {
		t.Fatalf("final recovery index state = %+v, want exact valid ready", final)
	}
}

func TestRecoverStage371LedgerSequenceIndexRefusesUnexpectedDefinition(t *testing.T) {
	fixture := newRecoveryFixture(t)
	if _, err := fixture.db.Exec(`CREATE INDEX transaction_entries_portfolio_ledger_sequence_uidx ON investment.transaction_entries (ledger_sequence, portfolio_id)`); err != nil {
		t.Fatalf("create unexpected index fixture: %v", err)
	}
	err := recoverStage371LedgerSequenceIndex(context.Background(), fixture.db)
	if err == nil || !strings.Contains(err.Error(), "does not have the expected schema, relation, or definition") {
		t.Fatalf("unexpected definition error = %v, want fail-closed rejection", err)
	}
	state, err := inspectStage371LedgerSequenceIndex(context.Background(), fixture.db)
	if err != nil {
		t.Fatalf("inspect unexpected index: %v", err)
	}
	if !state.present || state.exact {
		t.Fatalf("unexpected index was changed to %+v", state)
	}
}

func TestExecuteConcurrentIndexStatementRestoresBorrowedSessionTimeouts(t *testing.T) {
	fixture := newRecoveryFixture(t)
	fixture.db.SetMaxOpenConns(1)
	fixture.db.SetMaxIdleConns(1)
	if _, err := fixture.db.Exec(`
		SELECT
			set_config('lock_timeout', '1234ms', false),
			set_config('statement_timeout', '4321ms', false)
	`); err != nil {
		t.Fatalf("set fixture session timeouts: %v", err)
	}
	if err := executeConcurrentIndexStatement(context.Background(), fixture.db, `
		CREATE INDEX CONCURRENTLY transaction_entries_session_hygiene_idx
		ON investment.transaction_entries (ledger_sequence)
	`); err != nil {
		t.Fatalf("execute concurrent statement: %v", err)
	}

	var lockTimeout string
	var statementTimeout string
	if err := fixture.db.QueryRow(`
		SELECT current_setting('lock_timeout'), current_setting('statement_timeout')
	`).Scan(&lockTimeout, &statementTimeout); err != nil {
		t.Fatalf("read restored session timeouts: %v", err)
	}
	if lockTimeout != "1234ms" || statementTimeout != "4321ms" {
		t.Fatalf("borrowed connection leaked recovery timeouts: lock=%q statement=%q", lockTimeout, statementTimeout)
	}
}

func TestStage371RecoveryIndexDefinitionMatchesCanonicalMigration(t *testing.T) {
	root := repoRoot(t)
	migration, err := os.ReadFile(filepath.Join(root, "infrastructure", "postgres", "migrations", "000009_stage_03_71_ledger_sequence_unique.up.sql"))
	if err != nil {
		t.Fatalf("read canonical migration: %v", err)
	}
	canonicalStatement := "CREATE UNIQUE INDEX CONCURRENTLY transaction_entries_portfolio_ledger_sequence_uidx ON investment.transaction_entries (portfolio_id, ledger_sequence);"
	if !strings.Contains(string(migration), canonicalStatement) {
		t.Fatalf("canonical 000009 statement drifted from recovery contract")
	}
	wantEffectiveDefinition := "CREATE UNIQUE INDEX transaction_entries_portfolio_ledger_sequence_uidx ON investment.transaction_entries USING btree (portfolio_id, ledger_sequence)"
	if stage371IndexDefinition != wantEffectiveDefinition {
		t.Fatalf("recovery effective index definition = %q, want %q", stage371IndexDefinition, wantEffectiveDefinition)
	}
}

func TestCanonicalMigrationRunnerAppliesOrderedMigrationsAndHandlesStage371(t *testing.T) {
	adminURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if adminURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}
	parsed, err := url.Parse(adminURL)
	if err != nil {
		t.Fatalf("parse migration runner test database URL: %v", err)
	}
	databaseName := "openinvest_canonical_runner_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	adminDB, err := sql.Open("pgx", adminURL)
	if err != nil {
		t.Fatalf("open migration runner admin database: %v", err)
	}
	if _, err := adminDB.Exec(`CREATE DATABASE ` + databaseName); err != nil {
		_ = adminDB.Close()
		t.Fatalf("create isolated migration runner database: %v", err)
	}
	t.Cleanup(func() {
		if _, err := adminDB.Exec(`DROP DATABASE ` + databaseName + ` WITH (FORCE)`); err != nil {
			t.Errorf("drop isolated migration runner database: %v", err)
		}
		_ = adminDB.Close()
	})

	targetURL := *parsed
	targetURL.Path = "/" + databaseName
	root := repoRoot(t)
	command := exec.Command("bash", filepath.Join(root, "scripts", "apply-migrations.sh"))
	command.Dir = root
	command.Env = replaceEnvironment(os.Environ(), "OPENINVEST_DATABASE_OWNER_URL", targetURL.String())
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("canonical migration runner: %v\n%s", err, output)
	}
	if !bytes.Contains(output, []byte("MIGRATION_RUNNER_RESULT=PASS")) {
		t.Fatalf("canonical migration runner did not report success:\n%s", output)
	}

	entries, err := os.ReadDir(filepath.Join(root, "infrastructure", "postgres", "migrations"))
	if err != nil {
		t.Fatalf("read canonical migration directory: %v", err)
	}
	wantOrder := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.Type().IsRegular() && strings.HasSuffix(entry.Name(), ".up.sql") {
			wantOrder = append(wantOrder, entry.Name())
		}
	}
	sort.Strings(wantOrder)
	gotOrder := migrationRunnerOrder(string(output))
	if strings.Join(gotOrder, "\n") != strings.Join(wantOrder, "\n") {
		t.Fatalf("canonical migration runner order = %v, want %v", gotOrder, wantOrder)
	}

	db, err := sql.Open("pgx", targetURL.String())
	if err != nil {
		t.Fatalf("open canonical migration runner database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	var definition string
	if err := db.QueryRow(`SELECT pg_get_indexdef('investment.transaction_entries_portfolio_ledger_sequence_uidx'::regclass)`).Scan(&definition); err != nil {
		t.Fatalf("read effective Stage 3.71 index definition: %v", err)
	}
	if definition != stage371IndexDefinition {
		t.Fatalf("effective Stage 3.71 index definition = %q, want %q", definition, stage371IndexDefinition)
	}
}

func TestCanonicalMigrationRunnerRequiresExplicitOwnerURL(t *testing.T) {
	root := repoRoot(t)
	command := exec.Command("bash", filepath.Join(root, "scripts", "apply-migrations.sh"))
	command.Dir = root
	environment := replaceEnvironment(os.Environ(), "OPENINVEST_DATABASE_OWNER_URL", "")
	command.Env = replaceEnvironment(environment, "OPENINVEST_DATABASE_RUNTIME_URL", "postgres://runtime:runtime@127.0.0.1:5432/openinvest")
	output, err := command.CombinedOutput()
	if err == nil {
		t.Fatal("canonical migration runner accepted a runtime URL without an explicit owner URL")
	}
	if !bytes.Contains(output, []byte("OPENINVEST_DATABASE_OWNER_URL is required")) {
		t.Fatalf("owner URL rejection = %q", output)
	}
}

func migrationRunnerOrder(output string) []string {
	var order []string
	for _, line := range strings.Split(output, "\n") {
		if name, found := strings.CutPrefix(line, "MIGRATION_APPLY_FILE="); found {
			order = append(order, name)
		}
	}
	return order
}

func replaceEnvironment(environment []string, key, value string) []string {
	prefix := key + "="
	filtered := make([]string, 0, len(environment)+1)
	for _, entry := range environment {
		if !strings.HasPrefix(entry, prefix) {
			filtered = append(filtered, entry)
		}
	}
	return append(filtered, prefix+value)
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "infrastructure", "postgres", "migrations")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("repository root not found")
		}
		dir = parent
	}
}
