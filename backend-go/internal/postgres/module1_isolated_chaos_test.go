package postgres_test

// Real process/database crash experiments. These MUST run only on a disposable
// GitHub Actions PostgreSQL service and are never part of default go test.
import (
 "context"
 "crypto/sha256"
 "database/sql"
 "encoding/hex"
 "encoding/json"
 "fmt"
 "os"
 "os/exec"
 "path/filepath"
 "strings"
 "sync"
 "testing"
 "time"

 "github.com/google/uuid"
 "github.com/openinvest/openinvest/backend-go/internal/auth"
 "github.com/openinvest/openinvest/backend-go/internal/postgres"
)

func chaosAllowed(t *testing.T) {
 t.Helper()
 if os.Getenv("OPENINVEST_CHAOS_ISOLATED")!="YES_DISPOSABLE_GITHUB_RUNNER" {t.Fatal("NOT_EXECUTED: destructive isolation guard absent")}
 if os.Getenv("OPENINVEST_DATABASE_TEST_URL")=="" {t.Fatal("NOT_EXECUTED: ephemeral DB URL missing")}
}
func chaosHash(x string) string {a:=sha256.Sum256([]byte(x));return hex.EncodeToString(a[:])}
func chaosSQL(t *testing.T,db *sql.DB,q string,args ...any){t.Helper();if _,err:=db.Exec(q,args...);err!=nil{t.Fatal(err)}}
func chaosCount(t *testing.T,db *sql.DB,q string,args ...any) int {
 t.Helper();var n int;if e:=db.QueryRow(q,args...).Scan(&n);e!=nil{t.Fatal(e)};return n
}
func chaosDocker(t *testing.T,args ...string){
 t.Helper()
 name:=os.Getenv("OPENINVEST_CHAOS_PG_CONTAINER")
 if name==""{t.Fatal("NOT_EXECUTED: postgres service container id missing")}
 list:=append([]string{},args...);list=append(list,name)
 cmd:=exec.Command("docker",list...);cmd.Env=os.Environ()
 out,err:=cmd.CombinedOutput();if err!=nil{t.Fatalf("docker %v: %v: %s",list,err,out)}
}
func chaosDBReady(t *testing.T){
 t.Helper();url:=os.Getenv("OPENINVEST_DATABASE_TEST_URL")
 db,err:=sql.Open("pgx",url);if err!=nil{t.Fatal(err)};defer db.Close()
 deadline:=time.Now().Add(30*time.Second)
 for time.Now().Before(deadline){
  c,cancel:=context.WithTimeout(context.Background(),1200*time.Millisecond)
  err=db.PingContext(c);cancel()
  if err==nil{return}
  time.Sleep(350*time.Millisecond)
 }
 t.Fatal("RECOVERY_DEFECT: PostgreSQL did not become ready within 30 seconds")
}

func TestModule1ChaosWorkerSubprocess(t *testing.T){
 if os.Getenv("OPENINVEST_CHAOS_CHILD")!="1"{return}
 url:=os.Getenv("OPENINVEST_DATABASE_TEST_URL")
 store,err:=postgres.Open(url);if err!=nil{t.Fatal(err)};defer store.Close()
 digest:=chaosHash(fmt.Sprintf("%s-%d",os.Getenv("OPENINVEST_CHAOS_SEED"),os.Getpid()))
 lease,found,err:=store.LeaseVerificationDelivery(context.Background(),digest)
 if err!=nil||!found{t.Fatalf("CHILD_CLAIM_FAILED: %t %v",found,err)}
 data,err:=json.Marshal(lease);if err!=nil{t.Fatal(err)}
 if err:=os.WriteFile(os.Getenv("OPENINVEST_CHAOS_MARKER"),data,0600);err!=nil{t.Fatal(err)}
 for{time.Sleep(time.Hour)}
}

