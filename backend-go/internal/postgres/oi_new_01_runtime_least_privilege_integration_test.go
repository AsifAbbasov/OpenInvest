package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/openinvest/openinvest/backend-go/internal/auth"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"strings"
)

func TestOINew01RuntimeCapabilityMatrix(t *testing.T) {
	runtimeURL := os.Getenv("OPENINVEST_DATABASE_RUNTIME_TEST_URL")
	if runtimeURL == "" {
		t.Skip("OPENINVEST_DATABASE_RUNTIME_TEST_URL is not set")
	}
	ctx := context.Background()
	runtimeDB, err := sql.Open("pgx", runtimeURL)
	if err != nil {
		t.Fatalf("open runtime capability db: %v", err)
	}
	t.Cleanup(func() {
		if err := runtimeDB.Close(); err != nil {
			t.Errorf("close runtime capability db: %v", err)
		}
	})

	type capability struct {
		relation string
		selectOK bool
		insertOK bool
		updateOK bool
		deleteOK bool
	}
	expected := []capability{
		{relation: "identity.users", selectOK: true, insertOK: true},
		{relation: "identity.user_investment_links", selectOK: true, insertOK: true},
		{relation: "identity.credentials", selectOK: true, insertOK: true},
		{relation: "identity.privacy_settings", selectOK: true, insertOK: true},
		{relation: "identity.sessions", selectOK: true, insertOK: true, updateOK: true, deleteOK: true},
		{relation: "investment.subjects", insertOK: true},
		{relation: "investment.assets", selectOK: true, insertOK: true},
		{relation: "investment.portfolios", selectOK: true, insertOK: true},
		{relation: "investment.transaction_entries", selectOK: true, insertOK: true},
		{relation: "investment.command_deduplication", selectOK: true, insertOK: true, updateOK: true, deleteOK: true},
		{relation: "investment.outbox_events"},
		{relation: "investment.portfolio_manual_valuations", selectOK: true, insertOK: true, updateOK: true, deleteOK: true},
		{relation: "analytics.portfolio_snapshots", selectOK: true, insertOK: true},
		{relation: "analytics.snapshot_positions"},
		{relation: "analytics.calculation_runs"},
		{relation: "analytics.inbox_messages"},
		{relation: "audit.actors", insertOK: true},
		{relation: "audit.events", selectOK: true, insertOK: true},
		{relation: "audit.auth_security_event_deduplications", insertOK: true},
	}

	// The matrix is a maximum capability registry, not a complete schema-object inventory.
	// Therefore this test validates every KNOWN object exactly but intentionally does not assert
	// that protected schemas contain exactly len(expected) relations. Expand-phase zero-capability
	// objects may coexist with the old application.
	for _, want := range expected {
		var exists bool
		var sel, ins, upd, del bool
		var truncate, references, trigger, maintain bool
		var selGrant, insGrant, updGrant, delGrant, truncateGrant, referencesGrant, triggerGrant, maintainGrant bool
		err := runtimeDB.QueryRowContext(ctx, `
			SELECT
				to_regclass($1) IS NOT NULL,
				has_table_privilege(session_user, $1, 'SELECT'),
				has_table_privilege(session_user, $1, 'INSERT'),
				has_table_privilege(session_user, $1, 'UPDATE'),
				has_table_privilege(session_user, $1, 'DELETE'),
				has_table_privilege(session_user, $1, 'TRUNCATE'),
				has_table_privilege(session_user, $1, 'REFERENCES'),
				has_table_privilege(session_user, $1, 'TRIGGER'),
				has_table_privilege(session_user, $1, 'MAINTAIN'),
				has_table_privilege(session_user, $1, 'SELECT WITH GRANT OPTION'),
				has_table_privilege(session_user, $1, 'INSERT WITH GRANT OPTION'),
				has_table_privilege(session_user, $1, 'UPDATE WITH GRANT OPTION'),
				has_table_privilege(session_user, $1, 'DELETE WITH GRANT OPTION'),
				has_table_privilege(session_user, $1, 'TRUNCATE WITH GRANT OPTION'),
				has_table_privilege(session_user, $1, 'REFERENCES WITH GRANT OPTION'),
				has_table_privilege(session_user, $1, 'TRIGGER WITH GRANT OPTION'),
				has_table_privilege(session_user, $1, 'MAINTAIN WITH GRANT OPTION')
		`, want.relation).Scan(
			&exists, &sel, &ins, &upd, &del, &truncate, &references, &trigger, &maintain,
			&selGrant, &insGrant, &updGrant, &delGrant, &truncateGrant, &referencesGrant, &triggerGrant, &maintainGrant,
		)
		if err != nil {
			t.Fatalf("inspect %s capability: %v", want.relation, err)
		}
		if !exists {
			t.Fatalf("known runtime relation %s is missing", want.relation)
		}
		if sel != want.selectOK || ins != want.insertOK || upd != want.updateOK || del != want.deleteOK {
			t.Fatalf("%s capability mismatch: got S=%t I=%t U=%t D=%t, want S=%t I=%t U=%t D=%t",
				want.relation, sel, ins, upd, del, want.selectOK, want.insertOK, want.updateOK, want.deleteOK)
		}
		if truncate || references || trigger || maintain {
			t.Fatalf("%s has forbidden extra capability: truncate=%t references=%t trigger=%t maintain=%t",
				want.relation, truncate, references, trigger, maintain)
		}
		if selGrant || insGrant || updGrant || delGrant || truncateGrant || referencesGrant || triggerGrant || maintainGrant {
			t.Fatalf("%s unexpectedly has a table grant option", want.relation)
		}
	}

	var subjectIDSelect, subjectStateSelect, subjectIDGrant bool
	if err := runtimeDB.QueryRowContext(ctx, `
		SELECT
			has_column_privilege(session_user, 'investment.subjects', 'id', 'SELECT'),
			has_column_privilege(session_user, 'investment.subjects', 'subject_state', 'SELECT'),
			has_column_privilege(session_user, 'investment.subjects', 'id', 'SELECT WITH GRANT OPTION')
	`).Scan(&subjectIDSelect, &subjectStateSelect, &subjectIDGrant); err != nil {
		t.Fatalf("inspect investment.subjects column privileges: %v", err)
	}
	if !subjectIDSelect || subjectStateSelect || subjectIDGrant {
		t.Fatalf("investment.subjects column capability mismatch: id_select=%t subject_state_select=%t id_grant=%t",
			subjectIDSelect, subjectStateSelect, subjectIDGrant)
	}

	var actorIDSelect, actorKindSelect, actorIDGrant bool
	if err := runtimeDB.QueryRowContext(ctx, `
		SELECT
			has_column_privilege(session_user, 'audit.actors', 'id', 'SELECT'),
			has_column_privilege(session_user, 'audit.actors', 'actor_kind', 'SELECT'),
			has_column_privilege(session_user, 'audit.actors', 'id', 'SELECT WITH GRANT OPTION')
	`).Scan(&actorIDSelect, &actorKindSelect, &actorIDGrant); err != nil {
		t.Fatalf("inspect audit.actors column privileges: %v", err)
	}
	if !actorIDSelect || actorKindSelect || actorIDGrant {
		t.Fatalf("audit.actors column capability mismatch: id_select=%t actor_kind_select=%t id_grant=%t",
			actorIDSelect, actorKindSelect, actorIDGrant)
	}

	var dedupActionCodeSelect, dedupSessionIDSelect, dedupCreatedAtSelect, dedupActionCodeGrant bool
	if err := runtimeDB.QueryRowContext(ctx, `
		SELECT
			has_column_privilege(session_user, 'audit.auth_security_event_deduplications', 'action_code', 'SELECT'),
			has_column_privilege(session_user, 'audit.auth_security_event_deduplications', 'session_id', 'SELECT'),
			has_column_privilege(session_user, 'audit.auth_security_event_deduplications', 'created_at', 'SELECT'),
			has_column_privilege(session_user, 'audit.auth_security_event_deduplications', 'action_code', 'SELECT WITH GRANT OPTION')
	`).Scan(&dedupActionCodeSelect, &dedupSessionIDSelect, &dedupCreatedAtSelect, &dedupActionCodeGrant); err != nil {
		t.Fatalf("inspect audit deduplication column privileges: %v", err)
	}
	if !dedupActionCodeSelect || !dedupSessionIDSelect || dedupCreatedAtSelect || dedupActionCodeGrant {
		t.Fatalf("audit deduplication column capability mismatch: action_code_select=%t session_id_select=%t created_at_select=%t action_code_grant=%t",
			dedupActionCodeSelect, dedupSessionIDSelect, dedupCreatedAtSelect, dedupActionCodeGrant)
	}

	var portfolioTableUpdate, stateUpdate, nameUpdate, subjectUpdate, removedAtUpdate, stateUpdateGrant bool
	if err := runtimeDB.QueryRowContext(ctx, `
		SELECT
			has_table_privilege(session_user, 'investment.portfolios', 'UPDATE'),
			has_column_privilege(session_user, 'investment.portfolios', 'portfolio_state', 'UPDATE'),
			has_column_privilege(session_user, 'investment.portfolios', 'name', 'UPDATE'),
			has_column_privilege(session_user, 'investment.portfolios', 'subject_id', 'UPDATE'),
			has_column_privilege(session_user, 'investment.portfolios', 'removed_at', 'UPDATE'),
			has_column_privilege(session_user, 'investment.portfolios', 'portfolio_state', 'UPDATE WITH GRANT OPTION')
	`).Scan(&portfolioTableUpdate, &stateUpdate, &nameUpdate, &subjectUpdate, &removedAtUpdate, &stateUpdateGrant); err != nil {
		t.Fatalf("inspect investment.portfolios update scope: %v", err)
	}
	if portfolioTableUpdate || !stateUpdate || nameUpdate || subjectUpdate || removedAtUpdate || stateUpdateGrant {
		t.Fatalf("investment.portfolios update scope mismatch: table=%t state=%t name=%t subject=%t removed_at=%t state_grant=%t",
			portfolioTableUpdate, stateUpdate, nameUpdate, subjectUpdate, removedAtUpdate, stateUpdateGrant)
	}
}

