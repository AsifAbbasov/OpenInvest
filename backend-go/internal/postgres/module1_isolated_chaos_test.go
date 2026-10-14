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
 if os.Getenv("OPENINVEST_CHAOS_ISOLATED")!="YES_DISPOSABLE_GITHUB_RUNNER" {t.Skip("destructive audit excluded from ordinary CI; dedicated isolated workflow required")}
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

// A readiness probe is not a post-crash reconciliation connection: explicitly
// create and ping a separate pool that never existed before SIGKILL.
// Capture evidence for a real database-process death, rather than accepting
// a successful readiness probe as evidence of process restart or WAL replay.
func chaosDockerEvidence(t *testing.T, args ...string) string {
 t.Helper()
 name:=os.Getenv("OPENINVEST_CHAOS_PG_CONTAINER")
 if name==""{t.Fatal("ISOLATION_FAILURE: no disposable service ID")}
 ctx,cancel:=context.WithTimeout(context.Background(),20*time.Second)
 defer cancel()
 cmd:=exec.CommandContext(ctx,"docker",append(args,name)...)
 out,e:=cmd.CombinedOutput()
 if e!=nil{t.Fatalf("DOCKER_EVIDENCE_FAILED args=%v err=%v output=%s",args,e,out)}
 return strings.TrimSpace(string(out))
}
func chaosDatabaseIdentity(t *testing.T, db *sql.DB)(string,string,int){
 t.Helper()
 ctx,cancel:=context.WithTimeout(context.Background(),5*time.Second)
 defer cancel()
 var system,start string;var backendPID int
 err:=db.QueryRowContext(ctx,"SELECT system_identifier::text,pg_postmaster_start_time()::text,pg_backend_pid() FROM pg_control_system()").Scan(&system,&start,&backendPID)
 if err!=nil{t.Fatalf("DATABASE_IDENTITY_UNAVAILABLE: %v",err)}
 return system,start,backendPID
}
func chaosRealCrashEvidence(t *testing.T, before *sql.DB, round int)*sql.DB{
 t.Helper()
 if os.Getenv("GITHUB_ACTIONS")!="true"||os.Getenv("RUNNER_ENVIRONMENT")!="github-hosted"{
  t.Fatal("ISOLATION_FAILURE: not an ephemeral hosted GitHub runner")
 }
 systemBefore,startBefore,pidBefore:=chaosDatabaseIdentity(t,before)
 originalID:=chaosDockerEvidence(t,"inspect","--format","{{.Id}}")
 image:=chaosDockerEvidence(t,"inspect","--format","{{.Config.Image}}")
 if image!="postgres:18"{t.Fatalf("WRONG_CRASH_TARGET_IMAGE: %q",image)}
 priorLogs:=chaosDockerEvidence(t,"logs","--tail","500")
 recoveryCount:=strings.Count(priorLogs,"automatic recovery in progress")
 chaosDocker(t,"kill","--signal=KILL")
 exitState:=chaosDockerEvidence(t,"inspect","--format","{{.State.Status}}|{{.State.ExitCode}}")
 if exitState!="exited|137"{t.Fatalf("REAL_KILL_NOT_CONFIRMED: %s",exitState)}
 chaosDocker(t,"start")
 restartedID:=chaosDockerEvidence(t,"inspect","--format","{{.Id}}")
 status:=chaosDockerEvidence(t,"inspect","--format","{{.State.Status}}")
 if originalID!=restartedID||status!="running"{t.Fatal("SAME_POSTGRES_CONTAINER_NOT_RESTARTED")}
 chaosDBReady(t)
 recovered:=chaosRecoveryDB(t)
 systemAfter,startAfter,pidAfter:=chaosDatabaseIdentity(t,recovered)
 if systemBefore!=systemAfter||startBefore==startAfter{
  t.Fatalf("CLUSTER_PERSISTENCE_OR_POSTMASTER_RESTART_UNPROVEN: same_system=%t new_start=%t",systemBefore==systemAfter,startBefore!=startAfter)
 }
 logs:=chaosDockerEvidence(t,"logs","--tail","500")
 if strings.Count(logs,"automatic recovery in progress")<=recoveryCount||!strings.Contains(logs,"redo done at"){
  t.Fatal("WAL_RECOVERY_LOGS_NOT_OBSERVED")
 }
 t.Logf("ROUND=%d POSTMASTER_ACTUALLY_KILLED=YES DOCKER_SIGKILL_EXIT=137 SAME_CONTAINER_RESTARTED=YES SAME_CLUSTER=YES OLD_POSTMASTER_START=%s NEW_POSTMASTER_START=%s OLD_BACKEND_PID=%d NEW_BACKEND_PID=%d WAL_RECOVERY_OBSERVED=YES FRESH_CONNECTION_ESTABLISHED=YES",round,startBefore,startAfter,pidBefore,pidAfter)
 return recovered
}