// Kill real Go OS processes with SIGKILL after they obtain genuine DB leases;
// then recover with freshly opened storage connections and 24 concurrent workers.
func TestModule1ChaosEightRealSIGKILLWorkersAndRecovery(t *testing.T){
 chaosAllowed(t)
 svc,db:=ownershipService(t);independentDrainPreviousDeliveries(t,svc)
 const n=8
 emails:=make([]string,n);prefix:="killworker-"+uuid.NewString()+"-"
 for i:=range emails{
  emails[i]=fmt.Sprintf("%s%d@example.test",prefix,i)
  if err:=svc.BeginEmailRegistration(context.Background(),emails[i]);err!=nil{t.Fatal(err)}
 }
 t.Cleanup(func(){db.Exec("DELETE FROM identity.registration_intents WHERE email_normalized LIKE $1",prefix+"%")})
 type killed struct{JobID,LeaseID,Email string}
 var crashed []killed
 for i:=0;i<n;i++{
  marker:=filepath.Join(t.TempDir(),fmt.Sprintf("worker-%d.json",i))
  cmd:=exec.Command(os.Args[0],"-test.run=^TestModule1ChaosWorkerSubprocess$","-test.timeout=40s")
  cmd.Env=append(os.Environ(),"OPENINVEST_CHAOS_CHILD=1","OPENINVEST_CHAOS_MARKER="+marker,"OPENINVEST_CHAOS_SEED="+fmt.Sprintf("%s-%d",prefix,i))
  if err:=cmd.Start();err!=nil{t.Fatal(err)}
  deadline:=time.Now().Add(8*time.Second)
  var data []byte
  for time.Now().Before(deadline) {
   data,_=os.ReadFile(marker);if len(data)>0{break};time.Sleep(30*time.Millisecond)
  }
  if len(data)==0 {cmd.Process.Kill();cmd.Wait();t.Fatalf("worker %d never leased an intent",i)}
  var lease killed
  if err:=json.Unmarshal(data,&lease);err!=nil{cmd.Process.Kill();cmd.Wait();t.Fatal(err)}
  if err:=cmd.Process.Kill();err!=nil{t.Fatal(err)}
  _=cmd.Wait()
  crashed=append(crashed,lease)
 }
 unique:=map[string]bool{}
 for _,x:=range crashed {
  if unique[x.JobID]{t.Fatalf("duplicate job leased by multiple workers %s",x.JobID)}
  unique[x.JobID]=true
  var state string;var attempts int
  if err:=db.QueryRow("SELECT job_state,attempts FROM identity.email_delivery_jobs WHERE id=$1",x.JobID).Scan(&state,&attempts);err!=nil||state!="leased"||attempts!=1{t.Fatalf("lost lease after real SIGKILL: %s %d %v",state,attempts,err)}
 }
 chaosSQL(t,db,"UPDATE identity.email_delivery_jobs SET lease_until=statement_timestamp()-interval '2 seconds', available_at=statement_timestamp()-interval '2 seconds' WHERE intent_id IN (SELECT id FROM identity.registration_intents WHERE email_normalized LIKE $1)",prefix+"%")
 const contenders=24
 start:=make(chan struct{});var wg sync.WaitGroup;errCh:=make(chan error,contenders)
 recovered:=make(chan auth.VerificationDelivery,contenders)
 for i:=0;i<contenders;i++ {
  wg.Add(1);go func(i int){defer wg.Done();<-start
   store,err:=postgres.Open(os.Getenv("OPENINVEST_DATABASE_TEST_URL"));if err!=nil{errCh<-err;return};defer store.Close()
   lease,ok,err:=store.LeaseVerificationDelivery(context.Background(),chaosHash(fmt.Sprintf("%s-recovery-%d",prefix,i)))
   if err!=nil{errCh<-err;return}
   if ok {recovered<-lease}
  }(i)
 }
 close(start);wg.Wait();close(errCh);close(recovered)
 for err:=range errCh{t.Error(err)}
 winners:=0;for lease:=range recovered{winners++;if !strings.HasPrefix(lease.Email,prefix){t.Errorf("cross-fixture lease: %s",lease.Email)}}
 if winners!=n{t.Errorf("unexpected recovery ownership: got %d, want %d",winners,n)}
 for _,old:=range crashed {
  stale:=auth.VerificationDelivery{JobID:old.JobID,LeaseID:old.LeaseID,Email:old.Email}
  worker,err:=postgres.Open(os.Getenv("OPENINVEST_DATABASE_TEST_URL"));if err!=nil{t.Fatal(err)}
  if err:=worker.FinishVerificationDelivery(context.Background(),stale,true);err!=nil{t.Error(err)}
  worker.Close()
 }
 for _,e:=range emails {
  var state string;var attempts int
  if err:=db.QueryRow("SELECT j.job_state,j.attempts FROM identity.email_delivery_jobs j JOIN identity.registration_intents i ON i.id=j.intent_id WHERE i.email_normalized=$1",e).Scan(&state,&attempts);err!=nil||state!="leased"||attempts!=2{t.Errorf("stale callback won or retry budget drift %s: %s %d %v",e,state,attempts,err)}
  if got:=chaosCount(t,db,"SELECT count(*) FROM identity.users WHERE email_normalized=$1",e);got!=0{t.Errorf("pre-proof identity %s=%d",e,got)}
 }
 t.Logf("REAL_SIGKILL_WORKERS=%d RECOVERY_CONTENDERS=%d SUCCESSFUL_UNIQUE_RECOVERIES=%d PREPROOF_IDENTITIES=0",n,contenders,winners)
}

