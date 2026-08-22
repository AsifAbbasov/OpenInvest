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

func (limiter *argonWorkLimiter) run(work func()) error {
	select {
	case limiter.slots <- struct{}{}:
		defer func() { <-limiter.slots }()
		work()
		return nil
	default:
		return ErrAuthCapacity
	}
}

func (limiter *argonWorkLimiter) derive(password, salt []byte, timeCost, memoryKiB uint32, threads uint8, keyLen uint32) ([]byte, error) {
	var hash []byte
	if err := limiter.run(func() {
		hash = argon2.IDKey(password, salt, timeCost, memoryKiB, threads, keyLen)
	}); err != nil {
		return nil, err
	}
	return hash, nil
}

var processArgonWorkLimiter = newArgonWorkLimiter(argonMaxConcurrent)

func hashPassword(password string) (string, error) {
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash, err := processArgonWorkLimiter.derive([]byte(password), salt, argonTime, argonMemoryKiB, argonThreads, argonKeyLen)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argonMemoryKiB, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func verifyPassword(password string, encoded string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 5 || parts[0] != "argon2id" || parts[1] != "v=19" {
		return false, nil
	}
	params := map[string]uint32{}
	for _, pair := range strings.Split(parts[2], ",") {
		keyValue := strings.Split(pair, "=")
		if len(keyValue) != 2 {
			return false, nil
		}
		if keyValue[0] != "m" && keyValue[0] != "t" && keyValue[0] != "p" {
			return false, nil
		}
		if _, exists := params[keyValue[0]]; exists {
			return false, nil
		}
		value, err := strconv.ParseUint(keyValue[1], 10, 32)
		if err != nil {
			return false, nil
		}
		params[keyValue[0]] = uint32(value)
	}
	if len(params) != 3 || params["m"] != argonMemoryKiB || params["t"] != argonTime || params["p"] != uint32(argonThreads) {
		return false, nil
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil || len(salt) != argonSaltLen {
		return false, nil
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(expected) != argonKeyLen {
		return false, nil
	}
	actual, err := processArgonWorkLimiter.derive([]byte(password), salt, params["t"], params["m"], uint8(params["p"]), uint32(len(expected)))
	if err != nil {
		return false, err
	}
	return subtle.ConstantTimeCompare(actual, expected) == 1, nil
}

func verifyPasswordAgainstDummy(password string) error {
	salt := []byte("openinvest-auth")
	_, err := processArgonWorkLimiter.derive([]byte(password), salt, argonTime, argonMemoryKiB, argonThreads, argonKeyLen)
	return err
}
