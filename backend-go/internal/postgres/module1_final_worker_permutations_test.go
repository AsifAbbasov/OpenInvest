package postgres_test

import (
 "context"
 "crypto/sha256"
 "encoding/hex"
 "fmt"
 "os"
 "sync"
 "testing"

 "github.com/openinvest/openinvest/backend-go/internal/postgres"
 "github.com/openinvest/openinvest/backend-go/internal/auth"
)

func finalDigest(s string)string{sum:=sha256.Sum256([]byte(s));return hex.EncodeToString(sum[:])}

// This is independent of Builder's test implementation. Every attempt takes
// the real lock path in PostgreSQL 18, using a unique digest and a separate worker.
func TestEmailOwnershipFinalReplacementLeaseAndStaleCallbacks(t *testing.T){
 requireIndependentPG(t)
 for _,sent:=range []bool{true,false}{
  t.Run(fmt.Sprintf("stale_sent_%t",sent),func(t *testing.T){
   s,db:=ownershipService(t);independentDrainPreviousDeliveries(t,s)
   ctx:=context.Background();email:=independentAddress("replace-lease")
   if err:=s.BeginEmailRegistration(ctx,email);err!=nil{t.Fatal(err)}
   worker,err:=postgres.Open(os.Getenv("OPENINVEST_DATABASE_TEST_URL"));if err!=nil{t.Fatal(err)};defer worker.Close()
   first,found,err:=worker.LeaseVerificationDelivery(ctx,finalDigest(email+"-lease1"))
   if err!=nil||!found||first.Email!=email{t.Fatalf("first claim: %+v %t %v",first,found,err)}
   if _,err:=db.Exec("UPDATE identity.email_delivery_jobs SET lease_until=statement_timestamp()-interval '2 seconds',available_at=statement_timestamp()-interval '2 seconds' WHERE id=$1",first.JobID);err!=nil{t.Fatal(err)}
   second,found,err:=worker.LeaseVerificationDelivery(ctx,finalDigest(email+"-lease2"))
   if err!=nil||!found||second.Email!=email||second.JobID!=first.JobID||second.LeaseID==first.LeaseID{t.Fatalf("replacement lease missing or unchanged: %+v %+v %t %v",first,second,found,err)}
   if err:=worker.FinishVerificationDelivery(ctx,first,sent);err!=nil{t.Fatal(err)}
   var state,lease string;var attempts int
   if err:=db.QueryRow("SELECT job_state,lease_id,attempts FROM identity.email_delivery_jobs WHERE id=$1",second.JobID).Scan(&state,&lease,&attempts);err!=nil{t.Fatal(err)}
   if state!="leased"||lease!=second.LeaseID||attempts!=2{t.Fatalf("stale callback took ownership: sent=%t state=%s lease_ok=%t attempts=%d",sent,state,lease==second.LeaseID,attempts)}
   if err:=worker.FinishVerificationDelivery(ctx,second,true);err!=nil{t.Fatal(err)}
   if err:=db.QueryRow("SELECT job_state,attempts FROM identity.email_delivery_jobs WHERE id=$1",second.JobID).Scan(&state,&attempts);err!=nil||state!="delivered"||attempts!=2{t.Fatalf("valid current completion did not stick: %s %d %v",state,attempts,err)}
   if err:=worker.FinishVerificationDelivery(ctx,first,!sent);err!=nil{t.Fatal(err)}
   if err:=db.QueryRow("SELECT job_state,attempts FROM identity.email_delivery_jobs WHERE id=$1",second.JobID).Scan(&state,&attempts);err!=nil||state!="delivered"||attempts!=2{t.Fatalf("stale callback mutated finished job: %s %d %v",state,attempts,err)}
  })
 }
}