func TestOINew01ForbiddenRuntimeOperationsAreDeniedByPostgres(t *testing.T) {
	ownerURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	runtimeURL := os.Getenv("OPENINVEST_DATABASE_RUNTIME_TEST_URL")
	if ownerURL == "" || runtimeURL == "" {
		t.Skip("runtime privilege integration URLs are not set")
	}
	ctx := context.Background()
	ownerDB, err := sql.Open("pgx", ownerURL)
	if err != nil {
		t.Fatalf("open owner db: %v", err)
	}
	t.Cleanup(func() {
		if err := ownerDB.Close(); err != nil {
			t.Errorf("close owner db: %v", err)
		}
	})
	runtimeDB, err := sql.Open("pgx", runtimeURL)
	if err != nil {
		t.Fatalf("open runtime db: %v", err)
	}
	t.Cleanup(func() {
		if err := runtimeDB.Close(); err != nil {
			t.Errorf("close runtime db: %v", err)
		}
	})

	requireDenied := func(label string, query string, args ...any) {
		t.Helper()
		_, err := runtimeDB.ExecContext(ctx, query, args...)
		if err == nil {
			t.Fatalf("%s unexpectedly succeeded", label)
		}
		var state interface{ SQLState() string }
		if !errors.As(err, &state) || state.SQLState() != "42501" {
			t.Fatalf("%s: expected SQLSTATE 42501, got %v", label, err)
		}
	}
	requireSafe := func(label string) {
		t.Helper()
		store, err := postgres.OpenRuntime(runtimeURL)
		if err != nil {
			t.Fatalf("%s: OpenRuntime unexpectedly rejected safe runtime: %v", label, err)
		}
		_ = store.Close()
	}
	requireUnsafe := func(label string) {
		t.Helper()
		store, err := postgres.OpenRuntime(runtimeURL)
		if err == nil {
			_ = store.Close()
			t.Fatalf("%s: OpenRuntime unexpectedly accepted unsafe runtime", label)
		}
		if !errors.Is(err, postgres.ErrUnsafeRuntimeDatabaseRole) {
			t.Fatalf("%s: expected ErrUnsafeRuntimeDatabaseRole, got %v", label, err)
		}
	}

	// Canonical immutable boundaries remain denied.
	requireDenied("ledger UPDATE", `UPDATE investment.transaction_entries SET note = note`)
	requireDenied("ledger DELETE", `DELETE FROM investment.transaction_entries`)
	requireDenied("audit UPDATE", `UPDATE audit.events SET action_code = action_code`)
	requireDenied("audit DELETE", `DELETE FROM audit.events`)

	// Expand-compatible future table. Pre-clean makes the fixture idempotent after interrupted prior runs.
	const futureTable = "investment.oi_new_01_future_fixture"
	if _, err := ownerDB.ExecContext(ctx, `DROP TABLE IF EXISTS `+futureTable); err != nil {
		t.Fatalf("pre-clean future table: %v", err)
	}
	if _, err := ownerDB.ExecContext(ctx, `CREATE TABLE `+futureTable+` (id uuid PRIMARY KEY, note text)`); err != nil {
		t.Fatalf("create future table: %v", err)
	}
	t.Cleanup(func() {
		if _, err := ownerDB.ExecContext(context.Background(), `DROP TABLE IF EXISTS `+futureTable); err != nil {
			t.Errorf("cleanup future table: %v", err)
		}
	})
	requireSafe("future table with zero runtime capability")
	requireDenied("future table SELECT", `SELECT id FROM `+futureTable+` LIMIT 1`)
	requireDenied("future table INSERT", `INSERT INTO `+futureTable+` (id, note) VALUES ($1, 'x')`, uuid.NewString())
	requireDenied("future table UPDATE", `UPDATE `+futureTable+` SET note = note`)
	requireDenied("future table DELETE", `DELETE FROM `+futureTable)

	for _, privilege := range []string{"SELECT", "INSERT", "UPDATE", "DELETE", "TRUNCATE", "REFERENCES", "TRIGGER", "MAINTAIN"} {
		if _, err := ownerDB.ExecContext(ctx, `GRANT `+privilege+` ON `+futureTable+` TO openinvest_runtime`); err != nil {
			t.Fatalf("grant future table %s: %v", privilege, err)
		}
		requireUnsafe("future table unauthorized " + privilege)
		if _, err := ownerDB.ExecContext(ctx, `REVOKE `+privilege+` ON `+futureTable+` FROM openinvest_runtime`); err != nil {
			t.Fatalf("revoke future table %s: %v", privilege, err)
		}
		requireSafe("future table after revoking " + privilege)
	}
	if _, err := ownerDB.ExecContext(ctx, `GRANT SELECT (id) ON `+futureTable+` TO openinvest_runtime`); err != nil {
		t.Fatalf("grant future table column SELECT: %v", err)
	}
	requireUnsafe("future table unauthorized column SELECT")
	if _, err := ownerDB.ExecContext(ctx, `REVOKE SELECT (id) ON `+futureTable+` FROM openinvest_runtime`); err != nil {
		t.Fatalf("revoke future table column SELECT: %v", err)
	}
	requireSafe("future table after revoking column SELECT")
	if _, err := ownerDB.ExecContext(ctx, `GRANT SELECT (id) ON `+futureTable+` TO openinvest_runtime WITH GRANT OPTION`); err != nil {
		t.Fatalf("grant future table column SELECT grant option: %v", err)
	}
	requireUnsafe("future table column SELECT WITH GRANT OPTION")
	if _, err := ownerDB.ExecContext(ctx, `REVOKE SELECT (id) ON `+futureTable+` FROM openinvest_runtime CASCADE`); err != nil {
		t.Fatalf("revoke future table column grant option: %v", err)
	}
	if _, err := ownerDB.ExecContext(ctx, `GRANT SELECT ON `+futureTable+` TO openinvest_runtime WITH GRANT OPTION`); err != nil {
		t.Fatalf("grant future table table-level grant option: %v", err)
	}
	requireUnsafe("future table SELECT WITH GRANT OPTION")
	if _, err := ownerDB.ExecContext(ctx, `REVOKE SELECT ON `+futureTable+` FROM openinvest_runtime CASCADE`); err != nil {
		t.Fatalf("revoke future table table-level grant option: %v", err)
	}
	requireSafe("future table final zero capability")

	// Expand-compatible future sequence follows the same maximum-capability rule.
	const futureSequence = "investment.oi_new_01_future_sequence"
	if _, err := ownerDB.ExecContext(ctx, `DROP SEQUENCE IF EXISTS `+futureSequence); err != nil {
		t.Fatalf("pre-clean future sequence: %v", err)
	}
	if _, err := ownerDB.ExecContext(ctx, `CREATE SEQUENCE `+futureSequence); err != nil {
		t.Fatalf("create future sequence: %v", err)
	}
	t.Cleanup(func() {
		if _, err := ownerDB.ExecContext(context.Background(), `DROP SEQUENCE IF EXISTS `+futureSequence); err != nil {
			t.Errorf("cleanup future sequence: %v", err)
		}
	})
	requireSafe("future sequence with zero runtime capability")
	requireDenied("future sequence nextval", `SELECT nextval('`+futureSequence+`')`)
	for _, privilege := range []string{"USAGE", "SELECT", "UPDATE"} {
		if _, err := ownerDB.ExecContext(ctx, `GRANT `+privilege+` ON SEQUENCE `+futureSequence+` TO openinvest_runtime`); err != nil {
			t.Fatalf("grant future sequence %s: %v", privilege, err)
		}
		requireUnsafe("future sequence unauthorized " + privilege)
		if _, err := ownerDB.ExecContext(ctx, `REVOKE `+privilege+` ON SEQUENCE `+futureSequence+` FROM openinvest_runtime`); err != nil {
			t.Fatalf("revoke future sequence %s: %v", privilege, err)
		}
		requireSafe("future sequence after revoking " + privilege)
	}
	if _, err := ownerDB.ExecContext(ctx, `GRANT USAGE ON SEQUENCE `+futureSequence+` TO openinvest_runtime WITH GRANT OPTION`); err != nil {
		t.Fatalf("grant future sequence grant option: %v", err)
	}
	requireUnsafe("future sequence USAGE WITH GRANT OPTION")
	if _, err := ownerDB.ExecContext(ctx, `REVOKE USAGE ON SEQUENCE `+futureSequence+` FROM openinvest_runtime CASCADE`); err != nil {
		t.Fatalf("revoke future sequence grant option: %v", err)
	}
	requireSafe("future sequence final zero capability")
}

