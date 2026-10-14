package postgres_test

// A financial HTTP append races session logout. Already-issued stateless JWTs
// may remain valid until expiry by documented policy; this test deliberately
// does NOT assume immediate access-token revocation.
import (
 "context"
 "errors"
 "fmt"
 "sync"
 "testing"

 "github.com/google/uuid"
 "github.com/openinvest/openinvest/backend-go/internal/auth"
)

func TestModule1CampaignFinancialAppendRacingRefreshSessionLogout(t *testing.T){
 chaosAllowed(t)
 initial,db:=ownershipService(t)
 fixture:=chaosFinanceSeed(t,initial,db)
 request,security:=chaosFinancialHTTP(t)
 ctx:=context.Background()
 path:="/api/v1/portfolios/"+fixture.portfolioID+"/transactions"
 for n:=1;n<=4;n++{
  login,e:=security.Login(ctx,auth.LoginRequest{Email:fixture.ownerEmail,Password:fixture.password})
  if e!=nil{t.Fatalf("session login round=%d: %v",n,e)}
  amount:=fmt.Sprintf("%d.12345678",8100+n)
  body:=fmt.Sprintf(`{"transactionType":"DEPOSIT","ticker":null,"quantity":null,"unitPrice":null,"grossAmount":{"amount":%q,"currency":"RUB"},"commission":{"amount":"0.00000000","currency":"RUB"},"tax":{"amount":"0.00000000","currency":"RUB"},"tradeDate":"2026-03-01","settlementDate":null,"note":"concurrent session revocation audit"}`,amount)
  key:="campaign-fin-logout-"+uuid.NewString()
  start:=make(chan struct{})
  var wg sync.WaitGroup;wg.Add(1)
  type result struct{revoked bool;err error}
  outcome:=make(chan result,1)
  go func(){defer wg.Done();<-start;ok,e:=security.Logout(ctx,login.RefreshToken,login.Session.CSRFToken,false);outcome<-result{ok,e}}()
  close(start)
  status,response:=request("POST",path,login.Session.AccessToken,body,key)
  wg.Wait()
  finish:=<-outcome
  if finish.err!=nil||!finish.revoked{t.Fatalf("logout round=%d revoked=%t err=%v",n,finish.revoked,finish.err)}
  if status!=201&&status!=200{t.Fatalf("financial append unexpected status round=%d code=%d response=%s",n,status,response)}
  if _,e=security.Refresh(ctx,login.RefreshToken,login.Session.CSRFToken);!errors.Is(e,auth.ErrInvalidSession){
   t.Fatalf("revoked refresh still valid round=%d err=%v",n,e)
  }
  var count int
  e=db.QueryRow("SELECT count(*) FROM investment.transaction_entries WHERE portfolio_id=$1 AND transaction_type='DEPOSIT' AND gross_amount=$2::numeric",fixture.portfolioID,amount).Scan(&count)
  if e!=nil||count!=1{t.Fatalf("ACKNOWLEDGED_FINANCE_WRITE_LOST_OR_DUPLICATED round=%d count=%d err=%v",n,count,e)}
  newLogin,e:=security.Login(ctx,auth.LoginRequest{Email:fixture.ownerEmail,Password:fixture.password})
  if e!=nil{t.Fatalf("normal login after logout round=%d: %v",n,e)}
  code,payload:=request("POST",path,newLogin.Session.AccessToken,body,key)
  if code!=200&&code!=201{t.Fatalf("financial replay after logout failed round=%d status=%d response=%s",n,code,payload)}
  e=db.QueryRow("SELECT count(*) FROM investment.transaction_entries WHERE portfolio_id=$1 AND transaction_type='DEPOSIT' AND gross_amount=$2::numeric",fixture.portfolioID,amount).Scan(&count)
  if e!=nil||count!=1{t.Fatalf("FINANCIAL_DOUBLE_POST_AFTER_LOGOUT round=%d count=%d err=%v",n,count,e)}
  t.Logf("FINANCIAL_LOGOUT_RACE_ROUND=%d ACKNOWLEDGED_WRITE=1 REFRESH_REVOKED=YES NEW_LOGIN=PASS REPLAY_SAME_IDEMPOTENCY_KEY=PASS DUPLICATE_LEDGER_ENTRIES=0",n)
 }
}
