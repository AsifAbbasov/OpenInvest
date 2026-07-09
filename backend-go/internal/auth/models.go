package auth

import (
	"context"
	"errors"
	"time"
)

const (
	RefreshCookieName = "oi_refresh"
	LanguageRU        = "ru"
	LanguageEN        = "en"
	ThemeLight        = "light"
	ThemeDark         = "dark"
	ThemeSystem       = "system"
)

var (
	ErrInvalidInput       = errors.New("invalid auth input")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidSession     = errors.New("invalid session")
	ErrInvalidCSRF        = errors.New("invalid csrf token")
)

type Clock interface {
	Now() time.Time
}

type Store interface {
	RegisterUser(ctx context.Context, record RegistrationRecord) (StoredUser, error)
	FindUserByEmail(ctx context.Context, normalizedEmail string) (StoredUser, string, error)
	CreateSession(ctx context.Context, record SessionRecord) error
	RotateSession(ctx context.Context, currentRefreshTokenHash string, currentCSRFTokenHash string, next SessionRecord, now time.Time) (StoredUser, error)
	RevokeSession(ctx context.Context, refreshTokenHash string, csrfTokenHash string, allSessions bool, now time.Time) (bool, error)
}

type RegistrationRequest struct {
	Email    string
	Password string
	Language string
	Theme    string
	Timezone string
}

type LoginRequest struct {
	Email    string
	Password string
}

type RegistrationRecord struct {
	UserID              string
	InvestmentSubjectID string
	EmailNormalized     string
	PasswordHash        string
	Language            string
	Theme               string
	Timezone            string
	Now                 time.Time
	Session             SessionRecord
}

type SessionRecord struct {
	SessionID        string
	UserID           string
	RefreshTokenHash string
	CSRFTokenHash    string
	ExpiresAt        time.Time
	Now              time.Time
}

type StoredUser struct {
	ID                  string
	InvestmentSubjectID string
	Email               string
	Language            string
	Theme               string
	Timezone            string
	PrivacyMode         bool
	CreatedAt           time.Time
}

type AuthResult struct {
	User         StoredUser
	Session      ClientSession
	RefreshToken string
}

type ClientSession struct {
	AccessToken          string
	AccessTokenExpiresAt time.Time
	CSRFToken            string
}

type Config struct {
	AccessTokenSecret               []byte
	AccessTokenTTL                  time.Duration
	RefreshTokenTTL                 time.Duration
	RefreshCookieSecure             bool
	AllowDevelopmentBypass          bool
	AllowEphemeralAccessTokenSecret bool
}
