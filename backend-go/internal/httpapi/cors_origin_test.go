package httpapi

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/openinvest/openinvest/backend-go/internal/auth"
)

func TestCORSOriginCardinality(t *testing.T) {
	const allowed = "https://allowed.example.test"
	const disallowed = "https://other.example.test"
	t.Setenv("OPENINVEST_ALLOWED_WEB_ORIGINS", allowed)
	cases := []struct {
		name      string
		origins   []string
		ambiguous bool
		permitted bool
	}{
		{"single_allowed_origin", []string{allowed}, false, true},
		{"single_disallowed_origin", []string{disallowed}, false, false},
		{"origin_null", []string{"null"}, false, false},
		{"missing_origin", nil, false, false},
		{"duplicate_equal_origin", []string{allowed, allowed}, true, false},
		{"duplicate_conflicting_origin_allowed_first", []string{allowed, disallowed}, true, false},
		{"duplicate_conflicting_origin_disallowed_first", []string{disallowed, allowed}, true, false},
		{"three_origin_fields", []string{allowed, allowed, disallowed}, true, false},
		{"comma_combined_allowed_and_disallowed", []string{allowed + ", " + disallowed}, true, false},
		{"comma_combined_equal_origins", []string{allowed + ", " + allowed}, true, false},
	}
	for _, tc := range cases {
		for _, method := range []string{http.MethodPost, http.MethodOptions} {
			t.Run(tc.name+"/"+method, func(t *testing.T) {
				var fields, executed atomic.Int32
				app := fiber.New()
				app.Use(func(c fiber.Ctx) error {
					fields.Store(int32(len(c.Request().Header.PeekAll("Origin"))))
					return c.Next()
				})
				app.Use(localDevelopmentCORS)
				app.All("/mutation", func(c fiber.Ctx) error {
					executed.Add(1)
					return c.SendStatus(http.StatusOK)
				})
				listener, err := net.Listen("tcp", "127.0.0.1:0")
				if err != nil {
					t.Fatal(err)
				}
				serving := make(chan error, 1)
				go func() { serving <- app.Listener(listener) }()
				t.Cleanup(func() {
					ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
					defer cancel()
					if err := app.ShutdownWithContext(ctx); err != nil {
						t.Errorf("shutdown: %v", err)
					}
					_ = listener.Close()
					select {
					case <-serving:
					case <-ctx.Done():
						t.Error("listener did not stop")
					}
				})
				conn, err := net.DialTimeout("tcp", listener.Addr().String(), time.Second)
				if err != nil {
					t.Fatal(err)
				}
				defer conn.Close()
				if err := conn.SetDeadline(time.Now().Add(2 * time.Second)); err != nil {
					t.Fatal(err)
				}
				var request strings.Builder
				fmt.Fprintf(&request, "%s /mutation HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\nContent-Length: 0\r\n", method)
				for i, origin := range tc.origins {
					// Mixed casing also exercises HTTP's case-insensitive field names.
					name := "Origin"
					if i%2 == 1 {
						name = "oRiGiN"
					}
					fmt.Fprintf(&request, "%s: %s\r\n", name, origin)
				}
				if method == http.MethodOptions {
					request.WriteString("Access-Control-Request-Method: POST\r\nAccess-Control-Request-Headers: Content-Type\r\n")
				}
				request.WriteString("\r\n")
				if _, err := io.WriteString(conn, request.String()); err != nil {
					t.Fatal(err)
				}
				response, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: method})
				if err != nil {
					t.Fatal(err)
				}
				defer response.Body.Close()
				body, err := io.ReadAll(response.Body)
				if err != nil {
					t.Fatal(err)
				}
				if got := fields.Load(); got != int32(len(tc.origins)) {
					t.Errorf("physical Origin fields at Fiber boundary = %d, want %d", got, len(tc.origins))
				}
				wantStatus, wantExecuted := http.StatusOK, int32(1)
				if method == http.MethodOptions {
					wantStatus, wantExecuted = http.StatusNoContent, 0
				}
				if tc.ambiguous {
					wantStatus, wantExecuted = http.StatusBadRequest, 0
				}
				if response.StatusCode != wantStatus {
					t.Errorf("status = %d, want %d", response.StatusCode, wantStatus)
				}
				if executed.Load() != wantExecuted {
					t.Errorf("handler executed = %d, want %d", executed.Load(), wantExecuted)
				}
				wantOrigin, wantCredentials := "", ""
				if tc.permitted {
					wantOrigin, wantCredentials = allowed, "true"
				}
				if got := response.Header.Get("Access-Control-Allow-Origin"); got != wantOrigin {
					t.Errorf("ACAO = %q, want %q", got, wantOrigin)
				}
				if got := response.Header.Get("Access-Control-Allow-Credentials"); got != wantCredentials {
					t.Errorf("ACAC = %q, want %q", got, wantCredentials)
				}
				if tc.ambiguous {
					for _, origin := range tc.origins {
						if bytes.Contains(body, []byte(origin)) {
							t.Error("rejection reflects Origin")
						}
					}
					for name := range response.Header {
						if strings.HasPrefix(strings.ToLower(name), "access-control-allow-") {
							t.Errorf("ambiguous Origin emits %s", name)
						}
					}
				}
				if tc.permitted {
					if response.Header.Get("Vary") != "Origin" {
						t.Error("missing Vary: Origin")
					}
					if !strings.Contains(response.Header.Get("Access-Control-Allow-Methods"), "POST") {
						t.Error("missing POST permission")
					}
					if !strings.Contains(response.Header.Get("Access-Control-Allow-Headers"), "X-CSRF-Token") {
						t.Error("missing CSRF header permission")
					}
				}
				result := "PASS"
				if t.Failed() {
					result = "FAIL"
				}
				t.Logf("M12_CORS_ORIGIN_CASE=%s METHOD=%s PHYSICAL_ORIGIN_FIELDS=%d STATUS=%d ACAO=%q ACAC=%q HANDLER_EXECUTED=%d RESULT=%s", tc.name, method, fields.Load(), response.StatusCode, response.Header.Get("Access-Control-Allow-Origin"), response.Header.Get("Access-Control-Allow-Credentials"), executed.Load(), result)
			})
		}
	}
}