// Isolated Docker crash: postmaster SIGKILL, WAL recovery, new DB connection,
// durable lease/user invariants, successful post-crash registration.
func TestModule1ChaosPostgresAbruptKillWALRecovery(t *testing.T){
 chaosAllowed(t)
 svc,db:=ownershipService(t);independentDrainPreviousDeliveries(t,svc)
 prefix:="pgwal-"+uuid.NewString()+"-"
 t.Cleanup(func(){db.Exec("DELETE FROM identity.registration_intents WHERE email_normalized LIKE $1",prefix+"%")})
 for round:=1;round<=3;round++{
  address:=fmt.Sprintf("%s%d@example.test",prefix,round)
  if err:=svc.BeginEmailRegistration(context.Background(),address);err!=nil{t.Fatal(err)}
  store,err:=postgres.Open(os.Getenv("OPENINVEST_DATABASE_TEST_URL"));if err!=nil{t.Fatal(err)}
  lease,ok,err:=store.LeaseVerificationDelivery(context.Background(),chaosHash(address))
  if err!=nil||!ok||lease.Email!=address{store.Close();t.Fatalf("negative pre-crash control %t %v",ok,err)}
  store.Close()
  chaosDocker(t,"kill","--signal=KILL")
  chaosDocker(t,"start")
  chaosDBReady(t)
  var state string;var attempts int
  if err:=db.QueryRow("SELECT j.job_state,j.attempts FROM identity.email_delivery_jobs j JOIN identity.registration_intents i ON i.id=j.intent_id WHERE i.email_normalized=$1",address).Scan(&state,&attempts);err!=nil||state!="leased"||attempts!=1{t.Fatalf("durable lease lost after PostgreSQL SIGKILL: state=%s attempts=%d err=%v",state,attempts,err)}
  if count:=chaosCount(t,db,"SELECT count(*) FROM identity.users WHERE email_normalized=$1",address);count!=0{t.Fatalf("post-crash pre-proof user=%d",count)}
  chaosSQL(t,db,"UPDATE identity.email_delivery_jobs SET lease_until=statement_timestamp()-interval '2 seconds', available_at=statement_timestamp()-interval '2 seconds' WHERE id=$1",lease.JobID)
  fresh,err:=postgres.Open(os.Getenv("OPENINVEST_DATABASE_TEST_URL"));if err!=nil{t.Fatal(err)}
  replacement,ok,err:=fresh.LeaseVerificationDelivery(context.Background(),chaosHash(address+"-recover"))
  if err!=nil||!ok||replacement.Email!=address||replacement.LeaseID==lease.LeaseID{fresh.Close();t.Fatalf("WAL recovery failed new owner: %t %v",ok,err)}
  if err:=fresh.FinishVerificationDelivery(context.Background(),lease,true);err!=nil{t.Fatal(err)}
  var leaseID string
  if err:=db.QueryRow("SELECT lease_id FROM identity.email_delivery_jobs WHERE id=$1",replacement.JobID).Scan(&leaseID);err!=nil||leaseID!=replacement.LeaseID{t.Fatalf("old lease callback stole WAL successor: %q %v",leaseID,err)}
  fresh.Close()
 }
 t.Log("POSTGRES_REAL_SIGKILL_ROUNDS=3 WAL_DURABILITY=PASS STALE_LEASE_AFTER_RESTART=REJECTED")
}

