package main

import (
	"context"
	"flag"
	"github.com/openinvest/openinvest/backend-go/internal/auth"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verificationmail"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	enabled := flag.Bool("enable-external-delivery", false, "requires separately approved SMTP deployment configuration")
	flag.Parse()
	if !*enabled {
		log.Print("verification mail worker disabled")
		return
	}
	store, err := postgres.OpenRuntime(os.Getenv("OPENINVEST_DATABASE_URL"))
	if err != nil {
		log.Fatal("mail worker database unavailable")
	}
	defer store.Close()
	service, err := auth.NewService(store, nil, auth.Config{AllowEphemeralAccessTokenSecret: true})
	if err != nil {
		log.Fatal("mail worker configuration invalid")
	}
	mailer := verificationmail.SMTP{Host: os.Getenv("OPENINVEST_SMTP_HOST"), Port: os.Getenv("OPENINVEST_SMTP_PORT"), From: os.Getenv("OPENINVEST_SMTP_FROM"), Username: os.Getenv("OPENINVEST_SMTP_USER"), Password: os.Getenv("OPENINVEST_SMTP_PASSWORD"), VerificationURL: os.Getenv("OPENINVEST_EMAIL_VERIFICATION_URL")}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, err := service.DeliverOne(ctx, mailer)
			if err != nil {
				log.Print("verification delivery deferred")
			}
		}
	}
}
