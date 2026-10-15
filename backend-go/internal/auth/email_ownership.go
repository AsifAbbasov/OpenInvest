package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"
)

var ErrVerificationUnavailable = errors.New("email verification unavailable")
var ErrInvalidVerification = errors.New("invalid email verification")

// OwnershipStore is intentionally separate from the legacy session Store.
// Implementations must perform CompleteRegistration atomically and create no session.
type OwnershipStore interface {
	BeginRegistration(context.Context, string) error
	CompleteRegistration(context.Context, string, RegistrationRecord) error
	LeaseVerificationDelivery(context.Context, string) (VerificationDelivery, bool, error)
	FinishVerificationDelivery(context.Context, VerificationDelivery, bool) error
}
type VerificationDelivery struct {
	JobID, LeaseID, Email string
	ExpiresAt             time.Time
}
type VerificationMailer interface {
	SendVerification(context.Context, string, string) error
}

func verificationDigest(raw string) (string, error) {
	b, err := base64.RawURLEncoding.Strict().DecodeString(raw)
	if err != nil || len(b) != 32 || base64.RawURLEncoding.EncodeToString(b) != raw {
		return "", ErrInvalidVerification
	}
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:]), nil
}
func newVerificationToken() (string, string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	raw := base64.RawURLEncoding.EncodeToString(b)
	digest, err := verificationDigest(raw)
	return raw, digest, err
}
func (s *Service) BeginEmailRegistration(ctx context.Context, email string) error {
	email = normalizeEmail(email)
	if !validNormalizedEmail(email) {
		return ErrInvalidInput
	}
	store, ok := s.store.(OwnershipStore)
	if !ok {
		return ErrVerificationUnavailable
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return store.BeginRegistration(ctx, email)
}
func (s *Service) CompleteEmailRegistration(ctx context.Context, token string, request RegistrationRequest) error {
	digest, err := verificationDigest(token)
	if err != nil {
		return err
	}
	// Email is authoritative in the locked intent, never supplied by this caller.
	request.Email = "owner@example.test"
	if err := validateRegistration(request); err != nil {
		return err
	}
	store, ok := s.store.(OwnershipStore)
	if !ok {
		return ErrVerificationUnavailable
	}
	hash, err := hashPassword(request.Password)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return store.CompleteRegistration(ctx, digest, RegistrationRecord{PasswordHash: hash, Language: request.Language, Theme: request.Theme, Timezone: request.Timezone})
}

// DeliverOne issues a token only in memory, persisting its digest before sending.
// At-least-once delivery may issue another bounded token after an ambiguous SMTP result.
func (s *Service) DeliverOne(ctx context.Context, mailer VerificationMailer) (bool, error) {
	if mailer == nil {
		return false, ErrVerificationUnavailable
	}
	store, ok := s.store.(OwnershipStore)
	if !ok {
		return false, ErrVerificationUnavailable
	}
	raw, digest, err := newVerificationToken()
	if err != nil {
		return false, err
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	delivery, found, err := store.LeaseVerificationDelivery(ctx, digest)
	if err != nil || !found {
		return found, err
	}
	sent := mailer.SendVerification(ctx, delivery.Email, raw) == nil
	// Never expose transport errors that may contain message secrets.
	finishCtx, finishCancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer finishCancel()
	if err := store.FinishVerificationDelivery(finishCtx, delivery, sent); err != nil {
		return true, ErrVerificationUnavailable
	}
	if !sent {
		return true, ErrVerificationUnavailable
	}
	return true, nil
}
