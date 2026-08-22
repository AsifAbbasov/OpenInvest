package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

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
	if err := insertSession(ctx, tx, record.Session, record.Session.SessionID); err != nil {
		return auth.StoredUser{}, err
	}
	if err := recordAuthAudit(ctx, tx, record.UserID, "AUTH_REGISTER", "user", record.UserID, "success", record.Now); err != nil {
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
	if err := insertSession(ctx, tx, record, record.SessionID); err != nil {
		return err
	}
	if err := recordAuthAudit(ctx, tx, record.UserID, "AUTH_LOGIN", "session", record.SessionID, "success", record.Now); err != nil {
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

	ownerID, err := lookupRefreshSessionOwner(ctx, tx, currentRefreshTokenHash, currentCSRFTokenHash)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidSession) {
			if auditErr := recordAuthAudit(ctx, tx, "", "AUTH_REFRESH_REJECTED", "session", "", "failure", now); auditErr != nil {
				return auth.StoredUser{}, auditErr
			}
			if commitErr := tx.Commit(); commitErr != nil {
				return auth.StoredUser{}, commitErr
			}
		}
		return auth.StoredUser{}, err
	}
	if err := lockUserRefreshes(ctx, tx, ownerID); err != nil {
		return auth.StoredUser{}, err
	}

	current, err := lockRefreshSession(ctx, tx, currentRefreshTokenHash, currentCSRFTokenHash)
	if err != nil {
		return auth.StoredUser{}, err
	}
	if current.UserID != ownerID {
		return auth.StoredUser{}, auth.ErrInvalidSession
	}

	if current.SessionState != "active" {
		if current.SessionFamilyID.Valid {
			if err := revokeActiveSessionFamily(ctx, tx, current.SessionFamilyID.String, now); err != nil {
				return auth.StoredUser{}, err
			}
		} else {
			if err := revokeActiveUserSessions(ctx, tx, current.UserID, now); err != nil {
				return auth.StoredUser{}, err
			}
		}
		if err := recordAuthAudit(ctx, tx, current.UserID, "AUTH_REFRESH_REPLAY", "session", current.SessionID, "failure", now); err != nil {
			return auth.StoredUser{}, err
		}
		if err := tx.Commit(); err != nil {
			return auth.StoredUser{}, err
		}
		return auth.StoredUser{}, auth.ErrInvalidSession
	}

	if !current.ExpiresAt.After(now) {
		if err := recordAuthAudit(ctx, tx, current.UserID, "AUTH_REFRESH_REJECTED", "session", current.SessionID, "failure", now); err != nil {
			return auth.StoredUser{}, err
		}
		if err := tx.Commit(); err != nil {
			return auth.StoredUser{}, err
		}
		return auth.StoredUser{}, auth.ErrInvalidSession
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE identity.sessions
		SET session_state = 'revoked', revoked_at = $2
		WHERE id = $1 AND session_state = 'active'
	`, current.SessionID, now); err != nil {
		return auth.StoredUser{}, err
	}
	next.UserID = current.UserID
	familyID := current.SessionID
	if current.SessionFamilyID.Valid {
		familyID = current.SessionFamilyID.String
	}
	if err := insertSession(ctx, tx, next, familyID); err != nil {
		return auth.StoredUser{}, err
	}
	if err := recordAuthAudit(ctx, tx, current.UserID, "AUTH_REFRESH", "session", next.SessionID, "success", now); err != nil {
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

	ownerID, err := lookupRefreshSessionOwner(ctx, tx, refreshTokenHash, csrfTokenHash)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidSession) {
			if auditErr := recordAuthAudit(ctx, tx, "", "AUTH_LOGOUT_REJECTED", "session", "", "failure", now); auditErr != nil {
				return false, auditErr
			}
			if commitErr := tx.Commit(); commitErr != nil {
				return false, commitErr
			}
		}
		return false, err
	}
	if err := lockUserRefreshes(ctx, tx, ownerID); err != nil {
		return false, err
	}
	current, err := lockRefreshSession(ctx, tx, refreshTokenHash, csrfTokenHash)
	if err != nil {
		return false, err
	}
	if current.UserID != ownerID {
		return false, auth.ErrInvalidSession
	}

	if allSessions {
		if err := revokeActiveUserSessions(ctx, tx, current.UserID, now); err != nil {
			return false, err
		}
	} else if current.SessionState == "active" {
		if _, err := tx.ExecContext(ctx, `
			UPDATE identity.sessions
			SET session_state = 'revoked', revoked_at = $2
			WHERE id = $1 AND session_state = 'active'
		`, current.SessionID, now); err != nil {
			return false, err
		}
	} else if current.SessionFamilyID.Valid {
		if err := revokeActiveSessionFamily(ctx, tx, current.SessionFamilyID.String, now); err != nil {
			return false, err
		}
	} else {
		if err := revokeActiveUserSessions(ctx, tx, current.UserID, now); err != nil {
			return false, err
		}
	}

	if err := recordAuthAudit(ctx, tx, current.UserID, "AUTH_LOGOUT", "session", current.SessionID, "success", now); err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Store) RecordAuthEvent(ctx context.Context, record auth.AuthAuditRecord) error {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer rollback(tx)
	if err := recordAuthAudit(ctx, tx, record.ActorID, record.ActionCode, record.TargetKind, record.TargetID, record.Outcome, record.Now); err != nil {
		return err
	}
	return tx.Commit()
}

type storedSession struct {
	SessionID       string
	UserID          string
	SessionFamilyID sql.NullString
	SessionState    string
	ExpiresAt       time.Time
}

func insertSession(ctx context.Context, tx *sql.Tx, record auth.SessionRecord, sessionFamilyID string) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO identity.sessions (
			id, user_id, session_family_id, refresh_token_hash, csrf_token_hash, session_state,
			expires_at, created_at, last_rotated_at
		)
		VALUES ($1, $2, $3, $4, $5, 'active', $6, $7, $7)
	`, record.SessionID, record.UserID, sessionFamilyID, record.RefreshTokenHash, record.CSRFTokenHash, record.ExpiresAt, record.Now)
	return err
}

