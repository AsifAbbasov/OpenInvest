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
 "testing"

 "github.com/google/uuid"
 "github.com/openinvest/openinvest/backend-go/internal/auth"
 "github.com/openinvest/openinvest/backend-go/internal/httpapi"
 "github.com/openinvest/openinvest/backend-go/internal/postgres"
 "github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestEmailOwnershipFinalLegacyFinancialLedgerSurvivesVerificationAttack(t *testing.T){
 requireIndependentPG(t)
 s,db:=ownershipService(t)
 independentDrainPreviousDeliveries(t,s)
 ctx:=context.Background()
 oldEmail:=independentAddress("legacy-ledger-old")
 otherEmail:=independentAddress("legacy-ledger-other")
 old:=ownerRequest();old.Email=oldEmail;old.Password="legacy-financial-owner-password-2026"
 primary,err:=s.Register(ctx,old);if err!=nil{t.Fatal(err)}
 other:=ownerRequest();other.Email=otherEmail;other.Password="other-ledger-user-password-2026"
 otherAuth,err:=s.Register(ctx,other);if err!=nil{t.Fatal(err)}
 store,err:=postgres.Open(strings.TrimSpace(os.Getenv("OPENINVEST_DATABASE_TEST_URL")));if err!=nil{t.Fatal(err)};defer store.Close()
 app,err:=httpapi.New(verticalslice.NewService(store,verticalslice.SystemClock{}),s,[]byte("module1-final-ledger-audit-secret-32b"));if err!=nil{t.Fatal(err)}
 request:=func(method,path,bearer,body,key string)(int,[]byte){
  t.Helper()
  req:=httptest.NewRequest(method,path,strings.NewReader(body))
  if bearer!=""{req.Header.Set("Authorization","Bearer "+bearer)}
  if body!=""{req.Header.Set("Content-Type","application/json")}
  if key!=""{req.Header.Set("Idempotency-Key",key)}
  response,err:=app.Test(req);if err!=nil{t.Fatal(err)};defer response.Body.Close()
  raw,err:=io.ReadAll(response.Body);if err!=nil{t.Fatal(err)}
  return response.StatusCode,raw
 }
 payloadData:=func(raw []byte)json.RawMessage{
  t.Helper();var obj struct{Data json.RawMessage `json:"data"`};if err:=json.Unmarshal(raw,&obj);err!=nil{t.Fatal(err)};return obj.Data
 }
 portfolioKey:="final-financial-"+uuid.NewString()
 status,raw:=request("POST","/api/v1/portfolios",primary.Session.AccessToken,`{"name":"Immutable legacy ledger","baseCurrency":"RUB"}`,portfolioKey)
 if status!=http.StatusCreated{t.Fatalf("positive portfolio control: status=%d body=%s",status,raw)}
 var created struct{ID string `json:"id"`};if err:=json.Unmarshal(payloadData(raw),&created);err!=nil||created.ID==""{t.Fatalf("portfolio fixture invalid: %v %s",err,raw)}
 path:="/api/v1/portfolios/"+created.ID
 txBody:=`{"transactionType":"DEPOSIT","ticker":null,"quantity":null,"unitPrice":null,"grossAmount":{"amount":"1200.00000000","currency":"RUB"},"commission":{"amount":"0.00000000","currency":"RUB"},"tax":{"amount":"0.00000000","currency":"RUB"},"tradeDate":"2026-10-13","settlementDate":null,"note":"immutable-ledger-sentinel"}`
 status,raw=request("POST",path+"/transactions",primary.Session.AccessToken,txBody,"ledger-tx-"+uuid.NewString())
 if status!=http.StatusCreated{t.Fatalf("financial transaction fixture missing: %d %s",status,raw)}
 status,portfolioBefore:=request("GET",path,primary.Session.AccessToken,"","")
 if status!=http.StatusOK{t.Fatalf("owner cannot read original portfolio: %d",status)}
 status,txBefore:=request("GET",path+"/transactions",primary.Session.AccessToken,"","")
 if status!=http.StatusOK{t.Fatalf("owner cannot read financial history: %d",status)}
 status,raw=request("GET",path,otherAuth.Session.AccessToken,"","")
 if status!=http.StatusForbidden&&status!=http.StatusNotFound{t.Fatalf("negative principal control breached: status=%d body=%s",status,raw)}
 var oldPasswordHash,subject string;var sessionsBefore int
 if err:=db.QueryRow("SELECT c.password_hash,l.investment_subject_id FROM identity.credentials c JOIN identity.user_investment_links l ON c.user_id=l.user_id WHERE c.user_id=$1",primary.User.ID).Scan(&oldPasswordHash,&subject);err!=nil{t.Fatal(err)}
 if err:=db.QueryRow("SELECT count(*) FROM identity.sessions WHERE user_id=$1",primary.User.ID).Scan(&sessionsBefore);err!=nil{t.Fatal(err)}
 if err:=s.BeginEmailRegistration(ctx,oldEmail);err!=nil{t.Fatal(err)}
 mail:=&ownershipMail{};if found,err:=s.DeliverOne(ctx,mail);err!=nil||!found||mail.email!=oldEmail{t.Fatalf("ownership attack setup invalid: %t %s %v",found,mail.email,err)}
 change:=ownerRequest();change.Password="attempted-hijack-new-password-2026"
 if err:=s.CompleteEmailRegistration(ctx,mail.token,change);!errors.Is(err,auth.ErrInvalidVerification){t.Fatalf("existing financial account overwritten through ownership: %v",err)}
 var hashAfter,subjectAfter string;var sessionsAfter int
 if err:=db.QueryRow("SELECT c.password_hash,l.investment_subject_id FROM identity.credentials c JOIN identity.user_investment_links l ON c.user_id=l.user_id WHERE c.user_id=$1",primary.User.ID).Scan(&hashAfter,&subjectAfter);err!=nil{t.Fatal(err)}
 if err:=db.QueryRow("SELECT count(*) FROM identity.sessions WHERE user_id=$1",primary.User.ID).Scan(&sessionsAfter);err!=nil{t.Fatal(err)}
 if hashAfter!=oldPasswordHash||subjectAfter!=subject||sessionsAfter!=sessionsBefore{t.Fatalf("legacy financial identity mutated: hash=%t subject=%t sessions=%d/%d",hashAfter==oldPasswordHash,subjectAfter==subject,sessionsAfter,sessionsBefore)}
 status,portfolioAfter:=request("GET",path,primary.Session.AccessToken,"","")
 if status!=http.StatusOK||string(payloadData(portfolioBefore))!=string(payloadData(portfolioAfter)){t.Fatalf("financial portfolio mutated or inaccessible: status=%d before=%s after=%s",status,payloadData(portfolioBefore),payloadData(portfolioAfter))}
 status,txAfter:=request("GET",path+"/transactions",primary.Session.AccessToken,"","")
 if status!=http.StatusOK||string(payloadData(txBefore))!=string(payloadData(txAfter)){t.Fatalf("financial transaction ledger changed: %d before=%s after=%s",status,payloadData(txBefore),payloadData(txAfter))}
 status,raw=request("GET",path,otherAuth.Session.AccessToken,"","")
 if status!=http.StatusForbidden&&status!=http.StatusNotFound{t.Fatalf("new leak after verification attack: status=%d body=%s",status,raw)}
 t.Log("LEGACY_FINANCIAL_PORTFOLIO_AND_LEDGER_IMMUTABLE=PASS CROSS_PRINCIPAL_DENIED=PASS LOGIN_HASH_SESSION_SUBJECT_STABLE=PASS")
}
