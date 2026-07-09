package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/auth"
)

func (s *Store) RegisterUser(ctx context.Context, record auth.RegistrationRecord) (auth.StoredUser, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return auth.StoredUser{}, err
	}
	defer rollback(tx)

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO identity.users (
			id, email_normalized, account_state, language, theme, timezone, created_at, updated_at
		)
		VALUES ($1, $2, 'active', $3, $4, $5, $6, $6)
	`, record.UserID, record.EmailNormalized, record.Language, record.Theme, record.Timezone, record.Now); err != nil {
		if isUniqueViolation(err) {
			return auth.StoredUser{}, auth.ErrEmailAlreadyExists
		}
		return auth.StoredUser{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO investment.subjects (id, subject_state, created_at, updated_at)
		VALUES ($1, 'active', $2, $2)
	`, record.InvestmentSubjectID, record.Now); err != nil {
		return auth.StoredUser{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO identity.user_investment_links (user_id, investment_subject_id, created_at)
		VALUES ($1, $2, $3)
	`, record.UserID, record.InvestmentSubjectID, record.Now); err != nil {
		return auth.StoredUser{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO identity.credentials (user_id, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $3)
	`, record.UserID, record.PasswordHash, record.Now); err != nil {
		return auth.StoredUser{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO identity.privacy_settings (
			user_id, privacy_mode, tax_profile_enabled, notifications_enabled, analytics_mode,
			created_at, updated_at
		)
		VALUES ($1, true, false, false, 'anonymous', $2, $2)
	`, record.UserID, record.Now); err != nil {
		return auth.StoredUser{}, err
	}
	if err := insertSession(ctx, tx, record.Session); err != nil {
		return auth.StoredUser{}, err
	}

	user, err := scanAuthUser(tx.QueryRowContext(ctx, authUserSelectSQL()+` WHERE u.id = $1`, record.UserID))
	if err != nil {
		return auth.StoredUser{}, err
	}
	return user, tx.Commit()
}

func (s *Store) FindUserByEmail(ctx context.Context, normalizedEmail string) (auth.StoredUser, string, error) {
	var passwordHash string
	user, err := scanAuthUserWithPassword(s.db.QueryRowContext(ctx, authUserWithPasswordSelectSQL()+`
		WHERE u.email_normalized = $1 AND u.account_state = 'active'
	`, normalizedEmail), &passwordHash)
	if err != nil {
		return auth.StoredUser{}, "", err
	}
	return user, passwordHash, nil
}

func (s *Store) CreateSession(ctx context.Context, record auth.SessionRecord) error {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer rollback(tx)
	if err := insertSession(ctx, tx, record); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) RotateSession(ctx context.Context, currentRefreshTokenHash string, currentCSRFTokenHash string, next auth.SessionRecord, now time.Time) (auth.StoredUser, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return auth.StoredUser{}, err
	}
	defer rollback(tx)

	current, err := lockActiveSession(ctx, tx, currentRefreshTokenHash, currentCSRFTokenHash, now)
	if err != nil {
		return auth.StoredUser{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE identity.sessions
		SET session_state = 'revoked', revoked_at = $2
		WHERE id = $1 AND session_state = 'active'
	`, current.SessionID, now); err != nil {
		return auth.StoredUser{}, err
	}
	next.UserID = current.UserID
	if err := insertSession(ctx, tx, next); err != nil {
		return auth.StoredUser{}, err
	}
	user, err := scanAuthUser(tx.QueryRowContext(ctx, authUserSelectSQL()+` WHERE u.id = $1`, current.UserID))
	if err != nil {
		return auth.StoredUser{}, err
	}
	return user, tx.Commit()
}

func (s *Store) RevokeSession(ctx context.Context, refreshTokenHash string, csrfTokenHash string, allSessions bool, now time.Time) (bool, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return false, err
	}
	defer rollback(tx)

	current, err := lockActiveSession(ctx, tx, refreshTokenHash, csrfTokenHash, now)
	if err != nil {
		return false, err
	}
	var result sql.Result
	if allSessions {
		result, err = tx.ExecContext(ctx, `
			UPDATE identity.sessions
			SET session_state = 'revoked', revoked_at = $2
			WHERE user_id = $1 AND session_state = 'active'
		`, current.UserID, now)
	} else {
		result, err = tx.ExecContext(ctx, `
			UPDATE identity.sessions
			SET session_state = 'revoked', revoked_at = $2
			WHERE id = $1 AND session_state = 'active'
		`, current.SessionID, now)
	}
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, tx.Commit()
}

