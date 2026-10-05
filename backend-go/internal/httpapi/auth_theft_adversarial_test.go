package httpapi

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"

	"github.com/openinvest/openinvest/backend-go/internal/auth"
)

type theftSessionDTO struct {
	AccessToken string
	CSRFToken   string
}

func registerTheftSession(t *testing.T, app *fiber.App, email string, incomingRefreshCookie string) (*http.Cookie, theftSessionDTO) {
	t.Helper()
	response := authRequest(t, app, http.MethodPost, "/api/v1/auth/register", `{
		"email":"`+email+`",
		"password":"correct horse battery staple",
		"language":"en",
		"theme":"system",
		"timezone":"UTC"
	}`, incomingRefreshCookie, "")
	defer response.Body.Close()
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("register theft fixture status=%d", response.StatusCode)
	}
	cookie := requireCookie(t, response, auth.RefreshCookieName)
	var payload struct {
		Data struct {
			Session struct {
				AccessToken string `json:"accessToken"`
				CSRFToken   string `json:"csrfToken"`
			} `json:"session"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode theft registration response: %v", err)
	}
	if payload.Data.Session.AccessToken == "" || payload.Data.Session.CSRFToken == "" {
		t.Fatalf("registration did not issue complete client session")
	}
	return cookie, theftSessionDTO{
		AccessToken: payload.Data.Session.AccessToken,
		CSRFToken:   payload.Data.Session.CSRFToken,
	}
}

func loginTheftSession(t *testing.T, app *fiber.App, email string) (*http.Cookie, theftSessionDTO) {
	t.Helper()
	response := authRequest(t, app, http.MethodPost, "/api/v1/auth/login", `{
		"email":"`+email+`",
		"password":"correct horse battery staple"
	}`, "", "")
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("login theft fixture status=%d", response.StatusCode)
	}
	cookie := requireCookie(t, response, auth.RefreshCookieName)
	var payload struct {
		Data struct {
			Session struct {
				AccessToken string `json:"accessToken"`
				CSRFToken   string `json:"csrfToken"`
			} `json:"session"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode theft login response: %v", err)
	}
	return cookie, theftSessionDTO{
		AccessToken: payload.Data.Session.AccessToken,
		CSRFToken:   payload.Data.Session.CSRFToken,
	}
}

func TestAuthTheftStolenRefreshCookieAloneCannotRotateSession(t *testing.T) {
	app := newHTTPAuthApp(t, newHTTPAuthService(t, &httpAuthTestStore{}))
	cookie, session := registerTheftSession(t, app, "theft-cookie-only@example.com", "")

	for _, csrf := range []string{"", "stolen-client-wrong-csrf", session.CSRFToken + "-tampered"} {
		response := authRequest(t, app, http.MethodPost, "/api/v1/auth/refresh", "", cookie.Value, csrf)
		status := response.StatusCode
		response.Body.Close()
		if status != http.StatusUnauthorized {
			t.Fatalf("stolen cookie with csrf=%q was accepted: status=%d", csrf, status)
		}
	}

	// Failed theft attempts must not destroy the legitimate active session.
	legitimate := authRequest(t, app, http.MethodPost, "/api/v1/auth/refresh", "", cookie.Value, session.CSRFToken)
	defer legitimate.Body.Close()
	if legitimate.StatusCode != http.StatusOK {
		t.Fatalf("legitimate session was damaged by rejected theft attempts: status=%d", legitimate.StatusCode)
	}
}

func TestAuthTheftCrossSessionCookieCSRFMixingFailsClosed(t *testing.T) {
	store := &httpAuthTestStore{}
	app := newHTTPAuthApp(t, newHTTPAuthService(t, store))
	rootCookie, root := registerTheftSession(t, app, "theft-mix@example.com", "")
	loginCookie, independent := loginTheftSession(t, app, "theft-mix@example.com")

	cases := []struct {
		name   string
		cookie string
		csrf   string
	}{
		{"root cookie with independent csrf", rootCookie.Value, independent.CSRFToken},
		{"independent cookie with root csrf", loginCookie.Value, root.CSRFToken},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			response := authRequest(t, app, http.MethodPost, "/api/v1/auth/refresh", "", tc.cookie, tc.csrf)
			defer response.Body.Close()
			if response.StatusCode != http.StatusUnauthorized {
				t.Fatalf("cross-session cookie/csrf pair was accepted: status=%d", response.StatusCode)
			}
		})
	}

	for _, pair := range []struct {
		cookie string
		csrf   string
	}{
		{rootCookie.Value, root.CSRFToken},
		{loginCookie.Value, independent.CSRFToken},
	} {
		response := authRequest(t, app, http.MethodPost, "/api/v1/auth/refresh", "", pair.cookie, pair.csrf)
		status := response.StatusCode
		response.Body.Close()
		if status != http.StatusOK {
			t.Fatalf("valid session pair stopped working after cross-mixing rejection: status=%d", status)
		}
	}
}

