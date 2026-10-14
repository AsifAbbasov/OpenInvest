package postgres_test

import (
 "context"
 "database/sql"
 "encoding/json"
 "fmt"
 "io"
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

type chaosFinanceFixture struct {
 ownerEmail string
 outsiderEmail string
 password string
 ownerID string
 subjectID string
 portfolioID string
 entryID string
 portfolioJSON string
 entriesJSON string
}

func chaosFinancialHTTP(t *testing.T) (func(string,string,string,string,string)(int,[]byte), string, string) {
 t.Helper()
 store,err:=postgres.Open(os.Getenv("OPENINVEST_DATABASE_TEST_URL"));if err!=nil{t.Fatal(err)}
 t.Cleanup(func(){_ = store.Close()})
 security,err:=auth.NewService(store,nil,auth.Config{AccessTokenSecret:[]byte("financial-chaos-audit-secret-32-bytes")});if err!=nil{t.Fatal(err)}
 app,err:=httpapi.New(verticalslice.NewService(store,verticalslice.SystemClock{}),security,[]byte("financial-chaos-import-secret-32-bytes"));if err!=nil{t.Fatal(err)}
 run:=func(method,path,token,body,key string)(int,[]byte){
  t.Helper()
  req:=httptest.NewRequest(method,path,strings.NewReader(body))
  if token!=""{req.Header.Set("Authorization","Bearer "+token)}
  if body!=""{req.Header.Set("Content-Type","application/json")}
  if key!=""{req.Header.Set("Idempotency-Key",key)}
  res,e:=app.Test(req);if e!=nil{t.Fatal(e)}
  defer res.Body.Close()
  raw,e:=io.ReadAll(res.Body);if e!=nil{t.Fatal(e)}
  return res.StatusCode,raw
 }
 return run,"",""
}

func chaosFinanceStorage(t *testing.T,db *sql.DB,pid string)(string,string){
 t.Helper()
 ctx:=context.Background()
 var p,entries string
 if e:=db.QueryRowContext(ctx,"SELECT row_to_json(p)::text FROM investment.portfolios p WHERE p.id=$1",pid).Scan(&p);e!=nil{t.Fatal(e)}
 if e:=db.QueryRowContext(ctx,"SELECT COALESCE(json_agg(row_to_json(te) ORDER BY te.entry_id)::text,'[]') FROM investment.transaction_entries te WHERE te.portfolio_id=$1",pid).Scan(&entries);e!=nil{t.Fatal(e)}
 return p,entries
}

func chaosFinanceSeed(t *testing.T,svc *auth.Service,db *sql.DB)*chaosFinanceFixture{
 t.Helper()
 f:=&chaosFinanceFixture{ownerEmail:"finance-crash-owner-"+uuid.NewString()+"@example.test",outsiderEmail:"finance-crash-outsider-"+uuid.NewString()+"@example.test",password:"isolated-test-only-legacy-password"}
 owner:=ownerRequest();owner.Email=f.ownerEmail;owner.Password=f.password
 x,e:=svc.Register(context.Background(),owner);if e!=nil{t.Fatal(e)}
 outsider:=ownerRequest();outsider.Email=f.outsiderEmail;outsider.Password=f.password
 if _,e=svc.Register(context.Background(),outsider);e!=nil{t.Fatal(e)}
 f.ownerID=x.User.ID
 if e=db.QueryRow("SELECT investment_subject_id FROM identity.user_investment_links WHERE user_id=$1",f.ownerID).Scan(&f.subjectID);e!=nil{t.Fatal(e)}
 run,_,_:=chaosFinancialHTTP(t)
 store,e:=postgres.Open(os.Getenv("OPENINVEST_DATABASE_TEST_URL"));if e!=nil{t.Fatal(e)};defer store.Close()
 security,e:=auth.NewService(store,nil,auth.Config{AccessTokenSecret:[]byte("financial-chaos-audit-secret-32-bytes")});if e!=nil{t.Fatal(e)}
 login,e:=security.Login(context.Background(),auth.LoginRequest{Email:f.ownerEmail,Password:f.password});if e!=nil{t.Fatal(e)}
 token:=login.Session.AccessToken
 code,raw:=run("POST","/api/v1/portfolios",token,`{"name":"Crash durable financial portfolio","baseCurrency":"RUB"}`,"fin-create-"+uuid.NewString())
 if code!=201{t.Fatalf("FINANCE_CREATE_PORTFOLIO_FAILED code=%d body=%s",code,raw)}
 var p struct{Data struct{ID string `json:"id"`} `json:"data"`}
 if e=json.Unmarshal(raw,&p);e!=nil||p.Data.ID==""{t.Fatalf("FINANCE_PORTFOLIO_ID_ABSENT err=%v",e)}
 f.portfolioID=p.Data.ID
 deposit:=`{"transactionType":"DEPOSIT","ticker":null,"quantity":null,"unitPrice":null,"grossAmount":{"amount":"10000.00000000","currency":"RUB"},"commission":{"amount":"0.00000000","currency":"RUB"},"tax":{"amount":"0.00000000","currency":"RUB"},"tradeDate":"2026-01-09","settlementDate":"2026-01-09","note":"Crash audit financial preservation"}`
 code,raw=run("POST","/api/v1/portfolios/"+f.portfolioID+"/transactions",token,deposit,"fin-deposit-"+uuid.NewString())
 if code!=201{t.Fatalf("FINANCE_DEPOSIT_FAILED code=%d body=%s",code,raw)}
 f.portfolioJSON,f.entriesJSON=chaosFinanceStorage(t,db,f.portfolioID)
 if !strings.Contains(f.entriesJSON,"DEPOSIT"){t.Fatal("FINANCE_NEGATIVE_CONTROL_FAILED: no committed deposit")}
 var n int
 if e=db.QueryRow("SELECT count(*) FROM investment.transaction_entries WHERE portfolio_id=$1 AND transaction_type='DEPOSIT' AND gross_amount=10000",f.portfolioID).Scan(&n);e!=nil||n!=1{t.Fatalf("FINANCE_BASELINE_INCORRECT n=%d err=%v",n,e)}
 return f
}

func chaosFinanceReconcile(t *testing.T,db *sql.DB,f *chaosFinanceFixture,round int){
 t.Helper()
 var ownerID,subject string
 if e:=db.QueryRow("SELECT u.id,l.investment_subject_id FROM identity.users u JOIN identity.user_investment_links l ON l.user_id=u.id WHERE u.email_normalized=$1",f.ownerEmail).Scan(&ownerID,&subject);e!=nil{t.Fatal(e)}
 if ownerID!=f.ownerID||subject!=f.subjectID{t.Fatal("FINANCE_PRINCIPAL_SUBJECT_CHANGED")}
 var portfolioSubject string
 if e:=db.QueryRow("SELECT subject_id FROM investment.portfolios WHERE id=$1",f.portfolioID).Scan(&portfolioSubject);e!=nil||portfolioSubject!=f.subjectID{t.Fatal("FINANCE_PORTFOLIO_OWNERSHIP_CHANGED")}
 p,entries:=chaosFinanceStorage(t,db,f.portfolioID)
 if p!=f.portfolioJSON||entries!=f.entriesJSON{t.Fatalf("FINANCE_LEDGER_OR_PORTFOLIO_CHANGED after WAL round=%d",round)}
 t.Logf("FINANCIAL_ROUND=%d LEDGER_STORAGE_IDENTICAL=YES PORTFOLIO_OWNER_IDENTICAL=YES",round)
 _=fmt.Sprintf
}
