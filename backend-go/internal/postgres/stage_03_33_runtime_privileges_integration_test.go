package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestStage0333RuntimeRoleCanAppendButCannotMutateLedger(t *testing.T) {
	ownerURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	runtimeURL := os.Getenv("OPENINVEST_DATABASE_RUNTIME_TEST_URL")
	if ownerURL == "" || runtimeURL == "" {
		t.Skip("runtime privilege integration URLs are not set")
	}

	if ownerStore, err := postgres.OpenRuntime(ownerURL); err == nil {
		_ = ownerStore.Close()
		t.Fatal("schema owner unexpectedly passed runtime privilege validation")
	} else if !errors.Is(err, postgres.ErrUnsafeRuntimeDatabaseRole) {
		t.Fatalf("expected unsafe runtime role error for owner connection, got %v", err)
	}

	runtimeStore, err := postgres.OpenRuntime(runtimeURL)
	if err != nil {
		t.Fatalf("open least-privilege runtime store: %v", err)
	}
	defer runtimeStore.Close()

	ownerDB, err := sql.Open("pgx", ownerURL)
	if err != nil {
		t.Fatalf("open owner cleanup db: %v", err)
	}
	closeDBOnCleanup(t, ownerDB, "stage 3.33 runtime owner")

	runtimeDB, err := sql.Open("pgx", runtimeURL)
	if err != nil {
		t.Fatalf("open runtime verification db: %v", err)
	}
	closeDBOnCleanup(t, runtimeDB, "stage 3.33 runtime role")

	ctx := context.Background()
	service := verticalslice.NewService(runtimeStore, verticalslice.SystemClock{})
	subjectID := uuid.NewString()
	portfolio, err := service.CreatePortfolio(
		ctx,
		verticalslice.RequestContext{},
		subjectID,
		"stage-03-33-runtime-portfolio-key-01",
		"/api/v1/portfolios",
		verticalslice.CreatePortfolioRequest{Name: "Runtime privilege proof", BaseCurrency: verticalslice.RUB},
	)
	if err != nil {
		t.Fatalf("runtime role create portfolio: %v", err)
	}
	t.Cleanup(func() { cleanupPortfolioRows(t, ctx, ownerDB, portfolio.ID) })

	_, err = service.AppendTransaction(
		ctx,
		verticalslice.RequestContext{},
		subjectID,
		"stage-03-33-runtime-transaction-key-01",
		"/api/v1/portfolios/"+portfolio.ID+"/transactions",
		verticalslice.AppendTransactionRequest{
			PortfolioID:     portfolio.ID,
			TransactionType: "DEPOSIT",
			GrossAmount: &verticalslice.Money{
				Amount:   decimal.Must("100.00000000"),
				Currency: verticalslice.RUB,
			},
			Commission: verticalslice.Money{Amount: decimal.Zero(), Currency: verticalslice.RUB},
			Tax:        verticalslice.Money{Amount: decimal.Zero(), Currency: verticalslice.RUB},
			TradeDate:  "2026-08-23",
		},
	)
	if err != nil {
		t.Fatalf("runtime role append transaction: %v", err)
	}

	if _, err := runtimeDB.ExecContext(ctx, `
		UPDATE investment.transaction_entries
		SET note = note
		WHERE portfolio_id = $1
	`, portfolio.ID); err == nil {
		t.Fatal("runtime role unexpectedly updated append-only transaction_entries")
	}
	if _, err := runtimeDB.ExecContext(ctx, `
		DELETE FROM investment.transaction_entries
		WHERE portfolio_id = $1
	`, portfolio.ID); err == nil {
		t.Fatal("runtime role unexpectedly deleted append-only transaction_entries")
	}

	var canTruncate bool
	if err := runtimeDB.QueryRowContext(ctx, `
		SELECT has_table_privilege(session_user, 'investment.transaction_entries', 'TRUNCATE')
	`).Scan(&canTruncate); err != nil {
		t.Fatalf("inspect runtime truncate privilege: %v", err)
	}
	if canTruncate {
		t.Fatal("runtime role unexpectedly has TRUNCATE on append-only transaction_entries")
	}
}

