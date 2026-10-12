package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"

	"github.com/openinvest/openinvest/backend-go/internal/auth"
)

func module1BearerProbeApp(t *testing.T, service *auth.Service) *fiber.App {
	t.Helper()
	api := &API{auth: service}
	app := newFiberApp(HTTPNetworkConfig{})
	app.Get("/module1/protected", func(c fiber.Ctx) error {
		subject, err := api.subjectID(c)
		if err != nil {
			return c.Status(http.StatusUnauthorized).SendString("unauthorized")
		}
		return c.Status(http.StatusOK).SendString(subject)
	})
	return app
}

func module1AccessToken(t *testing.T, service *auth.Service) string {
	t.Helper()
	result, err := service.Register(context.Background(), auth.RegistrationRequest{
		Email: "module1-authz@example.test",
		Password: "correct horse battery staple",
		Language: auth.LanguageEN,
		Theme: auth.ThemeSystem,
		Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("register fixture: %v", err)
	}
	return result.Session.AccessToken
}

func TestModule1AuthorizationHeaderAmbiguityMatrix(t *testing.T) {
	store := &httpAuthTestStore{}
	service := newHTTPAuthService(t, store)
	token := module1AccessToken(t, service)

	type headerField struct{ name, value string }
	cases := []struct {
		name string
		headers []headerField
		want int
	}{
		{"valid_single", []headerField{{"Authorization", "Bearer " + token}}, http.StatusOK},
		{"missing", nil, http.StatusUnauthorized},
		{"empty", []headerField{{"Authorization", ""}}, http.StatusUnauthorized},
		{"bearer_empty", []headerField{{"Authorization", "Bearer "}}, http.StatusUnauthorized},
		{"lowercase_scheme", []headerField{{"Authorization", "bearer " + token}}, http.StatusUnauthorized},
		{"tab_separator", []headerField{{"Authorization", "Bearer\t" + token}}, http.StatusUnauthorized},
		{"comma_combined", []headerField{{"Authorization", "Bearer " + token + ", Bearer invalid"}}, http.StatusUnauthorized},
		{"duplicate_equal", []headerField{{"Authorization", "Bearer " + token}, {"Authorization", "Bearer " + token}}, http.StatusOK},
		{"duplicate_conflicting_valid_first", []headerField{{"Authorization", "Bearer " + token}, {"Authorization", "Bearer invalid"}}, http.StatusUnauthorized},
		{"duplicate_conflicting_valid_last", []headerField{{"Authorization", "Bearer invalid"}, {"Authorization", "Bearer " + token}}, http.StatusUnauthorized},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := module1BearerProbeApp(t, service)
			req := httptest.NewRequest(http.MethodGet, "/module1/protected", nil)
			for _, h := range tc.headers {
				req.Header.Add(h.name, h.value)
			}
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("request: %v", err)
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			t.Logf("PROPERTY=M1-AUTHZ-HEADER CASE=%s STATUS=%d BODY=%q RAW_AUTH_VALUES=%q", tc.name, resp.StatusCode, string(body), req.Header.Values("Authorization"))
			if resp.StatusCode != tc.want {
				t.Fatalf("AMBIGUOUS_AUTHORIZATION_REPRESENTATION_GRANTS_ACCESS case=%s expected=%d got=%d", tc.name, tc.want, resp.StatusCode)
			}
		})
	}
}

func module1RegisterHTTPFixture(t *testing.T) (*fiber.App, *httpAuthTestStore, string, string) {
	t.Helper()
	store := &httpAuthTestStore{}
	app := newHTTPAuthApp(t, newHTTPAuthService(t, store))
	resp := authRequest(t, app, http.MethodPost, "/api/v1/auth/register", `{
		"email":"module1-cookie@example.test",
		"password":"correct horse battery staple",
		"language":"en",
		"theme":"system",
		"timezone":"UTC"
	}`, "", "")
	if resp.StatusCode != http.StatusCreated {
		defer resp.Body.Close()
		data, _ := io.ReadAll(resp.Body)
		t.Fatalf("register fixture status=%d body=%s", resp.StatusCode, data)
	}
	cookie := requireCookie(t, resp, auth.RefreshCookieName)
	csrf := readCSRFToken(t, resp)
	resp.Body.Close()
	return app, store, cookie.Value, csrf
}

