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

type runtimeRelationCapability struct {
	Schema string
	Name   string
	Select bool
	Insert bool
	Update bool
	Delete bool
}

var runtimeRelationCapabilities = []runtimeRelationCapability{
	{Schema: "identity", Name: "users", Select: true, Insert: true},
	{Schema: "identity", Name: "user_investment_links", Select: true, Insert: true},
	{Schema: "identity", Name: "credentials", Select: true, Insert: true},
	{Schema: "identity", Name: "privacy_settings", Select: true, Insert: true},
	{Schema: "identity", Name: "sessions", Select: true, Insert: true, Update: true, Delete: true},
	{Schema: "investment", Name: "subjects", Insert: true},
	{Schema: "investment", Name: "assets", Select: true, Insert: true},
	{Schema: "investment", Name: "portfolios", Select: true, Insert: true},
	{Schema: "investment", Name: "transaction_entries", Select: true, Insert: true},
	{Schema: "investment", Name: "command_deduplication", Select: true, Insert: true, Update: true, Delete: true},
	{Schema: "investment", Name: "outbox_events"},
	{Schema: "investment", Name: "portfolio_manual_valuations", Select: true, Insert: true, Update: true, Delete: true},
	{Schema: "analytics", Name: "portfolio_snapshots", Select: true, Insert: true},
	{Schema: "analytics", Name: "snapshot_positions"},
	{Schema: "analytics", Name: "calculation_runs"},
	{Schema: "analytics", Name: "inbox_messages"},
	{Schema: "audit", Name: "actors", Insert: true},
	{Schema: "audit", Name: "events", Select: true, Insert: true},
	{Schema: "audit", Name: "auth_security_event_deduplications", Insert: true},
}

var runtimeSchemas = []string{"identity", "investment", "analytics", "audit"}

// Some production INSERT ... ON CONFLICT statements need SELECT only on their conflict-target
// column. These are deliberately column-scoped exceptions; they do not authorize table SELECT.
var runtimeColumnSelectCapabilities = map[string]map[string]struct{}{
	"investment.subjects":                      {"id": {}},
	"audit.actors":                             {"id": {}},
	"audit.auth_security_event_deduplications": {"action_code": {}, "session_id": {}},
}

// SELECT ... FOR UPDATE requires UPDATE privilege on at least one column. The API does not perform
// a business UPDATE of investment.portfolios, so this is a deliberately column-scoped lock capability.
var runtimeColumnUpdateCapabilities = map[string]map[string]struct{}{
	"investment.portfolios": {"portfolio_state": {}},
}

// OpenRuntime opens the application store and proves that the authenticated PostgreSQL LOGIN stays
// inside the maximum authorized runtime capability envelope. The registry describes capabilities, not
// the complete set of database objects allowed to exist: additive Expand-phase relations/sequences with
// zero runtime capability are compatible with old application versions. Any effective capability on an
// unknown user-schema object, any dangerous default ACL, ownership path, grant option, user-schema CREATE,
// predefined pg_* role path, explicit parameter ACL, non-origin replication state, database CREATE/ownership,
// or executable SECURITY DEFINER routine remains fail-closed. Every SET-reachable role is validated too.
// Migration/schema-owner connections use Open.
func OpenRuntime(databaseURL string) (*Store, error) {
	return OpenRuntimeWithCapability(databaseURL, RuntimeCapabilityR0)
}

