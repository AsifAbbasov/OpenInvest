package postgres_test

import(
 "bytes"
 "context"
 "encoding/json"
 "errors"
 "io"
 "os"
 "net/http/httptest"
 "strings"
 "testing"
 "github.com/google/uuid"
 "github.com/openinvest/openinvest/backend-go/internal/auth"
 "github.com/openinvest/openinvest/backend-go/internal/httpapi"
 "github.com/openinvest/openinvest/backend-go/internal/postgres"
 "github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestEmailOwnershipExtendedLegacyFinancialLedgerImmutableOnTakeover(t *testing.T){
 s,db:=ownershipService(t)
 independentDrainPreviousDeliveries(t,s)
 ctx:=context.Background()
 old:="financial-old-"+uuid.NewString()+"@example.test"
 attacker:="financial-outsider-"+uuid.NewString()+"@example.test"
 ownerSeed:=ownerRequest();ownerSeed.Email=old;ownerSeed.Password="legacy-owner-private-password-2026"
 owner,err:=s.Register(ctx,ownerSeed);if err!=nil{t.Fatal(err)}
 attackerSeed:=ownerRequest();attackerSeed.Email=attacker
 other,err:=s.Register(ctx,attackerSeed);if err!=nil{t.Fatal(err)}
 runtimeStore,err:=postgres.Open(getOwnershipURL(t));if err!=nil{t.Fatal(err)};defer runtimeStore.Close()
 runtimeAuth,err:=auth.NewService(runtimeStore,nil,auth.Config{AccessTokenSecret:[]byte("financial-ledger-audit-secret-over-32-bytes")});if err!=nil{t.Fatal(err)}
 // The original auth service and runtime auth must use the same JWT key for bearer proof.
 // Issue independent owner and outsider sessions via runtimeAuth's registered identities.
 runtimeOwner,err:=runtimeAuth.Login(ctx,auth.LoginRequest{Email:old,Password:ownerSeed.Password});if err!=nil{t.Fatal(err)}
 runtimeAttacker,err:=runtimeAuth.Login(ctx,auth.LoginRequest{Email:attacker,Password:attackerSeed.Password});if err!=nil{t.Fatal(err)}
 _=owner;_ =other
 app,err:=httpapi.New(verticalslice.NewService(runtimeStore,verticalslice.SystemClock{}),runtimeAuth,[]byte("financial-test-import-secret-long-enough"));if err!=nil{t.Fatal(err)}
 do:=func(method,path,token,body,key string)(int,[]byte){
  t.Helper()
  req:=httptest.NewRequest(method,path,strings.NewReader(body))
  if token!=""{req.Header.Set("Authorization","Bearer "+token)}
  if body!=""{req.Header.Set("Content-Type","application/json")}
  if key!=""{req.Header.Set("Idempotency-Key",key)}
  res,err:=app.Test(req);if err!=nil{t.Fatal(err)};defer res.Body.Close()
  buf,err:=io.ReadAll(res.Body);if err!=nil{t.Fatal(err)}
  return res.StatusCode,buf
 }
 ownerToken:=runtimeOwner.Session.AccessToken;outsiderToken:=runtimeAttacker.Session.AccessToken
 status,raw:=do("POST","/api/v1/portfolios",ownerToken,`{"name":"Legacy owned financial record","baseCurrency":"RUB"}`,"financial-create-"+uuid.NewString())
 if status!=201{t.Fatalf("negative control: owner cannot create portfolio: %d %s",status,raw)}
 var obj struct{Data struct{ID string `json:"id"`} `json:"data"`}
 if err:=json.Unmarshal(raw,&obj);err!=nil||obj.Data.ID==""{t.Fatalf("portfolio id absent: %v %s",err,raw)}
 basePath:="/api/v1/portfolios/"+obj.Data.ID
 deposit:=`{"transactionType":"DEPOSIT","ticker":null,"quantity":null,"unitPrice":null,"grossAmount":{"amount":"10000.00000000","currency":"RUB"},"commission":{"amount":"0.00000000","currency":"RUB"},"tax":{"amount":"0.00000000","currency":"RUB"},"tradeDate":"2026-01-09","settlementDate":"2026-01-09","note":"Audit financial preserve"}`
 status,raw=do("POST",basePath+"/transactions",ownerToken,deposit,"financial-deposit-"+uuid.NewString())
 if status!=201{t.Fatalf("negative control: deposit was not created: %d %s",status,raw)}
 readData:=func(endpoint string)[]byte {
  t.Helper()
  status,raw:=do("GET",endpoint,ownerToken,"","")
  if status!=200{t.Fatalf("owner read failed for %s: status=%d data=%s",endpoint,status,raw)}
  var e struct{Data json.RawMessage `json:"data"`}
  if err:=json.Unmarshal(raw,&e);err!=nil||len(e.Data)==0{t.Fatalf("snapshot parse: %v %s",err,raw)}
  var compact bytes.Buffer
  if err:=json.Compact(&compact,e.Data);err!=nil{t.Fatal(err)}
  return append([]byte(nil),compact.Bytes()...)
 }
 paths:=[]string{basePath,basePath+"/transactions",basePath+"/cash-flow"}
 baseline:=map[string][]byte{}
 for _,p:=range paths{baseline[p]=readData(p)}
 if !bytes.Contains(baseline[basePath+"/transactions"],[]byte("DEPOSIT")){t.Fatal("negative control: populated financial ledger missing")}
 var oldHash,subject string
 if err:=db.QueryRow("SELECT c.password_hash,l.investment_subject_id FROM identity.credentials c JOIN identity.user_investment_links l ON l.user_id=c.user_id JOIN identity.users u ON u.id=c.user_id WHERE u.email_normalized=$1",old).Scan(&oldHash,&subject);err!=nil{t.Fatal(err)}
 var nSessions int
 if err:=db.QueryRow("SELECT count(*) FROM identity.sessions WHERE user_id=$1",owner.User.ID).Scan(&nSessions);err!=nil||nSessions<1{t.Fatalf("owner session negative control: %d %v",nSessions,err)}
 if err:=s.BeginEmailRegistration(ctx,old);err!=nil{t.Fatal(err)}
 sink:=&ownershipMail{}
 if ok,err:=s.DeliverOne(ctx,sink);err!=nil||!ok||sink.email!=old{t.Fatalf("verification target mismatched: %t %v %q",ok,err,sink.email)}
 if err:=s.CompleteEmailRegistration(ctx,sink.token,ownerRequest());!errors.Is(err,auth.ErrInvalidVerification){t.Fatalf("legacy takeover admitted: %v",err)}
 for _,p:=range paths{
  status,_:=do("GET",p,outsiderToken,"","")
  if status==200{t.Fatalf("outside principal read legacy finance: %s",p)}
  if !bytes.Equal(baseline[p],readData(p)){t.Fatalf("legacy finance mutated after takeover attempt: %s",p)}
 }
 status,_=do("POST",basePath+"/transactions",outsiderToken,deposit,"financial-outsider-"+uuid.NewString())
 if status==200||status==201{t.Fatalf("outside principal changed ledger: %d",status)}
 for _,p:=range paths{if !bytes.Equal(baseline[p],readData(p)){t.Fatalf("cross-principal mutation of legacy finance: %s",p)}}
 var finalHash,finalSubject string
 if err:=db.QueryRow("SELECT c.password_hash,l.investment_subject_id FROM identity.credentials c JOIN identity.user_investment_links l ON l.user_id=c.user_id JOIN identity.users u ON u.id=c.user_id WHERE u.email_normalized=$1",old).Scan(&finalHash,&finalSubject);err!=nil{t.Fatal(err)}
 var finalSessions int
 if err:=db.QueryRow("SELECT count(*) FROM identity.sessions WHERE user_id=$1",owner.User.ID).Scan(&finalSessions);err!=nil{t.Fatal(err)}
 if oldHash!=finalHash||subject!=finalSubject||nSessions!=finalSessions{t.Fatal("legacy password, financial subject or session changed")}
 t.Log("LEGACY_FINANCIAL_SNAPSHOTS=3 TRANSACTION_SEEDED=YES OWNER_READ=PASS CROSS_PRINCIPAL_READ_AND_MUTATION=DENIED CREDENTIAL_SUBJECT_SESSION_INTACT=YES")
}
func getOwnershipURL(t *testing.T)string{t.Helper();url:=os.Getenv("OPENINVEST_DATABASE_TEST_URL");if url==""{t.Fatal("NOT_EXECUTED: PostgreSQL URL missing")};return url}
