package httpapi

import (
	"testing"
	"time"
)

func TestAuthLayeredAdmissionReducesSharedNATCouplingWithoutEmailRotationBypass(t *testing.T) {
	now := time.Date(2026, 10, 6, 9, 0, 0, 0, time.UTC)
	ipLimiter := newBoundedAuthRateLimiter(10, 1000, 1000, time.Minute)
	credentialLimiter := newBoundedAuthRateLimiter(2, 1000, 1000, time.Minute)
	ipKey := "/api/v1/auth/login|203.0.113.10"

	// Five independent credentials can each use two attempts behind one NAT before
	// the emergency IP ceiling is reached. The old single 2-attempt IP bucket
	// would have coupled them immediately.
	for user := 0; user < 5; user++ {
		email := "user-" + string(rune('a'+user)) + "@example.com"
		credentialKey := "/api/v1/auth/login|credential|" + normalizedAuthCredentialKey(email)
		for attempt := 0; attempt < 2; attempt++ {
			if !ipLimiter.allow(ipKey, now) {
				t.Fatalf("shared NAT emergency ceiling rejected user=%d attempt=%d too early", user, attempt)
			}
			if !credentialLimiter.allow(credentialKey, now) {
				t.Fatalf("credential bucket rejected user=%d attempt=%d too early", user, attempt)
			}
		}
	}
	if ipLimiter.allow(ipKey, now) {
		t.Fatal("rotating credentials bypassed the emergency IP ceiling")
	}
}

func TestAuthCredentialBucketCanonicalizesEmailAcrossIPAndCaseVariants(t *testing.T) {
	now := time.Date(2026, 10, 6, 9, 0, 0, 0, time.UTC)
	limiter := newBoundedAuthRateLimiter(2, 1000, 1000, time.Minute)
	left := "/api/v1/auth/login|credential|" + normalizedAuthCredentialKey(" Investor@Example.COM ")
	right := "/api/v1/auth/login|credential|" + normalizedAuthCredentialKey("investor@example.com")
	if left != right {
		t.Fatalf("email case/space variants produced distinct credential keys: %q %q", left, right)
	}
	if !limiter.allow(left, now) || !limiter.allow(right, now) {
		t.Fatal("credential bucket rejected inside limit")
	}
	if limiter.allow(left, now) {
		t.Fatal("same normalized credential bypassed per-credential budget")
	}
}
