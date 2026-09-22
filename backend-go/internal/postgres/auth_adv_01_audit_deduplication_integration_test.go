package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"sync"
	"testing"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/openinvest/openinvest/backend-go/internal/auth"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

const authAdv01Requests = 12

func TestAuthAdv01AnonymousRefreshAndLogoutRejectionsWriteNoDurableAudit(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}

	store, service := authAdv01StoreAndService(t, databaseURL)
	defer store.Close()
	db := authAdv01OwnerDB(t, databaseURL)
	defer db.Close()
	ctx := context.Background()

	before := authAdv01AllAuditCount(t, ctx, db)
	for range authAdv01Requests {
		if _, err := service.Refresh(ctx, "", "csrf-token"); !errors.Is(err, auth.ErrInvalidSession) {
			t.Fatalf("missing refresh token: %v", err)
		}
		if _, err := service.Refresh(ctx, "refresh-token", ""); !errors.Is(err, auth.ErrInvalidCSRF) {
			t.Fatalf("missing csrf token: %v", err)
		}
		if _, err := service.Refresh(ctx, "unknown-refresh-token", "unknown-csrf-token"); !errors.Is(err, auth.ErrInvalidSession) {
			t.Fatalf("unknown refresh pair: %v", err)
		}
		if _, err := service.Logout(ctx, "", "csrf-token", false); !errors.Is(err, auth.ErrInvalidSession) {
			t.Fatalf("missing logout refresh token: %v", err)
		}
		if _, err := service.Logout(ctx, "refresh-token", "", false); !errors.Is(err, auth.ErrInvalidCSRF) {
			t.Fatalf("missing logout csrf token: %v", err)
		}
		if _, err := service.Logout(ctx, "unknown-refresh-token", "unknown-csrf-token", false); !errors.Is(err, auth.ErrInvalidSession) {
			t.Fatalf("unknown logout pair: %v", err)
		}
	}
	after := authAdv01AllAuditCount(t, ctx, db)
	if after != before {
		t.Fatalf("anonymous auth rejections wrote durable audit rows: before=%d after=%d delta=%d", before, after, after-before)
	}
	t.Log("AUTH_ADV_01_MISSING_REFRESH_ZERO_DURABLE_AUDIT=PASS")
	t.Log("AUTH_ADV_01_MISSING_CSRF_ZERO_DURABLE_AUDIT=PASS")
	t.Log("AUTH_ADV_01_UNKNOWN_REFRESH_PAIR_ZERO_DURABLE_AUDIT=PASS")
	t.Log("AUTH_ADV_01_UNKNOWN_LOGOUT_PAIR_ZERO_DURABLE_AUDIT=PASS")
	t.Log("AUTH_ADV_01_UNKNOWN_PAIR_STILL_ZERO_DURABLE_AUDIT=PASS")
	t.Log("AUTH_ADV_01_MISSING_CSRF_STILL_ZERO_DURABLE_AUDIT=PASS")
}