func TestOINew01RequiredRuntimeAuthPersistenceStillWorks(t *testing.T) {
	ownerDB := openOINew01OwnerDB(t)
	runtimeURL := os.Getenv("OPENINVEST_DATABASE_RUNTIME_TEST_URL")
	runtimeStore, err := postgres.OpenRuntime(runtimeURL)
	if err != nil {
		t.Fatalf("open runtime store: %v", err)
	}
	defer runtimeStore.Close()

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)
	userID := uuid.NewString()
	subjectID := uuid.NewString()
	initialSessionID := uuid.NewString()
	email := "oi-new-01-" + uuid.NewString() + "@example.com"

	t.Cleanup(func() {
		cleanupCtx := context.Background()
		_, _ = ownerDB.ExecContext(cleanupCtx, `DELETE FROM audit.events WHERE actor_id = $1`, userID)
		_, _ = ownerDB.ExecContext(cleanupCtx, `DELETE FROM audit.actors WHERE id = $1`, userID)
		_, _ = ownerDB.ExecContext(cleanupCtx, `DELETE FROM identity.users WHERE id = $1`, userID)
		_, _ = ownerDB.ExecContext(cleanupCtx, `DELETE FROM investment.subjects WHERE id = $1`, subjectID)
	})

	record := auth.RegistrationRecord{
		UserID:              userID,
		InvestmentSubjectID: subjectID,
		EmailNormalized:     email,
		PasswordHash:        "argon2id$v=19$m=65536,t=3,p=1$AAECAwQFBgcICQoLDA0ODw$9Wc7VFOLWWIqyGrC4T85CfZU8wFkNPcB2GE1ympJUGU",
		Language:            auth.LanguageEN,
		Theme:               auth.ThemeSystem,
		Timezone:            "UTC",
		Now:                 now,
		Session: auth.SessionRecord{
			SessionID:        initialSessionID,
			UserID:           userID,
			RefreshTokenHash: "oi-new-01-register-refresh-" + uuid.NewString(),
			CSRFTokenHash:    "oi-new-01-register-csrf-" + uuid.NewString(),
			ExpiresAt:        now.Add(24 * time.Hour),
			Now:              now,
		},
	}
	if _, err := runtimeStore.RegisterUser(ctx, record); err != nil {
		t.Fatalf("runtime registration persistence: %v", err)
	}
	if _, _, err := runtimeStore.FindUserByEmail(ctx, email); err != nil {
		t.Fatalf("runtime login lookup persistence: %v", err)
	}

	loginSession := auth.SessionRecord{
		SessionID:        uuid.NewString(),
		UserID:           userID,
		RefreshTokenHash: "oi-new-01-login-refresh-" + uuid.NewString(),
		CSRFTokenHash:    "oi-new-01-login-csrf-" + uuid.NewString(),
		ExpiresAt:        now.Add(24 * time.Hour),
		Now:              now.Add(time.Second),
	}
	if err := runtimeStore.CreateSession(ctx, loginSession); err != nil {
		t.Fatalf("runtime create session: %v", err)
	}

	rotated := auth.SessionRecord{
		SessionID:        uuid.NewString(),
		RefreshTokenHash: "oi-new-01-rotated-refresh-" + uuid.NewString(),
		CSRFTokenHash:    "oi-new-01-rotated-csrf-" + uuid.NewString(),
		ExpiresAt:        now.Add(48 * time.Hour),
		Now:              now.Add(2 * time.Second),
	}
	if _, err := runtimeStore.RotateSession(
		ctx,
		loginSession.RefreshTokenHash,
		loginSession.CSRFTokenHash,
		rotated,
		now.Add(2*time.Second),
	); err != nil {
		t.Fatalf("runtime session rotation: %v", err)
	}
	revoked, err := runtimeStore.RevokeSession(
		ctx,
		rotated.RefreshTokenHash,
		rotated.CSRFTokenHash,
		false,
		now.Add(3*time.Second),
	)
	if err != nil {
		t.Fatalf("runtime logout/session revocation: %v", err)
	}
	if !revoked {
		t.Fatal("runtime logout/session revocation did not revoke the active session")
	}
}

func TestOINew01RequiredRuntimeManualValuationCRUDAndPortfolioLockWork(t *testing.T) {
	ownerDB := openOINew01OwnerDB(t)
	runtimeDB := openOINew01RuntimeDB(t)
	ctx := context.Background()

	subjectID := uuid.NewString()
	portfolioID := uuid.NewString()
	assetID := "00000000-0000-4000-8000-00000000a001"

	t.Cleanup(func() {
		cleanupCtx := context.Background()
		_, _ = ownerDB.ExecContext(cleanupCtx, `DELETE FROM investment.portfolio_manual_valuations WHERE portfolio_id = $1`, portfolioID)
		_, _ = ownerDB.ExecContext(cleanupCtx, `DELETE FROM investment.portfolios WHERE id = $1`, portfolioID)
		_, _ = ownerDB.ExecContext(cleanupCtx, `DELETE FROM investment.subjects WHERE id = $1`, subjectID)
	})

	if _, err := runtimeDB.ExecContext(ctx, `
		INSERT INTO investment.subjects (id, subject_state)
		VALUES ($1, 'active')
	`, subjectID); err != nil {
		t.Fatalf("runtime subject insert: %v", err)
	}
	if _, err := runtimeDB.ExecContext(ctx, `
		INSERT INTO investment.assets (
			id, ticker, asset_type, name, currency, market, lifecycle_status, isin, lot_size, updated_at
		)
		VALUES ($1, 'SBER', 'stock', 'Sberbank ordinary shares', 'RUB', 'MOEX', 'active', 'RU0009029540', 10, now())
		ON CONFLICT (ticker) DO NOTHING
	`, assetID); err != nil {
		t.Fatalf("runtime canonical asset insert: %v", err)
	}
	if _, err := runtimeDB.ExecContext(ctx, `
		INSERT INTO investment.portfolios (id, subject_id, name, base_currency)
		VALUES ($1, $2, 'OI-NEW-01 runtime portfolio', 'RUB')
	`, portfolioID, subjectID); err != nil {
		t.Fatalf("runtime portfolio insert: %v", err)
	}

	lockTx, err := runtimeDB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		t.Fatalf("begin portfolio lock proof: %v", err)
	}
	var lockedID string
	if err := lockTx.QueryRowContext(ctx, `
		SELECT id
		FROM investment.portfolios
		WHERE id = $1 AND subject_id = $2
		FOR UPDATE
	`, portfolioID, subjectID).Scan(&lockedID); err != nil {
		_ = lockTx.Rollback()
		t.Fatalf("runtime SELECT FOR UPDATE portfolio serialization: %v", err)
	}
	if err := lockTx.Rollback(); err != nil {
		t.Fatalf("rollback portfolio lock proof: %v", err)
	}

	if _, err := runtimeDB.ExecContext(ctx, `
		INSERT INTO investment.portfolio_manual_valuations (
			portfolio_id, asset_id, position_opened_ledger_sequence,
			price_amount, price_currency, as_of_date, source, updated_at
		)
		VALUES ($1, $2, 1, 100.00000000, 'RUB', '2026-09-15', 'USER_SUPPLIED', now())
	`, portfolioID, assetID); err != nil {
		t.Fatalf("runtime manual valuation insert: %v", err)
	}
	var price string
	if err := runtimeDB.QueryRowContext(ctx, `
		SELECT price_amount::text
		FROM investment.portfolio_manual_valuations
		WHERE portfolio_id = $1 AND asset_id = $2
	`, portfolioID, assetID).Scan(&price); err != nil {
		t.Fatalf("runtime manual valuation select: %v", err)
	}
	if _, err := runtimeDB.ExecContext(ctx, `
		UPDATE investment.portfolio_manual_valuations
		SET price_amount = 101.00000000, updated_at = now()
		WHERE portfolio_id = $1 AND asset_id = $2
	`, portfolioID, assetID); err != nil {
		t.Fatalf("runtime manual valuation update: %v", err)
	}
	if _, err := runtimeDB.ExecContext(ctx, `
		DELETE FROM investment.portfolio_manual_valuations
		WHERE portfolio_id = $1 AND asset_id = $2
	`, portfolioID, assetID); err != nil {
		t.Fatalf("runtime manual valuation delete: %v", err)
	}
}

