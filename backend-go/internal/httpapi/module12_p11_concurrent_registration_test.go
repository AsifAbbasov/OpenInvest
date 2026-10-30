package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/openinvest/openinvest/backend-go/internal/auth"
)

type module12RegistrationProbe struct {
	httpOwnershipFixture
	calls atomic.Int64
}

func (s *module12RegistrationProbe) BeginRegistration(context.Context, string) error {
	s.calls.Add(1)
	return nil
}

func TestModule12P11ConcurrentEmailRegistrationAcknowledgmentsHaveNoSessions(t *testing.T) {
	store := &module12RegistrationProbe{}
	service, err := auth.NewService(store, fixedHTTPClock{}, auth.Config{
		AccessTokenSecret:   []byte("01234567890123456789012345678901"),
		RefreshCookieSecure: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	app := newHTTPAuthApp(t, service)
	const attempts = 6
	var wg sync.WaitGroup
	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(`{"email":"module12-p11@example.test"}`))
			request.Header.Set("Content-Type", "application/json")
			response, err := app.Test(request)
			if err != nil {
				t.Errorf("parallel registration: %v", err)
				return
			}
			defer response.Body.Close()
			if response.StatusCode != http.StatusAccepted || len(response.Cookies()) != 0 || response.Header.Get("Cache-Control") != "no-store" {
				t.Errorf("registration granted an unexpected session or response: status=%d cookies=%d", response.StatusCode, len(response.Cookies()))
				return
			}
			var payload struct {
				Data map[string]any `json:"data"`
			}
			if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
				t.Errorf("decode registration acknowledgment: %v", err)
				return
			}
			if payload.Data["status"] != "REQUEST_ACCEPTED" || len(payload.Data) != 2 {
				t.Errorf("registration disclosed identity/session material")
			}
		}()
	}
	wg.Wait()
	if got := store.calls.Load(); got != attempts {
		t.Fatalf("expected %d independent registration acknowledgments, got %d", attempts, got)
	}
	if store.registerCalls != 0 || len(store.sessions) != 0 {
		t.Fatal("registration unexpectedly created identity or session before email proof")
	}
}
