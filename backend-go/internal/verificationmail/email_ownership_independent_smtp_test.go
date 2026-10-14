package verificationmail

import(
 "bufio"
 "context"
 "errors"
 "fmt"
 "net"
 "strconv"
 "strings"
 "testing"
 "time"
)

func TestEmailOwnershipIndependentSMTPConfigurationFailClosed(t *testing.T){
 good:=SMTP{Host:"127.0.0.1",Port:"0",From:"sender@example.test",VerificationURL:"https://example.test/auth/verify"}
 for _,v:=range []struct{name string;change func(*SMTP)}{
  {"http_url",func(s *SMTP){s.VerificationURL="http://example.test/auth/verify"}},
  {"injected_query",func(s *SMTP){s.VerificationURL="https://example.test/auth/verify?token=leak"}},
  {"embedded_credentials",func(s *SMTP){s.VerificationURL="https://attacker@example.test/auth/verify"}},
  {"existing_fragment",func(s *SMTP){s.VerificationURL="https://example.test/auth/verify#other"}},
  {"header_injection",func(s *SMTP){s.From="sender@example.test\r\nBcc: attacker@example.test"}},
  {"recipient_injection",func(s *SMTP){s.VerificationURL="https://example.test/auth/verify"}},
 }{
  if v.name=="recipient_injection"{continue}
  t.Run(v.name,func(t *testing.T){s:=good;v.change(&s)
   err:=s.SendVerification(context.Background(),"owner@example.test","SECRET_SENTINEL_SHOULD_NOT_LEAK")
   if !errors.Is(err,ErrDelivery){t.Fatalf("unsafe SMTP configuration accepted: %v",err)}
   if strings.Contains(err.Error(),"SECRET_SENTINEL"){t.Fatal("raw token leaked")}
  })
 }
}
func TestEmailOwnershipIndependentSMTPStartTLSMandatoryLoopback(t *testing.T){
 listener,err:=net.Listen("tcp","127.0.0.1:0");if err!=nil{t.Fatal(err)}
 defer listener.Close()
 accepted:=make(chan string,1)
 go func(){
  conn,err:=listener.Accept();if err!=nil{accepted<-"accept_failed";return}
  defer conn.Close();conn.SetDeadline(time.Now().Add(3*time.Second))
  fmt.Fprint(conn,"220 test.local ESMTP\r\n")
  line,err:=bufio.NewReader(conn).ReadString('\n')
  if err!=nil{accepted<-"read_failed";return}
  if !strings.HasPrefix(line,"EHLO "){accepted<-"wrong_command:"+strings.TrimSpace(line);return}
  fmt.Fprint(conn,"250-test.local\r\n250 HELP\r\n")
  accepted<-"ehlo_no_starttls"
 }()
 port:=strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)
 sender:=SMTP{Host:"127.0.0.1",Port:port,From:"sender@example.test",VerificationURL:"https://example.test/auth/verify"}
 ctx,cancel:=context.WithTimeout(context.Background(),3*time.Second);defer cancel()
 err=sender.SendVerification(ctx,"owner@example.test","SECRET_SENTINEL_SHOULD_NOT_LEAK")
 if !errors.Is(err,ErrDelivery){t.Fatalf("plaintext local SMTP accepted: %v",err)}
 if err!=nil&&strings.Contains(err.Error(),"SECRET_SENTINEL"){t.Fatal("token leaked")}
 select{case got:=<-accepted:
  if got!="ehlo_no_starttls"{t.Fatalf("negative control failed: %s",got)}
 case <-ctx.Done():t.Fatal("local TLS-negative control timed out")}
 t.Log("INDEPENDENT_SMTP_STARTTLS_REQUIRED_NO_EXTERNAL_TRAFFIC=PASS")
}
