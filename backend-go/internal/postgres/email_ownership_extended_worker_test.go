package postgres_test

import(
 "context"
 "crypto/sha256"
 "encoding/hex"
 "errors"
 "fmt"
 "testing"
 "github.com/google/uuid"
 "github.com/openinvest/openinvest/backend-go/internal/auth"
 "github.com/openinvest/openinvest/backend-go/internal/postgres"
)

func TestEmailOwnershipExtendedWorkerCrashOrderMatrix(t *testing.T){
 for _,mode:=range []string{"stale_completion_after_replacement","fourth_failed_delivery_immediately_terminal","intent_expiration_after_claim"}{
  t.Run(mode,func(t *testing.T){
   svc,db:=ownershipService(t);independentDrainPreviousDeliveries(t,svc)
   ctx:=context.Background()
   email:="worker-order-"+uuid.NewString()+"@example.test"
   if err:=svc.BeginEmailRegistration(ctx,email);err!=nil{t.Fatal(err)}
   if _,err:=db.Exec("UPDATE identity.registration_intents SET created_at=statement_timestamp()-interval '100 years' WHERE email_normalized=$1",email);err!=nil{t.Fatal(err)}
   t.Cleanup(func(){if _,err:=db.Exec("DELETE FROM identity.registration_intents WHERE email_normalized=$1",email);err!=nil{t.Error(err)}})
   worker,err:=postgres.Open(getOwnershipURL(t));if err!=nil{t.Fatal(err)};defer worker.Close()
   claim:=func(n int)auth.VerificationDelivery{
    t.Helper()
    h:=sha256.Sum256([]byte(fmt.Sprintf("%s-attempt-%d",email,n)))
    d,found,err:=worker.LeaseVerificationDelivery(ctx,hex.EncodeToString(h[:]))
    if err!=nil||!found||d.Email!=email{t.Fatalf("claim %d %t %v email=%s",n,found,err,d.Email)}
    return d
   }
   first:=claim(1)
   switch mode{
   case "stale_completion_after_replacement":
     if _,err:=db.Exec("UPDATE identity.email_delivery_jobs SET lease_until=statement_timestamp()-interval '1 second',available_at=statement_timestamp()-interval '1 second' WHERE id=$1",first.JobID);err!=nil{t.Fatal(err)}
     replacement:=claim(2)
     if replacement.LeaseID==first.LeaseID{t.Fatal("negative control: lease ID did not rotate")}
     if err:=worker.FinishVerificationDelivery(ctx,first,true);err!=nil{t.Fatal(err)}
     var current,leaseID string
     if err:=db.QueryRow("SELECT job_state,lease_id FROM identity.email_delivery_jobs WHERE id=$1",first.JobID).Scan(&current,&leaseID);err!=nil{t.Fatal(err)}
     if current!="leased"||leaseID!=replacement.LeaseID{t.Fatalf("stale worker stole active lease: %s %s",current,leaseID)}
     if err:=worker.FinishVerificationDelivery(ctx,replacement,false);err!=nil{t.Fatal(err)}
     if err:=db.QueryRow("SELECT job_state FROM identity.email_delivery_jobs WHERE id=$1",first.JobID).Scan(&current);err!=nil||current!="queued"{t.Fatalf("replacement failure did not queue: %s %v",current,err)}
   case "fourth_failed_delivery_immediately_terminal":
     current:=first
     for i:=1;i<=4;i++{
       if i>1{current=claim(i)}
       if err:=worker.FinishVerificationDelivery(ctx,current,false);err!=nil{t.Fatal(err)}
       if _,err:=db.Exec("UPDATE identity.email_delivery_jobs SET available_at=statement_timestamp()-interval '1 second' WHERE id=$1",current.JobID);err!=nil{t.Fatal(err)}
     }
     var state string;var attempts int
     if err:=db.QueryRow("SELECT job_state,attempts FROM identity.email_delivery_jobs WHERE id=$1",first.JobID).Scan(&state,&attempts);err!=nil{t.Fatal(err)}
     if state!="failed"||attempts!=4{t.Fatalf("exhaustion state %s attempts %d",state,attempts)}
     h:=sha256.Sum256([]byte(email+"-fifth"))
     _,found,err:=worker.LeaseVerificationDelivery(ctx,hex.EncodeToString(h[:]))
     if err!=nil||found{t.Fatalf("fifth attempt allowed or DB failure: %t %v",found,err)}
   case "intent_expiration_after_claim":
     if _,err:=db.Exec("UPDATE identity.registration_intents SET expires_at=statement_timestamp()-interval '1 second' WHERE email_normalized=$1",email);err!=nil{t.Fatal(err)}
     if _,err:=db.Exec("UPDATE identity.email_delivery_jobs SET lease_until=statement_timestamp()-interval '1 second',available_at=statement_timestamp()-interval '1 second' WHERE id=$1",first.JobID);err!=nil{t.Fatal(err)}
     h:=sha256.Sum256([]byte(email+"-expiry"))
     _,found,err:=worker.LeaseVerificationDelivery(ctx,hex.EncodeToString(h[:]))
     if err!=nil||found{t.Fatalf("expired intent received a fresh lease: %t %v",found,err)}
     if err:=worker.FinishVerificationDelivery(ctx,first,true);err!=nil{t.Fatal(err)}
   }
   var users,sessions,raw,intents int
   if err:=db.QueryRow("SELECT count(*) FROM identity.users WHERE email_normalized=$1",email).Scan(&users);err!=nil{t.Fatal(err)}
   if err:=db.QueryRow("SELECT count(*) FROM identity.sessions s JOIN identity.users u ON s.user_id=u.id WHERE u.email_normalized=$1",email).Scan(&sessions);err!=nil{t.Fatal(err)}
   if err:=db.QueryRow("SELECT count(*) FROM identity.email_verification_tokens t JOIN identity.registration_intents i ON t.intent_id=i.id WHERE i.email_normalized=$1 AND (length(t.token_hash)<>64 OR t.token_hash NOT SIMILAR TO '[0-9a-f]{64}')",email).Scan(&raw);err!=nil{t.Fatal(err)}
   if err:=db.QueryRow("SELECT count(*) FROM identity.email_verification_tokens t JOIN identity.registration_intents i ON t.intent_id=i.id WHERE i.email_normalized=$1",email).Scan(&intents);err!=nil{t.Fatal(err)}
   if users!=0||sessions!=0||raw!=0||intents==0{t.Fatalf("worker mutated authentication or stored raw proof: users=%d sessions=%d nonhash=%d tokenRows=%d",users,sessions,raw,intents)}
   if err:=svc.CompleteEmailRegistration(ctx,"definitely_invalid",ownerRequest());!errors.Is(err,auth.ErrInvalidVerification){t.Fatalf("invalid proof became valid: %v",err)}
   t.Logf("WORKER_CRASH_ORDER=%s USER_CREATION=NO SESSION=NO TOKEN_DIGEST_ONLY=YES",mode)
  })
 }
}
