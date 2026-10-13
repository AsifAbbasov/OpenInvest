package postgres_test

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/google/uuid"
	"github.com/openinvest/openinvest/backend-go/internal/auth"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
)

type authAdv02Mail struct {
	email string
	token string
	fail  bool
}

func (m *authAdv02Mail) SendVerification(_ context.Context, email, token string) error {
	m.email, m.token = email, token
	if m.fail {
		return errors.New("synthetic delivery failure")
	}
	return nil
}

func authAdv02Open(t *testing.T) (*auth.Service, *postgres.Store, *sql.DB) {
	t.Helper()
	url := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if url == "" {
		t.Skip("isolated PostgreSQL service required")
	}
	store, err := postgres.Open(url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	svc, err := auth.NewService(store, nil, auth.Config{AccessTokenSecret: []byte("01234567890123456789012345678901")})
	if err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if _, err := db.Exec(`DELETE FROM identity.registration_intents`); err != nil { t.Fatal(err) }
	return svc, store, db
}

func authAdv02OwnerRequest(password string) auth.RegistrationRequest {
	return auth.RegistrationRequest{Password: password, Language: "en", Theme: "system", Timezone: "UTC"}
}

func TestAuthAdv02NoIdentitySessionOrCredentialBeforeProof(t *testing.T) {
	svc, _, db := authAdv02Open(t)
	ctx := context.Background()
	email := "adv02-preproof-" + uuid.NewString() + "@example.test"
	if err := svc.BeginEmailRegistration(ctx, email); err != nil {
		t.Fatal(err)
	}
	var users, creds, sessions, links, subjects int
	if err := db.QueryRow(`SELECT count(*) FROM identity.users WHERE email_normalized=$1`, email).Scan(&users); err != nil { t.Fatal(err) }
	if err := db.QueryRow(`SELECT count(*) FROM identity.credentials c JOIN identity.users u ON u.id=c.user_id WHERE u.email_normalized=$1`, email).Scan(&creds); err != nil { t.Fatal(err) }
	if err := db.QueryRow(`SELECT count(*) FROM identity.sessions s JOIN identity.users u ON u.id=s.user_id WHERE u.email_normalized=$1`, email).Scan(&sessions); err != nil { t.Fatal(err) }
	if err := db.QueryRow(`SELECT count(*) FROM identity.user_investment_links l JOIN identity.users u ON u.id=l.user_id WHERE u.email_normalized=$1`, email).Scan(&links); err != nil { t.Fatal(err) }
	if err := db.QueryRow(`SELECT count(*) FROM investment.subjects s JOIN identity.user_investment_links l ON l.investment_subject_id=s.id JOIN identity.users u ON u.id=l.user_id WHERE u.email_normalized=$1`, email).Scan(&subjects); err != nil { t.Fatal(err) }
	if users+creds+sessions+links+subjects != 0 {
		t.Fatalf("pre-proof state created users=%d creds=%d sessions=%d links=%d subjects=%d", users, creds, sessions, links, subjects)
	}
	t.Log("AUTH_ADV_02_PREPROOF_ACTIVATION=BLOCKED")
}

func TestAuthAdv02TokenGenerationReplayAndSubstitutionFailClosed(t *testing.T) {
	svc, _, db := authAdv02Open(t)
	ctx := context.Background()
	email := "adv02-token-" + uuid.NewString() + "@example.test"
	if err := svc.BeginEmailRegistration(ctx, email); err != nil { t.Fatal(err) }
	mail := &authAdv02Mail{}
	if found, err := svc.DeliverOne(ctx, mail); err != nil || !found { t.Fatalf("deliver found=%t err=%v", found, err) }
	if mail.token == "" { t.Fatal("missing synthetic verification token") }

	// Supersede the generation without modifying production logic; the old token must become useless.
	if _, err := db.Exec(`UPDATE identity.registration_intents SET generation=generation+1 WHERE email_normalized=$1`, email); err != nil { t.Fatal(err) }
	if err := svc.CompleteEmailRegistration(ctx, mail.token, authAdv02OwnerRequest("owner-selected-password-1")); !errors.Is(err, auth.ErrInvalidVerification) {
		t.Fatalf("superseded token accepted: %v", err)
	}
	var users int
	if err := db.QueryRow(`SELECT count(*) FROM identity.users WHERE email_normalized=$1`, email).Scan(&users); err != nil || users != 0 {
		t.Fatalf("superseded token activated user=%d err=%v", users, err)
	}

	// Database purpose constraint must reject substituted token purposes.
	if _, err := db.Exec(`UPDATE identity.email_verification_tokens SET purpose='password_reset' WHERE intent_id=(SELECT id FROM identity.registration_intents WHERE email_normalized=$1)`, email); err == nil {
		t.Fatal("wrong-purpose verification token accepted by database constraint")
	}

	// Fresh intent/token activates once, then replay is rejected.
	if _, err := db.Exec(`UPDATE identity.registration_intents SET expires_at=statement_timestamp()-interval '1 second', intent_state='expired' WHERE email_normalized=$1`, email); err != nil { t.Fatal(err) }
	if err := svc.BeginEmailRegistration(ctx, email); err != nil { t.Fatal(err) }
	mail2 := &authAdv02Mail{}
	if found, err := svc.DeliverOne(ctx, mail2); err != nil || !found { t.Fatalf("redeliver found=%t err=%v", found, err) }
	if err := svc.CompleteEmailRegistration(ctx, mail2.token, authAdv02OwnerRequest("owner-selected-password-2")); err != nil { t.Fatal(err) }
	if err := svc.CompleteEmailRegistration(ctx, mail2.token, authAdv02OwnerRequest("owner-selected-password-3")); !errors.Is(err, auth.ErrInvalidVerification) {
		t.Fatalf("replayed token accepted: %v", err)
	}
	var sessions int
	if err := db.QueryRow(`SELECT count(*) FROM identity.sessions s JOIN identity.users u ON u.id=s.user_id WHERE u.email_normalized=$1`, email).Scan(&sessions); err != nil || sessions != 0 {
		t.Fatalf("verification issued session=%d err=%v", sessions, err)
	}
	t.Log("TOKEN_SUBSTITUTION=REJECTED TOKEN_REPLAY=REJECTED SUPERSEDED_GENERATION=REJECTED")
}

func TestAuthAdv02CrossPrincipalVerificationCannotTransferLegacyIdentity(t *testing.T) {
	svc, _, db := authAdv02Open(t)
	ctx := context.Background()
	email := "adv02-legacy-" + uuid.NewString() + "@example.test"
	legacyPassword := "legacy-owner-password-123"
	result, err := svc.Register(ctx, auth.RegistrationRequest{Email: email, Password: legacyPassword, Language: "en", Theme: "system", Timezone: "UTC"})
	if err != nil { t.Fatal(err) }
	legacyUser := result.User.ID
	legacySubject := result.User.InvestmentSubjectID
	var passwordHash string
	if err := db.QueryRow(`SELECT password_hash FROM identity.credentials WHERE user_id=$1`, legacyUser).Scan(&passwordHash); err != nil { t.Fatal(err) }
	var sessionCount int
	if err := db.QueryRow(`SELECT count(*) FROM identity.sessions WHERE user_id=$1`, legacyUser).Scan(&sessionCount); err != nil { t.Fatal(err) }

	if err := svc.BeginEmailRegistration(ctx, email); err != nil { t.Fatal(err) }
	mail := &authAdv02Mail{}
	if found, err := svc.DeliverOne(ctx, mail); err != nil || !found { t.Fatalf("deliver existing found=%t err=%v", found, err) }
	if err := svc.CompleteEmailRegistration(ctx, mail.token, authAdv02OwnerRequest("attacker-new-password-123")); !errors.Is(err, auth.ErrInvalidVerification) {
		t.Fatalf("existing identity overwritten by verification: %v", err)
	}

	var userID, subjectID, hashAfter string
	if err := db.QueryRow(`SELECT u.id,l.investment_subject_id,c.password_hash FROM identity.users u JOIN identity.user_investment_links l ON l.user_id=u.id JOIN identity.credentials c ON c.user_id=u.id WHERE u.email_normalized=$1`, email).Scan(&userID, &subjectID, &hashAfter); err != nil { t.Fatal(err) }
	var sessionsAfter int
	if err := db.QueryRow(`SELECT count(*) FROM identity.sessions WHERE user_id=$1`, legacyUser).Scan(&sessionsAfter); err != nil { t.Fatal(err) }
	if userID != legacyUser || subjectID != legacySubject || hashAfter != passwordHash || sessionsAfter != sessionCount {
		t.Fatalf("legacy identity mutated user=%s/%s subject=%s/%s sessions=%d/%d", userID, legacyUser, subjectID, legacySubject, sessionsAfter, sessionCount)
	}
	var consumed sql.NullTime
	if err := db.QueryRow(`SELECT consumed_at FROM identity.registration_intents WHERE email_normalized=$1`, email).Scan(&consumed); err != nil { t.Fatal(err) }
	if consumed.Valid { t.Fatal("failed legacy takeover consumed intent") }
	if _, err := svc.Login(ctx, auth.LoginRequest{Email: email, Password: legacyPassword}); err != nil { t.Fatalf("legacy password stopped working: %v", err) }
	if _, err := svc.Login(ctx, auth.LoginRequest{Email: email, Password: "attacker-new-password-123"}); !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("attacker password installed: %v", err)
	}
	t.Log("LEGACY_ACCOUNT_ISOLATION=PASS EXISTING_PASSWORD_AND_SUBJECT_PRESERVED=YES")
}

