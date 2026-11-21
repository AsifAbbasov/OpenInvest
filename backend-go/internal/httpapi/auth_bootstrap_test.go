package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/openinvest/openinvest/backend-go/internal/auth"
)

const bootstrapTestPath = "/api/v1/auth/bootstrap"

func bootstrapTestRequest(
	t *testing.T,
	app *fiber.App,
	cookie, origin, marker, accept, body, fetchSite string,
) *http.Response {
	t.Helper()

	request := httptest.NewRequest(
		http.MethodPost,
		bootstrapTestPath,
		strings.NewReader(body),
	)

	if cookie != "" {
		request.AddCookie(&http.Cookie{
			Name:  auth.RefreshCookieName,
			Value: cookie,
		})
	}
	if origin != "" {
		request.Header.Set("Origin", origin)
	}
	if marker != "" {
		request.Header.Set("X-OpenInvest-Bootstrap", marker)
	}
	if accept != "" {
		request.Header.Set("Accept", accept)
	}
	if fetchSite != "" {
		request.Header.Set("Sec-Fetch-Site", fetchSite)
	}

	response, err := app.Test(request, fiber.TestConfig{
		Timeout:       10 * time.Second,
		FailOnTimeout: true,
	})
	if err != nil {
		t.Fatalf("bootstrap request: %v", err)
	}

	return response
}

func TestBootstrapRestoresSessionWithoutRotatingRefreshToken(t *testing.T) {
	store := &httpAuthTestStore{}
	app := newHTTPAuthApp(t, newHTTPAuthService(t, store))

	registration := fixtureLegacyLogin(
		t,
		app,
		http.MethodPost,
		"/api/v1/auth/register",
		`{
			"email":"bootstrap@example.com",
			"password":"correct horse battery staple",
			"language":"en",
			"theme":"system",
			"timezone":"UTC"
		}`,
		"",
		"",
	)

	if registration.StatusCode != http.StatusOK {
		t.Fatalf(
			"registration status: got %d, want 200",
			registration.StatusCode,
		)
	}

	refreshCookie := requireCookie(
		t,
		registration,
		auth.RefreshCookieName,
	)
	originalCSRF := readCSRFToken(t, registration)
	registration.Body.Close()

	response := bootstrapTestRequest(
		t,
		app,
		refreshCookie.Value,
		"http://localhost:3000",
		"1",
		"application/json",
		"",
		"same-site",
	)
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf(
			"bootstrap status: got %d, want 200",
			response.StatusCode,
		)
	}

	if response.Header.Get("Set-Cookie") != "" {
		t.Fatal("bootstrap must not issue or rotate refresh cookie")
	}

	if response.Header.Get("Cache-Control") != "no-store" {
		t.Fatalf(
			"bootstrap Cache-Control: got %q, want no-store",
			response.Header.Get("Cache-Control"),
		)
	}

	var payload struct {
		Data struct {
			User struct {
				ID    string `json:"id"`
				Email string `json:"email"`
			} `json:"user"`

			Session struct {
				AccessToken string `json:"accessToken"`
				CSRFToken   string `json:"csrfToken"`
			} `json:"session"`
		} `json:"data"`
	}

	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode bootstrap response: %v", err)
	}

	if payload.Data.User.ID == "" ||
		payload.Data.User.Email != "bootstrap@example.com" {
		t.Fatalf(
			"bootstrap returned wrong user: %+v",
			payload.Data.User,
		)
	}

	if payload.Data.Session.AccessToken == "" ||
		payload.Data.Session.CSRFToken == "" {
		t.Fatal("bootstrap must return access and CSRF tokens")
	}

	if payload.Data.Session.CSRFToken == originalCSRF {
		t.Fatal("bootstrap must generate a new CSRF token")
	}

	oldCSRF := authRequest(
		t,
		app,
		http.MethodPost,
		"/api/v1/auth/refresh",
		"",
		refreshCookie.Value,
		originalCSRF,
	)
	oldCSRF.Body.Close()

	if oldCSRF.StatusCode != http.StatusUnauthorized {
		t.Fatalf(
			"old CSRF refresh: got %d, want 401",
			oldCSRF.StatusCode,
		)
	}

	currentCSRF := authRequest(
		t,
		app,
		http.MethodPost,
		"/api/v1/auth/refresh",
		"",
		refreshCookie.Value,
		payload.Data.Session.CSRFToken,
	)
	defer currentCSRF.Body.Close()

	if currentCSRF.StatusCode != http.StatusOK {
		t.Fatalf(
			"refresh after bootstrap: got %d, want 200",
			currentCSRF.StatusCode,
		)
	}

	rotated := requireCookie(
		t,
		currentCSRF,
		auth.RefreshCookieName,
	)
	if rotated.Value == refreshCookie.Value {
		t.Fatal("explicit refresh must still rotate the refresh cookie")
	}
}

