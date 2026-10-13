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
	if err != nil { t.Fatalf("sign seed: %v", err) }
	parts := strings.Split(valid, ".")
	if len(parts) != 3 { t.Fatalf("unexpected seed structure") }

	alphabet := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	findAlias := func(encoded string) (string, bool) {
		decoded, err := base64.RawURLEncoding.DecodeString(encoded)
		if err != nil || len(encoded) == 0 { return "", false }
		last := len(encoded)-1
		for i:=0;i<len(alphabet);i++ {
			if alphabet[i]==encoded[last] { continue }
			candidate:=encoded[:last]+string(alphabet[i])
			candidateDecoded, decodeErr:=base64.RawURLEncoding.DecodeString(candidate)
			if decodeErr==nil && hmac.Equal(decoded,candidateDecoded) { return candidate,true }
		}
		return "",false
	}

	alias, ok := findAlias(parts[2])
	if !ok {
		t.Log("JWT_CANONICALIZATION=NO_BYTE_EQUIVALENT_SIGNATURE_ALIAS_FOUND_BY_GO_DECODER")
		return
	}
	canonicalSig, _ := base64.RawURLEncoding.DecodeString(parts[2])
	aliasSig, err := base64.RawURLEncoding.DecodeString(alias)
	if err != nil || !hmac.Equal(canonicalSig, aliasSig) {
		t.Fatalf("alias oracle invalid: decoded signature bytes differ")
	}
	aliased := parts[0]+"."+parts[1]+"."+alias
	claims, err := verifyAccessToken(secret, aliased, now)
	if err != nil {
		t.Fatalf("byte-equivalent signature alias unexpectedly rejected: %v", err)
	}
	if claims.InvestmentSubjectID != user.InvestmentSubjectID || claims.UserID != user.ID {
		t.Fatalf("byte-equivalent alias changed authenticated principal: claims=%+v", claims)
	}
	if base64.RawURLEncoding.EncodeToString(aliasSig) != parts[2] {
		t.Fatalf("canonical re-encoding mismatch")
	}
	t.Logf("JWT_CANONICALIZATION_POLICY=NOT_SPECIFIED BYTE_EQUIVALENT_ALIAS_ACCEPTED=YES PRINCIPAL_UNCHANGED=YES canonical=%q alias=%q", parts[2], alias)
}


func FuzzModule1SessionLifecycleStateMachine(f *testing.F) {
	f.Add([]byte{0, 1, 0, 2, 0})
	f.Add([]byte{0, 0, 1, 2, 1})
	f.Add([]byte{2, 0, 1})
	f.Fuzz(func(t *testing.T, commands []byte) {
		if len(commands) == 0 || len(commands) > 32 {
			return
		}
		store := &memoryStore{}
		store.user = StoredUser{
			ID: "user-a",
			InvestmentSubjectID: "subject-a",
			Email: "state-machine@example.test",
			Language: LanguageEN,
			Theme: ThemeSystem,
			Timezone: "UTC",
			PrivacyMode: true,
		}
		rootRefresh := "root-refresh-synthetic"
		rootCSRF := "root-csrf-synthetic"
		rootRecord := SessionRecord{
			SessionID: "root-session",
			UserID: store.user.ID,
			RefreshTokenHash: tokenHash(rootRefresh),
			CSRFTokenHash: tokenHash(rootCSRF),
			ExpiresAt: time.Date(2026, 11, 13, 0, 0, 0, 0, time.UTC),
			Now: time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC),
		}
		store.sessions = map[string]SessionRecord{rootRecord.RefreshTokenHash: rootRecord}
		service, err := NewService(store, fixedClock{}, Config{
			AccessTokenSecret: []byte("01234567890123456789012345678901"),
			RefreshCookieSecure: true,
		})
		if err != nil {
			t.Fatalf("new state-machine service: %v", err)
		}

		currentRefresh, currentCSRF := rootRefresh, rootCSRF
		active := true
		type pair struct{ refresh, csrf string }
		old := []pair{}

		for step, raw := range commands {
			switch raw % 4 {
			case 0: // rotate current session
				result, err := service.Refresh(t.Context(), currentRefresh, currentCSRF)
				if !active {
					if err == nil {
						t.Fatalf("step=%d MODEL_DENY_IMPLEMENTATION_ALLOW refresh after logout", step)
					}
					continue
				}
				if err != nil {
					t.Fatalf("step=%d model allowed refresh but implementation denied: %v", step, err)
				}
				old = append(old, pair{currentRefresh, currentCSRF})
				currentRefresh, currentCSRF = result.RefreshToken, result.Session.CSRFToken
			case 1: // replay an obsolete credential if one exists
				if len(old) == 0 {
					continue
				}
				p := old[int(raw)%len(old)]
				if _, err := service.Refresh(t.Context(), p.refresh, p.csrf); err == nil {
					t.Fatalf("step=%d obsolete refresh credential replay authenticated", step)
				}
			case 2: // revoke the current session
				revoked, err := service.Logout(t.Context(), currentRefresh, currentCSRF, false)
				if !active {
					if err == nil || revoked {
						t.Fatalf("step=%d already-revoked session was revoked again as valid", step)
					}
					continue
				}
				if err != nil || !revoked {
					t.Fatalf("step=%d model allowed logout but implementation result revoked=%t err=%v", step, revoked, err)
				}
				old = append(old, pair{currentRefresh, currentCSRF})
				active = false
			case 3: // CSRF mismatch must never rotate
				beforeCount := len(store.sessions)
				if _, err := service.Refresh(t.Context(), currentRefresh, "wrong-csrf-synthetic"); err == nil {
					t.Fatalf("step=%d CSRF mismatch authenticated", step)
				}
				if len(store.sessions) != beforeCount {
					t.Fatalf("step=%d CSRF mismatch mutated session count before=%d after=%d", step, beforeCount, len(store.sessions))
				}
			}
		}

		for i, p := range old {
			if _, err := service.Refresh(t.Context(), p.refresh, p.csrf); err == nil {
				t.Fatalf("postcondition obsolete credential %d resurrected", i)
			}
		}
	})
}