func TestStage0333OpenRuntimeRejectsPrivilegedSessionMaskedBySetRole(t *testing.T) {
	ownerURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if ownerURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}

	maskedOwnerURL, err := databaseURLWithRole(ownerURL, "openinvest_runtime")
	if err != nil {
		t.Fatalf("build masked owner URL: %v", err)
	}

	maskedDB, err := sql.Open("pgx", maskedOwnerURL)
	if err != nil {
		t.Fatalf("open masked owner verification db: %v", err)
	}
	closeDBOnCleanup(t, maskedDB, "stage 3.33 masked owner")

	ctx := context.Background()
	var sessionUser string
	var currentUser string
	if err := maskedDB.QueryRowContext(ctx, `SELECT session_user::text, current_user::text`).Scan(&sessionUser, &currentUser); err != nil {
		t.Fatalf("verify masked owner identity: %v", err)
	}
	if sessionUser == currentUser || currentUser != "openinvest_runtime" {
		t.Fatalf("masked-owner fixture did not establish session/current split: session=%q current=%q", sessionUser, currentUser)
	}

	if store, err := postgres.OpenRuntime(maskedOwnerURL); err == nil {
		_ = store.Close()
		t.Fatal("privileged authenticated session masked by SET ROLE unexpectedly passed runtime validation")
	} else if !errors.Is(err, postgres.ErrUnsafeRuntimeDatabaseRole) {
		t.Fatalf("expected unsafe runtime role error for masked owner session, got %v", err)
	}
}

func TestStage0333OpenRuntimeRejectsLatentSetRoleMutationPath(t *testing.T) {
	ownerURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	runtimeURL := os.Getenv("OPENINVEST_DATABASE_RUNTIME_TEST_URL")
	if ownerURL == "" || runtimeURL == "" {
		t.Skip("runtime privilege integration URLs are not set")
	}

	ownerDB, err := sql.Open("pgx", ownerURL)
	if err != nil {
		t.Fatalf("open owner role-management db: %v", err)
	}
	closeDBOnCleanup(t, ownerDB, "stage 3.33 escalation owner")

	runtimeDB, err := sql.Open("pgx", runtimeURL)
	if err != nil {
		t.Fatalf("open runtime escalation verification db: %v", err)
	}
	closeDBOnCleanup(t, runtimeDB, "stage 3.33 escalation runtime")

	ctx := context.Background()
	var runtimeLoginQuoted string
	if err := runtimeDB.QueryRowContext(ctx, `SELECT quote_ident(session_user::text)`).Scan(&runtimeLoginQuoted); err != nil {
		t.Fatalf("read runtime login identifier: %v", err)
	}

	escalationRole := "openinvest_runtime_escalation_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := ownerDB.ExecContext(ctx, "CREATE ROLE "+escalationRole+" NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION"); err != nil {
		t.Fatalf("create escalation role: %v", err)
	}
	t.Cleanup(func() {
		_, _ = ownerDB.ExecContext(context.Background(), "REVOKE "+escalationRole+" FROM "+runtimeLoginQuoted)
		_, _ = ownerDB.ExecContext(context.Background(), "DROP OWNED BY "+escalationRole)
		if _, err := ownerDB.ExecContext(context.Background(), "DROP ROLE "+escalationRole); err != nil {
			t.Errorf("drop escalation role %s: %v", escalationRole, err)
		}
	})

	if _, err := ownerDB.ExecContext(ctx, "GRANT USAGE ON SCHEMA investment TO "+escalationRole); err != nil {
		t.Fatalf("grant escalation schema usage: %v", err)
	}
	if _, err := ownerDB.ExecContext(ctx, "GRANT UPDATE ON investment.transaction_entries TO "+escalationRole); err != nil {
		t.Fatalf("grant escalation ledger update: %v", err)
	}
	if _, err := ownerDB.ExecContext(ctx, "GRANT "+escalationRole+" TO "+runtimeLoginQuoted+" WITH INHERIT FALSE, SET TRUE"); err != nil {
		t.Fatalf("grant latent SET ROLE escalation path: %v", err)
	}

	var directlyCanUpdate bool
	var canSetEscalationRole bool
	if err := runtimeDB.QueryRowContext(ctx, `
		SELECT
			has_table_privilege(session_user, 'investment.transaction_entries', 'UPDATE'),
			pg_has_role(session_user, $1::name, 'SET')
	`, escalationRole).Scan(&directlyCanUpdate, &canSetEscalationRole); err != nil {
		t.Fatalf("verify latent escalation fixture: %v", err)
	}
	if directlyCanUpdate {
		t.Fatal("latent escalation fixture accidentally granted UPDATE through inheritance")
	}
	if !canSetEscalationRole {
		t.Fatal("latent escalation fixture did not grant SET ROLE capability")
	}

	if store, err := postgres.OpenRuntime(runtimeURL); err == nil {
		_ = store.Close()
		t.Fatal("runtime login with SET-reachable ledger mutation role unexpectedly passed validation")
	} else if !errors.Is(err, postgres.ErrUnsafeRuntimeDatabaseRole) {
		t.Fatalf("expected unsafe runtime role error for latent SET ROLE escalation, got %v", err)
	}
}

