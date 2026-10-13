package httpapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/openinvest/openinvest/backend-go/internal/auth"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type httpOwnershipFixture struct {
	httpAuthTestStore
	email, hash string
	used        bool
}

func (s *httpOwnershipFixture) BeginRegistration(_ context.Context, email string) error {
	s.email = email
	return nil
}
func (s *httpOwnershipFixture) LeaseVerificationDelivery(_ context.Context, hash string) (auth.VerificationDelivery, bool, error) {
	s.hash = hash
	return auth.VerificationDelivery{Email: s.email, ExpiresAt: time.Now().Add(time.Hour)}, true, nil
}
func (s *httpOwnershipFixture) FinishVerificationDelivery(context.Context, auth.VerificationDelivery, bool) error {
	return nil
}
func (s *httpOwnershipFixture) CompleteRegistration(ctx context.Context, hash string, r auth.RegistrationRecord) error {
	if s.used || hash != s.hash {
		return auth.ErrInvalidVerification
	}
	r.EmailNormalized = s.email
	r.UserID = uuid.NewString()
	r.InvestmentSubjectID = uuid.NewString()
	r.Now = time.Now()
	_, err := s.RegisterUser(ctx, r)
	s.sessions = map[string]auth.SessionRecord{}
	s.used = err == nil
	return err
}

type httpVerificationSink struct{ token string }

func (m *httpVerificationSink) SendVerification(_ context.Context, email, raw string) error {
	m.token = raw
	return nil
}
func TestEmailOwnershipHTTPAcknowledgmentAndActivation(t *testing.T) {
	store := &httpOwnershipFixture{}
	service, err := auth.NewService(store, nil, auth.Config{AccessTokenSecret: []byte("01234567890123456789012345678901"), RefreshCookieSecure: true})
	if err != nil {
		t.Fatal(err)
	}
	app, err := New(verticalslice.NewService(&importAPITestStore{}, fixedHTTPClock{}), service, []byte("test-import-review-token-secret-32-bytes"))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/api/v1/auth/register", "/api/v1/auth/email/resend"} {
		r := authRequest(t, app, "POST", path, `{"email":"owner@example.test"}`, "", "")
		if r.StatusCode != 202 || len(r.Cookies()) != 0 || r.Header.Get("Cache-Control") != "no-store" {
			t.Fatalf("unsafe acknowledgment status=%d", r.StatusCode)
		}
		var body struct{ Data map[string]string }
		if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		r.Body.Close()
		if body.Data["status"] != "REQUEST_ACCEPTED" || len(body.Data) != 2 {
			t.Fatal("acknowledgment leaked identity/session")
		}
		if store.registerCalls != 0 || len(store.sessions) != 0 {
			t.Fatal("identity/session before proof")
		}
	}
	sink := &httpVerificationSink{}
	if _, err = service.DeliverOne(context.Background(), sink); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256([]byte(sink.token))
	if store.hash != hex.EncodeToString(sum[:]) {
		t.Fatal("raw token storage")
	}
	payload, _ := json.Marshal(map[string]string{"token": sink.token, "password": "owner-selected-correct-horse", "language": "en", "theme": "system", "timezone": "UTC"})
	r := authRequest(t, app, "POST", "/api/v1/auth/email/verify", string(payload), "", "")
	defer r.Body.Close()
	if r.StatusCode != 200 || len(r.Cookies()) != 0 || len(store.sessions) != 0 || store.registerCalls != 1 {
		t.Fatal("activation must create exactly one account without session")
	}
	login := authRequest(t, app, "POST", "/api/v1/auth/login", `{"email":"owner@example.test","password":"owner-selected-correct-horse"}`, "", "")
	defer login.Body.Close()
	if login.StatusCode != 200 {
		t.Fatal("normal login unavailable")
	}
	requireCookie(t, login, auth.RefreshCookieName)
}
func TestEmailOwnershipHTTPUnavailableDoesNotFallBack(t *testing.T) {
	app := newHTTPAuthApp(t, newHTTPAuthService(t, &httpAuthTestStore{}))
	r := authRequest(t, app, http.MethodPost, "/api/v1/auth/register", `{"email":"owner@example.test"}`, "", "")
	defer r.Body.Close()
	if r.StatusCode != 503 || len(r.Cookies()) != 0 {
		t.Fatal("missing ownership backend must fail closed")
	}
}

func TestEmailOwnershipVerifyTimezoneSyntaxBeforePersistence(t *testing.T) {
	for _, tc := range []struct {
		zone           string
		status, writes int
	}{{" Asia/Baku", 400, 0}, {"UTC+04:00", 400, 0}, {"UTC", 200, 1}, {"Asia/Baku", 200, 1}} {
		t.Run(tc.zone, func(t *testing.T) {
			store := &httpOwnershipFixture{}
			service, err := auth.NewService(store, nil, auth.Config{AccessTokenSecret: []byte("01234567890123456789012345678901")})
			if err != nil {
				t.Fatal(err)
			}
			app := newHTTPAuthApp(t, service)
			if err := service.BeginEmailRegistration(context.Background(), "timezone@example.test"); err != nil {
				t.Fatal(err)
			}
			sink := &httpVerificationSink{}
			if ok, err := service.DeliverOne(context.Background(), sink); err != nil || !ok {
				t.Fatalf("delivery: %t %v", ok, err)
			}
			payload, _ := json.Marshal(map[string]string{"token": sink.token, "password": "owner-selected-correct-horse", "language": "en", "theme": "system", "timezone": tc.zone})
			request := httptest.NewRequest("POST", "/api/v1/auth/email/verify", bytes.NewReader(payload))
			request.Header.Set("Content-Type", "application/json")
			response, err := app.Test(request, fiber.TestConfig{Timeout: 5 * time.Second, FailOnTimeout: true})
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			if response.StatusCode != tc.status || store.registerCalls != tc.writes {
				t.Fatalf("status=%d writes=%d expected=%d/%d", response.StatusCode, store.registerCalls, tc.status, tc.writes)
			}
			if tc.writes == 0 && store.used {
				t.Fatal("invalid timezone consumed verification proof")
			}
		})
	}
	if !t.Failed() {
		t.Log("TIMEZONE_COMPLETION_NEGATIVE_AND_POSITIVE_CONTROLS=PASS")
	}
}