func openOINew01OwnerDB(t *testing.T) *sql.DB {
	t.Helper()
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open OI-NEW-01 owner DB: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close OI-NEW-01 owner DB: %v", err)
		}
	})
	if err := db.Ping(); err != nil {
		t.Fatalf("ping OI-NEW-01 owner DB: %v", err)
	}
	return db
}

func openOINew01RuntimeDB(t *testing.T) *sql.DB {
	t.Helper()
	databaseURL := os.Getenv("OPENINVEST_DATABASE_RUNTIME_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_RUNTIME_TEST_URL is not set")
	}
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open OI-NEW-01 runtime DB: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close OI-NEW-01 runtime DB: %v", err)
		}
	})
	if err := db.Ping(); err != nil {
		t.Fatalf("ping OI-NEW-01 runtime DB: %v", err)
	}
	return db
}

func TestOINew01CapabilityInventoryDescriptionStaysAuditable(t *testing.T) {
	lines := []string{
		"OI_NEW_01_CAPABILITY identity.users SELECT=true INSERT=true UPDATE=false DELETE=false",
		"OI_NEW_01_CAPABILITY identity.user_investment_links SELECT=true INSERT=true UPDATE=false DELETE=false",
		"OI_NEW_01_CAPABILITY identity.credentials SELECT=true INSERT=true UPDATE=false DELETE=false",
		"OI_NEW_01_CAPABILITY identity.privacy_settings SELECT=true INSERT=true UPDATE=false DELETE=false",
		"OI_NEW_01_CAPABILITY identity.sessions SELECT=true INSERT=true UPDATE=true DELETE=true",
		"OI_NEW_01_CAPABILITY investment.subjects SELECT=false INSERT=true UPDATE=false DELETE=false COLUMN_SELECT=id",
		"OI_NEW_01_CAPABILITY investment.assets SELECT=true INSERT=true UPDATE=false DELETE=false",
		"OI_NEW_01_CAPABILITY investment.portfolios SELECT=true INSERT=true UPDATE=false DELETE=false COLUMN_UPDATE=portfolio_state",
		"OI_NEW_01_CAPABILITY investment.transaction_entries SELECT=true INSERT=true UPDATE=false DELETE=false",
		"OI_NEW_01_CAPABILITY investment.command_deduplication SELECT=true INSERT=true UPDATE=true DELETE=true",
		"OI_NEW_01_CAPABILITY investment.outbox_events SELECT=false INSERT=false UPDATE=false DELETE=false",
		"OI_NEW_01_CAPABILITY investment.portfolio_manual_valuations SELECT=true INSERT=true UPDATE=true DELETE=true",
		"OI_NEW_01_CAPABILITY analytics.portfolio_snapshots SELECT=true INSERT=true UPDATE=false DELETE=false",
		"OI_NEW_01_CAPABILITY analytics.snapshot_positions SELECT=false INSERT=false UPDATE=false DELETE=false",
		"OI_NEW_01_CAPABILITY analytics.calculation_runs SELECT=false INSERT=false UPDATE=false DELETE=false",
		"OI_NEW_01_CAPABILITY analytics.inbox_messages SELECT=false INSERT=false UPDATE=false DELETE=false",
		"OI_NEW_01_CAPABILITY audit.actors SELECT=false INSERT=true UPDATE=false DELETE=false COLUMN_SELECT=id",
		"OI_NEW_01_CAPABILITY audit.events SELECT=true INSERT=true UPDATE=false DELETE=false",
		"OI_NEW_01_CAPABILITY audit.auth_security_event_deduplications SELECT=false INSERT=true UPDATE=false DELETE=false COLUMN_SELECT=action_code,session_id",
		"OI_NEW_01_UNKNOWN_RELATION zero_effective_capability=ALLOW any_effective_capability=REJECT",
		"OI_NEW_01_UNKNOWN_SEQUENCE zero_effective_capability=ALLOW any_effective_capability=REJECT",
	}
	for _, line := range lines {
		t.Log(line)
	}
}

func TestOINew01AuditEventsRemainReadAppendOnly(t *testing.T) {
	ownerURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	runtimeURL := os.Getenv("OPENINVEST_DATABASE_RUNTIME_TEST_URL")
	if ownerURL == "" || runtimeURL == "" {
		t.Skip("runtime privilege integration URLs are not set")
	}
	ctx := context.Background()
	ownerDB, err := sql.Open("pgx", ownerURL)
	if err != nil {
		t.Fatalf("open owner db: %v", err)
	}
	t.Cleanup(func() {
		if err := ownerDB.Close(); err != nil {
			t.Errorf("close owner db: %v", err)
		}
	})
	runtimeDB, err := sql.Open("pgx", runtimeURL)
	if err != nil {
		t.Fatalf("open runtime db: %v", err)
	}
	t.Cleanup(func() {
		if err := runtimeDB.Close(); err != nil {
			t.Errorf("close runtime db: %v", err)
		}
	})

	const action = "OI_NEW_01_RUNTIME_PRIVILEGE_PROOF"
	if _, err := ownerDB.ExecContext(ctx, `DELETE FROM audit.events WHERE action_code = $1`, action); err != nil {
		t.Fatalf("pre-clean audit proof rows: %v", err)
	}
	t.Cleanup(func() {
		if _, err := ownerDB.ExecContext(context.Background(), `DELETE FROM audit.events WHERE action_code = $1`, action); err != nil {
			t.Errorf("cleanup audit proof rows: %v", err)
		}
	})

	var before int
	if err := runtimeDB.QueryRowContext(ctx, `SELECT count(*) FROM audit.events`).Scan(&before); err != nil {
		t.Fatalf("runtime SELECT audit.events: %v", err)
	}
	id := uuid.NewString()
	if _, err := runtimeDB.ExecContext(ctx, `
		INSERT INTO audit.events (id, actor_id, action_code, target_kind, target_id, outcome, request_id, trace_id, occurred_at, schema_version)
		VALUES ($1, NULL, $2, 'runtime_privilege', NULL, 'success', NULL, NULL, now(), 1)
	`, id, action); err != nil {
		t.Fatalf("runtime INSERT audit.events: %v", err)
	}
	var got string
	if err := runtimeDB.QueryRowContext(ctx, `SELECT action_code FROM audit.events WHERE id = $1`, id).Scan(&got); err != nil {
		t.Fatalf("runtime SELECT inserted audit event: %v", err)
	}
	if got != action {
		t.Fatalf("unexpected audit action %q", got)
	}

	requireDenied := func(label string, query string) {
		t.Helper()
		_, err := runtimeDB.ExecContext(ctx, query, id)
		if err == nil {
			t.Fatalf("%s unexpectedly succeeded", label)
		}
		var state interface{ SQLState() string }
		if !errors.As(err, &state) || state.SQLState() != "42501" {
			t.Fatalf("%s: expected SQLSTATE 42501, got %v", label, err)
		}
	}
	requireDenied("audit.events UPDATE", `UPDATE audit.events SET action_code = action_code WHERE id = $1`)
	requireDenied("audit.events DELETE", `DELETE FROM audit.events WHERE id = $1`)
	var canTruncate, canReferences, canTrigger, canMaintain bool
	if err := runtimeDB.QueryRowContext(ctx, `
		SELECT
			has_table_privilege(session_user, 'audit.events', 'TRUNCATE'),
			has_table_privilege(session_user, 'audit.events', 'REFERENCES'),
			has_table_privilege(session_user, 'audit.events', 'TRIGGER'),
			has_table_privilege(session_user, 'audit.events', 'MAINTAIN')
	`).Scan(&canTruncate, &canReferences, &canTrigger, &canMaintain); err != nil {
		t.Fatalf("inspect audit.events forbidden capabilities: %v", err)
	}
	if canTruncate || canReferences || canTrigger || canMaintain {
		t.Fatalf("audit.events has forbidden capability: truncate=%t references=%t trigger=%t maintain=%t",
			canTruncate, canReferences, canTrigger, canMaintain)
	}
}