func TestAuthAdv01KnownRefreshWrongCSRFCreatesBoundedEvidenceWithoutMutation(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}

	storeOne, serviceOne := authAdv01StoreAndService(t, databaseURL)
	defer storeOne.Close()
	storeTwo, serviceTwo := authAdv01StoreAndService(t, databaseURL)
	defer storeTwo.Close()
	db := authAdv01OwnerDB(t, databaseURL)
	defer db.Close()
	ctx := context.Background()

	root := authAdv01Register(t, ctx, serviceOne, "wrong-csrf")
	rootSessionID := authAdv01SessionID(t, ctx, db, root.User.ID)
	rotated, err := serviceOne.Refresh(ctx, root.RefreshToken, root.Session.CSRFToken)
	if err != nil {
		t.Fatalf("prepare active descendant: %v", err)
	}
	activeSessionID := authAdv01LatestSessionID(t, ctx, db, root.User.ID)
	beforeSessionCount := authAdv01SessionCount(t, ctx, db, root.User.ID)
	beforeFamilyActive := authAdv01ActiveFamilyCount(t, ctx, db, rootSessionID)

	refreshBefore := authAdv01AuditCount(t, ctx, db, "AUTH_REFRESH_CSRF_REJECTED", activeSessionID)
	for range authAdv01Requests {
		if _, err := serviceOne.Refresh(ctx, rotated.RefreshToken, "wrong-non-empty-csrf-token"); !errors.Is(err, auth.ErrInvalidSession) {
			t.Fatalf("known refresh with wrong csrf: %v", err)
		}
	}
	if delta := authAdv01AuditCount(t, ctx, db, "AUTH_REFRESH_CSRF_REJECTED", activeSessionID) - refreshBefore; delta != 1 {
		t.Fatalf("wrong csrf refresh evidence grew beyond its durable bound: delta=%d", delta)
	}
	if count := authAdv01DeduplicationCount(t, ctx, db, "AUTH_REFRESH_CSRF_REJECTED", activeSessionID); count != 1 {
		t.Fatalf("expected one wrong csrf refresh deduplication gate, got %d", count)
	}

	logoutBefore := authAdv01AuditCount(t, ctx, db, "AUTH_LOGOUT_CSRF_REJECTED", activeSessionID)
	for range authAdv01Requests {
		if _, err := serviceOne.Logout(ctx, rotated.RefreshToken, "wrong-non-empty-csrf-token", false); !errors.Is(err, auth.ErrInvalidSession) {
			t.Fatalf("known logout with wrong csrf: %v", err)
		}
	}
	if delta := authAdv01AuditCount(t, ctx, db, "AUTH_LOGOUT_CSRF_REJECTED", activeSessionID) - logoutBefore; delta != 1 {
		t.Fatalf("wrong csrf logout evidence grew beyond its durable bound: delta=%d", delta)
	}
	if count := authAdv01DeduplicationCount(t, ctx, db, "AUTH_LOGOUT_CSRF_REJECTED", activeSessionID); count != 1 {
		t.Fatalf("expected one wrong csrf logout deduplication gate, got %d", count)
	}
	if state := authAdv01SessionState(t, ctx, db, activeSessionID); state != "active" {
		t.Fatalf("wrong csrf mutated active session state to %q", state)
	}
	if count := authAdv01SessionCount(t, ctx, db, root.User.ID); count != beforeSessionCount {
		t.Fatalf("wrong csrf unexpectedly rotated a session: before=%d after=%d", beforeSessionCount, count)
	}
	if active := authAdv01ActiveFamilyCount(t, ctx, db, rootSessionID); active != beforeFamilyActive {
		t.Fatalf("wrong csrf unexpectedly revoked the active family member: before=%d after=%d", beforeFamilyActive, active)
	}

	crossRoot := authAdv01Register(t, ctx, serviceOne, "wrong-csrf-cross-instance")
	crossRootSessionID := authAdv01SessionID(t, ctx, db, crossRoot.User.ID)
	crossRotated, err := serviceOne.Refresh(ctx, crossRoot.RefreshToken, crossRoot.Session.CSRFToken)
	if err != nil {
		t.Fatalf("prepare cross-instance active descendant: %v", err)
	}
	crossActiveSessionID := authAdv01LatestSessionID(t, ctx, db, crossRoot.User.ID)
	crossBefore := authAdv01AuditCount(t, ctx, db, "AUTH_REFRESH_CSRF_REJECTED", crossActiveSessionID)
	authAdv01Concurrent(t, func(index int) error {
		service := serviceOne
		if index%2 == 1 {
			service = serviceTwo
		}
		_, err := service.Refresh(ctx, crossRotated.RefreshToken, "wrong-non-empty-csrf-token")
		if !errors.Is(err, auth.ErrInvalidSession) {
			return err
		}
		return nil
	})
	if delta := authAdv01AuditCount(t, ctx, db, "AUTH_REFRESH_CSRF_REJECTED", crossActiveSessionID) - crossBefore; delta != 1 {
		t.Fatalf("cross-instance wrong csrf evidence grew beyond its durable bound: delta=%d", delta)
	}
	if state := authAdv01SessionState(t, ctx, db, crossActiveSessionID); state != "active" {
		t.Fatalf("cross-instance wrong csrf mutated active session state to %q", state)
	}
	if active := authAdv01ActiveFamilyCount(t, ctx, db, crossRootSessionID); active != 1 {
		t.Fatalf("cross-instance wrong csrf revoked active family members: %d", active)
	}

	t.Log("AUTH_ADV_01_KNOWN_REFRESH_WRONG_CSRF_REFRESH_AUDITED=PASS")
	t.Log("AUTH_ADV_01_KNOWN_REFRESH_WRONG_CSRF_LOGOUT_AUDITED=PASS")
	t.Log("AUTH_ADV_01_KNOWN_REFRESH_WRONG_CSRF_REFRESH_BOUND=PASS")
	t.Log("AUTH_ADV_01_KNOWN_REFRESH_WRONG_CSRF_LOGOUT_BOUND=PASS")
	t.Log("AUTH_ADV_01_KNOWN_REFRESH_WRONG_CSRF_CROSS_INSTANCE_BOUND=PASS")
	t.Log("AUTH_ADV_01_WRONG_CSRF_DOES_NOT_ROTATE=PASS")
	t.Log("AUTH_ADV_01_WRONG_CSRF_DOES_NOT_REVOKE=PASS")
	t.Log("AUTH_ADV_01_WRONG_CSRF_DOES_NOT_REVOKE_FAMILY=PASS")
}

