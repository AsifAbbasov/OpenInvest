package httpapi

import (
	"net/http"
	"testing"
)

func TestM11V2CSRFCrossSessionAndRotationMatrix(t *testing.T) {
	t.Run("cross session csrf", func(t *testing.T) {
		storeA := &httpAuthTestStore{}
		appA := newHTTPAuthApp(t, newHTTPAuthService(t, storeA))
		regA := authRequest(t, appA, http.MethodPost, "/api/v1/auth/register", `{
			"email":"m11-v2-cross-a@example.com",
			"password":"correct horse battery staple",
			"language":"en","theme":"system","timezone":"UTC"
		}`, "", "")
		cookieA := requireCookie(t, regA, "oi_refresh")
		csrfA := readCSRFToken(t, regA)
		regA.Body.Close()

		storeB := &httpAuthTestStore{}
		appB := newHTTPAuthApp(t, newHTTPAuthService(t, storeB))
		regB := authRequest(t, appB, http.MethodPost, "/api/v1/auth/register", `{
			"email":"m11-v2-cross-b@example.com",
			"password":"correct horse battery staple",
			"language":"en","theme":"system","timezone":"UTC"
		}`, "", "")
		csrfB := readCSRFToken(t, regB)
		regB.Body.Close()

		response := authRequest(t, appA, http.MethodPost, "/api/v1/auth/refresh", "", cookieA.Value, csrfB)
		response.Body.Close()
		if response.StatusCode != http.StatusUnauthorized {
			t.Fatalf("cross-session csrf status=%d", response.StatusCode)
		}
		if csrfA == csrfB {
			t.Fatal("independent sessions unexpectedly share csrf")
		}
	})

	t.Run("old pair replay after rotation", func(t *testing.T) {
		store := &httpAuthTestStore{}
		app := newHTTPAuthApp(t, newHTTPAuthService(t, store))
		reg := authRequest(t, app, http.MethodPost, "/api/v1/auth/register", `{
			"email":"m11-v2-rotation@example.com",
			"password":"correct horse battery staple",
			"language":"en","theme":"system","timezone":"UTC"
		}`, "", "")
		oldCookie := requireCookie(t, reg, "oi_refresh")
		oldCSRF := readCSRFToken(t, reg)
		reg.Body.Close()

		rotated := authRequest(t, app, http.MethodPost, "/api/v1/auth/refresh", "", oldCookie.Value, oldCSRF)
		rotated.Body.Close()
		if rotated.StatusCode != http.StatusOK {
			t.Fatalf("rotation status=%d", rotated.StatusCode)
		}
		replay := authRequest(t, app, http.MethodPost, "/api/v1/auth/refresh", "", oldCookie.Value, oldCSRF)
		replay.Body.Close()
		if replay.StatusCode != http.StatusUnauthorized {
			t.Fatalf("old refresh replay status=%d", replay.StatusCode)
		}
	})
	t.Log("M11_V2_CSRF_CROSS_SESSION_ROTATION=PASS")
}
