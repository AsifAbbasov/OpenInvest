package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
)

func requireModule12V2Adversarial(t *testing.T) {
	t.Helper()
	if os.Getenv("OPENINVEST_MODULE12_V2_ADVERSARIAL") != "1" {
		t.Skip("Module 12 V2 adversarial matrix runs only in the specialized workflow")
	}
}

type m12V2Listener struct {
	app *fiber.App
	ln net.Listener
	url string
}

func startM12V2Listener(t *testing.T, app *fiber.App) *m12V2Listener {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil { t.Fatal(err) }
	s := &m12V2Listener{app: app, ln: ln, url: "http://" + ln.Addr().String()}
	done := make(chan error, 1)
	go func() { done <- app.Listener(ln) }()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = app.ShutdownWithContext(ctx)
		_ = ln.Close()
		select { case <-done: case <-time.After(2*time.Second): }
	})
	return s
}

func m12V2Do(t *testing.T, method, url string, headers http.Header, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil { t.Fatal(err) }
	for k, vv := range headers {
		for _, v := range vv { req.Header.Add(k, v) }
	}
	resp, err := (&http.Client{Timeout: 3*time.Second}).Do(req)
	if err != nil { t.Fatalf("%s %s: %v", method, url, err) }
	return resp
}

func m12V2IdentityAppWithLimiter(t *testing.T, cfg HTTPNetworkConfig, limiter *authRateLimiter) *fiber.App {
	t.Helper()
	app := newFiberApp(cfg)
	app.Get("/m12/identity", func(c fiber.Ctx) error {
		ip, err := normalizedClientIP(c)
		if err != nil { return c.Status(http.StatusBadRequest).SendString("invalid") }
		return c.SendString(ip)
	})
	app.Get("/m12/rate", func(c fiber.Ctx) error {
		ip, err := normalizedClientIP(c)
		if err != nil { return c.Status(http.StatusBadRequest).SendString("invalid") }
		if !limiter.allow("m12|"+ip, time.Now().UTC()) {
			return c.Status(http.StatusTooManyRequests).SendString(ip)
		}
		return c.SendString(ip)
	})
	return app
}

func m12V2IdentityApp(t *testing.T, cfg HTTPNetworkConfig, limit int) *fiber.App {
	t.Helper()
	return m12V2IdentityAppWithLimiter(t, cfg, newBoundedAuthRateLimiter(limit, 100000, 10000, time.Minute))
}

