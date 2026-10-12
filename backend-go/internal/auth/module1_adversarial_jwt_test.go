package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

var module1JWTSecret = []byte("01234567890123456789012345678901")

func module1JWTUser() StoredUser {
	return StoredUser{
		ID: "11111111-1111-4111-8111-111111111111",
		InvestmentSubjectID: "22222222-2222-4222-8222-222222222222",
		Email: "module1-jwt@example.test",
	}
}

func module1SignedRawJWT(headerJSON, payloadJSON string) string {
	h := base64.RawURLEncoding.EncodeToString([]byte(headerJSON))
	p := base64.RawURLEncoding.EncodeToString([]byte(payloadJSON))
	unsigned := h + "." + p
	mac := hmac.New(sha256.New, module1JWTSecret)
	_, _ = mac.Write([]byte(unsigned))
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func module1ValidJWT(t testing.TB) string {
	t.Helper()
	now := time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC)
	token, err := signAccessToken(module1JWTSecret, module1JWTUser(), now, now.Add(15*time.Minute))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return token
}

func TestModule1JWTStructuralMatrix(t *testing.T) {
	now := time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC)
	valid := module1ValidJWT(t)
	if _, err := verifyAccessToken(module1JWTSecret, valid, now); err != nil {
		t.Fatalf("negative-control baseline: valid token rejected: %v", err)
	}

	parts := strings.Split(valid, ".")
	cases := map[string]string{
		"zero_segments": "",
		"one_segment": parts[0],
		"two_segments": parts[0] + "." + parts[1],
		"four_segments": valid + ".extra",
		"empty_header": "." + parts[1] + "." + parts[2],
		"empty_payload": parts[0] + ".." + parts[2],
		"empty_signature": parts[0] + "." + parts[1] + ".",
		"invalid_header_base64": "*" + parts[0][1:] + "." + parts[1] + "." + parts[2],
		"invalid_payload_base64": parts[0] + ".*" + parts[1][1:] + "." + parts[2],
		"invalid_signature_base64": parts[0] + "." + parts[1] + ".*",
		"padded_signature": valid + "=",
		"truncated_signature": parts[0] + "." + parts[1] + "." + parts[2][:len(parts[2])-1],
		"extended_signature": parts[0] + "." + parts[1] + "." + parts[2] + "A",
	}
	for name, token := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := verifyAccessToken(module1JWTSecret, token, now); !errors.Is(err, errInvalidToken) {
				t.Fatalf("PROPERTY=M1-JWT-STRUCT MALFORMED_STRUCTURE_AUTHENTICATES case=%s err=%v", name, err)
			}
		})
	}
}

func TestModule1JWTRejectsAmbiguousDuplicateCriticalJSONKeys(t *testing.T) {
	now := time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC)
	exp := now.Add(15 * time.Minute).Unix()
	uid := module1JWTUser().ID
	sid := module1JWTUser().InvestmentSubjectID

	cases := []struct {
		name, header, payload string
	}{
		{
			name: "duplicate_alg_conflicting",
			header: `{"alg":"none","alg":"HS256","typ":"JWT"}`,
			payload: fmt.Sprintf(`{"sub":%q,"uid":%q,"investmentSubjectId":%q,"exp":%d,"iat":%d}`, uid, uid, sid, exp, now.Unix()),
		},
		{
			name: "duplicate_uid_conflicting",
			header: `{"alg":"HS256","typ":"JWT"}`,
			payload: fmt.Sprintf(`{"sub":%q,"uid":"33333333-3333-4333-8333-333333333333","uid":%q,"investmentSubjectId":%q,"exp":%d,"iat":%d}`, uid, uid, sid, exp, now.Unix()),
		},
		{
			name: "duplicate_exp_conflicting",
			header: `{"alg":"HS256","typ":"JWT"}`,
			payload: fmt.Sprintf(`{"sub":%q,"uid":%q,"investmentSubjectId":%q,"exp":1,"exp":%d,"iat":%d}`, uid, uid, sid, exp, now.Unix()),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			token := module1SignedRawJWT(tc.header, tc.payload)
			if _, err := verifyAccessToken(module1JWTSecret, token, now); err == nil {
				t.Fatalf("PROPERTY=M1-JWT-JSON AMBIGUOUS_DUPLICATE_SECURITY_CLAIM_ACCEPTED case=%s", tc.name)
			}
		})
	}
}

func TestModule1JWTCanonicalTextualEncodingRejectsByteEquivalentAlias(t *testing.T) {
	now := time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC)
	valid := module1ValidJWT(t)
	parts := strings.Split(valid, ".")
	decoded, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		t.Fatal(err)
	}
	alphabet := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	var alias string
	for _, candidate := range alphabet {
		if byte(candidate) == parts[2][len(parts[2])-1] {
			continue
		}
		changed := parts[2][:len(parts[2])-1] + string(candidate)
		got, decodeErr := base64.RawURLEncoding.DecodeString(changed)
		if decodeErr == nil && hmac.Equal(got, decoded) {
			alias = parts[0] + "." + parts[1] + "." + changed
			break
		}
	}
	if alias == "" {
		t.Skip("no byte-equivalent textual alias representable by current Go base64 decoder")
	}
	if _, err := verifyAccessToken(module1JWTSecret, alias, now); err == nil {
		t.Fatalf("PROPERTY=M1-JWT-CANON BYTE_EQUIVALENT_NONCANONICAL_ALIAS_ACCEPTED")
	}
}