func TestBootstrapRejectsInvalidRequests(t *testing.T) {
	app := newHTTPAuthApp(
		t,
		newHTTPAuthService(t, &httpAuthTestStore{}),
	)

	tests := []struct {
		name      string
		cookie    string
		origin    string
		marker    string
		accept    string
		body      string
		fetchSite string
		want      int
	}{
		{
			name:   "missing cookie",
			origin: "http://localhost:3000",
			marker: "1",
			accept: "application/json",
			want:   http.StatusUnauthorized,
		},
		{
			name:   "unknown cookie",
			cookie: "unknown-refresh",
			origin: "http://localhost:3000",
			marker: "1",
			accept: "application/json",
			want:   http.StatusUnauthorized,
		},
		{
			name:   "missing origin",
			cookie: "unknown-refresh",
			marker: "1",
			accept: "application/json",
			want:   http.StatusForbidden,
		},
		{
			name:   "disallowed origin",
			cookie: "unknown-refresh",
			origin: "https://attacker.example",
			marker: "1",
			accept: "application/json",
			want:   http.StatusForbidden,
		},
		{
			name:   "null origin",
			cookie: "unknown-refresh",
			origin: "null",
			marker: "1",
			accept: "application/json",
			want:   http.StatusForbidden,
		},
		{
			name:      "cross site",
			cookie:    "unknown-refresh",
			origin:    "http://localhost:3000",
			marker:    "1",
			accept:    "application/json",
			fetchSite: "cross-site",
			want:      http.StatusForbidden,
		},
		{
			name:   "missing marker",
			cookie: "unknown-refresh",
			origin: "http://localhost:3000",
			accept: "application/json",
			want:   http.StatusBadRequest,
		},
		{
			name:   "incorrect marker",
			cookie: "unknown-refresh",
			origin: "http://localhost:3000",
			marker: "invalid",
			accept: "application/json",
			want:   http.StatusBadRequest,
		},
		{
			name:   "missing accept",
			cookie: "unknown-refresh",
			origin: "http://localhost:3000",
			marker: "1",
			want:   http.StatusBadRequest,
		},
		{
			name:   "nonempty body",
			cookie: "unknown-refresh",
			origin: "http://localhost:3000",
			marker: "1",
			accept: "application/json",
			body:   `{}`,
			want:   http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			response := bootstrapTestRequest(
				t,
				app,
				tc.cookie,
				tc.origin,
				tc.marker,
				tc.accept,
				tc.body,
				tc.fetchSite,
			)
			defer response.Body.Close()

			if response.StatusCode != tc.want {
				t.Fatalf(
					"status: got %d, want %d",
					response.StatusCode,
					tc.want,
				)
			}

			if response.Header.Get("Set-Cookie") != "" {
				t.Fatal("failed bootstrap must not set a cookie")
			}

			if response.Header.Get("Cache-Control") != "no-store" {
				t.Fatalf(
					"Cache-Control: got %q, want no-store",
					response.Header.Get("Cache-Control"),
				)
			}
		})
	}
}
