package auth

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestVerifyAccessTokenRejectsNonCanonicalSignatureEncoding(t *testing.T) {
	secret := []byte("01234567890123456789012345678901")
	issuedAt := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
	user := StoredUser{ID: "user-id", InvestmentSubjectID: "subject-id"}
	token, err := signAccessToken(secret, user, issuedAt, issuedAt.Add(time.Minute))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("token parts: %q", token)
	}

	// HS256 emits 32 signature bytes. The final character of its 43-character
	// unpadded Base64URL encoding has TWO unused low bits, not four.
	alphabet := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	last := parts[2][len(parts[2])-1]
	position := strings.IndexByte(alphabet, last)
	if position < 0 {
		t.Fatalf("signature last character is outside Base64URL alphabet: %q", last)
	}
	mutatedLast := alphabet[(position&^3)|((position+1)&3)]
	mutatedSignature := parts[2][:len(parts[2])-1] + string(mutatedLast)
	originalBytes, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		t.Fatalf("decode original: %v", err)
	}
	alteredBytes, err := base64.RawURLEncoding.DecodeString(mutatedSignature)
	if err != nil {
		t.Fatalf("decode altered: %v", err)
	}
	if !bytes.Equal(originalBytes, alteredBytes) {
		t.Fatal("the alternate spelling must decode to the SAME signed bytes")
	}
	mutated := parts[0] + "." + parts[1] + "." + mutatedSignature
	if mutated == token {
		t.Fatal("alternate representation did not change the token")
	}
	if _, err := verifyAccessToken(secret, mutated, issuedAt); !errors.Is(err, errInvalidToken) {
		t.Fatalf("non-canonical signed token must be rejected, got %v", err)
	}
}

func TestVerifyAccessTokenAcceptsCanonicalAndPreservesIdentityExpiry(t *testing.T) {
	secret := []byte("01234567890123456789012345678901")
	issuedAt := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
	expiresAt := issuedAt.Add(time.Minute)
	user := StoredUser{ID: "user-id", InvestmentSubjectID: "subject-id"}
	token, err := signAccessToken(secret, user, issuedAt, expiresAt)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	claims, err := verifyAccessToken(secret, token, issuedAt)
	if err != nil {
		t.Fatalf("valid canonical token rejected: %v", err)
	}
	if claims.SubjectID != user.ID || claims.UserID != user.ID || claims.InvestmentSubjectID != user.InvestmentSubjectID {
		t.Fatalf("signed identities did not survive verification: %+v", claims)
	}
	if _, err := verifyAccessToken(secret, token, expiresAt); !errors.Is(err, errInvalidToken) {
		t.Fatalf("expired token was accepted: %v", err)
	}
}

func TestVerifyAccessTokenRejectsSignedSubjectUserMismatch(t *testing.T) {
	secret := []byte("01234567890123456789012345678901")
	issuedAt := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
	user := StoredUser{ID: "user-id", InvestmentSubjectID: "subject-id"}
	token, err := signAccessToken(secret, user, issuedAt, issuedAt.Add(time.Minute))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	parts := strings.Split(token, ".")
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatal(err)
	}
	var claims accessClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		t.Fatal(err)
	}
	claims.SubjectID = "different-user"
	payload, err = json.Marshal(claims)
	if err != nil {
		t.Fatal(err)
	}
	unsigned := parts[0] + "." + base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, secret)
	if _, err := mac.Write([]byte(unsigned)); err != nil {
		t.Fatal(err)
	}
	forgedWithValidMAC := unsigned + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if _, err := verifyAccessToken(secret, forgedWithValidMAC, issuedAt); !errors.Is(err, errInvalidToken) {
		t.Fatalf("valid-MAC token with mismatched sub/uid must be rejected: %v", err)
	}
}