func TestAuthTheftRotatedAndLoggedOutCookiesCannotBeReused(t *testing.T) {
	t.Run("rotation replay", func(t *testing.T) {
		app := newHTTPAuthApp(t, newHTTPAuthService(t, &httpAuthTestStore{}))
		cookie, session := registerTheftSession(t, app, "theft-rotation@example.com", "")

		rotated := authRequest(t, app, http.MethodPost, "/api/v1/auth/refresh", "", cookie.Value, session.CSRFToken)
		if rotated.StatusCode != http.StatusOK {
			rotated.Body.Close()
			t.Fatalf("initial rotation failed: status=%d", rotated.StatusCode)
		}
		rotated.Body.Close()

		for i := 0; i < 20; i++ {
			replay := authRequest(t, app, http.MethodPost, "/api/v1/auth/refresh", "", cookie.Value, session.CSRFToken)
			status := replay.StatusCode
			replay.Body.Close()
			if status != http.StatusUnauthorized {
				t.Fatalf("stolen pre-rotation pair replay %d accepted: status=%d", i, status)
			}
		}
	})

	t.Run("logout replay", func(t *testing.T) {
		app := newHTTPAuthApp(t, newHTTPAuthService(t, &httpAuthTestStore{}))
		cookie, session := registerTheftSession(t, app, "theft-logout@example.com", "")

		logout := authRequest(t, app, http.MethodPost, "/api/v1/auth/logout", `{"allSessions":false}`, cookie.Value, session.CSRFToken)
		if logout.StatusCode != http.StatusOK {
			logout.Body.Close()
			t.Fatalf("logout failed: status=%d", logout.StatusCode)
		}
		logout.Body.Close()

		for i := 0; i < 20; i++ {
			replay := authRequest(t, app, http.MethodPost, "/api/v1/auth/refresh", "", cookie.Value, session.CSRFToken)
			status := replay.StatusCode
			replay.Body.Close()
			if status != http.StatusUnauthorized {
				t.Fatalf("stolen logged-out pair replay %d accepted: status=%d", i, status)
			}
		}
	})
}

func TestAuthTheftCookieFixationCannotChooseServerRefreshToken(t *testing.T) {
	app := newHTTPAuthApp(t, newHTTPAuthService(t, &httpAuthTestStore{}))
	const attackerChosen = "attacker-chosen-refresh-cookie-value"

	cookie, _ := registerTheftSession(t, app, "theft-fixation@example.com", attackerChosen)
	if cookie.Value == attackerChosen {
		t.Fatal("registration reused attacker-controlled refresh cookie")
	}
	if cookie.Value == "" {
		t.Fatal("registration failed to replace attacker-controlled cookie")
	}

	login := authRequest(t, app, http.MethodPost, "/api/v1/auth/login", `{
		"email":"theft-fixation@example.com",
		"password":"correct horse battery staple"
	}`, attackerChosen, "")
	defer login.Body.Close()
	if login.StatusCode != http.StatusOK {
		t.Fatalf("login with fixation cookie failed unexpectedly: status=%d", login.StatusCode)
	}
	loginCookie := requireCookie(t, login, auth.RefreshCookieName)
	if loginCookie.Value == attackerChosen || loginCookie.Value == cookie.Value {
		t.Fatal("login failed to issue an independent server-generated refresh token")
	}
}

func TestAuthTheftRefreshCookieAttributesRemainStrict(t *testing.T) {
	app := newHTTPAuthApp(t, newHTTPAuthService(t, &httpAuthTestStore{}))
	cookie, _ := registerTheftSession(t, app, "theft-cookie-attrs@example.com", "")

	if !cookie.HttpOnly {
		t.Fatal("refresh cookie lost HttpOnly")
	}
	if !cookie.Secure {
		t.Fatal("refresh cookie lost Secure")
	}
	if cookie.SameSite != http.SameSiteStrictMode {
		t.Fatalf("refresh cookie SameSite=%v want Strict", cookie.SameSite)
	}
	if cookie.Path != "/api/v1/auth" {
		t.Fatalf("refresh cookie path=%q want /api/v1/auth", cookie.Path)
	}
	if cookie.MaxAge <= 0 {
		t.Fatalf("refresh cookie max-age must be positive, got %d", cookie.MaxAge)
	}
}