func TestModule12V2ForwardedHeaderIdentityAdversarialMatrix(t *testing.T) {
	requireModule12V2Adversarial(t)
	trusted, err := NewHTTPNetworkConfig(true, []string{"127.0.0.1/32"})
	if err != nil { t.Fatal(err) }
	untrusted, err := NewHTTPNetworkConfig(true, []string{"127.0.0.2/32"})
	if err != nil { t.Fatal(err) }

	type variant struct { name string; h http.Header }
	variants := []variant{
		{"valid_ipv4", http.Header{"X-Forwarded-For": {"198.51.100.10"}}},
		{"valid_ipv6", http.Header{"X-Forwarded-For": {"2001:db8::10"}}},
		{"ipv4_mapped_ipv6", http.Header{"X-Forwarded-For": {"::ffff:198.51.100.11"}}},
		{"xff_single", http.Header{"X-Forwarded-For": {"198.51.100.12"}}},
		{"xff_chain", http.Header{"X-Forwarded-For": {"198.51.100.13, 203.0.113.4"}}},
		{"xff_leftmost_spoof", http.Header{"X-Forwarded-For": {"192.0.2.55, 198.51.100.14"}}},
		{"xff_rightmost_spoof", http.Header{"X-Forwarded-For": {"198.51.100.15, 192.0.2.56"}}},
		{"duplicate_xff_equal", http.Header{"X-Forwarded-For": {"198.51.100.16", "198.51.100.16"}}},
		{"duplicate_xff_conflicting", http.Header{"X-Forwarded-For": {"198.51.100.17", "203.0.113.17"}}},
		{"forwarded_single", http.Header{"Forwarded": {"for=198.51.100.18"}}},
		{"forwarded_multiple", http.Header{"Forwarded": {"for=198.51.100.19;proto=https, for=203.0.113.19"}}},
		{"forwarded_quoted_for", http.Header{"Forwarded": {"for=\"198.51.100.20\""}}},
		{"forwarded_unknown", http.Header{"Forwarded": {"for=unknown"}}},
		{"forwarded_obfuscated_identifier", http.Header{"Forwarded": {"for=_hidden"}}},
		{"empty_element", http.Header{"X-Forwarded-For": {"198.51.100.21, , 203.0.113.21"}}},
		{"whitespace_variants", http.Header{"X-Forwarded-For": {" 198.51.100.22 ,\t203.0.113.22 "}}},
		{"malformed_ip", http.Header{"X-Forwarded-For": {"not-an-ip"}}},
		{"mixed_ipv4_ipv6_chain", http.Header{"X-Forwarded-For": {"198.51.100.23, 2001:db8::23"}}},
		{"xff_vs_forwarded_conflict", http.Header{"X-Forwarded-For": {"198.51.100.24"}, "Forwarded": {"for=203.0.113.24"}}},
		{"xff_vs_cf_connecting_ip_conflict", http.Header{"X-Forwarded-For": {"198.51.100.25"}, "CF-Connecting-IP": {"203.0.113.25"}}},
		{"xff_vs_x_real_ip_conflict", http.Header{"X-Forwarded-For": {"198.51.100.26"}, "X-Real-IP": {"203.0.113.26"}}},
		{"host_proto_conflict", http.Header{"X-Forwarded-Host": {"evil.example"}, "X-Forwarded-Proto": {"http"}}},
	}
	modes := []struct{name string; cfg HTTPNetworkConfig}{
		{"DIRECT", HTTPNetworkConfig{}},
		{"TRUSTED_PROXY", trusted},
		{"UNTRUSTED_PEER", untrusted},
	}
	for _, mode := range modes {
		mode := mode
		t.Run(mode.name, func(t *testing.T) {
			s := startM12V2Listener(t, m12V2IdentityApp(t, mode.cfg, 12))
			for _, tc := range variants {
				tc := tc
				t.Run(tc.name, func(t *testing.T) {
					var got [2]string
					var status [2]int
					for i := 0; i < 2; i++ {
						resp := m12V2Do(t, http.MethodGet, s.url+"/m12/identity", tc.h, "")
						status[i] = resp.StatusCode
						b, _ := io.ReadAll(resp.Body)
						resp.Body.Close()
						got[i] = strings.TrimSpace(string(b))
					}
					if got[0] != got[1] { t.Errorf("non-deterministic resolved identity: %q vs %q", got[0], got[1]) }
					if got[0] != "invalid" {
						if _, err := netip.ParseAddr(got[0]); err != nil { t.Errorf("non-canonical identity %q", got[0]) }
					}
					if mode.name != "TRUSTED_PROXY" && got[0] != "127.0.0.1" {
						t.Errorf("untrusted/direct forwarding header spoof changed identity to %q", got[0])
					}
					t.Logf("CASE=%s MODE=%s RAW_HEADER_VARIANT=%q RESOLVED_CLIENT_IDENTITY=%s STATUS=%d RATE_LIMIT_ADMITTED=NOT_APPLICABLE PROVIDER_CALLS=0", tc.name, mode.name, tc.h, got[0], status[0])
				})
			}
		})
	}

	for _, replicas := range []int{1,2,4} {
		replicas := replicas
		t.Run(fmt.Sprintf("same_identity_replicas_%d", replicas), func(t *testing.T) {
			// This matrix is about forwarded-header identity stability across listeners.
			// Use one shared admission authority so the fixture does not manufacture a
			// false "distributed bypass" merely by constructing N independent limiters.
			// The repository's actual Redis-backed shared-budget behavior is exercised
			// separately by the shared-budget-cancellation specialized job.
			sharedLimiter := newBoundedAuthRateLimiter(12, 100000, 10000, time.Minute)
			servers := make([]*m12V2Listener, replicas)
			for i := range servers {
				servers[i] = startM12V2Listener(t, m12V2IdentityAppWithLimiter(t, trusted, sharedLimiter))
			}
			headers := http.Header{"X-Forwarded-For": {"198.51.100.88"}}
			admitted := 0
			for i := 0; i < 48; i++ {
				resp := m12V2Do(t, http.MethodGet, servers[i%replicas].url+"/m12/rate", headers, "")
				if resp.StatusCode == http.StatusOK { admitted++ }
				resp.Body.Close()
			}
			t.Logf("CASE=RATE_LIMIT_IDENTITY_CAMPAIGN MODE=TRUSTED_PROXY REPLICAS=%d SHARED_AUTHORITY=YES RAW_HEADER_VARIANT=X-Forwarded-For RESOLVED_CLIENT_IDENTITY=198.51.100.88 STATUS=MIXED_200_429 RATE_LIMIT_ADMITTED=%d PROVIDER_CALLS=0 EXPECTED_MAX=12", replicas, admitted)
			if admitted > 12 { t.Errorf("CROSS_REPLICA_IDENTITY_BUDGET_BYPASS=YES replicas=%d admitted=%d", replicas, admitted) }
		})
	}

	for _, campaign := range []struct {
		name string
		header func(int) http.Header
	}{
		{
			name: "trusted_proxy_spoof_prefix_rotation",
			header: func(i int) http.Header {
				return http.Header{"X-Forwarded-For": {fmt.Sprintf("198.51.100.%d, 203.0.113.99", 40+i)}}
			},
		},
		{
			name: "trusted_proxy_duplicate_xff_rotation",
			header: func(i int) http.Header {
				return http.Header{"X-Forwarded-For": {fmt.Sprintf("198.51.100.%d", 40+i), "203.0.113.99"}}
			},
		},
		{
			name: "trusted_proxy_conflicting_family_rotation",
			header: func(i int) http.Header {
				return http.Header{
					"X-Forwarded-For": {"203.0.113.99"},
					"Forwarded": {fmt.Sprintf("for=198.51.100.%d", 40+i)},
					"CF-Connecting-IP": {fmt.Sprintf("192.0.2.%d", 40+i)},
					"X-Real-IP": {fmt.Sprintf("198.18.0.%d", 40+i)},
				}
			},
		},
	} {
		campaign := campaign
		t.Run(campaign.name, func(t *testing.T) {
			server := startM12V2Listener(t, m12V2IdentityApp(t, trusted, 12))
			admitted := 0
			resolved := ""
			for i := 0; i < 36; i++ {
				h := campaign.header(i)
				identityResp := m12V2Do(t, http.MethodGet, server.url+"/m12/identity", h, "")
				b, _ := io.ReadAll(identityResp.Body)
				identityResp.Body.Close()
				resolved = strings.TrimSpace(string(b))
				resp := m12V2Do(t, http.MethodGet, server.url+"/m12/rate", h, "")
				if resp.StatusCode == http.StatusOK { admitted++ }
				resp.Body.Close()
			}
			t.Logf("CASE=%s MODE=TRUSTED_PROXY RESOLVED_CLIENT_IDENTITY=%s STATUS=MIXED_200_429 RATE_LIMIT_ADMITTED=%d PROVIDER_CALLS=0 EXPECTED_MAX=12", campaign.name, resolved, admitted)
			if admitted > 12 {
				t.Errorf("HEADER_IDENTITY_BUDGET_BYPASS=YES campaign=%s admitted=%d", campaign.name, admitted)
			}
		})
	}

	s := startM12V2Listener(t, m12V2IdentityApp(t, untrusted, 12))
	admitted := 0
	for i := 0; i < 36; i++ {
		h := http.Header{
			"X-Forwarded-For": {fmt.Sprintf("198.51.100.%d", 40+i)},
			"Forwarded": {fmt.Sprintf("for=203.0.113.%d", 40+i)},
			"CF-Connecting-IP": {fmt.Sprintf("192.0.2.%d", 40+i)},
			"X-Real-IP": {fmt.Sprintf("198.18.0.%d", 40+i)},
		}
		resp := m12V2Do(t, http.MethodGet, s.url+"/m12/rate", h, "")
		if resp.StatusCode == http.StatusOK { admitted++ }
		resp.Body.Close()
	}
	t.Logf("UNTRUSTED_PEER_ROTATION_ADMITTED=%d EXPECTED_MAX=12", admitted)
	if admitted > 12 { t.Errorf("UNTRUSTED_PEER_IDENTITY_SPOOF=YES admitted=%d", admitted) }
}

