package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/openinvest/openinvest/backend-go/internal/auth"
)

func requireModule1HTTPAdversarial(t *testing.T) {
	t.Helper()
	if os.Getenv("OPENINVEST_MODULE1_ADVERSARIAL") != "1" {
		t.Skip("OPENINVEST_MODULE1_ADVERSARIAL is not enabled")
	}
}

func TestModule1AuthorizationHeaderAmbiguityMatrix(t *testing.T) {
	requireModule1HTTPAdversarial(t)
	store := &httpAuthTestStore{}
	authService := newHTTPAuthService(t, store)
	regApp := newHTTPAuthApp(t, authService)

	register := authRequest(t, regApp, http.MethodPost, "/api/v1/auth/register", `{
		"email":"module1-authz@example.test",
		"password":"correct horse battery staple",
		"language":"en",
		"theme":"system",
		"timezone":"UTC"
	}`, "", "")
	defer register.Body.Close()
	var payload struct {
		Data struct {
			User struct {
				ID string `json:"id"`
			} `json:"user"`
			Session struct {
				AccessToken string `json:"accessToken"`
			} `json:"session"`
		} `json:"data"`
	}
	if err := json.NewDecoder(register.Body).Decode(&payload); err != nil {
		t.Fatalf("decode registration response: %v", err)
	}
	if payload.Data.Session.AccessToken == "" {
		t.Fatal("missing access token seed")
	}

	api := &API{auth: authService}
	app := fiber.New()
	app.Get("/protected", func(c fiber.Ctx) error {
		subject, err := api.subjectID(c)
		if err != nil {
			return c.SendStatus(http.StatusUnauthorized)
		}
		return c.Status(http.StatusOK).SendString(subject)
	})

	type tc struct {
		name    string
		values  []string
		wantOK  bool
	}
	valid := "Bearer " + payload.Data.Session.AccessToken
	cases := []tc{
		{name: "missing", values: nil, wantOK: false},
		{name: "empty", values: []string{""}, wantOK: false},
		{name: "bearer_empty", values: []string{"Bearer "}, wantOK: false},
		{name: "lowercase_scheme", values: []string{"bearer " + payload.Data.Session.AccessToken}, wantOK: false},
		{name: "basic_plus_bearer_combined", values: []string{"Basic synthetic, " + valid}, wantOK: false},
		{name: "single_valid", values: []string{valid}, wantOK: true},
		{name: "duplicate_equal", values: []string{valid, valid}, wantOK: true},
		{name: "duplicate_conflicting_valid_first", values: []string{valid, "Bearer synthetic.invalid"}, wantOK: false},
		{name: "duplicate_conflicting_valid_last", values: []string{"Bearer synthetic.invalid", valid}, wantOK: false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			for _, value := range test.values {
				req.Header.Add("Authorization", value)
			}
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("request: %v", err)
			}
			defer resp.Body.Close()
			gotOK := resp.StatusCode == http.StatusOK
			t.Logf("CASE=%s STATUS=%d RAW_AUTH_VALUES=%q", test.name, resp.StatusCode, test.values)
			if gotOK != test.wantOK {
				t.Fatalf("authorization ambiguity policy violated: wantOK=%t got status=%d values=%q", test.wantOK, resp.StatusCode, test.values)
			}
		})
	}
}

