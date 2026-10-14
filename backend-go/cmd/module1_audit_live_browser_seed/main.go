package main

// AUDIT-ONLY synthetic email proof generator for isolated GitHub Actions.
// It never sends email or prints the raw proof to logs.
import (
 "context"
 "encoding/json"
 "errors"
 "os"
 "time"

 "github.com/openinvest/openinvest/backend-go/internal/auth"
 "github.com/openinvest/openinvest/backend-go/internal/postgres"
)
type sink struct {raw string; email string}
func(s *sink)SendVerification(_ context.Context,email,raw string)error{s.email=email;s.raw=raw;return nil}
func main(){
 if os.Getenv("GITHUB_ACTIONS")!="true" || os.Getenv("RUNNER_ENVIRONMENT")!="github-hosted" || os.Getenv("OPENINVEST_CHAOS_ISOLATED")!="YES_DISPOSABLE_GITHUB_RUNNER"{panic("AUDIT_ISOLATION_NOT_PROVEN")}
 url:=os.Getenv("OPENINVEST_DATABASE_TEST_URL")
 if url==""{panic("DISPOSABLE_POSTGRES_MISSING")}
 path:=os.Getenv("OPENINVEST_AUDIT_PROOF_FILE")
 if path==""{panic("SECURE_PROOF_FILE_MISSING")}
 store,e:=postgres.Open(url);if e!=nil{panic(e)};defer store.Close()
 svc,e:=auth.NewService(store,nil,auth.Config{AccessTokenSecret:[]byte("audit-browser-test-key-separate-from-real-secrets-2026")})
 if e!=nil{panic(e)}
 ctx,cancel:=context.WithTimeout(context.Background(),15*time.Second);defer cancel()
 email:="module1-live-browser-proof@example.test"
 if e=svc.BeginEmailRegistration(ctx,email);e!=nil{panic(e)}
 collector:=&sink{}
 found,e:=svc.DeliverOne(ctx,collector)
 if e!=nil||!found||collector.raw==""||collector.email!=email{panic(errors.New("EPHEMERAL_MAIL_SINK_FAILED"))}
 serialized,e:=json.Marshal(struct{Email string `json:"email"`;Proof string `json:"proof"`}{email,collector.raw})
 if e!=nil{panic(e)}
 if e=os.WriteFile(path,serialized,0600);e!=nil{panic(e)}
 os.Stdout.WriteString("REAL_POSTGRES_PROOF_ISSUED=YES PLAIN_PROOF_LOGGED=NO EXTERNAL_EMAIL=NO\n")
}