func revokeActiveSessionFamily(ctx context.Context, tx *sql.Tx, sessionFamilyID string, now time.Time) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE identity.sessions
		SET session_state = 'revoked', revoked_at = $2
		WHERE session_family_id = $1
			AND session_state = 'active'
	`, sessionFamilyID, now)
	return err
}

func revokeActiveUserSessions(ctx context.Context, tx *sql.Tx, userID string, now time.Time) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE identity.sessions
		SET session_state = 'revoked', revoked_at = $2
		WHERE user_id = $1
			AND session_state = 'active'
	`, userID, now)
	return err
}

func recordAuthAudit(ctx context.Context, tx *sql.Tx, actorID string, actionCode string, targetKind string, targetID string, outcome string, now time.Time) error {
	if strings.TrimSpace(actorID) != "" {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO audit.actors (id, actor_kind, created_at)
			VALUES ($1, 'user', $2)
			ON CONFLICT (id) DO NOTHING
		`, actorID, now); err != nil {
			return err
		}
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO audit.events (
			id, actor_id, action_code, target_kind, target_id, outcome,
			request_id, trace_id, occurred_at, schema_version
		)
		VALUES (
			$1, NULLIF($2, '')::uuid, $3, $4, NULLIF($5, '')::uuid, $6,
			NULL, NULL, $7, 1
		)
	`, uuid.NewString(), actorID, actionCode, targetKind, targetID, outcome, now)
	return err
}

func lookupRefreshSessionOwner(ctx context.Context, tx *sql.Tx, refreshTokenHash string, csrfTokenHash string) (string, error) {
	var userID string
	err := tx.QueryRowContext(ctx, `
		SELECT user_id
		FROM identity.sessions
		WHERE refresh_token_hash = $1
			AND csrf_token_hash = $2
	`, refreshTokenHash, csrfTokenHash).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", auth.ErrInvalidSession
	}
	return userID, err
}

func lockUserRefreshes(ctx context.Context, tx *sql.Tx, userID string) error {
	_, err := tx.ExecContext(ctx, `
		SELECT pg_advisory_xact_lock(hashtextextended($1, 0))
	`, "openinvest/auth-refresh/"+userID)
	return err
}

func lockRefreshSession(ctx context.Context, tx *sql.Tx, refreshTokenHash string, csrfTokenHash string) (storedSession, error) {
	var session storedSession
	err := tx.QueryRowContext(ctx, `
		SELECT id, user_id, session_family_id, session_state, expires_at
		FROM identity.sessions
		WHERE refresh_token_hash = $1
			AND csrf_token_hash = $2
		FOR UPDATE
	`, refreshTokenHash, csrfTokenHash).Scan(
		&session.SessionID,
		&session.UserID,
		&session.SessionFamilyID,
		&session.SessionState,
		&session.ExpiresAt,
	)
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