func TestModule1RefreshCookieCollisionMatrix(t *testing.T) {
	requireModule1HTTPAdversarial(t)
	for _, order := range []struct {
		name    string
		validFirst bool
	}{
		{name: "valid_first_invalid_last", validFirst: true},
		{name: "invalid_first_valid_last", validFirst: false},
	} {
		t.Run(order.name, func(t *testing.T) {
			store := &httpAuthTestStore{}
			app := newHTTPAuthApp(t, newHTTPAuthService(t, store))
			register := authRequest(t, app, http.MethodPost, "/api/v1/auth/register", `{
				"email":"cookie-`+order.name+`@example.test",
				"password":"correct horse battery staple",
				"language":"en",
				"theme":"system",
				"timezone":"UTC"
			}`, "", "")
			refresh := requireCookie(t, register, auth.RefreshCookieName)
			csrf := readCSRFToken(t, register)
			register.Body.Close()

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
			req.Header.Set("X-CSRF-Token", csrf)
			if order.validFirst {
				req.Header.Add("Cookie", auth.RefreshCookieName+"="+refresh.Value+"; "+auth.RefreshCookieName+"=synthetic-invalid")
			} else {
				req.Header.Add("Cookie", auth.RefreshCookieName+"=synthetic-invalid; "+auth.RefreshCookieName+"="+refresh.Value)
			}
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("refresh collision request: %v", err)
			}
			defer resp.Body.Close()
			t.Logf("CASE=%s STATUS=%d", order.name, resp.StatusCode)
			if resp.StatusCode == http.StatusOK {
				t.Fatalf("conflicting duplicate refresh cookie was accepted")
			}
		})
	}
}

func TestModule1AuthJSONAmbiguityMatrix(t *testing.T) {
	requireModule1HTTPAdversarial(t)
	cases := []struct {
		name string
		path string
		body string
	}{
		{
			name: "duplicate_email",
			path: "/api/v1/auth/register",
			body: `{"email":"first@example.test","email":"second@example.test","password":"correct horse battery staple","language":"en","theme":"system","timezone":"UTC"}`,
		},
		{
			name: "duplicate_password",
			path: "/api/v1/auth/register",
			body: `{"email":"dup-password@example.test","password":"wrong wrong wrong","password":"correct horse battery staple","language":"en","theme":"system","timezone":"UTC"}`,
		},
		{
			name: "trailing_json",
			path: "/api/v1/auth/register",
			body: `{"email":"trailing@example.test","password":"correct horse battery staple","language":"en","theme":"system","timezone":"UTC"} {}`,
		},
		{
			name: "wrong_top_level_type",
			path: "/api/v1/auth/register",
			body: `[]`,
		},
		{
			name: "unknown_field",
			path: "/api/v1/auth/register",
			body: `{"email":"unknown@example.test","password":"correct horse battery staple","language":"en","theme":"system","timezone":"UTC","admin":true}`,
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			store := &httpAuthTestStore{}
			app := newHTTPAuthApp(t, newHTTPAuthService(t, store))
			req := httptest.NewRequest(http.MethodPost, test.path, strings.NewReader(test.body))
			req.Header.Set("Content-Type", "application/json")
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("request: %v", err)
			}
			defer resp.Body.Close()
			t.Logf("CASE=%s STATUS=%d REGISTER_CALLS=%d", test.name, resp.StatusCode, store.registerCalls)
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("ambiguous/strict JSON case must fail closed with 400, got %d", resp.StatusCode)
			}
			if store.registerCalls != 0 {
				t.Fatalf("ambiguous/invalid JSON reached persistence: calls=%d", store.registerCalls)
			}
		})
	}
}

func TestModule1BearerParserNegativeControl(t *testing.T) {
	requireModule1HTTPAdversarial(t)
	valid := "Bearer synthetic-token"
	if got := bearerToken(valid); got != "synthetic-token" {
		t.Fatalf("negative-control prerequisite failed: %q", got)
	}
	if got := bearerToken("Basic synthetic-token"); got != "" {
		t.Fatalf("oracle failed to reject deliberately invalid scheme: %q", got)
	}
}


func TestModule1ProtectedRouteInventoryRejectsAnonymous(t *testing.T) {
	requireModule1HTTPAdversarial(t)
	app := newHTTPAuthApp(t, newHTTPAuthService(t, &httpAuthTestStore{}))
	const portfolioID = "11111111-1111-4111-8111-111111111111"
	cases := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/portfolios"},
		{http.MethodGet, "/api/v1/portfolios/" + portfolioID},
		{http.MethodGet, "/api/v1/portfolios/" + portfolioID + "/summary"},
		{http.MethodGet, "/api/v1/portfolios/" + portfolioID + "/positions"},
		{http.MethodGet, "/api/v1/portfolios/" + portfolioID + "/returns"},
		{http.MethodGet, "/api/v1/portfolios/" + portfolioID + "/cash-flow"},
		{http.MethodGet, "/api/v1/portfolios/" + portfolioID + "/transactions"},
	}
	for _, test := range cases {
		name := test.method + "_" + strings.ReplaceAll(test.path, "/", "_")
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(test.method, test.path, nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("request: %v", err)
			}
			defer resp.Body.Close()
			t.Logf("PROTECTED_ROUTE=%s %s STATUS=%d", test.method, test.path, resp.StatusCode)
			if resp.StatusCode != http.StatusUnauthorized {
				t.Fatalf("protected route accepted anonymous request or rejected before auth guard: status=%d", resp.StatusCode)
			}
		})
	}
}


