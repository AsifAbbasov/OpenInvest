package httpapi

import (
	"fmt"
	"io"
	"log"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
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

func TestModule12SafeLogAndErrorConformance(t *testing.T) {
	values := map[string]string{}
	for _, k := range []string{"ACCESS_TOKEN", "REFRESH_COOKIE", "AUTHORIZATION_BEARER", "PROVIDER_TOKEN", "DATABASE_PASSWORD", "REDIS_PASSWORD", "IMPORT_REVIEW_TOKEN", "CSRF_TOKEN", "FINANCIAL_PAYLOAD_SENTINEL"} {
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
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
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
	output := restore()
	for k, v := range values {
		if strings.Contains(output, v) || strings.Contains(output, url.QueryEscape(v)) {
			failures = append(failures, "log disclosed "+k)
		}
	}
	for _, f := range failures {
		t.Error(f)
	}
	t.Logf("M12_SAFE_LOG_ERROR_CASES=5 CAPTURED_LOG_BYTES=%d SENTINEL_CATEGORIES=9 FINANCIAL_POLICY=NO_PAYLOAD_LOGGING_OBSERVED_IN_TESTED_PATHS", len(output))
}