func TestModule1JWTExpirationExactBoundaryUsesControlledClock(t *testing.T) {
	issued := time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC)
	exp := issued.Add(15 * time.Minute)
	token, err := signAccessToken(module1JWTSecret, module1JWTUser(), issued, exp)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := verifyAccessToken(module1JWTSecret, token, exp.Add(-time.Second)); err != nil {
		t.Fatalf("token must be valid immediately before exp: %v", err)
	}
	if _, err := verifyAccessToken(module1JWTSecret, token, exp); !errors.Is(err, errInvalidToken) {
		t.Fatalf("token must be invalid exactly at exp, got %v", err)
	}
	if _, err := verifyAccessToken(module1JWTSecret, token, exp.Add(time.Second)); !errors.Is(err, errInvalidToken) {
		t.Fatalf("token must remain invalid after exp, got %v", err)
	}
}

func FuzzModule1JWTMalformedStructureNeverAuthenticates(f *testing.F) {
	valid := module1ValidJWT(f)
	f.Add("extra")
	f.Add("")
	f.Add("A")
	now := time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC)
	f.Fuzz(func(t *testing.T, suffix string) {
		token := valid + "." + base64.RawURLEncoding.EncodeToString([]byte(suffix))
		if _, err := verifyAccessToken(module1JWTSecret, token, now); err == nil {
			t.Fatalf("PROPERTY=FZ-JWT-01 malformed four-segment token authenticated")
		}
	})
}

func FuzzModule1JWTByteChangingPayloadMutationNeverAuthenticates(f *testing.F) {
	valid := module1ValidJWT(f)
	parts := strings.Split(valid, ".")
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		f.Fatal(err)
	}
	f.Add(uint8(0), uint8(1))
	f.Add(uint8(len(payload)-1), uint8(0xff))
	now := time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC)
	f.Fuzz(func(t *testing.T, rawIndex uint8, mask uint8) {
		if len(payload) == 0 {
			t.Skip()
		}
		index := int(rawIndex) % len(payload)
		if mask == 0 {
			mask = 1
		}
		mutated := append([]byte(nil), payload...)
		mutated[index] ^= mask
		if string(mutated) == string(payload) {
			t.Skip()
		}
		token := parts[0] + "." + base64.RawURLEncoding.EncodeToString(mutated) + "." + parts[2]
		if _, err := verifyAccessToken(module1JWTSecret, token, now); err == nil {
			t.Fatalf("PROPERTY=FZ-JWT-02 byte-changing payload mutation authenticated index=%d mask=%d", index, mask)
		}
	})
}


func TestModule1JWTRejectsUnknownCriticalHeader(t *testing.T) {
	now := time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC)
	user := module1JWTUser()
	token := module1SignedRawJWT(
		`{"alg":"HS256","typ":"JWT","crit":["kid"],"kid":"synthetic-key-selector"}`,
		fmt.Sprintf(`{"sub":%q,"uid":%q,"investmentSubjectId":%q,"exp":%d,"iat":%d}`,
			user.ID, user.ID, user.InvestmentSubjectID, now.Add(15*time.Minute).Unix(), now.Unix()),
	)
	if _, err := verifyAccessToken(module1JWTSecret, token, now); err == nil {
		t.Fatal("PROPERTY=M1-JWT-CRIT UNKNOWN_CRITICAL_HEADER_ACCEPTED")
	}
}

func TestModule1EmailCanonicalizationPolicyMatrix(t *testing.T) {
	cases := []struct {
		name, input, expected string
	}{
		{"ascii_case", "Investor@Example.COM", "investor@example.com"},
		{"outer_whitespace", "  Investor@Example.COM  ", "investor@example.com"},
		{"plus_addressing_preserved", "Investor+Tag@Example.COM", "investor+tag@example.com"},
		{"dot_variant_preserved", "First.Last@Example.COM", "first.last@example.com"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeEmail(tc.input)
			if got != tc.expected {
				t.Fatalf("EMAIL_NORMALIZATION_POLICY case=%s expected=%q got=%q", tc.name, tc.expected, got)
			}
			if normalizeEmail(got) != got {
				t.Fatalf("normalization must be idempotent case=%s first=%q second=%q", tc.name, got, normalizeEmail(got))
			}
		})
	}
	if normalizeEmail("a+b@example.test") == normalizeEmail("ab@example.test") {
		t.Fatal("plus-addressing must not be silently collapsed without an explicit product policy")
	}
	if normalizeEmail("first.last@example.test") == normalizeEmail("firstlast@example.test") {
		t.Fatal("dot variants must not be silently collapsed without an explicit product policy")
	}
}

func FuzzModule1EmailNormalizationIdempotent(f *testing.F) {
	for _, seed := range []string{
		" Investor@Example.COM ",
		"a+b@example.test",
		"first.last@example.test",
		"δοκιμή@example.test",
		"",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input string) {
		once := normalizeEmail(input)
		twice := normalizeEmail(once)
		if once != twice {
			t.Fatalf("PROPERTY=FZ-REG-01 normalization not idempotent input=%q once=%q twice=%q", input, once, twice)
		}
		if strings.TrimSpace(once) != once {
			t.Fatalf("PROPERTY=FZ-REG-01 normalization retained outer whitespace input=%q normalized=%q", input, once)
		}
	})
}
