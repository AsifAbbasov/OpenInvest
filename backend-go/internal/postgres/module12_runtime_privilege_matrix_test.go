package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"
)

// Exercise negative privileges only against the disposable owner/runtime
// PostgreSQL pair supplied by integration CI. Every attempted DDL is rolled
// back independently, including if an unexpected privilege is discovered.
func TestModule12RuntimeRoleCannotModifySchemaOrGrantPrivileges(t *testing.T) {
	ownerURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	runtimeURL := os.Getenv("OPENINVEST_DATABASE_RUNTIME_TEST_URL")
	if ownerURL == "" || runtimeURL == "" {
		t.Skip("isolated owner/runtime PostgreSQL integration URLs are not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ownerDB, err := sql.Open("pgx", ownerURL)
	if err != nil {
		t.Fatal(err)
	}
	defer ownerDB.Close()
	runtimeDB, err := sql.Open("pgx", runtimeURL)
	if err != nil {
		t.Fatal(err)
	}
	defer runtimeDB.Close()

	var ownerRole, runtimeRole string
	if err := ownerDB.QueryRowContext(ctx, "SELECT session_user").Scan(&ownerRole); err != nil {
		t.Fatal(err)
	}
	if err := runtimeDB.QueryRowContext(ctx, "SELECT session_user").Scan(&runtimeRole); err != nil {
		t.Fatal(err)
	}
	if ownerRole == runtimeRole {
		t.Fatal("owner and runtime credentials resolved to the same database login")
	}

	operations := []struct {
		name  string
		query string
	}{
		{"create_table", "CREATE TABLE investment.module12_runtime_escape(id integer)"},
		{"alter_table", "ALTER TABLE investment.transaction_entries ADD COLUMN module12_runtime_escape integer"},
		{"drop_table", "DROP TABLE investment.transaction_entries"},
		{"create_schema", "CREATE SCHEMA module12_runtime_escape"},
		{"set_role_owner", fmt.Sprintf("SET ROLE %q", ownerRole)},
		{"set_session_authorization_owner", fmt.Sprintf("SET SESSION AUTHORIZATION %q", ownerRole)},
	}
	for _, operation := range operations {
		t.Run(operation.name, func(t *testing.T) {
			tx, err := runtimeDB.BeginTx(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback()
			_, err = tx.ExecContext(ctx, operation.query)
			if err == nil {
				t.Fatalf("runtime login %q unexpectedly executed privileged operation", runtimeRole)
			}
			var pgState interface{ SQLState() string }
			if !errors.As(err, &pgState) || pgState.SQLState() != "42501" {
				t.Fatalf("expected insufficient_privilege (42501), got %v", err)
			}
		})
	}
	// PostgreSQL may report an ineffective GRANT/REVOKE as a successful
	// statement with a warning. Compare the actual ACL within the same
	// transaction, before rollback, instead of assuming SQLSTATE 42501.
	for _, operation := range []struct {
		name  string
		query string
	}{
		{"grant", "GRANT SELECT ON investment.transaction_entries TO PUBLIC"},
		{"revoke", "REVOKE SELECT ON investment.transaction_entries FROM PUBLIC"},
	} {
		t.Run(operation.name+"_cannot_change_acl", func(t *testing.T) {
			tx, err := runtimeDB.BeginTx(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback()
			const aclQuery = "SELECT COALESCE(relacl::text, '<NULL>') FROM pg_class WHERE oid = 'investment.transaction_entries'::regclass"
			var before string
			if err := tx.QueryRowContext(ctx, aclQuery).Scan(&before); err != nil {
				t.Fatal(err)
			}
			if _, err := tx.ExecContext(ctx, operation.query); err != nil {
				var pgState interface{ SQLState() string }
				if !errors.As(err, &pgState) || pgState.SQLState() != "42501" {
					t.Fatalf("unexpected privilege command error: %v", err)
				}
				return
			}
			var after string
			if err := tx.QueryRowContext(ctx, aclQuery).Scan(&after); err != nil {
				t.Fatal(err)
			}
			if before != after {
				t.Fatalf("runtime login %q changed table ACL via %s", runtimeRole, operation.name)
			}
		})
	}
}
