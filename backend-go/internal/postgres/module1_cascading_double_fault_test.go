package postgres_test

// Independent double-fault experiments run ONLY on disposable GitHub Actions PostgreSQL.
// They intentionally chain distinct failure modes rather than replay earlier single crashes.
import (
 "context"
 "fmt"
 "io"
 "net/http/httptest"
 "os"
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

func cascadeDeposit(amount,note string)string {
 return fmt.Sprintf(`{"transactionType":"DEPOSIT","ticker":null,"quantity":null,"unitPrice":null,"grossAmount":{"amount":%q,"currency":"RUB"},"commission":{"amount":"0.00000000","currency":"RUB"},"tax":{"amount":"0.00000000","currency":"RUB"},"tradeDate":"2026-03-11","settlementDate":"2026-03-11","note":%q}`,amount,note)
}

// Durable server COMMIT + lost wire acknowledgment + actual postgres SIGKILL +
// WAL recovery + logical request retry. The earlier ACK and SIGKILL tests were separate.
func TestModule1CascadeLostFinancialAckAndPhysicalCrashSameOperation(t *testing.T){
 chaosAllowed(t)
 original,db:=ownershipService(t)
 f:=chaosFinanceSeed(t,original,db)
 direct,authService:=chaosFinancialHTTP(t)
 owner,err:=authService.Login(context.Background(),auth.LoginRequest{Email:f.ownerEmail,Password:f.password})
 if err!=nil{t.Fatal(err)}
 outsider,err:=authService.Login(context.Background(),auth.LoginRequest{Email:f.outsiderEmail,Password:f.password})
 if err!=nil{t.Fatal(err)}
 body:=cascadeDeposit("9271.12345678","double-fault-ack-"+uuid.NewString())
 key:="cascade-ack-"+uuid.NewString()
 path:="/api/v1/portfolios/"+f.portfolioID+"/transactions"
 proxy:=extendedStartAckProxy(t)
 store,err:=postgres.Open(proxy.databaseURL(t));if err!=nil{t.Fatal(err)}
 app,err:=httpapi.New(verticalslice.NewService(store,verticalslice.SystemClock{}),authService,[]byte("financial-chaos-import-secret-32-bytes"))
 if err!=nil{_ =store.Close();t.Fatal(err)}
 request:=httptest.NewRequest("POST",path,strings.NewReader(body))
 request.Header.Set("Authorization","Bearer "+owner.Session.AccessToken)
 request.Header.Set("Content-Type","application/json")
 request.Header.Set("Idempotency-Key",key)
 response,e:=app.Test(request)
 if e==nil {
  _,_=io.Copy(io.Discard,response.Body);_=response.Body.Close()
  if response.StatusCode==200||response.StatusCode==201{t.Fatalf("ACK_INTERCEPT_FAILURE: success returned status=%d",response.StatusCode)}
 }
 _=store.Close()
 select{case <-proxy.intercepted:case <-time.After(5*time.Second):t.Fatal("NOT_EXERCISED: actual COMMIT CommandComplete not intercepted")}
 select{case <-proxy.stopped:case <-time.After(5*time.Second):t.Fatal("proxy did not exit")}
 before:=chaosCount(t,db,"SELECT count(*) FROM investment.transaction_entries WHERE portfolio_id=$1 AND gross_amount=9271.12345678 AND transaction_type='DEPOSIT'",f.portfolioID)
 if before!=1{t.Fatalf("NEGATIVE_CONTROL: durable COMMIT missing/duplicated count=%d",before)}
 f.portfolioJSON,f.entriesJSON=chaosFinanceStorage(t,db,f.portfolioID)
 recovered:=chaosRealCrashEvidence(t,db,1);defer recovered.Close()
 chaosFinanceReconcile(t,recovered,f,1)
 retryAuth,security:=chaosFinancialHTTP(t)
 freshOwner,err:=security.Login(context.Background(),auth.LoginRequest{Email:f.ownerEmail,Password:f.password})
 if err!=nil{t.Fatal(err)}
 code,res:=retryAuth("POST",path,freshOwner.Session.AccessToken,body,key)
 if code!=200 && code!=201{t.Fatalf("RETRY_AFTER_ACK_LOSS_AND_SIGKILL_FAILED status=%d body=%s",code,res)}
 after:=chaosCount(t,recovered,"SELECT count(*) FROM investment.transaction_entries WHERE portfolio_id=$1 AND gross_amount=9271.12345678 AND transaction_type='DEPOSIT'",f.portfolioID)
 if after!=1{t.Fatalf("FINANCIAL_DOUBLE_FAULT_DUPLICATION count=%d",after)}
 _,entries:=chaosFinanceStorage(t,recovered,f.portfolioID)
 if entries!=f.entriesJSON{t.Fatal("CANONICAL_LEDGER_CHANGED_AFTER_IDENTICAL_REQUEST_RETRY")}
 code,_=retryAuth("POST",path,outsider.Session.AccessToken,body,key)
 if code==200||code==201{t.Fatal("CROSS_PRINCIPAL_REUSE_OF_OWNER_IDEMPOTENCY_KEY_SUCCEEDED")}
 _,entries=chaosFinanceStorage(t,recovered,f.portfolioID)
 if entries!=f.entriesJSON{t.Fatal("OUTSIDER_CHANGED_OWNER_LEDGER")}
 t.Log("DOUBLE_FAULT=REAL_COMMIT_ACK_WITHHELD_THEN_REAL_POSTMASTER_SIGKILL PHYSICAL_WAL_RECOVERY=YES SERVER_COMMITTED_ENTRY_PRESERVED=YES IDENTICAL_IDEMPOTENCY_KEY_RETRY=EXACTLY_ONCE FOREIGN_PRINCIPAL=DENIED")
}

type cascadeResponse struct { index int; status int; err error }

// Twelve true concurrent HTTP writes are parked in PostgreSQL on an
// ACCESS EXCLUSIVE lock before crash. Their outcomes are reconciled through
// independent SQL, and identical keys are retried after WAL recovery.
func TestModule1CascadeTwelveBlockedFinancialWritesAndRealSIGKILL(t *testing.T){
 chaosAllowed(t)
 original,db:=ownershipService(t)
 f:=chaosFinanceSeed(t,original,db)
 store,err:=postgres.Open(os.Getenv("OPENINVEST_DATABASE_TEST_URL"));if err!=nil{t.Fatal(err)}
 defer store.Close()
 security,err:=auth.NewService(store,nil,auth.Config{AccessTokenSecret:[]byte("financial-chaos-audit-secret-32-bytes")})
 if err!=nil{t.Fatal(err)}
 owner,err:=security.Login(context.Background(),auth.LoginRequest{Email:f.ownerEmail,Password:f.password});if err!=nil{t.Fatal(err)}
 app,err:=httpapi.New(verticalslice.NewService(store,verticalslice.SystemClock{}),security,[]byte("financial-chaos-import-secret-32-bytes"))
 if err!=nil{t.Fatal(err)}
 const n=12
 prefix:="cascade-wave-"+uuid.NewString()
 path:="/api/v1/portfolios/"+f.portfolioID+"/transactions"
 var keys [n]string
 var bodies [n]string
 for i:=0;i<n;i++{
  keys[i]="cascade-wave-key-"+uuid.NewString()
  bodies[i]=cascadeDeposit(fmt.Sprintf("201.%08d",i+1),fmt.Sprintf("%s-%02d",prefix,i))
 }
 blocker,err:=db.BeginTx(context.Background(),nil);if err!=nil{t.Fatal(err)}
 if _,err=blocker.Exec("LOCK TABLE investment.portfolios IN ACCESS EXCLUSIVE MODE");err!=nil{_ =blocker.Rollback();t.Fatal(err)}
 var wg sync.WaitGroup
 outcomes:=make(chan cascadeResponse,n)
 launch:=make(chan struct{})
 for i:=0;i<n;i++{
  wg.Add(1)
  go func(index int){
   defer wg.Done();<-launch
   req:=httptest.NewRequest("POST",path,strings.NewReader(bodies[index]))
   req.Header.Set("Authorization","Bearer "+owner.Session.AccessToken)
   req.Header.Set("Content-Type","application/json")
   req.Header.Set("Idempotency-Key",keys[index])
   response,e:=app.Test(req)
   if e!=nil{outcomes<-cascadeResponse{index:index,err:e};return}
   _,_=io.Copy(io.Discard,response.Body);_=response.Body.Close()
   outcomes<-cascadeResponse{index:index,status:response.StatusCode}
  }(i)
 }
 close(launch)
 // Positive injection oracle: real PostgreSQL sessions must be waiting
 // on a relation lock, not merely goroutines waiting in Go.
 observed:=false
 ctx,cancel:=context.WithTimeout(context.Background(),12*time.Second)
 defer cancel()
 for ctx.Err()==nil{
  var active int
  e=db.QueryRowContext(ctx,`SELECT count(*) FROM pg_stat_activity WHERE datname=current_database() AND wait_event_type='Lock' AND state='active' AND pid<>pg_backend_pid()`).Scan(&active)
  if e==nil&&active>=2{observed=true;t.Logf("PG_BLOCKED_ACTIVE_FINANCIAL_WRITERS=%d",active);break}
  time.Sleep(100*time.Millisecond)
 }
 if !observed{_=blocker.Rollback();t.Fatal("FAULT_NOT_INJECTED: multiple backend sessions blocked in SQL were not observed")}
 // Killing the single verified postgres:18 service also rolls back the lock holder.
 // No statement is made about each individual HTTP response while DB is unavailable.
 recovered:=chaosRealCrashEvidence(t,db,1);defer recovered.Close()
 // The lock-holding transaction was severed by the real postmaster SIGKILL.
 _=blocker.Rollback()
 drained:=make(chan struct{})
 go func(){wg.Wait();close(drained)}()
 select{case <-drained:case <-time.After(12*time.Second):t.Fatal("IN_FLIGHT_HTTP_REQUESTS_DID_NOT_SETTLE_AFTER_DATABASE_CRASH")}
 close(outcomes)
 failedOrAmbiguous:=0
 for result:=range outcomes{
  if result.err!=nil||result.status!=201 {failedOrAmbiguous++}
 }
 // The original operations may or may not be durable after reconnection:
 // only the independent DB snapshot decides. No duplicate operation is acceptable.
 freshHTTP,freshSecurity:=chaosFinancialHTTP(t)
 freshOwner,err:=freshSecurity.Login(context.Background(),auth.LoginRequest{Email:f.ownerEmail,Password:f.password});if err!=nil{t.Fatal(err)}
 for i:=0;i<n;i++{
  code,body:=freshHTTP("POST",path,freshOwner.Session.AccessToken,bodies[i],keys[i])
  if code!=200&&code!=201{t.Fatalf("RECOVERY_RETRY_REJECTED index=%d status=%d response=%s",i,code,body)}
 }
 var total,distinct int
 if err:=recovered.QueryRow(`SELECT count(*),count(DISTINCT note) FROM investment.transaction_entries WHERE portfolio_id=$1 AND note LIKE $2`,f.portfolioID,prefix+"%").Scan(&total,&distinct);err!=nil{t.Fatal(err)}
 if total!=n||distinct!=n{t.Fatalf("DUPLICATE_OR_LOST_OPERATIONS_AFTER_CRASH total=%d distinct=%d expected=%d",total,distinct,n)}
 for i:=0;i<n;i++{
  if x:=chaosCount(t,recovered,"SELECT count(*) FROM investment.transaction_entries WHERE portfolio_id=$1 AND note=$2 AND gross_amount=$3::numeric",f.portfolioID,fmt.Sprintf("%s-%02d",prefix,i),fmt.Sprintf("201.%08d",i+1));x!=1{
   t.Fatalf("DECIMAL_OR_IDENTITY_CORRUPTION case=%d count=%d",i,x)
  }
 }
 outsider,err:=freshSecurity.Login(context.Background(),auth.LoginRequest{Email:f.outsiderEmail,Password:f.password})
 if err!=nil{t.Fatal(err)}
 code,_:=freshHTTP("GET",path,outsider.Session.AccessToken,"","")
 if code==200{t.Fatal("CROSS_PRINCIPAL_LEDGER_READ_AFTER_MULTIPLE_INFLIGHT_CRASH")}
 t.Logf("CASCADE=12_REAL_HTTP_REQUESTS_BLOCKED_AT_POSTGRESQL_LOCK THEN_REAL_POSTMASTER_SIGKILL=YES WAL_RECOVERY=YES ORIGINAL_UNSUCCESSFUL_OR_AMBIGUOUS=%d RETRIES_ALL_SUCCESS=YES EXACT_DECIMAL_LEDGER_ROWS=%d DUPLICATES=0 CROSS_PRINCIPAL_READ=DENIED",failedOrAmbiguous,total)
}