func TestOINew01DangerousDefaultACLsRemainRejected(t *testing.T) {
	ownerURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	runtimeURL := os.Getenv("OPENINVEST_DATABASE_RUNTIME_TEST_URL")
	if ownerURL == "" || runtimeURL == "" {
		t.Skip("runtime privilege integration URLs are not set")
	}
	ctx := context.Background()
	ownerDB, err := sql.Open("pgx", ownerURL)
	if err != nil {
		t.Fatalf("open owner db: %v", err)
	}
	t.Cleanup(func() {
		if err := ownerDB.Close(); err != nil {
			t.Errorf("close owner db: %v", err)
		}
	})
	var ownerQuoted string
	if err := ownerDB.QueryRowContext(ctx, `SELECT quote_ident(session_user::text)`).Scan(&ownerQuoted); err != nil {
		t.Fatalf("read owner role: %v", err)
	}
	tableRevoke := `ALTER DEFAULT PRIVILEGES FOR ROLE ` + ownerQuoted + ` IN SCHEMA investment REVOKE SELECT ON TABLES FROM openinvest_runtime`
	sequenceRevoke := `ALTER DEFAULT PRIVILEGES FOR ROLE ` + ownerQuoted + ` IN SCHEMA investment REVOKE USAGE ON SEQUENCES FROM openinvest_runtime`
	// Pre-clean makes the test safe after any interrupted earlier process.
	if _, err := ownerDB.ExecContext(ctx, tableRevoke); err != nil {
		t.Fatalf("pre-clean relation default ACL: %v", err)
	}
	if _, err := ownerDB.ExecContext(ctx, sequenceRevoke); err != nil {
		t.Fatalf("pre-clean sequence default ACL: %v", err)
	}
	t.Cleanup(func() {
		if _, err := ownerDB.ExecContext(context.Background(), tableRevoke); err != nil {
			t.Errorf("cleanup relation default ACL: %v", err)
		}
		if _, err := ownerDB.ExecContext(context.Background(), sequenceRevoke); err != nil {
			t.Errorf("cleanup sequence default ACL: %v", err)
		}
	})

	requireUnsafe := func(label string) {
		t.Helper()
		store, err := postgres.OpenRuntime(runtimeURL)
		if err == nil {
			_ = store.Close()
			t.Fatalf("%s: OpenRuntime unexpectedly accepted dangerous default ACL", label)
		}
		if !errors.Is(err, postgres.ErrUnsafeRuntimeDatabaseRole) {
			t.Fatalf("%s: expected ErrUnsafeRuntimeDatabaseRole, got %v", label, err)
		}
	}
	requireSafe := func(label string) {
		t.Helper()
		store, err := postgres.OpenRuntime(runtimeURL)
		if err != nil {
			t.Fatalf("%s: OpenRuntime unexpectedly failed after cleanup: %v", label, err)
		}
		_ = store.Close()
	}

	if _, err := ownerDB.ExecContext(ctx, `ALTER DEFAULT PRIVILEGES FOR ROLE `+ownerQuoted+` IN SCHEMA investment GRANT SELECT ON TABLES TO openinvest_runtime`); err != nil {
		t.Fatalf("grant relation default ACL: %v", err)
	}
	requireUnsafe("future relation default SELECT")
	if _, err := ownerDB.ExecContext(ctx, tableRevoke); err != nil {
		t.Fatalf("revoke relation default ACL: %v", err)
	}
	requireSafe("relation default ACL cleanup")

	if _, err := ownerDB.ExecContext(ctx, `ALTER DEFAULT PRIVILEGES FOR ROLE `+ownerQuoted+` IN SCHEMA investment GRANT USAGE ON SEQUENCES TO openinvest_runtime`); err != nil {
		t.Fatalf("grant sequence default ACL: %v", err)
	}
	requireUnsafe("future sequence default USAGE")
	if _, err := ownerDB.ExecContext(ctx, sequenceRevoke); err != nil {
		t.Fatalf("revoke sequence default ACL: %v", err)
	}
	requireSafe("sequence default ACL cleanup")
}

