package postgres_test

// Registration state-machine attack combinations not covered by the accepted
// PR #261 three-way generation test. Real PostgreSQL and synthetic identities.
import (
 "context"
 "errors"
 "fmt"
 "strings"
 "sync"
 "testing"

 "github.com/google/uuid"
 "github.com/openinvest/openinvest/backend-go/internal/auth"
)

func TestModule1ContinuationActiveAccountReRegistrationCannotTransferFinancialIdentity(t *testing.T){
 chaosAllowed(t)
 service,db:=ownershipService(t)
 independentDrainPreviousDeliveries(t,service)
 ctx:=context.Background()
 original:=ownerRequest()
 original.Email="continuation-existing-"+uuid.NewString()+"@example.test"
 owner,err:=service.Register(ctx,original)
 if err!=nil{t.Fatal(err)}
 oldID:=owner.User.ID
 var originalSubject string
 if err=db.QueryRow("SELECT investment_subject_id::text FROM identity.user_investment_links WHERE user_id=$1",oldID).Scan(&originalSubject);err!=nil{t.Fatal(err)}
 startSessions:=chaosCount(t,db,"SELECT count(*) FROM identity.sessions WHERE user_id=$1",oldID)
 const n=8
 start:=make(chan struct{})
 errCh:=make(chan error,n)
 var wg sync.WaitGroup
 for i:=0;i<n;i++{
  wg.Add(1)
  go func(index int){defer wg.Done();<-start
   email:=original.Email
   if index%2==0{email="  "+strings.ToUpper(email)+"  "}
   errCh<-service.BeginEmailRegistration(ctx,email)
  }(i)
 }
 close(start);wg.Wait();close(errCh)
 for err:=range errCh{if err!=nil{t.Fatalf("GENERIC_REREGISTRATION_ERROR: %v",err)}}
 intentCount:=chaosCount(t,db,"SELECT count(*) FROM identity.registration_intents WHERE email_normalized=$1",strings.ToLower(original.Email))
 if intentCount>1{t.Fatalf("DUPLICATE_REGISTRATION_INTENT_FROM_EXISTING_ACCOUNT count=%d",intentCount)}
 // A token sent to the existing user's address must not reset that user's
 // existing password or transfer their canonical financial subject.
 if intentCount==1{
  sink:=&ownershipMail{}
  delivered,e:=service.DeliverOne(ctx,sink)
  if e!=nil{t.Fatalf("EXISTING_ACCOUNT_MAIL_DELIVERY_FAILED: %v",e)}
  if delivered{
   if sink.email!=original.Email||sink.token==""{t.Fatal("INCORRECT_DELIVERY_CONTROL")}
   replacement:=ownerRequest()
   replacement.Password="new-unauthorized-owner-password-2026"
   if err=service.CompleteEmailRegistration(ctx,sink.token,replacement);err==nil{
    t.Fatal("EXISTING_ACCOUNT_TAKEN_OVER_BY_EMAIL_PROOF")
   }
  }
 }
 if count:=chaosCount(t,db,"SELECT count(*) FROM identity.users WHERE email_normalized=$1",original.Email);count!=1{t.Fatalf("ACTIVE_ACCOUNT_DUPLICATED count=%d",count)}
 var userID,subject string
 if err=db.QueryRow("SELECT u.id::text,l.investment_subject_id::text FROM identity.users u JOIN identity.user_investment_links l ON l.user_id=u.id WHERE u.email_normalized=$1",original.Email).Scan(&userID,&subject);err!=nil{t.Fatal(err)}
 if userID!=oldID||subject!=originalSubject{t.Fatalf("FINANCIAL_SUBJECT_TRANSFERED user_same=%t subject_same=%t",userID==oldID,subject==originalSubject)}
 if count:=chaosCount(t,db,"SELECT count(*) FROM identity.sessions WHERE user_id=$1",oldID);count!=startSessions{t.Fatalf("ATTACK_CREATED_OR_REVOKED_SESSION before=%d after=%d",startSessions,count)}
 legitimate,err:=service.Login(ctx,auth.LoginRequest{Email:original.Email,Password:original.Password})
 if err!=nil||legitimate.User.ID!=oldID{t.Fatalf("ORIGINAL_PASSWORD_NO_LONGER_VALID: %v",err)}
 if _,err=service.Login(ctx,auth.LoginRequest{Email:original.Email,Password:"new-unauthorized-owner-password-2026"});!errors.Is(err,auth.ErrInvalidCredentials){t.Fatalf("ATTACKER_PASSWORD_ACCEPTED_OR_UNEXPECTED_ERROR: %v",err)}
 t.Logf("REAL_POSTGRES_CONCURRENT_REREGISTRATION=%d SINGLE_OR_ZERO_PENDING_INTENTS=%d EXISTING_USER_ID=UNCHANGED FINANCIAL_SUBJECT=UNCHANGED PASSWORD_TAKEOVER=DENIED NO_PREMATURE_SESSION=YES LEGITIMATE_LOGIN=PASS",n,intentCount)
}

