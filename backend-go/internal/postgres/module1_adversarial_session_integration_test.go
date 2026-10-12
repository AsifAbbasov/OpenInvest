package postgres_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/openinvest/openinvest/backend-go/internal/auth"
	"github.com/openinvest/openinvest/backend-go/internal/httpapi"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func module1TokenHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func module1Percentile(values []time.Duration, p float64) time.Duration {
	if len(values) == 0 {
		return 0
	}
	copyValues := append([]time.Duration(nil), values...)
	sort.Slice(copyValues, func(i, j int) bool { return copyValues[i] < copyValues[j] })
	index := int(float64(len(copyValues)-1) * p)
	return copyValues[index]
}

func module1AuthHTTPApp(t *testing.T, store *postgres.Store) *fiber.App {
	t.Helper()
	authService, err := auth.NewService(store, verticalslice.SystemClock{}, auth.Config{
		AccessTokenSecret: []byte("01234567890123456789012345678901"),
		RefreshCookieSecure: true,
	})
	if err != nil {
		t.Fatalf("auth service: %v", err)
	}
	app, err := httpapi.New(verticalslice.NewService(store, verticalslice.SystemClock{}), authService, []byte("module1-import-review-secret-32-bytes"))
	if err != nil {
		t.Fatalf("http app: %v", err)
	}
	return app
}

func module1JSONRequest(method, path, body string) *http.Request {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "198.51.100.25:44321"
	return req
}

