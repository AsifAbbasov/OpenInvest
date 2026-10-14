package postgres_test

// Financial ACK-loss evidence on real HTTP mutation and real PostgreSQL.
import (
 "context"
 "io"
 "net/http/httptest"
 "os"
 "strings"
 "testing"
 "time"

 "github.com/google/uuid"
 "github.com/openinvest/openinvest/backend-go/internal/auth"
 "github.com/openinvest/openinvest/backend-go/internal/httpapi"
 "github.com/openinvest/openinvest/backend-go/internal/postgres"
 "github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestModule1ExtendedFinancialHTTPLostCommitAckIdempotentRetry(t *testing.T){
 chaosAllowed(t)
 original,db:=ownershipService(t)
 fixture:=chaosFinanceSeed(t,original,db)
 direct,security:=chaosFinancialHTTP(t)
 ctx:=context.Background()
 ownerLogin,e:=security.Login(ctx,auth.LoginRequest{Email:fixture.ownerEmail,Password:fixture.password});if e!=nil{t.Fatal(e)}
 outsiderLogin,e:=security.Login(ctx,auth.LoginRequest{Email:fixture.outsiderEmail,Password:fixture.password});if e!=nil{t.Fatal(e)}
 token:=ownerLogin.Session.AccessToken
 var initial int
 e=db.QueryRow("SELECT count(*) FROM investment.transaction_entries WHERE portfolio_id=$1 AND transaction_type='DEPOSIT' AND gross_amount=4321.12345678",fixture.portfolioID).Scan(&initial)
 if e!=nil||initial!=0{t.Fatalf("invalid precontrol count=%d err=%v",initial,e)}
 proxy:=extendedStartAckProxy(t)
 store,e:=postgres.Open(proxy.databaseURL(t));if e!=nil{t.Fatal(e)}
 app,e:=httpapi.New(verticalslice.NewService(store,verticalslice.SystemClock{}),security,[]byte("financial-chaos-import-secret-32-bytes"))
 if e!=nil{_ =store.Close();t.Fatal(e)}
 body:="{\"transactionType\":\"DEPOSIT\",\"ticker\":null,\"quantity\":null,\"unitPrice\":null,\"grossAmount\":{\"amount\":\"4321.12345678\",\"currency\":\"RUB\"},\"commission\":{\"amount\":\"0.00000000\",\"currency\":\"RUB\"},\"tax\":{\"amount\":\"0.00000000\",\"currency\":\"RUB\"},\"tradeDate\":\"2026-01-11\",\"settlementDate\":\"2026-01-11\",\"note\":\"real lost COMMIT ack audit\"}"
 key:="finance-actual-ack-loss-"+uuid.NewString()
 path:="/api/v1/portfolios/"+fixture.portfolioID+"/transactions"
 req:=httptest.NewRequest("POST",path,strings.NewReader(body))
 req.Header.Set("Authorization","Bearer "+token)
 req.Header.Set("Content-Type","application/json")
 req.Header.Set("Idempotency-Key",key)
 resp,requestError:=app.Test(req,15000)
 if requestError==nil{
  _,_=io.Copy(io.Discard,resp.Body)
  _=resp.Body.Close()
  if resp.StatusCode==200||resp.StatusCode==201{t.Fatalf("FALSE_HTTP_SUCCESS_WITH_LOST_COMMIT_ACK status=%d",resp.StatusCode)}
 }
 _=store.Close()
 select{
 case <-proxy.intercepted:
 case <-time.After(5*time.Second):t.Fatal("NOT_EXERCISED: real financial COMMIT CommandComplete was not intercepted")
 }
 var durable int
 e=db.QueryRow("SELECT count(*) FROM investment.transaction_entries WHERE portfolio_id=$1 AND transaction_type='DEPOSIT' AND gross_amount=4321.12345678",fixture.portfolioID).Scan(&durable)
 if e!=nil||durable!=1{t.Fatalf("FINANCIAL_COMMIT_NOT_DURABLE_AFTER_LOST_ACK count=%d err=%v",durable,e)}
 code,replayed:=direct("POST",path,token,body,key)
 if code!=200&&code!=201{t.Fatalf("FINANCIAL_IDEMPOTENT_RETRY_FAILED status=%d response=%s",code,replayed)}
 var final int
 e=db.QueryRow("SELECT count(*) FROM investment.transaction_entries WHERE portfolio_id=$1 AND transaction_type='DEPOSIT' AND gross_amount=4321.12345678",fixture.portfolioID).Scan(&final)
 if e!=nil||final!=1{t.Fatalf("FINANCIAL_DUPLICATE_OR_LOST_AFTER_RETRY count=%d err=%v",final,e)}
 code,_=direct("GET",path,outsiderLogin.Session.AccessToken,"","")
 if code==200{t.Fatal("CROSS_PRINCIPAL_FINANCIAL_READ_ALLOWED_AFTER_ACK_LOSS")}
 t.Log("FINANCIAL_HTTP_WRITE=REAL_POSTGRES FINANCIAL_COMMIT_COMPLETE=INTERCEPTED CLIENT_RESPONSE=NOT_SUCCESSFUL FINANCIAL_ENTRY_DURABLE=YES SAME_IDEMPOTENCY_KEY_RETRY=SUCCESS DUPLICATE_LEDGER_ENTRIES=0 CROSS_PRINCIPAL_READ=DENIED")
}
