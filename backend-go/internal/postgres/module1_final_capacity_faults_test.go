package postgres_test

import (
 "context"
 "database/sql"
 "errors"
 "strings"
 "sync"
 "testing"
 "time"

 "github.com/google/uuid"
 "github.com/openinvest/openinvest/backend-go/internal/auth"
)

// All attacks use disposable PostgreSQL 18. Nothing is sent to external SMTP.
func TestEmailOwnershipFinalRegistrationSaturationAndRecovery(t *testing.T) {
 requireIndependentPG(t)
 s,db:=ownershipService(t)
 independentDrainPreviousDeliveries(t,s)
 ctx,cancel:=context.WithTimeout(context.Background(),35*time.Second);defer cancel()
 prefix:="cap-"+strings.ReplaceAll(uuid.NewString(),"-","")+"-"
 cleanup:=func(){c,cancel:=context.WithTimeout(context.Background(),25*time.Second);defer cancel();if _,err:=db.ExecContext(c,"DELETE FROM identity.registration_intents WHERE email_normalized LIKE $1",prefix+"%");err!=nil{t.Errorf("fixture cleanup: %v",err)}}
 t.Cleanup(cleanup)
 const limit=10000
 inserted,err:=db.ExecContext(ctx,`INSERT INTO identity.registration_intents(id,email_normalized,created_at,expires_at,next_delivery_at)
 SELECT gen_random_uuid(),$1 || n::text || '@example.test',statement_timestamp(),statement_timestamp()+interval '8 hours',statement_timestamp()
 FROM generate_series(1,10000) AS n`,prefix)
 if err!=nil{t.Fatal(err)}
 rows,err:=inserted.RowsAffected();if err!=nil||rows!=limit{t.Fatalf("fixture incomplete inserted=%d error=%v",rows,err)}
 var active int
 if err:=db.QueryRowContext(ctx,"SELECT count(*) FROM identity.registration_intents WHERE intent_state='pending' AND expires_at>statement_timestamp()").Scan(&active);err!=nil{t.Fatal(err)}
 if active<limit{t.Fatalf("negative control: no saturation, active=%d",active)}
 names:=make([]string,20);for i:=range names{names[i]=independentAddress("overcap")}
 outcomes:=make(chan error,len(names));var wg sync.WaitGroup
 for _,name:=range names {wg.Add(1);go func(email string){defer wg.Done();outcomes<-s.BeginEmailRegistration(ctx,email)}(name)}
 wg.Wait();close(outcomes)
 for err:=range outcomes{if err!=nil{t.Fatalf("saturated registration error reveals capacity: %v",err)}}
 for _,name:=range names{
  var n int
  if err:=db.QueryRowContext(ctx,"SELECT count(*) FROM identity.registration_intents WHERE email_normalized=$1",name).Scan(&n);err!=nil||n!=0{t.Fatalf("new intent admitted beyond cap: email=%q rows=%d err=%v",name,n,err)}
  if err:=db.QueryRowContext(ctx,"SELECT count(*) FROM identity.users WHERE email_normalized=$1",name).Scan(&n);err!=nil||n!=0{t.Fatalf("account created without proof under saturation: %d %v",n,err)}
 }
 if _,err:=db.ExecContext(ctx,"UPDATE identity.registration_intents SET expires_at=statement_timestamp()-interval '1 second' WHERE email_normalized LIKE $1",prefix+"%");err!=nil{t.Fatal(err)}
 recovery:=independentAddress("after-cap")
 if err:=s.BeginEmailRegistration(ctx,recovery);err!=nil{t.Fatalf("registration did not recover after expired capacity: %v",err)}
 var n int
 if err:=db.QueryRowContext(ctx,"SELECT count(*) FROM identity.registration_intents WHERE email_normalized=$1 AND intent_state='pending'",recovery).Scan(&n);err!=nil||n!=1{t.Fatalf("capacity recovery unavailable: %d %v",n,err)}
 if err:=db.QueryRowContext(ctx,"SELECT count(*) FROM identity.users WHERE email_normalized=$1",recovery).Scan(&n);err!=nil||n!=0{t.Fatalf("recovery wrongly activated user %d %v",n,err)}
 t.Logf("REGISTRATION_CAPACITY_LIMIT=%d SATURATION_ATTEMPTS=%d ADMITTED_BEYOND_CAP=0 EXPIRED_RECOVERY=PASS",limit,len(names))
}