func TestAuthAdv01WrongCSRFAndValidMutationsSerializeSafely(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}

	storeOne, serviceOne := authAdv01StoreAndService(t, databaseURL)
	defer storeOne.Close()
	storeTwo, serviceTwo := authAdv01StoreAndService(t, databaseURL)
	defer storeTwo.Close()
	db := authAdv01OwnerDB(t, databaseURL)
	defer db.Close()
	ctx := context.Background()

	t.Run("refresh", func(t *testing.T) {
		root := authAdv01Register(t, ctx, serviceOne, "wrong-csrf-race-refresh")
		rootSessionID := authAdv01SessionID(t, ctx, db, root.User.ID)
		before := authAdv01AuditCount(t, ctx, db, "AUTH_REFRESH_CSRF_REJECTED", rootSessionID)
		authAdv01ConcurrentWithValid(t,
			func() error {
				_, err := serviceOne.Refresh(ctx, root.RefreshToken, root.Session.CSRFToken)
				return err
			},
			func(index int) error {
				service := serviceOne
				if index%2 == 1 {
					service = serviceTwo
				}
				_, err := service.Refresh(ctx, root.RefreshToken, "wrong-non-empty-csrf-token")
				if !errors.Is(err, auth.ErrInvalidSession) {
					return err
				}
				return nil
			},
		)
		if delta := authAdv01AuditCount(t, ctx, db, "AUTH_REFRESH_CSRF_REJECTED", rootSessionID) - before; delta != 1 {
			t.Fatalf("refresh race wrong csrf evidence grew beyond its durable bound: delta=%d", delta)
		}
		if count := authAdv01SessionCount(t, ctx, db, root.User.ID); count != 2 {
			t.Fatalf("refresh race created %d sessions, expected one valid rotation only", count)
		}
		if active := authAdv01ActiveFamilyCount(t, ctx, db, rootSessionID); active != 1 {
			t.Fatalf("refresh race left %d active family sessions, expected one", active)
		}
	})

	t.Run("logout", func(t *testing.T) {
		root := authAdv01Register(t, ctx, serviceOne, "wrong-csrf-race-logout")
		rootSessionID := authAdv01SessionID(t, ctx, db, root.User.ID)
		before := authAdv01AuditCount(t, ctx, db, "AUTH_LOGOUT_CSRF_REJECTED", rootSessionID)
		authAdv01ConcurrentWithValid(t,
			func() error {
				revoked, err := serviceOne.Logout(ctx, root.RefreshToken, root.Session.CSRFToken, false)
				if err != nil || !revoked {
					return errors.New("valid logout did not revoke its session")
				}
				return nil
			},
			func(index int) error {
				service := serviceOne
				if index%2 == 1 {
					service = serviceTwo
				}
				_, err := service.Logout(ctx, root.RefreshToken, "wrong-non-empty-csrf-token", false)
				if !errors.Is(err, auth.ErrInvalidSession) {
					return err
				}
				return nil
			},
		)
		if delta := authAdv01AuditCount(t, ctx, db, "AUTH_LOGOUT_CSRF_REJECTED", rootSessionID) - before; delta != 1 {
			t.Fatalf("logout race wrong csrf evidence grew beyond its durable bound: delta=%d", delta)
		}
		if state := authAdv01SessionState(t, ctx, db, rootSessionID); state != "revoked" {
			t.Fatalf("valid logout race left session state %q", state)
		}
		if active := authAdv01ActiveFamilyCount(t, ctx, db, rootSessionID); active != 0 {
			t.Fatalf("logout race left %d active family sessions", active)
		}
	})

	t.Log("AUTH_ADV_01_WRONG_CSRF_VALID_REFRESH_RACE=PASS")
	t.Log("AUTH_ADV_01_WRONG_CSRF_VALID_LOGOUT_RACE=PASS")
}

