package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/openinvest/openinvest/backend-go/internal/auth"
)

func ownershipTx(ctx context.Context, db *sql.DB) (*sql.Tx, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, `SET LOCAL lock_timeout = '2s'; SET LOCAL statement_timeout = '4s'`); err != nil {
		tx.Rollback()
		return nil, err
	}
	return tx, nil
}
func (s *Store) BeginRegistration(ctx context.Context, email string) error {
	tx, err := ownershipTx(ctx, s.db)
	if err != nil {
		return err
	}
	defer rollback(tx)
	// Shared capacity and deterministic intake ordering, independent of membership.
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(150015)`); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM identity.registration_intents WHERE id IN (SELECT id FROM identity.registration_intents WHERE expires_at<statement_timestamp()-interval '7 days' LIMIT 100)`); err != nil {
		return err
	}
	var count int
	if err = tx.QueryRowContext(ctx, `SELECT count(*) FROM identity.registration_intents WHERE intent_state='pending' AND expires_at>statement_timestamp()`).Scan(&count); err != nil {
		return err
	}
	var id string
	var generation int64
	err = tx.QueryRowContext(ctx, `SELECT id,generation FROM identity.registration_intents WHERE email_normalized=$1 FOR UPDATE`, email).Scan(&id, &generation)
	if errors.Is(err, sql.ErrNoRows) {
		if count >= 10000 {
			return tx.Commit()
		}
		id = uuid.NewString()
		generation = 1
		_, err = tx.ExecContext(ctx, `INSERT INTO identity.registration_intents(id,email_normalized,created_at,expires_at,next_delivery_at) VALUES($1,$2,statement_timestamp(),statement_timestamp()+interval '24 hours',statement_timestamp())`, id, email)
	} else if err == nil {
		err = tx.QueryRowContext(ctx, `UPDATE identity.registration_intents SET generation=generation+1,intent_state='pending',created_at=statement_timestamp(),expires_at=statement_timestamp()+interval '24 hours',next_delivery_at=statement_timestamp(),consumed_at=NULL WHERE id=$1 AND expires_at<=statement_timestamp() AND intent_state<>'consumed' RETURNING generation`, id).Scan(&generation)
		if errors.Is(err, sql.ErrNoRows) {
			err = tx.QueryRowContext(ctx, `SELECT generation FROM identity.registration_intents WHERE id=$1`, id).Scan(&generation)
		}
	}
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO identity.email_delivery_jobs(id,intent_id,generation,job_state,available_at,expires_at) SELECT $1,id,generation,'queued',statement_timestamp(),expires_at FROM identity.registration_intents WHERE id=$2 AND intent_state='pending' AND expires_at>statement_timestamp() AND next_delivery_at<=statement_timestamp() ON CONFLICT(intent_id) DO UPDATE SET generation=EXCLUDED.generation,job_state='queued',attempts=CASE WHEN identity.email_delivery_jobs.generation<>EXCLUDED.generation THEN 0 ELSE identity.email_delivery_jobs.attempts END,available_at=EXCLUDED.available_at,expires_at=EXCLUDED.expires_at,lease_id=NULL,lease_until=NULL WHERE identity.email_delivery_jobs.generation<>EXCLUDED.generation OR (identity.email_delivery_jobs.job_state IN('delivered','failed','cancelled') AND identity.email_delivery_jobs.attempts<4)`, uuid.NewString(), id)
	if err != nil {
		return err
	}
	return tx.Commit()
}
func (s *Store) LeaseVerificationDelivery(ctx context.Context, digest string) (auth.VerificationDelivery, bool, error) {
	tx, err := ownershipTx(ctx, s.db)
	if err != nil {
		return auth.VerificationDelivery{}, false, err
	}
	defer rollback(tx)
	// All paths lock intent before jobs/tokens. Never hold a database lock over delivery.
	var id, email string
	var gen int64
	err = tx.QueryRowContext(ctx, `SELECT i.id,i.email_normalized,i.generation FROM identity.registration_intents i WHERE EXISTS(SELECT 1 FROM identity.email_delivery_jobs j WHERE j.intent_id=i.id AND j.generation=i.generation AND ((j.job_state='leased' AND j.lease_until<=statement_timestamp() AND j.attempts>=4) OR (i.intent_state='pending' AND i.expires_at>statement_timestamp() AND j.attempts<4 AND j.available_at<=statement_timestamp() AND (j.job_state='queued' OR (j.job_state='leased' AND j.lease_until<=statement_timestamp()))))) ORDER BY i.created_at FOR UPDATE OF i SKIP LOCKED LIMIT 1`).Scan(&id, &email, &gen)
	if errors.Is(err, sql.ErrNoRows) {
		return auth.VerificationDelivery{}, false, nil
	}
	if err != nil {
		return auth.VerificationDelivery{}, false, err
	}
	// Exhausted expired leases are terminalized under the same parent lock used
	// by leasing and activation. This bounded pass cannot issue a fifth token.
	result, err := tx.ExecContext(ctx, `UPDATE identity.email_delivery_jobs SET job_state='failed',lease_id=NULL,lease_until=NULL WHERE intent_id=$1 AND generation=$2 AND job_state='leased' AND attempts>=4 AND lease_until<=statement_timestamp()`, id, gen)
	if err != nil {
		return auth.VerificationDelivery{}, false, err
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return auth.VerificationDelivery{}, false, err
	}
	if changed != 0 {
		return auth.VerificationDelivery{}, false, tx.Commit()
	}
	lease := uuid.NewString()
	var job string
	var expiry time.Time
	err = tx.QueryRowContext(ctx, `UPDATE identity.email_delivery_jobs SET job_state='leased',lease_id=$2,lease_until=statement_timestamp()+interval '1 minute',attempts=attempts+1 WHERE intent_id=$1 AND generation=$3 AND attempts<4 AND available_at<=statement_timestamp() AND (job_state='queued' OR (job_state='leased' AND lease_until<=statement_timestamp())) RETURNING id`, id, lease, gen).Scan(&job)
	if errors.Is(err, sql.ErrNoRows) {
		return auth.VerificationDelivery{}, false, tx.Commit()
	}
	if err != nil {
		return auth.VerificationDelivery{}, false, err
	}
	err = tx.QueryRowContext(ctx, `INSERT INTO identity.email_verification_tokens(id,intent_id,generation,purpose,token_hash,created_at,expires_at) VALUES($1,$2,$3,'new_registration',$4,statement_timestamp(),LEAST(statement_timestamp()+interval '30 minutes',(SELECT expires_at FROM identity.registration_intents WHERE id=$2))) RETURNING expires_at`, uuid.NewString(), id, gen, digest).Scan(&expiry)
	if err != nil {
		return auth.VerificationDelivery{}, false, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE identity.registration_intents SET next_delivery_at=statement_timestamp()+interval '1 minute' WHERE id=$1`, id); err != nil {
		return auth.VerificationDelivery{}, false, err
	}
	if err = tx.Commit(); err != nil {
		return auth.VerificationDelivery{}, false, err
	}
	return auth.VerificationDelivery{JobID: job, LeaseID: lease, Email: email, ExpiresAt: expiry}, true, nil
}
func (s *Store) FinishVerificationDelivery(ctx context.Context, d auth.VerificationDelivery, sent bool) error {
	state := "queued"
	if sent {
		state = "delivered"
	}
	tx, err := ownershipTx(ctx, s.db)
	if err != nil {
		return err
	}
	defer rollback(tx)
	var intentID string
	err = tx.QueryRowContext(ctx, `SELECT i.id FROM identity.registration_intents i JOIN identity.email_delivery_jobs j ON j.intent_id=i.id WHERE j.id=$1 FOR UPDATE OF i`, d.JobID).Scan(&intentID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE identity.email_delivery_jobs SET job_state=CASE WHEN attempts>=4 AND $3='queued' THEN 'failed' ELSE $3 END,lease_id=NULL,lease_until=NULL,available_at=statement_timestamp()+interval '5 minutes' WHERE id=$1 AND lease_id=$2 AND job_state='leased' AND lease_until>statement_timestamp()`, d.JobID, d.LeaseID, state)
	if err != nil {
		return err
	}
	return tx.Commit()
}
func (s *Store) CompleteRegistration(ctx context.Context, digest string, record auth.RegistrationRecord) error {
	tx, err := ownershipTx(ctx, s.db)
	if err != nil {
		return err
	}
	defer rollback(tx)
	var id, email string
	var generation int64
	err = tx.QueryRowContext(ctx, `SELECT i.id,i.email_normalized,i.generation FROM identity.registration_intents i JOIN identity.email_verification_tokens t ON t.intent_id=i.id WHERE t.token_hash=$1 FOR UPDATE OF i`, digest).Scan(&id, &email, &generation)
	if errors.Is(err, sql.ErrNoRows) {
		return auth.ErrInvalidVerification
	}
	if err != nil {
		return err
	}
	var tokenID string
	err = tx.QueryRowContext(ctx, `SELECT t.id FROM identity.email_verification_tokens t JOIN identity.registration_intents i ON i.id=t.intent_id WHERE t.token_hash=$1 AND t.purpose='new_registration' AND t.generation=$2 AND t.consumed_at IS NULL AND t.expires_at>statement_timestamp() AND i.intent_state='pending' AND i.expires_at>statement_timestamp() FOR UPDATE OF t`, digest, generation).Scan(&tokenID)
	if errors.Is(err, sql.ErrNoRows) {
		return auth.ErrInvalidVerification
	}
	if err != nil {
		return err
	}
	var exists bool
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM identity.users WHERE email_normalized=$1 AND account_state<>'deleted')`, email).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return auth.ErrInvalidVerification
	}
	record.UserID = uuid.NewString()
	record.InvestmentSubjectID = uuid.NewString()
	record.EmailNormalized = email
	if err = tx.QueryRowContext(ctx, `SELECT statement_timestamp()`).Scan(&record.Now); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO identity.users (
			id, email_normalized, account_state, language, theme, timezone, created_at, updated_at
		)
		VALUES ($1, $2, 'active', $3, $4, $5, $6, $6)
	`, record.UserID, record.EmailNormalized, record.Language, record.Theme, record.Timezone, record.Now); err != nil {
		if isUniqueViolation(err) {
			return auth.ErrInvalidVerification
		}
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO investment.subjects (id, subject_state, created_at, updated_at)
		VALUES ($1, 'active', $2, $2)
	`, record.InvestmentSubjectID, record.Now); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO identity.user_investment_links (user_id, investment_subject_id, created_at)
		VALUES ($1, $2, $3)
	`, record.UserID, record.InvestmentSubjectID, record.Now); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO identity.credentials (user_id, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $3)
	`, record.UserID, record.PasswordHash, record.Now); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO identity.privacy_settings (
			user_id, privacy_mode, tax_profile_enabled, notifications_enabled, analytics_mode,
			created_at, updated_at
		)
		VALUES ($1, true, false, false, 'anonymous', $2, $2)
	`, record.UserID, record.Now); err != nil {
		return err
	}

	if _, err = tx.ExecContext(ctx, `INSERT INTO identity.user_email_verifications(user_id,evidence_state,verified_at) VALUES($1,'verified',$2)`, record.UserID, record.Now); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE identity.registration_intents SET intent_state='consumed',consumed_at=$2 WHERE id=$1`, id, record.Now); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE identity.email_verification_tokens SET consumed_at=$2 WHERE intent_id=$1 AND consumed_at IS NULL`, id, record.Now); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE identity.email_delivery_jobs SET job_state='cancelled',lease_id=NULL,lease_until=NULL WHERE intent_id=$1`, id); err != nil {
		return err
	}
	if err = recordAuthAudit(ctx, tx, record.UserID, "AUTH_EMAIL_VERIFIED", "user", record.UserID, "success", record.Now); err != nil {
		return err
	}
	return tx.Commit()
}