func TestEmailOwnershipFinalMultiPointTransactionalFaultInjection(t *testing.T) {
 requireIndependentPG(t)
 targets:=[]struct{table string}{
  {"identity.users"},{"investment.subjects"},{"identity.user_investment_links"},{"identity.credentials"},
  {"identity.privacy_settings"},{"identity.user_email_verifications"},
 }
 for _,tc:=range targets{
  t.Run(strings.ReplaceAll(tc.table,".","_"),func(t *testing.T){
   s,db:=ownershipService(t);independentDrainPreviousDeliveries(t,s)
   ctx,cancel:=context.WithTimeout(context.Background(),11*time.Second);defer cancel()
   email:=independentAddress("rollback-"+strings.ReplaceAll(tc.table,".","-"))
   if err:=s.BeginEmailRegistration(ctx,email);err!=nil{t.Fatal(err)}
   mail:=&ownershipMail{}
   if found,err:=s.DeliverOne(ctx,mail);err!=nil||!found||mail.email!=email{t.Fatalf("negative delivery control: found=%v recipient=%q err=%v",found,mail.email,err)}
   tx,err:=db.BeginTx(ctx,&sql.TxOptions{});if err!=nil{t.Fatal(err)}
   released:=false
   defer func(){if !released { _=tx.Rollback() }}()
   // Only static table names from the source literal allowlist.
   if _,err:=tx.ExecContext(ctx,"LOCK TABLE "+tc.table+" IN ACCESS EXCLUSIVE MODE");err!=nil{t.Fatal(err)}
   outcome:=make(chan error,1)
   go func(){outcome<-s.CompleteEmailRegistration(ctx,mail.token,ownerRequest())}()
   select{
   case err=<-outcome:if err==nil{t.Fatalf("activation committed while %s was locked",tc.table)}
   case <-time.After(7*time.Second):t.Fatalf("activation did not fail within bounded lock timeout at %s",tc.table)
   }
   if err:=tx.Rollback();err!=nil{t.Fatal(err)};released=true
   var n int
   for _,query:=range []string{
    "SELECT count(*) FROM identity.users WHERE email_normalized=$1",
    "SELECT count(*) FROM identity.sessions s JOIN identity.users u ON s.user_id=u.id WHERE u.email_normalized=$1",
    "SELECT count(*) FROM investment.subjects x JOIN identity.user_investment_links l ON x.id=l.investment_subject_id JOIN identity.users u ON u.id=l.user_id WHERE u.email_normalized=$1",
   }{
    if err:=db.QueryRowContext(ctx,query,email).Scan(&n);err!=nil||n!=0{t.Fatalf("partial database state at %s: rows=%d err=%v",tc.table,n,err)}
   }
   var consumed sql.NullTime
   if err:=db.QueryRowContext(ctx,`SELECT t.consumed_at FROM identity.email_verification_tokens t JOIN identity.registration_intents i ON i.id=t.intent_id WHERE i.email_normalized=$1 ORDER BY t.created_at DESC LIMIT 1`,email).Scan(&consumed);err!=nil||consumed.Valid{t.Fatalf("verification token consumed on rollback at %s: %v %v",tc.table,consumed,err)}
   if err:=s.CompleteEmailRegistration(ctx,mail.token,ownerRequest());err!=nil{t.Fatalf("valid retry after fault at %s did not commit: %v",tc.table,err)}
   if err:=db.QueryRowContext(ctx,"SELECT count(*) FROM identity.users WHERE email_normalized=$1 AND account_state='active'",email).Scan(&n);err!=nil||n!=1{t.Fatalf("retry at %s caused wrong final users=%d err=%v",tc.table,n,err)}
   if err:=s.CompleteEmailRegistration(ctx,mail.token,ownerRequest());!errors.Is(err,auth.ErrInvalidVerification){t.Fatalf("replayed proof after successful retry at %s: %v",tc.table,err)}
   t.Logf("FAULT_AT=%s ROLLBACK_NO_PARTIAL_STATE=YES PROOF_REUSABLE_AFTER_ABORT=YES SINGLE_FINAL_USER=YES",tc.table)
  })
 }
}

func TestEmailOwnershipFinalIntentExpiryRejectsLateProof(t *testing.T){
 requireIndependentPG(t)
 s,db:=ownershipService(t);independentDrainPreviousDeliveries(t,s)
 ctx:=context.Background();email:=independentAddress("intent-expiry")
 if err:=s.BeginEmailRegistration(ctx,email);err!=nil{t.Fatal(err)}
 mail:=&ownershipMail{};if found,err:=s.DeliverOne(ctx,mail);err!=nil||!found||mail.email!=email{t.Fatalf("no prior verified delivery: %t %v",found,err)}
 if _,err:=db.Exec("UPDATE identity.registration_intents SET expires_at=statement_timestamp()-interval '1 second' WHERE email_normalized=$1",email);err!=nil{t.Fatal(err)}
 if err:=s.CompleteEmailRegistration(ctx,mail.token,ownerRequest());!errors.Is(err,auth.ErrInvalidVerification){t.Fatalf("expired intent accepted: %v",err)}
 var n int
 if err:=db.QueryRow("SELECT count(*) FROM identity.users WHERE email_normalized=$1",email).Scan(&n);err!=nil||n!=0{t.Fatalf("expired intent activated user: %v %d",err,n)}
 if err:=s.BeginEmailRegistration(ctx,email);err!=nil{t.Fatalf("new generation unavailable: %v",err)}
 if err:=s.CompleteEmailRegistration(ctx,mail.token,ownerRequest());!errors.Is(err,auth.ErrInvalidVerification){t.Fatalf("old generation proof reclaimed new intent: %v",err)}
 if err:=db.QueryRow("SELECT generation FROM identity.registration_intents WHERE email_normalized=$1",email).Scan(&n);err!=nil||n!=2{t.Fatalf("new generation not issued: gen=%d err=%v",n,err)}
 t.Log("INTENT_EXPIRED_AND_OLD_PROOF_CANNOT_ACTIVATE=PASS")
}
