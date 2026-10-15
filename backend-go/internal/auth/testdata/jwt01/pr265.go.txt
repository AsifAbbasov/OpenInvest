package auth

// Audit-only JWT contract and structured token-boundary tests.
// Signed abnormal inputs represent compromised/alternate signers, NOT external forgery.
import (
 "crypto/hmac"
 "crypto/sha256"
 "encoding/base64"
 "fmt"
 "math/rand"
 "strings"
 "testing"
 "time"
)

func gapJWT(secret []byte, header, payload string) string {
 a:=base64.RawURLEncoding.EncodeToString([]byte(header))
 b:=base64.RawURLEncoding.EncodeToString([]byte(payload))
 unsigned:=a+"."+b
 m:=hmac.New(sha256.New,secret);_,_=m.Write([]byte(unsigned))
 return unsigned+"."+base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}

func TestAUTHGAPJWTCanonicalSignedIdentityContract(t *testing.T){
 secret:=[]byte("audit-independent-hmac-secret-32-bytes")
 now:=time.Unix(1800000000,0).UTC()
 user:=StoredUser{ID:"user-A",InvestmentSubjectID:"subject-A"}
 legitimate,e:=signAccessToken(secret,user,now,now.Add(time.Hour))
 if e!=nil{t.Fatal(e)}
 claims,e:=verifyAccessToken(secret,legitimate,now)
 if e!=nil||claims.UserID!=user.ID||claims.SubjectID!=user.ID||claims.InvestmentSubjectID!=user.InvestmentSubjectID{t.Fatalf("POSITIVE_ISSUANCE_CONTROL_FAILED err=%v claims=%+v",e,claims)}
 abnormal:=gapJWT(secret,`{"alg":"HS256","typ":"JWT"}`,fmt.Sprintf(`{"sub":"user-B","uid":"user-A","investmentSubjectId":"subject-A","exp":%d,"iat":%d}`,now.Add(time.Hour).Unix(),now.Unix()))
 actual,err:=verifyAccessToken(secret,abnormal,now)
 if err==nil{t.Fatalf("CONFIRMED_SIGNED_CLAIM_POLICY_FAILURE verifier accepted sub!=uid; claims=%+v; external HMAC forgery NOT demonstrated",actual)}
}

func TestAUTHGAPJWTFutureIssuedAtPolicy(t *testing.T){
 secret:=[]byte("audit-independent-hmac-secret-32-bytes")
 now:=time.Unix(1800000000,0).UTC()
 user:=StoredUser{ID:"user-A",InvestmentSubjectID:"subject-A"}
 normal,e:=signAccessToken(secret,user,now,now.Add(time.Hour))
 if e!=nil{t.Fatal(e)}
 if _,e=verifyAccessToken(secret,normal,now);e!=nil{t.Fatalf("POSITIVE_CONTROL_FAILED %v",e)}
 future:=gapJWT(secret,`{"alg":"HS256","typ":"JWT"}`,fmt.Sprintf(`{"sub":"user-A","uid":"user-A","investmentSubjectId":"subject-A","iat":%d,"exp":%d}`,now.Add(24*time.Hour).Unix(),now.Add(25*time.Hour).Unix()))
 result,e:=verifyAccessToken(secret,future,now)
 if e==nil{t.Fatalf("CONFIRMED_SIGNED_FUTURE_IAT_POLICY_FAILURE verifier accepted token issued +24h; claims=%+v; external forgery NOT demonstrated",result)}
}

func TestAUTHGAPJWTStructuredInvalidTokenBoundaries(t *testing.T){
 secret:=[]byte("audit-independent-hmac-secret-32-bytes")
 now:=time.Unix(1800000000,0).UTC()
 header:=`{"alg":"HS256","typ":"JWT"}`
 good:=fmt.Sprintf(`{"sub":"user-A","uid":"user-A","investmentSubjectId":"subject-A","iat":%d,"exp":%d}`,now.Unix(),now.Add(time.Hour).Unix())
 positive:=gapJWT(secret,header,good)
 if _,e:=verifyAccessToken(secret,positive,now);e!=nil{t.Fatalf("POSITIVE=%v",e)}
 patterns:=[]struct{name,header,payload string}{
  {"bad-alg-none",`{"alg":"none","typ":"JWT"}`,good},
  {"bad-alg-HS512",`{"alg":"HS512","typ":"JWT"}`,good},
  {"wrong-purpose",`{"alg":"HS256","typ":"EMAIL_VERIFY"}`,good},
  {"missing-user",header,fmt.Sprintf(`{"sub":"user-A","investmentSubjectId":"subject-A","exp":%d}`,now.Add(time.Hour).Unix())},
  {"empty-user",header,fmt.Sprintf(`{"sub":"user-A","uid":"","investmentSubjectId":"subject-A","exp":%d}`,now.Add(time.Hour).Unix())},
  {"missing-subject-id",header,fmt.Sprintf(`{"uid":"user-A","exp":%d}`,now.Add(time.Hour).Unix())},
  {"expired-now",header,fmt.Sprintf(`{"uid":"user-A","investmentSubjectId":"subject-A","exp":%d}`,now.Unix())},
  {"expired-past",header,fmt.Sprintf(`{"uid":"user-A","investmentSubjectId":"subject-A","exp":%d}`,now.Unix()-1)},
  {"overflow-exp",header,`{"uid":"user-A","investmentSubjectId":"subject-A","exp":9223372036854775808}`},
  {"fraction-exp",header,`{"uid":"user-A","investmentSubjectId":"subject-A","exp":1800003600.5}`},
  {"string-exp",header,`{"uid":"user-A","investmentSubjectId":"subject-A","exp":"1800003600"}`},
  {"null-uid",header,`{"uid":null,"investmentSubjectId":"subject-A","exp":1800003600}`},
  {"wrong-uid-type",header,`{"uid":42,"investmentSubjectId":"subject-A","exp":1800003600}`},
  {"wrong-i-at-type",header,`{"uid":"user-A","investmentSubjectId":"subject-A","iat":"yesterday","exp":1800003600}`},
  {"invalid-JSON",header,`{this is not json}`},
 }
 for _,p:=range patterns {
  t.Run(p.name,func(t *testing.T){tok:=gapJWT(secret,p.header,p.payload);if c,e:=verifyAccessToken(secret,tok,now);e==nil{t.Fatalf("MALFORMED_SIGNED_TOKEN_ACCEPTED claims=%+v",c)}})
 }
 if _,e:=verifyAccessToken(secret,strings.Replace(positive,".","..",1),now);e==nil{t.Fatal("MALFORMED_SEGMENTS_ACCEPTED")}
 const seed int64=26101502
 rng:=rand.New(rand.NewSource(seed))
 total:=512
 for i:=0;i<total;i++{
  b:=[]byte(positive)
  at:=rng.Intn(len(b))
  alphabet:="abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_"
  orig:=b[at]
  for b[at]==orig{b[at]=alphabet[rng.Intn(len(alphabet))]}
  if c,e:=verifyAccessToken(secret,string(b),now);e==nil{t.Fatalf("TAMPERED_UNSIGNED_OR_SIGNATURE_ACCEPTED index=%d claims=%+v",i,c)}
 }
 t.Logf("AUDIT_JWT_STRUCTURED_NEGATIVE=%d RANDOM_SINGLE_BYTE_TAMPERS=%d SEED=%d ORIGINAL_SIGNED_POSITIVE=PASS",len(patterns)+1,total,seed)
}
