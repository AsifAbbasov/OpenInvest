package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/auth"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func FuzzM11RouteUUIDAuthorization(f *testing.F) {
	f.Add("00000000-0000-4000-8000-000000000002", true)
	f.Add("00000000-0000-4000-8000-000000000002", false)
	f.Add("not-a-uuid", true)
	f.Add("../00000000-0000-4000-8000-000000000002", true)
	f.Add("00000000000040008000000000000002", true)

	authStore := &httpAuthTestStore{}
	authService, err := auth.NewService(authStore, fixedHTTPClock{}, auth.Config{
		AccessTokenSecret: []byte("m11-v2-route-fuzz-access-secret-32"),
		RefreshCookieSecure: true,
	})
	if err != nil { f.Fatal(err) }
	session, err := authService.Register(context.Background(), auth.RegistrationRequest{
		Email: "m11-v2-route-fuzz@example.com",
		Password: "correct horse battery staple",
		Language: auth.LanguageEN,
		Theme: auth.ThemeSystem,
		Timezone: "UTC",
	})
	if err != nil { f.Fatal(err) }

	store := &importAPITestStore{}
	app, err := New(
		verticalslice.NewService(store, fixedHTTPClock{}),
		authService,
		[]byte("m11-v2-route-fuzz-import-secret-32"),
	)
	if err != nil { f.Fatal(err) }

	f.Fuzz(func(t *testing.T, raw string, authenticated bool) {
		if len(raw) > 256 { t.Skip() }
		store.listTransactionsCalls = 0
		req := httptest.NewRequest(http.MethodGet, "/api/v1/portfolios/"+url.PathEscape(raw)+"/transactions", nil)
		if authenticated {
			req.Header.Set("Authorization", "Bearer "+session.Session.AccessToken)
		}
		resp, err := app.Test(req)
		if err != nil { t.Fatalf("request: %v", err) }
		resp.Body.Close()

		_, parseErr := uuid.Parse(raw)
		if !authenticated {
			if store.listTransactionsCalls != 0 || resp.StatusCode != http.StatusUnauthorized {
				t.Fatalf("unauthenticated raw=%q calls=%d status=%d", raw, store.listTransactionsCalls, resp.StatusCode)
			}
			return
		}
		if parseErr != nil {
			if store.listTransactionsCalls != 0 {
				t.Fatalf("invalid UUID reached store raw=%q calls=%d", raw, store.listTransactionsCalls)
			}
			if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusNotFound {
				t.Fatalf("invalid UUID status=%d raw=%q", resp.StatusCode, raw)
			}
			return
		}
		if store.listTransactionsCalls != 1 || resp.StatusCode != http.StatusOK {
			t.Fatalf("valid authenticated UUID raw=%q calls=%d status=%d", raw, store.listTransactionsCalls, resp.StatusCode)
		}
	})
}
