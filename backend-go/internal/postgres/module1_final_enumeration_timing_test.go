package postgres_test

import(
 "context"
 "encoding/json"
 "io"
 "net/http"
 "net/http/httptest"
 "sort"
 "strings"
 "testing"
 "time"

 "github.com/openinvest/openinvest/backend-go/internal/httpapi"
 "github.com/openinvest/openinvest/backend-go/internal/postgres"
 "github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func finalMedian(x []time.Duration)time.Duration{
 y:=append([]time.Duration(nil),x...);sort.Slice(y,func(i,j int)bool{return y[i]<y[j]});return (y[(len(y)-1)/2]+y[len(y)/2])/2
}
func TestEmailOwnershipFinalRandomizedEnumerationTimingAndBodyParity(t *testing.T){
 requireIndependentPG(t)
 s,db:=ownershipService(t);independentDrainPreviousDeliveries(t,s)
 ctx:=context.Background()
 existing,pending,expired:=independentAddress("timing-legacy"),independentAddress("timing-pending"),independentAddress("timing-expired")
 seed:=ownerRequest();seed.Email=existing;if _,err:=s.Register(ctx,seed);err!=nil{t.Fatal(err)}
 for _,v:=range []string{pending,expired}{if err:=s.BeginEmailRegistration(ctx,v);err!=nil{t.Fatal(err)}}
 if _,err:=db.Exec("UPDATE identity.registration_intents SET expires_at=statement_timestamp()-interval '1 second' WHERE email_normalized=$1",expired);err!=nil{t.Fatal(err)}
 app,err:=httpapi.New(verticalslice.NewService(&postgres.Store{},verticalslice.SystemClock{}),s,[]byte("module1-enumeration-timing-audit-32b"));if err!=nil{t.Fatal(err)}
 const samples=22
 names:=[]string{"fresh","legacy","pending","expired"}
 byName:=map[string][]time.Duration{}
 var baseline string
 // Latin-square interleaving avoids chronologically comparing all fresh
 // calls to all legacy calls. Exactly 88 calls stay below the 100/IP limit.
 for i:=0;i<samples;i++{
  for offset:=0;offset<len(names);offset++{
   label:=names[(i+offset)%len(names)]
   email:=map[string]string{"legacy":existing,"pending":pending,"expired":expired}[label]
   if label=="fresh"{email=independentAddress("timing-fresh")}
   req:=httptest.NewRequest(http.MethodPost,"/api/v1/auth/register",strings.NewReader(`{"email":"`+email+`"}`))
   req.Header.Set("Content-Type","application/json")
   start:=time.Now()
   response,err:=app.Test(req)
   elapsed:=time.Since(start)
   if err!=nil{t.Fatal(err)}
   body,err:=io.ReadAll(response.Body);response.Body.Close();if err!=nil{t.Fatal(err)}
   if response.StatusCode!=202||len(response.Cookies())!=0||response.Header.Get("Cache-Control")!="no-store"{
    t.Fatalf("enumeration leakage before timing: label=%s status=%d cookies=%d cache=%q body=%s",label,response.StatusCode,len(response.Cookies()),response.Header.Get("Cache-Control"),body)
   }
   var envelope struct{Data json.RawMessage `json:"data"`}
   if err:=json.Unmarshal(body,&envelope);err!=nil{t.Fatal(err)}
   if baseline==""{baseline=string(envelope.Data)}
   if string(envelope.Data)!=baseline{t.Fatalf("different registration response by identity state: label=%s",label)}
   byName[label]=append(byName[label],elapsed)
  }
 }
 reference:=finalMedian(byName["fresh"])
 for _,label:=range names{
  duration:=finalMedian(byName[label]);if duration<=0{t.Fatalf("clock error for %s",label)}
  t.Logf("TIMING_OBSERVATION class=%s N=%d median_us=%d reference_fresh_us=%d",label,len(byName[label]),duration.Microseconds(),reference.Microseconds())
  // This is a coarse regression tripwire, NOT a statistical proof of
  // indistinguishability on a real network. The deliberately wide limit
  // only catches catastrophic membership-specific branches.
  if duration>reference*12+70*time.Millisecond{t.Fatalf("extreme timing side channel candidate %s: median=%s fresh=%s",label,duration,reference)}
 }
 t.Log("REGISTRATION_OBSERVABLE_STATUS_BODY_COOKIE_PARITY=PASS LOOPBACK_TIMING_CATASTROPHIC_OUTLIER=NO PRODUCTION_TIMING=NOT_VERIFIED")
}