func TestAuthAdv01KnownSessionSecurityAuditIsDatabaseBounded(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}

	store, service := authAdv01StoreAndService(t, databaseURL)
	defer store.Close()
	db := authAdv01OwnerDB(t, databaseURL)
	defer db.Close()
	ctx := context.Background()

	root := authAdv01Register(t, ctx, service, "known-refresh")
	rootSessionID := authAdv01SessionID(t, ctx, db, root.User.ID)
	validRefreshBefore := authAdv01ActorAuditCount(t, ctx, db, "AUTH_REFRESH", root.User.ID)
	if _, err := service.Refresh(ctx, root.RefreshToken, root.Session.CSRFToken); err != nil {
		t.Fatalf("valid refresh: %v", err)
	}
	if delta := authAdv01ActorAuditCount(t, ctx, db, "AUTH_REFRESH", root.User.ID) - validRefreshBefore; delta != 1 {
		t.Fatalf("valid refresh audit was not preserved: delta=%d", delta)
	}

	replayBefore := authAdv01AuditCount(t, ctx, db, "AUTH_REFRESH_REPLAY", rootSessionID)
	for range authAdv01Requests {
		if _, err := service.Refresh(ctx, root.RefreshToken, root.Session.CSRFToken); !errors.Is(err, auth.ErrInvalidSession) {
			t.Fatalf("replayed refresh: %v", err)
		}
	}
	if delta := authAdv01AuditCount(t, ctx, db, "AUTH_REFRESH_REPLAY", rootSessionID) - replayBefore; delta != 1 {
		t.Fatalf("refresh replay evidence grew beyond its durable bound: delta=%d", delta)
	}
	if count := authAdv01DeduplicationCount(t, ctx, db, "AUTH_REFRESH_REPLAY", rootSessionID); count != 1 {
		t.Fatalf("expected one replay deduplication gate, got %d", count)
	}

	logoutReplayBefore := authAdv01AuditCount(t, ctx, db, "AUTH_LOGOUT_REPLAY", rootSessionID)
	for range authAdv01Requests {
		revoked, err := service.Logout(ctx, root.RefreshToken, root.Session.CSRFToken, false)
		if err != nil || !revoked {
			t.Fatalf("revoked-session logout: revoked=%t err=%v", revoked, err)
		}
	}
	if delta := authAdv01AuditCount(t, ctx, db, "AUTH_LOGOUT_REPLAY", rootSessionID) - logoutReplayBefore; delta != 1 {
		t.Fatalf("revoked logout evidence grew beyond its durable bound: delta=%d", delta)
	}
	if count := authAdv01DeduplicationCount(t, ctx, db, "AUTH_LOGOUT_REPLAY", rootSessionID); count != 1 {
		t.Fatalf("expected one logout replay deduplication gate, got %d", count)
	}

	expiredRefresh := authAdv01Register(t, ctx, service, "expired-refresh")
	expiredRefreshSessionID := authAdv01SessionID(t, ctx, db, expiredRefresh.User.ID)
	authAdv01ExpireSession(t, ctx, db, expiredRefreshSessionID)
	expiredRefreshBefore := authAdv01AuditCount(t, ctx, db, "AUTH_REFRESH_REJECTED", expiredRefreshSessionID)
	for range authAdv01Requests {
		if _, err := service.Refresh(ctx, expiredRefresh.RefreshToken, expiredRefresh.Session.CSRFToken); !errors.Is(err, auth.ErrInvalidSession) {
			t.Fatalf("expired refresh: %v", err)
		}
	}
	if delta := authAdv01AuditCount(t, ctx, db, "AUTH_REFRESH_REJECTED", expiredRefreshSessionID) - expiredRefreshBefore; delta != 1 {
		t.Fatalf("expired refresh evidence grew beyond its durable bound: delta=%d", delta)
	}
	if count := authAdv01DeduplicationCount(t, ctx, db, "AUTH_REFRESH_REJECTED", expiredRefreshSessionID); count != 1 {
		t.Fatalf("expected one expired refresh deduplication gate, got %d", count)
	}

	expiredLogout := authAdv01Register(t, ctx, service, "expired-logout")
	expiredLogoutSessionID := authAdv01SessionID(t, ctx, db, expiredLogout.User.ID)
	authAdv01ExpireSession(t, ctx, db, expiredLogoutSessionID)
	expiredLogoutBefore := authAdv01AuditCount(t, ctx, db, "AUTH_LOGOUT_REJECTED", expiredLogoutSessionID)
	for range authAdv01Requests {
		if _, err := service.Logout(ctx, expiredLogout.RefreshToken, expiredLogout.Session.CSRFToken, false); !errors.Is(err, auth.ErrInvalidSession) {
			t.Fatalf("expired logout: %v", err)
		}
	}
	if delta := authAdv01AuditCount(t, ctx, db, "AUTH_LOGOUT_REJECTED", expiredLogoutSessionID) - expiredLogoutBefore; delta != 1 {
		t.Fatalf("expired logout evidence grew beyond its durable bound: delta=%d", delta)
	}
	if count := authAdv01DeduplicationCount(t, ctx, db, "AUTH_LOGOUT_REJECTED", expiredLogoutSessionID); count != 1 {
		t.Fatalf("expected one expired logout deduplication gate, got %d", count)
	}

	logoutRoot := authAdv01Register(t, ctx, service, "valid-logout")
	logoutSessionID := authAdv01SessionID(t, ctx, db, logoutRoot.User.ID)
	validLogoutBefore := authAdv01AuditCount(t, ctx, db, "AUTH_LOGOUT", logoutSessionID)
	revoked, err := service.Logout(ctx, logoutRoot.RefreshToken, logoutRoot.Session.CSRFToken, false)
	if err != nil || !revoked {
		t.Fatalf("valid logout: revoked=%t err=%v", revoked, err)
	}
	if delta := authAdv01AuditCount(t, ctx, db, "AUTH_LOGOUT", logoutSessionID) - validLogoutBefore; delta != 1 {
		t.Fatalf("valid logout audit was not preserved: delta=%d", delta)
	}
	t.Log("AUTH_ADV_01_REFRESH_REPLAY_DURABLE_BOUND=PASS")
	t.Log("AUTH_ADV_01_REVOKED_LOGOUT_DURABLE_BOUND=PASS")
	t.Log("AUTH_ADV_01_EXPIRED_REFRESH_DURABLE_BOUND=PASS")
	t.Log("AUTH_ADV_01_EXPIRED_LOGOUT_DURABLE_BOUND=PASS")
	t.Log("AUTH_ADV_01_VALID_REFRESH_AUDIT_PRESERVED=PASS")
	t.Log("AUTH_ADV_01_VALID_LOGOUT_AUDIT_PRESERVED=PASS")
}