func TestModule1SessionFixationPreAuthCookieNotReused(t *testing.T) {
	requireModule1HTTPAdversarial(t)
	store := &httpAuthTestStore{}
	app := newHTTPAuthApp(t, newHTTPAuthService(t, store))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(`{
		"email":"fixation@example.test",
		"password":"correct horse battery staple",
		"language":"en",
		"theme":"system",
		"timezone":"UTC"
	}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Cookie", auth.RefreshCookieName+"=attacker-controlled-preauth-cookie")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("registration with preauth cookie: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("registration status=%d", resp.StatusCode)
	}
	cookie := requireCookie(t, resp, auth.RefreshCookieName)
	if cookie.Value == "" || cookie.Value == "attacker-controlled-preauth-cookie" {
		t.Fatalf("attacker-controlled preauth cookie was reused as authenticated session")
	}
	if _, exists := store.sessions[tokenHashForModule1("attacker-controlled-preauth-cookie")]; exists {
		t.Fatalf("attacker-controlled preauth refresh value reached durable session map")
	}
	t.Log("SESSION_FIXATION_PREAUTH_COOKIE_REUSED=NO")
}

func TestModule1CrossPrincipalCSRFCannotRefreshOtherSession(t *testing.T) {
	requireModule1HTTPAdversarial(t)

	storeA := &httpAuthTestStore{}
	appA := newHTTPAuthApp(t, newHTTPAuthService(t, storeA))
	registerA := authRequest(t, appA, http.MethodPost, "/api/v1/auth/register", `{
		"email":"csrf-a@example.test",
		"password":"correct horse battery staple",
		"language":"en",
		"theme":"system",
		"timezone":"UTC"
	}`, "", "")
	cookieA := requireCookie(t, registerA, auth.RefreshCookieName)
	csrfA := readCSRFToken(t, registerA)
	registerA.Body.Close()

	storeB := &httpAuthTestStore{}
	appB := newHTTPAuthApp(t, newHTTPAuthService(t, storeB))
	registerB := authRequest(t, appB, http.MethodPost, "/api/v1/auth/register", `{
		"email":"csrf-b@example.test",
		"password":"correct horse battery staple",
		"language":"en",
		"theme":"system",
		"timezone":"UTC"
	}`, "", "")
	csrfB := readCSRFToken(t, registerB)
	registerB.Body.Close()

	before := len(storeA.sessions)
	resp := authRequest(t, appA, http.MethodPost, "/api/v1/auth/refresh", "", cookieA.Value, csrfB)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("cross-principal csrf mismatch status=%d, want 401", resp.StatusCode)
	}
	if len(storeA.sessions) != before {
		t.Fatalf("cross-principal csrf mismatch mutated A session state before=%d after=%d", before, len(storeA.sessions))
	}

	valid := authRequest(t, appA, http.MethodPost, "/api/v1/auth/refresh", "", cookieA.Value, csrfA)
	defer valid.Body.Close()
	if valid.StatusCode != http.StatusOK {
		t.Fatalf("valid A refresh failed after rejected B csrf: %d", valid.StatusCode)
	}
	t.Log("CSRF_SESSION_BINDING_BYPASS=NO")
}

func tokenHashForModule1(value string) string {
	// Test-side independent persistence check: the store keys are SHA-256 hex digests
	// of raw refresh tokens. Reuse the public behavior via a tiny local computation.
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
