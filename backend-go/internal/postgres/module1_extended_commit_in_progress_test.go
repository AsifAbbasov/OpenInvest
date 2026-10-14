package postgres_test

// PostgreSQL process is killed while pg_stat_activity independently proves
// it is actively executing COMMIT. A deferred constraint trigger ensures the
// observation interval is real, not a timer-only timing guess.
// This DOES NOT prove internal WAL critical-section overlap.
import (
 "context"
 "database/sql"
 "fmt"
 "strings"
 "testing"
 "time"

 "github.com/google/uuid"
)
func TestModule1ExtendedRealKillWhileCOMMITExecutingDeferredTrigger(t *testing.T){
 chaosAllowed(t)
 _,db:=ownershipService(t)
 suffix:=strings.ReplaceAll(uuid.NewString(),"-","")
 table:="oi_commit_probe_"+suffix
 fn:="oi_commit_delay_"+suffix
 if _,e:=db.Exec("CREATE TABLE "+table+" (id bigint PRIMARY KEY)");e!=nil{t.Fatal(e)}
 if _,e:=db.Exec(fmt.Sprintf("CREATE FUNCTION %s() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN PERFORM pg_sleep(25); RETURN NEW; END $$",fn));e!=nil{t.Fatal(e)}
 if _,e:=db.Exec(fmt.Sprintf("CREATE CONSTRAINT TRIGGER deferred_pause AFTER INSERT ON %s DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION %s()",table,fn));e!=nil{t.Fatal(e)}
 defer func(){
  ctx,cancel:=context.WithTimeout(context.Background(),8*time.Second);defer cancel()
  _,_=db.ExecContext(ctx,"DROP TABLE IF EXISTS "+table+" CASCADE")
  _,_=db.ExecContext(ctx,"DROP FUNCTION IF EXISTS "+fn+"()")
 }()
 ctx,cancel:=context.WithTimeout(context.Background(),40*time.Second);defer cancel()
 conn,e:=db.Conn(ctx);if e!=nil{t.Fatal(e)}
 defer conn.Close()
 var pid int
 if e=conn.QueryRowContext(ctx,"SELECT pg_backend_pid()").Scan(&pid);e!=nil{t.Fatal(e)}
 tx,e:=conn.BeginTx(ctx,&sql.TxOptions{});if e!=nil{t.Fatal(e)}
 if _,e=tx.ExecContext(ctx,"INSERT INTO "+table+" (id) VALUES (1)");e!=nil{_ =tx.Rollback();t.Fatal(e)}
 commitDone:=make(chan error,1)
 go func(){commitDone<-tx.Commit()}()
 // The active COMMIT backend is observed by a second independent connection.
 // The sleep is a deferred constraint trigger *inside the COMMIT statement*.
 // We explicitly do NOT equate this with PostgreSQL WAL critical sections.
 observed:=false
 deadline:=time.Now().Add(8*time.Second)
 for time.Now().Before(deadline){
  var state,query string
  e=db.QueryRowContext(ctx,"SELECT state,query FROM pg_stat_activity WHERE pid=$1",pid).Scan(&state,&query)
  if e==nil&&state=="active"&&strings.HasPrefix(strings.ToLower(strings.TrimSpace(query)),"commit"){
   observed=true;break
  }
  time.Sleep(40*time.Millisecond)
 }
 if !observed {t.Fatal("NOT_EXERCISED: cannot prove active COMMIT via pg_stat_activity")}
 recovered:=chaosRealCrashEvidence(t,db,1);defer recovered.Close()
 var result error
 select{
 case result=<-commitDone:
 case <-time.After(8*time.Second):t.Fatal("COMMIT_CLIENT_RESULT_TIMEOUT_AFTER_REAL_KILL")
 }
 if result==nil{t.Fatal("FALSE_PASS: killed COMMIT returned success")}
 var rows int
 e=recovered.QueryRowContext(context.Background(),"SELECT count(*) FROM "+table).Scan(&rows)
 if e!=nil||rows!=0{t.Fatalf("DEFERRED_UNCOMMITTED_PROBE_SURVIVED_KILL rows=%d err=%v",rows,e)}
 t.Logf("REAL_COMMIT_IN_PROGRESS=CONFIRMED_BY_PG_STAT_ACTIVITY BACKEND_PID=%d POSTMASTER_SIGKILL=CONFIRMED DOCKER_EXIT=137 WAL_RECOVERY=YES DEFERRED_TRIGGER_ROLLED_BACK=YES COMMIT_CLIENT_RESPONSE=ERROR INTERNAL_WAL_CRITICAL_SECTION_OVERLAP=NOT_PROVEN",pid)
}
