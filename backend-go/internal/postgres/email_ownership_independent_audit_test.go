package postgres_test

import (
    "context"
    "crypto/sha256"
    "encoding/hex"
    "errors"
    "fmt"
    "os"
    "sync"
    "testing"

    "github.com/openinvest/openinvest/backend-go/internal/auth"
    "github.com/openinvest/openinvest/backend-go/internal/postgres"
)

// Independent adversarial tests: test-only verification branch, never production.
func requireIndependentPG(t *testing.T) {
    t.Helper()
    if os.Getenv("OPENINVEST_DATABASE_TEST_URL") == "" {
        t.Fatal("NOT_EXECUTED: independent PostgreSQL test URL is mandatory")
    }
}

func TestEmailOwnershipIndependentLegacyAccountCannotBeReclaimed(t *testing.T) {
    requireIndependentPG(t)
    service, db := ownershipService(t)
    ctx := context.Background()
    email := "independent-legacy@example.test"
    existing := ownerRequest()
    existing.Email = email
    existing.Password = "legacy-owner-private-password-2026"
    registered, err := service.Register(ctx, existing) // Controlled legacy-fixture seed, NOT public registration.
    if err != nil { t.Fatal(err) }
    var initialHash, initialSubject string
    if err := db.QueryRow("SELECT c.password_hash, l.investment_subject_id FROM identity.credentials c JOIN identity.user_investment_links l ON l.user_id=c.user_id WHERE c.user_id=$1", registered.User.ID).Scan(&initialHash, &initialSubject); err != nil {t.Fatal(err)}
    var initialSessions int
    if err := db.QueryRow("SELECT count(*) FROM identity.sessions WHERE user_id=$1", registered.User.ID).Scan(&initialSessions); err != nil { t.Fatal(err) }
    if initialSessions == 0 {t.Fatal("negative control invalid: legacy session absent")}
    if err := service.BeginEmailRegistration(ctx,email); err != nil {t.Fatal(err)}
    sink := &ownershipMail{}
    if ok,err := service.DeliverOne(ctx,sink); err != nil || !ok || sink.email != email {t.Fatalf("wrong delivery target %q: %t %v",sink.email,ok,err)}
    if err := service.CompleteEmailRegistration(ctx,sink.token,ownerRequest()); !errors.Is(err,auth.ErrInvalidVerification) {t.Fatalf("existing account takeover allowed or unexpected error: %v",err)}
    var id,hash,subject string
    if err := db.QueryRow("SELECT u.id, c.password_hash, l.investment_subject_id FROM identity.users u JOIN identity.credentials c ON c.user_id=u.id JOIN identity.user_investment_links l ON l.user_id=u.id WHERE u.email_normalized=$1", email).Scan(&id,&hash,&subject); err != nil {t.Fatal(err)}
    var sessions int
    if err := db.QueryRow("SELECT count(*) FROM identity.sessions WHERE user_id=$1",registered.User.ID).Scan(&sessions); err != nil {t.Fatal(err)}
    if id != registered.User.ID || hash != initialHash || subject != initialSubject || sessions != initialSessions {
        t.Fatalf("legacy identity/credential/subject/session changed: id_same=%t hash_same=%t subject_same=%t sessions=%d/%d",id==registered.User.ID,hash==initialHash,subject==initialSubject,sessions,initialSessions)
    }
    t.Log("INDEPENDENT_LEGACY_ACCOUNT_ISOLATION=PASS")
}

func TestEmailOwnershipIndependentGenerationInvalidatesOldToken(t *testing.T) {
    requireIndependentPG(t)
    service, db := ownershipService(t)
    ctx := context.Background()
    email := "independent-generation@example.test"
    if err := service.BeginEmailRegistration(ctx,email); err != nil {t.Fatal(err)}
    old := &ownershipMail{}
    if ok,err := service.DeliverOne(ctx,old); err != nil || !ok || old.email != email {t.Fatalf("initial delivery %t %v %q",ok,err,old.email)}
    if _,err := db.Exec("UPDATE identity.registration_intents SET expires_at=statement_timestamp()-interval '1 second' WHERE email_normalized=$1",email); err != nil {t.Fatal(err)}
    if err := service.BeginEmailRegistration(ctx,email); err != nil {t.Fatal(err)}
    if err := service.CompleteEmailRegistration(ctx,old.token,ownerRequest()); !errors.Is(err,auth.ErrInvalidVerification) {t.Fatalf("superseded token accepted: %v",err)}
    newer := &ownershipMail{}
    if ok,err := service.DeliverOne(ctx,newer); err != nil || !ok || newer.email != email {t.Fatalf("new delivery %t %v %q",ok,err,newer.email)}
    if newer.token == old.token {t.Fatal("token replayed by worker")}
    if err := service.CompleteEmailRegistration(ctx,newer.token,ownerRequest()); err != nil {t.Fatalf("new generation unable to activate: %v",err)}
    var n int
    if err := db.QueryRow("SELECT count(*) FROM identity.users WHERE email_normalized=$1",email).Scan(&n); err != nil || n!=1 {t.Fatalf("not exactly one account: %d %v",n,err)}
    t.Log("INDEPENDENT_TOKEN_GENERATION=PASS")
}

