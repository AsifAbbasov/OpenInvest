package moexiss

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type module10MOEXClock struct{ now time.Time }

func (c module10MOEXClock) Now() time.Time { return c.now }

func module10MOEXProvider(t *testing.T, h http.Handler) (*Provider, *httptest.Server) {
	t.Helper()
	s := httptest.NewServer(h)
	p, err := newQuoteProvider(
		s.Client(),
		module10MOEXClock{now: time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC)},
		s.URL,
	)
	if err != nil {
		t.Fatal(err)
	}
	return p, s
}

func TestModule10MOEXDormantAdapterSlowProviderAndBounds(t *testing.T) {
	t.Run("caller_deadline_and_upstream_cancel", func(t *testing.T) {
		cancelSeen := make(chan struct{}, 1)
		p, s := module10MOEXProvider(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			<-r.Context().Done()
			cancelSeen <- struct{}{}
		}))
		defer s.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
		defer cancel()
		_, err := p.Quote(ctx, "SBER")
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("deadline not preserved: %v", err)
		}
		select {
		case <-cancelSeen:
		case <-time.After(time.Second):
			t.Fatal("upstream cancellation not observed")
		}
	})

	t.Run("oversized_response", func(t *testing.T) {
		body := strings.Repeat("x", int(maxResponseBodyBytes)+1)
		p, s := module10MOEXProvider(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.WriteString(w, body)
		}))
		defer s.Close()
		if _, err := p.Quote(context.Background(), "SBER"); err == nil {
			t.Fatal("oversized MOEX response accepted")
		}
	})

	t.Run("redirect_not_followed", func(t *testing.T) {
		reached := make(chan struct{}, 1)
		target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reached <- struct{}{}
		}))
		defer target.Close()

		p, source := module10MOEXProvider(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, target.URL, http.StatusFound)
		}))
		defer source.Close()
		if _, err := p.Quote(context.Background(), "SBER"); err == nil {
			t.Fatal("redirect response accepted")
		}
		select {
		case <-reached:
			t.Fatal("redirect target reached")
		default:
		}
	})
}
