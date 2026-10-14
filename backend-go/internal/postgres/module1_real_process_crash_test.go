package postgres_test

import (
 "bufio"
 "context"
 "database/sql"
 "errors"
 "fmt"
 "os"
 "os/exec"
 "strings"
 "syscall"
 "testing"
 "time"

 "github.com/google/uuid"
 "github.com/openinvest/openinvest/backend-go/internal/postgres"
)

// The child executes real production Store.LeaseVerificationDelivery and then
// dies with SIGKILL without invoking FinishVerificationDelivery.
func TestOpenInvestRealProcessLeaseCrashChild(t *testing.T){
 if os.Getenv("OPENINVEST_CHILD_SIGKILL")!="yes"{t.Skip("child helper only")}
 email:=os.Getenv("OPENINVEST_CHILD_EMAIL")
 if email==""{t.Fatal("missing child email")}
 store,err:=postgres.Open(os.Getenv("OPENINVEST_DATABASE_TEST_URL"));if err!=nil{t.Fatal(err)}
 claim,found,err:=store.LeaseVerificationDelivery(context.Background(),finalDigest(email+"-child-process"))
 if err!=nil||!found||claim.Email!=email{t.Fatalf("child could not lease %s found=%t err=%v",email,found,err)}
 fmt.Printf("REAL_CHILD_LEASE_CLAIMED email=%s job=%s lease=%s\n",email,claim.JobID,claim.LeaseID)
 os.Stdout.Sync()
 syscall.Kill(syscall.Getpid(),syscall.SIGKILL)
 t.Fatal("SIGKILL unexpectedly returned")
}
func TestOpenInvestRealProcessLeaseCrashAndRecovery(t *testing.T){
 requireIndependentPG(t)
 s,db:=ownershipService(t);independentDrainPreviousDeliveries(t,s)
 ctx:=context.Background();email:=independentAddress("real-process-sigkill")
 if err:=s.BeginEmailRegistration(ctx,email);err!=nil{t.Fatal(err)}
 cmd:=exec.Command(os.Args[0],"-test.run=^TestOpenInvestRealProcessLeaseCrashChild$","-test.v")
 cmd.Env=append(os.Environ(),"OPENINVEST_CHILD_SIGKILL=yes","OPENINVEST_CHILD_EMAIL="+email)
 pipe,err:=cmd.StdoutPipe();if err!=nil{t.Fatal(err)}
 cmd.Stderr=os.Stderr
 if err:=cmd.Start();err!=nil{t.Fatal(err)}
 scanner:=bufio.NewScanner(pipe);claimed:=""
 for scanner.Scan(){line:=scanner.Text();if strings.HasPrefix(line,"REAL_CHILD_LEASE_CLAIMED"){claimed=line;break}}
 err=cmd.Wait()
 if claimed==""{t.Fatalf("real child did not lease: scanner=%v wait=%v",scanner.Err(),err)}
 var exit *exec.ExitError
 if !errors.As(err,&exit)||exit.ExitCode()!=-1{t.Fatalf("child exit not signal-killed: %v",err)}
 if st,ok:=exit.Sys().(syscall.WaitStatus);!ok||!st.Signaled()||st.Signal()!=syscall.SIGKILL{t.Fatalf("child not killed by SIGKILL: %v",err)}
 var job,lease string;var n int
 if err:=db.QueryRow(`SELECT j.id,j.lease_id,j.attempts FROM identity.email_delivery_jobs j JOIN identity.registration_intents i ON i.id=j.intent_id WHERE i.email_normalized=$1`,email).Scan(&job,&lease,&n);err!=nil{t.Fatal(err)}
 if n!=1||lease==""{t.Fatalf("crashed worker lost lease evidence: attempts=%d lease=%q",n,lease)}
 // Move server time forward *for this fixture's lease only*; all operations
 // use actual worker SQL and PostgreSQL locks on the next claim.
 if _,err:=db.Exec("UPDATE identity.email_delivery_jobs SET lease_until=statement_timestamp()-interval '1 second',available_at=statement_timestamp()-interval '1 second' WHERE id=$1",job);err!=nil{t.Fatal(err)}
 worker,err:=postgres.Open(os.Getenv("OPENINVEST_DATABASE_TEST_URL"));if err!=nil{t.Fatal(err)};defer worker.Close()
 recovered,found,err:=worker.LeaseVerificationDelivery(ctx,finalDigest(email+"-fresh-worker"))
 if err!=nil||!found||recovered.Email!=email||recovered.LeaseID==lease{t.Fatalf("worker unable to recover real SIGKILL: found=%t err=%v",found,err)}
 if err:=worker.FinishVerificationDelivery(ctx,recovered,true);err!=nil{t.Fatal(err)}
 var state string
 if err:=db.QueryRow("SELECT job_state,attempts FROM identity.email_delivery_jobs WHERE id=$1",job).Scan(&state,&n);err!=nil||state!="delivered"||n!=2{t.Fatalf("state after process restart %s attempts=%d err=%v",state,n,err)}
 var users,sessions int
 if err:=db.QueryRow("SELECT count(*) FROM identity.users WHERE email_normalized=$1",email).Scan(&users);err!=nil||users!=0{t.Fatalf("unproved identity after SIGKILL %d %v",users,err)}
 if err:=db.QueryRow("SELECT count(*) FROM identity.sessions s JOIN identity.users u ON u.id=s.user_id WHERE u.email_normalized=$1",email).Scan(&sessions);err!=nil||sessions!=0{t.Fatalf("unproved session after SIGKILL %d %v",sessions,err)}
 t.Log("REAL_GO_PROCESS_SIGKILL=YES LEASE_RECOVERED=YES NO_PREPROOF_ACCOUNT=YES")
}