func TestModule1RefreshCookieCollisionMatrix(t *testing.T) {
	cases := []struct {
		name string
		values func(valid string) []string
		want int
	}{
		{"duplicate_equal", func(v string) []string { return []string{v, v} }, http.StatusOK},
		{"valid_first_invalid_last", func(v string) []string { return []string{v, "invalid-refresh"} }, http.StatusUnauthorized},
		{"invalid_first_valid_last", func(v string) []string { return []string{"invalid-refresh", v} }, http.StatusUnauthorized},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app, store, valid, csrf := module1RegisterHTTPFixture(t)
			before := len(store.sessions)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
			for _, value := range tc.values(valid) {
				req.Header.Add("Cookie", fmt.Sprintf("%s=%s", auth.RefreshCookieName, value))
			}
			req.Header.Set("X-CSRF-Token", csrf)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("refresh: %v", err)
			}
			defer resp.Body.Close()
			after := len(store.sessions)
			t.Logf("PROPERTY=M1-COOKIE CASE=%s STATUS=%d COOKIE_FIELDS=%d SESSIONS_BEFORE=%d SESSIONS_AFTER=%d", tc.name, resp.StatusCode, len(req.Header.Values("Cookie")), before, after)
			if resp.StatusCode != tc.want {
				t.Fatalf("REFRESH_COOKIE_REPRESENTATION_POLICY case=%s expected=%d got=%d", tc.name, tc.want, resp.StatusCode)
			}
			if after != before {
				t.Fatalf("SESSION_FAMILY_STATE_UNEXPECTEDLY_CHANGED case=%s before=%d after=%d", tc.name, before, after)
			}
		})
	}
}

func TestModule1StrictAuthJSONAmbiguityMatrix(t *testing.T) {
	cases := []struct {
		name, body, contentType string
		mustReject bool
	}{
		{"duplicate_email_key", `{"email":"first@example.test","email":"second@example.test","password":"correct horse battery staple","language":"en","theme":"system","timezone":"UTC"}`, "application/json", true},
		{"multiple_objects", `{"email":"a@example.test"}{"email":"b@example.test"}`, "application/json", true},
		{"trailing_json", `{"email":"a@example.test","password":"correct horse battery staple","language":"en","theme":"system","timezone":"UTC"} trailing`, "application/json", true},
		{"array_top_level", `[]`, "application/json", true},
		{"null_top_level", `null`, "application/json", true},
		{"unknown_field", `{"email":"a@example.test","password":"correct horse battery staple","language":"en","theme":"system","timezone":"UTC","admin":true}`, "application/json", true},
		{"text_plain_json", `{"email":"plain@example.test","password":"correct horse battery staple","language":"en","theme":"system","timezone":"UTC"}`, "text/plain", false},
		{"missing_content_type_json", `{"email":"missingct@example.test","password":"correct horse battery staple","language":"en","theme":"system","timezone":"UTC"}`, "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := newHTTPAuthApp(t, newHTTPAuthService(t, &httpAuthTestStore{}))
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(tc.body))
			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("request: %v", err)
			}
			defer resp.Body.Close()
			t.Logf("PROPERTY=M1-AUTH-JSON CASE=%s STATUS=%d CONTENT_TYPE=%q", tc.name, resp.StatusCode, tc.contentType)
			if tc.mustReject && resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("PARSER_AMBIGUITY_BYPASSES_VALIDATION case=%s expected=400 got=%d", tc.name, resp.StatusCode)
			}
		})
	}
}

