package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
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

	register := fixtureLegacyLogin(t, regApp, http.MethodPost, "/api/v1/auth/register", `{
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
		name       string
		values     []string
		wantStatus int
	}
	valid := "Bearer " + payload.Data.Session.AccessToken
	cases := []tc{
		{name: "missing", values: nil, wantStatus: http.StatusUnauthorized},
		{name: "empty", values: []string{""}, wantStatus: http.StatusUnauthorized},
		{name: "bearer_empty", values: []string{"Bearer "}, wantStatus: http.StatusUnauthorized},
		{name: "lowercase_scheme", values: []string{"bearer " + payload.Data.Session.AccessToken}, wantStatus: http.StatusUnauthorized},
		{name: "basic_plus_bearer_combined", values: []string{"Basic synthetic, " + valid}, wantStatus: http.StatusUnauthorized},
		{name: "single_valid", values: []string{valid}, wantStatus: http.StatusOK},
		{name: "duplicate_equal", values: []string{valid, valid}, wantStatus: http.StatusOK},
		// Fiber currently exposes the first physical Authorization value through c.Get.
		// The security oracle is therefore principal integrity, not an undocumented
		// policy that every duplicate field must be rejected.
		{name: "duplicate_conflicting_valid_first", values: []string{valid, "Bearer synthetic.invalid"}, wantStatus: http.StatusOK},
		{name: "duplicate_conflicting_valid_last", values: []string{"Bearer synthetic.invalid", valid}, wantStatus: http.StatusUnauthorized},
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
			body, _ := io.ReadAll(resp.Body)
			t.Logf("CASE=%s STATUS=%d RAW_AUTH_VALUES=%q RESOLVED_SUBJECT=%q", test.name, resp.StatusCode, test.values, strings.TrimSpace(string(body)))
			if resp.StatusCode != test.wantStatus {
				t.Fatalf("authorization parser behavior changed: want status=%d got=%d values=%q", test.wantStatus, resp.StatusCode, test.values)
			}
			if resp.StatusCode == http.StatusOK && strings.TrimSpace(string(body)) != store.user.InvestmentSubjectID {
				t.Fatalf("authorization ambiguity resolved to unexpected principal: got=%q want=%q", strings.TrimSpace(string(body)), store.user.InvestmentSubjectID)
			}
		})
	}
	t.Log("AUTHORIZATION_DUPLICATE_POLICY=FIRST_PHYSICAL_VALUE; UNAUTHORIZED_PRINCIPAL_OBSERVED=NO")
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
			register := fixtureLegacyLogin(t, app, http.MethodPost, "/api/v1/auth/register", `{
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
			after := len(store.sessions)
			t.Logf("CASE=%s STATUS=%d SESSION_COUNT_BEFORE=1 SESSION_COUNT_AFTER=%d", order.name, resp.StatusCode, after)
			if order.validFirst {
				if resp.StatusCode != http.StatusOK {
					t.Fatalf("first valid refresh cookie was not selected: status=%d", resp.StatusCode)
				}
				if after != 1 {
					t.Fatalf("successful rotation must replace exactly one session, got session count=%d", after)
				}
				if _, exists := store.sessions[tokenHashForModule1(refresh.Value)]; exists {
					t.Fatalf("old refresh credential remained active after rotation")
				}
			} else {
				if resp.StatusCode != http.StatusUnauthorized {
					t.Fatalf("first invalid refresh cookie must not fall through to later valid cookie: status=%d", resp.StatusCode)
				}
				if after != 1 {
					t.Fatalf("rejected duplicate-cookie request mutated session state: count=%d", after)
				}
				if _, exists := store.sessions[tokenHashForModule1(refresh.Value)]; !exists {
					t.Fatalf("rejected duplicate-cookie request revoked or rotated the valid session")
				}
			}
		})
	}
}

func newModule1OwnershipService(t *testing.T, store *httpOwnershipFixture) *auth.Service {
	t.Helper()
	service, err := auth.NewService(store, fixedHTTPClock{}, auth.Config{
		AccessTokenSecret: []byte("01234567890123456789012345678901"),
		RefreshCookieSecure: true,
	})
	if err != nil { t.Fatal(err) }
	return service
}

