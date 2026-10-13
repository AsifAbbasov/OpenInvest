package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/openinvest/openinvest/backend-go/internal/auth"
)

func TestModule1AuthLogAndErrorResponseSecretSentinels(t *testing.T) {
	requireModule1HTTPAdversarial(t)

	const passwordSentinel = "M1_PASSWORD_SENTINEL_7f93b05b"
	const invalidBearerSentinel = "M1_BEARER_SENTINEL_d02f2ca1"
	const invalidCSRFSentinel = "M1_CSRF_SENTINEL_91afda8e"

	var captured bytes.Buffer
	originalWriter := log.Writer()
	originalFlags := log.Flags()
	originalPrefix := log.Prefix()
	log.SetOutput(&captured)
	log.SetFlags(0)
	log.SetPrefix("")
	defer func() {
		log.SetOutput(originalWriter)
		log.SetFlags(originalFlags)
		log.SetPrefix(originalPrefix)
	}()

	store := &httpAuthTestStore{}
	service := newHTTPAuthService(t, store)
	app := newHTTPAuthApp(t, service)

	registerBody := `{
		"email":"module1-log-sentinel@example.test",
		"password":"` + passwordSentinel + `",
		"language":"en",
		"theme":"system",
		"timezone":"UTC"
	}`
	register := authRequest(t, app, http.MethodPost, "/api/v1/auth/register", registerBody, "", "")
	if register.StatusCode != http.StatusCreated {
		t.Fatalf("registration status=%d", register.StatusCode)
	}
	refreshCookie := requireCookie(t, register, auth.RefreshCookieName)
	var payload struct {
		Data struct {
			Session struct {
				AccessToken string `json:"accessToken"`
				CSRFToken   string `json:"csrfToken"`
			} `json:"session"`
		} `json:"data"`
	}
	if err := json.NewDecoder(register.Body).Decode(&payload); err != nil {
		register.Body.Close()
		t.Fatalf("decode registration response: %v", err)
	}
	register.Body.Close()
	csrf := payload.Data.Session.CSRFToken
	if payload.Data.Session.AccessToken == "" || refreshCookie.Value == "" || csrf == "" {
		t.Fatal("missing generated auth secret sentinel prerequisites")
	}

	// Login failure exercises password handling without making the password part of an error.
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(
		`{"email":"module1-log-sentinel@example.test","password":"wrong-`+passwordSentinel+`"}`,
	))
	loginReq.Header.Set("Content-Type", "application/json")
	loginResp, err := app.Test(loginReq)
	if err != nil {
		t.Fatalf("login failure request: %v", err)
	}

	// Refresh failure exercises a real refresh credential paired with a synthetic invalid CSRF marker.
	refreshReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	refreshReq.Header.Set("Cookie", auth.RefreshCookieName+"="+refreshCookie.Value)
	refreshReq.Header.Set("X-CSRF-Token", invalidCSRFSentinel)
	refreshResp, err := app.Test(refreshReq)
	if err != nil {
		t.Fatalf("refresh failure request: %v", err)
	}

	// Protected endpoint with invalid bearer exercises Authorization parsing/error handling.
	protected := newHTTPAuthApp(t, service)
	bearerReq := httptest.NewRequest(http.MethodGet, "/api/v1/portfolios", nil)
	bearerReq.Header.Set("Authorization", "Bearer "+invalidBearerSentinel)
	bearerResp, err := protected.Test(bearerReq)
	if err != nil {
		t.Fatalf("invalid bearer request: %v", err)
	}

	// Logout failure uses the real refresh credential with the invalid CSRF marker.
	logoutReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", strings.NewReader(`{"allSessions":false}`))
	logoutReq.Header.Set("Content-Type", "application/json")
	logoutReq.Header.Set("Cookie", auth.RefreshCookieName+"="+refreshCookie.Value)
	logoutReq.Header.Set("X-CSRF-Token", invalidCSRFSentinel)
	logoutResp, err := app.Test(logoutReq)
	if err != nil {
		t.Fatalf("logout failure request: %v", err)
	}

	readResponse := func(name string, resp *http.Response) string {
		t.Helper()
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("%s read response: %v", name, err)
		}
		var header strings.Builder
		for key, values := range resp.Header {
			if strings.EqualFold(key, "Set-Cookie") {
				// Set-Cookie on successful issuance intentionally carries the refresh credential.
				// Failure responses used here should not contain it, but do not treat the
				// success-path cookie contract as a logging leak.
				continue
			}
			header.WriteString(key)
			header.WriteString(":")
			header.WriteString(strings.Join(values, ","))
			header.WriteString("\n")
		}
		return string(body) + "\n" + header.String()
	}

	errorSurface := strings.Join([]string{
		readResponse("login", loginResp),
		readResponse("refresh", refreshResp),
		readResponse("bearer", bearerResp),
		readResponse("logout", logoutResp),
	}, "\n")

	secrets := map[string]string{
		"password": passwordSentinel,
		"access_token": payload.Data.Session.AccessToken,
		"refresh_token": refreshCookie.Value,
		"csrf_token": csrf,
		"invalid_bearer_marker": invalidBearerSentinel,
		"invalid_csrf_marker": invalidCSRFSentinel,
	}
	logText := captured.String()
	for name, secret := range secrets {
		if secret == "" {
			t.Fatalf("%s sentinel is empty", name)
		}
		if strings.Contains(logText, secret) {
			t.Fatalf("raw %s leaked to repository-owned stdlib log output", name)
		}
		if strings.Contains(errorSurface, secret) {
			t.Fatalf("raw %s leaked to auth error response surface", name)
		}
	}
	t.Log("AUTH_LOG_RAW_SECRET_LEAK=NO_FOR_TESTED_REPOSITORY_OWNED_PATHS")
	t.Log("AUTH_ERROR_RESPONSE_RAW_SECRET_LEAK=NO_FOR_TESTED_PATHS")
	t.Log("EXTERNAL_LOG_AGGREGATOR=NOT_VERIFIED")
}
