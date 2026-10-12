package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
		{"duplicate_equal", []headerField{{"Authorization", "Bearer " + token}, {"Authorization", "Bearer " + token}}, http.StatusUnauthorized},
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
	}{
		{"duplicate_equal", func(v string) []string { return []string{v, v} }},
		{"valid_first_invalid_last", func(v string) []string { return []string{v, "invalid-refresh"} }},
		{"invalid_first_valid_last", func(v string) []string { return []string{"invalid-refresh", v} }},
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
			if resp.StatusCode != http.StatusUnauthorized {
				t.Fatalf("AMBIGUOUS_REFRESH_COOKIE_ACCEPTED case=%s status=%d", tc.name, resp.StatusCode)
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
