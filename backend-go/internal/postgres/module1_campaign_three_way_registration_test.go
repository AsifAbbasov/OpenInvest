package postgres_test

// New three-way transition schedule: renewal of an expired registration,
// attempted old-token completion, and stale worker acknowledgment. All use
// actual PostgreSQL state, not a mock store or replayed historical result.
import (
 "bytes"
 "context"
 "crypto/sha256"
 "encoding/base64"
 "encoding/hex"
 "errors"
 "fmt"
 "sync"
 "testing"

 "github.com/google/uuid"
 "github.com/openinvest/openinvest/backend-go/internal/auth"
 "github.com/openinvest/openinvest/backend-go/internal/postgres"
)

func campaignToken(seed byte) (string,string){
 raw:=base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{seed},32))
 h:=sha256.Sum256([]byte(raw))
 return raw,hex.EncodeToString(h[:])
}
func TestModule1CampaignRenewalOldProofAndStaleWorkerThreeWay(t *testing.T){
 chaosAllowed(t)
 service,db:=ownershipService(t)
 ctx:=context.Background()
 email:="campaign-threeway-"+uuid.NewString()+"@example.test"
 if e:=service.BeginEmailRegistration(ctx,email);e!=nil{t.Fatal(e)}
 store,e:=postgres.Open(getOwnershipURL(t));if e!=nil{t.Fatal(e)}
 defer store.Close()
 oldRaw,oldHash:=campaignToken(41)
 oldLease,found,e:=store.LeaseVerificationDelivery(ctx,oldHash)
 if e!=nil||!found||oldLease.Email!=email{t.Fatalf("negative control lease A failed found=%t err=%v",found,e)}
 if _,e=db.Exec("UPDATE identity.registration_intents SET expires_at=statement_timestamp()-interval '1 second' WHERE email_normalized=$1",email);e!=nil{t.Fatal(e)}
 start:=make(chan struct{})
 var wg sync.WaitGroup
 var renewalErr,oldProofErr,lateWorkerErr error
 wg.Add(3)
 go func(){defer wg.Done();<-start;renewalErr=service.BeginEmailRegistration(ctx,email)}()
 go func(){defer wg.Done();<-start;oldProofErr=service.CompleteEmailRegistration(ctx,oldRaw,ownerRequest())}()
 go func(){defer wg.Done();<-start;lateWorkerErr=store.FinishVerificationDelivery(ctx,oldLease,true)}()
 close(start);wg.Wait()
 if renewalErr!=nil{t.Fatalf("renewal failed: %v",renewalErr)}
 if !errors.Is(oldProofErr,auth.ErrInvalidVerification){t.Fatalf("expired old proof accepted/unexpected error: %v",oldProofErr)}
 if lateWorkerErr!=nil{t.Fatalf("stale worker completion error: %v",lateWorkerErr)}
 var gen,jobGen int64;var jobState string;var users,sessions int
 e=db.QueryRow("SELECT i.generation,j.generation,j.job_state,(SELECT count(*) FROM identity.users WHERE email_normalized=$1),(SELECT count(*) FROM identity.sessions s JOIN identity.users u ON u.id=s.user_id WHERE u.email_normalized=$1) FROM identity.registration_intents i JOIN identity.email_delivery_jobs j ON j.intent_id=i.id WHERE i.email_normalized=$1",email).Scan(&gen,&jobGen,&jobState,&users,&sessions)
 if e!=nil||gen!=2||jobGen!=2||users!=0||sessions!=0||jobState!="queued"{t.Fatalf("THREE_WAY_STATE_INVARIANT gen=%d jobGen=%d state=%s users=%d sessions=%d err=%v",gen,jobGen,jobState,users,sessions,e)}
 newRaw,newHash:=campaignToken(42)
 newLease,found,e:=store.LeaseVerificationDelivery(ctx,newHash)
 if e!=nil||!found||newLease.Email!=email||newLease.LeaseID==oldLease.LeaseID{t.Fatalf("NEW_GENERATION_LEASE_MISSING found=%t err=%v",found,e)}
 if e=store.FinishVerificationDelivery(ctx,oldLease,true);e!=nil{t.Fatal(e)}
 var leaseID string
 e=db.QueryRow("SELECT j.lease_id FROM identity.email_delivery_jobs j JOIN identity.registration_intents i ON i.id=j.intent_id WHERE i.email_normalized=$1",email).Scan(&leaseID)
 if e!=nil||leaseID!=newLease.LeaseID{t.Fatalf("STALE_CALLBACK_STOLE_NEW_LEASE got=%s err=%v",leaseID,e)}
 if e=store.FinishVerificationDelivery(ctx,newLease,true);e!=nil{t.Fatal(e)}
 if e=service.CompleteEmailRegistration(ctx,oldRaw,ownerRequest());!errors.Is(e,auth.ErrInvalidVerification){t.Fatalf("OLD_TOKEN_AFTER_NEW_DELIVERY_ACCEPTED: %v",e)}
 if e=service.CompleteEmailRegistration(ctx,newRaw,ownerRequest());e!=nil{t.Fatalf("NEW_GENERATION_OWNER_ACTIVATION_FAILED: %v",e)}
 if e=service.CompleteEmailRegistration(ctx,newRaw,ownerRequest());!errors.Is(e,auth.ErrInvalidVerification){t.Fatalf("NEW_TOKEN_REPLAY_ACCEPTED: %v",e)}
 var active int
 e=db.QueryRow("SELECT count(*) FROM identity.users WHERE email_normalized=$1 AND account_state='active'",email).Scan(&active)
 if e!=nil||active!=1{t.Fatalf("ACTIVE_ACCOUNT_COUNT=%d err=%v",active,e)}
 if n:=chaosCount(t,db,"SELECT count(*) FROM identity.sessions s JOIN identity.users u ON u.id=s.user_id WHERE u.email_normalized=$1",email);n!=0{t.Fatalf("premature session count=%d",n)}
 login,e:=service.Login(ctx,auth.LoginRequest{Email:email,Password:ownerRequest().Password})
 if e!=nil||login.User.ID==""{t.Fatalf("normal login after three-way race failed: %v",e)}
 t.Log(fmt.Sprintf("REAL_POSTGRES_REGISTRATION_3_WAY=PASS GENERATION=%d STALE_CALLBACK_REJECTED=YES OLD_PROOF_REJECTED=YES SINGLE_ACTIVE_IDENTITY=YES NO_PREPROOF_SESSION=YES NORMAL_LOGIN=PASS",gen))
}