func TestAuthAdv01ConcurrentMultiInstanceSecurityAuditBound(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}

	storeOne, serviceOne := authAdv01StoreAndService(t, databaseURL)
	defer storeOne.Close()
	storeTwo, serviceTwo := authAdv01StoreAndService(t, databaseURL)
	defer storeTwo.Close()
	db := authAdv01OwnerDB(t, databaseURL)
	defer db.Close()
	ctx := context.Background()

	root := authAdv01Register(t, ctx, serviceOne, "concurrent-refresh")
	rootSessionID := authAdv01SessionID(t, ctx, db, root.User.ID)
	if _, err := serviceOne.Refresh(ctx, root.RefreshToken, root.Session.CSRFToken); err != nil {
		t.Fatalf("prepare revoked refresh session: %v", err)
	}
	replayBefore := authAdv01AuditCount(t, ctx, db, "AUTH_REFRESH_REPLAY", rootSessionID)
	authAdv01Concurrent(t, func(index int) error {
		service := serviceOne
		if index%2 == 1 {
			service = serviceTwo
		}
		_, err := service.Refresh(ctx, root.RefreshToken, root.Session.CSRFToken)
		if !errors.Is(err, auth.ErrInvalidSession) {
			return err
		}
		return nil
	})
	if delta := authAdv01AuditCount(t, ctx, db, "AUTH_REFRESH_REPLAY", rootSessionID) - replayBefore; delta != 1 {
		t.Fatalf("concurrent refresh replay evidence grew beyond its durable bound: delta=%d", delta)
	}
	var familyActive int
	if err := db.QueryRowContext(ctx, `
		SELECT count(*)
		FROM identity.sessions
		WHERE session_family_id = $1
			AND session_state = 'active'
	`, rootSessionID).Scan(&familyActive); err != nil {
		t.Fatalf("query compromised family: %v", err)
	}
	if familyActive != 0 {
		t.Fatalf("concurrent replay left %d active sessions in the compromised family", familyActive)
	}

	logoutRoot := authAdv01Register(t, ctx, serviceOne, "concurrent-logout")
	logoutSessionID := authAdv01SessionID(t, ctx, db, logoutRoot.User.ID)
	if revoked, err := serviceOne.Logout(ctx, logoutRoot.RefreshToken, logoutRoot.Session.CSRFToken, false); err != nil || !revoked {
		t.Fatalf("prepare revoked logout session: revoked=%t err=%v", revoked, err)
	}
	logoutReplayBefore := authAdv01AuditCount(t, ctx, db, "AUTH_LOGOUT_REPLAY", logoutSessionID)
	authAdv01Concurrent(t, func(index int) error {
		service := serviceOne
		if index%2 == 1 {
			service = serviceTwo
		}
		revoked, err := service.Logout(ctx, logoutRoot.RefreshToken, logoutRoot.Session.CSRFToken, false)
		if err != nil {
			return err
		}
		if !revoked {
			return errors.New("concurrent revoked logout did not preserve idempotent response")
		}
		return nil
	})
	if delta := authAdv01AuditCount(t, ctx, db, "AUTH_LOGOUT_REPLAY", logoutSessionID) - logoutReplayBefore; delta != 1 {
		t.Fatalf("concurrent revoked logout evidence grew beyond its durable bound: delta=%d", delta)
	}
	t.Log("AUTH_ADV_01_CONCURRENT_REPLAY_DURABLE_BOUND=PASS")
	t.Log("AUTH_ADV_01_CONCURRENT_REVOKED_LOGOUT_BOUND=PASS")
	t.Log("AUTH_ADV_01_MULTI_INSTANCE_DB_BOUND=PASS")
	t.Log("AUTH_ADV_01_SESSION_FAMILY_CONTAINMENT_PRESERVED=PASS")
}

