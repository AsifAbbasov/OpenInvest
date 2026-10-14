package postgres_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"github.com/google/uuid"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"os"
	"testing"
)

func renewalDigest(s string) string { h := sha256.Sum256([]byte(s)); return hex.EncodeToString(h[:]) }
func TestEmailOwnershipFinalRenewalWhilePriorLeaseUnfinished(t *testing.T) {

	for _, mode := range []string{"delivered_old_generation_control", "unfinished_old_generation"} {
		t.Run(mode, func(t *testing.T) {
			s, db := ownershipService(t)
			ctx := context.Background()
			email := "renew-" + uuid.NewString() + "@example.test"
			if err := s.BeginEmailRegistration(ctx, email); err != nil {
				t.Fatal(err)
			}
			if _, err := db.Exec("UPDATE identity.registration_intents SET created_at=statement_timestamp()-interval '100 years' WHERE email_normalized=$1", email); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if _, err := db.Exec("DELETE FROM identity.registration_intents WHERE email_normalized=$1", email); err != nil {
					t.Error(err)
				}
			})
			worker, err := postgres.Open(os.Getenv("OPENINVEST_DATABASE_TEST_URL"))
			if err != nil {
				t.Fatal(err)
			}
			defer worker.Close()
			first, found, err := worker.LeaseVerificationDelivery(ctx, renewalDigest(email+"-gen1"))
			if err != nil || !found || first.Email != email {
				t.Fatalf("generation 1 claim unavailable: %t %v", found, err)
			}
			if mode == "delivered_old_generation_control" {
				if err := worker.FinishVerificationDelivery(ctx, first, true); err != nil {
					t.Fatal(err)
				}
			}
			// Simulate the realistic abandoned-worker condition: both the 24h intent
			// and the one-minute lease are expired before renewal is requested.
			if _, err := db.Exec("UPDATE identity.email_delivery_jobs SET lease_until=statement_timestamp()-interval '2 seconds' WHERE id=$1", first.JobID); err != nil {
				t.Fatal(err)
			}
			if _, err := db.Exec("UPDATE identity.registration_intents SET expires_at=statement_timestamp()-interval '2 seconds' WHERE email_normalized=$1", email); err != nil {
				t.Fatal(err)
			}
			if err := s.BeginEmailRegistration(ctx, email); err != nil {
				t.Fatal(err)
			}
			var generation, jobGeneration int64
			if err := db.QueryRow(`SELECT i.generation,j.generation FROM identity.registration_intents i JOIN identity.email_delivery_jobs j ON j.intent_id=i.id WHERE i.email_normalized=$1`, email).Scan(&generation, &jobGeneration); err != nil {
				t.Fatal(err)
			}
			if generation != 2 {
				t.Fatalf("renewal did not advance generation: %d", generation)
			}
			// A renewed eligible account must be able to request new instructions,
			// even if the previous generation crashed while leased.
			if jobGeneration != 2 {
				t.Fatalf("renewed intent stranded behind old %s lease: intent_gen=%d job_gen=%d", mode, generation, jobGeneration)
			}
			next, found, err := worker.LeaseVerificationDelivery(ctx, renewalDigest(email+"-gen2"))
			if err != nil || !found || next.Email != email {
				t.Fatalf("renewed generation delivery blocked after %s: found=%t email=%s err=%v", mode, found, next.Email, err)
			}
			if err := worker.FinishVerificationDelivery(ctx, first, true); err != nil {
				t.Fatal(err)
			}
			var lease string
			if err := db.QueryRow("SELECT lease_id FROM identity.email_delivery_jobs WHERE id=$1", next.JobID).Scan(&lease); err != nil || lease != next.LeaseID {
				t.Fatalf("old generation completion stole successor lease: %s %v", lease, err)
			}
		})
	}
}
