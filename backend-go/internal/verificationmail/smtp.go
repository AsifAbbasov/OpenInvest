// Package verificationmail implements the explicitly configured email boundary.
package verificationmail

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"net/url"
	"strings"
	"time"
)

var ErrDelivery = errors.New("verification delivery failed")

type SMTP struct{ Host, Port, Username, Password, From, VerificationURL string }

func (s SMTP) SendVerification(ctx context.Context, email, token string) error {
	from, err := mail.ParseAddress(s.From)
	if err != nil || from.Address != s.From {
		return ErrDelivery
	}
	to, err := mail.ParseAddress(email)
	if err != nil || to.Address != email {
		return ErrDelivery
	}
	if strings.ContainsAny(s.From+email, "\r\n") {
		return ErrDelivery
	}
	u, err := url.Parse(s.VerificationURL)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return ErrDelivery
	}
	u.Fragment = token
	conn, err := (&net.Dialer{Timeout: 10 * time.Second}).DialContext(ctx, "tcp", net.JoinHostPort(s.Host, s.Port))
	if err != nil {
		return ErrDelivery
	}
	defer conn.Close()
	deadline := time.Now().Add(30 * time.Second)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	if err = conn.SetDeadline(deadline); err != nil {
		return ErrDelivery
	}
	client, err := smtp.NewClient(conn, s.Host)
	if err != nil {
		return ErrDelivery
	}
	defer client.Close()
	if err = client.StartTLS(&tls.Config{ServerName: s.Host, MinVersion: tls.VersionTLS12}); err != nil {
		return ErrDelivery
	}
	if s.Username != "" {
		if err = client.Auth(smtp.PlainAuth("", s.Username, s.Password, s.Host)); err != nil {
			return ErrDelivery
		}
	}
	if err = client.Mail(s.From); err != nil {
		return ErrDelivery
	}
	if err = client.Rcpt(email); err != nil {
		return ErrDelivery
	}
	w, err := client.Data()
	if err != nil {
		return ErrDelivery
	}
	_, err = fmt.Fprintf(w, "From: %s\r\nTo: %s\r\nSubject: Complete your OpenInvest registration\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\nIf you requested registration, open %s and choose your password.\r\nOtherwise ignore this message.\r\n", s.From, email, u.String())
	if err != nil {
		return ErrDelivery
	}
	if err = w.Close(); err != nil {
		return ErrDelivery
	}
	if err = client.Quit(); err != nil {
		return ErrDelivery
	}
	return nil
}
