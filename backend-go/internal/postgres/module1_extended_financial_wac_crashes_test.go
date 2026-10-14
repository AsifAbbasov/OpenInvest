package postgres_test

// Real money, buy/sell, WAC, idempotency and ownership across repeated WAL replay.
import (
 "context"
 "encoding/json"
 "fmt"
 "testing"

 "github.com/google/uuid"
 "github.com/openinvest/openinvest/backend-go/internal/auth"
)

type extendedPositionEnvelope struct{
 Data struct {
  Items []struct{
   Ticker string `json:"ticker"`
   Quantity string `json:"quantity"`
   WeightedAverageCost struct{
    Amount string `json:"amount"`
   } `json:"weightedAverageCost"`
  } `json:"items"`
 } `json:"data"`
}
func extendedAssertSBERPosition(t *testing.T,body []byte){
 t.Helper()
 var response extendedPositionEnvelope
 if e:=json.Unmarshal(body,&response);e!=nil{t.Fatalf("positions JSON invalid: %v",e)}
 matched:=0
 for _,p:=range response.Data.Items{
  if p.Ticker!="SBER"{continue}
  matched++
  if p.Quantity!="15.00000000"||p.WeightedAverageCost.Amount!="150.00000000"{
   t.Fatalf("WAC_POSITION_DRIFT quantity=%q WAC=%q",p.Quantity,p.WeightedAverageCost.Amount)
  }
 }
 if matched!=1{t.Fatalf("expected one SBER position, found %d",matched)}
}
func TestModule1ExtendedFinancialTradesWACAcrossThreeRealCrashes(t *testing.T){
 chaosAllowed(t)
 svc,db:=ownershipService(t)
 f:=chaosFinanceSeed(t,svc,db)
 do,security:=chaosFinancialHTTP(t)
 login,e:=security.Login(context.Background(),auth.LoginRequest{Email:f.ownerEmail,Password:f.password});if e!=nil{t.Fatal(e)}
 token:=login.Session.AccessToken
 base:="/api/v1/portfolios/"+f.portfolioID
 transaction:=func(kind,quantity,price,tradeDate string)string{
  return fmt.Sprintf(`{"transactionType":%q,"ticker":"SBER","quantity":%q,"unitPrice":{"amount":%q,"currency":"RUB"},"grossAmount":null,"commission":{"amount":"0.00000000","currency":"RUB"},"tax":{"amount":"0.00000000","currency":"RUB"},"tradeDate":%q,"settlementDate":null,"note":"WAC crash test"}`,kind,quantity,price,tradeDate)
 }
 scenarios:=[]struct{kind,qty,price,date string}{
  {"BUY","10.00000000","100.00000000","2026-02-01"},
  {"BUY","10.00000000","200.00000000","2026-02-02"},
  {"SELL","5.00000000","150.00000000","2026-02-03"},
 }
 var lastKey,lastBody string
 for i,s:=range scenarios{
  body:=transaction(s.kind,s.qty,s.price,s.date)
  key:="finance-wac-real-crash-"+uuid.NewString()
  code,response:=do("POST",base+"/transactions",token,body,key)
  if code!=201{t.Fatalf("REAL_FINANCIAL_TRANSACTION_REJECTED index=%d type=%s status=%d body=%s",i,s.kind,code,response)}
  if i==len(scenarios)-1{lastKey,lastBody=key,body}
 }
 path:=base+"/positions?asOfDate=2026-02-03"
 code,response:=do("GET",path,token,"","")
 if code!=200{t.Fatalf("baseline positions not readable status=%d",code)}
 extendedAssertSBERPosition(t,response)
 f.portfolioJSON,f.entriesJSON=chaosFinanceStorage(t,db,f.portfolioID)
 if n:=chaosCount(t,db,"SELECT count(*) FROM investment.transaction_entries WHERE portfolio_id=$1 AND transaction_type IN ('BUY','SELL')",f.portfolioID);n!=3{t.Fatalf("baseline trade ledger count=%d",n)}
 previous:=db
 for round:=1;round<=3;round++ {
  recovered:=chaosRealCrashEvidence(t,previous,round)
  defer recovered.Close()
  chaosFinanceReconcile(t,recovered,f,round)
  read,securityAfter:=chaosFinancialHTTP(t)
  owner,e:=securityAfter.Login(context.Background(),auth.LoginRequest{Email:f.ownerEmail,Password:f.password});if e!=nil{t.Fatalf("owner login after crash %d: %v",round,e)}
  tokenAfter:=owner.Session.AccessToken
  code,data:=read("GET",path,tokenAfter,"","")
  if code!=200{t.Fatalf("positions after crash %d: %d",round,code)}
  extendedAssertSBERPosition(t,data)
  code,data=read("POST",base+"/transactions",tokenAfter,lastBody,lastKey)
  if code!=200&&code!=201{t.Fatalf("idempotent trade retry after crash %d status=%d response=%s",round,code,data)}
  if n:=chaosCount(t,recovered,"SELECT count(*) FROM investment.transaction_entries WHERE portfolio_id=$1 AND transaction_type IN ('BUY','SELL')",f.portfolioID);n!=3{t.Fatalf("duplicate financial trades after crash %d: %d",round,n)}
  _,afterEntries:=chaosFinanceStorage(t,recovered,f.portfolioID)
  if afterEntries!=f.entriesJSON{t.Fatalf("canonical ledger changed after retry in round %d",round)}
  t.Logf("FINANCIAL_WAC_CRASH_ROUND=%d BUY_SELL_LEDGER_IDENTICAL=YES SBER_QUANTITY=15.00000000 EXACT_WAC=150.00000000 IDEMPOTENT_SELL_RETRY=PASS DUPLICATE_TRADES=0",round)
  previous=recovered
 }
}