func TestModule1RegistrationEnumerationAndPreRegistrationSquatting(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}
	store, err := postgres.Open(databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	app := module1AuthHTTPApp(t, store)

	email := "module1-enum-" + uuid.NewString() + "@example.test"
	password := "correct horse battery staple"
	registerBody := func(address string) string {
		return fmt.Sprintf(`{"email":%q,"password":%q,"language":"en","theme":"system","timezone":"UTC"}`, address, password)
	}

	first, err := app.Test(module1JSONRequest(http.MethodPost, "/api/v1/auth/register", registerBody(email)))
	if err != nil {
		t.Fatal(err)
	}
	first.Body.Close()
	if first.StatusCode != http.StatusCreated {
		t.Fatalf("AUTH-ADV-02 fixture registration expected=201 got=%d", first.StatusCode)
	}

	login, err := app.Test(module1JSONRequest(http.MethodPost, "/api/v1/auth/login", fmt.Sprintf(`{"email":%q,"password":%q}`, email, password)))
	if err != nil {
		t.Fatal(err)
	}
	login.Body.Close()
	if login.StatusCode != http.StatusOK {
		t.Fatalf("PRE_REGISTRATION_SQUATTING active account expected immediate login=200 got=%d", login.StatusCode)
	}

	duplicate, err := app.Test(module1JSONRequest(http.MethodPost, "/api/v1/auth/register", registerBody(email)))
	if err != nil {
		t.Fatal(err)
	}
	var duplicatePayload map[string]any
	_ = json.NewDecoder(duplicate.Body).Decode(&duplicatePayload)
	duplicate.Body.Close()
	if duplicate.StatusCode != http.StatusConflict {
		t.Fatalf("AUTH_ADV_02_DIRECT_RESPONSE_ENUMERATION duplicate expected=409 got=%d", duplicate.StatusCode)
	}
	encoded, _ := json.Marshal(duplicatePayload)
	if !bytes.Contains(encoded, []byte("Email is already registered")) {
		t.Fatalf("AUTH_ADV_02 direct enumeration message changed: %s", encoded)
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var state string
	if err := db.QueryRow(`SELECT account_state FROM identity.users WHERE email_normalized=$1`, email).Scan(&state); err != nil {
		t.Fatal(err)
	}
	if state != "active" {
		t.Fatalf("expected immediate active account, got %q", state)
	}
	t.Logf("AUTH_ADV_02_DIRECT_RESPONSE_ENUMERATION=CONFIRMED EMAIL_OWNERSHIP_VERIFICATION=ABSENT ACCOUNT_STATE=%s PRE_REGISTRATION_SQUATTING=CONFIRMED", state)

	const samples = 10
	newDurations := make([]time.Duration, 0, samples)
	duplicateDurations := make([]time.Duration, 0, samples)
	for i := 0; i < samples; i++ {
		newEmail := fmt.Sprintf("module1-enum-sample-%s-%d@example.test", uuid.NewString(), i)
		req := module1JSONRequest(http.MethodPost, "/api/v1/auth/register", registerBody(newEmail))
		req.RemoteAddr = fmt.Sprintf("198.51.100.%d:44321", 50+i)
		start := time.Now()
		resp, runErr := app.Test(req)
		newDurations = append(newDurations, time.Since(start))
		if runErr != nil {
			t.Fatal(runErr)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("new sample %d expected=201 got=%d", i, resp.StatusCode)
		}

		dupReq := module1JSONRequest(http.MethodPost, "/api/v1/auth/register", registerBody(email))
		dupReq.RemoteAddr = fmt.Sprintf("198.51.100.%d:44321", 100+i)
		start = time.Now()
		dupResp, runErr := app.Test(dupReq)
		duplicateDurations = append(duplicateDurations, time.Since(start))
		if runErr != nil {
			t.Fatal(runErr)
		}
		dupResp.Body.Close()
		if dupResp.StatusCode != http.StatusConflict {
			t.Fatalf("duplicate sample %d expected=409 got=%d", i, dupResp.StatusCode)
		}
	}
	t.Logf("AUTH_ADV_02_TIMING_ENUMERATION=MEASURED_NOT_CLASSIFIED samples=%d NEW_P50=%s NEW_P95=%s DUP_P50=%s DUP_P95=%s",
		samples,
		module1Percentile(newDurations, .50), module1Percentile(newDurations, .95),
		module1Percentile(duplicateDurations, .50), module1Percentile(duplicateDurations, .95))
}

func TestModule1RefreshHighContentionPersistentStateOracle(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}
	store, err := postgres.Open(databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	service, err := auth.NewService(store, verticalslice.SystemClock{}, auth.Config{
		AccessTokenSecret: []byte("01234567890123456789012345678901"),
		RefreshCookieSecure: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var baselineMem runtime.MemStats
	runtime.ReadMemStats(&baselineMem)
	goroutineBaseline := runtime.NumGoroutine()

	levels := []int{2, 4, 8, 16, 32}
	const repetitions = 25
	totalCalls := 0
	for _, contenders := range levels {
		levelSuccess := 0
		levelInvalid := 0
		for repetition := 0; repetition < repetitions; repetition++ {
			now := time.Now().UTC()
			userID := uuid.NewString()
			subjectID := uuid.NewString()
			sessionID := uuid.NewString()
			refresh := "m1-refresh-" + uuid.NewString()
			csrf := "m1-csrf-" + uuid.NewString()
			email := "m1-contention-" + uuid.NewString() + "@example.test"
			_, err := store.RegisterUser(context.Background(), auth.RegistrationRecord{
				UserID: userID,
				InvestmentSubjectID: subjectID,
				EmailNormalized: email,
				PasswordHash: "synthetic-not-used-by-refresh-test",
				Language: auth.LanguageEN,
				Theme: auth.ThemeSystem,
				Timezone: "UTC",
				Now: now,
				Session: auth.SessionRecord{
					SessionID: sessionID,
					UserID: userID,
					RefreshTokenHash: module1TokenHash(refresh),
					CSRFTokenHash: module1TokenHash(csrf),
					ExpiresAt: now.Add(time.Hour),
					Now: now,
				},
			})
			if err != nil {
				t.Fatalf("register direct fixture: %v", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			var wg sync.WaitGroup
			start := make(chan struct{})
			results := make(chan error, contenders)
			wg.Add(contenders)
			for i := 0; i < contenders; i++ {
				go func() {
					defer wg.Done()
					<-start
					_, refreshErr := service.Refresh(ctx, refresh, csrf)
					results <- refreshErr
				}()
			}
			close(start)
			wg.Wait()
			cancel()
			close(results)

			successes := 0
			invalid := 0
			for refreshErr := range results {
				totalCalls++
				if refreshErr == nil {
					successes++
					continue
				}
				if refreshErr == auth.ErrInvalidSession {
					invalid++
					continue
				}
				t.Fatalf("contenders=%d repetition=%d unexpected refresh error: %v", contenders, repetition, refreshErr)
			}
			levelSuccess += successes
			levelInvalid += invalid
			if successes > 1 {
				t.Fatalf("RACE_CREATES_MORE_VALID_SESSION_LINEAGES_THAN_SERIAL_POLICY_ALLOWS contenders=%d repetition=%d successes=%d", contenders, repetition, successes)
			}

			var active int
			if err := db.QueryRow(`SELECT count(*) FROM identity.sessions WHERE user_id=$1 AND session_state='active'`, userID).Scan(&active); err != nil {
				t.Fatal(err)
			}
			if active != 0 {
				t.Fatalf("persistent-state oracle: contenders=%d repetition=%d active_sessions_after=%d", contenders, repetition, active)
			}
		}
		var connections int
		_ = db.QueryRow(`SELECT count(*) FROM pg_stat_activity WHERE datname=current_database()`).Scan(&connections)
		t.Logf("PROPERTY=M1-SESSION-CONTENTION CONTENDERS=%d REPETITIONS=%d SUCCESS_COUNT=%d AUTH_FAILURE_COUNT=%d ACTIVE_SESSIONS_AFTER=0 DB_CONNECTIONS_OBSERVED=%d", contenders, repetitions, levelSuccess, levelInvalid, connections)
	}

	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	var recoveryMem runtime.MemStats
	runtime.ReadMemStats(&recoveryMem)
	goroutineRecovery := runtime.NumGoroutine()
	t.Logf("RESOURCE_EVIDENCE TOTAL_REFRESH_CALLS=%d GOROUTINES_BASELINE=%d GOROUTINES_RECOVERY=%d MEMORY_BASELINE=%d MEMORY_RECOVERY=%d",
		totalCalls, goroutineBaseline, goroutineRecovery, baselineMem.Alloc, recoveryMem.Alloc)
	if goroutineRecovery > goroutineBaseline+16 {
		t.Fatalf("unbounded goroutine recovery delta baseline=%d recovery=%d", goroutineBaseline, goroutineRecovery)
	}
}
