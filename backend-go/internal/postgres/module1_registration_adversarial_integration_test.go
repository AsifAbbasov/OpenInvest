package postgres_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/openinvest/openinvest/backend-go/internal/auth"
	"github.com/openinvest/openinvest/backend-go/internal/httpapi"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestModule1RegistrationEnumerationAndPreRegistrationSquattingEvidence(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}

	store, err := postgres.Open(databaseURL)
	if err != nil {
		t.Fatalf("open postgres store: %v", err)
	}
	defer store.Close()

	authService, err := auth.NewService(store, verticalslice.SystemClock{}, auth.Config{
		AccessTokenSecret:   []byte("module1-audit-access-token-secret-32bytes"),
		RefreshCookieSecure: true,
	})
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}
	app, err := httpapi.New(
		verticalslice.NewService(store, verticalslice.SystemClock{}),
		authService,
		[]byte("module1-audit-import-review-secret-32bytes"),
	)
	if err != nil {
		t.Fatalf("new http app: %v", err)
	}

	suffix := uuid.NewString()
	email := "module1-" + suffix + "@example.test"
	registerBody := func(value string) string {
		payload := map[string]any{
			"email": value,
			"password": "correct horse battery staple",
			"language": "en",
			"theme": "system",
			"timezone": "UTC",
		}
		raw, _ := json.Marshal(payload)
		return string(raw)
	}
	send := func(body string) *http.Response {
		t.Helper()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, requestErr := app.Test(req)
		if requestErr != nil {
			t.Fatalf("registration request: %v", requestErr)
		}
		return resp
	}

	first := send(registerBody(email))
	defer first.Body.Close()
	if first.StatusCode != http.StatusCreated {
		t.Fatalf("new synthetic registration status=%d, want 201", first.StatusCode)
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open sql db: %v", err)
	}
	defer db.Close()
	var accountState string
	var normalized string
	if err := db.QueryRowContext(context.Background(),
		"SELECT account_state, email_normalized FROM identity.users WHERE email_normalized=$1",
		email,
	).Scan(&accountState, &normalized); err != nil {
		t.Fatalf("query registered identity: %v", err)
	}
	if accountState != "active" {
		t.Fatalf("new registration is not immediately active: %q", accountState)
	}
	if normalized != email {
		t.Fatalf("unexpected normalized email: %q", normalized)
	}

	duplicateInput := "  " + strings.ToUpper(email) + "  "
	second := send(registerBody(duplicateInput))
	defer second.Body.Close()
	if second.StatusCode != http.StatusConflict {
		t.Fatalf("duplicate normalized registration status=%d, want 409", second.StatusCode)
	}
	var conflict struct {
		Error struct {
			Code string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(second.Body).Decode(&conflict); err != nil {
		t.Fatalf("decode duplicate response: %v", err)
	}
	if conflict.Error.Code != "CONFLICT" || conflict.Error.Message != "Email is already registered" {
		t.Fatalf("duplicate response lost direct enumeration signal: %+v", conflict.Error)
	}

	var userCount int
	if err := db.QueryRowContext(context.Background(),
		"SELECT count(*) FROM identity.users WHERE email_normalized=$1",
		email,
	).Scan(&userCount); err != nil {
		t.Fatalf("count normalized identities: %v", err)
	}
	if userCount != 1 {
		t.Fatalf("normalization collision created %d identities, want 1", userCount)
	}

	t.Log("AUTH_ADV_02_DIRECT_RESPONSE_ENUMERATION=CONFIRMED")
	t.Log("EMAIL_OWNERSHIP_VERIFICATION=ABSENT_BY_RUNTIME_REGISTRATION_FLOW")
	t.Log("PRE_REGISTRATION_SQUATTING=CONFIRMED")
	t.Logf("ACCOUNT_STATE_AFTER_REGISTRATION=%s NORMALIZED_IDENTITY_ROWS=%d", accountState, userCount)
}

func TestModule1RegistrationTimingEvidenceBounded(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}
	store, err := postgres.Open(databaseURL)
	if err != nil {
		t.Fatalf("open postgres store: %v", err)
	}
	defer store.Close()
	service, err := auth.NewService(store, verticalslice.SystemClock{}, auth.Config{
		AccessTokenSecret:   []byte("module1-audit-access-token-secret-32bytes"),
		RefreshCookieSecure: true,
	})
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}
	ctx := context.Background()
	password := "correct horse battery staple"
	base := "module1-timing-" + uuid.NewString() + "@example.test"
	if _, err := service.Register(ctx, auth.RegistrationRequest{
		Email: base, Password: password, Language: auth.LanguageEN, Theme: auth.ThemeSystem, Timezone: "UTC",
	}); err != nil {
		t.Fatalf("seed duplicate identity: %v", err)
	}

	const samples = 6
	newDurations := make([]time.Duration, 0, samples)
	duplicateDurations := make([]time.Duration, 0, samples)
	for i := 0; i < samples; i++ {
		start := time.Now()
		_, err := service.Register(ctx, auth.RegistrationRequest{
			Email: "module1-timing-new-" + uuid.NewString() + "@example.test",
			Password: password, Language: auth.LanguageEN, Theme: auth.ThemeSystem, Timezone: "UTC",
		})
		newDurations = append(newDurations, time.Since(start))
		if err != nil {
			t.Fatalf("new timing registration: %v", err)
		}

		start = time.Now()
		_, err = service.Register(ctx, auth.RegistrationRequest{
			Email: base, Password: password, Language: auth.LanguageEN, Theme: auth.ThemeSystem, Timezone: "UTC",
		})
		duplicateDurations = append(duplicateDurations, time.Since(start))
		if err != auth.ErrEmailAlreadyExists {
			t.Fatalf("duplicate timing registration err=%v", err)
		}
	}
	median := func(values []time.Duration) time.Duration {
		copyValues := append([]time.Duration(nil), values...)
		for i := 1; i < len(copyValues); i++ {
			for j := i; j > 0 && copyValues[j] < copyValues[j-1]; j-- {
				copyValues[j], copyValues[j-1] = copyValues[j-1], copyValues[j]
			}
		}
		return copyValues[len(copyValues)/2]
	}
	t.Logf("AUTH_ADV_02_TIMING_SAMPLES_PER_CLASS=%d NEW_MEDIAN=%s DUPLICATE_MEDIAN=%s", samples, median(newDurations), median(duplicateDurations))
	t.Log("AUTH_ADV_02_TIMING_ENUMERATION=EVIDENCE_COLLECTED_NOT_CLASSIFIED_FROM_SMALL_SAMPLE")
}
