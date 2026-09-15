package httpapi

import (
	"fmt"
	"net/netip"
	"sort"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

const (
	defaultHTTPReadTimeout  = 15 * time.Second
	defaultHTTPWriteTimeout = 30 * time.Second
	defaultHTTPIdleTimeout  = 60 * time.Second
)

// HTTPNetworkConfig is the validated HTTP peer/proxy trust boundary used when building Fiber.
// Its fields are intentionally private so callers cannot construct an unvalidated trusted-proxy mode.
type HTTPNetworkConfig struct {
	trustProxy        bool
	trustedProxyCIDRs []string
}

// NewHTTPNetworkConfig validates and canonicalizes the explicit proxy allowlist. The zero value
// and trustProxy=false both represent direct mode; no forwarding header is trusted in that mode.
func NewHTTPNetworkConfig(trustProxy bool, trustedProxyCIDRs []string) (HTTPNetworkConfig, error) {
	normalized := make([]string, 0, len(trustedProxyCIDRs))
	prefixes := make([]netip.Prefix, 0, len(trustedProxyCIDRs))
	seen := make(map[string]struct{}, len(trustedProxyCIDRs))

	for _, raw := range trustedProxyCIDRs {
		entry := strings.TrimSpace(raw)
		if entry == "" {
			return HTTPNetworkConfig{}, fmt.Errorf("trusted proxy allowlist contains a blank entry")
		}
		canonical, prefix, err := canonicalTrustedProxyEntry(entry)
		if err != nil {
			return HTTPNetworkConfig{}, fmt.Errorf("invalid trusted proxy %q: %w", entry, err)
		}
		if _, exists := seen[canonical]; exists {
			continue
		}
		seen[canonical] = struct{}{}
		normalized = append(normalized, canonical)
		prefixes = append(prefixes, prefix)
	}

	if trustProxy && len(normalized) == 0 {
		return HTTPNetworkConfig{}, fmt.Errorf("trusted proxy mode requires OPENINVEST_TRUSTED_PROXY_CIDRS")
	}
	if proxyAllowlistTrustsEntireFamily(prefixes, true) {
		return HTTPNetworkConfig{}, fmt.Errorf("trusted proxy allowlist must not cover the entire IPv4 address space")
	}
	if proxyAllowlistTrustsEntireFamily(prefixes, false) {
		return HTTPNetworkConfig{}, fmt.Errorf("trusted proxy allowlist must not cover the entire IPv6 address space")
	}

	sort.Strings(normalized)
	if !trustProxy {
		return HTTPNetworkConfig{}, nil
	}
	return HTTPNetworkConfig{trustProxy: true, trustedProxyCIDRs: normalized}, nil
}

func canonicalTrustedProxyEntry(entry string) (string, netip.Prefix, error) {
	if strings.Contains(entry, "/") {
		prefix, err := netip.ParsePrefix(entry)
		if err != nil {
			return "", netip.Prefix{}, err
		}
		if prefix.Addr().Is4In6() {
			if prefix.Bits() < 96 {
				return "", netip.Prefix{}, fmt.Errorf("IPv4-mapped prefix is broader than the IPv4-mapped address space")
			}
			prefix = netip.PrefixFrom(prefix.Addr().Unmap(), prefix.Bits()-96)
		}
		prefix = prefix.Masked()
		return prefix.String(), prefix, nil
	}

	addr, err := netip.ParseAddr(entry)
	if err != nil {
		return "", netip.Prefix{}, err
	}
	addr = addr.Unmap()
	bits := 128
	if addr.Is4() {
		bits = 32
	}
	return addr.String(), netip.PrefixFrom(addr, bits), nil
}

// proxyAllowlistTrustsEntireFamily detects not only an explicit /0, but also an equivalent
// union such as two sibling /1 prefixes. This prevents a split allowlist from becoming trust-all.
func proxyAllowlistTrustsEntireFamily(prefixes []netip.Prefix, ipv4 bool) bool {
	current := make(map[netip.Prefix]struct{}, len(prefixes))
	for _, prefix := range prefixes {
		prefix = prefix.Masked()
		if prefix.Addr().Is4() != ipv4 {
			continue
		}
		current[prefix] = struct{}{}
	}

	for {
		for prefix := range current {
			if prefix.Bits() == 0 {
				return true
			}
		}

		siblings := make(map[netip.Prefix][]netip.Prefix)
		for prefix := range current {
			if prefix.Bits() == 0 {
				continue
			}
			parent := netip.PrefixFrom(prefix.Addr(), prefix.Bits()-1).Masked()
			siblings[parent] = append(siblings[parent], prefix)
		}

		merged := false
		for parent, children := range siblings {
			if len(children) != 2 {
				continue
			}
			delete(current, children[0])
			delete(current, children[1])
			current[parent] = struct{}{}
			merged = true
		}
		if !merged {
			return false
		}
	}
}

func newFiberApp(networkConfig HTTPNetworkConfig) *fiber.App {
	return fiber.New(fiberConfig(networkConfig))
}

func fiberConfig(networkConfig HTTPNetworkConfig) fiber.Config {
	config := fiber.Config{
		AppName:      "OpenInvest API",
		ReadTimeout:  defaultHTTPReadTimeout,
		WriteTimeout: defaultHTTPWriteTimeout,
		IdleTimeout:  defaultHTTPIdleTimeout,
	}
	if !networkConfig.trustProxy {
		return config
	}

	config.TrustProxy = true
	config.ProxyHeader = fiber.HeaderXForwardedFor
	config.EnableIPValidation = true
	config.TrustProxyConfig = fiber.TrustProxyConfig{
		Proxies: append([]string(nil), networkConfig.trustedProxyCIDRs...),
	}
	return config
}
