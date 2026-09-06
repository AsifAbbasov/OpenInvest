package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const ownerDatabaseEnv = "OPENINVEST_DATABASE_OWNER_URL"

func main() {
	databaseURL := os.Getenv(ownerDatabaseEnv)
	if databaseURL == "" {
		log.Fatalf("%s is required; the runtime DATABASE_URL must not be used for this owner-only backfill", ownerDatabaseEnv)
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	if err := backfill(ctx, db); err != nil {
		log.Fatal(err)
	}
	log.Print("Stage 3.71 ledger sequence population verified")
}

func backfill(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `LOCK TABLE investment.transaction_entries IN SHARE ROW EXCLUSIVE MODE`); err != nil {
		return err
	}
	if err := verifyStage371UniqueIndex(ctx, tx); err != nil {
		return err
	}

	var totalRows int64
	var nullRows int64
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*), COUNT(*) FILTER (WHERE ledger_sequence IS NULL)
		FROM investment.transaction_entries
	`).Scan(&totalRows, &nullRows); err != nil {
		return err
	}

	if nullRows > 0 && nullRows != totalRows {
		return errors.New("refusing partial ledger_sequence population: expected all pre-activation rows NULL or all rows already populated")
	}

	if nullRows == 0 {
		if err := verifyStructuralIntegrity(ctx, tx); err != nil {
			return err
		}
		return tx.Commit()
	}

	var historicalSell int64
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM investment.transaction_entries
		WHERE transaction_type = 'SELL'
	`).Scan(&historicalSell); err != nil {
		return err
	}
	if historicalSell != 0 {
		return fmt.Errorf("historical SELL premise failed: found %d pre-Stage-3.71 SELL rows", historicalSell)
	}

	var unsupportedRevisionRows int64
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM investment.transaction_entries
		WHERE revision <> 1
			OR prior_entry_id IS NOT NULL
			OR reverses_transaction_id IS NOT NULL
	`).Scan(&unsupportedRevisionRows); err != nil {
		return err
	}
	if unsupportedRevisionRows != 0 {
		return fmt.Errorf("correction/reversal premise failed: found %d unsupported historical rows", unsupportedRevisionRows)
	}

	result, err := tx.ExecContext(ctx, `
		WITH ranked AS (
			SELECT
				entry_id,
				ROW_NUMBER() OVER (
					PARTITION BY portfolio_id
					ORDER BY
						trade_date ASC,
						created_at ASC,
						transaction_id ASC,
						revision ASC,
						entry_id ASC
				) AS ledger_sequence
			FROM investment.transaction_entries
		)
		UPDATE investment.transaction_entries te
		SET ledger_sequence = ranked.ledger_sequence
		FROM ranked
		WHERE te.entry_id = ranked.entry_id
			AND te.ledger_sequence IS NULL
	`)
	if err != nil {
		return err
	}
	updatedRows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if updatedRows != totalRows {
		return fmt.Errorf("backfill updated %d rows; expected %d", updatedRows, totalRows)
	}

	var rankingMismatches int64
	if err := tx.QueryRowContext(ctx, `
		WITH ranked AS (
			SELECT
				entry_id,
				ROW_NUMBER() OVER (
					PARTITION BY portfolio_id
					ORDER BY
						trade_date ASC,
						created_at ASC,
						transaction_id ASC,
						revision ASC,
						entry_id ASC
				) AS expected_sequence
			FROM investment.transaction_entries
		)
		SELECT COUNT(*)
		FROM investment.transaction_entries te
		JOIN ranked ON ranked.entry_id = te.entry_id
		WHERE te.ledger_sequence <> ranked.expected_sequence
	`).Scan(&rankingMismatches); err != nil {
		return err
	}
	if rankingMismatches != 0 {
		return fmt.Errorf("deterministic backfill verification failed for %d rows", rankingMismatches)
	}

	if err := verifyStructuralIntegrity(ctx, tx); err != nil {
		return err
	}
	return tx.Commit()
}

func verifyStage371UniqueIndex(ctx context.Context, tx *sql.Tx) error {
	var ready bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM pg_class index_class
			JOIN pg_namespace index_namespace ON index_namespace.oid = index_class.relnamespace
			JOIN pg_index index_meta ON index_meta.indexrelid = index_class.oid
			WHERE index_namespace.nspname = 'investment'
				AND index_class.relname = 'transaction_entries_portfolio_ledger_sequence_uidx'
				AND index_meta.indrelid = 'investment.transaction_entries'::regclass
				AND index_meta.indisunique
				AND index_meta.indisvalid
				AND index_meta.indisready
				AND index_meta.indpred IS NULL
				AND index_meta.indexprs IS NULL
				AND index_meta.indnkeyatts = 2
				AND index_meta.indnatts = 2
				AND (
					SELECT array_agg(attribute.attname::text ORDER BY key.ordinality)
					FROM unnest(index_meta.indkey) WITH ORDINALITY AS key(attnum, ordinality)
					JOIN pg_attribute attribute
						ON attribute.attrelid = index_meta.indrelid
						AND attribute.attnum = key.attnum
				) = ARRAY['portfolio_id', 'ledger_sequence']::text[]
		)
	`).Scan(&ready); err != nil {
		return err
	}
	if !ready {
		return errors.New("Stage 3.71 ledger sequence unique index is missing, invalid, not ready, or structurally incorrect; apply migration 000009 before backfill")
	}
	return nil
}

func verifyStructuralIntegrity(ctx context.Context, tx *sql.Tx) error {
	var invalidRows int64
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM investment.transaction_entries
		WHERE ledger_sequence IS NULL OR ledger_sequence <= 0
	`).Scan(&invalidRows); err != nil {
		return err
	}
	if invalidRows != 0 {
		return fmt.Errorf("ledger sequence integrity failed: %d NULL/non-positive rows", invalidRows)
	}

	var duplicateGroups int64
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM (
			SELECT portfolio_id, ledger_sequence
			FROM investment.transaction_entries
			GROUP BY portfolio_id, ledger_sequence
			HAVING COUNT(*) > 1
		) duplicates
	`).Scan(&duplicateGroups); err != nil {
		return err
	}
	if duplicateGroups != 0 {
		return fmt.Errorf("ledger sequence integrity failed: %d duplicate portfolio-local sequence groups", duplicateGroups)
	}
	return nil
}