func m12V2Hash(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func TestModule12V2CacheCORSHostBoundaryAdversarialMatrix(t *testing.T) {
	requireModule12V2Adversarial(t)
	t.Setenv("OPENINVEST_ALLOWED_WEB_ORIGINS", "https://allowed.example.test")
	type tc struct {
		name, method, path, origin, host, body string
		headers http.Header
		allowed bool
		sensitive bool
	}
	cases := []tc{
		{"configured_allowed_origin","POST","/api/v1/auth/register","https://allowed.example.test","api.example.test","{\"email\":\"a@example.test\",\"password\":\"correct horse battery staple\",\"language\":\"en\",\"theme\":\"system\",\"timezone\":\"UTC\"}",nil,true,true},
		{"disallowed_origin","POST","/api/v1/auth/register","https://evil.example.test","api.example.test","{}",nil,false,true},
		{"origin_null","POST","/api/v1/auth/register","null","api.example.test","{}",nil,false,true},
		{"missing_origin","GET","/api/v1/portfolios","","api.example.test","",nil,false,true},
		{"duplicate_equal_origin","OPTIONS","/api/v1/auth/register","","api.example.test","",http.Header{"Origin":{"https://allowed.example.test","https://allowed.example.test"},"Access-Control-Request-Method":{"POST"}},true,false},
		{"duplicate_conflicting_origin","OPTIONS","/api/v1/auth/register","","api.example.test","",http.Header{"Origin":{"https://allowed.example.test","https://evil.example.test"},"Access-Control-Request-Method":{"POST"}},false,false},
		{"comma_combined_origin","OPTIONS","/api/v1/auth/register","https://allowed.example.test, https://evil.example.test","api.example.test","",http.Header{"Access-Control-Request-Method":{"POST"}},false,false},
		{"arbitrary_host","POST","/api/v1/auth/register","https://allowed.example.test","attacker.example.test","{}",nil,true,true},
		{"x_forwarded_host","POST","/api/v1/auth/register","https://allowed.example.test","api.example.test","{}",http.Header{"X-Forwarded-Host":{"attacker.example.test"}},true,true},
		{"forwarded_host","POST","/api/v1/auth/register","https://allowed.example.test","api.example.test","{}",http.Header{"Forwarded":{"host=attacker.example.test;proto=http"}},true,true},
		{"x_forwarded_proto_http","GET","/api/v1/portfolios","","api.example.test","",http.Header{"X-Forwarded-Proto":{"http"}},false,true},
		{"x_forwarded_proto_https","GET","/api/v1/portfolios","","api.example.test","",http.Header{"X-Forwarded-Proto":{"https"}},false,true},
		{"forwarded_proto_http","GET","/api/v1/portfolios","","api.example.test","",http.Header{"Forwarded":{"proto=http"}},false,true},
		{"forwarded_proto_https","GET","/api/v1/portfolios","","api.example.test","",http.Header{"Forwarded":{"proto=https"}},false,true},
		{"router_headers","GET","/api/v1/portfolios?limit=10","","api.example.test","",http.Header{"RSC":{"1"},"Next-Router-State-Tree":{"synthetic"},"Next-Router-Prefetch":{"1"},"Accept-Encoding":{"identity"}},false,true},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			app := newHTTPAuthApp(t, newHTTPAuthService(t, &httpAuthTestStore{}))
			req := httptest.NewRequest(c.method, "https://api.example.test"+c.path, strings.NewReader(c.body))
			req.Host = c.host
			if c.origin != "" { req.Header.Set("Origin", c.origin) }
			if c.body != "" { req.Header.Set("Content-Type", "application/json") }
			for k, vv := range c.headers { for _, v := range vv { req.Header.Add(k,v) } }
			resp, err := app.Test(req)
			if err != nil { t.Fatal(err) }
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			acao, acac := resp.Header.Get("Access-Control-Allow-Origin"), resp.Header.Get("Access-Control-Allow-Credentials")
			if c.allowed {
				expectedOrigin := c.origin
				if expectedOrigin == "" { expectedOrigin = "https://allowed.example.test" }
				if acao != expectedOrigin || acac != "true" { t.Errorf("allowed origin mismatch ACAO=%q ACAC=%q", acao, acac) }
			} else if acao != "" || acac != "" {
				t.Errorf("ambiguous/disallowed origin authorized ACAO=%q ACAC=%q", acao, acac)
			}
			if resp.Header.Get("Location") != "" { t.Errorf("host/proto controlled redirect %q", resp.Header.Get("Location")) }
			for _, cookie := range resp.Cookies() {
				if cookie.Domain != "" { t.Errorf("host controlled cookie domain %q", cookie.Domain) }
				if !cookie.Secure { t.Error("cookie Secure weakened") }
			}
			if c.sensitive && !strings.Contains(strings.ToLower(resp.Header.Get("Cache-Control")), "no-store") {
				t.Errorf("sensitive response lacks no-store: %q", resp.Header.Get("Cache-Control"))
			}
			t.Logf("CASE=%s STATUS=%d ACCESS_CONTROL_ALLOW_ORIGIN=%q ACCESS_CONTROL_ALLOW_CREDENTIALS=%q LOCATION=%q SET_COOKIE=%d CACHE_CONTROL=%q VARY=%q ETAG=%q BODY_HASH=%s", c.name, resp.StatusCode, acao, acac, resp.Header.Get("Location"), len(resp.Cookies()), resp.Header.Get("Cache-Control"), resp.Header.Get("Vary"), resp.Header.Get("ETag"), m12V2Hash(body))
		})
	}

	variations := []http.Header{
		{"X-Forwarded-Host":{"evil.example.test"}},
		{"Forwarded":{"host=evil.example.test"}},
		{"Origin":{"https://evil.example.test"}},
		{"Authorization":{"Bearer synthetic.invalid"}},
		{"Cookie":{"refresh=synthetic"}},
		{"Accept":{"text/html"}},
		{"Accept-Encoding":{"identity"}},
		{"RSC":{"1"},"Next-Router-State-Tree":{"synthetic"},"Next-Router-Prefetch":{"1"}},
	}
	for i, h := range variations {
		app := newHTTPAuthApp(t, newHTTPAuthService(t, &httpAuthTestStore{}))
		req := httptest.NewRequest(http.MethodGet, "/api/v1/portfolios?limit=10&case="+fmt.Sprint(i), nil)
		for k, vv := range h { for _, v := range vv { req.Header.Add(k,v) } }
		resp, err := app.Test(req)
		if err != nil { t.Fatal(err) }
		body, _ := io.ReadAll(resp.Body); resp.Body.Close()
		if !strings.Contains(strings.ToLower(resp.Header.Get("Cache-Control")), "no-store") {
			t.Errorf("CACHE_VARIATION=%d public-cache risk: %q", i, resp.Header.Get("Cache-Control"))
		}
		t.Logf("CACHE_VARIATION=%d STATUS=%d CACHE_CONTROL=%q VARY=%q ETAG=%q SET_COOKIE=%d LOCATION=%q BODY_HASH=%s", i, resp.StatusCode, resp.Header.Get("Cache-Control"), resp.Header.Get("Vary"), resp.Header.Get("ETag"), len(resp.Cookies()), resp.Header.Get("Location"), m12V2Hash(body))
	}
}