func OpenRuntimeWithCapability(databaseURL string, profile RuntimeCapabilityProfile) (*Store, error) {
	if !profile.valid() {
		return nil, fmt.Errorf("%w: unknown expected runtime capability profile %q", ErrUnsafeRuntimeDatabaseRole, profile)
	}
	store, err := Open(databaseURL)
	if err != nil {
		return nil, err
	}
	store.runtimeCapabilityProfile = profile

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := store.ValidateRuntimePrivileges(ctx); err != nil {
		_ = store.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) ValidateRuntimePrivileges(ctx context.Context) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("%w: database store is not initialized", ErrUnsafeRuntimeDatabaseRole)
	}

	profile := s.runtimeCapabilityProfile
	if !profile.valid() {
		return fmt.Errorf("%w: invalid expected runtime capability profile %q", ErrUnsafeRuntimeDatabaseRole, profile)
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
		return fmt.Errorf("%w: authenticated principal %s is masked by effective role %s", ErrUnsafeRuntimeDatabaseRole, sessionUser, currentUser)
	}
	if err := validateRuntimeSessionState(ctx, conn); err != nil {
		return err
	}

	if err := validateRuntimeRoleAttributes(ctx, conn, sessionUser, "authenticated principal"); err != nil {
		return err
	}
	if err := validateNoPredefinedRoleCapabilities(ctx, conn, sessionUser, "authenticated principal"); err != nil {
		return err
	}
	if err := validateNoRuntimeParameterPrivileges(ctx, conn, sessionUser, "authenticated principal"); err != nil {
		return err
	}
	if err := validateRuntimeRoleCapabilities(ctx, conn, sessionUser, profile, true, "authenticated principal"); err != nil {
		return err
	}
	if err := validateNoRoleAdministration(ctx, conn, sessionUser, "authenticated principal"); err != nil {
		return err
	}
	if err := validateNoRuntimeDefaultPrivileges(ctx, conn, sessionUser, "authenticated principal"); err != nil {
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
		if err := validateNoPredefinedRoleCapabilities(ctx, conn, roleName, "SET-reachable role"); err != nil {
			return err
		}
		if err := validateNoRuntimeParameterPrivileges(ctx, conn, roleName, "SET-reachable role"); err != nil {
			return err
		}
		if err := validateRuntimeRoleCapabilities(ctx, conn, roleName, profile, false, "SET-reachable role"); err != nil {
			return err
		}
		if err := validateNoRoleAdministration(ctx, conn, roleName, "SET-reachable role"); err != nil {
			return err
		}
		if err := validateNoRuntimeDefaultPrivileges(ctx, conn, roleName, "SET-reachable role"); err != nil {
			return err
		}
	}

	return nil
}

func validateRuntimeRoleAttributes(ctx context.Context, conn *sql.Conn, roleName string, roleKind string) error {
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
			ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, superuser, createDB, createRole, replication, bypassRLS,
		)
	}
	return nil
}

func validateRuntimeRoleCapabilities(ctx context.Context, conn *sql.Conn, roleName string, profile RuntimeCapabilityProfile, requireExact bool, roleKind string) error {
	if err := validateRuntimeDatabase(ctx, conn, roleName, requireExact, roleKind); err != nil {
		return err
	}
	if err := validateRuntimeSchemas(ctx, conn, roleName, requireExact, roleKind); err != nil {
		return err
	}
	if err := validateRuntimeRelations(ctx, conn, roleName, profile, requireExact, roleKind); err != nil {
		return err
	}
	if err := validateRuntimeColumns(ctx, conn, roleName, profile, requireExact, roleKind); err != nil {
		return err
	}
	if err := validateRuntimeSequences(ctx, conn, roleName, roleKind); err != nil {
		return err
	}
	if err := validateRuntimeSecurityDefinerRoutines(ctx, conn, roleName, roleKind); err != nil {
		return err
	}
	return nil
}