func TestEmailOwnershipFinalExhaustionConcurrentRecoveryAndLateCompletion(t *testing.T){
 requireIndependentPG(t)
 s,db:=ownershipService(t);independentDrainPreviousDeliveries(t,s)
 ctx:=context.Background();email:=independentAddress("recovery-race")
 if err:=s.BeginEmailRegistration(ctx,email);err!=nil{t.Fatal(err)}
 worker,err:=postgres.Open(os.Getenv("OPENINVEST_DATABASE_TEST_URL"));if err!=nil{t.Fatal(err)};defer worker.Close()
 var finalLeaseID,jobID string
 for attempt:=1;attempt<=4;attempt++{
  d,found,err:=worker.LeaseVerificationDelivery(ctx,finalDigest(fmt.Sprintf("%s-attempt-%d",email,attempt)))
  if err!=nil||!found||d.Email!=email{t.Fatalf("lease %d found=%t email=%s err=%v",attempt,found,d.Email,err)}
  jobID=d.JobID;finalLeaseID=d.LeaseID
  if _,err:=db.Exec("UPDATE identity.email_delivery_jobs SET lease_until=statement_timestamp()-interval '2 seconds', available_at=statement_timestamp()-interval '2 seconds' WHERE id=$1",d.JobID);err!=nil{t.Fatal(err)}
 }
 const contenders=16
 var wg sync.WaitGroup;start:=make(chan struct{})
 results:=make(chan error,contenders)
 for i:=0;i<contenders;i++{wg.Add(1);go func(n int){defer wg.Done();<-start
  claim,found,err:=worker.LeaseVerificationDelivery(ctx,finalDigest(fmt.Sprintf("%s-postlimit-%d",email,n)))
  if err!=nil{results<-err}else if found&&claim.Email==email{results<-fmt.Errorf("fifth claim admitted for %s",email)}else{results<-nil}
 }(i)}
 close(start);wg.Wait();close(results)
 for err:=range results{if err!=nil{t.Fatal(err)}}
 var state string;var attempts int;var lease *string
 if err:=db.QueryRow("SELECT job_state,attempts,lease_id FROM identity.email_delivery_jobs WHERE id=$1",jobID).Scan(&state,&attempts,&lease);err!=nil{t.Fatal(err)}
 if state!="failed"||attempts!=4||lease!=nil{t.Fatalf("terminalization wrong after %d concurrent recoveries: state=%s attempts=%d lease=%v",contenders,state,attempts,lease)}
 // Late successful completion from the expired final owner may never
 // resurrect an exhausted job.
 if err:=worker.FinishVerificationDelivery(ctx,auth.VerificationDelivery{JobID:jobID,LeaseID:finalLeaseID,Email:email},true);err!=nil{t.Fatal(err)}
 if err:=db.QueryRow("SELECT job_state,attempts FROM identity.email_delivery_jobs WHERE id=$1",jobID).Scan(&state,&attempts);err!=nil||state!="failed"||attempts!=4{t.Fatalf("stale final owner resurrected job: %s %d %v",state,attempts,err)}
 var sessions,users int
 if err:=db.QueryRow("SELECT count(*) FROM identity.users WHERE email_normalized=$1",email).Scan(&users);err!=nil||users!=0{t.Fatalf("crashed delivery activated user: %d %v",users,err)}
 if err:=db.QueryRow("SELECT count(*) FROM identity.sessions s JOIN identity.users u ON u.id=s.user_id WHERE u.email_normalized=$1",email).Scan(&sessions);err!=nil||sessions!=0{t.Fatalf("unproved session created: %d %v",sessions,err)}
 t.Logf("RECOVERY_CONTENDERS=%d TERMINAL_STATE=failed ATTEMPTS=4 FIFTH_ATTEMPTS=0 STALE_COMPLETION_BLOCKED=YES",contenders)
}

func TestEmailOwnershipFinalRenewalWhilePriorLeaseUnfinished(t *testing.T){
 requireIndependentPG(t)
 for _,mode:=range []string{"delivered_old_generation_control","unfinished_old_generation"}{
  t.Run(mode,func(t *testing.T){
   s,db:=ownershipService(t);independentDrainPreviousDeliveries(t,s)
   ctx:=context.Background();email:=independentAddress("renew-"+mode)
   if err:=s.BeginEmailRegistration(ctx,email);err!=nil{t.Fatal(err)}
   worker,err:=postgres.Open(os.Getenv("OPENINVEST_DATABASE_TEST_URL"));if err!=nil{t.Fatal(err)};defer worker.Close()
   first,found,err:=worker.LeaseVerificationDelivery(ctx,finalDigest(email+"-gen1"))
   if err!=nil||!found||first.Email!=email{t.Fatalf("generation 1 claim unavailable: %t %v",found,err)}
   if mode=="delivered_old_generation_control"{
    if err:=worker.FinishVerificationDelivery(ctx,first,true);err!=nil{t.Fatal(err)}
   }
   if _,err:=db.Exec("UPDATE identity.registration_intents SET expires_at=statement_timestamp()-interval '2 seconds' WHERE email_normalized=$1",email);err!=nil{t.Fatal(err)}
   if err:=s.BeginEmailRegistration(ctx,email);err!=nil{t.Fatal(err)}
   var generation,jobGeneration int64
   if err:=db.QueryRow(`SELECT i.generation,j.generation FROM identity.registration_intents i JOIN identity.email_delivery_jobs j ON j.intent_id=i.id WHERE i.email_normalized=$1`,email).Scan(&generation,&jobGeneration);err!=nil{t.Fatal(err)}
   if generation!=2{t.Fatalf("renewal did not advance generation: %d",generation)}
   // A renewed eligible account must be able to request new instructions,
   // even if the previous generation crashed while leased.
   if jobGeneration!=2{t.Fatalf("renewed intent stranded behind old %s lease: intent_gen=%d job_gen=%d",mode,generation,jobGeneration)}
   next,found,err:=worker.LeaseVerificationDelivery(ctx,finalDigest(email+"-gen2"))
   if err!=nil||!found||next.Email!=email{t.Fatalf("renewed generation delivery blocked after %s: found=%t email=%s err=%v",mode,found,next.Email,err)}
   if err:=worker.FinishVerificationDelivery(ctx,first,true);err!=nil{t.Fatal(err)}
   var lease string
   if err:=db.QueryRow("SELECT lease_id FROM identity.email_delivery_jobs WHERE id=$1",next.JobID).Scan(&lease);err!=nil||lease!=next.LeaseID{t.Fatalf("old generation completion stole successor lease: %s %v",lease,err)}
  })
 }
}