// Server-initiated physical socket closure while a transaction is open must
// roll back all uncommitted writes. Retry must create a single coherent intent.
func TestModule1ChaosNetworkDisconnectRollsBackAndCanRetry(t *testing.T){
 chaosAllowed(t)
 _,db:=ownershipService(t)
 const rounds=6
 for n:=0;n<rounds;n++{
  email:=independentAddress("network-drop")
  ctx:=context.Background()
  connection,err:=db.Conn(ctx);if err!=nil{t.Fatal(err)}
  var pid int
  if err:=connection.QueryRowContext(ctx,"SELECT pg_backend_pid()").Scan(&pid);err!=nil{t.Fatal(err)}
  tx,err:=connection.BeginTx(ctx,nil);if err!=nil{t.Fatal(err)}
  _,err=tx.ExecContext(ctx,"INSERT INTO identity.registration_intents(id,email_normalized,created_at,expires_at,next_delivery_at) VALUES($1,$2,statement_timestamp(),statement_timestamp()+interval '24 hours',statement_timestamp())",uuid.NewString(),email)
  if err!=nil{t.Fatal(err)}
  var terminated bool
  if err:=db.QueryRow("SELECT pg_terminate_backend($1)",pid).Scan(&terminated);err!=nil||!terminated{t.Fatalf("pg_backend termination failed %v %t",err,terminated)}
  if err:=tx.Commit();err==nil{t.Fatalf("network-closed transaction committed unexpectedly")}
  connection.Close()
  if x:=chaosCount(t,db,"SELECT count(*) FROM identity.registration_intents WHERE email_normalized=$1",email);x!=0{t.Fatalf("uncommitted row survived network disconnect: %d",x)}
  svc,_:=ownershipService(t)
  if err:=svc.BeginEmailRegistration(ctx,email);err!=nil{t.Fatal(err)}
  if x:=chaosCount(t,db,"SELECT count(*) FROM identity.registration_intents WHERE email_normalized=$1",email);x!=1{t.Fatalf("recovered retry not unique=%d",x)}
  chaosSQL(t,db,"DELETE FROM identity.registration_intents WHERE email_normalized=$1",email)
 }
 t.Logf("REAL_PG_SOCKET_TERMINATIONS=%d ROLLBACK_AT_DISCONNECT=PASS IDEMPOTENT_RETRY=PASS",rounds)
}

// Deliberately query both outcomes after a client-side ambiguous COMMIT: either
// exactly one committed intent OR zero; never partial/multiple membership.
// Does not falsely claim deterministic SIGKILL in the COMMIT critical section.
func TestModule1ChaosCommitAmbiguityAfterPostmasterCrash(t *testing.T){
 chaosAllowed(t)
 _,db:=ownershipService(t)
 prefix:="commit-ambig-"+uuid.NewString()+"-"
 t.Cleanup(func(){db.Exec("DELETE FROM identity.registration_intents WHERE email_normalized LIKE $1",prefix+"%")})
 for n:=0;n<3;n++ {
  email:=fmt.Sprintf("%s%d@example.test",prefix,n)
  tx,err:=db.BeginTx(context.Background(),nil);if err!=nil{t.Fatal(err)}
  chaosID:=uuid.NewString()
  _,err=tx.Exec("INSERT INTO identity.registration_intents(id,email_normalized,created_at,expires_at,next_delivery_at) VALUES($1,$2,statement_timestamp(),statement_timestamp()+interval '24 hours',statement_timestamp())",chaosID,email)
  if err!=nil{t.Fatal(err)}
  done:=make(chan error,1)
  go func(){done<-tx.Commit()}()
  chaosDocker(t,"kill","--signal=KILL")
  chaosDocker(t,"start")
  chaosDBReady(t)
  result:=<-done
  count:=chaosCount(t,db,"SELECT count(*) FROM identity.registration_intents WHERE email_normalized=$1",email)
  if count>1 {t.Fatalf("commit ambiguity duplicated identity intent: %d",count)}
  if result==nil&&count!=1{t.Fatalf("COMMIT acknowledged but WAL lost it: rows=%d",count)}
  s,_:=ownershipService(t)
  if err:=s.BeginEmailRegistration(context.Background(),email);err!=nil{t.Fatal(err)}
  count=chaosCount(t,db,"SELECT count(*) FROM identity.registration_intents WHERE email_normalized=$1",email)
  if count!=1{t.Fatalf("client uncertainty led to duplicate/missing intent: %d",count)}
  t.Logf("COMMIT_RACE_ROUND=%d CLIENT_COMMIT_ERROR=%t COMMITTED_COUNT_AFTER_CRASH=%d RETRIED_COUNT=%d",n+1,result!=nil,count,1)
 }
 t.Log("COMMIT_AMBIGUITY_RECONCILIATION=PASS CRITICAL_SECTION_OVERLAP=NON_DETERMINISTIC")
}
