package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func requireModule1Adversarial(t *testing.T) {
	t.Helper()
	if os.Getenv("OPENINVEST_MODULE1_ADVERSARIAL") != "1" {
		t.Skip("OPENINVEST_MODULE1_ADVERSARIAL is not enabled")
	}
}

func TestModule1JWTStructuralSignatureAndTimeMatrix(t *testing.T) {
	requireModule1Adversarial(t)
	secret := []byte("01234567890123456789012345678901")
	now := time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC)
	user := StoredUser{ID: "user-a", InvestmentSubjectID: "subject-a"}

	valid, err := signAccessToken(secret, user, now.Add(-time.Minute), now.Add(time.Minute))
	if err != nil {
		t.Fatalf("sign valid token: %v", err)
	}
	if got, err := verifyAccessToken(secret, valid, now); err != nil || got.InvestmentSubjectID != "subject-a" {
		t.Fatalf("valid token rejected or principal changed: claims=%+v err=%v", got, err)
	}

	parts := strings.Split(valid, ".")
	if len(parts) != 3 {
		t.Fatalf("seed token structure: %q", valid)
	}

	structural := map[string]string{
		"zero_segments": "",
		"one_segment": "abc",
		"two_segments": "abc.def",
		"four_segments": valid + ".extra",
		"empty_header": "." + parts[1] + "." + parts[2],
		"empty_payload": parts[0] + ".." + parts[2],
		"empty_signature": parts[0] + "." + parts[1] + ".",
		"invalid_header_b64": "*" + parts[0][1:] + "." + parts[1] + "." + parts[2],
		"invalid_payload_b64": parts[0] + ".*" + parts[1][1:] + "." + parts[2],
		"invalid_signature_b64": parts[0] + "." + parts[1] + ".*",
		"padded_signature": parts[0] + "." + parts[1] + "." + parts[2] + "=",
	}
	for name, token := range structural {
		t.Run(name, func(t *testing.T) {
			if claims, err := verifyAccessToken(secret, token, now); err == nil {
				t.Fatalf("malformed token authenticated: claims=%+v token=%q", claims, token)
			}
		})
	}

	for _, index := range []int{0, 1, 2} {
		t.Run(fmt.Sprintf("tamper_segment_%d", index), func(t *testing.T) {
			mutated := append([]string(nil), parts...)
			if len(mutated[index]) == 0 {
				t.Fatal("unexpected empty segment")
			}
			replacement := byte('A')
			if mutated[index][0] == replacement {
				replacement = 'B'
			}
			mutated[index] = string(replacement) + mutated[index][1:]
			if claims, err := verifyAccessToken(secret, strings.Join(mutated, "."), now); err == nil {
				t.Fatalf("byte-changing JWT mutation authenticated: claims=%+v", claims)
			}
		})
	}

	expired, err := signAccessToken(secret, user, now.Add(-2*time.Minute), now)
	if err != nil {
		t.Fatalf("sign exp boundary token: %v", err)
	}
	if claims, err := verifyAccessToken(secret, expired, now); err == nil {
		t.Fatalf("token with exp==now authenticated: %+v", claims)
	}

	notExpired, err := signAccessToken(secret, user, now.Add(-time.Minute), now.Add(time.Second))
	if err != nil {
		t.Fatalf("sign future-exp token: %v", err)
	}
	if claims, err := verifyAccessToken(secret, notExpired, now); err != nil || claims.InvestmentSubjectID != user.InvestmentSubjectID {
		t.Fatalf("token immediately before expiry rejected: claims=%+v err=%v", claims, err)
	}
}

