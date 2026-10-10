package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"
)

func TestModule12SafeRuntimePermissionConformance(t *testing.T) {
	ownerURL, runtimeURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL"), os.Getenv("OPENINVEST_DATABASE_RUNTIME_TEST_URL")
	if ownerURL == "" || runtimeURL == "" {
		t.Skip("isolated PostgreSQL service URLs required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	owner, err := sql.Open("pgx", ownerURL)
	if err != nil {
		t.Fatal(err)
	}
	defer owner.Close()
	runtime, err := sql.Open("pgx", runtimeURL)
	if err != nil {
		t.Fatal(err)
	}
	defer runtime.Close()
	if _, err = owner.ExecContext(ctx, `CREATE SCHEMA m12_safe_owner; CREATE TABLE m12_safe_owner.fixture (id integer)`); err != nil {
		t.Fatal(err)
	}
	defer owner.ExecContext(context.Background(), `DROP SCHEMA IF EXISTS m12_safe_owner CASCADE`)
	var ownerRole string
	if err = owner.QueryRowContext(ctx, `SELECT current_user`).Scan(&ownerRole); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, query string }{
		{"CREATE_TABLE", `CREATE TABLE investment.m12_safe_new (id integer)`},
		{"ALTER_TABLE", `ALTER TABLE m12_safe_owner.fixture ADD COLUMN extra integer`},
		{"DROP_TABLE", `DROP TABLE m12_safe_owner.fixture`},
		{"CREATE_SCHEMA", `CREATE SCHEMA m12_safe_runtime`},
		{"DROP_SCHEMA", `DROP SCHEMA m12_safe_owner CASCADE`},
		{"CREATE_EXTENSION", `CREATE EXTENSION file_fdw`},
		{"ALTER_ROLE", `ALTER ROLE "` + ownerRole + `" NOLOGIN`},
		{"GRANT", `GRANT SELECT ON m12_safe_owner.fixture TO PUBLIC`},
		{"REVOKE", `REVOKE ALL ON m12_safe_owner.fixture FROM PUBLIC`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tx, e := runtime.BeginTx(ctx, nil)
			if e != nil {
				t.Fatal(e)
			}
			defer tx.Rollback()
			_, e = tx.ExecContext(ctx, tc.query)
			var state interface{ SQLState() string }
			if e == nil {
				t.Error("expected permission denial")
			} else if !errors.As(e, &state) || state.SQLState() != "42501" {
				t.Errorf("expected SQLSTATE 42501, got %v", e)
			}
			if state != nil {
				t.Logf("M12_SAFE_RUNTIME_OPERATION=%s SQLSTATE=%s", tc.name, state.SQLState())
			}
		})
	}
	var role string
	if err = runtime.QueryRowContext(ctx, `SELECT current_user`).Scan(&role); err != nil {
		t.Error(err)
	}
	if role == ownerRole {
		t.Error("owner and runtime coincide")
	}
	t.Log("M12_SAFE_OWNER_RUNTIME_SEPARATION=YES")
}
