package postgres_test

// Cross-system auth-to-import controls: a signed import review token may not
// cross principal/portfolio boundaries; logout/new-login replay must remain
// idempotent in real PostgreSQL. No mocks or real brokerage data.
import(
 "context"
 "encoding/json"
 "fmt"
 "testing"

 "github.com/google/uuid"
 "github.com/openinvest/openinvest/backend-go/internal/auth"
)

func TestModule1ContinuationRealHTTPImportReviewPrincipalAndSessionReplay(t *testing.T){
 chaosAllowed(t)
 initial,db:=ownershipService(t)
 f:=chaosFinanceSeed(t,initial,db)
 run,security:=chaosFinancialHTTP(t)
 owner,err:=security.Login(context.Background(),auth.LoginRequest{Email:f.ownerEmail,Password:f.password});if err!=nil{t.Fatal(err)}
 outsider,err:=security.Login(context.Background(),auth.LoginRequest{Email:f.outsiderEmail,Password:f.password});if err!=nil{t.Fatal(err)}
 label:="Independent audit source "+uuid.NewString()
 brokerOperation:="audit-import-"+uuid.NewString()
 csv:="transaction_type,ticker,quantity,unit_price,gross_amount,commission,tax,trade_date,settlement_date,currency,broker_operation_id,note\n"+
  fmt.Sprintf("DEPOSIT,,,,3456.12345678,0.00000000,0.00000000,2026-07-01,,RUB,%s,isolated import identity replay\n",brokerOperation)
 reviewPath:="/api/v1/portfolios/"+f.portfolioID+"/imports/review"
 appendPath:="/api/v1/portfolios/"+f.portfolioID+"/imports/append"
 reviewRaw,_:=json.Marshal(map[string]any{"sourceAccountLabel":label,"csvPayload":csv})
 code,payload:=run("POST",reviewPath,owner.Session.AccessToken,string(reviewRaw),"")
 if code!=200{t.Fatalf("OWNER_IMPORT_REVIEW_POSITIVE_FAILED status=%d response=%s",code,payload)}
 var review struct{Data struct{
  SourceFileHash string `json:"sourceFileHash"`
  ReviewToken string `json:"reviewToken"`
  Rows []struct{RowNumber int `json:"rowNumber"`;RowHash string `json:"rowHash"`;Status string `json:"status"`} `json:"rows"`
 } `json:"data"`}
 if err=json.Unmarshal(payload,&review);err!=nil||review.Data.ReviewToken==""||len(review.Data.Rows)!=1||review.Data.Rows[0].Status!="APPENDABLE"{t.Fatalf("OWNER_IMPORT_REVIEW_MALFORMED status=%d err=%v",code,err)}
 decision:=map[string]any{"rowNumber":review.Data.Rows[0].RowNumber,"rowHash":review.Data.Rows[0].RowHash,"action":"APPROVE"}
 appendRaw,_:=json.Marshal(map[string]any{"sourceAccountLabel":label,"sourceFileHash":review.Data.SourceFileHash,"reviewToken":review.Data.ReviewToken,"csvPayload":csv,"decisions":[]any{decision}})
 key:="continuation-import-"+uuid.NewString()
 // Foreign authenticated subject must not use a genuine owner-issued review.
 ownerBefore:=chaosCount(t,db,"SELECT count(*) FROM investment.transaction_entries WHERE portfolio_id=$1 AND gross_amount=3456.12345678",f.portfolioID)
 if ownerBefore!=0{t.Fatalf("OWNER_IMPORT_BASELINE_COLLISION=%d",ownerBefore)}
 code,_=run("POST",appendPath,outsider.Session.AccessToken,string(appendRaw),key)
 if code==200||code==201{t.Fatal("CROSS_PRINCIPAL_IMPORT_APPEND_SUCCEEDED_WITH_OWNER_REVIEW_TOKEN")}
 if n:=chaosCount(t,db,"SELECT count(*) FROM investment.transaction_entries WHERE portfolio_id=$1 AND gross_amount=3456.12345678",f.portfolioID);n!=0{t.Fatalf("FOREIGN_IMPORT_MUTATED_OWNER_LEDGER=%d",n)}
 // A distinct synthetic user is allowed to create a distinct portfolio, but
 // never to import under another portfolio's signed review authority.
 code,payload=run("POST","/api/v1/portfolios",outsider.Session.AccessToken,"{\"name\":\"Import review token isolation\",\"baseCurrency\":\"RUB\"}","continuation-outside-portfolio-"+uuid.NewString())
 if code!=201{t.Fatalf("OUTSIDER_PORTFOLIO_POSITIVE_FAILED status=%d body=%s",code,payload)}
 var portfolio struct{Data struct{ID string `json:"id"`} `json:"data"`}
 if err=json.Unmarshal(payload,&portfolio);err!=nil||portfolio.Data.ID==""{t.Fatalf("OUTSIDER_PORTFOLIO_PARSE_FAILED %v",err)}
 outsiderAppend:="/api/v1/portfolios/"+portfolio.Data.ID+"/imports/append"
 code,_=run("POST",outsiderAppend,outsider.Session.AccessToken,string(appendRaw),"continuation-import-other-"+uuid.NewString())
 if code==200||code==201{t.Fatal("OWNER_SIGNED_REVIEW_TOKEN_ACCEPTED_IN_OUTSIDER_PORTFOLIO")}
 if n:=chaosCount(t,db,"SELECT count(*) FROM investment.transaction_entries WHERE portfolio_id=$1",portfolio.Data.ID);n!=0{t.Fatalf("OUTSIDER_PORTFOLIO_MUTATED_WITH_FOREIGN_REVIEW=%d",n)}
 code,payload=run("POST",appendPath,owner.Session.AccessToken,string(appendRaw),key)
 if code!=201{t.Fatalf("OWNER_LEGITIMATE_IMPORT_APPEND_FAILED status=%d body=%s",code,payload)}
 if n:=chaosCount(t,db,"SELECT count(*) FROM investment.transaction_entries WHERE portfolio_id=$1 AND gross_amount=3456.12345678",f.portfolioID);n!=1{t.Fatalf("IMPORT_FINANCIAL_COUNT_AFTER_POSITIVE=%d",n)}
 _,before:=chaosFinanceStorage(t,db,f.portfolioID)
 revoked,err:=security.Logout(context.Background(),owner.RefreshToken,owner.Session.CSRFToken,false)
 if err!=nil||!revoked{t.Fatalf("OWNER_LOGOUT_BEFORE_IMPORT_REPLAY_FAILED revoked=%t err=%v",revoked,err)}
 if _,err=security.Refresh(context.Background(),owner.RefreshToken,owner.Session.CSRFToken);err==nil{t.Fatal("REVOKED_REFRESH_STILL_WORKS")}
 relogin,err:=security.Login(context.Background(),auth.LoginRequest{Email:f.ownerEmail,Password:f.password})
 if err!=nil{t.Fatal(err)}
 code,payload=run("POST",appendPath,relogin.Session.AccessToken,string(appendRaw),key)
 if code!=200&&code!=201{t.Fatalf("LEGITIMATE_IMPORT_REPLAY_AFTER_SESSION_CHANGE_FAILED status=%d response=%s",code,payload)}
 if n:=chaosCount(t,db,"SELECT count(*) FROM investment.transaction_entries WHERE portfolio_id=$1 AND gross_amount=3456.12345678",f.portfolioID);n!=1{t.Fatalf("IMPORT_REPLAY_DUPLICATION_OR_LOSS=%d",n)}
 _,after:=chaosFinanceStorage(t,db,f.portfolioID)
 if before!=after{t.Fatal("IMPORT_REPLAY_ALTERED_CANONICAL_LEDGER_AFTER_LOGOUT_AND_LOGIN")}
 t.Log("REAL_POSTGRES_SIGNED_IMPORT_REVIEW=PASS FOREIGN_PRINCIPAL_REVIEW_TOKEN_REUSE=DENIED FOREIGN_PORTFOLIO_REVIEW_TOKEN_REUSE=DENIED OWNER_AUTHORIZED_IMPORT=PASS OLD_REFRESH_REVOKED=YES RELOGIN_IMPORT_IDEMPOTENCY=PASS EXACT_DECIMAL_IMPORTED_ONCE=YES NO_LEDGER_MUTATION_ON_REPLAY=YES")
}
