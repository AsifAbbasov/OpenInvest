package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	store  Store
	clock  Clock
	config Config
}

type systemClock struct{}

func (systemClock) Now() time.Time {
	return time.Now().UTC()
}

func NewService(store Store, clock Clock, config Config) (*Service, error) {
	if clock == nil {
		clock = systemClock{}
	}
	if config.AccessTokenTTL == 0 {
		config.AccessTokenTTL = 15 * time.Minute
	}
	if config.RefreshTokenTTL == 0 {
		config.RefreshTokenTTL = 30 * 24 * time.Hour
	}
	if len(config.AccessTokenSecret) < 32 {
		if !config.AllowEphemeralAccessTokenSecret {
			return nil, fmt.Errorf("%w: access token secret must be at least 32 bytes", ErrInvalidInput)
		}
		secret := make([]byte, 32)
		if _, err := rand.Read(secret); err != nil {
			return nil, err
		}
		config.AccessTokenSecret = secret
	}
	return &Service{store: store, clock: clock, config: config}, nil
}

func (s *Service) Register(ctx context.Context, request RegistrationRequest) (AuthResult, error) {
	request.Email = normalizeEmail(request.Email)
	if err := validateRegistration(request); err != nil {
		return AuthResult{}, err
	}
	passwordHash, err := hashPassword(request.Password)
	if err != nil {
		return AuthResult{}, err
	}
	userID := uuid.NewString()
	refreshToken, csrfToken, session, err := s.newSessionRecord(userID)
	if err != nil {
		return AuthResult{}, err
	}
	user, err := s.store.RegisterUser(ctx, RegistrationRecord{
		UserID:              userID,
		InvestmentSubjectID: uuid.NewString(),
		EmailNormalized:     request.Email,
		PasswordHash:        passwordHash,
		Language:            request.Language,
		Theme:               request.Theme,
		Timezone:            request.Timezone,
		Now:                 s.clock.Now(),
		Session:             session,
	})
	if err != nil {
		return AuthResult{}, err
	}
	return s.result(user, refreshToken, csrfToken)
}

func (s *Service) Login(ctx context.Context, request LoginRequest) (AuthResult, error) {
	email := normalizeEmail(request.Email)
	if !validNormalizedEmail(email) || strings.TrimSpace(request.Password) == "" {
		return AuthResult{}, ErrInvalidCredentials
	}
	user, passwordHash, err := s.store.FindUserByEmail(ctx, email)
	if err != nil {
		if capacityErr := verifyPasswordAgainstDummy(request.Password); capacityErr != nil {
			return AuthResult{}, capacityErr
		}
		return AuthResult{}, ErrInvalidCredentials
	}
	verified, err := verifyPassword(request.Password, passwordHash)
	if err != nil {
		return AuthResult{}, err
	}
	if !verified {
		return AuthResult{}, ErrInvalidCredentials
	}
	refreshToken, csrfToken, session, err := s.newSessionRecord(user.ID)
	if err != nil {
		return AuthResult{}, err
	}
	if err := s.store.CreateSession(ctx, session); err != nil {
		return AuthResult{}, err
	}
	return s.result(user, refreshToken, csrfToken)
}

func (s *Service) Refresh(ctx context.Context, refreshToken string, csrfToken string) (AuthResult, error) {
	if strings.TrimSpace(refreshToken) == "" {
		if err := s.recordRejectedAuthEvent(ctx, "AUTH_REFRESH_REJECTED"); err != nil {
			return AuthResult{}, err
		}
		return AuthResult{}, ErrInvalidSession
	}
	if strings.TrimSpace(csrfToken) == "" {
		if err := s.recordRejectedAuthEvent(ctx, "AUTH_REFRESH_REJECTED"); err != nil {
			return AuthResult{}, err
		}
		return AuthResult{}, ErrInvalidCSRF
	}
	nextRefreshToken, nextCSRFToken, nextSession, err := s.newSessionRecord("")
	if err != nil {
		return AuthResult{}, err
	}
	user, err := s.store.RotateSession(ctx, tokenHash(refreshToken), tokenHash(csrfToken), nextSession, s.clock.Now())
	if err != nil {
		return AuthResult{}, ErrInvalidSession
	}
	return s.result(user, nextRefreshToken, nextCSRFToken)
}

