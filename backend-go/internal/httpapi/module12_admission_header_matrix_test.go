package httpapi

import (
	"errors"
	"fmt"
	"testing"

	"github.com/gofiber/fiber/v3"
)

// The trusted X-Forwarded-For identity must not be replaced by untrusted,
// conflicting forwarding-header families when the admission bucket is chosen.
func TestModule12TrustedProxyConflictingHeaderFamiliesShareAdmissionBucket(t *testing.T) {
	config, err := NewHTTPNetworkConfig(true, []string{"127.0.0.1/32"})
	if err != nil {
		t.Fatal(err)
	}
	api, app := newOINew0506RateLimitApp(2, config)
	for attempt := 0; attempt < 9; attempt++ {
		headers := map[string]string{
			fiber.HeaderXForwardedFor: "203.0.113.71",
			"Forwarded":               fmt.Sprintf("for=198.51.100.%d;proto=http", attempt+40),
			"CF-Connecting-IP":        fmt.Sprintf("192.0.2.%d", attempt+40),
			"X-Real-IP":               fmt.Sprintf("198.18.0.%d", attempt+40),
		}
		identity, admissionErr := oiNew0506RateLimitAttempt(app, api, "127.0.0.1", headers)
		if identity != "203.0.113.71" {
			t.Fatalf("attempt %d changed client identity to %q", attempt+1, identity)
		}
		if attempt < 2 && admissionErr != nil {
			t.Fatalf("attempt %d unexpectedly rejected: %v", attempt+1, admissionErr)
		}
		if attempt >= 2 && !errors.Is(admissionErr, errAuthRateLimited) {
			t.Fatalf("attempt %d rotated the admission bucket: %v", attempt+1, admissionErr)
		}
	}
}

// A peer absent from the explicit proxy allowlist cannot escape its admission
// limit by rotating either X-Forwarded-For or any alternate identity header.
func TestModule12UntrustedProxyConflictingHeaderFamiliesShareAdmissionBucket(t *testing.T) {
	config, err := NewHTTPNetworkConfig(true, []string{"127.0.0.1/32"})
	if err != nil {
		t.Fatal(err)
	}
	api, app := newOINew0506RateLimitApp(2, config)
	for attempt := 0; attempt < 9; attempt++ {
		headers := map[string]string{
			fiber.HeaderXForwardedFor: fmt.Sprintf("198.51.100.%d", attempt+40),
			"Forwarded":               fmt.Sprintf("for=192.0.2.%d", attempt+40),
			"CF-Connecting-IP":        fmt.Sprintf("203.0.113.%d", attempt+40),
			"X-Real-IP":               fmt.Sprintf("198.18.0.%d", attempt+40),
		}
		identity, admissionErr := oiNew0506RateLimitAttempt(app, api, "127.0.0.2", headers)
		if identity != "127.0.0.2" {
			t.Fatalf("attempt %d changed untrusted peer identity to %q", attempt+1, identity)
		}
		if attempt < 2 && admissionErr != nil {
			t.Fatalf("attempt %d unexpectedly rejected: %v", attempt+1, admissionErr)
		}
		if attempt >= 2 && !errors.Is(admissionErr, errAuthRateLimited) {
			t.Fatalf("attempt %d rotated the admission bucket: %v", attempt+1, admissionErr)
		}
	}
}