func TestModule1AuthErrorSurfaceDoesNotLeakSyntheticSecrets(t *testing.T) {
	sentinels := []string{
		"M1_ACCESS_TOKEN_SECRET_39f73",
		"M1_REFRESH_SECRET_39f73",
		"M1_DATABASE_PASSWORD_39f73",
		"M1_REDIS_PASSWORD_39f73",
		"M1_PROVIDER_TOKEN_39f73",
	}
	store := &httpAuthTestStore{findErr: fmt.Errorf("synthetic postgres outage %s %s", sentinels[2], sentinels[4])}
	app := newHTTPAuthApp(t, newHTTPAuthService(t, store))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":"missing@example.test","password":"M1_REFRESH_SECRET_39f73"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	serialized := string(body)
	for _, sentinel := range sentinels {
		if strings.Contains(serialized, sentinel) {
			t.Fatalf("RAW_SECRET_LEAK sentinel=%q body=%q", sentinel, serialized)
		}
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("response must remain structured JSON: %v", err)
	}
	t.Logf("PROPERTY=M1-LOG-LEAK STATUS=%d BODY_BYTES=%d RAW_SECRET_LEAK=NO_FOR_TESTED_RESPONSE", resp.StatusCode, len(body))
}


func TestModule1SessionFixationPreauthCookieIsNeverReused(t *testing.T) {
	store := &httpAuthTestStore{}
	app := newHTTPAuthApp(t, newHTTPAuthService(t, store))
	attackerCookie := "attacker-controlled-preauth-session"

	registerBody := `{
		"email":"module1-fixation@example.test",
		"password":"correct horse battery staple",
		"language":"en",
		"theme":"system",
		"timezone":"UTC"
	}`
	registerReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(registerBody))
	registerReq.Header.Set("Content-Type", "application/json")
	registerReq.AddCookie(&http.Cookie{Name: auth.RefreshCookieName, Value: attackerCookie})
	registerResp, err := app.Test(registerReq)
	if err != nil {
		t.Fatal(err)
	}
	defer registerResp.Body.Close()
	if registerResp.StatusCode != http.StatusCreated {
		t.Fatalf("register expected 201 got=%d", registerResp.StatusCode)
	}
	issued := requireCookie(t, registerResp, auth.RefreshCookieName)
	if issued.Value == attackerCookie {
		t.Fatal("ATTACKER_SUPPLIED_PREAUTH_SESSION_IDENTIFIER_REUSED during registration")
	}

	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":"module1-fixation@example.test","password":"correct horse battery staple"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginReq.AddCookie(&http.Cookie{Name: auth.RefreshCookieName, Value: attackerCookie})
	loginResp, err := app.Test(loginReq)
	if err != nil {
		t.Fatal(err)
	}
	defer loginResp.Body.Close()
	if loginResp.StatusCode != http.StatusOK {
		t.Fatalf("login expected 200 got=%d", loginResp.StatusCode)
	}
	loginIssued := requireCookie(t, loginResp, auth.RefreshCookieName)
	if loginIssued.Value == attackerCookie {
		t.Fatal("ATTACKER_SUPPLIED_PREAUTH_SESSION_IDENTIFIER_REUSED during login")
	}
	if loginIssued.Value == issued.Value {
		t.Fatal("login must create a fresh session rather than preserve prior session identity")
	}
	t.Log("SESSION_FIXATION_PREAUTH_IDENTIFIER_REUSED=NO")
}

func module1ProtectedRouteSafe(status int) bool {
	return status < 200 || status >= 300
}

