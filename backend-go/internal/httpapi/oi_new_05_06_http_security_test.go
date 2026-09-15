package httpapi

import (
	"errors"
	"net"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/valyala/fasthttp"
)

func TestOINew0506HTTPTimeoutsUseActualAppConstruction(t *testing.T) {
	t.Parallel()

	apps := map[string]*fiber.App{
		"standard": newApp(&API{}),
		"replay":   newReplayApp(&API{}),
	}
	for name, app := range apps {
		app := app
		t.Run(name, func(t *testing.T) {
			config := app.Config()
			if config.ReadTimeout != 15*time.Second || config.WriteTimeout != 30*time.Second || config.IdleTimeout != 60*time.Second {
				t.Fatalf("unexpected HTTP timeouts: read=%s write=%s idle=%s", config.ReadTimeout, config.WriteTimeout, config.IdleTimeout)
			}
			if config.ReadTimeout == 0 || config.WriteTimeout == 0 || config.IdleTimeout == 0 {
				t.Fatalf("HTTP timeouts must all be bounded: read=%s write=%s idle=%s", config.ReadTimeout, config.WriteTimeout, config.IdleTimeout)
			}
			if config.TrustProxy || config.ProxyHeader != "" {
				t.Fatalf("default constructor must remain direct mode: trustProxy=%t proxyHeader=%q", config.TrustProxy, config.ProxyHeader)
			}
		})
	}
}

func TestOINew0506TrustedProxyFiberConfig(t *testing.T) {
	t.Parallel()

	networkConfig, err := NewHTTPNetworkConfig(true, []string{"127.0.0.1/32"})
	if err != nil {
		t.Fatalf("build trusted proxy config: %v", err)
	}
	app := newApp(&API{httpNetworkConfig: networkConfig})
	config := app.Config()
	if !config.TrustProxy {
		t.Fatal("trusted proxy mode must set TrustProxy")
	}
	if config.ProxyHeader != fiber.HeaderXForwardedFor {
		t.Fatalf("unexpected proxy header: %q", config.ProxyHeader)
	}
	if !config.EnableIPValidation {
		t.Fatal("trusted proxy mode must enable IP validation")
	}
	if len(config.TrustProxyConfig.Proxies) != 1 || config.TrustProxyConfig.Proxies[0] != "127.0.0.1/32" {
		t.Fatalf("unexpected trusted proxy allowlist: %#v", config.TrustProxyConfig.Proxies)
	}
	if config.TrustProxyConfig.Private || config.TrustProxyConfig.Loopback || config.TrustProxyConfig.LinkLocal {
		t.Fatalf("broad convenience trust must remain disabled: %#v", config.TrustProxyConfig)
	}
}

func TestOINew0506ProxyConfigurationFailsClosed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		trust bool
		list  []string
		ok    bool
	}{
		{name: "direct-empty", trust: false, list: nil, ok: true},
		{name: "trusted-empty", trust: true, list: nil},
		{name: "trusted-blank", trust: true, list: []string{"   "}},
		{name: "malformed-ip", trust: true, list: []string{"999.1.1.1"}},
		{name: "malformed-cidr", trust: true, list: []string{"10.0.0.0/99"}},
		{name: "ipv4-trust-all", trust: true, list: []string{"0.0.0.0/0"}},
		{name: "ipv6-trust-all", trust: true, list: []string{"::/0"}},
		{name: "ipv4-split-trust-all", trust: true, list: []string{"0.0.0.0/1", "128.0.0.0/1"}},
		{name: "ipv6-split-trust-all", trust: true, list: []string{"::/1", "8000::/1"}},
		{name: "mapped-ipv4-trust-all", trust: true, list: []string{"::ffff:0:0/96"}},
		{name: "explicit-proxies", trust: true, list: []string{"127.0.0.1/32", "10.10.1.0/24"}, ok: true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			_, err := NewHTTPNetworkConfig(test.trust, test.list)
			if test.ok && err != nil {
				t.Fatalf("expected valid config, got %v", err)
			}
			if !test.ok && err == nil {
				t.Fatal("expected configuration to be rejected")
			}
		})
	}
}

func TestOINew0506ClientIPCanonicalization(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"192.0.2.44":        "192.0.2.44",
		"::ffff:192.0.2.44": "192.0.2.44",
		"2001:db8:0:0::44":  "2001:db8::44",
	}
	for input, want := range tests {
		got, err := canonicalClientIP(input)
		if err != nil {
			t.Fatalf("canonicalize %q: %v", input, err)
		}
		if got != want {
			t.Fatalf("canonicalize %q: got %q want %q", input, got, want)
		}
	}
	if _, err := canonicalClientIP("not-an-ip"); err == nil {
		t.Fatal("invalid resolved client IP must fail closed")
	}
}

