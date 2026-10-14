package postgres_test

// Independent actual PostgreSQL + production Fiber handler authorization
// boundary. Distinct from PR #261's direct auth.Service contention tests.
import (
 "context"
 "encoding/json"
 "io"
 "net/http"
 "net/http/httptest"
 "strings"
 "fmt"
 "testing"

 "github.com/google/uuid"
 "github.com/openinvest/openinvest/backend-go/internal/auth"
 "github.com/openinvest/openinvest/backend-go/internal/httpapi"
 "github.com/openinvest/openinvest/backend-go/internal/postgres"
 "github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func continuationAuthHTTP(t *testing.T,app interface{Test(*http.Request)(*http.Response,error)},method,path,body,cookie,csrf string)(int,[]byte,*http.Response){
 t.Helper()
 req:=httptest.NewRequest(method,path,strings.NewReader(body))
 if body!=""{req.Header.Set("Content-Type","application/json")}
 if cookie!=""{req.AddCookie(&http.Cookie{Name:auth.RefreshCookieName,Value:cookie})}
 if csrf!=""{req.Header.Set("X-CSRF-Token",csrf)}
 res,err:=app.Test(req);if err!=nil{t.Fatal(err)}
 raw,err:=io.ReadAll(res.Body);_ =res.Body.Close();if err!=nil{t.Fatal(err)}
 return res.StatusCode,raw,res
}

func TestModule1ContinuationRealHTTPMultiPrincipalCookieCSRFAndAllSessions(t *testing.T){
 chaosAllowed(t)
 initial,db:=ownershipService(t)
 f:=chaosFinanceSeed(t,initial,db)
 store,err:=postgres.Open(getOwnershipURL(t));if err!=nil{t.Fatal(err)}
 defer store.Close()
 security,err:=auth.NewService(store,nil,auth.Config{AccessTokenSecret:[]byte("continuation-postgres-auth-safety-secret-32-bytes"),RefreshCookieSecure:true})
 if err!=nil{t.Fatal(err)}
 app,err:=httpapi.New(verticalslice.NewService(store,verticalslice.SystemClock{}),security,[]byte("continuation-import-review-token-key-32-bytes"))
 if err!=nil{t.Fatal(err)}
 ctx:=context.Background()
 owner,err:=security.Login(ctx,auth.LoginRequest{Email:f.ownerEmail,Password:f.password});if err!=nil{t.Fatal(err)}
 ownerSecond,err:=security.Login(ctx,auth.LoginRequest{Email:f.ownerEmail,Password:f.password});if err!=nil{t.Fatal(err)}
 foreign,err:=security.Login(ctx,auth.LoginRequest{Email:f.outsiderEmail,Password:f.password});if err!=nil{t.Fatal(err)}
 ownerSessionsBefore:=chaosCount(t,db,"SELECT count(*) FROM identity.sessions WHERE user_id=$1 AND session_state='active'",f.ownerID)
 if ownerSessionsBefore<2{t.Fatalf("MISSING_MULTI_SESSION_POSITIVE_CONTROL count=%d",ownerSessionsBefore)}
 outsiderBefore:=chaosCount(t,db,"SELECT count(*) FROM identity.sessions s JOIN identity.users u ON u.id=s.user_id WHERE u.email_normalized=$1 AND s.session_state='active'",f.outsiderEmail)
 if outsiderBefore<1{t.Fatalf("MISSING_OUTSIDER_SESSION_POSITIVE_CONTROL count=%d",outsiderBefore)}
 refresh:="/api/v1/auth/refresh"
 logout:="/api/v1/auth/logout"
 code,_,_:=continuationAuthHTTP(t,app,"POST",refresh,"",owner.RefreshToken,"")
 if code==200{t.Fatal("HTTP_REFRESH_WITHOUT_CSRF_ACCEPTED")}
 code,_,_ = continuationAuthHTTP(t,app,"POST",refresh,"",owner.RefreshToken,foreign.Session.CSRFToken)
 if code==200{t.Fatal("HTTP_FOREIGN_PRINCIPAL_CSRF_ACCEPTED")}
 code,_,_ = continuationAuthHTTP(t,app,"POST",refresh,"",foreign.RefreshToken,owner.Session.CSRFToken)
 if code==200{t.Fatal("HTTP_COOKIE_HEADER_CROSS_PRINCIPAL_MISMATCH_ACCEPTED")}
 code,_,_ = continuationAuthHTTP(t,app,"POST",logout,"{\"allSessions\":true}",owner.RefreshToken,foreign.Session.CSRFToken)
 if code==200{t.Fatal("HTTP_LOGOUT_WITH_FOREIGN_CSRF_ACCEPTED")}
 // Invalid cross-cookie/CSRF operations must not revoke a valid active session.
 active:=chaosCount(t,db,"SELECT count(*) FROM identity.sessions WHERE user_id=$1 AND session_state='active'",f.ownerID)
 if active!=ownerSessionsBefore{t.Fatalf("FORBIDDEN_COOKIE_CSRF_MUTATED_OWNER_SESSIONS before=%d after=%d",ownerSessionsBefore,active)}
 code,body,response:=continuationAuthHTTP(t,app,"POST",refresh,"",owner.RefreshToken,owner.Session.CSRFToken)
 if code!=200{t.Fatalf("HTTP_LEGITIMATE_REFRESH_NOT_ALLOWED status=%d body=%s",code,body)}
 var rotated string
 for _,c:=range response.Cookies(){if c.Name==auth.RefreshCookieName {
  if !c.HttpOnly||!c.Secure||c.Path!="/api/v1/auth"||c.SameSite!=http.SameSiteStrictMode {t.Fatalf("REFRESH_COOKIE_SECURITY_FLAGS_BROKEN %+v",c)}
  rotated=c.Value
 }}
 if rotated==""||rotated==owner.RefreshToken{t.Fatal("HTTP_REFRESH_TOKEN_NOT_ROTATED")}
 var obj struct{Data struct{CSRFToken string `json:"csrfToken"`} `json:"data"`}
 if err=json.Unmarshal(body,&obj);err!=nil||obj.Data.CSRFToken==""{t.Fatalf("NO_ROTATED_CSRF_IN_RESPONSE: %v",err)}
 code,_,_=continuationAuthHTTP(t,app,"POST",refresh,"",owner.RefreshToken,owner.Session.CSRFToken)
 if code==200{t.Fatal("ROTATED_REFRESH_REPLAY_HTTP_ACCEPTED")}
 code,_,_=continuationAuthHTTP(t,app,"POST",logout,"{\"allSessions\":true}",rotated,obj.Data.CSRFToken)
 if code!=200{t.Fatalf("HTTP_OWNER_ALL_SESSIONS_LOGOUT_FAILED code=%d",code)}
 if n:=chaosCount(t,db,"SELECT count(*) FROM identity.sessions WHERE user_id=$1 AND session_state='active'",f.ownerID);n!=0{t.Fatalf("OWNER_SESSION_FAMILY_STILL_ACTIVE=%d",n)}
 code,_,_=continuationAuthHTTP(t,app,"POST",refresh,"",ownerSecond.RefreshToken,ownerSecond.Session.CSRFToken)
 if code==200{t.Fatal("SECOND_OWNER_SESSION_SURVIVED_ALL_SESSIONS_LOGOUT")}
 if n:=chaosCount(t,db,"SELECT count(*) FROM identity.sessions s JOIN identity.users u ON u.id=s.user_id WHERE u.email_normalized=$1 AND s.session_state='active'",f.outsiderEmail);n!=outsiderBefore{t.Fatalf("OTHER_PRINCIPAL_SESSION_REVOKED_BY_OWNER_LOGOUT before=%d after=%d",outsiderBefore,n)}
 // A legitimate independent user can still rotate their session.
 if _,err=security.Refresh(ctx,foreign.RefreshToken,foreign.Session.CSRFToken);err!=nil{t.Fatalf("LEGITIMATE_FOREIGN_PRINCIPAL_REFRESH_DISRUPTED: %v",err)}
 t.Log("REAL_FIBER_HTTP_PLUS_POSTGRES=YES MISSING_CSRF=DENIED FOREIGN_CSRF=DENIED COOKIE_HEADER_MISMATCH=DENIED INVALID_LOGOUT_NO_MUTATION=YES VALID_REFRESH_ROTATED=YES HTTP_ONLY_SECURE_STRICT_COOKIE=YES REPLAY=DENIED ALL_SESSIONS_LOGOUT=PASS FOREIGN_SESSION_PRESERVED=YES")
}

