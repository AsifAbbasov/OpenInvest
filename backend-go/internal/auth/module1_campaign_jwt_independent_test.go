package auth

// Independent campaign tests: malformed JWTs and signed-claim consistency.
// A deliberately valid HMAC over abnormal claims probes verifier assumptions;
// it does NOT claim that external attackers can sign arbitrary JWTs.
import (
 "crypto/hmac"
 "crypto/sha256"
 "encoding/base64"
 "encoding/json"
 "math/rand"
 "strings"
 "testing"
 "time"
)
func campaignSignedJWT(t *testing.T, secret []byte, header map[string]any, claims map[string]any) string {
 t.Helper()
 h,e:=json.Marshal(header);if e!=nil{t.Fatal(e)}
 c,e:=json.Marshal(claims);if e!=nil{t.Fatal(e)}
 body:=base64.RawURLEncoding.EncodeToString(h)+"."+base64.RawURLEncoding.EncodeToString(c)
 mac:=hmac.New(sha256.New,secret);_,_=mac.Write([]byte(body))
 return body+"."+base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
func TestModule1CampaignJWTMalformedAndTamperedBoundary(t *testing.T){
 secret:=[]byte("independent-module1-jwt-testing-key-32-bytes")
 now:=time.Date(2026,10,14,15,0,0,0,time.UTC)
 user:=StoredUser{ID:"user-campaign-01",InvestmentSubjectID:"financial-subject-01"}
 valid,e:=signAccessToken(secret,user,now.Add(-time.Second),now.Add(time.Hour))
 if e!=nil{t.Fatal(e)}
 got,e:=verifyAccessToken(secret,valid,now)
 if e!=nil||got.UserID!=user.ID||got.InvestmentSubjectID!=user.InvestmentSubjectID{t.Fatalf("positive signed JWT rejected: %v",e)}
 header:=map[string]any{"alg":"HS256","typ":"JWT"}
 normal:=map[string]any{"sub":user.ID,"uid":user.ID,"investmentSubjectId":user.InvestmentSubjectID,"iat":now.Unix(),"exp":now.Add(time.Hour).Unix()}
 cases:=map[string]string{
  "alg-none":campaignSignedJWT(t,secret,map[string]any{"alg":"none","typ":"JWT"},normal),
  "alg-HS512":campaignSignedJWT(t,secret,map[string]any{"alg":"HS512","typ":"JWT"},normal),
  "typ-refresh":campaignSignedJWT(t,secret,map[string]any{"alg":"HS256","typ":"refresh"},normal),
  "no-uid":campaignSignedJWT(t,secret,header,map[string]any{"sub":user.ID,"investmentSubjectId":user.InvestmentSubjectID,"iat":now.Unix(),"exp":now.Add(time.Hour).Unix()}),
  "no-financial-subject":campaignSignedJWT(t,secret,header,map[string]any{"sub":user.ID,"uid":user.ID,"iat":now.Unix(),"exp":now.Add(time.Hour).Unix()}),
  "expired":campaignSignedJWT(t,secret,header,map[string]any{"sub":user.ID,"uid":user.ID,"investmentSubjectId":user.InvestmentSubjectID,"iat":now.Unix()-100,"exp":now.Unix()-1}),
  "invalid-segments":"e30.e30",
  "empty":"",
  "invalid-b64":"%.%.%",
 }
 parts:=strings.Split(valid,".")
 mutated:=[]byte(parts[2]);if len(mutated)==0{t.Fatal("invalid positive fixture")};if mutated[0]=='a'{mutated[0]='b'}else{mutated[0]='a'}
 cases["signature-tamper"]=parts[0]+"."+parts[1]+"."+string(mutated)
 for label,token:=range cases{
  if _,e:=verifyAccessToken(secret,token,now);e==nil{t.Errorf("rejected JWT policy bypass: %s",label)}
 }
 rng:=rand.New(rand.NewSource(260142026))
 for i:=0;i<128;i++{
  x:=[]byte(valid);index:=rng.Intn(len(x));switch x[index]{case '.': index=(index+1)%len(x)}
  if x[index]=='a'{x[index]='b'}else{x[index]='a'}
  if _,e:=verifyAccessToken(secret,string(x),now);e==nil{t.Errorf("tampered JWT unexpectedly accepted seed=%d iteration=%d",260142026,i)}
 }
 t.Logf("JWT_BOUNDARY_POSITIVE=YES NEGATIVE_CASES=%d DETERMINISTIC_TAMPERS=128 SEED=260142026",len(cases))
}
func TestModule1CampaignJWTRequiresCanonicalSignedSubject(t *testing.T){
 secret:=[]byte("independent-module1-jwt-testing-key-32-bytes")
 now:=time.Date(2026,10,14,15,0,0,0,time.UTC)
 header:=map[string]any{"alg":"HS256","typ":"JWT"}
 claims:=map[string]any{"uid":"user-A","sub":"user-B","investmentSubjectId":"subject-A","iat":now.Unix(),"exp":now.Add(time.Hour).Unix()}
 token:=campaignSignedJWT(t,secret,header,claims)
 if _,e:=verifyAccessToken(secret,token,now);e==nil {
  t.Fatal("SIGNED_CLAIM_INCONSISTENCY: verifier accepted sub != uid; requires independent assessment of exploitability")
 }
}
func TestModule1CampaignJWTRejectsFarFutureIssuedAt(t *testing.T){
 secret:=[]byte("independent-module1-jwt-testing-key-32-bytes")
 now:=time.Date(2026,10,14,15,0,0,0,time.UTC)
 header:=map[string]any{"alg":"HS256","typ":"JWT"}
 claims:=map[string]any{"uid":"user-A","sub":"user-A","investmentSubjectId":"subject-A","iat":now.Add(24*time.Hour).Unix(),"exp":now.Add(25*time.Hour).Unix()}
 token:=campaignSignedJWT(t,secret,header,claims)
 if _,e:=verifyAccessToken(secret,token,now);e==nil {
  t.Fatal("SIGNED_FUTURE_ISSUANCE_ACCEPTED: verifier does not reject token issued 24h in the future; no external forgery demonstrated")
 }
}
