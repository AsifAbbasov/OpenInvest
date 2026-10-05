package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

type authTheftClock struct{ now time.Time }

func (clock *authTheftClock) Now() time.Time { return clock.now }

func TestAuthTheftAccessTokenTamperingAndAlgorithmConfusionRejected(t *testing.T) {
	secret := []byte("01234567890123456789012345678901")
	now := time.Date(2026, 10, 5, 10, 0, 0, 0, time.UTC)
	user := StoredUser{
		ID:                  "00000000-0000-4000-8000-000000000111",
		InvestmentSubjectID: "00000000-0000-4000-8000-000000000222",
	}
	valid, err := signAccessToken(secret, user, now, now.Add(15*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := verifyAccessToken(secret, valid, now); err != nil {
		t.Fatalf("valid access token rejected: %v", err)
	}

	parts := strings.Split(valid, ".")
	if len(parts) != 3 {
		t.Fatalf("unexpected valid token shape")
	}

	for _, tc := range []struct {
		name  string
		token string
	}{
		{"payload byte changed", parts[0] + "." + mutateTokenPart(parts[1]) + "." + parts[2]},
		{"signature byte changed", parts[0] + "." + parts[1] + "." + mutateTokenPart(parts[2])},
		{"header byte changed", mutateTokenPart(parts[0]) + "." + parts[1] + "." + parts[2]},
		{"truncated token", parts[0] + "." + parts[1]},
		{"extra segment", valid + ".extra"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := verifyAccessToken(secret, tc.token, now); err == nil {
				t.Fatal("tampered bearer token was accepted")
			}
		})
	}

	for _, alg := range []string{"none", "HS512", "RS256", ""} {
		t.Run("algorithm "+alg, func(t *testing.T) {
			header, err := json.Marshal(accessHeader{Algorithm: alg, Type: "JWT"})
			if err != nil {
				t.Fatal(err)
			}
			unsigned := base64.RawURLEncoding.EncodeToString(header) + "." + parts[1]
			mac := hmac.New(sha256.New, secret)
			_, _ = mac.Write([]byte(unsigned))
			signedWrongAlgorithm := unsigned + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
			if _, err := verifyAccessToken(secret, signedWrongAlgorithm, now); err == nil {
				t.Fatalf("token with algorithm %q was accepted despite valid HMAC", alg)
			}
		})
	}
}

func TestAuthTheftAccessTokenExpiryAndLogoutExposureWindowIsBounded(t *testing.T) {
	clock := &authTheftClock{now: time.Date(2026, 10, 5, 10, 0, 0, 0, time.UTC)}
	store := &memoryStore{}
	service, err := NewService(store, clock, Config{
		AccessTokenSecret:   []byte("01234567890123456789012345678901"),
		AccessTokenTTL:      15 * time.Minute,
		RefreshTokenTTL:     30 * 24 * time.Hour,
		RefreshCookieSecure: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := service.Register(t.Context(), RegistrationRequest{
		Email:    "access-theft@example.com",
		Password: "correct horse battery staple",
		Language: LanguageEN,
		Theme:    ThemeSystem,
		Timezone: "UTC",
	})
	if err != nil {
		t.Fatal(err)
	}
	if ttl := result.Session.AccessTokenExpiresAt.Sub(clock.now); ttl != 15*time.Minute {
		t.Fatalf("access token exposure window drifted: ttl=%s", ttl)
	}
	subjectID, err := service.AuthenticateAccessToken(result.Session.AccessToken)
	if err != nil || subjectID != result.User.InvestmentSubjectID {
		t.Fatalf("fresh access token authentication failed: subject=%q err=%v", subjectID, err)
	}

	revoked, err := service.Logout(t.Context(), result.RefreshToken, result.Session.CSRFToken, false)
	if err != nil || !revoked {
		t.Fatalf("logout failed: revoked=%t err=%v", revoked, err)
	}

	// Access tokens are intentionally stateless. Logout revokes refresh capability,
	// while a stolen bearer remains usable only until its short expiry. Make that
	// residual window explicit and fail if it silently grows.
	if _, err := service.AuthenticateAccessToken(result.Session.AccessToken); err != nil {
		t.Fatalf("short-lived access token unexpectedly became stateful after logout: %v", err)
	}

	clock.now = result.Session.AccessTokenExpiresAt
	if _, err := service.AuthenticateAccessToken(result.Session.AccessToken); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("access token remained valid at expiry boundary: %v", err)
	}
}

func FuzzAuthTheftMutatedSignedBearerNeverAuthenticates(f *testing.F) {
	f.Add(uint16(0), byte(1))
	f.Add(uint16(17), byte(0xff))
	f.Add(uint16(250), byte(0x80))

	secret := []byte("01234567890123456789012345678901")
	now := time.Date(2026, 10, 5, 10, 0, 0, 0, time.UTC)
	user := StoredUser{
		ID:                  "00000000-0000-4000-8000-000000000111",
		InvestmentSubjectID: "00000000-0000-4000-8000-000000000222",
	}
	valid, err := signAccessToken(secret, user, now, now.Add(15*time.Minute))
	if err != nil {
		f.Fatal(err)
	}

	f.Fuzz(func(t *testing.T, index uint16, xor byte) {
		raw := []byte(valid)
		position := int(index) % len(raw)
		if xor == 0 {
			xor = 1
		}
		raw[position] ^= xor
		if string(raw) == valid {
			t.Fatal("mutation failed to change token")
		}
		if _, err := verifyAccessToken(secret, string(raw), now); err == nil {
			t.Fatalf("mutated signed bearer authenticated: index=%d xor=%d", position, xor)
		}
	})
}

func mutateTokenPart(value string) string {
	if value == "" {
		return "A"
	}
	raw := []byte(value)
	if raw[len(raw)-1] == 'A' {
		raw[len(raw)-1] = 'B'
	} else {
		raw[len(raw)-1] = 'A'
	}
	return string(raw)
}
