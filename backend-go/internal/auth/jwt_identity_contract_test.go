package auth

// Independent JWT-01 functions and helpers retain the original assertions. Full
// original sources are preserved verbatim in testdata/jwt01. PR261 blob dfe275965e8702b1d2abec96ff5579373875a57b;
// PR265 blob 21eac313f4b3d05cb55472ff627c83ecc66408e5.
import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"
)

func campaignSignedJWT(t *testing.T, secret []byte, header map[string]any, claims map[string]any) string {
	t.Helper()
	h, e := json.Marshal(header)
	if e != nil {
		t.Fatal(e)
	}
	c, e := json.Marshal(claims)
	if e != nil {
		t.Fatal(e)
	}
	body := base64.RawURLEncoding.EncodeToString(h) + "." + base64.RawURLEncoding.EncodeToString(c)
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(body))
	return body + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
func TestModule1CampaignJWTRequiresCanonicalSignedSubject(t *testing.T) {
	secret := []byte("independent-module1-jwt-testing-key-32-bytes")
	now := time.Date(2026, 10, 14, 15, 0, 0, 0, time.UTC)
	header := map[string]any{"alg": "HS256", "typ": "JWT"}
	claims := map[string]any{"uid": "user-A", "sub": "user-B", "investmentSubjectId": "subject-A", "iat": now.Unix(), "exp": now.Add(time.Hour).Unix()}
	token := campaignSignedJWT(t, secret, header, claims)
	if _, e := verifyAccessToken(secret, token, now); e == nil {
		t.Fatal("SIGNED_CLAIM_INCONSISTENCY: verifier accepted sub != uid; requires independent assessment of exploitability")
	}
}
func gapJWT(secret []byte, header, payload string) string {
	a := base64.RawURLEncoding.EncodeToString([]byte(header))
	b := base64.RawURLEncoding.EncodeToString([]byte(payload))
	unsigned := a + "." + b
	m := hmac.New(sha256.New, secret)
	_, _ = m.Write([]byte(unsigned))
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}
func TestAUTHGAPJWTCanonicalSignedIdentityContract(t *testing.T) {
	secret := []byte("audit-independent-hmac-secret-32-bytes")
	now := time.Unix(1800000000, 0).UTC()
	user := StoredUser{ID: "user-A", InvestmentSubjectID: "subject-A"}
	legitimate, e := signAccessToken(secret, user, now, now.Add(time.Hour))
	if e != nil {
		t.Fatal(e)
	}
	claims, e := verifyAccessToken(secret, legitimate, now)
	if e != nil || claims.UserID != user.ID || claims.SubjectID != user.ID || claims.InvestmentSubjectID != user.InvestmentSubjectID {
		t.Fatalf("POSITIVE_ISSUANCE_CONTROL_FAILED err=%v claims=%+v", e, claims)
	}
	abnormal := gapJWT(secret, `{"alg":"HS256","typ":"JWT"}`, fmt.Sprintf(`{"sub":"user-B","uid":"user-A","investmentSubjectId":"subject-A","exp":%d,"iat":%d}`, now.Add(time.Hour).Unix(), now.Unix()))
	actual, err := verifyAccessToken(secret, abnormal, now)
	if err == nil {
		t.Fatalf("CONFIRMED_SIGNED_CLAIM_POLICY_FAILURE verifier accepted sub!=uid; claims=%+v; external HMAC forgery NOT demonstrated", actual)
	}
}