func TestOINew01PortfolioRowLockUsesOnlyColumnUpdateCapability(t *testing.T) {
	ownerURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	runtimeURL := os.Getenv("OPENINVEST_DATABASE_RUNTIME_TEST_URL")
	if ownerURL == "" || runtimeURL == "" {
		t.Skip("runtime privilege integration URLs are not set")
	}
	ctx := context.Background()
	ownerDB, err := sql.Open("pgx", ownerURL)
	if err != nil {
		t.Fatalf("open owner db: %v", err)
	}
	t.Cleanup(func() {
		if err := ownerDB.Close(); err != nil {
			t.Errorf("close owner db: %v", err)
		}
	})
	runtimeDB, err := sql.Open("pgx", runtimeURL)
	if err != nil {
		t.Fatalf("open runtime db: %v", err)
	}
	t.Cleanup(func() {
		if err := runtimeDB.Close(); err != nil {
			t.Errorf("close runtime db: %v", err)
		}
	})

	subjectID := uuid.NewString()
	portfolioID := uuid.NewString()
	t.Cleanup(func() {
		if _, err := ownerDB.ExecContext(context.Background(), `DELETE FROM investment.portfolios WHERE id = $1`, portfolioID); err != nil {
			t.Errorf("cleanup portfolio: %v", err)
		}
		if _, err := ownerDB.ExecContext(context.Background(), `DELETE FROM investment.subjects WHERE id = $1`, subjectID); err != nil {
			t.Errorf("cleanup subject: %v", err)
		}
	})
	if _, err := ownerDB.ExecContext(ctx, `INSERT INTO investment.subjects (id) VALUES ($1)`, subjectID); err != nil {
		t.Fatalf("seed subject: %v", err)
	}
	if _, err := ownerDB.ExecContext(ctx, `
		INSERT INTO investment.portfolios (id, subject_id, name, base_currency, portfolio_state, version, created_at, updated_at)
		VALUES ($1, $2, 'OI-NEW-01 lock proof', 'RUB', 'active', 1, now(), now())
	`, portfolioID, subjectID); err != nil {
		t.Fatalf("seed portfolio: %v", err)
	}

	var tableUpdate, stateUpdate, stateGrant bool
	if err := runtimeDB.QueryRowContext(ctx, `SELECT
		has_table_privilege(session_user, 'investment.portfolios', 'UPDATE'),
		has_column_privilege(session_user, 'investment.portfolios', 'portfolio_state', 'UPDATE'),
		has_column_privilege(session_user, 'investment.portfolios', 'portfolio_state', 'UPDATE WITH GRANT OPTION')
	`).Scan(&tableUpdate, &stateUpdate, &stateGrant); err != nil {
		t.Fatalf("inspect portfolio update privileges: %v", err)
	}
	if tableUpdate || !stateUpdate || stateGrant {
		t.Fatalf("unexpected portfolio update capability: table=%t state=%t state_grant=%t", tableUpdate, stateUpdate, stateGrant)
	}

	requireRuntimeSafe := func(label string) {
		t.Helper()
		store, err := postgres.OpenRuntime(runtimeURL)
		if err != nil {
			t.Fatalf("%s: OpenRuntime unexpectedly rejected runtime: %v", label, err)
		}
		_ = store.Close()
	}
	requireRuntimeUnsafe := func(label string) {
		t.Helper()
		store, err := postgres.OpenRuntime(runtimeURL)
		if err == nil {
			_ = store.Close()
			t.Fatalf("%s: OpenRuntime unexpectedly accepted expanded portfolio capability", label)
		}
		if !errors.Is(err, postgres.ErrUnsafeRuntimeDatabaseRole) {
			t.Fatalf("%s: expected ErrUnsafeRuntimeDatabaseRole, got %v", label, err)
		}
	}
	requireRuntimeSafe("baseline column-update envelope")
	if _, err := ownerDB.ExecContext(ctx, `GRANT UPDATE (name) ON investment.portfolios TO openinvest_runtime`); err != nil {
		t.Fatalf("grant unauthorized portfolio name update: %v", err)
	}
	requireRuntimeUnsafe("unauthorized portfolio name column UPDATE")
	if _, err := ownerDB.ExecContext(ctx, `REVOKE UPDATE (name) ON investment.portfolios FROM openinvest_runtime`); err != nil {
		t.Fatalf("revoke unauthorized portfolio name update: %v", err)
	}
	requireRuntimeSafe("after portfolio name UPDATE cleanup")
	if _, err := ownerDB.ExecContext(ctx, `GRANT UPDATE (portfolio_state) ON investment.portfolios TO openinvest_runtime WITH GRANT OPTION`); err != nil {
		t.Fatalf("grant portfolio_state update grant option: %v", err)
	}
	requireRuntimeUnsafe("portfolio_state UPDATE WITH GRANT OPTION")
	if _, err := ownerDB.ExecContext(ctx, `REVOKE GRANT OPTION FOR UPDATE (portfolio_state) ON investment.portfolios FROM openinvest_runtime`); err != nil {
		t.Fatalf("revoke portfolio_state update grant option: %v", err)
	}
	requireRuntimeSafe("after portfolio_state grant-option cleanup")

	tx, err := runtimeDB.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin runtime lock tx: %v", err)
	}
	var lockedID string
	err = tx.QueryRowContext(ctx, `
		SELECT id FROM investment.portfolios
		WHERE id = $1 AND subject_id = $2 AND portfolio_state = 'active'
		FOR UPDATE
	`, portfolioID, subjectID).Scan(&lockedID)
	if err != nil {
		_ = tx.Rollback()
		t.Fatalf("production SELECT FOR UPDATE shape failed: %v", err)
	}
	if lockedID != portfolioID {
		_ = tx.Rollback()
		t.Fatalf("locked unexpected portfolio %s", lockedID)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback lock proof: %v", err)
	}

	requireDenied := func(label, query string) {
		t.Helper()
		_, err := runtimeDB.ExecContext(ctx, query, portfolioID)
		if err == nil {
			t.Fatalf("%s unexpectedly succeeded", label)
		}
		var state interface{ SQLState() string }
		if !errors.As(err, &state) || state.SQLState() != "42501" {
			t.Fatalf("%s: expected SQLSTATE 42501, got %v", label, err)
		}
	}
	for _, tc := range []struct{ label, query string }{
		{"name", `UPDATE investment.portfolios SET name = name WHERE id = $1`},
		{"base_currency", `UPDATE investment.portfolios SET base_currency = base_currency WHERE id = $1`},
		{"version", `UPDATE investment.portfolios SET version = version WHERE id = $1`},
		{"subject_id", `UPDATE investment.portfolios SET subject_id = subject_id WHERE id = $1`},
		{"id", `UPDATE investment.portfolios SET id = id WHERE id = $1`},
		{"created_at", `UPDATE investment.portfolios SET created_at = created_at WHERE id = $1`},
		{"updated_at", `UPDATE investment.portfolios SET updated_at = updated_at WHERE id = $1`},
		{"removed_at", `UPDATE investment.portfolios SET removed_at = removed_at WHERE id = $1`},
	} {
		requireDenied("portfolio "+tc.label+" UPDATE", tc.query)
	}

	_, err = runtimeDB.ExecContext(ctx, `UPDATE investment.portfolios SET portfolio_state = 'removed_from_active_use' WHERE id = $1`, portfolioID)
	if err == nil {
		t.Fatal("portfolio_state-only active->removed transition unexpectedly succeeded")
	}
	var state interface{ SQLState() string }
	if !errors.As(err, &state) || state.SQLState() != "23514" {
		t.Fatalf("active->removed: expected CHECK SQLSTATE 23514, got %v", err)
	}

	if _, err := ownerDB.ExecContext(ctx, `UPDATE investment.portfolios SET portfolio_state='removed_from_active_use', removed_at=now(), updated_at=now() WHERE id=$1`, portfolioID); err != nil {
		t.Fatalf("owner seed removed portfolio: %v", err)
	}
	_, err = runtimeDB.ExecContext(ctx, `UPDATE investment.portfolios SET portfolio_state = 'active' WHERE id = $1`, portfolioID)
	if err == nil {
		t.Fatal("portfolio_state-only removed->active transition unexpectedly succeeded")
	}
	if !errors.As(err, &state) || state.SQLState() != "23514" {
		t.Fatalf("removed->active: expected CHECK SQLSTATE 23514, got %v", err)
	}
}

func TestOINew01ACLResetIsTransactionallyInvisibleUntilCommit(t *testing.T) {
	ownerURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	runtimeURL := os.Getenv("OPENINVEST_DATABASE_RUNTIME_TEST_URL")
	if ownerURL == "" || runtimeURL == "" {
		t.Skip("runtime privilege integration URLs are not set")
	}
	ctx := context.Background()
	ownerDB, err := sql.Open("pgx", ownerURL)
	if err != nil {
		t.Fatalf("open owner db: %v", err)
	}
	t.Cleanup(func() {
		if err := ownerDB.Close(); err != nil {
			t.Errorf("close owner db: %v", err)
		}
	})
	runtimeDB, err := sql.Open("pgx", runtimeURL)
	if err != nil {
		t.Fatalf("open runtime db: %v", err)
	}
	t.Cleanup(func() {
		if err := runtimeDB.Close(); err != nil {
			t.Errorf("close runtime db: %v", err)
		}
	})

	checkRuntimeStillSeesCommittedACL := func(label string) {
		t.Helper()
		var usersSelect, auditSelect, portfolioTableUpdate, portfolioStateUpdate bool
		if err := runtimeDB.QueryRowContext(ctx, `SELECT
			has_table_privilege(session_user, 'identity.users', 'SELECT'),
			has_table_privilege(session_user, 'audit.events', 'SELECT'),
			has_table_privilege(session_user, 'investment.portfolios', 'UPDATE'),
			has_column_privilege(session_user, 'investment.portfolios', 'portfolio_state', 'UPDATE')
		`).Scan(&usersSelect, &auditSelect, &portfolioTableUpdate, &portfolioStateUpdate); err != nil {
			t.Fatalf("%s inspect runtime ACL: %v", label, err)
		}
		if !usersSelect || !auditSelect || portfolioTableUpdate || !portfolioStateUpdate {
			t.Fatalf("%s observed partial/unexpected ACL: users_select=%t audit_select=%t portfolio_table_update=%t portfolio_state_update=%t",
				label, usersSelect, auditSelect, portfolioTableUpdate, portfolioStateUpdate)
		}
	}

	checkRuntimeStillSeesCommittedACL("before owner reset")
	tx, err := ownerDB.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin owner ACL reset tx: %v", err)
	}
	if _, err := tx.ExecContext(ctx, `REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA identity, investment, analytics, audit FROM openinvest_runtime`); err != nil {
		_ = tx.Rollback()
		t.Fatalf("transactional broad revoke: %v", err)
	}
	// Another connection must continue observing the last committed capability matrix while the reset is uncommitted.
	checkRuntimeStillSeesCommittedACL("during uncommitted owner reset")
	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback owner ACL reset: %v", err)
	}
	checkRuntimeStillSeesCommittedACL("after rollback")
}