func TestStage0333OpenRuntimeRejectsLatentAdminOptionMutationPath(t *testing.T) {
	ownerURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	runtimeURL := os.Getenv("OPENINVEST_DATABASE_RUNTIME_TEST_URL")
	if ownerURL == "" || runtimeURL == "" {
		t.Skip("runtime privilege integration URLs are not set")
	}

	ownerDB, err := sql.Open("pgx", ownerURL)
	if err != nil {
		t.Fatalf("open owner role-admin db: %v", err)
	}
	closeDBOnCleanup(t, ownerDB, "stage 3.33 admin escalation owner")

	runtimeDB, err := sql.Open("pgx", runtimeURL)
	if err != nil {
		t.Fatalf("open runtime admin escalation db: %v", err)
	}
	closeDBOnCleanup(t, runtimeDB, "stage 3.33 admin escalation runtime")

	ctx := context.Background()
	var runtimeLoginQuoted string
	if err := runtimeDB.QueryRowContext(ctx, `SELECT quote_ident(session_user::text)`).Scan(&runtimeLoginQuoted); err != nil {
		t.Fatalf("read runtime login identifier: %v", err)
	}

	adminRole := "openinvest_runtime_admin_escalation_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := ownerDB.ExecContext(ctx, "CREATE ROLE "+adminRole+" NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION"); err != nil {
		t.Fatalf("create admin escalation role: %v", err)
	}
	t.Cleanup(func() {
		_, _ = ownerDB.ExecContext(context.Background(), "REVOKE "+adminRole+" FROM "+runtimeLoginQuoted)
		_, _ = ownerDB.ExecContext(context.Background(), "DROP OWNED BY "+adminRole)
		if _, err := ownerDB.ExecContext(context.Background(), "DROP ROLE "+adminRole); err != nil {
			t.Errorf("drop admin escalation role %s: %v", adminRole, err)
		}
	})

	if _, err := ownerDB.ExecContext(ctx, "GRANT USAGE ON SCHEMA investment TO "+adminRole); err != nil {
		t.Fatalf("grant admin escalation schema usage: %v", err)
	}
	if _, err := ownerDB.ExecContext(ctx, "GRANT UPDATE ON investment.transaction_entries TO "+adminRole); err != nil {
		t.Fatalf("grant admin escalation ledger update: %v", err)
	}
	if _, err := ownerDB.ExecContext(ctx, "GRANT "+adminRole+" TO "+runtimeLoginQuoted+" WITH ADMIN TRUE, INHERIT FALSE, SET FALSE"); err != nil {
		t.Fatalf("grant latent ADMIN OPTION escalation path: %v", err)
	}

	var directlyCanUpdate bool
	var canSetAdminRole bool
	var hasAdminOption bool
	if err := runtimeDB.QueryRowContext(ctx, `
		SELECT
			has_table_privilege(session_user, 'investment.transaction_entries', 'UPDATE'),
			pg_has_role(session_user, $1::name, 'SET'),
			pg_has_role(session_user, $1::name, 'MEMBER WITH ADMIN OPTION')
	`, adminRole).Scan(&directlyCanUpdate, &canSetAdminRole, &hasAdminOption); err != nil {
		t.Fatalf("verify latent ADMIN OPTION fixture: %v", err)
	}
	if directlyCanUpdate {
		t.Fatal("admin escalation fixture accidentally granted UPDATE through inheritance")
	}
	if canSetAdminRole {
		t.Fatal("admin escalation fixture accidentally granted SET ROLE capability")
	}
	if !hasAdminOption {
		t.Fatal("admin escalation fixture did not establish ADMIN OPTION capability")
	}

	if store, err := postgres.OpenRuntime(runtimeURL); err == nil {
		_ = store.Close()
		t.Fatal("runtime login with latent ADMIN OPTION ledger escalation unexpectedly passed validation")
	} else if !errors.Is(err, postgres.ErrUnsafeRuntimeDatabaseRole) {
		t.Fatalf("expected unsafe runtime role error for latent ADMIN OPTION escalation, got %v", err)
	}
}

func databaseURLWithRole(databaseURL string, role string) (string, error) {
	parsed, err := url.Parse(databaseURL)
	if err != nil {
		return "", err
	}
	query := parsed.Query()
	query.Set("options", "-c role="+role)
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}
