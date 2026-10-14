package postgres_test

// Independent, bounded NEW combinations: concurrent refresh reuse, mixed
// cross-principal CSRF, and equal idempotency keys on separate portfolios.
import (
 "context"
 "errors"
 "fmt"
 "encoding/json"
 "sync"
 "testing"

 "github.com/google/uuid"
 "github.com/openinvest/openinvest/backend-go/internal/auth"
)

func TestModule1CampaignSixteenWayRefreshReplayAndForeignCSRF(t *testing.T){
 chaosAllowed(t)
 service,db:=ownershipService(t)
 ctx:=context.Background()
 original:=ownerRequest()
 original.Email="campaign-refresh-"+uuid.NewString()+"@example.test"
 owner,e:=service.Register(ctx,original);if e!=nil{t.Fatal(e)}
 foreign:=ownerRequest()
 foreign.Email="campaign-other-"+uuid.NewString()+"@example.test"
 outsider,e:=service.Register(ctx,foreign);if e!=nil{t.Fatal(e)}
 if _,e=service.Refresh(ctx,owner.RefreshToken,outsider.Session.CSRFToken);!errors.Is(e,auth.ErrInvalidSession){
  t.Fatalf("FOREIGN_CSRF_ACCEPTED_OR_WRONG_ERROR: %v",e)
 }
 if _,e=service.Refresh(ctx,owner.RefreshToken,"");!errors.Is(e,auth.ErrInvalidCSRF){
  t.Fatalf("MISSING_CSRF_ACCEPTED_OR_WRONG_ERROR: %v",e)
 }
 const contenders=16
 start:=make(chan struct{})
 results:=make(chan struct{auth.AuthResult;error},contenders)
 var wg sync.WaitGroup
 for i:=0;i<contenders;i++{
  wg.Add(1);go func(){
   defer wg.Done();<-start
   result,e:=service.Refresh(ctx,owner.RefreshToken,owner.Session.CSRFToken)
   results<-struct{auth.AuthResult;error}{result,e}
  }()
 }
 close(start);wg.Wait();close(results)
 success:=0;var winner auth.AuthResult;invalid:=0
 for v:=range results{
  if v.error==nil{success++;winner=v.AuthResult;continue}
  if errors.Is(v.error,auth.ErrInvalidSession){invalid++;continue}
  t.Fatalf("UNEXPECTED_REFRESH_RACE_ERROR: %v",v.error)
 }
 if success!=1||invalid!=contenders-1{t.Fatalf("REFRESH_DOUBLE_ROTATION winners=%d invalid=%d",success,invalid)}
 if _,e=service.Refresh(ctx,owner.RefreshToken,owner.Session.CSRFToken);!errors.Is(e,auth.ErrInvalidSession){t.Fatalf("REPLAY_OF_ROTATED_REFRESH_ACCEPTED: %v",e)}
 revoked,e:=service.Logout(ctx,winner.RefreshToken,winner.Session.CSRFToken,true)
 if e!=nil||!revoked{t.Fatalf("winner logout failed: revoked=%t err=%v",revoked,e)}
 if _,e=service.Refresh(ctx,winner.RefreshToken,winner.Session.CSRFToken);!errors.Is(e,auth.ErrInvalidSession){t.Fatalf("REVOKED_DESCENDANT_ACCEPTED: %v",e)}
 var active int
 e=db.QueryRow("SELECT count(*) FROM identity.sessions WHERE user_id=$1 AND session_state='active'",owner.User.ID).Scan(&active)
 if e!=nil||active!=0{t.Fatalf("REFRESH_FAMILY_ACTIVE_AFTER_LOGOUT active=%d err=%v",active,e)}
 // A legitimate other principal's active session must remain functional.
 if _,e=service.Refresh(ctx,outsider.RefreshToken,outsider.Session.CSRFToken);e!=nil{t.Fatalf("FOREIGN_PRINCIPAL_SESSION_WRONGLY_REVOKED: %v",e)}
 t.Logf("REFRESH_CONCURRENT_OPERATIONS=%d SINGLE_WINNER=YES REPLAY_REJECTED=YES CSRF_CROSS_PRINCIPAL_DENIED=YES LOGOUT_REVOKES_DESCENDANT=YES OTHER_SESSION_PRESERVED=YES",contenders)
}

