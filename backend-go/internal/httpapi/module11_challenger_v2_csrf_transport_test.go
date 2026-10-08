package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openinvest/openinvest/backend-go/internal/auth"
)

func TestM11V2CSRFStaleCombinationMatrix(t *testing.T) {
	for _, mode := range []string{"old_cookie_new_csrf", "new_cookie_old_csrf"} {
		t.Run(mode, func(t *testing.T) {
			store := &httpAuthTestStore{}
			app := newHTTPAuthApp(t, newHTTPAuthService(t, store))
			reg := authRequest(t, app, http.MethodPost, "/api/v1/auth/register", `{
				"email":"m11-v2-`+mode+`@example.com",
				"password":"correct horse battery staple",
				"language":"en","theme":"system","timezone":"UTC"
			}`, "", "")
			oldCookie := requireCookie(t, reg, auth.RefreshCookieName)
			oldCSRF := readCSRFToken(t, reg)
			reg.Body.Close()

			rotated := authRequest(t, app, http.MethodPost, "/api/v1/auth/refresh", "", oldCookie.Value, oldCSRF)
			newCookie := requireCookie(t, rotated, auth.RefreshCookieName)
			newCSRF := readCSRFToken(t, rotated)
			rotated.Body.Close()

			cookie, csrf := oldCookie.Value, newCSRF
			if mode == "new_cookie_old_csrf" {
				cookie, csrf = newCookie.Value, oldCSRF
			}
			response := authRequest(t, app, http.MethodPost, "/api/v1/auth/refresh", "", cookie, csrf)
			response.Body.Close()
			if response.StatusCode != http.StatusUnauthorized {
				t.Fatalf("%s status=%d", mode, response.StatusCode)
			}
		})
	}
	t.Log("M11_V2_CSRF_STALE_COOKIE_TOKEN_COMBINATIONS=PASS")
}

func TestM11V2LogoutMixedCaseAndDuplicateCookie(t *testing.T) {
	t.Run("logout replay", func(t *testing.T) {
		store := &httpAuthTestStore{}
		app := newHTTPAuthApp(t, newHTTPAuthService(t, store))
		reg := authRequest(t, app, http.MethodPost, "/api/v1/auth/register", `{
			"email":"m11-v2-logout-replay@example.com",
			"password":"correct horse battery staple",
			"language":"en","theme":"system","timezone":"UTC"
		}`, "", "")
		cookie := requireCookie(t, reg, auth.RefreshCookieName)
		csrf := readCSRFToken(t, reg)
		reg.Body.Close()
		logout := authRequest(t, app, http.MethodPost, "/api/v1/auth/logout", `{"allSessions":false}`, cookie.Value, csrf)
		logout.Body.Close()
		if logout.StatusCode != http.StatusOK {
			t.Fatalf("logout status=%d", logout.StatusCode)
		}
		replay := authRequest(t, app, http.MethodPost, "/api/v1/auth/refresh", "", cookie.Value, csrf)
		replay.Body.Close()
		if replay.StatusCode != http.StatusUnauthorized {
			t.Fatalf("refresh after logout status=%d", replay.StatusCode)
		}
	})

	t.Run("mixed case csrf header", func(t *testing.T) {
		store := &httpAuthTestStore{}
		app := newHTTPAuthApp(t, newHTTPAuthService(t, store))
		reg := authRequest(t, app, http.MethodPost, "/api/v1/auth/register", `{
			"email":"m11-v2-mixed-case@example.com",
			"password":"correct horse battery staple",
			"language":"en","theme":"system","timezone":"UTC"
		}`, "", "")
		cookie := requireCookie(t, reg, auth.RefreshCookieName)
		csrf := readCSRFToken(t, reg)
		reg.Body.Close()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
		req.AddCookie(&http.Cookie{Name: auth.RefreshCookieName, Value: cookie.Value})
		req.Header.Set("x-cSrF-ToKeN", csrf)
		resp, err := app.Test(req)
		if err != nil { t.Fatal(err) }
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("mixed-case csrf status=%d", resp.StatusCode)
		}
	})

	t.Run("duplicate cookie same session", func(t *testing.T) {
		store := &httpAuthTestStore{}
		app := newHTTPAuthApp(t, newHTTPAuthService(t, store))
		reg := authRequest(t, app, http.MethodPost, "/api/v1/auth/register", `{
			"email":"m11-v2-duplicate-cookie@example.com",
			"password":"correct horse battery staple",
			"language":"en","theme":"system","timezone":"UTC"
		}`, "", "")
		cookie := requireCookie(t, reg, auth.RefreshCookieName)
		csrf := readCSRFToken(t, reg)
		reg.Body.Close()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
		req.Header.Add("Cookie", auth.RefreshCookieName+"="+cookie.Value)
		req.Header.Add("Cookie", auth.RefreshCookieName+"="+cookie.Value)
		req.Header.Set("X-CSRF-Token", csrf)
		resp, err := app.Test(req)
		if err != nil { t.Fatal(err) }
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("duplicate Cookie status=%d", resp.StatusCode)
		}
		t.Logf("M11_V2_DUPLICATE_COOKIE_STATUS=%d", resp.StatusCode)
	})
}
