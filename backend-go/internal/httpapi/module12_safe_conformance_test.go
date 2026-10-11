package httpapi

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestModule12SafeCORSHostConformance(t *testing.T) {
	t.Setenv("OPENINVEST_ALLOWED_WEB_ORIGINS", "https://web.example.test")
	for _, tc := range []struct {
		name, origin, host, proto string
		preflight, allowed        bool
	}{
		{"allowed", "https://web.example.test", "api.example.test", "", false, true},
		{"disallowed", "https://other.example.test", "api.example.test", "", false, false},
		{"null", "null", "api.example.test", "", false, false},
		{"missing", "", "api.example.test", "", false, false},
		{"allowed_preflight", "https://web.example.test", "api.example.test", "", true, true},
		{"disallowed_preflight", "https://other.example.test", "api.example.test", "", true, false},
		{"arbitrary_host", "https://web.example.test", "other.example.test", "", false, true},
		{"forwarded_http", "https://web.example.test", "api.example.test", "http", false, true},
		{"forwarded_https", "https://web.example.test", "api.example.test", "https", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			app := newHTTPAuthApp(t, newHTTPAuthService(t, &httpAuthTestStore{}))
			method, body := "POST", `{"email":"safe@example.test","password":"correct horse battery staple","language":"en","theme":"system","timezone":"UTC"}`
			if tc.preflight {
				method, body = "OPTIONS", ""
			}
			req := httptest.NewRequest(method, "https://api.example.test/api/v1/auth/register", strings.NewReader(body))
			req.Host = tc.host
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Origin", tc.origin)
			if tc.proto != "" {
				req.Header.Set("X-Forwarded-Proto", tc.proto)
				req.Header.Set("X-Forwarded-Host", "other.example.test")
			}
			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			expected := 201
			if tc.preflight {
				expected = 204
			}
			if resp.StatusCode != expected {
				t.Errorf("status %d expected %d", resp.StatusCode, expected)
			}
			if tc.allowed {
				if resp.Header.Get("Access-Control-Allow-Origin") != tc.origin || resp.Header.Get("Access-Control-Allow-Credentials") != "true" {
					t.Error("CORS mismatch")
				}
			} else if resp.Header.Get("Access-Control-Allow-Origin") != "" || resp.Header.Get("Access-Control-Allow-Credentials") != "" {
				t.Error("unexpected CORS permission")
			}
			if resp.Header.Get("Location") != "" {
				t.Error("unexpected redirect")
			}
			for _, c := range resp.Cookies() {
				if c.Domain != "" || !c.Secure {
					t.Error("cookie attributes changed")
				}
			}
			if !tc.preflight && !strings.Contains(resp.Header.Get("Cache-Control"), "no-store") {
				t.Error("missing no-store")
			}
			t.Logf("M12_SAFE_CORS_HOST case=%s status=%d ACAO=%q ACAC=%q CACHE=%q VARY=%q LOCATION=%q CSP=%q ETAG=%q SET_COOKIE_COUNT=%d", tc.name, resp.StatusCode, resp.Header.Get("Access-Control-Allow-Origin"), resp.Header.Get("Access-Control-Allow-Credentials"), resp.Header.Get("Cache-Control"), resp.Header.Get("Vary"), resp.Header.Get("Location"), resp.Header.Get("Content-Security-Policy"), resp.Header.Get("ETag"), len(resp.Cookies()))
		})
	}
}

type module12LogFixtureProvider struct{ err error }

func (p module12LogFixtureProvider) CorporateActions(context.Context, verticalslice.CorporateActionQuery) ([]verticalslice.CorporateActionEvent, error) {
	return nil, p.err
}

type module12UnavailableBudget struct{ err error }

func (b module12UnavailableBudget) Ping(context.Context) error { return b.err }
func (b module12UnavailableBudget) AdmitEndpoint(context.Context, string, int, int, time.Duration) (bool, error) {
	return false, b.err
}
func (b module12UnavailableBudget) AdmitProvider(context.Context, int, time.Duration) (bool, error) {
	return false, b.err
}
func (b module12UnavailableBudget) ObserveProviderAllowance(context.Context, int, time.Duration) error {
	return b.err
}
func (b module12UnavailableBudget) Close() error { return nil }