func TestModule1AuthJSONAmbiguityMatrix(t *testing.T) {
	requireModule1HTTPAdversarial(t)

	t.Run("duplicate_email_last_member_wins_without_activation", func(t *testing.T) {
		store := &httpOwnershipFixture{}
		app := newHTTPAuthApp(t, newModule1OwnershipService(t, store))
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(`{"email":"first@example.test","email":"second@example.test"}`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil { t.Fatal(err) }
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusAccepted || store.email != "second@example.test" {
			t.Fatalf("duplicate email resolution changed status=%d email=%q", resp.StatusCode, store.email)
		}
		if store.registerCalls != 0 || len(store.sessions) != 0 {
			t.Fatal("duplicate registration representation activated identity/session before proof")
		}
	})

	for _, tc := range []struct{name, body string}{
		{"trailing_json", `{"email":"trailing@example.test"} {}`},
		{"wrong_top_level_type", `[]`},
		{"unknown_field", `{"email":"unknown@example.test","admin":true}`},
		{"legacy_password_field_rejected", `{"email":"legacy@example.test","password":"synthetic"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := &httpOwnershipFixture{}
			app := newHTTPAuthApp(t, newModule1OwnershipService(t, store))
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			resp, err := app.Test(req)
			if err != nil { t.Fatal(err) }
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusBadRequest || store.registerCalls != 0 || len(store.sessions) != 0 {
				t.Fatalf("strict registration JSON failed closed status=%d registerCalls=%d sessions=%d", resp.StatusCode, store.registerCalls, len(store.sessions))
			}
		})
	}
	t.Log("REGISTRATION_JSON_AMBIGUITY=NO_PREPROOF_ACTIVATION; STRICT_UNKNOWN_AND_TRAILING_REJECTION=PASS")
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
	store := &httpOwnershipFixture{}
	app := newHTTPAuthApp(t, newModule1OwnershipService(t, store))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(`{"email":"fixation@example.test"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Cookie", auth.RefreshCookieName+"=attacker-controlled-preauth-cookie")
	resp, err := app.Test(req)
	if err != nil { t.Fatal(err) }
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted || len(resp.Cookies()) != 0 {
		t.Fatalf("pre-proof registration status=%d cookies=%d", resp.StatusCode, len(resp.Cookies()))
	}
	if len(store.sessions) != 0 || store.registerCalls != 0 {
		t.Fatalf("pre-proof attacker cookie promoted to session/account sessions=%d registerCalls=%d", len(store.sessions), store.registerCalls)
	}
	t.Log("SESSION_FIXATION_PREAUTH_COOKIE_REUSED=NO SESSION_BEFORE_VERIFICATION=NO")
}

func TestModule1CrossPrincipalCSRFCannotRefreshOtherSession(t *testing.T) {
	requireModule1HTTPAdversarial(t)

	storeA := &httpAuthTestStore{}
	appA := newHTTPAuthApp(t, newHTTPAuthService(t, storeA))
	registerA := fixtureLegacyLogin(t, appA, http.MethodPost, "/api/v1/auth/register", `{
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
	registerB := fixtureLegacyLogin(t, appB, http.MethodPost, "/api/v1/auth/register", `{
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


func TestModule1AuthorizationRepresentationExtendedMatrix(t *testing.T) {
	requireModule1HTTPAdversarial(t)
	store := &httpAuthTestStore{}
	service := newHTTPAuthService(t, store)
	appReg := newHTTPAuthApp(t, service)
	register := fixtureLegacyLogin(t, appReg, http.MethodPost, "/api/v1/auth/register", `{
		"email":"module1-authz-extended@example.test",
		"password":"correct horse battery staple",
		"language":"en",
		"theme":"system",
		"timezone":"UTC"
	}`, "", "")
	defer register.Body.Close()
	var payload struct {
		Data struct {
			Session struct { AccessToken string `json:"accessToken"` } `json:"session"`
		} `json:"data"`
	}
	if err := json.NewDecoder(register.Body).Decode(&payload); err != nil { t.Fatal(err) }
	valid := "Bearer " + payload.Data.Session.AccessToken

	api := &API{auth: service}
	app := fiber.New()
	app.Get("/protected", func(c fiber.Ctx) error {
		subject, err := api.subjectID(c)
		if err != nil { return c.SendStatus(http.StatusUnauthorized) }
		return c.Status(http.StatusOK).SendString(subject)
	})

	cases := []struct{
		name string
		values []string
		want int
	}{
		{"leading_whitespace", []string{" " + valid}, http.StatusOK}, // net/http normalizes leading OWS before Fiber
		{"trailing_whitespace", []string{valid + " "}, http.StatusOK},
		{"double_space_after_scheme", []string{"Bearer  " + payload.Data.Session.AccessToken}, http.StatusOK},
		{"tab_after_scheme", []string{"Bearer\t" + payload.Data.Session.AccessToken}, http.StatusUnauthorized},
		{"uppercase_scheme", []string{"BEARER " + payload.Data.Session.AccessToken}, http.StatusUnauthorized},
		{"mixed_case_scheme", []string{"BeArEr " + payload.Data.Session.AccessToken}, http.StatusUnauthorized},
		{"three_values_valid_first", []string{valid, "Bearer synthetic.invalid", valid}, http.StatusOK},
		{"three_values_invalid_first", []string{"Bearer synthetic.invalid", valid, valid}, http.StatusUnauthorized},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			for _, v := range tc.values { req.Header.Add("Authorization", v) }
			resp, err := app.Test(req)
			if err != nil { t.Fatalf("request: %v", err) }
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			if resp.StatusCode != tc.want {
				t.Fatalf("status=%d want=%d values=%q", resp.StatusCode, tc.want, tc.values)
			}
			if resp.StatusCode == http.StatusOK && strings.TrimSpace(string(body)) != store.user.InvestmentSubjectID {
				t.Fatalf("resolved principal mismatch got=%q want=%q", strings.TrimSpace(string(body)), store.user.InvestmentSubjectID)
			}
			t.Logf("AUTH_REP_CASE=%s STATUS=%d", tc.name, resp.StatusCode)
		})
	}
}

func TestModule1RefreshCookiePhysicalHeaderCollisionMatrix(t *testing.T) {
	requireModule1HTTPAdversarial(t)
	for _, tc := range []struct{
		name string
		headers []string
		wantOK bool
	}{
		{"two_headers_valid_then_invalid", []string{"VALID", "INVALID"}, true},
		{"two_headers_invalid_then_valid", []string{"INVALID", "VALID"}, false},
		{"duplicate_equal", []string{"VALID", "VALID"}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := &httpAuthTestStore{}
			app := newHTTPAuthApp(t, newModule1OwnershipService(t, store))
			register := fixtureLegacyLogin(t, app, http.MethodPost, "/api/v1/auth/register", `{
				"email":"physical-cookie-`+tc.name+`@example.test",
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
			for _, marker := range tc.headers {
				value := "synthetic-invalid"
				if marker == "VALID" { value = refresh.Value }
				req.Header.Add("Cookie", auth.RefreshCookieName+"="+value)
			}
			resp, err := app.Test(req)
			if err != nil { t.Fatalf("refresh: %v", err) }
			defer resp.Body.Close()
			ok := resp.StatusCode == http.StatusOK
			if ok != tc.wantOK { t.Fatalf("status=%d wantOK=%t", resp.StatusCode, tc.wantOK) }
			if ok {
				if _, exists := store.sessions[tokenHashForModule1(refresh.Value)]; exists {
					t.Fatal("successful refresh left old credential active")
				}
				if len(store.sessions) != 1 { t.Fatalf("rotated session count=%d", len(store.sessions)) }
			} else {
				if _, exists := store.sessions[tokenHashForModule1(refresh.Value)]; !exists {
					t.Fatal("rejected cookie collision mutated valid session")
				}
			}
			t.Logf("COOKIE_HEADER_CASE=%s STATUS=%d ACTIVE_SESSIONS=%d", tc.name, resp.StatusCode, len(store.sessions))
		})
	}
}


func TestModule1MalformedCookieAndContentTypeMatrix(t *testing.T) {
	requireModule1HTTPAdversarial(t)

	t.Run("refresh_malformed_cookie_rejected_without_state_mutation", func(t *testing.T) {
		store := &httpAuthTestStore{}
		app := newHTTPAuthApp(t, newHTTPAuthService(t, store))
		register := fixtureLegacyLogin(t, app, http.MethodPost, "/api/v1/auth/register", `{
			"email":"malformed-cookie@example.test",
			"password":"correct horse battery staple",
			"language":"en",
			"theme":"system",
			"timezone":"UTC"
		}`, "", "")
		refresh := requireCookie(t, register, auth.RefreshCookieName)
		csrf := readCSRFToken(t, register)
		register.Body.Close()

		before := len(store.sessions)
		for _, raw := range []string{
			auth.RefreshCookieName,
			auth.RefreshCookieName + "=",
			auth.RefreshCookieName + "=synthetic-invalid",
			"unrelated=synthetic; " + auth.RefreshCookieName + "=synthetic-invalid",
		} {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
			req.Header.Set("X-CSRF-Token", csrf)
			req.Header.Add("Cookie", raw)
			resp, err := app.Test(req)
			if err != nil { t.Fatalf("cookie=%q request: %v", raw, err) }
			resp.Body.Close()
			if resp.StatusCode != http.StatusUnauthorized {
				t.Fatalf("cookie=%q status=%d want=401", raw, resp.StatusCode)
			}
			if len(store.sessions) != before {
				t.Fatalf("cookie=%q mutated session count before=%d after=%d", raw, before, len(store.sessions))
			}
			if _, ok := store.sessions[tokenHashForModule1(refresh.Value)]; !ok {
				t.Fatalf("cookie=%q altered valid session while rejecting malformed material", raw)
			}
		}
		t.Log("MALFORMED_REFRESH_COOKIE_PROMOTED_TO_VALID_SESSION=NO")
	})

	t.Run("registration_content_type_representation", func(t *testing.T) {
		body := `{"email":"content-type@example.test"}`
		cases := []struct{
			name string
			contentType string
		}{
			{"application_json", "application/json"},
			{"application_json_charset", "application/json; charset=utf-8"},
			{"text_plain", "text/plain"},
			{"missing", ""},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				store := &httpOwnershipFixture{}
				app := newHTTPAuthApp(t, newModule1OwnershipService(t, store))
				req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(body))
				if tc.contentType != "" { req.Header.Set("Content-Type", tc.contentType) }
				resp, err := app.Test(req)
				if err != nil { t.Fatalf("request: %v", err) }
				defer resp.Body.Close()

				// The current handler decodes the JSON bytes directly and does not use Content-Type
				// as an authentication/authorization decision. Record that transport policy, while
				// requiring the resolved identity to remain exactly the submitted synthetic identity.
				if resp.StatusCode != http.StatusAccepted || store.registerCalls != 0 || len(store.sessions) != 0 {
					t.Fatalf("content-type=%q status=%d registerCalls=%d sessions=%d", tc.contentType, resp.StatusCode, store.registerCalls, len(store.sessions))
				}
				if store.email != "content-type@example.test" {
					t.Fatalf("content-type=%q changed resolved registration intent: %q", tc.contentType, store.email)
				}
				t.Logf("AUTH_CONTENT_TYPE_CASE=%s STATUS=%d RESOLVED_EMAIL=%q", tc.name, resp.StatusCode, store.email)
			})
		}
		t.Log("AUTH_CONTENT_TYPE_POLICY=BODY_JSON_DECODE_INDEPENDENT_OF_CONTENT_TYPE; PRINCIPAL_CONFUSION_OBSERVED=NO")
	})
}