func TestOINew0506DirectModeForwardingHeadersCannotCreateFreshBuckets(t *testing.T) {
	t.Parallel()

	api, app := newOINew0506RateLimitApp(2, HTTPNetworkConfig{})
	peer := "198.51.100.20"

	for _, headers := range []map[string]string{
		{fiber.HeaderXForwardedFor: "1.2.3.4", "X-Real-IP": "5.6.7.8", "Forwarded": "for=9.10.11.12", "CF-Connecting-IP": "13.14.15.16"},
		{fiber.HeaderXForwardedFor: "17.18.19.20"},
	} {
		ip, err := oiNew0506RateLimitAttempt(app, api, peer, headers)
		if err != nil {
			t.Fatalf("direct-mode request unexpectedly rejected: %v", err)
		}
		if ip != peer {
			t.Fatalf("direct mode trusted a caller header: got %q want peer %q", ip, peer)
		}
	}

	ip, err := oiNew0506RateLimitAttempt(app, api, peer, map[string]string{fiber.HeaderXForwardedFor: "21.22.23.24"})
	if ip != peer {
		t.Fatalf("third request identity changed: got %q want %q", ip, peer)
	}
	if !errors.Is(err, errAuthRateLimited) {
		t.Fatalf("changing X-Forwarded-For created a fresh limiter bucket: %v", err)
	}
}

func TestOINew0506TrustedAndUntrustedProxyResolution(t *testing.T) {
	t.Parallel()

	networkConfig, err := NewHTTPNetworkConfig(true, []string{"127.0.0.1/32"})
	if err != nil {
		t.Fatalf("build trusted config: %v", err)
	}

	trustedAPI, trustedApp := newOINew0506RateLimitApp(5, networkConfig)
	ip, err := oiNew0506RateLimitAttempt(trustedApp, trustedAPI, "127.0.0.1", map[string]string{fiber.HeaderXForwardedFor: "203.0.113.10"})
	if err != nil {
		t.Fatalf("trusted proxy request: %v", err)
	}
	if ip != "203.0.113.10" {
		t.Fatalf("trusted proxy did not resolve client: got %q", ip)
	}

	untrustedAPI, untrustedApp := newOINew0506RateLimitApp(5, networkConfig)
	ip, err = oiNew0506RateLimitAttempt(untrustedApp, untrustedAPI, "127.0.0.2", map[string]string{fiber.HeaderXForwardedFor: "203.0.113.11"})
	if err != nil {
		t.Fatalf("untrusted proxy request: %v", err)
	}
	if ip != "127.0.0.2" {
		t.Fatalf("untrusted proxy header was not ignored: got %q", ip)
	}
}

func TestOINew0506TrustedProxyBucketsAndForwardedChain(t *testing.T) {
	t.Parallel()

	networkConfig, err := NewHTTPNetworkConfig(true, []string{"127.0.0.1/32", "10.0.0.2/32", "10.0.0.3/32"})
	if err != nil {
		t.Fatalf("build trusted config: %v", err)
	}
	api, app := newOINew0506RateLimitApp(2, networkConfig)

	for i := 0; i < 2; i++ {
		ip, err := oiNew0506RateLimitAttempt(app, api, "127.0.0.1", map[string]string{fiber.HeaderXForwardedFor: "203.0.113.20"})
		if err != nil || ip != "203.0.113.20" {
			t.Fatalf("client A attempt %d: ip=%q err=%v", i+1, ip, err)
		}
	}
	if _, err := oiNew0506RateLimitAttempt(app, api, "127.0.0.1", map[string]string{fiber.HeaderXForwardedFor: "203.0.113.20"}); !errors.Is(err, errAuthRateLimited) {
		t.Fatalf("same trusted client did not stay in one bucket: %v", err)
	}

	ip, err := oiNew0506RateLimitAttempt(app, api, "127.0.0.1", map[string]string{fiber.HeaderXForwardedFor: "203.0.113.21"})
	if err != nil || ip != "203.0.113.21" {
		t.Fatalf("client B must have an independent bucket: ip=%q err=%v", ip, err)
	}

	chainAPI, chainApp := newOINew0506RateLimitApp(5, networkConfig)
	ip, err = oiNew0506RateLimitAttempt(chainApp, chainAPI, "127.0.0.1", map[string]string{
		fiber.HeaderXForwardedFor: "198.51.100.66, 203.0.113.30, 10.0.0.3, 10.0.0.2",
	})
	if err != nil {
		t.Fatalf("multi-hop trusted proxy chain: %v", err)
	}
	if ip != "203.0.113.30" {
		t.Fatalf("forwarded chain accepted attacker-controlled left element: got %q want %q", ip, "203.0.113.30")
	}
}

func newOINew0506RateLimitApp(limit int, networkConfig HTTPNetworkConfig) (*API, *fiber.App) {
	now := time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)
	api := &API{
		authLimiter:       newBoundedAuthRateLimiter(limit, 1000, 1000, time.Minute),
		httpNetworkConfig: networkConfig,
		now:               func() time.Time { return now },
	}
	return api, newApp(api)
}

func oiNew0506RateLimitAttempt(app *fiber.App, api *API, peer string, headers map[string]string) (string, error) {
	var request fasthttp.Request
	request.Header.SetMethod(fiber.MethodPost)
	request.SetRequestURI("/api/v1/auth/login")
	for name, value := range headers {
		request.Header.Set(name, value)
	}

	remoteIP := net.ParseIP(peer)
	if remoteIP == nil {
		return "", errors.New("test peer is not an IP")
	}
	var requestCtx fasthttp.RequestCtx
	requestCtx.Init(&request, &net.TCPAddr{IP: remoteIP, Port: 4242}, nil)
	ctx := app.AcquireCtx(&requestCtx)
	defer app.ReleaseCtx(ctx)

	ip, err := normalizedClientIP(ctx)
	if err != nil {
		return "", err
	}
	return ip, api.checkAuthRateLimit(ctx)
}