func TestOpenInvestRealBackendDisconnectRollback(t *testing.T){
 requireIndependentPG(t)
 _,db:=ownershipService(t)
 ctx,cancel:=context.WithTimeout(context.Background(),12*time.Second);defer cancel()
 table:="independent_abort_"+strings.ReplaceAll(uuid.NewString(),"-","")
 if _,err:=db.ExecContext(ctx,"CREATE TABLE public."+table+" (id INT PRIMARY KEY,amount INT NOT NULL)");err!=nil{t.Fatal(err)}
 defer db.Exec("DROP TABLE IF EXISTS public."+table)
 tx,err:=db.BeginTx(ctx,&sql.TxOptions{});if err!=nil{t.Fatal(err)}
 var pid int
 if err:=tx.QueryRowContext(ctx,"SELECT pg_backend_pid()").Scan(&pid);err!=nil{t.Fatal(err)}
 if _,err:=tx.ExecContext(ctx,"INSERT INTO public."+table+" VALUES (1,987654)");err!=nil{t.Fatal(err)}
 var killed bool
 if err:=db.QueryRowContext(ctx,"SELECT pg_terminate_backend($1)",pid).Scan(&killed);err!=nil||!killed{t.Fatalf("network/connection close failed: pid=%d killed=%t err=%v",pid,killed,err)}
 if err:=tx.Commit();err==nil{t.Fatal("terminated PostgreSQL connection committed transaction")}
 var n int
 if err:=db.QueryRowContext(ctx,"SELECT count(*) FROM public."+table).Scan(&n);err!=nil||n!=0{t.Fatalf("uncommitted marker became durable n=%d err=%v",n,err)}
 if _,err:=db.ExecContext(ctx,"INSERT INTO public."+table+" VALUES (1,987654)");err!=nil{t.Fatalf("retry after disconnect unavailable: %v",err)}
 if err:=db.QueryRowContext(ctx,"SELECT count(*) FROM public."+table+" WHERE id=1 AND amount=987654").Scan(&n);err!=nil||n!=1{t.Fatalf("post-disconnect retry did not persist exactly once: %d %v",n,err)}
 t.Log("TERMINATED_BACKEND_ROLLBACK=YES UNCOMMITTED_LEDGER_NOT_DURABLE=YES RETRY=ONE_COMMIT")
}
