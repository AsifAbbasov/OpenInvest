package auth

import (
	"context"
	"errors"
	"testing"
	"time"
)

type ownershipFixture struct {
	memoryStore
	begin     int
	hash      string
	completed bool
	delivery  bool
}

func (s *ownershipFixture) BeginRegistration(_ context.Context, email string) error {
	s.begin++
	return nil
}
func (s *ownershipFixture) CompleteRegistration(_ context.Context, hash string, r RegistrationRecord) error {
	if hash != s.hash || s.completed {
		return ErrInvalidVerification
	}
	s.completed = true
	if r.PasswordHash == "" || r.EmailNormalized != "" {
		return ErrInvalidInput
	}
	return nil
}
func (s *ownershipFixture) LeaseVerificationDelivery(_ context.Context, hash string) (VerificationDelivery, bool, error) {
	if s.delivery {
		return VerificationDelivery{}, false, nil
	}
	s.hash = hash
	s.delivery = true
	return VerificationDelivery{Email: "owner@example.test", ExpiresAt: time.Now().Add(time.Minute)}, true, nil
}
func (s *ownershipFixture) FinishVerificationDelivery(context.Context, VerificationDelivery, bool) error {
	return nil
}

type mailFixture struct {
	raw     string
	failure bool
}

func (m *mailFixture) SendVerification(_ context.Context, email, raw string) error {
	m.raw = raw
	if m.failure {
		return errors.New("synthetic transport diagnostic secret")
	}
	return nil
}
func TestEmailOwnershipLifecycleDoesNotIssueSession(t *testing.T) {
	store := &ownershipFixture{}
	s, err := NewService(store, nil, Config{AccessTokenSecret: []byte("01234567890123456789012345678901")})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := s.BeginEmailRegistration(ctx, " OWNER@example.test "); err != nil {
		t.Fatal(err)
	}
	if store.user.ID != "" {
		t.Fatal("begin created user")
	}
	mail := &mailFixture{}
	if _, err := s.DeliverOne(ctx, mail); err != nil {
		t.Fatal(err)
	}
	if len(mail.raw) != 43 || store.hash == mail.raw || len(store.hash) != 64 {
		t.Fatal("token entropy/storage contract")
	}
	request := RegistrationRequest{Password: "owner selected strong password", Language: "en", Theme: "system", Timezone: "UTC"}
	if err := s.CompleteEmailRegistration(ctx, mail.raw, request); err != nil {
		t.Fatal(err)
	}
	if err := s.CompleteEmailRegistration(ctx, mail.raw, request); !errors.Is(err, ErrInvalidVerification) {
		t.Fatal("not single-use")
	}
}
func TestEmailOwnershipDeliveryFailureDoesNotActivate(t *testing.T) {
	store := &ownershipFixture{}
	s, err := NewService(store, nil, Config{AccessTokenSecret: []byte("01234567890123456789012345678901")})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.BeginEmailRegistration(context.Background(), "owner@example.test"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.DeliverOne(context.Background(), &mailFixture{failure: true}); !errors.Is(err, ErrVerificationUnavailable) {
		t.Fatal("delivery must fail closed")
	}
	if store.completed || store.user.ID != "" {
		t.Fatal("failed delivery activated")
	}
}
