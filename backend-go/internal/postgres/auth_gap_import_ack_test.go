package postgres_test

// Real PostgreSQL wire loss of COMMIT acknowledgment for an ATOMIC two-row import.
// This does not kill the postmaster: the fault is the client's unknown outcome.
import (
 "context"
 "strings"
 "testing"
 "time"

 "github.com/google/uuid"
 "github.com/openinvest/openinvest/backend-go/internal/decimal"
 "github.com/openinvest/openinvest/backend-go/internal/postgres"
 "github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestAUTHGAPImportTwoRowAtomicLostCommitAckAndReplay(t *testing.T){
 chaosAllowed(t)
 original,db:=ownershipService(t)
 f:=chaosFinanceSeed(t,original,db)
 first,_:=transitionRequests(t,f.portfolioID)
 first.TradeDate="2026-08-11"
 first.GrossAmount=&verticalslice.Money{Amount:decimal.Must("817.12345678"),Currency:verticalslice.RUB}
 first.ImportProvenance=nil
 second:=first
 second.TradeDate="2026-08-12"
 second.GrossAmount=&verticalslice.Money{Amount:decimal.Must("913.87654321"),Currency:verticalslice.RUB}
 second.ImportProvenance=nil
 batch:=verticalslice.AppendImportBatchRequest{
  PortfolioID:f.portfolioID,Transactions:[]verticalslice.AppendTransactionRequest{first,second},
  SourceKind:"USER_UPLOADED_FILE",
  SourceAccountLabel:"audit-import-ack-"+uuid.NewString(),
  SourceFileHash:strings.Repeat("f",64),
 }
 key:="audit-import-two-row-ack-"+uuid.NewString()
 path:="/api/v1/portfolios/"+f.portfolioID+"/imports/append"
 builder:=func(outcome verticalslice.ImportAppendOutcome)(verticalslice.CommandReplayArtifact,error){
  if len(outcome.Transactions)!=2{return verticalslice.CommandReplayArtifact{},verticalslice.ErrInvalidInput}
  return verticalslice.CommandReplayArtifact{StatusCode:201,Body:[]byte(`{"data":{"imported":2}}`),RequestID:"audit-request-1"},nil
 }
 proxy:=extendedStartAckProxy(t)
 store,e:=postgres.Open(proxy.databaseURL(t));if e!=nil{t.Fatal(e)}
 service:=verticalslice.NewService(store,verticalslice.SystemClock{})
 ctx,cancel:=context.WithTimeout(context.Background(),12*time.Second)
 _,_,lostAckErr:=service.AppendImportedTransactionsWithOutcomeReplay(ctx,verticalslice.RequestContext{},f.subjectID,key,path,batch,builder)
 cancel();_=store.Close()
 select{case <-proxy.intercepted:case <-time.After(5*time.Second):t.Fatal("FAULT_NOT_EXERCISED: backend COMMIT CommandComplete not intercepted")}
 if lostAckErr==nil{t.Fatal("FALSE_SUCCESS: lost server COMMIT acknowledgment returned success")}
 query:=`SELECT count(*) FROM investment.transaction_entries WHERE portfolio_id=$1 AND source_account_label=$2`
 before:=chaosCount(t,db,query,f.portfolioID,batch.SourceAccountLabel)
 if before!=2{t.Fatalf("ATOMIC_TWO_ROW_IMPORT_FAILED after lost ack: %d rows",before)}
 var total string
 if e:=db.QueryRow(`SELECT sum(gross_amount)::numeric(28,8)::text FROM investment.transaction_entries WHERE portfolio_id=$1 AND source_account_label=$2`,f.portfolioID,batch.SourceAccountLabel).Scan(&total);e!=nil||total!="1730.99999999"{t.Fatalf("IMPORTED_MONEY_CORRUPTED amount=%s err=%v",total,e)}
 beforePortfolio,beforeEntries:=chaosFinanceStorage(t,db,f.portfolioID)
 fresh,e:=postgres.Open(getOwnershipURL(t));if e!=nil{t.Fatal(e)}
 defer fresh.Close()
 newBuilderCalls:=0
 _,artifact,e:=verticalslice.NewService(fresh,verticalslice.SystemClock{}).AppendImportedTransactionsWithOutcomeReplay(context.Background(),verticalslice.RequestContext{},f.subjectID,key,path,batch,func(outcome verticalslice.ImportAppendOutcome)(verticalslice.CommandReplayArtifact,error){
  newBuilderCalls++
  return builder(outcome)
 })
 if e!=nil||artifact.StatusCode!=201||string(artifact.Body)!=`{"data":{"imported":2}}`{t.Fatalf("COMMITTED_IMPORT_REPLAY_FAILED err=%v artifact=%+v",e,artifact)}
 if newBuilderCalls!=0{t.Fatalf("COMPLETED_IMPORT_REPLAY_REEXECUTED_MUTATION builder_calls=%d",newBuilderCalls)}
 afterPortfolio,afterEntries:=chaosFinanceStorage(t,db,f.portfolioID)
 if beforePortfolio!=afterPortfolio||beforeEntries!=afterEntries{t.Fatal("TWO_ROW_IMPORT_CHANGED_AFTER_EXACT_RETRY")}
 _,_,wrongSubjectErr:=verticalslice.NewService(fresh,verticalslice.SystemClock{}).AppendImportedTransactionsWithOutcomeReplay(context.Background(),verticalslice.RequestContext{},uuid.NewString(),key,path,batch,builder)
 if wrongSubjectErr==nil{t.Fatal("UNAUTHORIZED_FOREIGN_SUBJECT_IMPORTED_OR_REPLAYED")}
 if chaosCount(t,db,query,f.portfolioID,batch.SourceAccountLabel)!=2{t.Fatal("FOREIGN_REPLAY_MUTATED_IMPORT")}
 t.Log("AUTH_GAP_IMPORT_COMMIT_ACK_WITHHELD=YES TWO_ROW_ATOMIC_COMMIT=YES EXACT_DECIMAL_TOTAL=1730.99999999 SAME_IDEMPOTENCY_REPLAY=EXACT_NO_REBUILD CROSS_PRINCIPAL_DENIED=YES NO_DUPLICATE_IMPORT=YES")
}