type storedSession struct {
	SessionID string
	UserID    string
}

func insertSession(ctx context.Context, tx *sql.Tx, record auth.SessionRecord) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO identity.sessions (
			id, user_id, refresh_token_hash, csrf_token_hash, session_state,
			expires_at, created_at, last_rotated_at
		)
		VALUES ($1, $2, $3, $4, 'active', $5, $6, $6)
	`, record.SessionID, record.UserID, record.RefreshTokenHash, record.CSRFTokenHash, record.ExpiresAt, record.Now)
	return err
}

func lockActiveSession(ctx context.Context, tx *sql.Tx, refreshTokenHash string, csrfTokenHash string, now time.Time) (storedSession, error) {
	var session storedSession
	err := tx.QueryRowContext(ctx, `
		SELECT id, user_id
		FROM identity.sessions
		WHERE refresh_token_hash = $1
			AND csrf_token_hash = $2
			AND session_state = 'active'
			AND expires_at > $3
		FOR UPDATE
	`, refreshTokenHash, csrfTokenHash, now).Scan(&session.SessionID, &session.UserID)
	if errors.Is(err, sql.ErrNoRows) {
		return storedSession{}, auth.ErrInvalidSession
	}
	return session, err
}

func authUserSelectSQL() string {
	return `
		SELECT
			u.id, l.investment_subject_id, u.email_normalized, u.language, u.theme, u.timezone,
			ps.privacy_mode, u.created_at
		FROM identity.users u
		JOIN identity.user_investment_links l ON l.user_id = u.id
		JOIN identity.privacy_settings ps ON ps.user_id = u.id
	`
}

func authUserWithPasswordSelectSQL() string {
	return `
		SELECT
			u.id, l.investment_subject_id, u.email_normalized, u.language, u.theme, u.timezone,
			ps.privacy_mode, u.created_at, c.password_hash
		FROM identity.users u
		JOIN identity.user_investment_links l ON l.user_id = u.id
		JOIN identity.privacy_settings ps ON ps.user_id = u.id
		JOIN identity.credentials c ON c.user_id = u.id
	`
}

func scanAuthUser(row scanner) (auth.StoredUser, error) {
	var user auth.StoredUser
	err := row.Scan(
		&user.ID,
		&user.InvestmentSubjectID,
		&user.Email,
		&user.Language,
		&user.Theme,
		&user.Timezone,
		&user.PrivacyMode,
		&user.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return auth.StoredUser{}, auth.ErrInvalidCredentials
	}
	return user, err
}

func scanAuthUserWithPassword(row scanner, passwordHash *string) (auth.StoredUser, error) {
	var user auth.StoredUser
	err := row.Scan(
		&user.ID,
		&user.InvestmentSubjectID,
		&user.Email,
		&user.Language,
		&user.Theme,
		&user.Timezone,
		&user.PrivacyMode,
		&user.CreatedAt,
		passwordHash,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return auth.StoredUser{}, auth.ErrInvalidCredentials
	}
	return user, err
}

func isUniqueViolation(err error) bool {
	var state interface{ SQLState() string }
	return errors.As(err, &state) && state.SQLState() == "23505" || strings.Contains(err.Error(), "SQLSTATE 23505")
}