func TestEmailOwnershipIndependentConcurrentTokenConsumption(t *testing.T) {
    requireIndependentPG(t)
    service, db := ownershipService(t)
    ctx := context.Background()
    email := "independent-concurrency@example.test"
    if err := service.BeginEmailRegistration(ctx,email); err != nil {t.Fatal(err)}
    sink := &ownershipMail{}
    if ok,err := service.DeliverOne(ctx,sink); err != nil || !ok || sink.email != email {t.Fatalf("delivery %t %v %q",ok,err,sink.email)}
    const attempts=2
    start:=make(chan struct{})
    var wg sync.WaitGroup
    result:=make(chan error,attempts)
    for i:=0;i<attempts;i++ {wg.Add(1); go func(){defer wg.Done();<-start;result<-service.CompleteEmailRegistration(ctx,sink.token,ownerRequest())}()}
    close(start);wg.Wait();close(result)
    success,invalid:=0,0
    for err:=range result {if err==nil {success++} else if errors.Is(err,auth.ErrInvalidVerification) {invalid++} else {t.Fatalf("inconclusive admission/infrastructure outcome: %v",err)}}
    var users int
    if err := db.QueryRow("SELECT count(*) FROM identity.users WHERE email_normalized=$1",email).Scan(&users); err != nil {t.Fatal(err)}
    if success!=1 || invalid!=1 || users!=1 {t.Fatalf("atomic single-consumption failed: success=%d invalid=%d users=%d",success,invalid,users)}
    t.Log("INDEPENDENT_ATOMIC_ACTIVATION=PASS")
}

// Property oracle: after the last lease expires, a restart must not leave a
// permanently unclaimable job marked 'leased'. No transport/network is called.
func TestEmailOwnershipIndependentFourthCrashTerminalState(t *testing.T) {
    requireIndependentPG(t)
    service, db := ownershipService(t)
    ctx := context.Background()
    email := "independent-final-crash@example.test"
    if err := service.BeginEmailRegistration(ctx,email); err != nil {t.Fatal(err)}
    worker,err := postgres.Open(os.Getenv("OPENINVEST_DATABASE_TEST_URL"))
    if err != nil {t.Fatal(err)}
    defer worker.Close()
    for i:=1;i<=4;i++ {
        sum:=sha256.Sum256([]byte(fmt.Sprintf("independent-crash-%d",i)))
        lease,found,err:=worker.LeaseVerificationDelivery(ctx,hex.EncodeToString(sum[:]))
        if err!=nil || !found || lease.Email!=email {t.Fatalf("lease %d failed: found=%t email=%q err=%v",i,found,lease.Email,err)}
        // Deterministically simulate a crash without FinishVerificationDelivery,
        // and advance the lease expiration in the disposable PostgreSQL fixture.
        if _,err:=db.Exec("UPDATE identity.email_delivery_jobs SET lease_until=statement_timestamp()-interval '1 second',available_at=statement_timestamp()-interval '1 second' WHERE id=$1",lease.JobID); err!=nil {t.Fatal(err)}
    }
    sum:=sha256.Sum256([]byte("independent-fifth-attempt"))
    _,found,err:=worker.LeaseVerificationDelivery(ctx,hex.EncodeToString(sum[:]))
    if err!=nil {t.Fatal(err)}
    var state string
    var attempts int
    if err:=db.QueryRow("SELECT j.job_state,j.attempts FROM identity.email_delivery_jobs j JOIN identity.registration_intents i ON i.id=j.intent_id WHERE i.email_normalized=$1",email).Scan(&state,&attempts);err!=nil{t.Fatal(err)}
    t.Logf("INDEPENDENT_FOURTH_CRASH_COUNTEREXAMPLE state=%s attempts=%d eligible_for_recovery=%t",state,attempts,found)
    if found || attempts!=4 {t.Fatalf("retry budget oracle mismatch: found=%t attempts=%d",found,attempts)}
    if state=="leased" {t.Fatal("PRODUCT_RELIABILITY_DEFECT: exhausted lease remains 'leased' forever until intent renewal/cleanup; worker cannot retry or terminally mark failed")}
    if state!="failed" {t.Fatalf("unexpected terminal state %q",state)}
}
