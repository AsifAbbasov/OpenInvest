package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/google/uuid"
	"github.com/openinvest/openinvest/backend-go/internal/auth"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
)

func TestAuthAdv02RegistrationEnumerationHTTPGenericAcrossMembershipStates(t *testing.T) {
	url := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if url == "" { t.Skip("isolated PostgreSQL service required") }
	store, err := postgres.Open(url)
	if err != nil { t.Fatal(err) }
	defer store.Close()
	service, err := auth.NewService(store, fixedHTTPClock{}, auth.Config{
		AccessTokenSecret: []byte("01234567890123456789012345678901"),
		RefreshCookieSecure: true,
	})
	if err != nil { t.Fatal(err) }
	app := newHTTPAuthApp(t, service)
	db, err := sql.Open("pgx", url)
	if err != nil { t.Fatal(err) }
	defer db.Close()
	if _, err := db.Exec(`DELETE FROM identity.registration_intents`); err != nil { t.Fatal(err) }

	newEmail := "adv02-http-new-" + uuid.NewString() + "@example.test"
	existingEmail := "adv02-http-existing-" + uuid.NewString() + "@example.test"
	expiredEmail := "adv02-http-expired-" + uuid.NewString() + "@example.test"
	if _, err := service.Register(context.Background(), auth.RegistrationRequest{
		Email: existingEmail, Password: "legacy-existing-password-123", Language: "en", Theme: "system", Timezone: "UTC",
	}); err != nil { t.Fatal(err) }

	type observed struct {
		status int
		data map[string]string
		cache string
		cookies int
	}
	request := func(email string) observed {
		r := authRequest(t, app, http.MethodPost, "/api/v1/auth/register", `{"email":"`+email+`"}`, "", "")
		defer r.Body.Close()
		var payload struct { Data map[string]string `json:"data"` }
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil { t.Fatal(err) }
		return observed{r.StatusCode, payload.Data, r.Header.Get("Cache-Control"), len(r.Cookies())}
	}
	assertGeneric := func(name string, got observed) {
		t.Helper()
		if got.status != http.StatusAccepted || got.cookies != 0 || got.cache != "no-store" {
			t.Fatalf("%s unsafe envelope status=%d cookies=%d cache=%q", name, got.status, got.cookies, got.cache)
		}
		if len(got.data) != 2 || got.data["status"] != "REQUEST_ACCEPTED" || got.data["message"] != "If this address is eligible, instructions will be sent." {
			t.Fatalf("%s membership-distinguishing body=%v", name, got.data)
		}
	}

	firstNew := request(newEmail); assertGeneric("new", firstNew)
	pending := request(newEmail); assertGeneric("pending", pending)
	existing := request(existingEmail); assertGeneric("existing", existing)
	firstExpired := request(expiredEmail); assertGeneric("expired_seed", firstExpired)
	if _, err := db.Exec(`UPDATE identity.registration_intents SET expires_at=statement_timestamp()-interval '1 second', intent_state='expired' WHERE email_normalized=$1`, expiredEmail); err != nil { t.Fatal(err) }
	expired := request(expiredEmail); assertGeneric("expired", expired)

	for name, got := range map[string]observed{"pending":pending,"existing":existing,"expired":expired} {
		if got.status != firstNew.status || got.cache != firstNew.cache || got.cookies != firstNew.cookies ||
			got.data["status"] != firstNew.data["status"] || got.data["message"] != firstNew.data["message"] {
			t.Fatalf("%s differs from new registration: new=%+v got=%+v", name, firstNew, got)
		}
	}

	var newUsers int
	if err := db.QueryRow(`SELECT count(*) FROM identity.users WHERE email_normalized IN ($1,$2)`, newEmail, expiredEmail).Scan(&newUsers); err != nil { t.Fatal(err) }
	if newUsers != 0 { t.Fatalf("registration acknowledgment activated unverified users=%d", newUsers) }
	t.Log("REGISTRATION_ENUMERATION=NOT_OBSERVED_FOR_NEW_EXISTING_PENDING_EXPIRED_HTTP_ENVELOPES")
}

func TestAuthAdv02InvalidExpiredAndConsumedVerificationHTTPFailClosed(t *testing.T) {
	url := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if url == "" { t.Skip("isolated PostgreSQL service required") }
	store, err := postgres.Open(url)
	if err != nil { t.Fatal(err) }
	defer store.Close()
	service, err := auth.NewService(store, fixedHTTPClock{}, auth.Config{
		AccessTokenSecret: []byte("01234567890123456789012345678901"),
		RefreshCookieSecure: true,
	})
	if err != nil { t.Fatal(err) }
	app := newHTTPAuthApp(t, service)
	db, err := sql.Open("pgx", url)
	if err != nil { t.Fatal(err) }
	defer db.Close()
	if _, err := db.Exec(`DELETE FROM identity.registration_intents`); err != nil { t.Fatal(err) }

	postVerify := func(token string) *http.Response {
		body, _ := json.Marshal(map[string]string{
			"token": token, "password":"owner-selected-password-123", "language":"en", "theme":"system", "timezone":"UTC",
		})
		return authRequest(t, app, http.MethodPost, "/api/v1/auth/email/verify", string(body), "", "")
	}
	assertInvalid := func(name string, r *http.Response) {
		t.Helper(); defer r.Body.Close()
		var payload struct {
			Error struct {
				Code string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil { t.Fatal(err) }
		if r.StatusCode != http.StatusBadRequest || len(r.Cookies()) != 0 || payload.Error.Code != "INVALID_VERIFICATION" {
			t.Fatalf("%s status=%d cookies=%d code=%q", name, r.StatusCode, len(r.Cookies()), payload.Error.Code)
		}
	}

	assertInvalid("invalid", postVerify("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"))

	email := "adv02-http-token-" + uuid.NewString() + "@example.test"
	if err := service.BeginEmailRegistration(context.Background(), email); err != nil { t.Fatal(err) }
	sink := &httpVerificationSink{}
	if found, err := service.DeliverOne(context.Background(), sink); err != nil || !found { t.Fatalf("deliver found=%t err=%v", found, err) }
	if _, err := db.Exec(`UPDATE identity.email_verification_tokens SET expires_at=statement_timestamp()-interval '1 second' WHERE intent_id=(SELECT id FROM identity.registration_intents WHERE email_normalized=$1)`, email); err != nil { t.Fatal(err) }
	assertInvalid("expired", postVerify(sink.token))

	if _, err := db.Exec(`UPDATE identity.registration_intents SET expires_at=statement_timestamp()-interval '1 second', intent_state='expired' WHERE email_normalized=$1`, email); err != nil { t.Fatal(err) }
	if err := service.BeginEmailRegistration(context.Background(), email); err != nil { t.Fatal(err) }
	sink2 := &httpVerificationSink{}
	if found, err := service.DeliverOne(context.Background(), sink2); err != nil || !found { t.Fatalf("redeliver found=%t err=%v", found, err) }
	ok := postVerify(sink2.token)
	if ok.StatusCode != http.StatusOK || len(ok.Cookies()) != 0 { ok.Body.Close(); t.Fatalf("valid activation status=%d cookies=%d", ok.StatusCode, len(ok.Cookies())) }
	ok.Body.Close()
	assertInvalid("consumed_replay", postVerify(sink2.token))
	t.Log("TOKEN_EXPIRATION=PASS TOKEN_REPLAY=PASS INVALID_TOKEN=PASS")
}
