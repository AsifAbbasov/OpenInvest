package postgres_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/google/uuid"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"os"
	"strings"
	"sync"
	"testing"
)

func renewalDigest(s string) string { h := sha256.Sum256([]byte(s)); return hex.EncodeToString(h[:]) }
func TestEmailOwnershipBuilderRenewalWhilePriorLeaseUnfinished(t *testing.T) {

	for _, mode := range []string{"delivered_old_generation_control", "unfinished_old_generation", "failed_old_generation", "still_valid_old_lease"} {
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
			if mode == "failed_old_generation" {
				if _, err := db.Exec("UPDATE identity.email_delivery_jobs SET job_state='failed',attempts=4,lease_id=NULL,lease_until=NULL WHERE id=$1", first.JobID); err != nil {
					t.Fatal(err)
				}
			}
			// Simulate the realistic abandoned-worker condition: both the 24h intent
			// and the one-minute lease are expired before renewal is requested.
			if mode != "still_valid_old_lease" {
				if _, err := db.Exec("UPDATE identity.email_delivery_jobs SET lease_until=statement_timestamp()-interval '2 seconds' WHERE id=$1", first.JobID); err != nil {
					t.Fatal(err)
				}
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
			for _, sent := range []bool{true, false} {
				if err := worker.FinishVerificationDelivery(ctx, first, sent); err != nil {
					t.Fatal(err)
				}
			}
			var lease string
			if err := db.QueryRow("SELECT lease_id FROM identity.email_delivery_jobs WHERE id=$1", next.JobID).Scan(&lease); err != nil || lease != next.LeaseID {
				t.Fatalf("old generation completion stole successor lease: %s %v", lease, err)
			}
		})
	}
}

func TestEmailOwnershipConcurrentRenewalAndRecovery(t *testing.T) {
	service, db := ownershipService(t)
	ctx := context.Background()
	email := "renew-race-" + uuid.NewString() + "@example.test"
	if err := service.BeginEmailRegistration(ctx, email); err != nil {
		t.Fatal(err)
	}
	worker, err := postgres.Open(os.Getenv("OPENINVEST_DATABASE_TEST_URL"))
	if err != nil {
		t.Fatal(err)
	}
	defer worker.Close()
	if _, err := db.Exec(`UPDATE identity.registration_intents SET created_at=statement_timestamp()-interval '100 years' WHERE email_normalized=$1`, email); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := db.Exec(`DELETE FROM identity.registration_intents WHERE email_normalized=$1`, email); err != nil {
			t.Error(err)
		}
	})
	lease, found, err := worker.LeaseVerificationDelivery(ctx, renewalDigest(email+"-old"))
	if err != nil || !found || lease.Email != email {
		t.Fatalf("initial lease: %t %v", found, err)
	}
	if _, err := db.Exec(`UPDATE identity.email_delivery_jobs SET attempts=4,lease_until=statement_timestamp()-interval '1 second' WHERE id=$1`, lease.JobID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE identity.registration_intents SET expires_at=statement_timestamp()-interval '1 second' WHERE email_normalized=$1`, email); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	start := make(chan struct{})
	for n := 0; n < 4; n++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			<-start
			if err := service.BeginEmailRegistration(ctx, email); err != nil {
				t.Error(err)
			}
		}()
		go func(n int) {
			defer wg.Done()
			<-start
			if _, _, err := worker.LeaseVerificationDelivery(ctx, renewalDigest(fmt.Sprintf("%s-recovery-%d", email, n))); err != nil {
				t.Error(err)
			}
		}(n)
	}
	close(start)
	wg.Wait()
	var generation, jobGeneration, attempts, users, sessions int
	if err := db.QueryRow(`SELECT i.generation,j.generation,j.attempts FROM identity.registration_intents i JOIN identity.email_delivery_jobs j ON j.intent_id=i.id WHERE i.email_normalized=$1`, email).Scan(&generation, &jobGeneration, &attempts); err != nil {
		t.Fatal(err)
	}
	if generation != 2 || jobGeneration != 2 || attempts > 1 {
		t.Fatalf("generation/ownership inconsistent: %d/%d attempts=%d", generation, jobGeneration, attempts)
	}
	for _, sent := range []bool{true, false} {
		if err := worker.FinishVerificationDelivery(ctx, lease, sent); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.QueryRow(`SELECT count(*) FROM identity.users WHERE email_normalized=$1`, email).Scan(&users); err != nil || users != 0 {
		t.Fatalf("unauthorized users=%d %v", users, err)
	}
	if err := db.QueryRow(`SELECT count(*) FROM identity.sessions s JOIN identity.users u ON u.id=s.user_id WHERE u.email_normalized=$1`, email).Scan(&sessions); err != nil || sessions != 0 {
		t.Fatalf("unauthorized sessions=%d %v", sessions, err)
	}
	t.Log("CONCURRENT_RENEWAL_GENERATION=2 SINGLE_DELIVERY_OWNERSHIP=PASS NO_PREPROOF_ACTIVATION=PASS")
}

func TestEmailOwnershipRenewalRollbackOnJobFailure(t *testing.T) {
	service, db := ownershipService(t)
	ctx := context.Background()
	email := "renew-rollback-" + uuid.NewString() + "@example.test"
	if err := service.BeginEmailRegistration(ctx, email); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := db.Exec(`DELETE FROM identity.registration_intents WHERE email_normalized=$1`, email); err != nil {
			t.Error(err)
		}
	})
	if _, err := db.Exec(`UPDATE identity.registration_intents SET expires_at=statement_timestamp()-interval '1 second' WHERE email_normalized=$1`, email); err != nil {
		t.Fatal(err)
	}
	var id string
	if err := db.QueryRow(`SELECT id FROM identity.registration_intents WHERE email_normalized=$1`, email).Scan(&id); err != nil {
		t.Fatal(err)
	}
	// Disposable owner-role fixture only. Trigger affects this unique intent only.
	name := "renewal_fault_" + uuid.NewString()
	name = strings.ReplaceAll(name, "-", "")
	if _, err := db.Exec(fmt.Sprintf(`CREATE FUNCTION public.%s() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'synthetic renewal job failure'; END $$`, name)); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Exec("DROP FUNCTION IF EXISTS public." + name + "() CASCADE") })
	if _, err := db.Exec(fmt.Sprintf(`CREATE TRIGGER %s BEFORE UPDATE ON identity.email_delivery_jobs FOR EACH ROW WHEN (OLD.intent_id='%s' AND NEW.generation<>OLD.generation) EXECUTE FUNCTION public.%s()`, name, id, name)); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Exec("DROP TRIGGER IF EXISTS " + name + " ON identity.email_delivery_jobs") })
	if err := service.BeginEmailRegistration(ctx, email); err == nil {
		t.Fatal("injected failure was not observed")
	}
	var generation, jobGeneration int
	if err := db.QueryRow(`SELECT i.generation,j.generation FROM identity.registration_intents i JOIN identity.email_delivery_jobs j ON j.intent_id=i.id WHERE i.id=$1`, id).Scan(&generation, &jobGeneration); err != nil || generation != 1 || jobGeneration != 1 {
		t.Fatalf("partial transaction committed: %d/%d %v", generation, jobGeneration, err)
	}
	if _, err := db.Exec("DROP TRIGGER " + name + " ON identity.email_delivery_jobs"); err != nil {
		t.Fatal(err)
	}
	if err := service.BeginEmailRegistration(ctx, email); err != nil {
		t.Fatal(err)
	}
	t.Log("RENEWAL_JOB_FAILURE_TRANSACTION_ROLLBACK=PASS")
}
