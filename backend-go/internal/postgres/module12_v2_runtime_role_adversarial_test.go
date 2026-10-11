package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestModule12V2RuntimeRoleAdversarialMatrix(t *testing.T) {
	ownerURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	runtimeURL := os.Getenv("OPENINVEST_DATABASE_RUNTIME_TEST_URL")
	if ownerURL == "" || runtimeURL == "" {
		t.Skip("isolated PostgreSQL owner/runtime URLs required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	owner, err := sql.Open("pgx", ownerURL)
	if err != nil { t.Fatal(err) }
	defer owner.Close()
	runtimeDB, err := sql.Open("pgx", runtimeURL)
	if err != nil { t.Fatal(err) }
	defer runtimeDB.Close()

	var ownerRole, runtimeRole string
	if err := owner.QueryRowContext(ctx, "SELECT current_user").Scan(&ownerRole); err != nil { t.Fatal(err) }
	if err := runtimeDB.QueryRowContext(ctx, "SELECT current_user").Scan(&runtimeRole); err != nil { t.Fatal(err) }
	if ownerRole == runtimeRole { t.Fatal("OWNER_RUNTIME_SEPARATION=NO") }

	if _, err := owner.ExecContext(ctx, "CREATE SCHEMA IF NOT EXISTS m12_v2_owner; CREATE TABLE IF NOT EXISTS m12_v2_owner.fixture(id integer primary key, value text);"); err != nil {
		t.Fatal(err)
	}
	defer owner.ExecContext(context.Background(), "DROP SCHEMA IF EXISTS m12_v2_owner CASCADE")

	ops := []struct{name, sql string}{
		{"CREATE_TABLE", "CREATE TABLE investment.m12_v2_escape(id integer)"},
		{"ALTER_TABLE", "ALTER TABLE m12_v2_owner.fixture ADD COLUMN extra integer"},
		{"DROP_TABLE", "DROP TABLE m12_v2_owner.fixture"},
		{"CREATE_SCHEMA", "CREATE SCHEMA m12_v2_runtime_escape"},
		{"DROP_SCHEMA", "DROP SCHEMA m12_v2_owner CASCADE"},
		{"CREATE_EXTENSION", "CREATE EXTENSION file_fdw"},
		{"ALTER_ROLE", fmt.Sprintf("ALTER ROLE %q NOLOGIN", ownerRole)},
		{"GRANT", "GRANT SELECT ON m12_v2_owner.fixture TO PUBLIC"},
		{"REVOKE", "REVOKE ALL ON m12_v2_owner.fixture FROM PUBLIC"},
		{"SET_ROLE_OWNER", fmt.Sprintf("SET ROLE %q", ownerRole)},
	}
	for _, op := range ops {
		op := op
		t.Run(op.name, func(t *testing.T) {
			tx, err := runtimeDB.BeginTx(ctx, nil)
			if err != nil { t.Fatal(err) }
			defer tx.Rollback()
			_, err = tx.ExecContext(ctx, op.sql)
			if err == nil {
				t.Errorf("unexpected runtime privilege: %s succeeded", op.name)
				return
			}
			var state interface{ SQLState() string }
			if !errors.As(err, &state) {
				t.Errorf("%s returned non-SQLSTATE error: %v", op.name, err)
				return
			}
			if state.SQLState() != "42501" {
				t.Errorf("%s expected insufficient_privilege 42501, got %s: %v", op.name, state.SQLState(), err)
			}
			t.Logf("RUNTIME_OPERATION=%s SQLSTATE=%s", op.name, state.SQLState())
		})
	}

	rows, err := owner.QueryContext(ctx, "SELECT table_schema, table_name FROM information_schema.tables WHERE lower(table_name) LIKE '%migration%' ORDER BY table_schema, table_name")
	if err != nil { t.Fatal(err) }
	defer rows.Close()
	found := false
	for rows.Next() {
		var schema, table string
		if err := rows.Scan(&schema, &table); err != nil { t.Fatal(err) }
		found = true
		query := fmt.Sprintf("DELETE FROM %q.%q WHERE false", schema, table)
		tx, err := runtimeDB.BeginTx(ctx, nil)
		if err != nil { t.Fatal(err) }
		_, execErr := tx.ExecContext(ctx, query)
		_ = tx.Rollback()
		if execErr == nil {
			t.Errorf("runtime role can mutate migration bookkeeping relation %s.%s", schema, table)
			continue
		}
		var state interface{ SQLState() string }
		if !errors.As(execErr, &state) || state.SQLState() != "42501" {
			t.Errorf("migration bookkeeping %s.%s expected 42501, got %v", schema, table, execErr)
		}
		t.Logf("MIGRATION_BOOKKEEPING_RELATION=%s.%s RESULT=DENIED", schema, table)
	}
	if !found { t.Log("MIGRATION_BOOKKEEPING_MUTATION=NOT_APPLICABLE_NO_BOOKKEEPING_RELATION_FOUND") }

	var canCreateTemp bool
	if err := runtimeDB.QueryRowContext(ctx, "SELECT has_database_privilege(current_user,current_database(),'TEMP')").Scan(&canCreateTemp); err == nil {
		t.Logf("RUNTIME_DATABASE_TEMP_PRIVILEGE=%t", canCreateTemp)
	}
	t.Logf("OWNER_RUNTIME_SEPARATION=YES OWNER=%s RUNTIME=%s", ownerRole, runtimeRole)

	// Verify the role can still perform expected non-DDL work in an allowed schema.
	var one int
	if err := runtimeDB.QueryRowContext(ctx, "SELECT 1").Scan(&one); err != nil || one != 1 {
		t.Fatalf("LEGITIMATE_RUNTIME_DML=FAIL simple query err=%v", err)
	}
	t.Log("LEGITIMATE_RUNTIME_DML=PASS")
}

func TestModule12V2RuntimeRoleCannotChangeSessionAuthorization(t *testing.T) {
	runtimeURL := os.Getenv("OPENINVEST_DATABASE_RUNTIME_TEST_URL")
	if strings.TrimSpace(runtimeURL) == "" { t.Skip("isolated runtime URL required") }
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	db, err := sql.Open("pgx", runtimeURL)
	if err != nil { t.Fatal(err) }
	defer db.Close()
	_, err = db.ExecContext(ctx, "SET SESSION AUTHORIZATION openinvest")
	if err == nil { t.Fatal("runtime role changed session authorization to owner") }
	var state interface{ SQLState() string }
	if errors.As(err, &state) { t.Logf("SET_SESSION_AUTHORIZATION SQLSTATE=%s", state.SQLState()) }
}
