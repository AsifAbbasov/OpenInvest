package postgres_test

// Seeds one REAL PostgreSQL-backed email verification proof for the actual
// Next.js + live Go API + Chromium audit. The plaintext proof is never logged.
import (
 "context"
 "encoding/json"
 "os"
 "strings"
 "testing"

 "github.com/google/uuid"
)

func TestAUTHGAPSeedRealBackendBrowserProof(t *testing.T){
 chaosAllowed(t)
 output:=os.Getenv("AUTHGAP_BROWSER_PROOF_FILE")
 if !strings.HasPrefix(output,"/tmp/") {t.Fatal("AUDIT_ISOLATION_FAILURE: a temporary output path under /tmp is required")}
 s,db:=ownershipService(t)
 independentDrainPreviousDeliveries(t,s)
 email:="browser-real-backend-"+uuid.NewString()+"@example.test"
 if err:=s.BeginEmailRegistration(context.Background(),email);err!=nil{t.Fatal(err)}
 if n:=chaosCount(t,db,"SELECT count(*) FROM identity.users WHERE email_normalized=$1",email);n!=0{t.Fatal("PRE_PROOF_IDENTITY_EXISTS")}
 sink:=&ownershipMail{}
 found,err:=s.DeliverOne(context.Background(),sink)
 if err!=nil||!found||sink.token==""||sink.email!=email{t.Fatalf("REAL_DATABASE_PROOF_DELIVERY_SETUP_FAILED found=%t matching=%t err=%v",found,sink.email==email,err)}
 var plaintextStored int
 if err:=db.QueryRow("SELECT count(*) FROM identity.email_verification_tokens t JOIN identity.registration_intents i ON t.intent_id=i.id WHERE i.email_normalized=$1 AND t.token_hash=$2",email,sink.token).Scan(&plaintextStored);err!=nil||plaintextStored!=0{t.Fatal("PLAINTEXT_PROOF_STORED_IN_DATABASE")}
 payload,err:=json.Marshal(map[string]string{"token":sink.token,"email":email})
 if err!=nil{t.Fatal(err)}
 if err:=os.WriteFile(output,payload,0600);err!=nil{t.Fatal(err)}
 t.Log("REAL_PG18_PROOF_ISSUED=YES SYNTHETIC_EMAIL_ONLY=YES PREPROOF_USERS=0 PLAINTEXT_TOKEN_STORAGE=NO SECURE_TEMP_OUTPUT=YES")
}