func validateRuntimeSchemas(ctx context.Context, conn *sql.Conn, roleName string, requireExact bool, roleKind string) error {
	rows, err := conn.QueryContext(ctx, `
		SELECT
			n.nspname::text,
			($1::name = owner_role.rolname::name) OR pg_has_role($1::name, n.nspowner, 'MEMBER'),
			has_schema_privilege($1::name, n.oid, 'USAGE'),
			has_schema_privilege($1::name, n.oid, 'CREATE'),
			has_schema_privilege($1::name, n.oid, 'USAGE WITH GRANT OPTION'),
			has_schema_privilege($1::name, n.oid, 'CREATE WITH GRANT OPTION')
		FROM pg_namespace n
		JOIN pg_roles owner_role ON owner_role.oid = n.nspowner
		WHERE n.nspname <> 'pg_catalog'
			AND n.nspname <> 'information_schema'
			AND n.nspname NOT LIKE 'pg_toast%'
			AND n.nspname NOT LIKE 'pg_temp_%'
		ORDER BY n.nspname
	`, roleName)
	if err != nil {
		return fmt.Errorf("%w: inspect %s %s schema privileges: %v", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, err)
	}
	defer rows.Close()

	knownSchemas := make(map[string]struct{}, len(runtimeSchemas))
	for _, schema := range runtimeSchemas {
		knownSchemas[schema] = struct{}{}
	}
	seen := map[string]bool{}
	for rows.Next() {
		var schema string
		var owns, usage, create, usageGrant, createGrant bool
		if err := rows.Scan(&schema, &owns, &usage, &create, &usageGrant, &createGrant); err != nil {
			return fmt.Errorf("%w: scan %s %s schema privileges: %v", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, err)
		}
		seen[schema] = true
		if owns {
			return fmt.Errorf("%w: %s %s owns or is a member of the owner role for user schema %s", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, schema)
		}
		if create || usageGrant || createGrant {
			return fmt.Errorf("%w: %s %s has forbidden schema capability on %s (create=%t usage_grant=%t create_grant=%t)", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, schema, create, usageGrant, createGrant)
		}
		_, known := knownSchemas[schema]
		if requireExact && known && !usage {
			return fmt.Errorf("%w: %s %s lacks required USAGE on schema %s", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, schema)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("%w: read %s %s schema privileges: %v", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, err)
	}
	for _, schema := range runtimeSchemas {
		if !seen[schema] {
			return fmt.Errorf("%w: runtime schema %s is unavailable", ErrUnsafeRuntimeDatabaseRole, schema)
		}
	}
	return nil
}

func validateRuntimeRelations(ctx context.Context, conn *sql.Conn, roleName string, profile RuntimeCapabilityProfile, requireExact bool, roleKind string) error {
	capabilities := runtimeRelationCapabilitiesForProfile(profile)
	expected := make(map[string]runtimeRelationCapability, len(capabilities))
	for _, capability := range capabilities {
		expected[capability.Schema+"."+capability.Name] = capability
	}

	rows, err := conn.QueryContext(ctx, `
		SELECT
			n.nspname::text,
			c.relname::text,
			($1::name = owner_role.rolname::name) OR pg_has_role($1::name, c.relowner, 'MEMBER'),
			has_table_privilege($1::name, c.oid, 'SELECT'),
			has_table_privilege($1::name, c.oid, 'INSERT'),
			has_table_privilege($1::name, c.oid, 'UPDATE'),
			has_table_privilege($1::name, c.oid, 'DELETE'),
			has_table_privilege($1::name, c.oid, 'TRUNCATE'),
			has_table_privilege($1::name, c.oid, 'REFERENCES'),
			has_table_privilege($1::name, c.oid, 'TRIGGER'),
			CASE WHEN current_setting('server_version_num')::int >= 170000
				THEN has_table_privilege($1::name, c.oid, 'MAINTAIN') ELSE false END,
			has_table_privilege($1::name, c.oid, 'SELECT WITH GRANT OPTION'),
			has_table_privilege($1::name, c.oid, 'INSERT WITH GRANT OPTION'),
			has_table_privilege($1::name, c.oid, 'UPDATE WITH GRANT OPTION'),
			has_table_privilege($1::name, c.oid, 'DELETE WITH GRANT OPTION'),
			has_table_privilege($1::name, c.oid, 'TRUNCATE WITH GRANT OPTION'),
			has_table_privilege($1::name, c.oid, 'REFERENCES WITH GRANT OPTION'),
			has_table_privilege($1::name, c.oid, 'TRIGGER WITH GRANT OPTION'),
			CASE WHEN current_setting('server_version_num')::int >= 170000
				THEN has_table_privilege($1::name, c.oid, 'MAINTAIN WITH GRANT OPTION') ELSE false END
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		JOIN pg_roles owner_role ON owner_role.oid = c.relowner
		WHERE n.nspname <> 'pg_catalog'
			AND n.nspname <> 'information_schema'
			AND n.nspname NOT LIKE 'pg_toast%'
			AND n.nspname NOT LIKE 'pg_temp_%'
			AND c.relkind IN ('r', 'p', 'v', 'm', 'f')
		ORDER BY n.nspname, c.relname
	`, roleName)
	if err != nil {
		return fmt.Errorf("%w: inspect %s %s relation privileges: %v", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, err)
	}
	defer rows.Close()

	seen := map[string]bool{}
	for rows.Next() {
		var schema, relation string
		var owns bool
		var sel, ins, upd, del, trunc, refs, trigger, maintain bool
		var selGrant, insGrant, updGrant, delGrant, truncGrant, refsGrant, triggerGrant, maintainGrant bool
		if err := rows.Scan(
			&schema, &relation, &owns,
			&sel, &ins, &upd, &del, &trunc, &refs, &trigger, &maintain,
			&selGrant, &insGrant, &updGrant, &delGrant, &truncGrant, &refsGrant, &triggerGrant, &maintainGrant,
		); err != nil {
			return fmt.Errorf("%w: scan %s %s relation privileges: %v", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, err)
		}
		key := schema + "." + relation
		capability, known := expected[key]
		seen[key] = true
		if owns {
			return fmt.Errorf("%w: %s %s owns or is a member of the owner role for %s", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, key)
		}
		if trunc || refs || trigger || maintain || selGrant || insGrant || updGrant || delGrant || truncGrant || refsGrant || triggerGrant || maintainGrant {
			return fmt.Errorf("%w: %s %s has forbidden relation/grant-option capability on %s", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, key)
		}
		if !known {
			if sel || ins || upd || del {
				return fmt.Errorf(
					"%w: %s %s has unauthorized capability on unknown relation %s (select=%t insert=%t update=%t delete=%t)",
					ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, key, sel, ins, upd, del,
				)
			}
			continue
		}
		if requireExact {
			if sel != capability.Select || ins != capability.Insert || upd != capability.Update || del != capability.Delete {
				return fmt.Errorf("%w: %s %s capability mismatch on %s (select=%t insert=%t update=%t delete=%t)", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, key, sel, ins, upd, del)
			}
		} else if (sel && !capability.Select) || (ins && !capability.Insert) || (upd && !capability.Update) || (del && !capability.Delete) {
			return fmt.Errorf("%w: %s %s exceeds allowed capability on %s", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, key)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("%w: read %s %s relation privileges: %v", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, err)
	}
	for key := range expected {
		if !seen[key] {
			return fmt.Errorf("%w: required runtime relation %s is unavailable", ErrUnsafeRuntimeDatabaseRole, key)
		}
	}
	return nil
}

func validateRuntimeColumns(ctx context.Context, conn *sql.Conn, roleName string, profile RuntimeCapabilityProfile, requireExact bool, roleKind string) error {
	capabilities := runtimeRelationCapabilitiesForProfile(profile)
	expected := make(map[string]runtimeRelationCapability, len(capabilities))
	for _, capability := range capabilities {
		expected[capability.Schema+"."+capability.Name] = capability
	}

	rows, err := conn.QueryContext(ctx, `
		SELECT
			n.nspname::text,
			c.relname::text,
			a.attname::text,
			has_column_privilege($1::name, c.oid, a.attnum, 'SELECT'),
			has_column_privilege($1::name, c.oid, a.attnum, 'INSERT'),
			has_column_privilege($1::name, c.oid, a.attnum, 'UPDATE'),
			has_column_privilege($1::name, c.oid, a.attnum, 'REFERENCES'),
			has_column_privilege($1::name, c.oid, a.attnum, 'SELECT WITH GRANT OPTION'),
			has_column_privilege($1::name, c.oid, a.attnum, 'INSERT WITH GRANT OPTION'),
			has_column_privilege($1::name, c.oid, a.attnum, 'UPDATE WITH GRANT OPTION'),
			has_column_privilege($1::name, c.oid, a.attnum, 'REFERENCES WITH GRANT OPTION')
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		JOIN pg_attribute a ON a.attrelid = c.oid AND a.attnum > 0 AND NOT a.attisdropped
		WHERE n.nspname <> 'pg_catalog'
			AND n.nspname <> 'information_schema'
			AND n.nspname NOT LIKE 'pg_toast%'
			AND n.nspname NOT LIKE 'pg_temp_%'
			AND c.relkind IN ('r', 'p', 'v', 'm', 'f')
		ORDER BY n.nspname, c.relname, a.attnum
	`, roleName)
	if err != nil {
		return fmt.Errorf("%w: inspect %s %s column privileges: %v", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, err)
	}
	defer rows.Close()

	for rows.Next() {
		var schema, relation, column string
		var sel, ins, upd, refs bool
		var selGrant, insGrant, updGrant, refsGrant bool
		if err := rows.Scan(&schema, &relation, &column, &sel, &ins, &upd, &refs, &selGrant, &insGrant, &updGrant, &refsGrant); err != nil {
			return fmt.Errorf("%w: scan %s %s column privileges: %v", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, err)
		}
		key := schema + "." + relation
		capability, known := expected[key]
		expectedSelect := false
		expectedInsert := false
		expectedUpdate := false
		if known {
			expectedSelect = capability.Select || hasRuntimeColumnSelectCapability(key, column)
			expectedInsert = capability.Insert
			expectedUpdate = capability.Update || hasRuntimeColumnUpdateCapability(key, column)
		}
		if refs || selGrant || insGrant || updGrant || refsGrant {
			return fmt.Errorf("%w: %s %s has forbidden column/grant-option capability on %s.%s", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, key, column)
		}
		if requireExact {
			if sel != expectedSelect || ins != expectedInsert || upd != expectedUpdate {
				return fmt.Errorf("%w: %s %s column capability mismatch on %s.%s (select=%t insert=%t update=%t)", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, key, column, sel, ins, upd)
			}
		} else if (sel && !expectedSelect) || (ins && !expectedInsert) || (upd && !expectedUpdate) {
			return fmt.Errorf("%w: %s %s exceeds allowed column capability on %s.%s", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, key, column)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("%w: read %s %s column privileges: %v", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, err)
	}
	return nil
}

func validateRuntimeSequences(ctx context.Context, conn *sql.Conn, roleName string, roleKind string) error {
	rows, err := conn.QueryContext(ctx, `
		SELECT
			n.nspname::text,
			c.relname::text,
			($1::name = owner_role.rolname::name) OR pg_has_role($1::name, c.relowner, 'MEMBER'),
			has_sequence_privilege($1::name, c.oid, 'USAGE'),
			has_sequence_privilege($1::name, c.oid, 'SELECT'),
			has_sequence_privilege($1::name, c.oid, 'UPDATE'),
			has_sequence_privilege($1::name, c.oid, 'USAGE WITH GRANT OPTION'),
			has_sequence_privilege($1::name, c.oid, 'SELECT WITH GRANT OPTION'),
			has_sequence_privilege($1::name, c.oid, 'UPDATE WITH GRANT OPTION')
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		JOIN pg_roles owner_role ON owner_role.oid = c.relowner
		WHERE n.nspname <> 'pg_catalog'
			AND n.nspname <> 'information_schema'
			AND n.nspname NOT LIKE 'pg_toast%'
			AND n.nspname NOT LIKE 'pg_temp_%'
			AND c.relkind = 'S'
		ORDER BY n.nspname, c.relname
	`, roleName)
	if err != nil {
		return fmt.Errorf("%w: inspect %s %s sequence privileges: %v", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, err)
	}
	defer rows.Close()
	for rows.Next() {
		var schema, sequence string
		var owns bool
		var usage, sel, upd, usageGrant, selGrant, updGrant bool
		if err := rows.Scan(&schema, &sequence, &owns, &usage, &sel, &upd, &usageGrant, &selGrant, &updGrant); err != nil {
			return fmt.Errorf("%w: scan %s %s sequence privileges: %v", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, err)
		}
		key := schema + "." + sequence
		if owns {
			return fmt.Errorf("%w: %s %s owns or is a member of the owner role for unknown/future sequence %s", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, key)
		}
		if usage || sel || upd || usageGrant || selGrant || updGrant {
			return fmt.Errorf(
				"%w: %s %s has unauthorized capability on unknown/future sequence %s (usage=%t select=%t update=%t usage_grant=%t select_grant=%t update_grant=%t)",
				ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, key, usage, sel, upd, usageGrant, selGrant, updGrant,
			)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("%w: read %s %s sequence privileges: %v", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, err)
	}
	return nil
}

func validateNoRuntimeDefaultPrivileges(ctx context.Context, conn *sql.Conn, roleName string, roleKind string) error {
	var ownerName, schemaName, objectType, granteeName, privilegeType string
	var grantable bool
	err := conn.QueryRowContext(ctx, `
		SELECT
			owner_role.rolname::text,
			COALESCE(namespace_row.nspname::text, '*'),
			CASE default_acl.defaclobjtype WHEN 'r' THEN 'relations' WHEN 'S' THEN 'sequences' ELSE default_acl.defaclobjtype::text END,
			COALESCE(grantee_role.rolname::text, 'PUBLIC'),
			exploded.privilege_type,
			exploded.is_grantable
		FROM pg_default_acl default_acl
		JOIN pg_roles owner_role ON owner_role.oid = default_acl.defaclrole
		LEFT JOIN pg_namespace namespace_row ON namespace_row.oid = default_acl.defaclnamespace
		CROSS JOIN LATERAL aclexplode(default_acl.defaclacl) exploded
		LEFT JOIN pg_roles grantee_role ON grantee_role.oid = exploded.grantee
		WHERE default_acl.defaclobjtype IN ('r', 'S')
			AND (
				default_acl.defaclnamespace = 0
				OR (
					namespace_row.nspname <> 'pg_catalog'
					AND namespace_row.nspname <> 'information_schema'
					AND namespace_row.nspname NOT LIKE 'pg_toast%'
					AND namespace_row.nspname NOT LIKE 'pg_temp_%'
				)
			)
			AND (
				exploded.grantee = 0
				OR exploded.grantee = (SELECT oid FROM pg_roles WHERE rolname = $1)
				OR pg_has_role($1::name, NULLIF(exploded.grantee, 0), 'MEMBER')
			)
		ORDER BY 1, 2, 3, 4, 5
		LIMIT 1
	`, roleName).Scan(&ownerName, &schemaName, &objectType, &granteeName, &privilegeType, &grantable)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("%w: inspect %s %s default privileges: %v", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, err)
	}
	return fmt.Errorf(
		"%w: %s %s is covered by default privilege %s on future %s (owner=%s schema=%s grantee=%s grantable=%t)",
		ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, privilegeType, objectType, ownerName, schemaName, granteeName, grantable,
	)
}

func validateRuntimeSessionState(ctx context.Context, conn *sql.Conn) error {
	var replicationRole string
	if err := conn.QueryRowContext(ctx, `SELECT current_setting('session_replication_role')`).Scan(&replicationRole); err != nil {
		return fmt.Errorf("%w: inspect session_replication_role: %v", ErrUnsafeRuntimeDatabaseRole, err)
	}
	if strings.TrimSpace(replicationRole) != "origin" {
		return fmt.Errorf("%w: session_replication_role must be origin, got %q", ErrUnsafeRuntimeDatabaseRole, replicationRole)
	}
	return nil
}

func validateNoPredefinedRoleCapabilities(ctx context.Context, conn *sql.Conn, roleName string, roleKind string) error {
	var predefinedRole string
	var member, settable bool
	err := conn.QueryRowContext(ctx, `
		SELECT target.rolname::text,
			pg_has_role($1::name, target.oid, 'MEMBER'),
			pg_has_role($1::name, target.oid, 'SET')
		FROM pg_roles target
		WHERE target.rolname LIKE 'pg\_%' ESCAPE '\'
			AND (
				pg_has_role($1::name, target.oid, 'MEMBER')
				OR pg_has_role($1::name, target.oid, 'SET')
			)
		ORDER BY target.rolname
		LIMIT 1
	`, roleName).Scan(&predefinedRole, &member, &settable)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("%w: inspect %s %s predefined-role capabilities: %v", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, err)
	}
	return fmt.Errorf("%w: %s %s reaches PostgreSQL predefined role %s (member=%t set=%t)", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, predefinedRole, member, settable)
}

func validateNoRuntimeParameterPrivileges(ctx context.Context, conn *sql.Conn, roleName string, roleKind string) error {
	var parameter, grantee, privilege string
	var grantable bool
	err := conn.QueryRowContext(ctx, `
		SELECT parameter_acl.parname::text,
			COALESCE(grantee_role.rolname::text, 'PUBLIC'),
			exploded.privilege_type,
			exploded.is_grantable
		FROM pg_parameter_acl parameter_acl
		CROSS JOIN LATERAL aclexplode(parameter_acl.paracl) exploded
		LEFT JOIN pg_roles grantee_role ON grantee_role.oid = exploded.grantee
		WHERE exploded.grantee = 0
			OR exploded.grantee = (SELECT oid FROM pg_roles WHERE rolname = $1)
			OR pg_has_role($1::name, NULLIF(exploded.grantee, 0), 'MEMBER')
		ORDER BY parameter_acl.parname, exploded.privilege_type
		LIMIT 1
	`, roleName).Scan(&parameter, &grantee, &privilege, &grantable)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("%w: inspect %s %s parameter ACL: %v", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, err)
	}
	return fmt.Errorf("%w: %s %s has explicit PostgreSQL parameter capability %s on %s via %s (grantable=%t)", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, privilege, parameter, grantee, grantable)
}

func validateRuntimeDatabase(ctx context.Context, conn *sql.Conn, roleName string, requireConnect bool, roleKind string) error {
	var database string
	var owns, connect, create, connectGrant, createGrant bool
	if err := conn.QueryRowContext(ctx, `
		SELECT db.datname::text,
			($1::name = owner_role.rolname::name) OR pg_has_role($1::name, db.datdba, 'MEMBER'),
			has_database_privilege($1::name, db.oid, 'CONNECT'),
			has_database_privilege($1::name, db.oid, 'CREATE'),
			has_database_privilege($1::name, db.oid, 'CONNECT WITH GRANT OPTION'),
			has_database_privilege($1::name, db.oid, 'CREATE WITH GRANT OPTION')
		FROM pg_database db
		JOIN pg_roles owner_role ON owner_role.oid = db.datdba
		WHERE db.datname = current_database()
	`, roleName).Scan(&database, &owns, &connect, &create, &connectGrant, &createGrant); err != nil {
		return fmt.Errorf("%w: inspect %s %s current-database privileges: %v", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, err)
	}
	if owns {
		return fmt.Errorf("%w: %s %s owns or is a member of the owner role for current database %s", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, database)
	}
	if create || connectGrant || createGrant {
		return fmt.Errorf("%w: %s %s has forbidden current-database capability on %s (create=%t connect_grant=%t create_grant=%t)", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, database, create, connectGrant, createGrant)
	}
	if requireConnect && !connect {
		return fmt.Errorf("%w: authenticated principal %s lacks CONNECT on current database %s", ErrUnsafeRuntimeDatabaseRole, roleName, database)
	}
	return nil
}

func validateRuntimeSecurityDefinerRoutines(ctx context.Context, conn *sql.Conn, roleName string, roleKind string) error {
	rows, err := conn.QueryContext(ctx, `
		SELECT n.nspname::text,
			p.proname::text,
			p.oid::bigint,
			owner_role.rolname::text,
			($1::name = owner_role.rolname::name) OR pg_has_role($1::name, p.proowner, 'MEMBER'),
			has_function_privilege($1::name, p.oid, 'EXECUTE'),
			has_function_privilege($1::name, p.oid, 'EXECUTE WITH GRANT OPTION')
		FROM pg_proc p
		JOIN pg_namespace n ON n.oid = p.pronamespace
		JOIN pg_roles owner_role ON owner_role.oid = p.proowner
		WHERE p.prosecdef
			AND n.nspname <> 'pg_catalog'
			AND n.nspname <> 'information_schema'
			AND n.nspname NOT LIKE 'pg_toast%'
			AND n.nspname NOT LIKE 'pg_temp_%'
		ORDER BY n.nspname, p.proname, p.oid
	`, roleName)
	if err != nil {
		return fmt.Errorf("%w: inspect %s %s SECURITY DEFINER routines: %v", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, err)
	}
	defer rows.Close()
	for rows.Next() {
		var schema, routine, owner string
		var oid int64
		var owns, execute, executeGrant bool
		if err := rows.Scan(&schema, &routine, &oid, &owner, &owns, &execute, &executeGrant); err != nil {
			return fmt.Errorf("%w: scan %s %s SECURITY DEFINER routine capability: %v", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, err)
		}
		if owns || execute || executeGrant {
			return fmt.Errorf("%w: %s %s has privileged SECURITY DEFINER path %s.%s oid=%d owner=%s (owner_path=%t execute=%t execute_grant=%t)", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, schema, routine, oid, owner, owns, execute, executeGrant)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("%w: read %s %s SECURITY DEFINER routine capabilities: %v", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, err)
	}
	return nil
}

func validateNoRoleAdministration(ctx context.Context, conn *sql.Conn, roleName string, roleKind string) error {
	var administrableRole string
	err := conn.QueryRowContext(ctx, `
		SELECT target_role.rolname::text
		FROM pg_roles target_role
		WHERE target_role.rolname <> $1::name
			AND pg_has_role($1::name, target_role.oid, 'MEMBER WITH ADMIN OPTION')
		ORDER BY target_role.rolname
		LIMIT 1
	`, roleName).Scan(&administrableRole)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("%w: inspect %s %s role administration capability: %v", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, err)
	}
	return fmt.Errorf("%w: %s %s can administer PostgreSQL role membership for %s", ErrUnsafeRuntimeDatabaseRole, roleKind, roleName, strings.TrimSpace(administrableRole))
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
		if roleName = strings.TrimSpace(roleName); roleName != "" {
			roles = append(roles, roleName)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: read SET-reachable roles for %s: %v", ErrUnsafeRuntimeDatabaseRole, sessionUser, err)
	}
	return roles, nil
}

func hasRuntimeColumnSelectCapability(relation string, column string) bool {
	columns, ok := runtimeColumnSelectCapabilities[relation]
	if !ok {
		return false
	}
	_, ok = columns[column]
	return ok
}

func hasRuntimeColumnUpdateCapability(relation string, column string) bool {
	columns, ok := runtimeColumnUpdateCapabilities[relation]
	if !ok {
		return false
	}
	_, ok = columns[column]
	return ok
}

// pqTextArray returns a PostgreSQL text[] literal. The values are compile-time schema names, so this
// helper is deliberately small and avoids adding a driver-specific array dependency to validation.
func pqTextArray(values []string) string {
	quoted := make([]string, 0, len(values))
	for _, value := range values {
		quoted = append(quoted, `"`+strings.ReplaceAll(value, `"`, `\"`)+`"`)
	}
	return `{` + strings.Join(quoted, `,`) + `}`
}