func TestModule12SafeLogAndErrorConformance(t *testing.T) {
	values := map[string]string{}
	for _, k := range []string{"ACCESS_TOKEN", "REFRESH_COOKIE", "AUTHORIZATION_BEARER", "PROVIDER_TOKEN", "DATABASE_PASSWORD", "REDIS_PASSWORD", "IMPORT_REVIEW_TOKEN", "CSRF_TOKEN", "FINANCIAL_PAYLOAD_SENTINEL", "USER_EMAIL_SENTINEL"} {
		values[k] = "M12_SAFE_" + k + "_7654c9"
	}
	oldOut, oldErr, oldLog := os.Stdout, os.Stderr, log.Writer()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	os.Stderr = w
	log.SetOutput(w)
	done := make(chan string, 1)
	go func() { b, _ := io.ReadAll(r); done <- string(b) }()
	restored := false
	restore := func() string {
		os.Stdout = oldOut
		os.Stderr = oldErr
		log.SetOutput(oldLog)
		w.Close()
		s := <-done
		r.Close()
		restored = true
		return s
	}
	defer func() {
		if !restored {
			restore()
		}
	}()
	var failures []string
	for _, tc := range []struct {
		name, method, path, body string
		status                   int
		failure                  bool
	}{
		{"success", "POST", "/api/v1/auth/register", `{"email":"logs@example.test","password":"correct horse battery staple","language":"en","theme":"system","timezone":"UTC"}`, 201, false},
		{"validation", "POST", "/api/v1/auth/register", `{}`, 400, false},
		{"malformed_request", "POST", "/api/v1/auth/register", `{`, 400, false},
		{"invalid_authentication", "POST", "/api/v1/auth/login", `{"email":"logs@example.test","password":"wrong password material"}`, 401, false},
		{"unauthorized", "GET", "/api/v1/portfolios", "", 401, false},
		{"not_found", "GET", "/api/v1/not-present", "", 404, false},
		{"dependency_error", "POST", "/api/v1/auth/login", `{"email":"logs@example.test","password":"correct horse battery staple"}`, 500, true},
	} {
		store := &httpAuthTestStore{}
		if tc.failure {
			var all []string
			for _, v := range values {
				all = append(all, v)
			}
			store.findErr = fmt.Errorf("synthetic dependency unavailable postgres://fixture:password@db.internal:5432/db redis://redis.internal:6379 /workspace/private/store.go SELECT secret: %s", strings.Join(all, " "))
		}
		app := newHTTPAuthApp(t, newHTTPAuthService(t, store))
		body := strings.ReplaceAll(tc.body, "logs@example.test", values["USER_EMAIL_SENTINEL"]+"@example.test")
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+values["AUTHORIZATION_BEARER"])
		req.Header.Set("Cookie", "refresh="+values["REFRESH_COOKIE"])
		req.Header.Set("X-CSRF-Token", values["CSRF_TOKEN"])
		resp, e := app.Test(req)
		if e != nil {
			failures = append(failures, tc.name+": request failed")
			continue
		}
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != tc.status {
			failures = append(failures, fmt.Sprintf("%s status %d expected %d", tc.name, resp.StatusCode, tc.status))
		}
		for k, v := range values {
			if strings.Contains(string(b), v) {
				failures = append(failures, tc.name+": response disclosed "+k)
			}
		}
		for _, v := range []string{"panic:", "goroutine ", "/workspace/", "SELECT ", "postgres://", "redis://"} {
			if strings.Contains(string(b), v) {
				failures = append(failures, tc.name+": metadata disclosed")
			}
		}
	}

	// Controlled 403/409 serializer fixtures use the repository's production
	// error writer and cache middleware; no product behavior is changed.
	fixtureApp := fiber.New()
	fixtureApp.Use(sensitiveResponseCachePolicy)
	fixtureApp.Get("/m12/forbidden", func(c fiber.Ctx) error {
		return writeErrorWithMeta(c, requestMeta(c), http.StatusForbidden, "FORBIDDEN", "Forbidden")
	})
	fixtureApp.Get("/m12/conflict", func(c fiber.Ctx) error {
		return writeErrorWithMeta(c, requestMeta(c), http.StatusConflict, "CONFLICT", "Conflict")
	})
	for _, fixture := range []struct {
		name, path string
		status int
	}{
		{"forbidden", "/m12/forbidden", http.StatusForbidden},
		{"conflict", "/m12/conflict", http.StatusConflict},
	} {
		req := httptest.NewRequest(http.MethodGet, fixture.path, nil)
		req.Header.Set("Authorization", "Bearer "+values["AUTHORIZATION_BEARER"])
		req.Header.Set("X-CSRF-Token", values["CSRF_TOKEN"])
		resp, e := fixtureApp.Test(req)
		if e != nil {
			failures = append(failures, fixture.name+": request failed")
			continue
		}
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != fixture.status {
			failures = append(failures, fmt.Sprintf("%s status %d expected %d", fixture.name, resp.StatusCode, fixture.status))
		}
		for k, v := range values {
			if strings.Contains(string(b), v) { failures = append(failures, fixture.name+": response disclosed "+k) }
		}
	}

	// Exercise the real per-credential auth limiter until a repository-owned
	// 429 is returned. Registration creates the synthetic identity; repeated
	// wrong-password logins are bounded and never use a real credential.
	rateStore := &httpAuthTestStore{}
	rateApp := newHTTPAuthApp(t, newHTTPAuthService(t, rateStore))
	registerBody := `{"email":"m12-rate@example.test","password":"correct horse battery staple","language":"en","theme":"system","timezone":"UTC"}`
	registerReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(registerBody))
	registerReq.Header.Set("Content-Type", "application/json")
	registerResp, registerErr := rateApp.Test(registerReq)
	if registerErr != nil || registerResp.StatusCode != http.StatusCreated {
		failures = append(failures, "rate_limit: synthetic registration failed")
	}
	if registerResp != nil { registerResp.Body.Close() }
	rateStatus := 0
	for i := 0; i < defaultAuthCredentialLimit+2; i++ {
		body := `{"email":"m12-rate@example.test","password":"definitely wrong synthetic password"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+values["AUTHORIZATION_BEARER"])
		resp, e := rateApp.Test(req)
		if e != nil { failures = append(failures, "rate_limit: request failed"); break }
		rateStatus = resp.StatusCode
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		for k, v := range values {
			if strings.Contains(string(b), v) { failures = append(failures, "rate_limit: response disclosed "+k) }
		}
		if rateStatus == http.StatusTooManyRequests { break }
	}
	if rateStatus != http.StatusTooManyRequests {
		failures = append(failures, fmt.Sprintf("rate_limit status %d expected %d", rateStatus, http.StatusTooManyRequests))
	}

	providerCases := []struct {
		name string
		provider verticalslice.CorporateActionProvider
		budget module12UnavailableBudget
		useBudget bool
	}{
		{"provider_error", module12LogFixtureProvider{err: errors.New("synthetic provider failure")}, module12UnavailableBudget{}, false},
		{"provider_timeout", module12LogFixtureProvider{err: context.DeadlineExceeded}, module12UnavailableBudget{}, false},
		{"redis_unavailable", module12LogFixtureProvider{}, module12UnavailableBudget{err: fmt.Errorf("synthetic redis unavailable: %s", values["REDIS_PASSWORD"])}, true},
	}
	for _, pc := range providerCases {
		api := &API{
			corporateActionProvider: pc.provider,
			corporateActionLimiter: newCorporateActionProjectionRateLimiter(),
			corporateActionCoalescer: newCorporateActionProjectionCoalescer(),
		}
		if pc.useBudget { api.corporateActionBudget = pc.budget }
		app := newReplayApp(api)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/corporate-actions/projection?instrumentId=SBER&from=2026-01-01&to=2026-12-31", nil)
		req.Header.Set("Authorization", "Bearer "+values["AUTHORIZATION_BEARER"])
		resp, e := app.Test(req)
		if e != nil {
			failures = append(failures, pc.name+": request failed")
			continue
		}
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode < 500 || resp.StatusCode > 503 {
			failures = append(failures, fmt.Sprintf("%s status %d expected 5xx fail-closed", pc.name, resp.StatusCode))
		}
		for k, v := range values {
			if strings.Contains(string(b), v) { failures = append(failures, pc.name+": response disclosed "+k) }
		}
		for _, needle := range []string{"panic:", "goroutine ", "/workspace/", "SELECT ", "postgres://", "redis://"} {
			if strings.Contains(string(b), needle) { failures = append(failures, pc.name+": metadata disclosed") }
		}
	}

	output := restore()
	for k, v := range values {
		if strings.Contains(output, v) || strings.Contains(output, url.QueryEscape(v)) {
			failures = append(failures, "log disclosed "+k)
		}
	}
	for _, f := range failures {
		t.Error(f)
	}
	t.Logf("M12_SAFE_LOG_ERROR_CASES=14 CAPTURED_LOG_BYTES=%d SENTINEL_CATEGORIES=10 FINANCIAL_POLICY=NO_PAYLOAD_LOGGING_OBSERVED_IN_TESTED_PATHS", len(output))
}
