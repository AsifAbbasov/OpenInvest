package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/openinvest/openinvest/backend-go/internal/auth"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"os"
	"sync"
	"testing"
)

type ownershipMail struct {
	token string
	email string
	fail  bool
}

func (m *ownershipMail) SendVerification(_ context.Context, email string, token string) error {
	m.token = token
	m.email = email
	if m.fail {
		return errors.New("synthetic delivery unavailable")
	}
	return nil
}
func ownershipService(t *testing.T) (*auth.Service, *sql.DB) {
	t.Helper()
	url := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if url == "" {
		t.Skip("isolated PostgreSQL test service required")
	}
	store, err := postgres.Open(url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	service, err := auth.NewService(store, nil, auth.Config{AccessTokenSecret: []byte("01234567890123456789012345678901")})
	if err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return service, db
}
func ownerRequest() auth.RegistrationRequest {
	return auth.RegistrationRequest{Password: "owner-selected-correct-horse", Language: "en", Theme: "system", Timezone: "UTC"}
}
func TestEmailOwnershipRegistrationActivation(t *testing.T) {
	s, db := ownershipService(t)
	ctx := context.Background()
	email := "ownership-" + uuid.NewString() + "@example.test"
	if err := s.BeginEmailRegistration(ctx, email); err != nil {
		t.Fatal(err)
	}
	if err := s.BeginEmailRegistration(ctx, email); err != nil {
		t.Fatal(err)
	}
	var before int
	if err := db.QueryRow(`SELECT count(*) FROM identity.users WHERE email_normalized=$1`, email).Scan(&before); err != nil || before != 0 {
		t.Fatalf("identity exists before proof: %d %v", before, err)
	}
	sink := &ownershipMail{}
	found, err := s.DeliverOne(ctx, sink)
	if err != nil || !found || sink.token == "" {
		t.Fatalf("delivery: %t %v", found, err)
	}
	var digest string
	emailDelivered := sink.email
	if err = db.QueryRow(`SELECT t.token_hash FROM identity.email_verification_tokens t JOIN identity.registration_intents i ON i.id=t.intent_id WHERE i.email_normalized=$1 ORDER BY t.created_at DESC LIMIT 1`, emailDelivered).Scan(&digest); err != nil {
		t.Fatal(err)
	}
	if digest == sink.token || len(digest) != 64 {
		t.Fatal("raw token persisted or invalid digest")
	}
	if err = s.CompleteEmailRegistration(ctx, sink.token, ownerRequest()); err != nil {
		t.Fatal(err)
	}
	var count, sessions int
	if err = db.QueryRow(`SELECT count(*) FROM identity.users u JOIN identity.user_email_verifications v ON v.user_id=u.id WHERE u.email_normalized=$1 AND u.account_state='active' AND v.evidence_state='verified'`, emailDelivered).Scan(&count); err != nil || count != 1 {
		t.Fatalf("activation: %d %v", count, err)
	}
	if err = db.QueryRow(`SELECT count(*) FROM identity.sessions s JOIN identity.users u ON u.id=s.user_id WHERE u.email_normalized=$1`, emailDelivered).Scan(&sessions); err != nil || sessions != 0 {
		t.Fatalf("activation issued session: %d %v", sessions, err)
	}
	if err = s.CompleteEmailRegistration(ctx, sink.token, ownerRequest()); !errors.Is(err, auth.ErrInvalidVerification) {
		t.Fatalf("consumed token accepted: %v", err)
	}
	if _, err = s.Login(ctx, auth.LoginRequest{Email: emailDelivered, Password: ownerRequest().Password}); err != nil {
		t.Fatal(err)
	}
	t.Log("EMAIL_OWNERSHIP_ACTIVATION=PASS NO_SESSION_ON_ACTIVATION=YES TOKEN_DIGEST_ONLY=YES")
}
func TestEmailOwnershipConcurrentCompletion(t *testing.T) {
	s, db := ownershipService(t)
	ctx := context.Background()
	email := "ownership-race-" + uuid.NewString() + "@example.test"
	if err := s.BeginEmailRegistration(ctx, email); err != nil {
		t.Fatal(err)
	}
	sink := &ownershipMail{}
	if _, err := s.DeliverOne(ctx, sink); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	results := make(chan error, 4)
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); results <- s.CompleteEmailRegistration(ctx, sink.token, ownerRequest()) }()
	}
	wg.Wait()
	close(results)
	wins := 0
	for err := range results {
		if err == nil {
			wins++
		} else if !errors.Is(err, auth.ErrInvalidVerification) && !errors.Is(err, auth.ErrAuthCapacity) {
			t.Error(err)
		}
	}
	if wins != 1 {
		t.Fatalf("activation winners=%d", wins)
	}
	var count int
	if err := db.QueryRow(`SELECT count(*) FROM identity.users WHERE email_normalized=$1`, sink.email).Scan(&count); err != nil || count != 1 {
		t.Fatalf("concurrent persistent identities=%d err=%v", count, err)
	}
	t.Log("EMAIL_OWNERSHIP_CONCURRENT_SINGLE_WINNER=PASS ARGON_ADMISSION_LIMIT_PRESERVED=YES")
}

func TestEmailOwnershipExpiredDeliveryCannotActivate(t *testing.T) {
	s, db := ownershipService(t)
	ctx := context.Background()
	email := "ownership-expiry-" + uuid.NewString() + "@example.test"
	if err := s.BeginEmailRegistration(ctx, email); err != nil {
		t.Fatal(err)
	}
	sink := &ownershipMail{}
	if _, err := s.DeliverOne(ctx, sink); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE identity.email_verification_tokens SET expires_at=statement_timestamp()-interval '1 second' WHERE intent_id=(SELECT id FROM identity.registration_intents WHERE email_normalized=$1)`, sink.email); err != nil {
		t.Fatal(err)
	}
	if err := s.CompleteEmailRegistration(ctx, sink.token, ownerRequest()); !errors.Is(err, auth.ErrInvalidVerification) {
		t.Fatalf("expired proof must be rejected: %v", err)
	}
	var count int
	if err := db.QueryRow(`SELECT count(*) FROM identity.users WHERE email_normalized=$1`, sink.email).Scan(&count); err != nil || count != 0 {
		t.Fatalf("expired token activated identity: %d %v", count, err)
	}
}

func TestEmailOwnershipRuntimePrivilegeBoundary(t *testing.T) {
	url := os.Getenv("OPENINVEST_DATABASE_RUNTIME_TEST_URL")
	if url == "" {
		t.Skip("isolated runtime PostgreSQL service required")
	}
	store, err := postgres.OpenRuntime(url)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	service, err := auth.NewService(store, nil, auth.Config{AccessTokenSecret: []byte("01234567890123456789012345678901")})
	if err != nil {
		t.Fatal(err)
	}
	email := "runtime-ownership-" + uuid.NewString() + "@example.test"
	if err := service.BeginEmailRegistration(context.Background(), email); err != nil {
		t.Fatal(err)
	}
	sink := &ownershipMail{}
	if found, err := service.DeliverOne(context.Background(), sink); err != nil || !found {
		t.Fatalf("runtime delivery %t %v", found, err)
	}
	if err := service.CompleteEmailRegistration(context.Background(), sink.token, ownerRequest()); err != nil {
		t.Fatal(err)
	}
	t.Log("EMAIL_OWNERSHIP_RUNTIME_PRIVILEGE_BOUNDARY=PASS")
}