func chaosRecoveryDB(t *testing.T) *sql.DB {
 t.Helper()
 db,err:=sql.Open("pgx",os.Getenv("OPENINVEST_DATABASE_TEST_URL"));if err!=nil{t.Fatal(err)}
 ctx,cancel:=context.WithTimeout(context.Background(),5*time.Second)
 defer cancel()
 if err=db.PingContext(ctx);err!=nil{db.Close();t.Fatalf("FRESH_POSTCRASH_POOL_PING_FAILED: %v",err)}
 return db
}

type chaosRowEvidence struct{
 intentID string
 generation int64
 intentState string
 jobID string
 jobGeneration int64
 jobState string
 attempts int
 leaseID string
 jobCount int
 userCount int
 sessionCount int
}
func chaosReadPersistentState(t *testing.T,db *sql.DB,email string)chaosRowEvidence{
 t.Helper()
 ctx,cancel:=context.WithTimeout(context.Background(),5*time.Second)
 defer cancel()
 var v chaosRowEvidence
 err:=db.QueryRowContext(ctx,`SELECT i.id::text,i.generation,i.intent_state,j.id::text,j.generation,j.job_state,j.attempts,COALESCE(j.lease_id::text,''),
  (SELECT count(*) FROM identity.email_delivery_jobs k WHERE k.intent_id=i.id),
  (SELECT count(*) FROM identity.users u WHERE u.email_normalized=i.email_normalized),
  (SELECT count(*) FROM identity.sessions x JOIN identity.users u ON u.id=x.user_id WHERE u.email_normalized=i.email_normalized)
  FROM identity.registration_intents i JOIN identity.email_delivery_jobs j ON j.intent_id=i.id
  WHERE i.email_normalized=$1`,email).Scan(&v.intentID,&v.generation,&v.intentState,&v.jobID,&v.jobGeneration,&v.jobState,&v.attempts,&v.leaseID,&v.jobCount,&v.userCount,&v.sessionCount)
 if err!=nil{t.Fatalf("FRESH_PERSISTENT_STATE_QUERY_FAILED: %v",err)}
 if v.generation!=v.jobGeneration||v.jobCount!=1||v.userCount!=0||v.sessionCount!=0{
  t.Fatalf("INTENT_JOB_PREPROOF_INVARIANTS_BROKEN: %+v",v)
 }
 return v
}
func chaosOldPoolAndApplicationEvidence(t *testing.T,db *sql.DB,svc *auth.Service,fresh *sql.DB,round int){
 t.Helper()
 query:=func()(int,error){
  ctx,cancel:=context.WithTimeout(context.Background(),4*time.Second)
  defer cancel()
  var one int
  e:=db.QueryRowContext(ctx,"SELECT 1").Scan(&one)
  return one,e
 }
 first,firstErr:=query()
 second,secondErr:=query()
 t.Logf("ROUND=%d PRECRASH_POOL_FIRST_QUERY=%d ERROR=%v PRECRASH_POOL_SECOND_QUERY=%d SECOND_ERROR=%v",round,first,firstErr,second,secondErr)
 if secondErr!=nil||second!=1{t.Fatalf("DATABASE_SQL_OLD_POOL_PERMANENTLY_UNAVAILABLE: %v",secondErr)}
 address:="reconnect-"+uuid.NewString()+"@example.test"
 makeRegistration:=func()error{
  ctx,cancel:=context.WithTimeout(context.Background(),5*time.Second)
  defer cancel()
  return svc.BeginEmailRegistration(ctx,address)
 }
 firstAppError:=makeRegistration()
 t.Logf("ROUND=%d PRECRASH_APPLICATION_STORE_FIRST_ATTEMPT_ERROR=%v",round,firstAppError)
 if firstAppError!=nil{
  secondAppError:=makeRegistration()
  t.Logf("ROUND=%d PRECRASH_APPLICATION_STORE_SECOND_ATTEMPT_ERROR=%v",round,secondAppError)
  if secondAppError!=nil{t.Fatalf("APPLICATION_STORE_RECOVERY_FAILURE: %v",secondAppError)}
 }
 ctx,cancel:=context.WithTimeout(context.Background(),5*time.Second)
 defer cancel()
 var n int
 if e:=fresh.QueryRowContext(ctx,"SELECT count(*) FROM identity.registration_intents WHERE email_normalized=$1",address).Scan(&n);e!=nil||n!=1{
  t.Fatalf("APPLICATION_RECONNECTION_DUPLICATE_OR_LOST_REGISTRATION count=%d error=%v",n,e)
 }
 if _,e:=fresh.ExecContext(ctx,"DELETE FROM identity.registration_intents WHERE email_normalized=$1",address);e!=nil{t.Fatal(e)}
 t.Logf("ROUND=%d APPLICATION_STORE_RECOVERY=PASS RECONNECTION_NO_DUPLICATE=YES",round)
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
 financial:=chaosFinanceSeed(t,svc,db)
 prefix:="pgwal-"+uuid.NewString()+"-"
 t.Cleanup(func(){db.Exec("DELETE FROM identity.registration_intents WHERE email_normalized LIKE $1",prefix+"%")})
 for round:=1;round<=3;round++{
  address:=fmt.Sprintf("%s%d@example.test",prefix,round)
  if err:=svc.BeginEmailRegistration(context.Background(),address);err!=nil{t.Fatal(err)}
  store,err:=postgres.Open(os.Getenv("OPENINVEST_DATABASE_TEST_URL"));if err!=nil{t.Fatal(err)}
  lease,ok,err:=store.LeaseVerificationDelivery(context.Background(),chaosHash(address))
  if err!=nil||!ok||lease.Email!=address{store.Close();t.Fatalf("negative pre-crash control %t %v",ok,err)}

  store.Close()
  before:=chaosReadPersistentState(t,db,address)
  if before.jobID!=lease.JobID||before.leaseID!=lease.LeaseID||before.attempts!=1||before.jobState!="leased"||before.intentState!="pending"{t.Fatalf("INVALID_ACKNOWLEDGED_PRECRASH_STATE: %+v",before)}
  competitor,err:=postgres.Open(os.Getenv("OPENINVEST_DATABASE_TEST_URL"));if err!=nil{t.Fatal(err)}
  stolen,claimed,err:=competitor.LeaseVerificationDelivery(context.Background(),chaosHash(address+"-competitor"))
  competitor.Close()
  if err!=nil||claimed&&stolen.JobID==lease.JobID{t.Fatalf("ACTIVE_LEASE_EXCLUSION_NEGATIVE_CONTROL_FAILED claimed=%t err=%v",claimed,err)}
  recovered:=chaosRealCrashEvidence(t,db,round)
  defer recovered.Close()

  after:=chaosReadPersistentState(t,recovered,address)
  if after!=before{t.Fatalf("ACKNOWLEDGED_COMMIT_OR_LEASE_STATE_CHANGED_AFTER_WAL before=%+v after=%+v",before,after)}
  t.Logf("WAL_ROUND=%d ACKNOWLEDGED_COMMIT_PRESERVED=YES DELIVERY_STATE_PRESERVED=YES NO_PREMATURE_AUTH=YES",round)
  chaosFinanceReconcile(t,recovered,financial,round)
  chaosOldPoolAndApplicationEvidence(t,db,svc,recovered,round)
  var state string;var attempts int
  if err:=recovered.QueryRow("SELECT j.job_state,j.attempts FROM identity.email_delivery_jobs j JOIN identity.registration_intents i ON i.id=j.intent_id WHERE i.email_normalized=$1",address).Scan(&state,&attempts);err!=nil||state!="leased"||attempts!=1{t.Fatalf("durable lease lost after PostgreSQL SIGKILL: state=%s attempts=%d err=%v",state,attempts,err)}
  if count:=chaosCount(t,recovered,"SELECT count(*) FROM identity.users WHERE email_normalized=$1",address);count!=0{t.Fatalf("post-crash pre-proof user=%d",count)}
  chaosSQL(t,recovered,"UPDATE identity.email_delivery_jobs SET lease_until=statement_timestamp()-interval '2 seconds', available_at=statement_timestamp()-interval '2 seconds' WHERE id=$1",lease.JobID)
  fresh,err:=postgres.Open(os.Getenv("OPENINVEST_DATABASE_TEST_URL"));if err!=nil{t.Fatal(err)}
  replacement,ok,err:=fresh.LeaseVerificationDelivery(context.Background(),chaosHash(address+"-recover"))
  if err!=nil||!ok||replacement.Email!=address||replacement.LeaseID==lease.LeaseID{fresh.Close();t.Fatalf("WAL recovery failed new owner: %t %v",ok,err)}
  if err:=fresh.FinishVerificationDelivery(context.Background(),lease,true);err!=nil{t.Fatal(err)}
  successor:=chaosReadPersistentState(t,recovered,address)
  if successor.intentID!=before.intentID||successor.generation!=before.generation||successor.jobID!=before.jobID||successor.jobGeneration!=before.jobGeneration||successor.jobState!="leased"||successor.attempts!=2||successor.leaseID!=replacement.LeaseID||successor.leaseID==before.leaseID{t.Fatalf("STALE_LEASE_OWNERSHIP_OR_RETRY_BUDGET_INVALID: %+v",successor)}
  t.Logf("WAL_ROUND=%d STALE_CALLBACK_REJECTED=YES NEW_WORKER_OWNERSHIP=YES RETRY_BUDGET_BOUNDED=YES",round)
  var leaseID string
  if err:=recovered.QueryRow("SELECT lease_id FROM identity.email_delivery_jobs WHERE id=$1",replacement.JobID).Scan(&leaseID);err!=nil||leaseID!=replacement.LeaseID{t.Fatalf("old lease callback stole WAL successor: %q %v",leaseID,err)}
  fresh.Close()
 }
 t.Log("POSTGRES_REAL_SIGKILL_ROUNDS=3 WAL_DURABILITY=PASS STALE_LEASE_AFTER_RESTART=REJECTED FINANCIAL_LEDGER_PRESERVED=YES")
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
  recovered:=chaosRealCrashEvidence(t,db,n+1)
  defer recovered.Close()
  result:=<-done
  count:=chaosCount(t,recovered,"SELECT count(*) FROM identity.registration_intents WHERE email_normalized=$1",email)
  if count>1 {t.Fatalf("commit ambiguity duplicated identity intent: %d",count)}
  if result==nil&&count!=1{t.Fatalf("COMMIT acknowledged but WAL lost it: rows=%d",count)}
  s,_:=ownershipService(t)
  if err:=s.BeginEmailRegistration(context.Background(),email);err!=nil{t.Fatal(err)}
  count=chaosCount(t,recovered,"SELECT count(*) FROM identity.registration_intents WHERE email_normalized=$1",email)
  if count!=1{t.Fatalf("client uncertainty led to duplicate/missing intent: %d",count)}
  t.Logf("COMMIT_RACE_ROUND=%d CLIENT_COMMIT_ERROR=%t COMMITTED_COUNT_AFTER_CRASH=%d RETRIED_COUNT=%d",n+1,result!=nil,count,1)
 }
 t.Log("COMMIT_AMBIGUITY_RECONCILIATION=PASS CRITICAL_SECTION_OVERLAP=NON_DETERMINISTIC")
}