func TestJWTIdentityContractRegression(t *testing.T) {
	secret := []byte("jwt-identity-contract-local-key-32-bytes")
	now := time.Unix(1800000000, 0).UTC()
	user := StoredUser{ID: "user-A", InvestmentSubjectID: "financial-A"}
	legitimate, err := signAccessToken(secret, user, now, now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	claims, err := verifyAccessToken(secret, legitimate, now)
	if err != nil || claims.SubjectID != user.ID || claims.UserID != user.ID || claims.InvestmentSubjectID != user.InvestmentSubjectID {
		t.Fatalf("issuer/financial positive control: %v", err)
	}
	header := map[string]any{"alg": "HS256", "typ": "JWT"}
	normal := map[string]any{"sub": user.ID, "uid": user.ID, "investmentSubjectId": user.InvestmentSubjectID, "exp": now.Add(time.Hour).Unix(), "iat": now.Unix()}
	for _, name := range []string{"missing-sub", "empty-sub", "mismatch", "missing-uid", "empty-uid", "missing-financial-subject", "empty-financial-subject", "wrong-key", "invalid-signature", "wrong-algorithm", "wrong-type", "expired"} {
		t.Run(name, func(t *testing.T) {
			c := map[string]any{}
			for k, v := range normal {
				c[k] = v
			}
			h := map[string]any{}
			for k, v := range header {
				h[k] = v
			}
			key := secret
			switch name {
			case "missing-sub":
				delete(c, "sub")
			case "empty-sub":
				c["sub"] = ""
			case "mismatch":
				c["sub"] = "user-B"
			case "missing-uid":
				delete(c, "uid")
			case "empty-uid":
				c["uid"] = ""
			case "missing-financial-subject":
				delete(c, "investmentSubjectId")
			case "empty-financial-subject":
				c["investmentSubjectId"] = ""
			case "wrong-key":
				key = []byte("other-local-signing-secret-32-bytes")
			case "wrong-algorithm":
				h["alg"] = "HS512"
			case "wrong-type":
				h["typ"] = "refresh"
			case "expired":
				c["exp"] = now.Unix()
			}
			tok := campaignSignedJWT(t, key, h, c)
			if name == "invalid-signature" {
				parts := []byte(tok)
				parts[len(parts)-5] ^= 1
				tok = string(parts)
			}
			if _, err := verifyAccessToken(secret, tok, now); err == nil {
				t.Fatal("invalid token accepted")
			}
			svc := &Service{config: Config{AccessTokenSecret: secret}, clock: jwtIdentityClock{now: now}}
			if _, err := svc.AuthenticateAccessToken(tok); err != ErrInvalidSession {
				t.Fatalf("external error contract changed: %v", err)
			}
		})
	}
}

func TestJWTIdentityOriginalSourceBlobs(t *testing.T) {
	// Keep full historical sources, including unapproved JWT-02 RED assertions,
	// outside the ordinary approved-policy test suite. Never rewrite those files.
	for _, tc := range []struct{ path, blob string }{{"testdata/jwt01/pr261.go.txt", "dfe275965e8702b1d2abec96ff5579373875a57b"}, {"testdata/jwt01/pr265.go.txt", "21eac313f4b3d05cb55472ff627c83ecc66408e5"}} {
		data, err := os.ReadFile(tc.path)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(data, []byte("Future")) {
			t.Fatal("historical future-iat evidence absent")
		}
		h := sha1.New()
		fmt.Fprintf(h, "blob %d%c", len(data), byte(0))
		h.Write(data)
		if fmt.Sprintf("%x", h.Sum(nil)) != tc.blob {
			t.Fatal("independent source blob changed")
		}
	}
}

func TestEmailOwnershipJWTIdentityContractEvidence(t *testing.T) {
	t.Run("TestModule1CampaignJWTRequiresCanonicalSignedSubject", TestModule1CampaignJWTRequiresCanonicalSignedSubject)
	t.Run("TestAUTHGAPJWTCanonicalSignedIdentityContract", TestAUTHGAPJWTCanonicalSignedIdentityContract)
	t.Run("IdentityRegression", TestJWTIdentityContractRegression)
	t.Run("OriginalSourceBlobs", TestJWTIdentityOriginalSourceBlobs)
	t.Log("JWT01_APPROVED_IDENTITY_CONTRACT=VERIFIED JWT02_NEGATIVE_POLICY=UNCHANGED_RED_NOT_REMEDIATED JWT03_PARSER=UNCHANGED")
}

type jwtIdentityClock struct{ now time.Time }

func (c jwtIdentityClock) Now() time.Time { return c.now }
