package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	ownerDatabaseEnv = "OPENINVEST_DATABASE_OWNER_URL"

	stage371IndexSchema     = "investment"
	stage371IndexName       = "transaction_entries_portfolio_ledger_sequence_uidx"
	stage371IndexDefinition = "CREATE UNIQUE INDEX transaction_entries_portfolio_ledger_sequence_uidx ON investment.transaction_entries USING btree (portfolio_id, ledger_sequence)"

	stage371RecoveryLockTimeout      = "5s"
	stage371RecoveryStatementTimeout = "360s"
)

type stage371IndexState struct {
	present bool
	exact   bool
	valid   bool
	ready   bool
}

func main() {
	databaseURL := strings.TrimSpace(os.Getenv(ownerDatabaseEnv))
	if databaseURL == "" {
		log.Fatalf("%s is required; the runtime DATABASE_URL must not be used for this owner-only recovery", ownerDatabaseEnv)
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
	defer cancel()
	if err := recoverStage371LedgerSequenceIndex(ctx, db); err != nil {
		log.Fatal(err)
	}
	log.Print("Stage 3.71 ledger sequence unique index is exact, valid, and ready")
}

func recoverStage371LedgerSequenceIndex(ctx context.Context, db *sql.DB) error {
	state, err := inspectStage371LedgerSequenceIndex(ctx, db)
	if err != nil {
		return err
	}
	if state.present && !state.exact {
		return errors.New("refusing Stage 3.71 index recovery: existing index name does not have the expected schema, relation, or definition")
	}
	if state.present && state.valid && state.ready {
		return nil
	}
	if err := preflightStage371LedgerSequenceIndex(ctx, db); err != nil {
		return err
	}

	if state.present {
		if err := executeConcurrentIndexStatement(ctx, db, `DROP INDEX CONCURRENTLY investment.transaction_entries_portfolio_ledger_sequence_uidx`); err != nil {
			return fmt.Errorf("drop invalid Stage 3.71 ledger sequence index: %w", err)
		}
	}
	if err := executeConcurrentIndexStatement(ctx, db, `CREATE UNIQUE INDEX CONCURRENTLY transaction_entries_portfolio_ledger_sequence_uidx ON investment.transaction_entries (portfolio_id, ledger_sequence)`); err != nil {
		return fmt.Errorf("create Stage 3.71 ledger sequence index: %w", err)
	}
	if err := preflightStage371LedgerSequenceIndex(ctx, db); err != nil {
		return fmt.Errorf("post-build Stage 3.71 ledger sequence preflight: %w", err)
	}

	state, err = inspectStage371LedgerSequenceIndex(ctx, db)
	if err != nil {
		return err
	}
	if !state.present || !state.exact || !state.valid || !state.ready {
		return errors.New("Stage 3.71 ledger sequence index recovery did not produce an exact, valid, ready index")
	}
	return nil
}

func preflightStage371LedgerSequenceIndex(ctx context.Context, db *sql.DB) error {
	var duplicateGroups int64
	var nonNullRows int64
	if err := db.QueryRowContext(ctx, `
		SELECT
			(
				SELECT COUNT(*)
				FROM (
					SELECT portfolio_id, ledger_sequence
					FROM investment.transaction_entries
					WHERE ledger_sequence IS NOT NULL
					GROUP BY portfolio_id, ledger_sequence
					HAVING COUNT(*) > 1
				) duplicate_groups
			),
			COUNT(*) FILTER (WHERE ledger_sequence IS NOT NULL)
		FROM investment.transaction_entries
	`).Scan(&duplicateGroups, &nonNullRows); err != nil {
		return fmt.Errorf("preflight Stage 3.71 ledger sequence index: %w", err)
	}
	if duplicateGroups != 0 {
		return fmt.Errorf("refusing Stage 3.71 index recovery: %d duplicate (portfolio_id, ledger_sequence) group(s) require operator correction", duplicateGroups)
	}
	if nonNullRows != 0 {
		return fmt.Errorf("refusing Stage 3.71 index recovery: %d populated ledger_sequence row(s) indicate Stage 3.71 activation has started", nonNullRows)
	}
	return nil
}

func executeConcurrentIndexStatement(ctx context.Context, db *sql.DB, statement string) (returnErr error) {
	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("acquire Stage 3.71 recovery connection: %w", err)
	}
	defer conn.Close()

	var originalLockTimeout string
	var originalStatementTimeout string
	if err := conn.QueryRowContext(ctx, `
		SELECT current_setting('lock_timeout'), current_setting('statement_timeout')
	`).Scan(&originalLockTimeout, &originalStatementTimeout); err != nil {
		return fmt.Errorf("read Stage 3.71 recovery session timeouts: %w", err)
	}
	defer func() {
		// Concurrent index operations cannot use SET LOCAL. Restore the borrowed
		// session explicitly so the sql.DB pool cannot hand altered limits to a caller.
		restoreCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := conn.ExecContext(restoreCtx, `
			SELECT
				set_config('lock_timeout', $1, false),
				set_config('statement_timeout', $2, false)
		`, originalLockTimeout, originalStatementTimeout); err != nil && returnErr == nil {
			returnErr = fmt.Errorf("restore Stage 3.71 recovery session timeouts: %w", err)
		}
	}()

	if _, err := conn.ExecContext(ctx, `SELECT set_config('lock_timeout', $1, false)`, stage371RecoveryLockTimeout); err != nil {
		return fmt.Errorf("set Stage 3.71 recovery lock timeout: %w", err)
	}
	if _, err := conn.ExecContext(ctx, `SELECT set_config('statement_timeout', $1, false)`, stage371RecoveryStatementTimeout); err != nil {
		return fmt.Errorf("set Stage 3.71 recovery statement timeout: %w", err)
	}
	if _, err := conn.ExecContext(ctx, statement); err != nil {
		return err
	}
	return nil
}