func (s *Service) Logout(ctx context.Context, refreshToken string, csrfToken string, allSessions bool) (bool, error) {
	if strings.TrimSpace(refreshToken) == "" {
		if err := s.recordRejectedAuthEvent(ctx, "AUTH_LOGOUT_REJECTED"); err != nil {
			return false, err
		}
		return false, ErrInvalidSession
	}
	if strings.TrimSpace(csrfToken) == "" {
		if err := s.recordRejectedAuthEvent(ctx, "AUTH_LOGOUT_REJECTED"); err != nil {
			return false, err
		}
		return false, ErrInvalidCSRF
	}
	revoked, err := s.store.RevokeSession(ctx, tokenHash(refreshToken), tokenHash(csrfToken), allSessions, s.clock.Now())
	if err != nil {
		return false, ErrInvalidSession
	}
	return revoked, nil
}

func (s *Service) AuthenticateAccessToken(token string) (string, error) {
	claims, err := verifyAccessToken(s.config.AccessTokenSecret, strings.TrimSpace(token), s.clock.Now())
	if err != nil {
		return "", ErrInvalidSession
	}
	return claims.InvestmentSubjectID, nil
}

func (s *Service) AllowsDevelopmentBypass() bool {
	return s.config.AllowDevelopmentBypass
}

func (s *Service) RefreshCookieSecure() bool {
	return s.config.RefreshCookieSecure
}

func (s *Service) RefreshCookieMaxAgeSeconds() int {
	return int(s.config.RefreshTokenTTL / time.Second)
}

func (s *Service) result(user StoredUser, refreshToken string, csrfToken string) (AuthResult, error) {
	now := s.clock.Now()
	expiresAt := now.Add(s.config.AccessTokenTTL)
	accessToken, err := signAccessToken(s.config.AccessTokenSecret, user, now, expiresAt)
	if err != nil {
		return AuthResult{}, err
	}
	return AuthResult{
		User:         user,
		RefreshToken: refreshToken,
		Session: ClientSession{
			AccessToken:          accessToken,
			AccessTokenExpiresAt: expiresAt,
			CSRFToken:            csrfToken,
		},
	}, nil
}

func (s *Service) newSessionRecord(userID string) (string, string, SessionRecord, error) {
	refreshToken, err := randomToken(32)
	if err != nil {
		return "", "", SessionRecord{}, err
	}
	csrfToken, err := randomToken(32)
	if err != nil {
		return "", "", SessionRecord{}, err
	}
	now := s.clock.Now()
	return refreshToken, csrfToken, SessionRecord{
		SessionID:        uuid.NewString(),
		UserID:           userID,
		RefreshTokenHash: tokenHash(refreshToken),
		CSRFTokenHash:    tokenHash(csrfToken),
		ExpiresAt:        now.Add(s.config.RefreshTokenTTL),
		Now:              now,
	}, nil
}

func (s *Service) recordRejectedAuthEvent(ctx context.Context, actionCode string) error {
	return s.store.RecordAuthEvent(ctx, AuthAuditRecord{
		ActionCode: actionCode,
		TargetKind: "session",
		Outcome:    "failure",
		Now:        s.clock.Now(),
	})
}

func validateRegistration(request RegistrationRequest) error {
	if !validNormalizedEmail(request.Email) {
		return fmt.Errorf("%w: email is invalid", ErrInvalidInput)
	}
	if len(request.Password) < 12 || len(request.Password) > 256 {
		return fmt.Errorf("%w: password must be 12..256 characters", ErrInvalidInput)
	}
	if request.Language != LanguageRU && request.Language != LanguageEN {
		return fmt.Errorf("%w: language must be ru or en", ErrInvalidInput)
	}
	if request.Theme != ThemeLight && request.Theme != ThemeDark && request.Theme != ThemeSystem {
		return fmt.Errorf("%w: theme is invalid", ErrInvalidInput)
	}
	if strings.TrimSpace(request.Timezone) == "" || len(request.Timezone) > 64 {
		return fmt.Errorf("%w: timezone is required", ErrInvalidInput)
	}
	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validNormalizedEmail(email string) bool {
	if email == "" || len(email) > 254 {
		return false
	}
	parsed, err := mail.ParseAddress(email)
	return err == nil && parsed.Name == "" && parsed.Address == email
}

func IsInvalidInput(err error) bool {
	return errors.Is(err, ErrInvalidInput)
}