func TestAuthAdv01RuntimeAuditPrivilegesRemainAppendOnly(t *testing.T) {
	ownerURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	runtimeURL := os.Getenv("OPENINVEST_DATABASE_RUNTIME_TEST_URL")
	if ownerURL == "" || runtimeURL == "" {
		t.Skip("runtime privilege integration URLs are not set")
	}
	ctx := context.Background()
	ownerDB := authAdv01OwnerDB(t, ownerURL)
	defer ownerDB.Close()
	runtimeDB, err := sql.Open("pgx", runtimeURL)
	if err != nil {
		t.Fatalf("open runtime db: %v", err)
	}
	defer runtimeDB.Close()

	actionCode := "AUTH_ADV_01_RUNTIME_" + uuid.NewString()
	sessionID := uuid.NewString()
	if _, err := runtimeDB.ExecContext(ctx, `
		INSERT INTO audit.auth_security_event_deduplications (action_code, session_id, created_at)
		VALUES ($1, $2, now())
		ON CONFLICT (action_code, session_id) DO NOTHING
	`, actionCode, sessionID); err != nil {
		t.Fatalf("runtime INSERT deduplication gate: %v", err)
	}
	t.Cleanup(func() {
		_, _ = ownerDB.ExecContext(context.Background(), `
			DELETE FROM audit.auth_security_event_deduplications
			WHERE action_code = $1 AND session_id = $2
		`, actionCode, sessionID)
	})

	for name, query := range map[string]string{
		"audit.events update":          `UPDATE audit.events SET action_code = action_code`,
		"audit.events delete":          `DELETE FROM audit.events`,
		"audit.events truncate":        `TRUNCATE audit.events`,
		"audit deduplication update":   `UPDATE audit.auth_security_event_deduplications SET action_code = action_code`,
		"audit deduplication delete":   `DELETE FROM audit.auth_security_event_deduplications`,
		"audit deduplication truncate": `TRUNCATE audit.auth_security_event_deduplications`,
	} {
		if _, err := runtimeDB.ExecContext(ctx, query); err == nil {
			t.Fatalf("%s unexpectedly succeeded", name)
		}
	}
	t.Log("AUTH_ADV_01_AUDIT_APPEND_ONLY_PRIVILEGES_PRESERVED=PASS")
}