func TestOINew01Round4PrivilegeEnvelope(t *testing.T) {
	ownerURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	runtimeURL := os.Getenv("OPENINVEST_DATABASE_RUNTIME_TEST_URL")
	if ownerURL == "" || runtimeURL == "" {
		t.Skip("runtime privilege integration URLs are not set")
	}
	ctx := context.Background()
	ownerDB, err := sql.Open("pgx", ownerURL)
	if err != nil {
		t.Fatalf("open owner db: %v", err)
	}
	t.Cleanup(func() {
		if err := ownerDB.Close(); err != nil {
			t.Errorf("close owner db: %v", err)
		}
	})
	runtimeDB, err := sql.Open("pgx", runtimeURL)
	if err != nil {
		t.Fatalf("open runtime db: %v", err)
	}
	t.Cleanup(func() {
		if err := runtimeDB.Close(); err != nil {
			t.Errorf("close runtime db: %v", err)
		}
	})

	var runtimeLogin string
	if err := runtimeDB.QueryRowContext(ctx, `SELECT session_user::text`).Scan(&runtimeLogin); err != nil {
		t.Fatalf("inspect runtime login: %v", err)
	}
	quoteIdent := func(value string) string {
		return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
	}
	runtimeIdent := quoteIdent(runtimeLogin)

	requireSafe := func(st *testing.T, label string) {
		st.Helper()
		store, err := postgres.OpenRuntime(runtimeURL)
		if err != nil {
			st.Fatalf("%s: OpenRuntime unexpectedly rejected runtime: %v", label, err)
		}
		if err := store.Close(); err != nil {
			st.Fatalf("%s: close runtime store: %v", label, err)
		}
	}
	requireUnsafe := func(st *testing.T, label string) {
		st.Helper()
		store, err := postgres.OpenRuntime(runtimeURL)
		if err == nil {
			_ = store.Close()
			st.Fatalf("%s: OpenRuntime unexpectedly accepted unsafe runtime", label)
		}
		if !errors.Is(err, postgres.ErrUnsafeRuntimeDatabaseRole) {
			st.Fatalf("%s: expected ErrUnsafeRuntimeDatabaseRole, got %v", label, err)
		}
	}

	requireSafe(t, "round4 clean baseline")

	t.Run("predefined_role_direct", func(st *testing.T) {
		if _, err := ownerDB.ExecContext(ctx, `GRANT pg_execute_server_program TO `+runtimeIdent); err != nil {
			st.Fatalf("grant predefined role directly: %v", err)
		}
		st.Cleanup(func() {
			if _, err := ownerDB.ExecContext(context.Background(), `REVOKE pg_execute_server_program FROM `+runtimeIdent); err != nil {
				st.Errorf("cleanup predefined direct grant: %v", err)
			}
		})
		requireUnsafe(st, "direct pg_execute_server_program")
		if _, err := ownerDB.ExecContext(ctx, `REVOKE pg_execute_server_program FROM `+runtimeIdent); err != nil {
			st.Fatalf("revoke predefined direct grant: %v", err)
		}
		requireSafe(st, "after predefined direct cleanup")
	})

	t.Run("predefined_role_transitive", func(st *testing.T) {
		mid := "oi_new_01_predef_mid_" + strings.ReplaceAll(uuid.NewString(), "-", "")
		midIdent := quoteIdent(mid)
		if _, err := ownerDB.ExecContext(ctx, `CREATE ROLE `+midIdent+` NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS`); err != nil {
			st.Fatalf("create predefined intermediate role: %v", err)
		}
		st.Cleanup(func() {
			_, _ = ownerDB.ExecContext(context.Background(), `REVOKE `+midIdent+` FROM `+runtimeIdent)
			_, _ = ownerDB.ExecContext(context.Background(), `REVOKE pg_execute_server_program FROM `+midIdent)
			if _, err := ownerDB.ExecContext(context.Background(), `DROP ROLE IF EXISTS `+midIdent); err != nil {
				st.Errorf("drop predefined intermediate role: %v", err)
			}
		})
		if _, err := ownerDB.ExecContext(ctx, `GRANT pg_execute_server_program TO `+midIdent); err != nil {
			st.Fatalf("grant predefined role to intermediate: %v", err)
		}
		if _, err := ownerDB.ExecContext(ctx, `GRANT `+midIdent+` TO `+runtimeIdent); err != nil {
			st.Fatalf("grant intermediate role to runtime: %v", err)
		}
		requireUnsafe(st, "transitive pg_execute_server_program")
		if _, err := ownerDB.ExecContext(ctx, `REVOKE `+midIdent+` FROM `+runtimeIdent); err != nil {
			st.Fatalf("revoke intermediate role from runtime: %v", err)
		}
		if _, err := ownerDB.ExecContext(ctx, `REVOKE pg_execute_server_program FROM `+midIdent); err != nil {
			st.Fatalf("revoke predefined role from intermediate: %v", err)
		}
		if _, err := ownerDB.ExecContext(ctx, `DROP ROLE `+midIdent); err != nil {
			st.Fatalf("drop predefined intermediate role: %v", err)
		}
		requireSafe(st, "after predefined transitive cleanup")
	})

	t.Run("parameter_acl_direct", func(st *testing.T) {
		if _, err := ownerDB.ExecContext(ctx, `GRANT SET ON PARAMETER session_replication_role TO `+runtimeIdent); err != nil {
			st.Fatalf("grant SET parameter privilege: %v", err)
		}
		st.Cleanup(func() {
			_, _ = ownerDB.ExecContext(context.Background(), `REVOKE SET ON PARAMETER session_replication_role FROM `+runtimeIdent)
		})
		requireUnsafe(st, "direct parameter SET privilege")
		if _, err := ownerDB.ExecContext(ctx, `REVOKE SET ON PARAMETER session_replication_role FROM `+runtimeIdent); err != nil {
			st.Fatalf("revoke SET parameter privilege: %v", err)
		}
		requireSafe(st, "after parameter SET cleanup")

		if _, err := ownerDB.ExecContext(ctx, `GRANT ALTER SYSTEM ON PARAMETER session_replication_role TO `+runtimeIdent); err != nil {
			st.Fatalf("grant ALTER SYSTEM parameter privilege: %v", err)
		}
		st.Cleanup(func() {
			_, _ = ownerDB.ExecContext(context.Background(), `REVOKE ALTER SYSTEM ON PARAMETER session_replication_role FROM `+runtimeIdent)
		})
		requireUnsafe(st, "direct parameter ALTER SYSTEM privilege")
		if _, err := ownerDB.ExecContext(ctx, `REVOKE ALTER SYSTEM ON PARAMETER session_replication_role FROM `+runtimeIdent); err != nil {
			st.Fatalf("revoke ALTER SYSTEM parameter privilege: %v", err)
		}
		requireSafe(st, "after parameter ALTER SYSTEM cleanup")
	})

	t.Run("parameter_acl_transitive", func(st *testing.T) {
		mid := "oi_new_01_param_mid_" + strings.ReplaceAll(uuid.NewString(), "-", "")
		midIdent := quoteIdent(mid)
		if _, err := ownerDB.ExecContext(ctx, `CREATE ROLE `+midIdent+` NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS`); err != nil {
			st.Fatalf("create parameter intermediate role: %v", err)
		}
		st.Cleanup(func() {
			_, _ = ownerDB.ExecContext(context.Background(), `REVOKE `+midIdent+` FROM `+runtimeIdent)
			_, _ = ownerDB.ExecContext(context.Background(), `REVOKE SET ON PARAMETER session_replication_role FROM `+midIdent)
			if _, err := ownerDB.ExecContext(context.Background(), `DROP ROLE IF EXISTS `+midIdent); err != nil {
				st.Errorf("drop parameter intermediate role: %v", err)
			}
		})
		if _, err := ownerDB.ExecContext(ctx, `GRANT SET ON PARAMETER session_replication_role TO `+midIdent); err != nil {
			st.Fatalf("grant parameter privilege to intermediate: %v", err)
		}
		if _, err := ownerDB.ExecContext(ctx, `GRANT `+midIdent+` TO `+runtimeIdent); err != nil {
			st.Fatalf("grant parameter intermediate role to runtime: %v", err)
		}
		requireUnsafe(st, "transitive parameter SET privilege")
		if _, err := ownerDB.ExecContext(ctx, `REVOKE `+midIdent+` FROM `+runtimeIdent); err != nil {
			st.Fatalf("revoke parameter intermediate role from runtime: %v", err)
		}
		if _, err := ownerDB.ExecContext(ctx, `REVOKE SET ON PARAMETER session_replication_role FROM `+midIdent); err != nil {
			st.Fatalf("revoke parameter privilege from intermediate: %v", err)
		}
		if _, err := ownerDB.ExecContext(ctx, `DROP ROLE `+midIdent); err != nil {
			st.Fatalf("drop parameter intermediate role: %v", err)
		}
		requireSafe(st, "after parameter transitive cleanup")
	})

	t.Run("session_replication_role_startup_state", func(st *testing.T) {
		st.Cleanup(func() {
			_, _ = ownerDB.ExecContext(context.Background(), `ALTER ROLE `+runtimeIdent+` RESET session_replication_role`)
		})
		if _, err := ownerDB.ExecContext(ctx, `ALTER ROLE `+runtimeIdent+` SET session_replication_role = replica`); err != nil {
			st.Fatalf("set dangerous runtime startup state: %v", err)
		}
		requireUnsafe(st, "session_replication_role=replica startup state")
		if _, err := ownerDB.ExecContext(ctx, `ALTER ROLE `+runtimeIdent+` RESET session_replication_role`); err != nil {
			st.Fatalf("reset dangerous runtime startup state: %v", err)
		}
		requireSafe(st, "after session_replication_role reset")
	})

	var databaseName, databaseOwner string
	if err := ownerDB.QueryRowContext(ctx, `SELECT current_database(), pg_get_userbyid(datdba) FROM pg_database WHERE datname=current_database()`).Scan(&databaseName, &databaseOwner); err != nil {
		t.Fatalf("inspect current database owner: %v", err)
	}
	databaseIdent := quoteIdent(databaseName)
	databaseOwnerIdent := quoteIdent(databaseOwner)

	t.Run("database_create", func(st *testing.T) {
		st.Cleanup(func() {
			_, _ = ownerDB.ExecContext(context.Background(), `REVOKE CREATE ON DATABASE `+databaseIdent+` FROM `+runtimeIdent)
		})
		if _, err := ownerDB.ExecContext(ctx, `GRANT CREATE ON DATABASE `+databaseIdent+` TO `+runtimeIdent); err != nil {
			st.Fatalf("grant current-database CREATE: %v", err)
		}
		requireUnsafe(st, "current-database CREATE")
		if _, err := ownerDB.ExecContext(ctx, `REVOKE CREATE ON DATABASE `+databaseIdent+` FROM `+runtimeIdent); err != nil {
			st.Fatalf("revoke current-database CREATE: %v", err)
		}
		requireSafe(st, "after database CREATE cleanup")
	})

	t.Run("database_owner", func(st *testing.T) {
		st.Cleanup(func() {
			_, _ = ownerDB.ExecContext(context.Background(), `ALTER DATABASE `+databaseIdent+` OWNER TO `+databaseOwnerIdent)
		})
		if _, err := ownerDB.ExecContext(ctx, `ALTER DATABASE `+databaseIdent+` OWNER TO `+runtimeIdent); err != nil {
			st.Fatalf("make runtime current-database owner: %v", err)
		}
		requireUnsafe(st, "current-database ownership")
		if _, err := ownerDB.ExecContext(ctx, `ALTER DATABASE `+databaseIdent+` OWNER TO `+databaseOwnerIdent); err != nil {
			st.Fatalf("restore current-database owner: %v", err)
		}
		requireSafe(st, "after database owner cleanup")
	})

	t.Run("public_schema_create", func(st *testing.T) {
		st.Cleanup(func() {
			_, _ = ownerDB.ExecContext(context.Background(), `REVOKE CREATE ON SCHEMA public FROM `+runtimeIdent)
		})
		if _, err := ownerDB.ExecContext(ctx, `GRANT CREATE ON SCHEMA public TO `+runtimeIdent); err != nil {
			st.Fatalf("grant public schema CREATE: %v", err)
		}
		requireUnsafe(st, "public schema CREATE")
		if _, err := ownerDB.ExecContext(ctx, `REVOKE CREATE ON SCHEMA public FROM `+runtimeIdent); err != nil {
			st.Fatalf("revoke public schema CREATE: %v", err)
		}
		requireSafe(st, "after public schema CREATE cleanup")
	})

	t.Run("future_user_schema", func(st *testing.T) {
		schema := "oi_new_01_future_schema_" + strings.ReplaceAll(uuid.NewString(), "-", "")
		schemaIdent := quoteIdent(schema)
		if _, err := ownerDB.ExecContext(ctx, `CREATE SCHEMA `+schemaIdent); err != nil {
			st.Fatalf("create future user schema: %v", err)
		}
		st.Cleanup(func() {
			_, _ = ownerDB.ExecContext(context.Background(), `REVOKE CREATE ON SCHEMA `+schemaIdent+` FROM `+runtimeIdent)
			if _, err := ownerDB.ExecContext(context.Background(), `DROP SCHEMA IF EXISTS `+schemaIdent+` CASCADE`); err != nil {
				st.Errorf("drop future user schema: %v", err)
			}
		})
		requireSafe(st, "future schema with zero CREATE")
		if _, err := ownerDB.ExecContext(ctx, `GRANT CREATE ON SCHEMA `+schemaIdent+` TO `+runtimeIdent); err != nil {
			st.Fatalf("grant future user schema CREATE: %v", err)
		}
		requireUnsafe(st, "future schema CREATE")
		if _, err := ownerDB.ExecContext(ctx, `REVOKE CREATE ON SCHEMA `+schemaIdent+` FROM `+runtimeIdent); err != nil {
			st.Fatalf("revoke future user schema CREATE: %v", err)
		}
		requireSafe(st, "after future schema CREATE cleanup")
	})

	t.Run("security_definer_execute", func(st *testing.T) {
		fn := "oi_new_01_sd_" + strings.ReplaceAll(uuid.NewString(), "-", "")
		fnIdent := quoteIdent(fn)
		qualified := `investment.` + fnIdent + `()`
		if _, err := ownerDB.ExecContext(ctx, `CREATE FUNCTION `+qualified+` RETURNS integer LANGUAGE sql SECURITY DEFINER AS 'SELECT 1'`); err != nil {
			st.Fatalf("create harmless SECURITY DEFINER fixture: %v", err)
		}
		st.Cleanup(func() {
			_, _ = ownerDB.ExecContext(context.Background(), `REVOKE EXECUTE ON FUNCTION `+qualified+` FROM `+runtimeIdent)
			_, _ = ownerDB.ExecContext(context.Background(), `REVOKE EXECUTE ON FUNCTION `+qualified+` FROM PUBLIC`)
			if _, err := ownerDB.ExecContext(context.Background(), `DROP FUNCTION IF EXISTS `+qualified); err != nil {
				st.Errorf("drop SECURITY DEFINER fixture: %v", err)
			}
		})
		requireUnsafe(st, "SECURITY DEFINER with default PUBLIC EXECUTE")
		if _, err := ownerDB.ExecContext(ctx, `REVOKE EXECUTE ON FUNCTION `+qualified+` FROM PUBLIC`); err != nil {
			st.Fatalf("revoke default PUBLIC EXECUTE: %v", err)
		}
		requireSafe(st, "SECURITY DEFINER without runtime EXECUTE")
		if _, err := ownerDB.ExecContext(ctx, `GRANT EXECUTE ON FUNCTION `+qualified+` TO `+runtimeIdent); err != nil {
			st.Fatalf("grant SECURITY DEFINER EXECUTE to runtime: %v", err)
		}
		requireUnsafe(st, "SECURITY DEFINER explicit runtime EXECUTE")
		if _, err := ownerDB.ExecContext(ctx, `REVOKE EXECUTE ON FUNCTION `+qualified+` FROM `+runtimeIdent); err != nil {
			st.Fatalf("revoke SECURITY DEFINER EXECUTE from runtime: %v", err)
		}
		requireSafe(st, "after SECURITY DEFINER EXECUTE cleanup")
	})
}
