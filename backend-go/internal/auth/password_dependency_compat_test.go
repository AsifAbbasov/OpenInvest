package auth

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestStage344HistoricalArgon2idHashRemainsCompatible(t *testing.T) {
	const (
		password = "Stage344 historical password!"
		historical = "argon2id$v=19$m=65536,t=3,p=1$AAECAwQFBgcICQoLDA0ODw$9Wc7VFOLWWIqyGrC4T85CfZU8wFkNPcB2GE1ympJUGU"
	)

	verified, err := verifyPassword(password, historical)
	if err != nil {
		t.Fatalf("verify historical pre-update hash: %v", err)
	}
	if !verified {
		t.Fatal("historical pre-update Argon2id hash no longer verifies")
	}

	wrongVerified, err := verifyPassword(password+"-wrong", historical)
	if err != nil {
		t.Fatalf("verify wrong password against historical hash: %v", err)
	}
	if wrongVerified {
		t.Fatal("wrong password unexpectedly verifies against historical hash")
	}

	if argonMemoryKiB != 64*1024 {
		t.Fatalf("argonMemoryKiB changed: got %d want %d", argonMemoryKiB, 64*1024)
	}
	if argonTime != 3 {
		t.Fatalf("argonTime changed: got %d want 3", argonTime)
	}
	if argonThreads != 1 {
		t.Fatalf("argonThreads changed: got %d want 1", argonThreads)
	}
	if argonKeyLen != 32 {
		t.Fatalf("argonKeyLen changed: got %d want 32", argonKeyLen)
	}
	if argonSaltLen != 16 {
		t.Fatalf("argonSaltLen changed: got %d want 16", argonSaltLen)
	}

	encoded, err := hashPassword(password)
	if err != nil {
		t.Fatalf("hash password using current implementation: %v", err)
	}
	const prefix = "argon2id$v=19$m=65536,t=3,p=1$"
	if !strings.HasPrefix(encoded, prefix) {
		t.Fatalf("encoded Argon2id format/parameters changed: %q", encoded)
	}

	parts := strings.Split(encoded, "$")
	if len(parts) != 5 {
		t.Fatalf("encoded Argon2id field count changed: got %d", len(parts))
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		t.Fatalf("decode generated salt: %v", err)
	}
	if len(salt) != argonSaltLen {
		t.Fatalf("generated salt length changed: got %d want %d", len(salt), argonSaltLen)
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		t.Fatalf("decode generated hash: %v", err)
	}
	if len(hash) != argonKeyLen {
		t.Fatalf("generated hash length changed: got %d want %d", len(hash), argonKeyLen)
	}

	newVerified, err := verifyPassword(password, encoded)
	if err != nil {
		t.Fatalf("verify newly encoded password: %v", err)
	}
	if !newVerified {
		t.Fatal("newly encoded password does not round-trip")
	}
}