func authAdv01StoreAndService(t *testing.T, databaseURL string) (*postgres.Store, *auth.Service) {
	t.Helper()
	store, err := postgres.Open(databaseURL)
	if err != nil {
		t.Fatalf("open postgres store: %v", err)
	}
	service, err := auth.NewService(store, verticalslice.SystemClock{}, auth.Config{
		AccessTokenSecret:   []byte("01234567890123456789012345678901"),
		RefreshCookieSecure: true,
	})
	if err != nil {
		_ = store.Close()
		t.Fatalf("new auth service: %v", err)
	}
	return store, service
}

func authAdv01OwnerDB(t *testing.T, databaseURL string) *sql.DB {
	t.Helper()
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open owner db: %v", err)
	}
	return db
}

func authAdv01Register(t *testing.T, ctx context.Context, service *auth.Service, label string) auth.AuthResult {
	t.Helper()
	result, err := service.Register(ctx, auth.RegistrationRequest{
		Email:    "auth-adv-01-" + label + "-" + uuid.NewString() + "@example.com",
		Password: "correct horse battery staple",
		Language: auth.LanguageEN,
		Theme:    auth.ThemeSystem,
		Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("register %s: %v", label, err)
	}
	return result
}

func authAdv01SessionID(t *testing.T, ctx context.Context, db *sql.DB, userID string) string {
	t.Helper()
	var sessionID string
	if err := db.QueryRowContext(ctx, `
		SELECT id
		FROM identity.sessions
		WHERE user_id = $1
		ORDER BY created_at, id
		LIMIT 1
	`, userID).Scan(&sessionID); err != nil {
		t.Fatalf("lookup session id: %v", err)
	}
	return sessionID
}

func authAdv01LatestSessionID(t *testing.T, ctx context.Context, db *sql.DB, userID string) string {
	t.Helper()
	var sessionID string
	if err := db.QueryRowContext(ctx, `
		SELECT id
		FROM identity.sessions
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT 1
	`, userID).Scan(&sessionID); err != nil {
		t.Fatalf("lookup latest session id: %v", err)
	}
	return sessionID
}

func authAdv01SessionCount(t *testing.T, ctx context.Context, db *sql.DB, userID string) int {
	t.Helper()
	var count int
	if err := db.QueryRowContext(ctx, `
		SELECT count(*)
		FROM identity.sessions
		WHERE user_id = $1
	`, userID).Scan(&count); err != nil {
		t.Fatalf("count user sessions: %v", err)
	}
	return count
}

func authAdv01SessionState(t *testing.T, ctx context.Context, db *sql.DB, sessionID string) string {
	t.Helper()
	var state string
	if err := db.QueryRowContext(ctx, `
		SELECT session_state
		FROM identity.sessions
		WHERE id = $1
	`, sessionID).Scan(&state); err != nil {
		t.Fatalf("read session state: %v", err)
	}
	return state
}

func authAdv01ActiveFamilyCount(t *testing.T, ctx context.Context, db *sql.DB, sessionFamilyID string) int {
	t.Helper()
	var count int
	if err := db.QueryRowContext(ctx, `
		SELECT count(*)
		FROM identity.sessions
		WHERE session_family_id = $1
			AND session_state = 'active'
	`, sessionFamilyID).Scan(&count); err != nil {
		t.Fatalf("count active session family: %v", err)
	}
	return count
}

func authAdv01ExpireSession(t *testing.T, ctx context.Context, db *sql.DB, sessionID string) {
	t.Helper()
	result, err := db.ExecContext(ctx, `
		UPDATE identity.sessions
		SET expires_at = created_at + interval '1 microsecond'
		WHERE id = $1
	`, sessionID)
	if err != nil {
		t.Fatalf("expire session %s: %v", sessionID, err)
	}
	changed, err := result.RowsAffected()
	if err != nil {
		t.Fatalf("read expired-session row count: %v", err)
	}
	if changed != 1 {
		t.Fatalf("expire session %s changed %d rows", sessionID, changed)
	}
}

func authAdv01AllAuditCount(t *testing.T, ctx context.Context, db *sql.DB) int {
	t.Helper()
	var count int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM audit.events`).Scan(&count); err != nil {
		t.Fatalf("count audit events: %v", err)
	}
	return count
}

func authAdv01AuditCount(t *testing.T, ctx context.Context, db *sql.DB, actionCode string, sessionID string) int {
	t.Helper()
	var count int
	if err := db.QueryRowContext(ctx, `
		SELECT count(*)
		FROM audit.events
		WHERE action_code = $1 AND target_kind = 'session' AND target_id = $2
	`, actionCode, sessionID).Scan(&count); err != nil {
		t.Fatalf("count %s audit events: %v", actionCode, err)
	}
	return count
}

func authAdv01ActorAuditCount(t *testing.T, ctx context.Context, db *sql.DB, actionCode string, actorID string) int {
	t.Helper()
	var count int
	if err := db.QueryRowContext(ctx, `
		SELECT count(*)
		FROM audit.events
		WHERE action_code = $1 AND actor_id = $2
	`, actionCode, actorID).Scan(&count); err != nil {
		t.Fatalf("count %s actor audit events: %v", actionCode, err)
	}
	return count
}

func authAdv01DeduplicationCount(t *testing.T, ctx context.Context, db *sql.DB, actionCode string, sessionID string) int {
	t.Helper()
	var count int
	if err := db.QueryRowContext(ctx, `
		SELECT count(*)
		FROM audit.auth_security_event_deduplications
		WHERE action_code = $1 AND session_id = $2
	`, actionCode, sessionID).Scan(&count); err != nil {
		t.Fatalf("count %s deduplication gates: %v", actionCode, err)
	}
	return count
}

func authAdv01Concurrent(t *testing.T, call func(int) error) {
	t.Helper()
	start := make(chan struct{})
	errs := make(chan error, authAdv01Requests)
	var ready sync.WaitGroup
	ready.Add(authAdv01Requests)
	for index := range authAdv01Requests {
		go func(index int) {
			ready.Done()
			<-start
			errs <- call(index)
		}(index)
	}
	ready.Wait()
	close(start)
	for range authAdv01Requests {
		if err := <-errs; err != nil {
			t.Fatalf("concurrent auth call failed: %v", err)
		}
	}
}

func authAdv01ConcurrentWithValid(t *testing.T, valid func() error, wrong func(int) error) {
	t.Helper()
	start := make(chan struct{})
	errs := make(chan error, authAdv01Requests+1)
	var ready sync.WaitGroup
	ready.Add(authAdv01Requests + 1)
	go func() {
		ready.Done()
		<-start
		errs <- valid()
	}()
	for index := range authAdv01Requests {
		go func(index int) {
			ready.Done()
			<-start
			errs <- wrong(index)
		}(index)
	}
	ready.Wait()
	close(start)
	for range authAdv01Requests + 1 {
		if err := <-errs; err != nil {
			t.Fatalf("concurrent auth mutation failed: %v", err)
		}
	}
}
