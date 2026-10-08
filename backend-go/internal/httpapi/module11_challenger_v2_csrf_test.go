package httpapi

import (
	"io"
	"net/http"
	"testing"
)

func TestM11V2CSRFSessionMatrix(t *testing.T) {
	cases := []struct {
		name string
		csrf string
	}{
		{name: "missing", csrf: ""},
		{name: "malformed", csrf: "malformed"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := &httpAuthTestStore{}
			app := newHTTPAuthApp(t, newHTTPAuthService(t, store))
			register := authRequest(t, app, http.MethodPost, "/api/v1/auth/register", `{
				"email":"m11-v2-`+tc.name+`@example.com",
				"password":"correct horse battery staple",
				"language":"en",
				"theme":"system",
				"timezone":"UTC"
			}`, "", "")
			cookie := requireCookie(t, register, "oi_refresh")
			csrf := readCSRFToken(t, register)
			register.Body.Close()
			if tc.csrf != "" {
				csrf = tc.csrf
			} else {
				csrf = ""
			}
			before := len(store.sessions)
			response := authRequest(t, app, http.MethodPost, "/api/v1/auth/refresh", "", cookie.Value, csrf)
			defer response.Body.Close()
			body, _ := io.ReadAll(response.Body)
			if response.StatusCode != http.StatusUnauthorized {
				t.Fatalf("status=%d body=%s", response.StatusCode, string(body))
			}
			if len(store.sessions) != before {
				t.Fatalf("rejected CSRF mutated session state")
			}
		})
	}
	t.Log("M11_V2_CSRF_MISSING_MALFORMED=PASS")
}