func TestModule1JWTSignedDuplicateSecurityClaimBehaviorIsRecorded(t *testing.T) {
	requireModule1Adversarial(t)
	secret := []byte("01234567890123456789012345678901")
	now := time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC)
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"user-a","uid":"user-a","investmentSubjectId":"subject-a","investmentSubjectId":"subject-b","iat":1791849540,"exp":1893456000}`))
	unsigned := header + "." + payload
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(unsigned))
	token := unsigned + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	claims, err := verifyAccessToken(secret, token, now)
	if err != nil {
		t.Logf("DUPLICATE_CLAIM_POLICY=REJECTED err=%v", err)
		return
	}
	t.Logf("DUPLICATE_CLAIM_POLICY=ACCEPTED RESOLVED_INVESTMENT_SUBJECT=%q", claims.InvestmentSubjectID)
	if claims.InvestmentSubjectID != "subject-b" {
		t.Fatalf("unexpected duplicate-claim resolution: %+v", claims)
	}
	// This is evidence about parser behavior, not by itself a vulnerability: constructing
	// a valid signed token requires the signing secret. The security oracle remains that an
	// unsigned/byte-changing attacker mutation must never authenticate.
}

func FuzzModule1JWTByteChangingMutationNeverAuthenticates(f *testing.F) {
	secret := []byte("01234567890123456789012345678901")
	now := time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC)
	user := StoredUser{ID: "user-a", InvestmentSubjectID: "subject-a"}
	valid, err := signAccessToken(secret, user, now.Add(-time.Minute), now.Add(time.Hour))
	if err != nil {
		f.Fatalf("sign seed: %v", err)
	}
	f.Add([]byte(valid), uint16(0))
	f.Add([]byte(valid), uint16(len(valid)/2))
	f.Add([]byte(valid), uint16(len(valid)-1))

	f.Fuzz(func(t *testing.T, raw []byte, pos uint16) {
		if len(raw) == 0 || len(raw) > 4096 {
			return
		}
		seed := []byte(valid)
		index := int(pos) % len(seed)
		mutated := append([]byte(nil), seed...)
		mutated[index] ^= 0x01
		if string(mutated) == valid {
			t.Fatal("mutation oracle failed to change token bytes")
		}
		if claims, err := verifyAccessToken(secret, string(mutated), now); err == nil {
			t.Fatalf("byte-changing token mutation authenticated at index=%d claims=%+v", index, claims)
		}
	})
}

func FuzzModule1JWTStructuralParserNeverPanics(f *testing.F) {
	secret := []byte("01234567890123456789012345678901")
	now := time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC)
	for _, seed := range []string{"", ".", "..", "a.b.c", strings.Repeat("A", 512)} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, token string) {
		if len(token) > 8192 {
			return
		}
		_, _ = verifyAccessToken(secret, token, now)
	})
}


func TestModule1JWTCanonicalTextualEncoding(t *testing.T) {
	requireModule1Adversarial(t)
	secret := []byte("01234567890123456789012345678901")
	now := time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC)
	user := StoredUser{ID: "user-a", InvestmentSubjectID: "subject-a"}
	valid, err := signAccessToken(secret, user, now.Add(-time.Minute), now.Add(time.Hour))
	if err != nil {
		t.Fatalf("sign seed: %v", err)
	}
	parts := strings.Split(valid, ".")
	if len(parts) != 3 {
		t.Fatalf("unexpected seed structure")
	}

	alphabet := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	findAlias := func(encoded string) (string, bool) {
		decoded, err := base64.RawURLEncoding.DecodeString(encoded)
		if err != nil || len(encoded) == 0 {
			return "", false
		}
		last := len(encoded) - 1
		for i := 0; i < len(alphabet); i++ {
			if alphabet[i] == encoded[last] {
				continue
			}
			candidate := encoded[:last] + string(alphabet[i])
			candidateDecoded, decodeErr := base64.RawURLEncoding.DecodeString(candidate)
			if decodeErr == nil && hmac.Equal(decoded, candidateDecoded) {
				return candidate, true
			}
		}
		return "", false
	}

	alias, ok := findAlias(parts[2])
	if !ok {
		t.Log("JWT_CANONICALIZATION=NO_BYTE_EQUIVALENT_SIGNATURE_ALIAS_FOUND_BY_GO_DECODER")
		return
	}
	aliased := parts[0] + "." + parts[1] + "." + alias
	if claims, err := verifyAccessToken(secret, aliased, now); err == nil {
		t.Fatalf("noncanonical byte-equivalent signature encoding authenticated: subject=%q canonical=%q alias=%q", claims.InvestmentSubjectID, parts[2], alias)
	}
	t.Log("JWT_CANONICALIZATION=NONCANONICAL_SIGNATURE_ALIAS_REJECTED")
}
