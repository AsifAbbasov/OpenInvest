package postgres_test

import (
 "context"
 "fmt"
 "math"
 "sort"
 "testing"
 "time"

 "github.com/google/uuid"
 "github.com/openinvest/openinvest/backend-go/internal/auth"
)

// Bounded observable membership-timing diagnostic on a disposable PostgreSQL instance.
// Coarse guard only: passing does NOT prove absence of subtle timing enumeration.
func TestEmailOwnershipExtendedTimingEnumerationBoundedDiagnostic(t *testing.T) {
 svc,db:=ownershipService(t)
 independentDrainPreviousDeliveries(t,svc)
 ctx:=context.Background()
 legacy:="timing-legacy-"+uuid.NewString()+"@example.test"
 req:=ownerRequest();req.Email=legacy
 if _,err:=svc.Register(ctx,req);err!=nil{t.Fatal(err)}
 const n=24
 measured:=func(email string) float64 {
  start:=time.Now()
  if err:=svc.BeginEmailRegistration(ctx,email);err!=nil{t.Fatalf("begin registration: %v",err)}
  return float64(time.Since(start).Microseconds())/1000
 }
 var existing,fresh []float64
 for i:=0;i<n;i++ {
  unique:=fmt.Sprintf("timing-new-%s-%d@example.test",uuid.NewString(),i)
  if i%2==0 {
    existing=append(existing,measured(legacy));fresh=append(fresh,measured(unique))
  } else {
    fresh=append(fresh,measured(unique));existing=append(existing,measured(legacy))
  }
 }
 median:=func(values []float64) float64 {out:=append([]float64(nil),values...);sort.Float64s(out);return (out[len(out)/2-1]+out[len(out)/2])/2}
 mLegacy,mFresh:=median(existing),median(fresh)
 if mLegacy<=0||mFresh<=0{t.Fatalf("invalid measured timing distribution: %.4f %.4f",mLegacy,mFresh)}
 // Calibration: a deliberately injected 25 ms membership signal MUST be detectable.
 artificial:=append([]float64(nil),existing...)
 for i:=range artificial{artificial[i]+=25}
 if median(artificial)-mLegacy<24.9{t.Fatal("HARNESS_ORACLE_DEFECT: failed to detect calibrated timing leak")}
 ratio:=math.Max(mLegacy,mFresh)/math.Min(mLegacy,mFresh)
 t.Logf("TIMING_DIAGNOSTIC samples_per_class=%d legacy_median_ms=%.4f fresh_median_ms=%.4f median_ratio=%.3f injected_25ms_control=DETECTED FINE_GRAIN_TIMING=NOT_VERIFIED",n,mLegacy,mFresh,ratio)
 if math.Abs(mLegacy-mFresh)>12 && ratio>6{t.Fatalf("MEMBERSHIP_TIMING_COARSE_ORACLE_REJECTED: medians differ materially: %.3f vs %.3f ms",mLegacy,mFresh)}
 // Persistent negative control: the public acknowledgement must never create a second user.
 var users int
 if err:=db.QueryRow("SELECT count(*) FROM identity.users WHERE email_normalized=$1",legacy).Scan(&users);err!=nil||users!=1{t.Fatalf("legacy account changed: users=%d err=%v",users,err)}
}

func TestEmailOwnershipExtendedCapacity10000FailClosedAndRecovery(t *testing.T) {
 svc,db:=ownershipService(t)
 ctx:=context.Background()
 prefix:="saturation-"+uuid.NewString()+"-"
 blocked:="saturation-blocked-"+uuid.NewString()+"@example.test"
 t.Cleanup(func(){
  if _,err:=db.Exec("DELETE FROM identity.registration_intents WHERE email_normalized LIKE $1 OR email_normalized=$2",prefix+"%",blocked);err!=nil{t.Errorf("synthetic saturation cleanup failed: %v",err)}
 })
 var before int
 if err:=db.QueryRow("SELECT count(*) FROM identity.registration_intents WHERE intent_state='pending' AND expires_at>statement_timestamp()").Scan(&before);err!=nil{t.Fatal(err)}
 if before>=10000{t.Fatalf("fixture already saturated: %d",before)}
 count:=10000-before
 // One bounded, set-based PostgreSQL fixture seed, not 10,000 HTTP requests.
 result,err:=db.Exec(`INSERT INTO identity.registration_intents(id,email_normalized,created_at,expires_at,next_delivery_at)
 SELECT md5($1::text || g::text)::uuid,$1::text || g::text || '@example.test',
 statement_timestamp(),statement_timestamp()+interval '24 hours',statement_timestamp()
 FROM generate_series(1,$2::integer) AS g`,prefix,count)
 if err!=nil{t.Fatal(err)}
 inserted,_:=result.RowsAffected()
 if inserted!=int64(count){t.Fatalf("fixture rows=%d expected=%d",inserted,count)}
 var active int
 if err:=db.QueryRow("SELECT count(*) FROM identity.registration_intents WHERE intent_state='pending' AND expires_at>statement_timestamp()").Scan(&active);err!=nil{t.Fatal(err)}
 if active!=10000{t.Fatalf("capacity calibration: %d expected 10000",active)}
 // HTTP-level positive control: successful generic acknowledgement even when saturated.
 if err:=svc.BeginEmailRegistration(ctx,blocked);err!=nil{t.Fatalf("saturation leaked internal capacity: %v",err)}
 var created,users,sessions int
 if err:=db.QueryRow("SELECT count(*) FROM identity.registration_intents WHERE email_normalized=$1",blocked).Scan(&created);err!=nil||created!=0{t.Fatalf("saturated intake allowed extra intent: %d %v",created,err)}
 if err:=db.QueryRow("SELECT count(*) FROM identity.users WHERE email_normalized=$1",blocked).Scan(&users);err!=nil||users!=0{t.Fatalf("saturated intake activated user: %d %v",users,err)}
 if err:=db.QueryRow("SELECT count(*) FROM identity.sessions s JOIN identity.users u ON s.user_id=u.id WHERE u.email_normalized=$1",blocked).Scan(&sessions);err!=nil||sessions!=0{t.Fatalf("saturated intake issued session: %d %v",sessions,err)}
 // Existing pending intent remains addressable at the boundary without expanding intake.
 seeded:=prefix+"1@example.test"
 if err:=svc.BeginEmailRegistration(ctx,seeded);err!=nil{t.Fatalf("existing intent falsely rejected: %v",err)}
 if _,err:=db.Exec("UPDATE identity.registration_intents SET expires_at=statement_timestamp()-interval '1 second' WHERE email_normalized=$1",seeded);err!=nil{t.Fatal(err)}
 if err:=svc.BeginEmailRegistration(ctx,blocked);err!=nil{t.Fatalf("one freed slot could not recover intake: %v",err)}
 if err:=db.QueryRow("SELECT count(*) FROM identity.registration_intents WHERE email_normalized=$1 AND intent_state='pending'",blocked).Scan(&created);err!=nil||created!=1{t.Fatalf("freed capacity did not admit exactly one: %d %v",created,err)}
 if err:=db.QueryRow("SELECT count(*) FROM identity.users WHERE email_normalized=$1",blocked).Scan(&users);err!=nil||users!=0{t.Fatalf("unproved registration became active: %d %v",users,err)}
 t.Logf("SATURATION_CAPACITY=10000 SEED_ROWS=%d OVERFLOW_ADMITTED=NO ACKNOWLEDGEMENT_POLICY=GENERIC CAPACITY_RECOVERY=PASS AUTH_BEFORE_PROOF=NO",inserted)
 _=auth.ErrInvalidVerification
}
