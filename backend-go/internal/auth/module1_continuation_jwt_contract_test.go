package auth

// Independent continuation: characterize original PR #261's signed-claim RED
// without misrepresenting possession of the HMAC signing key as an exploit.
import (
 "crypto/hmac"
 "crypto/sha256"
 "encoding/base64"
 "encoding/json"
 "strings"
 "testing"
 "time"
)

func continuationJWT(t *testing.T, secret []byte, payload string) string {
 t.Helper()
 header := base64.RawURLEncoding.EncodeToString([]byte("{\"alg\":\"HS256\",\"typ\":\"JWT\"}"))
 body := header+"."+base64.RawURLEncoding.EncodeToString([]byte(payload))
 mac:=hmac.New(sha256.New,secret)
 _,_=mac.Write([]byte(body))
 return body+"."+base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func TestModule1ContinuationJWTRealIssuanceAndCannotForgeClaims(t *testing.T) {
 secret:=[]byte("continuation-jwt-immutable-fixture-secret-32-bytes")
 now:=time.Date(2026,10,14,15,30,0,0,time.UTC)
 user:=StoredUser{ID:"user-real-01",InvestmentSubjectID:"financial-owner-01"}
 token,err:=signAccessToken(secret,user,now,now.Add(15*time.Minute))
 if err!=nil{t.Fatal(err)}
 claims,err:=verifyAccessToken(secret,token,now)
 if err!=nil||claims.SubjectID!=user.ID||claims.UserID!=user.ID||claims.InvestmentSubjectID!=user.InvestmentSubjectID||claims.IssuedAt!=now.Unix()||claims.ExpiresAt!=now.Add(15*time.Minute).Unix(){t.Fatalf("REAL_ISSUANCE_CONTRACT_FAILED claims=%+v err=%v",claims,err)}
 parts:=strings.Split(token,".")
 if len(parts)!=3{t.Fatal("issued token malformed")}
 if _,err=verifyAccessToken(secret,token,now.Add(15*time.Minute));err==nil{t.Fatal("EXPIRATION_BOUNDARY_ACCEPTED")}
 newOwner:=StoredUser{ID:"user-attacker-02",InvestmentSubjectID:"financial-foreign-02"}
 differentKeyToken,err:=signAccessToken([]byte("other-key-different-principal-32-bytes"),newOwner,now,now.Add(time.Hour))
 if err!=nil{t.Fatal(err)}
 if _,err=verifyAccessToken(secret,differentKeyToken,now);err==nil{t.Fatal("CROSS_SECRET_SIGNATURE_ACCEPTED")}
 raw,err:=base64.RawURLEncoding.DecodeString(parts[1]);if err!=nil{t.Fatal(err)}
 var payload map[string]any
 if err=json.Unmarshal(raw,&payload);err!=nil{t.Fatal(err)}
 payload["investmentSubjectId"]="financial-foreign-02"
 forgedRaw,err:=json.Marshal(payload);if err!=nil{t.Fatal(err)}
 forgedBody:=parts[0]+"."+base64.RawURLEncoding.EncodeToString(forgedRaw)+"."+parts[2]
 if _,err=verifyAccessToken(secret,forgedBody,now);err==nil{t.Fatal("UNSIGNED_FINANCIAL_PRINCIPAL_SUBSTITUTION_ACCEPTED")}
 t.Log("REAL_ISSUER_SUB_EQUALS_UID=YES REAL_ISSUER_IAT_NOW=YES REAL_ISSUER_EXP_15_MINUTES=YES UNAUTHORIZED_PAYLOAD_CHANGE=DENIED CROSS_SECRET_HMAC=DENIED EXP_BOUNDARY=DENIED")
}

// Observe abnormal *correctly signed* tokens and explicitly report, rather
// than silently weakening the historical tests that demand stricter semantics.
func TestModule1ContinuationJWTClaimPolicyAndTypeBoundary(t *testing.T){
 secret:=[]byte("continuation-jwt-immutable-fixture-secret-32-bytes")
 now:=time.Date(2026,10,14,15,30,0,0,time.UTC)
 cases:=[]struct{name,payload string;invalid bool}{
  {"sub-uid-mismatch","{\"sub\":\"user-B\",\"uid\":\"user-A\",\"investmentSubjectId\":\"subject-A\",\"iat\":1791991800,\"exp\":1791995400}",false},
  {"future-iat","{\"sub\":\"user-A\",\"uid\":\"user-A\",\"investmentSubjectId\":\"subject-A\",\"iat\":1792078200,\"exp\":1792081800}",false},
  {"null-uid","{\"sub\":\"user-A\",\"uid\":null,\"investmentSubjectId\":\"subject-A\",\"exp\":1791995400}",true},
  {"bool-uid","{\"sub\":\"user-A\",\"uid\":true,\"investmentSubjectId\":\"subject-A\",\"exp\":1791995400}",true},
  {"float-exp","{\"sub\":\"user-A\",\"uid\":\"user-A\",\"investmentSubjectId\":\"subject-A\",\"exp\":1791995400.5}",true},
  {"huge-exp","{\"sub\":\"user-A\",\"uid\":\"user-A\",\"investmentSubjectId\":\"subject-A\",\"exp\":9999999999999999999999999999}",true},
  {"array-subject","{\"sub\":\"user-A\",\"uid\":\"user-A\",\"investmentSubjectId\":[\"subject-A\"],\"exp\":1791995400}",true},
  {"negative-exp","{\"sub\":\"user-A\",\"uid\":\"user-A\",\"investmentSubjectId\":\"subject-A\",\"exp\":-1}",true},
  {"zero-exp","{\"sub\":\"user-A\",\"uid\":\"user-A\",\"investmentSubjectId\":\"subject-A\",\"exp\":0}",true},
 }
 for _,tc:=range cases{
  token:=continuationJWT(t,secret,tc.payload)
  _,err:=verifyAccessToken(secret,token,now)
  if tc.invalid&&err==nil{t.Errorf("MALFORMED_TYPED_CLAIM_ACCEPTED case=%s",tc.name)}
  if !tc.invalid {t.Logf("SIGNED_CLAIM_POLICY_OBSERVATION case=%s verifier_accepted=%t externally_forgeable=NOT_DEMONSTRATED",tc.name,err==nil)}
 }
 // Duplicate object members are non-canonical JSON. Observe without claiming
 // a protocol contract not proven by the frozen requirements.
 duplicate:=continuationJWT(t,secret,"{\"sub\":\"user-A\",\"uid\":\"user-A\",\"investmentSubjectId\":\"subject-B\",\"investmentSubjectId\":\"subject-A\",\"iat\":1791991800,\"exp\":1791995400}")
 claims,err:=verifyAccessToken(secret,duplicate,now)
 t.Logf("SIGNED_DUPLICATE_JWT_CLAIM accepted=%t decoded_financial_subject=%q CONTRACT_DECISION=GATEKEEPER",err==nil,claims.InvestmentSubjectID)
}
