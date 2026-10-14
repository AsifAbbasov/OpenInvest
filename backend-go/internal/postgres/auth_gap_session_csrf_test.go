package postgres_test

// Negative CSRF/cookie/session matrix against real HTTP handlers + real PG18.
import(
 "context"
 "encoding/json"
 "fmt"
 "math/rand"
 "time"
 "io"
 "net/http"
 "net/http/httptest"
 "strings"
 "sync"
 "testing"

 "github.com/google/uuid"
 "github.com/openinvest/openinvest/backend-go/internal/auth"
 "github.com/openinvest/openinvest/backend-go/internal/httpapi"
 "github.com/openinvest/openinvest/backend-go/internal/postgres"
 "github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestAUTHGAPRealPostgresCSRFMatrixAndSessionIsolation(t *testing.T){
 chaosAllowed(t)
 svc,db:=ownershipService(t)
 f:=chaosFinanceSeed(t,svc,db)
 store,e:=postgres.Open(getOwnershipURL(t));if e!=nil{t.Fatal(e)};defer store.Close()
 security,e:=auth.NewService(store,nil,auth.Config{AccessTokenSecret:[]byte("financial-chaos-audit-secret-32-bytes")});if e!=nil{t.Fatal(e)}
 app,e:=httpapi.New(verticalslice.NewService(store,verticalslice.SystemClock{}),security,[]byte("financial-chaos-import-secret-32-bytes"));if e!=nil{t.Fatal(e)}
 owner,e:=security.Login(context.Background(),auth.LoginRequest{Email:f.ownerEmail,Password:f.password});if e!=nil{t.Fatal(e)}
 other,e:=security.Login(context.Background(),auth.LoginRequest{Email:f.outsiderEmail,Password:f.password});if e!=nil{t.Fatal(e)}
 makeReq:=func(path,cookie,csrf,body string)(int,[]byte,error){
  req:=httptest.NewRequest("POST",path,strings.NewReader(body))
  req.Header.Set("Content-Type","application/json")
  if cookie!=""{req.AddCookie(&http.Cookie{Name:auth.RefreshCookieName,Value:cookie})}
  if csrf!=""{req.Header.Set("X-CSRF-Token",csrf)}
  res,err:=app.Test(req);if err!=nil{return 0,nil,err}
  defer res.Body.Close()
  raw,err:=io.ReadAll(res.Body);return res.StatusCode,raw,err
 }
 cases:=[]struct{name,path,cookie,csrf,body string}{
  {"refresh-missing-cookie","/api/v1/auth/refresh","",owner.Session.CSRFToken,""},
  {"refresh-missing-csrf","/api/v1/auth/refresh",owner.RefreshToken,"",""},
  {"refresh-foreign-csrf","/api/v1/auth/refresh",owner.RefreshToken,other.Session.CSRFToken,""},
  {"refresh-foreign-cookie","/api/v1/auth/refresh",other.RefreshToken,owner.Session.CSRFToken,""},
  {"refresh-malformed-cookie","/api/v1/auth/refresh","malformed-fixture-not-a-session",owner.Session.CSRFToken,""},
  {"logout-missing-cookie","/api/v1/auth/logout","",owner.Session.CSRFToken,`{"allSessions":false}`},
  {"logout-missing-csrf","/api/v1/auth/logout",owner.RefreshToken,"",`{"allSessions":false}`},
  {"logout-foreign-csrf","/api/v1/auth/logout",owner.RefreshToken,other.Session.CSRFToken,`{"allSessions":false}`},
  {"logout-foreign-cookie","/api/v1/auth/logout",other.RefreshToken,owner.Session.CSRFToken,`{"allSessions":true}`},
  {"logout-malformed-cookie","/api/v1/auth/logout","malformed-fixture-not-a-session",owner.Session.CSRFToken,`{"allSessions":false}`},
  {"logout-missing-contract-body","/api/v1/auth/logout",owner.RefreshToken,owner.Session.CSRFToken,""},
 }
 for _,c:=range cases{
  t.Run(c.name,func(t *testing.T){
   status,raw,err:=makeReq(c.path,c.cookie,c.csrf,c.body)
   if err!=nil{t.Fatalf("HTTP failed: %v",err)}
   if status>=200&&status<300{t.Fatalf("CROSS_SESSION_CSRF_OR_COOKIE_BYPASS status=%d response=%s",status,raw)}
  })
 }
 status,raw,err:=makeReq("/api/v1/auth/refresh",owner.RefreshToken,owner.Session.CSRFToken,"")
 if err!=nil||status!=200{t.Fatalf("LEGITIMATE_REFRESH_DESTROYED_BY_NEGATIVE_PROBES status=%d error=%v body=%s",status,err,raw)}
 var response struct{Data struct{CSRFToken string `json:"csrfToken"`} `json:"data"`}
 if e:=json.Unmarshal(raw,&response);e!=nil{t.Fatal(e)}
 status,_,err=makeReq("/api/v1/auth/refresh",owner.RefreshToken,owner.Session.CSRFToken,"")
 if err!=nil||status>=200&&status<300{t.Fatal("ROTATED_REFRESH_TOKEN_REPLAY_ACCEPTED")}
 status,_,err=makeReq("/api/v1/auth/refresh",other.RefreshToken,other.Session.CSRFToken,"")
 if err!=nil||status!=200{t.Fatalf("OTHER_PRINCIPAL_SESSION_CORRUPTED status=%d err=%v",status,err)}
 t.Logf("POSTGRES_AUTH_CSRF_NEGATIVE_CASES=%d ALL_REJECTED=YES LEGITIMATE_REFRESH=PASS OLD_REFRESH_REPLAY=DENIED INDEPENDENT_PRINCIPAL_SESSION=PASS",len(cases))
}

