package auth

import (
 "crypto/sha256"
 "encoding/base64"
 "encoding/hex"
 "strings"
 "testing"
)

func TestEmailOwnershipIndependentTokenNegativeControls(t *testing.T){
 raw,digest,err:=newVerificationToken()
 if err!=nil{t.Fatal(err)}
 if len(raw)!=43||len(digest)!=64{t.Fatalf("unexpected token envelope raw=%d digest=%d",len(raw),len(digest))}
 if got,err:=verificationDigest(raw);err!=nil||got!=digest{t.Fatalf("valid token oracle failed: %q %v",got,err)}
 bad:=[]string{"",raw+"=",raw[:42],raw+"A",raw[:10]+"."+raw[11:],strings.Repeat("/",43),strings.Repeat("a",4096)}
 for _,candidate:=range bad{
   if _,err:=verificationDigest(candidate);err==nil{t.Fatalf("malformed token was accepted length=%d",len(candidate))}
 }
 canonical:=base64.RawURLEncoding.EncodeToString(make([]byte,32))
 if _,err:=verificationDigest(canonical);err!=nil{t.Fatal(err)}
 t.Log("INDEPENDENT_TOKEN_LENGTH_ENCODING_DIGEST_NEGATIVE_CONTROLS=PASS")
}
func FuzzEmailOwnershipIndependentCanonicalDigest(f *testing.F){
 for _,seed:=range []string{"",strings.Repeat("A",43),strings.Repeat("a",43),strings.Repeat("A",42)+"_","AQIDBA","%41",strings.Repeat("A",44)}{f.Add(seed)}
 f.Fuzz(func(t *testing.T,raw string){
  if len(raw)>4096{return}
  got,err:=verificationDigest(raw)
  decoded,decodeErr:=base64.RawURLEncoding.Strict().DecodeString(raw)
  valid:=decodeErr==nil&&len(decoded)==32&&base64.RawURLEncoding.EncodeToString(decoded)==raw
  if valid{
   if err!=nil{t.Fatalf("valid strict canonical token rejected: %v",err)}
   sum:=sha256.Sum256([]byte(raw))
   if got!=hex.EncodeToString(sum[:]){t.Fatal("digest is not SHA256(raw token)")}
  }else if err==nil{t.Fatalf("malformed token accepted len=%d",len(raw))}
 })
}