func TestCORSOriginCompatibility(t *testing.T) {
	const allowed = "https://allowed.example.test"
	t.Setenv("OPENINVEST_ALLOWED_WEB_ORIGINS", allowed)
	for _, variant := range []struct{ name, host, forwardedHost, forwarded string }{
		{"ordinary", "localhost", "", ""},
		{"host", "alternative.example.test", "", ""},
		{"forwarded_host", "localhost", "proxy.example.test", ""},
		{"forwarded_http", "localhost", "", "host=proxy.example.test;proto=http"},
		{"forwarded_https", "localhost", "", "host=proxy.example.test;proto=https"},
	} {
		t.Run(variant.name, func(t *testing.T) {
			app := newHTTPAuthApp(t, newHTTPAuthService(t, &httpAuthTestStore{}))
			seedLegacyFixture(t, app, `{"email":"cors-compat@example.test","password":"correct horse battery staple","language":"en","theme":"system","timezone":"UTC"}`)
			request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":"cors-compat@example.test","password":"correct horse battery staple"}`))
			request.Host = variant.host
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Origin", allowed)
			if variant.forwardedHost != "" {
				request.Header.Set("X-Forwarded-Host", variant.forwardedHost)
			}
			if variant.forwarded != "" {
				request.Header.Set("Forwarded", variant.forwarded)
			}
			response, err := app.Test(request)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			assertNoStoreResponse(t, response, http.StatusOK)
			cookie := requireCookie(t, response, auth.RefreshCookieName)
			if !cookie.Secure || !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode || cookie.Domain != "" {
				t.Errorf("refresh cookie policy changed: %+v", cookie)
			}
			if response.Header.Get("Access-Control-Allow-Origin") != allowed || response.Header.Get("Access-Control-Allow-Credentials") != "true" {
				t.Error("single-Origin auth CORS permission changed")
			}
			if response.Header.Get("Location") != "" {
				t.Error("unexpected redirect")
			}
			result := "PASS"
			if t.Failed() {
				result = "FAIL"
			}
			t.Logf("M12_CORS_COMPATIBILITY_CASE=%s AUTH_REGISTRATION_STATUS=%d COOKIE_SECURE=%t COOKIE_HTTPONLY=%t COOKIE_SAMESITE=%d CACHE_CONTROL=%q RESULT=%s", variant.name, response.StatusCode, cookie.Secure, cookie.HttpOnly, cookie.SameSite, response.Header.Get("Cache-Control"), result)
		})
	}
}