func inspectStage371LedgerSequenceIndex(ctx context.Context, db *sql.DB) (stage371IndexState, error) {
	var state stage371IndexState
	var relationSchema string
	var relationName string
	var definition string
	var unique bool
	var noPredicate bool
	var noExpressions bool
	var keyCount int
	var attributeCount int
	var keys string
	err := db.QueryRowContext(ctx, `
		SELECT
			table_namespace.nspname,
			table_class.relname,
			pg_get_indexdef(index_meta.indexrelid),
			index_meta.indisunique,
			index_meta.indisvalid,
			index_meta.indisready,
			index_meta.indpred IS NULL,
			index_meta.indexprs IS NULL,
			index_meta.indnkeyatts,
			index_meta.indnatts,
			array_to_string(ARRAY(
				SELECT attribute.attname::text
				FROM unnest(index_meta.indkey) WITH ORDINALITY AS key(attnum, ordinality)
				JOIN pg_attribute attribute
					ON attribute.attrelid = index_meta.indrelid
					AND attribute.attnum = key.attnum
				ORDER BY key.ordinality
			), ',')
		FROM pg_class index_class
		JOIN pg_namespace index_namespace ON index_namespace.oid = index_class.relnamespace
		JOIN pg_index index_meta ON index_meta.indexrelid = index_class.oid
		JOIN pg_class table_class ON table_class.oid = index_meta.indrelid
		JOIN pg_namespace table_namespace ON table_namespace.oid = table_class.relnamespace
		WHERE index_namespace.nspname = $1
			AND index_class.relname = $2
	`, stage371IndexSchema, stage371IndexName).Scan(
		&relationSchema,
		&relationName,
		&definition,
		&unique,
		&state.valid,
		&state.ready,
		&noPredicate,
		&noExpressions,
		&keyCount,
		&attributeCount,
		&keys,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return stage371IndexState{}, nil
	}
	if err != nil {
		return stage371IndexState{}, fmt.Errorf("inspect Stage 3.71 ledger sequence index: %w", err)
	}
	state.present = true
	state.exact = relationSchema == "investment" &&
		relationName == "transaction_entries" &&
		definition == stage371IndexDefinition &&
		unique &&
		noPredicate &&
		noExpressions &&
		keyCount == 2 &&
		attributeCount == 2 &&
		keys == "portfolio_id,ledger_sequence"
	return state, nil
}