func TestAUTHGAPRefreshLogoutRacingFinancialWrites(t *testing.T){
 chaosAllowed(t)
 seed,db:=ownershipService(t)
 f:=chaosFinanceSeed(t,seed,db)
 store,e:=postgres.Open(getOwnershipURL(t));if e!=nil{t.Fatal(e)};defer store.Close()
 security,e:=auth.NewService(store,nil,auth.Config{AccessTokenSecret:[]byte("financial-chaos-audit-secret-32-bytes")});if e!=nil{t.Fatal(e)}
 app,e:=httpapi.New(verticalslice.NewService(store,verticalslice.SystemClock{}),security,[]byte("financial-chaos-import-secret-32-bytes"));if e!=nil{t.Fatal(e)}
 owner,e:=security.Login(context.Background(),auth.LoginRequest{Email:f.ownerEmail,Password:f.password});if e!=nil{t.Fatal(e)}
 outsider,e:=security.Login(context.Background(),auth.LoginRequest{Email:f.outsiderEmail,Password:f.password});if e!=nil{t.Fatal(e)}
 path:="/api/v1/portfolios/"+f.portfolioID+"/transactions"
 const writes=12
 type result struct{label string;status int;err error}
 results:=make(chan result,writes+2)
 start:=make(chan struct{})
 var wg sync.WaitGroup
 do:=func(label,method,path,body,token,cookie,csrf,key string){
  defer wg.Done();<-start
  req:=httptest.NewRequest(method,path,strings.NewReader(body))
  if body!=""{req.Header.Set("Content-Type","application/json")}
  if token!=""{req.Header.Set("Authorization","Bearer "+token)}
  if cookie!=""{req.AddCookie(&http.Cookie{Name:auth.RefreshCookieName,Value:cookie})}
  if csrf!=""{req.Header.Set("X-CSRF-Token",csrf)}
  if key!=""{req.Header.Set("Idempotency-Key",key)}
  resp,err:=app.Test(req)
  if err!=nil{results<-result{label:label,err:err};return}
  _,_=io.Copy(io.Discard,resp.Body);_=resp.Body.Close()
  results<-result{label:label,status:resp.StatusCode}
 }
 prefix:="authgap-session-race-"+uuid.NewString()
 for i:=0;i<writes;i++{
  amount:=""+strings.Repeat("0",0)
  amount=string([]byte{'1','0','0'})+"."+strings.Repeat("0",7)+string('0'+rune(i%10))
  body:=`{"transactionType":"DEPOSIT","ticker":null,"quantity":null,"unitPrice":null,"grossAmount":{"amount":"`+amount+`","currency":"RUB"},"commission":{"amount":"0.00000000","currency":"RUB"},"tax":{"amount":"0.00000000","currency":"RUB"},"tradeDate":"2026-08-13","settlementDate":"2026-08-13","note":"`+prefix+"-"+uuid.NewString()+`"}`
  wg.Add(1);go do("write","POST",path,body,owner.Session.AccessToken,"","","race-"+uuid.NewString())
 }
 wg.Add(2)
 go do("refresh","POST","/api/v1/auth/refresh","", "",owner.RefreshToken,owner.Session.CSRFToken,"")
 go do("logout","POST","/api/v1/auth/logout",`{"allSessions":false}`,"",owner.RefreshToken,owner.Session.CSRFToken,"")
 close(start);wg.Wait();close(results)
 succeeded:=0
 for x:=range results{
  if x.err!=nil{t.Fatalf("HTTP_TRANSPORT_ERROR label=%s err=%v",x.label,x.err)}
  if x.label=="write" && x.status==201{succeeded++}
  if x.label=="write"&&x.status>=500{t.Fatalf("FINANCIAL_WRITE_SERVER_ERROR status=%d",x.status)}
 }
 n:=chaosCount(t,db,"SELECT count(*) FROM investment.transaction_entries WHERE portfolio_id=$1 AND note LIKE $2",f.portfolioID,prefix+"%")
 if n!=succeeded{t.Fatalf("ACKED_FINANCIAL_ROWS_MISMATCH successes=%d durable=%d",succeeded,n)}
 run,_:=chaosFinancialHTTP(t)
 if status,_:=run("GET",path,outsider.Session.AccessToken,"","");status==200{t.Fatal("FOREIGN_LEDGER_READ_AFTER_SESSION_RACE")}
 if status,_:=run("POST",path,outsider.Session.AccessToken,`{"transactionType":"DEPOSIT","grossAmount":{"amount":"1.00000000","currency":"RUB"},"commission":{"amount":"0.00000000","currency":"RUB"},"tax":{"amount":"0.00000000","currency":"RUB"},"tradeDate":"2026-08-14","settlementDate":"2026-08-14"}`,"foreign-"+uuid.NewString());status==201{t.Fatal("FOREIGN_LEDGER_WRITE_AFTER_SESSION_RACE")}
 t.Logf("POSTGRES_SESSION_FINANCE_RACE_OPS=14 HTTP_ACKED_WRITES=%d INDEPENDENT_DURABLE_ROWS=%d FOREIGN_READ_WRITE=DENIED ACCESS_TOKEN_IS_STATELESS_NO_IMMEDIATE_REVOCATION_ASSUMED",succeeded,n)
}


