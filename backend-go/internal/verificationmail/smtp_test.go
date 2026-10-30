package verificationmail

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"math/big"
	"net"
	"strings"
	"testing"
	"time"
)

type localCA struct {
	cert *x509.Certificate
	key  *ecdsa.PrivateKey
	pool *x509.CertPool
}

func makeLocalCA(t *testing.T) localCA {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "OpenInvest SMTP test CA"},
		NotBefore: now.Add(-time.Minute), NotAfter: now.Add(time.Hour),
		IsCA: true, BasicConstraintsValid: true,
		KeyUsage: x509.KeyUsageCertSign,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	pool := x509.NewCertPool()
	pool.AddCert(cert)
	return localCA{cert: cert, key: key, pool: pool}
}

func makeLocalServerCertificate(t *testing.T, authority localCA, address string) tls.Certificate {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(2), Subject: pkix.Name{CommonName: "SMTP test"},
		NotBefore: time.Now().Add(-time.Minute), NotAfter: time.Now().Add(time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: []net.IP{net.ParseIP(address)},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, authority.cert, &key.PublicKey, authority.key)
	if err != nil {
		t.Fatal(err)
	}
	return tls.Certificate{Certificate: [][]byte{der, authority.cert.Raw}, PrivateKey: key}
}

func startLocalSMTP(t *testing.T, cert tls.Certificate) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
		_, _ = fmt.Fprint(conn, "220 localhost ESMTP\r\n")
		reader := bufio.NewReader(conn)
		secure := false
		inData := false
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			command := strings.TrimSpace(line)
			if inData {
				if command == "." {
					inData = false
					_, _ = fmt.Fprint(conn, "250 queued\r\n")
				}
				continue
			}
			switch {
			case strings.HasPrefix(command, "EHLO "):
				if secure {
					_, _ = fmt.Fprint(conn, "250 localhost\r\n")
				} else {
					_, _ = fmt.Fprint(conn, "250-localhost\r\n250 STARTTLS\r\n")
				}
			case command == "STARTTLS":
				_, _ = fmt.Fprint(conn, "220 ready\r\n")
				upgraded := tls.Server(conn, &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12})
				if err := upgraded.Handshake(); err != nil {
					return
				}
				conn, reader, secure = upgraded, bufio.NewReader(upgraded), true
			case strings.HasPrefix(command, "MAIL FROM:") || strings.HasPrefix(command, "RCPT TO:"):
				_, _ = fmt.Fprint(conn, "250 OK\r\n")
			case command == "DATA":
				inData = true
				_, _ = fmt.Fprint(conn, "354 end with dot\r\n")
			case command == "QUIT":
				_, _ = fmt.Fprint(conn, "221 goodbye\r\n")
				return
			default:
				_, _ = fmt.Fprint(conn, "500 unsupported\r\n")
				return
			}
		}
	}()
	t.Cleanup(func() {
		_ = listener.Close()
		select {
		case <-done:
		case <-time.After(4 * time.Second):
			t.Error("SMTP test listener did not terminate")
		}
	})
	_, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	return port
}

func localSMTP(port string, roots *x509.CertPool) SMTP {
	return SMTP{
		Host: "127.0.0.1", Port: port, From: "sender@example.test",
		VerificationURL: "https://example.test/verify", rootCAs: roots,
	}
}

func TestSMTPStartTLSCertificateVerification(t *testing.T) {
	trusted := makeLocalCA(t)
	other := makeLocalCA(t)
	for _, tc := range []struct {
		name      string
		cert      tls.Certificate
		wantError bool
	}{
		{name: "trusted_cert", cert: makeLocalServerCertificate(t, trusted, "127.0.0.1")},
		{name: "untrusted_cert", cert: makeLocalServerCertificate(t, other, "127.0.0.1"), wantError: true},
		{name: "wrong_hostname", cert: makeLocalServerCertificate(t, trusted, "127.0.0.2"), wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			port := startLocalSMTP(t, tc.cert)
			err := localSMTP(port, trusted.pool).SendVerification(context.Background(), "person@example.test", "local-proof")
			if tc.wantError && !errors.Is(err, ErrDelivery) {
				t.Fatalf("expected ErrDelivery for rejected TLS certificate, got %v", err)
			}
			if !tc.wantError && err != nil {
				t.Fatalf("trusted TLS certificate was rejected: %v", err)
			}
		})
	}
}

func TestSMTPStartTLSCancelledAndDeadline(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := localSMTP("1", nil).SendVerification(ctx, "person@example.test", "proof"); !errors.Is(err, ErrDelivery) {
		t.Fatalf("cancelled context: expected ErrDelivery, got %v", err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err == nil {
			defer conn.Close()
			_, _ = fmt.Fprint(conn, "220 localhost ESMTP\r\n")
			// Intentionally do not answer EHLO to exercise the in-flight deadline.
			_, _ = bufio.NewReader(conn).ReadString('\n')
			var one [1]byte
			_, _ = conn.Read(one[:])
		}
	}()
	_, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	deadlineCtx, deadlineCancel := context.WithTimeout(context.Background(), 75*time.Millisecond)
	defer deadlineCancel()
	if err := localSMTP(port, nil).SendVerification(deadlineCtx, "person@example.test", "proof"); !errors.Is(err, ErrDelivery) {
		t.Fatalf("in-flight deadline: expected ErrDelivery, got %v", err)
	}
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("deadline did not close the SMTP connection")
	}
}