func TestModule1ContinuationCompetingValidProofAndMalformedProofs(t *testing.T){
 chaosAllowed(t)
 service,db:=ownershipService(t)
 independentDrainPreviousDeliveries(t,service)
 ctx:=context.Background()
 email:="continuation-proof-race-"+uuid.NewString()+"@example.test"
 if err:=service.BeginEmailRegistration(ctx,email);err!=nil{t.Fatal(err)}
 sink:=&ownershipMail{}
 delivered,err:=service.DeliverOne(ctx,sink)
 if err!=nil||!delivered||sink.email!=email||sink.token==""{t.Fatalf("LEGITIMATE_MAIL_FIXTURE_UNAVAILABLE delivered=%t err=%v",delivered,err)}
 if n:=chaosCount(t,db,"SELECT count(*) FROM identity.users WHERE email_normalized=$1",email);n!=0{t.Fatalf("PREPROOF_USER_EXISTED n=%d",n)}
 var dbHash string
 if err=db.QueryRow("SELECT t.token_hash FROM identity.email_verification_tokens t JOIN identity.registration_intents i ON t.intent_id=i.id WHERE i.email_normalized=$1",email).Scan(&dbHash);err!=nil{t.Fatal(err)}
 if dbHash==sink.token||strings.Contains(dbHash,sink.token){t.Fatal("RAW_VERIFICATION_TOKEN_PERSISTED")}
 const invalid=7
 start:=make(chan struct{})
 errCh:=make(chan error,invalid+1)
 var wg sync.WaitGroup
 for i:=0;i<invalid;i++{
  wg.Add(1)
  go func(n int){defer wg.Done();<-start
   token:=sink.token
   index:=n%len(token)
   changed:=[]byte(token)
   if changed[index]=='A'{changed[index]='B'}else{changed[index]='A'}
   token=string(changed)
   e:=service.CompleteEmailRegistration(ctx,token,ownerRequest())
   if !errors.Is(e,auth.ErrInvalidVerification){errCh<-fmt.Errorf("BAD_PROOF_ACCEPTED_OR_WRONG_ERROR i=%d error=%v",n,e)}
  }(i)
 }
 wg.Add(1)
 go func(){defer wg.Done();<-start
  if e:=service.CompleteEmailRegistration(ctx,sink.token,ownerRequest());e!=nil{errCh<-fmt.Errorf("LEGITIMATE_PROOF_REJECTED: %w",e)}
 }()
 close(start);wg.Wait();close(errCh)
 for e:=range errCh{t.Error(e)}
 if n:=chaosCount(t,db,"SELECT count(*) FROM identity.users WHERE email_normalized=$1 AND account_state='active'",email);n!=1{t.Fatalf("PROOF_RACE_ACCOUNT_COUNT=%d",n)}
 if n:=chaosCount(t,db,"SELECT count(*) FROM identity.sessions s JOIN identity.users u ON s.user_id=u.id WHERE u.email_normalized=$1",email);n!=0{t.Fatalf("PROOF_RACE_AUTO_SESSION_COUNT=%d",n)}
 if e:=service.CompleteEmailRegistration(ctx,sink.token,ownerRequest());!errors.Is(e,auth.ErrInvalidVerification){t.Fatalf("CONSUMED_PROOF_REPLAY_ACCEPTED: %v",e)}
 t.Log("REAL_POSTGRES_PROOF_RACE=8_CONCURRENT_OPERATIONS ATTACKER_MUTATED_PROOFS=7 ALL_DENIED=YES REAL_EMAIL_PROOF_SINGLE_ACTIVATION=YES DIGEST_ONLY_STORAGE=YES AUTOMATIC_SESSION=NO CONSUMED_REPLAY=DENIED")
}