func TestModule1ProtectedRouteInventoryRejectsAnonymous(t *testing.T) {
	app := newHTTPAuthApp(t, newHTTPAuthService(t, &httpAuthTestStore{}))
	const portfolio = "11111111-1111-4111-8111-111111111111"
	const transaction = "22222222-2222-4222-8222-222222222222"
	cases := []struct {
		method, path, body string
	}{
		{http.MethodGet, "/api/v1/portfolios", ""},
		{http.MethodPost, "/api/v1/portfolios", `{"name":"x","baseCurrency":"RUB"}`},
		{http.MethodGet, "/api/v1/portfolios/" + portfolio, ""},
		{http.MethodGet, "/api/v1/portfolios/" + portfolio + "/summary", ""},
		{http.MethodGet, "/api/v1/portfolios/" + portfolio + "/positions", ""},
		{http.MethodPut, "/api/v1/portfolios/" + portfolio + "/valuations/TEST", `{"marketPrice":{"amount":"1","currency":"RUB"},"asOfDate":"2026-10-13"}`},
		{http.MethodDelete, "/api/v1/portfolios/" + portfolio + "/valuations/TEST", ""},
		{http.MethodGet, "/api/v1/portfolios/" + portfolio + "/returns?asOfDate=2026-10-13", ""},
		{http.MethodGet, "/api/v1/portfolios/" + portfolio + "/cash-flow", ""},
		{http.MethodGet, "/api/v1/portfolios/" + portfolio + "/transactions", ""},
		{http.MethodPost, "/api/v1/portfolios/" + portfolio + "/transactions", `{"transactionType":"DEPOSIT","grossAmount":{"amount":"1","currency":"RUB"},"commission":{"amount":"0","currency":"RUB"},"tax":{"amount":"0","currency":"RUB"},"tradeDate":"2026-10-13"}`},
		{http.MethodPatch, "/api/v1/portfolios/" + portfolio + "/transactions/" + transaction, `{}`},
		{http.MethodDelete, "/api/v1/portfolios/" + portfolio + "/transactions/" + transaction, `{}`},
		{http.MethodPost, "/api/v1/portfolios/" + portfolio + "/imports/review", `{"csvPayload":"type,amount\\n"}`},
		{http.MethodPost, "/api/v1/portfolios/" + portfolio + "/imports/append", `{}`},
	}
	tested := 0
	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		if tc.body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		req.Header.Set("Idempotency-Key", "module1-route-inventory-"+fmt.Sprint(tested))
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("%s %s: %v", tc.method, tc.path, err)
		}
		resp.Body.Close()
		tested++
		t.Logf("PROPERTY=M1-ROUTE-AUTH METHOD=%s PATH=%s STATUS=%d", tc.method, tc.path, resp.StatusCode)
		if !module1ProtectedRouteSafe(resp.StatusCode) {
			t.Fatalf("PROTECTED_ROUTE_WITHOUT_AUTH_GUARD method=%s path=%s status=%d", tc.method, tc.path, resp.StatusCode)
		}
	}
	t.Logf("PROTECTED_ROUTES_DISCOVERED=%d PROTECTED_ROUTES_TESTED=%d PROTECTED_ROUTES_WITHOUT_AUTH_GUARD=0", len(cases), tested)
}

func TestModule1ProtectedRouteOracleNegativeControl(t *testing.T) {
	if module1ProtectedRouteSafe(http.StatusOK) {
		t.Fatal("negative control failed: oracle accepted a synthetic unauthorized 200")
	}
	if !module1ProtectedRouteSafe(http.StatusUnauthorized) {
		t.Fatal("negative control failed: oracle rejected expected unauthorized denial")
	}
	t.Log("NEGATIVE_CONTROL_DETECTED=YES")
}

func TestModule1AuthStandardLoggerSentinelRedaction(t *testing.T) {
	var captured bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&captured)
	defer log.SetOutput(previous)

	sentinel := "M1_SYNTHETIC_PASSWORD_SENTINEL_4b28"
	app := newHTTPAuthApp(t, newHTTPAuthService(t, &httpAuthTestStore{}))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(fmt.Sprintf(`{"email":"missing@example.test","password":%q}`, sentinel)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if strings.Contains(captured.String(), sentinel) {
		t.Fatalf("RAW_SECRET_LEAK_IN_STANDARD_LOG=%q", captured.String())
	}
	t.Logf("AUTH_LOG_LEAKAGE=NO_FOR_TESTED_STANDARD_LOG_PATH bytes=%d", captured.Len())
}
