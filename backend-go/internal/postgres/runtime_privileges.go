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

var runtimeSchemas = []string{"identity", "investment", "analytics", "audit"}

// OpenRuntime opens the application store and verifies that the authenticated database LOGIN and
// every role it can enter with SET ROLE are incapable of mutating the protected append-only tables.
// Migration/schema-owner connections must use Open instead.
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

// ValidateRuntimePrivileges proves the runtime boundary against PostgreSQL's authenticated session
// principal, not merely the current effective role. The complete check runs on one physical database
// connection so session_user/current_user and all capability queries describe the same session.
func (s *Store) ValidateRuntimePrivileges(ctx context.Context) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("%w: database store is not initialized", ErrUnsafeRuntimeDatabaseRole)
	}

	conn, err := s.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("%w: acquire validation connection: %v", ErrUnsafeRuntimeDatabaseRole, err)
	}
	defer conn.Close()

	var sessionUser string
	var currentUser string
	if err := conn.QueryRowContext(ctx, `SELECT session_user::text, current_user::text`).Scan(&sessionUser, &currentUser); err != nil {
		return fmt.Errorf("%w: inspect database session identity: %v", ErrUnsafeRuntimeDatabaseRole, err)
	}
	sessionUser = strings.TrimSpace(sessionUser)
	currentUser = strings.TrimSpace(currentUser)
	if sessionUser == "" || currentUser == "" {
		return fmt.Errorf("%w: database session identity is empty", ErrUnsafeRuntimeDatabaseRole)
	}
	if sessionUser != currentUser {
		return fmt.Errorf(
			"%w: authenticated principal %s is masked by effective role %s",
			ErrUnsafeRuntimeDatabaseRole,
			sessionUser,
			currentUser,
		)
	}

	if err := validateRuntimeRoleAttributes(ctx, conn, sessionUser, "authenticated principal"); err != nil {
		return err
	}
	if err := validateRuntimeRoleCapabilities(ctx, conn, sessionUser, true, "authenticated principal"); err != nil {
		return err
	}

	setReachableRoles, err := listSetReachableRoles(ctx, conn, sessionUser)
	if err != nil {
		return err
	}
	for _, roleName := range setReachableRoles {
		if err := validateRuntimeRoleAttributes(ctx, conn, roleName, "SET-reachable role"); err != nil {
			return err
		}
		if err := validateRuntimeRoleCapabilities(ctx, conn, roleName, false, "SET-reachable role"); err != nil {
			return err
		}
	}

	return nil
}

func validateRuntimeRoleAttributes(
	ctx context.Context,
	conn *sql.Conn,
	roleName string,
	roleKind string,
) error {
	var superuser bool
	var createDB bool
	var createRole bool
	var replication bool
	var bypassRLS bool
	if err := conn.QueryRowContext(ctx, `
		SELECT rolsuper, rolcreatedb, rolcreaterole, rolreplication, rolbypassrls
		FROM pg_roles
		WHERE rolname = $1
	`, roleName).Scan(&superuser, &createDB, &createRole, &replication, &bypassRLS); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: %s %s is unavailable", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName)
		}
		return fmt.Errorf("%w: inspect %s %s attributes: %v", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, err)
	}

	if superuser || createDB || createRole || replication || bypassRLS {
		return fmt.Errorf(
			"%w: %s %s has elevated role attributes (superuser=%t createdb=%t createrole=%t replication=%t bypassrls=%t)",
			ErrUnsafeRuntimeDatabaseRole,
			roleKind,
			roleName,
			superuser,
			createDB,
			createRole,
			replication,
			bypassRLS,
		)
	}
	return nil
}

