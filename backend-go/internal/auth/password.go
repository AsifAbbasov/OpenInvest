package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonMemoryKiB     = 64 * 1024
	argonTime          = 3
	argonThreads       = 1
	argonKeyLen        = 32
	argonSaltLen       = 16
	argonMaxConcurrent = 2
)

type argonWorkLimiter struct{ slots chan struct{} }

func newArgonWorkLimiter(maxConcurrent int) *argonWorkLimiter {
	if maxConcurrent < 1 {
		panic("argon max concurrency must be positive")
	}
	return &argonWorkLimiter{slots: make(chan struct{}, maxConcurrent)}
}

func (limiter *argonWorkLimiter) run(work func()) {
	limiter.slots <- struct{}{}
	defer func() { <-limiter.slots }()
	work()
}

func (limiter *argonWorkLimiter) derive(password, salt []byte, timeCost, memoryKiB uint32, threads uint8, keyLen uint32) []byte {
	var hash []byte
	limiter.run(func() {
		hash = argon2.IDKey(password, salt, timeCost, memoryKiB, threads, keyLen)
	})
	return hash
}

var processArgonWorkLimiter = newArgonWorkLimiter(argonMaxConcurrent)

func hashPassword(password string) (string, error) {
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := processArgonWorkLimiter.derive([]byte(password), salt, argonTime, argonMemoryKiB, argonThreads, argonKeyLen)
	return fmt.Sprintf("argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argonMemoryKiB, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func verifyPassword(password string, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 5 || parts[0] != "argon2id" || parts[1] != "v=19" {
		return false
	}
	params := map[string]uint32{}
	for _, pair := range strings.Split(parts[2], ",") {
		keyValue := strings.Split(pair, "=")
		if len(keyValue) != 2 {
			return false
		}
		if keyValue[0] != "m" && keyValue[0] != "t" && keyValue[0] != "p" {
			return false
		}
		if _, exists := params[keyValue[0]]; exists {
			return false
		}
		value, err := strconv.ParseUint(keyValue[1], 10, 32)
		if err != nil {
			return false
		}
		params[keyValue[0]] = uint32(value)
	}
	if len(params) != 3 || params["m"] != argonMemoryKiB || params["t"] != argonTime || params["p"] != uint32(argonThreads) {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil || len(salt) != argonSaltLen {
		return false
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(expected) != argonKeyLen {
		return false
	}
	actual := processArgonWorkLimiter.derive([]byte(password), salt, params["t"], params["m"], uint8(params["p"]), uint32(len(expected)))
	return subtle.ConstantTimeCompare(actual, expected) == 1
}

func verifyPasswordAgainstDummy(password string) {
	salt := []byte("openinvest-auth")
	_ = processArgonWorkLimiter.derive([]byte(password), salt, argonTime, argonMemoryKiB, argonThreads, argonKeyLen)
}
