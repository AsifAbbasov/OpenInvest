package postgres_test

// This suite exercises real public HTTP handlers and genuine PostgreSQL 18
// SIGKILL/WAL recovery on a disposable GitHub-hosted runner.
import (
 "context"
 "encoding/json"
 "fmt"
 "io"
 "net/http"
 "net/http/httptest"
 "os"
 "strings"
 "testing"

 "github.com/google/uuid"
 "github.com/openinvest/openinvest/backend-go/internal/auth"
 "github.com/openinvest/openinvest/backend-go/internal/httpapi"
 "github.com/openinvest/openinvest/backend-go/internal/postgres"
 "github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)
func extendedRealHTTP(t *testing.T)(func(string,string,string)(int,[]byte,http.Header),*auth.Service){
 t.Helper()
 store,e:=postgres.Open(os.Getenv("OPENINVEST_DATABASE_TEST_URL"));if e!=nil{t.Fatal(e)}
 t.Cleanup(func(){_ =store.Close()})
 svc,e:=auth.NewService(store,nil,auth.Config{AccessTokenSecret:[]byte("extended-real-http-chaos-test-secret-32")})
 if e!=nil{t.Fatal(e)}
 app,e:=httpapi.New(verticalslice.NewService(store,verticalslice.SystemClock{}),svc,[]byte("extended-real-http-import-review-secret-32"))
 if e!=nil{t.Fatal(e)}
 request:=func(method,path,body string)(int,[]byte,http.Header){
  t.Helper()
  req:=httptest.NewRequest(method,path,strings.NewReader(body))
  if body!="" {req.Header.Set("Content-Type","application/json")}
  response,e:=app.Test(req);if e!=nil{t.Fatal(e)}
  defer response.Body.Close()
  data,e:=io.ReadAll(response.Body);if e!=nil{t.Fatal(e)}
  return response.StatusCode,data,response.Header.Clone()
 }
 return request,svc
}
func TestModule1ExtendedRealHTTPActivationThroughThreePostgresKills(t *testing.T){
 chaosAllowed(t)
 initial,db:=ownershipService(t)
 independentDrainPreviousDeliveries(t,initial)
 // Test genuine HTTP-seeded financial records, not a mocked ledger.
 fixture:=chaosFinanceSeed(t,initial,db)
 email:="http-3-real-crashes-"+uuid.NewString()+"@example.test"
 request,_:=extendedRealHTTP(t)
 code,_,h:=request("POST","/api/v1/auth/register",fmt.Sprintf("{\"email\":%q}",email))
 if code!=202||len(h.Values("Set-Cookie"))!=0 {t.Fatalf("REGISTRATION_HTTP_ACK_FAILED code=%d",code)}
 if n:=chaosCount(t,db,"SELECT count(*) FROM identity.users WHERE email_normalized=$1",email);n!=0{t.Fatalf("preproof user=%d",n)}
 if n:=chaosCount(t,db,"SELECT count(*) FROM identity.registration_intents WHERE email_normalized=$1",email);n!=1{t.Fatalf("precrash intent=%d",n)}
 // First real PostgreSQL SIGKILL after request is acknowledged.
 db1:=chaosRealCrashEvidence(t,db,1);defer db1.Close()
 chaosFinanceReconcile(t,db1,fixture,1)
 if n:=chaosCount(t,db1,"SELECT count(*) FROM identity.registration_intents WHERE email_normalized=$1",email);n!=1{t.Fatalf("postcrash intent=%d",n)}
 _,mailerSvc:=extendedRealHTTP(t)
 sink:=&ownershipMail{}
 delivered,e:=mailerSvc.DeliverOne(context.Background(),sink)
 if e!=nil||!delivered||sink.email!=email||sink.token=="" {t.Fatalf("postcrash delivery not completed found=%t recipient_match=%t err=%v",delivered,sink.email==email,e)}
 if n:=chaosCount(t,db1,"SELECT count(*) FROM identity.users WHERE email_normalized=$1",email);n!=0{t.Fatalf("preproof user after delivery=%d",n)}
 // Second real crash after durable proof issuance, before proof consumption.
 db2:=chaosRealCrashEvidence(t,db1,2);defer db2.Close()
 chaosFinanceReconcile(t,db2,fixture,2)
 verify,_:=extendedRealHTTP(t)
 payload,e:=json.Marshal(map[string]string{"token":sink.token,"password":"owner-selected-correct-horse","language":"en","theme":"system","timezone":"UTC"})
 if e!=nil{t.Fatal(e)}
 code,_,h=verify("POST","/api/v1/auth/email/verify",string(payload))
 if code!=200||len(h.Values("Set-Cookie"))!=0{t.Fatalf("postcrash activate HTTP status=%d cookies=%d",code,len(h.Values("Set-Cookie")))}
 if n:=chaosCount(t,db2,"SELECT count(*) FROM identity.users WHERE email_normalized=$1 AND account_state='active'",email);n!=1{t.Fatalf("activated users=%d",n)}
 if n:=chaosCount(t,db2,"SELECT count(*) FROM identity.user_email_verifications v JOIN identity.users u ON u.id=v.user_id WHERE u.email_normalized=$1 AND v.evidence_state='verified'",email);n!=1{t.Fatalf("verified evidence=%d",n)}
 if n:=chaosCount(t,db2,"SELECT count(*) FROM identity.sessions s JOIN identity.users u ON s.user_id=u.id WHERE u.email_normalized=$1",email);n!=0{t.Fatalf("activation minted session=%d",n)}
 // Third real crash after activation, before first normal HTTP login.
 db3:=chaosRealCrashEvidence(t,db2,3);defer db3.Close()
 chaosFinanceReconcile(t,db3,fixture,3)
 login,_:=extendedRealHTTP(t)
 code,_,h=login("POST","/api/v1/auth/login",fmt.Sprintf("{\"email\":%q,\"password\":\"owner-selected-correct-horse\"}",email))
 if code!=200||len(h.Values("Set-Cookie"))==0{t.Fatalf("postcrash login failed status=%d cookie_count=%d",code,len(h.Values("Set-Cookie")))}
 if n:=chaosCount(t,db3,"SELECT count(*) FROM identity.sessions s JOIN identity.users u ON s.user_id=u.id WHERE u.email_normalized=$1",email);n!=1{t.Fatalf("normal login sessions=%d",n)}
 code,_,_=login("POST","/api/v1/auth/email/verify",string(payload))
 if code==200{t.Fatal("consumed proof reused after real crash")}
 t.Log("THREE_REAL_POSTGRES_SIGKILLS=YES HTTP_REGISTER_202=PASS MAILER_PROOF_AFTER_RESTART=YES HTTP_ACTIVATION_200=PASS NO_SESSION_ON_ACTIVATION=YES HTTP_LOGIN_AFTER_RESTART_200=PASS TOKEN_REPLAY_REJECTED=YES FINANCIAL_LEDGER_AFTER_EACH_CRASH=IMMUTABLE")
}