func validateRuntimeRoleCapabilities(
	ctx context.Context,
	conn *sql.Conn,
	roleName string,
	requireAppendRead bool,
	roleKind string,
) error {
	for _, schemaName := range runtimeSchemas {
		var canCreate bool
		if err := conn.QueryRowContext(ctx, `
			SELECT has_schema_privilege($1::name, $2, 'CREATE')
		`, roleName, schemaName).Scan(&canCreate); err != nil {
			return fmt.Errorf("%w: inspect %s %s schema %s privileges: %v", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, schemaName, err)
		}
		if canCreate {
			return fmt.Errorf("%w: %s %s can CREATE in protected runtime schema %s", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, schemaName)
		}
	}

	for _, protected := range protectedAppendOnlyRuntimeTables {
		qualified := protected.Schema + "." + protected.Table
		var ownsOrIsMemberOfOwner bool
		var schemaUsage bool
		var canSelect bool
		var canInsert bool
		var canUpdate bool
		var canDelete bool
		var canTruncate bool
		var canTrigger bool

		err := conn.QueryRowContext(ctx, `
			SELECT
				($1::name = pg_get_userbyid(table_row.relowner))
					OR pg_has_role($1::name, table_row.relowner, 'MEMBER'),
				has_schema_privilege($1::name, $3, 'USAGE'),
				has_table_privilege($1::name, $2, 'SELECT'),
				has_table_privilege($1::name, $2, 'INSERT'),
				has_table_privilege($1::name, $2, 'UPDATE'),
				has_table_privilege($1::name, $2, 'DELETE'),
				has_table_privilege($1::name, $2, 'TRUNCATE'),
				has_table_privilege($1::name, $2, 'TRIGGER')
			FROM pg_class table_row
			WHERE table_row.oid = to_regclass($2)
		`, roleName, qualified, protected.Schema).Scan(
			&ownsOrIsMemberOfOwner,
			&schemaUsage,
			&canSelect,
			&canInsert,
			&canUpdate,
			&canDelete,
			&canTruncate,
			&canTrigger,
		)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("%w: protected table %s is unavailable", ErrUnsafeRuntimeDatabaseRole, qualified)
			}
			return fmt.Errorf("%w: inspect %s %s privileges on %s: %v", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, qualified, err)
		}

		if ownsOrIsMemberOfOwner {
			return fmt.Errorf("%w: %s %s owns or is a member of the owner role for %s", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, qualified)
		}
		if canUpdate || canDelete || canTruncate || canTrigger {
			return fmt.Errorf(
				"%w: %s %s has forbidden mutable capability on %s (update=%t delete=%t truncate=%t trigger=%t)",
				ErrUnsafeRuntimeDatabaseRole,
				roleKind,
				roleName,
				qualified,
				canUpdate,
				canDelete,
				canTruncate,
				canTrigger,
			)
		}
		if requireAppendRead && (!schemaUsage || !canSelect || !canInsert) {
			return fmt.Errorf(
				"%w: %s %s lacks required append/read capability on %s",
				ErrUnsafeRuntimeDatabaseRole,
				roleKind,
				roleName,
				qualified,
			)
		}
	}
	return nil
}

func listSetReachableRoles(ctx context.Context, conn *sql.Conn, sessionUser string) ([]string, error) {
	rows, err := conn.QueryContext(ctx, `
		SELECT role_row.rolname::text
		FROM pg_roles role_row
		WHERE role_row.rolname <> $1::name
			AND pg_has_role($1::name, role_row.oid, 'SET')
		ORDER BY role_row.rolname
	`, sessionUser)
	if err != nil {
		return nil, fmt.Errorf("%w: inspect SET-reachable roles for %s: %v", ErrUnsafeRuntimeDatabaseRole, sessionUser, err)
	}
	defer rows.Close()

	roles := []string{}
	for rows.Next() {
		var roleName string
		if err := rows.Scan(&roleName); err != nil {
			return nil, fmt.Errorf("%w: scan SET-reachable role for %s: %v", ErrUnsafeRuntimeDatabaseRole, sessionUser, err)
		}
		roleName = strings.TrimSpace(roleName)
		if roleName != "" {
			roles = append(roles, roleName)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: read SET-reachable roles for %s: %v", ErrUnsafeRuntimeDatabaseRole, sessionUser, err)
	}
	return roles, nil
}
