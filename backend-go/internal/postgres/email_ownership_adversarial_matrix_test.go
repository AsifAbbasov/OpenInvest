package postgres_test

import (
 "context"
 "os"
 "encoding/json"
 "errors"
 "io"
 "net/http"
 "net/http/httptest"
 "strings"
 "sync"
 "testing"
 "time"

 "github.com/google/uuid"
 "github.com/openinvest/openinvest/backend-go/internal/auth"
 "github.com/openinvest/openinvest/backend-go/internal/httpapi"
 "github.com/openinvest/openinvest/backend-go/internal/postgres"
 "github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

// Drain synthetic, already-queued fake deliveries before beginning a new probe.
// Never calls external SMTP; bound prevents a polluted fixture from passing.
func independentDrainPreviousDeliveries(t *testing.T,svc *auth.Service) {
 t.Helper()
 for i:=0;i<512;i++ {
  found,err:=svc.DeliverOne(context.Background(),&ownershipMail{})
  if err!=nil{t.Fatalf("fixture precondition: %v",err)}
  if !found{return}
 }
 t.Fatal("TEST_FIXTURE_DEFECT: residual delivery work exceeds finite fixture budget")
}
func independentAddress(label string) string {return "adv02-"+label+"-"+uuid.NewString()+"@example.test"}
func TestEmailOwnershipAdversarialHTTPMembershipParity(t *testing.T) {
 s,db:=ownershipService(t)
 independentDrainPreviousDeliveries(t,s)
 app,err:=httpapi.New(verticalslice.NewService(&postgres.Store{},verticalslice.SystemClock{}),s,[]byte("verification-independent-32-byte-secret"))
 if err!=nil {t.Fatal(err)}
 fresh:=independentAddress("fresh")
 legacy:=independentAddress("legacy")
 seed:=ownerRequest();seed.Email=legacy;seed.Password="legacy-original-password-2026"
 if _,err:=s.Register(context.Background(),seed);err!=nil{t.Fatal(err)}
 pending:=independentAddress("pending")
 expired:=independentAddress("expired")
 for _,email:=range []string{pending,expired}{if err:=s.BeginEmailRegistration(context.Background(),email);err!=nil{t.Fatal(err)}}
 if _,err:=db.Exec("UPDATE identity.registration_intents SET expires_at=statement_timestamp()-interval '1 second' WHERE email_normalized=$1",expired);err!=nil{t.Fatal(err)}
 cases:=[]struct{name,email string}{{"fresh",fresh},{"legacy",legacy},{"pending",pending},{"expired",expired},{"normalized","  "+strings.ToUpper(fresh)+"  "}}
 var canonical string
 for _,tc:=range cases{
  t.Run(tc.name,func(t *testing.T){
   raw,_:=json.Marshal(map[string]string{"email":tc.email})
   req:=httptest.NewRequest(http.MethodPost,"/api/v1/auth/register",strings.NewReader(string(raw)))
   req.Header.Set("Content-Type","application/json")
   resp,err:=app.Test(req);if err!=nil{t.Fatal(err)};defer resp.Body.Close()
   body,err:=io.ReadAll(resp.Body);if err!=nil{t.Fatal(err)}
   var envelope map[string]json.RawMessage
   if err:=json.Unmarshal(body,&envelope);err!=nil{t.Fatal(err)}
   var data map[string]any
   if err:=json.Unmarshal(envelope["data"],&data);err!=nil{t.Fatal(err)}
   msg:=data["message"]
   if resp.StatusCode!=202 || data["status"]!="REQUEST_ACCEPTED" || len(data)!=2 || len(resp.Cookies())!=0 {
      t.Fatalf("membership leak %s status=%d data=%v cookie=%d",tc.name,resp.StatusCode,data,len(resp.Cookies()))
   }
   if resp.Header.Get("Cache-Control")!="no-store" {t.Fatalf("cacheable registration response: %q",resp.Header.Get("Cache-Control"))}
   if canonical==""{canonical=msg.(string)} else if canonical!=msg{t.Fatalf("membership-specific response text: %q vs %q",canonical,msg)}
  })
 }
 for _,email:=range []string{fresh,pending,expired}{
  var n int
  if err:=db.QueryRow("SELECT count(*) FROM identity.users WHERE email_normalized=$1",email).Scan(&n);err!=nil||n!=0{t.Fatalf("premature account %s rows=%d err=%v",email,n,err)}
 }
 t.Log("INDEPENDENT_REGISTRATION_BODY_STATUS_COOKIE_PARITY=PASS")
}

func TestEmailOwnershipAdversarialCrossIntentAndPurpose(t *testing.T){
 s,db:=ownershipService(t);independentDrainPreviousDeliveries(t,s);ctx:=context.Background()
 a,b:=independentAddress("a"),independentAddress("b")
 for _,v:=range []string{a,b}{if err:=s.BeginEmailRegistration(ctx,v);err!=nil{t.Fatal(err)}}
 ma:=&ownershipMail{};mb:=&ownershipMail{}
 if ok,err:=s.DeliverOne(ctx,ma);err!=nil||!ok{t.Fatalf("send A %t %v",ok,err)}
 if ok,err:=s.DeliverOne(ctx,mb);err!=nil||!ok{t.Fatalf("send B %t %v",ok,err)}
 if ma.email==mb.email || ma.token==mb.token {t.Fatal("negative control invalid: same principal or token")}
 valid:=[]string{ma.token,mb.token}
 for _,tok:=range valid{
  for _,bad:=range []string{tok[:len(tok)-1]+"*",tok+"=",tok[:10]+"."+tok[11:],"",strings.Repeat("A",43)} {
   if err:=s.CompleteEmailRegistration(ctx,bad,ownerRequest());!errors.Is(err,auth.ErrInvalidVerification) {t.Fatalf("invalid token accepted or unexpected error: %v",err)}
  }
 }
 // The SQL purpose constraint must reject repurposing at the data boundary.
 if _,err:=db.Exec("UPDATE identity.email_verification_tokens SET purpose='password_reset' WHERE purpose='new_registration'");err==nil {t.Fatal("purpose CHECK allowed illegal token purpose")}
 if err:=s.CompleteEmailRegistration(ctx,ma.token,ownerRequest());err!=nil{t.Fatal(err)}
 var usersA,usersB int
 if err:=db.QueryRow("SELECT count(*) FROM identity.users WHERE email_normalized=$1",ma.email).Scan(&usersA);err!=nil{t.Fatal(err)}
 if err:=db.QueryRow("SELECT count(*) FROM identity.users WHERE email_normalized=$1",mb.email).Scan(&usersB);err!=nil{t.Fatal(err)}
 if usersA!=1||usersB!=0{t.Fatalf("cross-principal activation mismatch A=%d B=%d",usersA,usersB)}
 if err:=s.CompleteEmailRegistration(ctx,mb.token,ownerRequest());err!=nil{t.Fatal(err)}
 if err:=db.QueryRow("SELECT count(*) FROM identity.users WHERE email_normalized=$1",mb.email).Scan(&usersB);err!=nil||usersB!=1{t.Fatalf("second owner lost: %d %v",usersB,err)}
 t.Log("INDEPENDENT_CROSS_INTENT_BOUND_SECRET_AND_PURPOSE_CHECK=PASS")
}

func TestEmailOwnershipAdversarialRollbackMidActivation(t *testing.T){
 s,db:=ownershipService(t);independentDrainPreviousDeliveries(t,s);ctx:=context.Background()
 email:=independentAddress("rollback")
 if err:=s.BeginEmailRegistration(ctx,email);err!=nil{t.Fatal(err)}
 sink:=&ownershipMail{};if ok,err:=s.DeliverOne(ctx,sink);!ok||err!=nil{t.Fatalf("send %t %v",ok,err)}
 lock,err:=db.BeginTx(ctx,nil);if err!=nil{t.Fatal(err)}
 if _,err=lock.Exec("LOCK TABLE identity.credentials IN ACCESS EXCLUSIVE MODE");err!=nil{t.Fatal(err)}
 done:=make(chan error,1)
 go func(){done<-s.CompleteEmailRegistration(ctx,sink.token,ownerRequest())}()
 // An intentional lock timeout must abort all intermediate INSERTs.
 select{case err=<-done:
   if err==nil {t.Fatal("activation unexpectedly committed under blocked credentials")}
 case <-time.After(7*time.Second):t.Fatal("activation exceeded bounded transaction budget")
 }
 if err:=lock.Rollback();err!=nil{t.Fatal(err)}
 var rows int
 for _,query:=range []string{
  "SELECT count(*) FROM identity.users WHERE email_normalized=$1",
  "SELECT count(*) FROM identity.user_email_verifications v JOIN identity.users u ON u.id=v.user_id WHERE u.email_normalized=$1",
  "SELECT count(*) FROM identity.user_investment_links l JOIN identity.users u ON u.id=l.user_id WHERE u.email_normalized=$1",
 }{
  if err:=db.QueryRow(query,email).Scan(&rows);err!=nil||rows!=0{t.Fatalf("partial persisted after rollback: rows=%d err=%v",rows,err)}
 }
 if err:=s.CompleteEmailRegistration(ctx,sink.token,ownerRequest());err!=nil{t.Fatalf("token consumed despite rollback: %v",err)}
 t.Log("INDEPENDENT_TRANSACTION_ROLLBACK_AFTER_MID_INSERT_FAILURE=PASS")
}

func TestEmailOwnershipAdversarialDeliveryRetriesAndWorkers(t *testing.T){
 s,db:=ownershipService(t);independentDrainPreviousDeliveries(t,s);ctx:=context.Background()
 email:=independentAddress("deliveries")
 if err:=s.BeginEmailRegistration(ctx,email);err!=nil{t.Fatal(err)}
 // Two workers concurrently see one queued job.
 var wg sync.WaitGroup
 type outcome struct{found bool;err error;mail *ownershipMail}
 ch:=make(chan outcome,2)
 for i:=0;i<2;i++{wg.Add(1);go func(){defer wg.Done();m:=&ownershipMail{};ok,err:=s.DeliverOne(ctx,m);ch<-outcome{ok,err,m}}()}
 wg.Wait();close(ch)
 wins:=0
 for out:=range ch{if out.err!=nil{t.Fatal(out.err)};if out.found {wins++;if out.mail.email!=email{t.Fatalf("unexpected recipient %s",out.mail.email)}}}
 if wins!=1{t.Fatalf("concurrent workers both claimed same email: %d",wins)}
 // Restart/new worker after a failed send on a new generation.
 second:=independentAddress("failed-retry")
 if err:=s.BeginEmailRegistration(ctx,second);err!=nil{t.Fatal(err)}
 failed:=&ownershipMail{fail:true}
 if ok,err:=s.DeliverOne(ctx,failed);!ok||!errors.Is(err,auth.ErrVerificationUnavailable) {t.Fatalf("synthetic transport failure not isolated: %t %v",ok,err)}
 var state string;var attempts int
 if err:=db.QueryRow("SELECT j.job_state,j.attempts FROM identity.email_delivery_jobs j JOIN identity.registration_intents i ON i.id=j.intent_id WHERE i.email_normalized=$1",second).Scan(&state,&attempts);err!=nil{t.Fatal(err)}
 if state!="queued"||attempts!=1{t.Fatalf("not correctly queued for retry: state=%s attempts=%d",state,attempts)}
 if _,err:=db.Exec("UPDATE identity.email_delivery_jobs SET available_at=statement_timestamp()-interval '1 second' WHERE intent_id=(SELECT id FROM identity.registration_intents WHERE email_normalized=$1)",second);err!=nil{t.Fatal(err)}
 // Force first email job out of discovery to make the oracle deterministic.
 freshWorker,err:=postgres.Open(os.Getenv("OPENINVEST_DATABASE_TEST_URL"));if err!=nil{t.Fatal(err)};defer freshWorker.Close()
 restarted,err:=auth.NewService(freshWorker,nil,auth.Config{AccessTokenSecret:[]byte("01234567890123456789012345678901")});if err!=nil{t.Fatal(err)}
 recovered:=&ownershipMail{}
 if ok,err:=restarted.DeliverOne(ctx,recovered);err!=nil||!ok{t.Fatalf("restart cannot retry failed delivery: %t %v",ok,err)}
 if recovered.email!=second{t.Fatalf("wrong retry target %s wanted %s",recovered.email,second)}
 if err:=db.QueryRow("SELECT j.job_state,j.attempts FROM identity.email_delivery_jobs j JOIN identity.registration_intents i ON i.id=j.intent_id WHERE i.email_normalized=$1",second).Scan(&state,&attempts);err!=nil{t.Fatal(err)}
 if state!="delivered"||attempts!=2{t.Fatalf("restart state mismatch: %s %d",state,attempts)}
 t.Log("INDEPENDENT_CONCURRENT_WORKER_SINGLE_CLAIM_RETRY_AFTER_RESTART=PASS")
}
