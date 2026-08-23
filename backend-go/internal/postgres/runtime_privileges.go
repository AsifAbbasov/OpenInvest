package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrUnsafeRuntimeDatabaseRole = errors.New("unsafe PostgreSQL runtime role")

var protectedAppendOnlyRuntimeTables = []struct {
	Schema string
	Table  string
}{
	{Schema: "investment", Table: "transaction_entries"},
	{Schema: "audit", Table: "events"},
}

// OpenRuntime opens the application store and verifies that the effective database principal is a
// dedicated least-privilege runtime role. Migration/schema-owner connections must use Open instead.
func OpenRuntime(databaseURL string) (*Store, error) {
	store, err := Open(databaseURL)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := store.ValidateRuntimePrivileges(ctx); err != nil {
		_ = store.Close()
		return nil, err
	}
	return store, nil
}

// ValidateRuntimePrivileges proves the append-only runtime boundary against PostgreSQL's effective
// ACLs. It deliberately checks capabilities rather than a hard-coded login name so provider-managed
// LOGIN roles can inherit the repository-owned openinvest_runtime capability role.
func (s *Store) ValidateRuntimePrivileges(ctx context.Context) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("%w: database store is not initialized", ErrUnsafeRuntimeDatabaseRole)
	}

	for _, protected := range protectedAppendOnlyRuntimeTables {
		qualified := protected.Schema + "." + protected.Table
		var currentUser string
		var superuser bool
		var memberOfOwner bool
		var schemaUsage bool
		var canSelect bool
		var canInsert bool
		var canUpdate bool
		var canDelete bool
		var canTruncate bool

		err := s.db.QueryRowContext(ctx, `
			SELECT
				current_user,
				role_row.rolsuper,
				pg_has_role(current_user, table_row.relowner, 'MEMBER'),
				has_schema_privilege(current_user, $2, 'USAGE'),
				has_table_privilege(current_user, $1, 'SELECT'),
				has_table_privilege(current_user, $1, 'INSERT'),
				has_table_privilege(current_user, $1, 'UPDATE'),
				has_table_privilege(current_user, $1, 'DELETE'),
				has_table_privilege(current_user, $1, 'TRUNCATE')
			FROM pg_roles role_row
			JOIN pg_class table_row ON table_row.oid = to_regclass($1)
			WHERE role_row.rolname = current_user
		`, qualified, protected.Schema).Scan(
			&currentUser,
			&superuser,
			&memberOfOwner,
			&schemaUsage,
			&canSelect,
			&canInsert,
			&canUpdate,
			&canDelete,
			&canTruncate,
		)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("%w: protected table %s is unavailable", ErrUnsafeRuntimeDatabaseRole, qualified)
			}
			return fmt.Errorf("%w: inspect %s privileges: %v", ErrUnsafeRuntimeDatabaseRole, qualified, err)
		}

		currentUser = strings.TrimSpace(currentUser)
		if currentUser == "" {
			return fmt.Errorf("%w: current database principal is empty", ErrUnsafeRuntimeDatabaseRole)
		}
		if superuser {
			return fmt.Errorf("%w: principal %s is a superuser", ErrUnsafeRuntimeDatabaseRole, currentUser)
		}
		if memberOfOwner {
			return fmt.Errorf("%w: principal %s owns or inherits the owner of %s", ErrUnsafeRuntimeDatabaseRole, currentUser, qualified)
		}
		if !schemaUsage || !canSelect || !canInsert {
			return fmt.Errorf("%w: principal %s lacks required append/read capability on %s", ErrUnsafeRuntimeDatabaseRole, currentUser, qualified)
		}
		if canUpdate || canDelete || canTruncate {
			return fmt.Errorf("%w: principal %s has mutable ledger capability on %s", ErrUnsafeRuntimeDatabaseRole, currentUser, qualified)
		}
	}
	return nil
}
