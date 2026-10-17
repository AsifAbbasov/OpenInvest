package auth

// Provenance: PR #229 at 657c34f813ac7faca1a9aa567b664dee5bc01be8; auth_theft_adversarial_test.go.
// Byte-identical signature aliases remain a separate, unresolved canonicalization contract.
import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"
	"time"
)

func FuzzAccessTokenByteMutation(f *testing.F) {
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
		mutated := string(raw)
		if _, err := verifyAccessToken(secret, mutated, now); err == nil {
			originalParts := strings.Split(valid, ".")
			mutatedParts := strings.Split(mutated, ".")
			if len(mutatedParts) != 3 {
				t.Fatalf("structurally mutated signed bearer authenticated: index=%d xor=%d", position, xor)
			}
			semanticChange := false
			for i := 0; i < 3; i++ {
				originalDecoded, originalErr := base64.RawURLEncoding.DecodeString(originalParts[i])
				mutatedDecoded, mutatedErr := base64.RawURLEncoding.DecodeString(mutatedParts[i])
				if originalErr != nil || mutatedErr != nil || !bytes.Equal(originalDecoded, mutatedDecoded) {
					semanticChange = true
					break
				}
			}
			if semanticChange {
				t.Fatalf("cryptographically distinct mutated signed bearer authenticated: index=%d xor=%d", position, xor)
			}
			// Different textual Base64URL spellings that decode to identical bytes are
			// a canonicalization concern, not a signature-forgery result. A dedicated
			// breaker test below treats acceptance of such aliases as hardening debt.
		}
	})
}