func TestModule1ContinuationRealHTTPBearerSubjectSwapAndLedgerNoMutation(t *testing.T){
 chaosAllowed(t)
 initial,db:=ownershipService(t)
 f:=chaosFinanceSeed(t,initial,db)
 run,security:=chaosFinancialHTTP(t)
 owner,err:=security.Login(context.Background(),auth.LoginRequest{Email:f.ownerEmail,Password:f.password});if err!=nil{t.Fatal(err)}
 other,err:=security.Login(context.Background(),auth.LoginRequest{Email:f.outsiderEmail,Password:f.password});if err!=nil{t.Fatal(err)}
 ownerPath:="/api/v1/portfolios/"+f.portfolioID+"/transactions"
 positive,data:=run("GET",ownerPath,owner.Session.AccessToken,"","")
 if positive!=200{t.Fatalf("LEGITIMATE_OWNER_BEARER_REJECTED code=%d body=%s",positive,data)}
 _,before:=chaosFinanceStorage(t,db,f.portfolioID)
 body:=fmt.Sprintf("{\"transactionType\":\"DEPOSIT\",\"ticker\":null,\"quantity\":null,\"unitPrice\":null,\"grossAmount\":{\"amount\":\"813.12345678\",\"currency\":\"RUB\"},\"commission\":{\"amount\":\"0.00000000\",\"currency\":\"RUB\"},\"tax\":{\"amount\":\"0.00000000\",\"currency\":\"RUB\"},\"tradeDate\":\"2026-03-05\",\"settlementDate\":null,\"note\":%q}","jwt-principal-probe-"+uuid.NewString())
 for _,x:=range []struct{name,token string}{{"foreign-issued",other.Session.AccessToken},{"missing",""},{"not-jwt","eyJhbGciOiJub25lIn0.e30."},{"tampered",owner.Session.AccessToken+"x"}} {
  code,_:=run("GET",ownerPath,x.token,"","")
  if code==200{t.Fatalf("FOREIGN_READ_SUCCEEDED variant=%s",x.name)}
  code,_=run("POST",ownerPath,x.token,body,"continuation-"+uuid.NewString())
  if code==200||code==201{t.Fatalf("FOREIGN_FINANCIAL_MUTATION variant=%s",x.name)}
 }
 _,after:=chaosFinanceStorage(t,db,f.portfolioID)
 if before!=after{t.Fatal("FOREIGN_OR_MALFORMED_BEARER_MUTATED_CANONICAL_LEDGER")}
 t.Log("REAL_HTTP_FINANCIAL_OWNER_POSITIVE=PASS CROSS_PRINCIPAL_BEARER=DENIED JWT_NONE=DENIED MISSING_JWT=DENIED TAMPERED_JWT=DENIED NO_FOREIGN_MUTATION=YES")
}