func TestModule1JWTExtendedStructuralBoundaries(t *testing.T) {
	requireModule1Adversarial(t)
	secret := []byte("01234567890123456789012345678901")
	now := time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC)
	user := StoredUser{ID: "user-a", InvestmentSubjectID: "subject-a"}
	valid, err := signAccessToken(secret, user, now.Add(-time.Minute), now.Add(time.Hour))
	if err != nil { t.Fatal(err) }
	parts := strings.Split(valid, ".")
	if len(parts) != 3 { t.Fatal("unexpected JWT seed") }

	cases := map[string]string{
		"five_segments": valid + ".x.y",
		"padded_header": parts[0] + "=" + "." + parts[1] + "." + parts[2],
		"padded_payload": parts[0] + "." + parts[1] + "=" + "." + parts[2],
		"oversized_header": strings.Repeat("A", 8192) + "." + parts[1] + "." + parts[2],
		"oversized_payload": parts[0] + "." + strings.Repeat("A", 8192) + "." + parts[2],
		"oversized_signature": parts[0] + "." + parts[1] + "." + strings.Repeat("A", 8192),
		"signature_one_byte": parts[0] + "." + parts[1] + ".AA",
	}
	for name, token := range cases {
		t.Run(name, func(t *testing.T) {
			if claims, err := verifyAccessToken(secret, token, now); err == nil {
				t.Fatalf("invalid structural boundary authenticated claims=%+v", claims)
			}
		})
	}
}

func TestModule1JWTIssuedAtBehaviorRecorded(t *testing.T) {
	requireModule1Adversarial(t)
	secret := []byte("01234567890123456789012345678901")
	now := time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC)
	user := StoredUser{ID: "user-a", InvestmentSubjectID: "subject-a"}
	futureIssued, err := signAccessToken(secret, user, now.Add(24*time.Hour), now.Add(25*time.Hour))
	if err != nil { t.Fatal(err) }
	claims, err := verifyAccessToken(secret, futureIssued, now)
	if err != nil {
		t.Logf("JWT_IAT_FUTURE_POLICY=REJECTED err=%v", err)
		return
	}
	if claims.InvestmentSubjectID != user.InvestmentSubjectID {
		t.Fatalf("future iat changed principal: %+v", claims)
	}
	t.Log("JWT_IAT_FUTURE_POLICY=ACCEPTED_NOT_CURRENTLY_VALIDATED; NO_ATTACKER_SIGNATURE_PATH_DEMONSTRATED")
}
