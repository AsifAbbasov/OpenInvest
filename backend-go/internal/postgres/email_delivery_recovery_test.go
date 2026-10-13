package postgres_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/google/uuid"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"os"
	"testing"
)

func TestEmailOwnershipDeliveryRetryAndStaleCompletion(t *testing.T) {
	service, db := ownershipService(t)
	ctx := context.Background()
	worker, err := postgres.Open(os.Getenv("OPENINVEST_DATABASE_TEST_URL"))
	if err != nil {
		t.Fatal(err)
	}
	defer worker.Close()
	for _, mode := range []string{"failed_delivery", "expired_lease"} {
		t.Run(mode, func(t *testing.T) {
			email := "retry-" + uuid.NewString() + "@example.test"
			if err := service.BeginEmailRegistration(ctx, email); err != nil {
				t.Fatal(err)
			}
			if _, err := db.Exec(`UPDATE identity.registration_intents SET created_at=statement_timestamp()-interval '100 years' WHERE email_normalized=$1`, email); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if _, err := db.Exec(`DELETE FROM identity.registration_intents WHERE email_normalized=$1`, email); err != nil {
					t.Error(err)
				}
			})
			var previousJob, previousLease string
			for attempt := 1; attempt <= 4; attempt++ {
				sum := sha256.Sum256([]byte(fmt.Sprintf("%s-%d", email, attempt)))
				current, found, err := worker.LeaseVerificationDelivery(ctx, hex.EncodeToString(sum[:]))
				if err != nil || !found || current.Email != email {
					t.Fatalf("lease %d %t %v", attempt, found, err)
				}
				if attempt > 1 {
					stale := current
					stale.JobID = previousJob
					stale.LeaseID = previousLease
					if err := worker.FinishVerificationDelivery(ctx, stale, true); err != nil {
						t.Fatal(err)
					}
					var leaseID string
					if err := db.QueryRow(`SELECT lease_id FROM identity.email_delivery_jobs WHERE id=$1`, current.JobID).Scan(&leaseID); err != nil || leaseID != current.LeaseID {
						t.Fatalf("stale completion changed active lease: %v", err)
					}
				}
				previousJob, previousLease = current.JobID, current.LeaseID
				if mode == "failed_delivery" {
					if err := worker.FinishVerificationDelivery(ctx, current, false); err != nil {
						t.Fatal(err)
					}
				}
				if _, err := db.Exec(`UPDATE identity.email_delivery_jobs SET available_at=statement_timestamp()-interval '1 second',lease_until=statement_timestamp()-interval '1 second' WHERE id=$1`, current.JobID); err != nil {
					t.Fatal(err)
				}
				if mode == "expired_lease" {
					if err := worker.FinishVerificationDelivery(ctx, current, true); err != nil {
						t.Fatal(err)
					}
				}
			}
			sum := sha256.Sum256([]byte(email + "-exhausted"))
			if lease, found, err := worker.LeaseVerificationDelivery(ctx, hex.EncodeToString(sum[:])); err != nil || found && lease.Email == email {
				t.Fatalf("exhausted lease %t %v", found, err)
			}
			var state string
			var attempts, tokens, users int
			if err := db.QueryRow(`SELECT j.job_state,j.attempts FROM identity.email_delivery_jobs j JOIN identity.registration_intents i ON i.id=j.intent_id WHERE i.email_normalized=$1`, email).Scan(&state, &attempts); err != nil || state != "failed" || attempts != 4 {
				t.Fatalf("state=%s attempts=%d err=%v", state, attempts, err)
			}
			if err := db.QueryRow(`SELECT count(*) FROM identity.email_verification_tokens t JOIN identity.registration_intents i ON i.id=t.intent_id WHERE i.email_normalized=$1 AND length(t.token_hash)=64`, email).Scan(&tokens); err != nil || tokens != 4 {
				t.Fatalf("digest-only tokens=%d err=%v", tokens, err)
			}
			if err := db.QueryRow(`SELECT count(*) FROM identity.users WHERE email_normalized=$1`, email).Scan(&users); err != nil || users != 0 {
				t.Fatalf("unexpected activation=%d err=%v", users, err)
			}
		})
	}
}
