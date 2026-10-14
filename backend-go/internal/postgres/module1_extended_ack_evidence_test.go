package postgres_test

import (
 "bytes"
 "context"
 "database/sql"
 "encoding/binary"
 "fmt"
 "io"
 "net"
 "net/url"
 "os"
 "sync"
 "testing"
 "time"

 "github.com/google/uuid"
 "github.com/openinvest/openinvest/backend-go/internal/auth"
 "github.com/openinvest/openinvest/backend-go/internal/postgres"
)

// A real PostgreSQL wire proxy drops CommandComplete(COMMIT) after the backend
// has sent it. This tests unknown CLIENT outcome, not PostgreSQL process death.
type extendedAckProxy struct {
 l net.Listener
 intercepted chan struct{}
 stopped chan struct{}
 once sync.Once
}
func extendedStartAckProxy(t *testing.T)*extendedAckProxy{
 t.Helper()
 l,e:=net.Listen("tcp","127.0.0.1:0");if e!=nil{t.Fatal(e)}
 p:=&extendedAckProxy{l:l,intercepted:make(chan struct{}),stopped:make(chan struct{})}
 t.Cleanup(func(){_ =l.Close()})
 go func(){
  defer close(p.stopped)
  client,e:=l.Accept();if e!=nil{return}
  server,e:=net.DialTimeout("tcp","127.0.0.1:5432",5*time.Second)
  if e!=nil{_ =client.Close();return}
  defer client.Close();defer server.Close()
  go func(){_,_=io.Copy(server,client);_ =server.Close()}()
  for {
   var hdr [5]byte
   if _,e=io.ReadFull(server,hdr[:]);e!=nil{return}
   n:=binary.BigEndian.Uint32(hdr[1:])
   if n<4||n>16*1024*1024{return}
   payload:=make([]byte,int(n)-4)
   if _,e=io.ReadFull(server,payload);e!=nil{return}
   if hdr[0]=='C' && bytes.Equal(payload,[]byte{'C','O','M','M','I','T',0}) {
    p.once.Do(func(){close(p.intercepted)})
    return
   }
   if _,e=client.Write(append(hdr[:],payload...));e!=nil{return}
  }
 }()
 return p
}
func (p *extendedAckProxy) databaseURL(t *testing.T)string{
 t.Helper()
 u,e:=url.Parse(os.Getenv("OPENINVEST_DATABASE_TEST_URL"))
 if e!=nil||u.Scheme!="postgres"{t.Fatalf("bad isolated db URL: %v",e)}
 u.Host=p.l.Addr().String()
 q:=u.Query();q.Set("sslmode","disable");u.RawQuery=q.Encode()
 return u.String()
}
func TestModule1ExtendedAckLossRealRegistrationStore(t *testing.T){
 chaosAllowed(t)
 _,direct:=ownershipService(t)
 for n:=1;n<=3;n++ {
  email:=fmt.Sprintf("real-ack-loss-%s-%d@example.test",uuid.NewString(),n)
  proxy:=extendedStartAckProxy(t)
  store,e:=postgres.Open(proxy.databaseURL(t));if e!=nil{t.Fatal(e)}
  service,e:=auth.NewService(store,nil,auth.Config{AccessTokenSecret:[]byte("real-ack-loss-test-secret-32-bytes-long")})
  if e!=nil{_ =store.Close();t.Fatal(e)}
  ctx,cancel:=context.WithTimeout(context.Background(),10*time.Second)
  result:=service.BeginEmailRegistration(ctx,email)
  cancel();_ =store.Close()
  select {
  case <-proxy.intercepted:
  case <-time.After(5*time.Second):t.Fatal("NOT_EXERCISED: server COMMIT CommandComplete not intercepted")
  }
  if result==nil {t.Fatal("FALSE_PASS: client received COMMIT success")}
  select {
  case <-proxy.stopped:
  case <-time.After(5*time.Second):t.Fatal("proxy did not close")
  }
  var intentID,jobID string;var gen,jobGen int64
  e=direct.QueryRow("SELECT i.id::text,i.generation,j.id::text,j.generation FROM identity.registration_intents i JOIN identity.email_delivery_jobs j ON j.intent_id=i.id WHERE i.email_normalized=$1",email).Scan(&intentID,&gen,&jobID,&jobGen)
  if e!=nil||intentID==""||jobID==""||gen!=jobGen{t.Fatalf("DURABLE_COMMIT_AND_JOB_MISSING: %v",e)}
  fresh,e:=postgres.Open(os.Getenv("OPENINVEST_DATABASE_TEST_URL"));if e!=nil{t.Fatal(e)}
  s,e:=auth.NewService(fresh,nil,auth.Config{AccessTokenSecret:[]byte("real-ack-loss-test-secret-32-bytes-long")})
  if e!=nil{_ =fresh.Close();t.Fatal(e)}
  e=s.BeginEmailRegistration(context.Background(),email)
  _=fresh.Close()
  if e!=nil{t.Fatalf("APPLICATION_RETRY_AFTER_UNKNOWN_COMMIT_FAILED: %v",e)}
  var count,users,sessions int;var finalIntent,finalJob string
  e=direct.QueryRow("SELECT i.id::text,j.id::text,(SELECT count(*) FROM identity.registration_intents x WHERE x.email_normalized=$1),(SELECT count(*) FROM identity.users u WHERE u.email_normalized=$1),(SELECT count(*) FROM identity.sessions s JOIN identity.users u ON u.id=s.user_id WHERE u.email_normalized=$1) FROM identity.registration_intents i JOIN identity.email_delivery_jobs j ON j.intent_id=i.id WHERE i.email_normalized=$1",email).Scan(&finalIntent,&finalJob,&count,&users,&sessions)
  if e!=nil||count!=1||finalIntent!=intentID||finalJob!=jobID||users!=0||sessions!=0{t.Fatalf("ACK_LOSS_RECONCILIATION_FAILED count=%d users=%d sessions=%d same_intent=%t same_job=%t err=%v",count,users,sessions,finalIntent==intentID,finalJob==jobID,e)}
  t.Logf("ROUND=%d REAL_SERVER_COMMIT_COMPLETE=OBSERVED CLIENT_COMMIT_ACK=WITHHELD CLIENT_OUTCOME=ERROR DURABLE_INTENT=YES APPLICATION_RETRY=SUCCESS DUPLICATES=0 PREPROOF_USERS=0",n)
 }
 t.Log("REAL_LOST_COMMIT_ACKNOWLEDGMENT=PASS CRITICAL_SECTION_POSTMASTER_KILL=NOT_PROVEN")
}
func TestModule1ExtendedNormalCommitControl(t *testing.T){
 chaosAllowed(t)
 db,e:=sql.Open("pgx",os.Getenv("OPENINVEST_DATABASE_TEST_URL"));if e!=nil{t.Fatal(e)}
 defer db.Close()
 ctx,cancel:=context.WithTimeout(context.Background(),5*time.Second);defer cancel()
 tx,e:=db.BeginTx(ctx,nil);if e!=nil{t.Fatal(e)}
 if _,e=tx.ExecContext(ctx,"SELECT 1");e!=nil{_ =tx.Rollback();t.Fatal(e)}
 if e=tx.Commit();e!=nil{t.Fatalf("normal COMMIT acknowledgment control failed: %v",e)}
 t.Log("NORMAL_SERVER_COMMIT_ACK=CONFIRMED WITHOUT_INTERCEPTION")
}
