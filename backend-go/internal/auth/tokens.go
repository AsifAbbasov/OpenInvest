package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var errInvalidToken = errors.New("invalid token")

type accessClaims struct {
	SubjectID           string `json:"sub"`
	UserID              string `json:"uid"`
	InvestmentSubjectID string `json:"investmentSubjectId"`
	ExpiresAt           int64  `json:"exp"`
	IssuedAt            int64  `json:"iat"`
}

type accessHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
}

func randomToken(byteLength int) (string, error) {
	value := make([]byte, byteLength)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func tokenHash(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}

func signAccessToken(secret []byte, user StoredUser, issuedAt time.Time, expiresAt time.Time) (string, error) {
	header := accessHeader{Algorithm: "HS256", Type: "JWT"}
	claims := accessClaims{
		SubjectID:           user.ID,
		UserID:              user.ID,
		InvestmentSubjectID: user.InvestmentSubjectID,
		IssuedAt:            issuedAt.Unix(),
		ExpiresAt:           expiresAt.Unix(),
	}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	unsigned := base64.RawURLEncoding.EncodeToString(headerJSON) + "." + base64.RawURLEncoding.EncodeToString(claimsJSON)
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(unsigned))
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

func verifyAccessToken(secret []byte, token string, now time.Time) (accessClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return accessClaims{}, errInvalidToken
	}
	unsigned := parts[0] + "." + parts[1]
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return accessClaims{}, errInvalidToken
	}
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(unsigned))
	if !hmac.Equal(signature, mac.Sum(nil)) {
		return accessClaims{}, errInvalidToken
	}
	headerPayload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return accessClaims{}, errInvalidToken
	}
	var header accessHeader
	if err := json.Unmarshal(headerPayload, &header); err != nil {
		return accessClaims{}, errInvalidToken
	}
	if header.Algorithm != "HS256" || header.Type != "JWT" {
		return accessClaims{}, errInvalidToken
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return accessClaims{}, errInvalidToken
	}
	var claims accessClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return accessClaims{}, errInvalidToken
	}
	if claims.UserID == "" || claims.InvestmentSubjectID == "" || claims.ExpiresAt <= now.Unix() {
		return accessClaims{}, errInvalidToken
	}
	return claims, nil
}