// Eight distinct randomized launch orders, each with ten competing HTTP
// commands: 8 exact idempotent financial writes + refresh + logout.
// Revalidate storage after each schedule; not an exhaustive scheduler proof.
func TestAUTHGAPEightSessionFinanceIdempotencySchedules(t *testing.T){
 chaosAllowed(t)
 seed,db:=ownershipService(t)
 f:=chaosFinanceSeed(t,seed,db)
 store,err:=postgres.Open(getOwnershipURL(t));if err!=nil{t.Fatal(err)};defer store.Close()
 security,err:=auth.NewService(store,nil,auth.Config{AccessTokenSecret:[]byte("financial-chaos-audit-secret-32-bytes")});if err!=nil{t.Fatal(err)}
 app,err:=httpapi.New(verticalslice.NewService(store,verticalslice.SystemClock{}),security,[]byte("financial-chaos-import-secret-32-bytes"));if err!=nil{t.Fatal(err)}
 const rounds=8
 const writers=8
 const seedValue int64=26101508
 rng:=rand.New(rand.NewSource(seedValue))
 for round:=0;round<rounds;round++{
  owner,err:=security.Login(context.Background(),auth.LoginRequest{Email:f.ownerEmail,Password:f.password});if err!=nil{t.Fatal(err)}
  label:=fmt.Sprintf("authgap-multischedule-%d-%s",round,uuid.NewString())
  key:="authgap-exact-idempotency-"+uuid.NewString()
  amount:=fmt.Sprintf("%d.00000000",300+round)
  body:=fmt.Sprintf(`{"transactionType":"DEPOSIT","ticker":null,"quantity":null,"unitPrice":null,"grossAmount":{"amount":%q,"currency":"RUB"},"commission":{"amount":"0.00000000","currency":"RUB"},"tax":{"amount":"0.00000000","currency":"RUB"},"tradeDate":"2026-08-16","settlementDate":"2026-08-16","note":%q}`,amount,label)
  path:="/api/v1/portfolios/"+f.portfolioID+"/transactions"
  type work struct {method,path,body,key,token,cookie,csrf,kind string}
  jobs:=make([]work,0,writers+2)
  for i:=0;i<writers;i++{jobs=append(jobs,work{method:"POST",path:path,body:body,key:key,token:owner.Session.AccessToken,kind:"financial"})}
  jobs=append(jobs,work{method:"POST",path:"/api/v1/auth/refresh",cookie:owner.RefreshToken,csrf:owner.Session.CSRFToken,kind:"refresh"})
  jobs=append(jobs,work{method:"POST",path:"/api/v1/auth/logout",body:`{"allSessions":false}`,cookie:owner.RefreshToken,csrf:owner.Session.CSRFToken,kind:"logout"})
  rng.Shuffle(len(jobs),func(i,j int){jobs[i],jobs[j]=jobs[j],jobs[i]})
  results:=make(chan struct{kind string;status int;err error},len(jobs))
  start:=make(chan struct{})
  var wg sync.WaitGroup
  for i,j:=range jobs{
   wg.Add(1)
   go func(index int,w work){defer wg.Done();<-start
    // Bounded launch skew intentionally explores different interleavings.
    time.Sleep(time.Duration(index%4)*time.Millisecond)
    req:=httptest.NewRequest(w.method,w.path,strings.NewReader(w.body))
    if w.body!=""{req.Header.Set("Content-Type","application/json")}
    if w.token!=""{req.Header.Set("Authorization","Bearer "+w.token)}
    if w.key!=""{req.Header.Set("Idempotency-Key",w.key)}
    if w.cookie!=""{req.AddCookie(&http.Cookie{Name:auth.RefreshCookieName,Value:w.cookie})}
    if w.csrf!=""{req.Header.Set("X-CSRF-Token",w.csrf)}
    res,e:=app.Test(req)
    if e!=nil{results<-struct{kind string;status int;err error}{w.kind,0,e};return}
    _,_=io.Copy(io.Discard,res.Body);_=res.Body.Close()
    results<-struct{kind string;status int;err error}{w.kind,res.StatusCode,nil}
   }(i,j)
  }
  close(start);wg.Wait();close(results)
  positive:=0
  for x:=range results{
   if x.err!=nil{t.Fatalf("REAL_HTTP_FAILURE round=%d kind=%s err=%v",round,x.kind,x.err)}
   if x.kind=="financial"{
    if x.status==201{positive++}
    if x.status>=500{t.Fatalf("FINANCIAL_RACE_SERVER_FAILURE round=%d status=%d",round,x.status)}
   }
  }
  if positive==0{t.Fatalf("NO_AUTHORIZED_FINANCIAL_POSITIVE_CONTROL round=%d",round)}
  count:=chaosCount(t,db,"SELECT count(*) FROM investment.transaction_entries WHERE portfolio_id=$1 AND note=$2",f.portfolioID,label)
  if count!=1{t.Fatalf("DUPLICATE_OR_LOST_IDENTICAL_CONCURRENT_FINANCIAL_OPERATION round=%d count=%d successes=%d",round,count,positive)}
  t.Logf("INDEPENDENT_SCHEDULE=%d CONCURRENT_OPS=%d EXACT_KEY_ACK_RESPONSES=%d DURABLE_FINANCIAL_ROWS=1",round,len(jobs),positive)
 }
 t.Logf("AUTH_GAP_SESSION_IDEMPOTENCY_LAUNCH_ORDERS=%d TOTAL_CONCURRENT_HTTP_OPERATIONS=%d FIXED_SEED=%d ALL_DURABLE_EXACTLY_ONCE=YES",rounds,rounds*(writers+2),seedValue)
}