func TestAuthAdv02DeliveryLeaseRaceExpiryAndAttemptBound(t *testing.T) {
	_, store, db := authAdv02Open(t)
	ctx := context.Background()
	email := "adv02-delivery-" + uuid.NewString() + "@example.test"
	if err := store.BeginRegistration(ctx, email); err != nil { t.Fatal(err) }

	digests := make([]string, 6)
	for i := range digests {
		sum := sha256.Sum256([]byte(email + string(rune('A'+i))))
		digests[i] = hex.EncodeToString(sum[:])
	}
	var wg sync.WaitGroup
	type leaseResult struct{ found bool; err error }
	results := make(chan leaseResult, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, found, err := store.LeaseVerificationDelivery(ctx, digests[i])
			results <- leaseResult{found, err}
		}(i)
	}
	wg.Wait(); close(results)
	winners := 0
	for r := range results {
		if r.err != nil { t.Fatal(r.err) }
		if r.found { winners++ }
	}
	if winners != 1 { t.Fatalf("concurrent lease winners=%d", winners) }

	for attempt := 2; attempt <= 4; attempt++ {
		if _, err := db.Exec(`UPDATE identity.email_delivery_jobs SET lease_until=statement_timestamp()-interval '1 second', available_at=statement_timestamp()-interval '1 second' WHERE intent_id=(SELECT id FROM identity.registration_intents WHERE email_normalized=$1)`, email); err != nil { t.Fatal(err) }
		if _, found, err := store.LeaseVerificationDelivery(ctx, digests[attempt]); err != nil || !found { t.Fatalf("lease recovery attempt=%d found=%t err=%v", attempt, found, err) }
	}
	if _, err := db.Exec(`UPDATE identity.email_delivery_jobs SET lease_until=statement_timestamp()-interval '1 second', available_at=statement_timestamp()-interval '1 second' WHERE intent_id=(SELECT id FROM identity.registration_intents WHERE email_normalized=$1)`, email); err != nil { t.Fatal(err) }
	if _, found, err := store.LeaseVerificationDelivery(ctx, digests[5]); err != nil || found { t.Fatalf("attempt limit bypass found=%t err=%v", found, err) }
	var attempts int
	if err := db.QueryRow(`SELECT attempts FROM identity.email_delivery_jobs WHERE intent_id=(SELECT id FROM identity.registration_intents WHERE email_normalized=$1)`, email).Scan(&attempts); err != nil { t.Fatal(err) }
	if attempts != 4 { t.Fatalf("attempts=%d want=4", attempts) }
	var users int
	if err := db.QueryRow(`SELECT count(*) FROM identity.users WHERE email_normalized=$1`, email).Scan(&users); err != nil || users != 0 { t.Fatalf("delivery race activated user=%d err=%v", users, err) }
	t.Log("DELIVERY_WORKER_RACES=PASS LEASE_SINGLE_WINNER=YES LEASE_EXPIRY_RECOVERY=YES ATTEMPTS_BOUNDED_4=YES")
}

func TestAuthAdv02RuntimeRoleCannotMutateLedgerAuditOrDDL(t *testing.T) {
	url := os.Getenv("OPENINVEST_DATABASE_RUNTIME_TEST_URL")
	if url == "" { t.Skip("isolated runtime role required") }
	db, err := sql.Open("pgx", url)
	if err != nil { t.Fatal(err) }
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for name, query := range map[string]string{
		"ledger_delete": `DELETE FROM investment.transaction_entries WHERE false`,
		"audit_delete": `DELETE FROM audit.events WHERE false`,
		"ddl_create": `CREATE TABLE identity.auth_adv_02_forbidden(id integer)`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := db.ExecContext(ctx, query); err == nil {
				t.Fatalf("runtime role unexpectedly allowed %s", name)
			}
		})
	}
	t.Log("RUNTIME_ROLE_PRIVILEGES=PASS_FOR_DANGEROUS_NEGATIVE_CONTROLS")
}
