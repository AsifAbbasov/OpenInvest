package httpapi

import (
	"errors"
	"fmt"
	"net/netip"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
)

var (
	errAuthRateLimited = errors.New("auth rate limited")
	errInvalidClientIP = errors.New("invalid resolved client IP")
)

type authRateLimiter struct {
	mu             sync.Mutex
	limit          int
	globalLimit    int
	maxKeys        int
	window         time.Duration
	attempts       map[string][]time.Time
	globalAttempts []time.Time
	lastSweep      time.Time
}

func newAuthRateLimiter(limit int, window time.Duration) *authRateLimiter {
	return newBoundedAuthRateLimiter(limit, defaultAuthRateLimiterGlobalLimit, defaultAuthRateLimiterMaxKeys, window)
}

func newBoundedAuthRateLimiter(limit int, globalLimit int, maxKeys int, window time.Duration) *authRateLimiter {
	return &authRateLimiter{
		limit:       nonNegativeInt(limit),
		globalLimit: nonNegativeInt(globalLimit),
		maxKeys:     nonNegativeInt(maxKeys),
		window:      window,
		attempts:    map[string][]time.Time{},
	}
}

func (limiter *authRateLimiter) allow(key string, now time.Time) bool {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	if limiter.window <= 0 || limiter.limit == 0 || limiter.globalLimit == 0 || limiter.maxKeys == 0 {
		return false
	}

	cutoff := now.Add(-limiter.window)
	limiter.globalAttempts = retainAuthAttempts(limiter.globalAttempts, cutoff)

	attempts, exists := limiter.attempts[key]
	if exists {
		attempts = retainAuthAttempts(attempts, cutoff)
		if len(attempts) == 0 {
			delete(limiter.attempts, key)
			exists = false
		} else {
			limiter.attempts[key] = attempts
		}
	}

	if !exists && (len(limiter.attempts) >= limiter.maxKeys || limiter.shouldSweep(now)) {
		limiter.sweepExpired(cutoff, now)
		attempts = limiter.attempts[key]
		exists = len(attempts) > 0
	}

	if len(limiter.globalAttempts) >= limiter.globalLimit {
		return false
	}
	if !exists && len(limiter.attempts) >= limiter.maxKeys {
		return false
	}
	if len(attempts) >= limiter.limit {
		return false
	}

	attempts = append(attempts, now)
	limiter.attempts[key] = attempts
	limiter.globalAttempts = append(limiter.globalAttempts, now)
	return true
}

func (limiter *authRateLimiter) shouldSweep(now time.Time) bool {
	return limiter.lastSweep.IsZero() || !now.Before(limiter.lastSweep.Add(limiter.window))
}

func (limiter *authRateLimiter) sweepExpired(cutoff time.Time, now time.Time) {
	for key, attempts := range limiter.attempts {
		retained := retainAuthAttempts(attempts, cutoff)
		if len(retained) == 0 {
			delete(limiter.attempts, key)
			continue
		}
		limiter.attempts[key] = retained
	}
	limiter.lastSweep = now
}

func retainAuthAttempts(attempts []time.Time, cutoff time.Time) []time.Time {
	retained := attempts[:0]
	for _, attempt := range attempts {
		if attempt.After(cutoff) {
			retained = append(retained, attempt)
		}
	}
	return retained
}

func nonNegativeInt(value int) int {
	if value < 0 {
		return 0
	}
	return value
}

func canonicalClientIP(raw string) (string, error) {
	addr, err := netip.ParseAddr(strings.TrimSpace(raw))
	if err != nil {
		return "", fmt.Errorf("%w: %q", errInvalidClientIP, raw)
	}
	return addr.Unmap().String(), nil
}

func normalizedClientIP(c fiber.Ctx) (string, error) {
	return canonicalClientIP(c.IP())
}

func (api *API) checkAuthRateLimit(c fiber.Ctx) error {
	if api.authLimiter == nil {
		return nil
	}
	clientIP, err := normalizedClientIP(c)
	if err != nil {
		return err
	}
	key := c.Path() + "|" + clientIP
	if !api.authLimiter.allow(key, api.nowUTC()) {
		return errAuthRateLimited
	}
	return nil
}