func TestModule1CampaignCrossPrincipalSameIdempotencyKeyCannotCrossLedger(t *testing.T){
 chaosAllowed(t)
 service,db:=ownershipService(t)
 fixture:=chaosFinanceSeed(t,service,db)
 run,security:=chaosFinancialHTTP(t)
 ctx:=context.Background()
 owner,e:=security.Login(ctx,auth.LoginRequest{Email:fixture.ownerEmail,Password:fixture.password});if e!=nil{t.Fatal(e)}
 other,e:=security.Login(ctx,auth.LoginRequest{Email:fixture.outsiderEmail,Password:fixture.password});if e!=nil{t.Fatal(e)}
 ownerToken:=owner.Session.AccessToken
 outsiderToken:=other.Session.AccessToken
 key:="campaign-same-key-"+uuid.NewString()
 ownerPath:="/api/v1/portfolios/"+fixture.portfolioID+"/transactions"
 deposit:=func(amount string)string{
  return fmt.Sprintf("{\"transactionType\":\"DEPOSIT\",\"ticker\":null,\"quantity\":null,\"unitPrice\":null,\"grossAmount\":{\"amount\":%q,\"currency\":\"RUB\"},\"commission\":{\"amount\":\"0.00000000\",\"currency\":\"RUB\"},\"tax\":{\"amount\":\"0.00000000\",\"currency\":\"RUB\"},\"tradeDate\":\"2026-02-10\",\"settlementDate\":null,\"note\":\"cross-principal idempotency campaign\"}",amount)
 }
 code,data:=run("POST",ownerPath,ownerToken,deposit("2345.67890123"),key)
 if code!=201{t.Fatalf("OWNER_POSITIVE_CONTROL_FAILED code=%d response=%s",code,data)}
 _,ownerBefore:=chaosFinanceStorage(t,db,fixture.portfolioID)
 code,_=run("POST",ownerPath,outsiderToken,deposit("9876.54321098"),key)
 if code==200||code==201{t.Fatal("FOREIGN_PRINCIPAL_REPLAYED_OWNER_FINANCE_OPERATION")}
 _,ownerAfter:=chaosFinanceStorage(t,db,fixture.portfolioID)
 if ownerBefore!=ownerAfter{t.Fatal("CROSS_PRINCIPAL_REPLAY_MUTATED_OWNER_LEDGER")}
 code,data=run("POST","/api/v1/portfolios",outsiderToken,`{"name":"outsider-campaign-portfolio","baseCurrency":"RUB"}`,"campaign-outsider-portfolio-"+uuid.NewString())
 if code!=201{t.Fatalf("OUTSIDER_PORTFOLIO_POSITIVE_CONTROL_FAILED status=%d response=%s",code,data)}
 var obj struct{Data struct{ID string `json:"id"`} `json:"data"`}
 if e=json.Unmarshal(data,&obj);e!=nil||obj.Data.ID==""{t.Fatalf("outsider portfolio parse: %v",e)}
 outsiderPath:="/api/v1/portfolios/"+obj.Data.ID+"/transactions"
 code,data=run("POST",outsiderPath,outsiderToken,deposit("4321.12345678"),key)
 if code!=201{t.Fatalf("CROSS_PRINCIPAL_SAME_KEY_WAS_NOT_SCOPED status=%d response=%s",code,data)}
 code,data=run("POST",outsiderPath,outsiderToken,deposit("4321.12345678"),key)
 if code!=200&&code!=201{t.Fatalf("OUTSIDER_IDEMPOTENT_RETRY_FAILED status=%d response=%s",code,data)}
 var n1,n2 int
 e=db.QueryRow("SELECT count(*) FROM investment.transaction_entries WHERE portfolio_id=$1 AND gross_amount=2345.67890123",fixture.portfolioID).Scan(&n1)
 if e!=nil||n1!=1{t.Fatalf("OWNER_LEDGER_REPLAY_COUNT count=%d err=%v",n1,e)}
 e=db.QueryRow("SELECT count(*) FROM investment.transaction_entries WHERE portfolio_id=$1 AND gross_amount=4321.12345678",obj.Data.ID).Scan(&n2)
 if e!=nil||n2!=1{t.Fatalf("OUTSIDER_LEDGER_REPLAY_COUNT count=%d err=%v",n2,e)}
 t.Log("SHARED_IDEMPOTENCY_KEY_ACROSS_PRINCIPALS=ISOLATED OWNED_PORTFOLIO_READ_WRITE=PASS FOREIGN_PORTFOLIO_WRITE=DENIED DUPLICATE_TRANSACTIONS=0 EXACT_DECIMALS=PASS")
}
