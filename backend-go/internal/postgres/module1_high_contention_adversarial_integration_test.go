package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"runtime"
	"sync"
	"testing"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/openinvest/openinvest/backend-go/internal/auth"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestModule1BoundedHighContentionRefreshCampaign(t *testing.T) {
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
		AccessTokenSecret:   []byte("module1-bounded-contention-secret-32b"),
		RefreshCookieSecure: true,
	})
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	baselineGoroutines := runtime.NumGoroutine()
	baselineFDs := module1FDCount(t)
	var baselineMem runtime.MemStats
	runtime.ReadMemStats(&baselineMem)
	peakGoroutines := baselineGoroutines
	totalAttempts := 0
	totalSuccesses := 0

	for _, contenders := range []int{2, 4, 8, 16, 32} {
		t.Run("contenders_"+itoaModule1(contenders), func(t *testing.T) {
			for repetition := 0; repetition < 25; repetition++ {
				root, err := service.Register(ctx, auth.RegistrationRequest{
					Email:    "module1-contention-" + uuid.NewString() + "@example.test",
					Password: "correct horse battery staple",
					Language: auth.LanguageEN,
					Theme:    auth.ThemeSystem,
					Timezone: "UTC",
				})
				if err != nil {
					t.Fatalf("register repetition=%d: %v", repetition, err)
				}

				type outcome struct {
					result auth.AuthResult
					err    error
				}
				outcomes := make(chan outcome, contenders)
				start := make(chan struct{})
				var ready sync.WaitGroup
				ready.Add(contenders)
				for i := 0; i < contenders; i++ {
					go func() {
						ready.Done()
						<-start
						result, err := service.Refresh(ctx, root.RefreshToken, root.Session.CSRFToken)
						outcomes <- outcome{result: result, err: err}
					}()
				}
				ready.Wait()
				if current := runtime.NumGoroutine(); current > peakGoroutines {
					peakGoroutines = current
				}
				close(start)

				successes := 0
				var successful []auth.AuthResult
				for i := 0; i < contenders; i++ {
					got := <-outcomes
					totalAttempts++
					if got.err == nil {
						successes++
						totalSuccesses++
						successful = append(successful, got.result)
						continue
					}
					if !errors.Is(got.err, auth.ErrInvalidSession) {
						t.Fatalf("contenders=%d repetition=%d unexpected refresh error: %v", contenders, repetition, got.err)
					}
				}
				if successes > 1 {
					t.Fatalf("contenders=%d repetition=%d created %d successful descendants from one refresh credential", contenders, repetition, successes)
				}

				var active int
				if err := db.QueryRowContext(ctx, `
					SELECT count(*)
					FROM identity.sessions
					WHERE user_id = $1
					  AND session_state = 'active'
				`, root.User.ID).Scan(&active); err != nil {
					t.Fatalf("query active sessions contenders=%d repetition=%d: %v", contenders, repetition, err)
				}
				if active > 1 {
					t.Fatalf("contenders=%d repetition=%d left %d active session lineages", contenders, repetition, active)
				}

				// Independent usability oracle: among any returned successful descendants,
				// no more than one may still be usable after the replay/contention outcome.
				usable := 0
				for _, candidate := range successful {
					if _, err := service.Refresh(ctx, candidate.RefreshToken, candidate.Session.CSRFToken); err == nil {
						usable++
					} else if !errors.Is(err, auth.ErrInvalidSession) {
						t.Fatalf("contenders=%d repetition=%d descendant validation error: %v", contenders, repetition, err)
					}
				}
				if usable > 1 {
					t.Fatalf("contenders=%d repetition=%d left %d usable descendant credentials", contenders, repetition, usable)
				}
			}
		})
	}

	// Post-pressure recovery: legitimate registration and refresh must still work.
	recovery, err := service.Register(ctx, auth.RegistrationRequest{
		Email:    "module1-contention-recovery-" + uuid.NewString() + "@example.test",
		Password: "correct horse battery staple",
		Language: auth.LanguageEN,
		Theme:    auth.ThemeSystem,
		Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("post-contention registration failed: %v", err)
	}
	if _, err := service.Refresh(ctx, recovery.RefreshToken, recovery.Session.CSRFToken); err != nil {
		t.Fatalf("post-contention legitimate refresh failed: %v", err)
	}

	afterGoroutines := runtime.NumGoroutine()
	afterFDs := module1FDCount(t)
	var afterMem runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&afterMem)
	t.Logf("CONTENTION_LEVELS=2,4,8,16,32 REPETITIONS_PER_LEVEL=25 TOTAL_ATTEMPTS=%d TOTAL_SUCCESSES=%d", totalAttempts, totalSuccesses)
	t.Logf("GOROUTINES_BASELINE=%d GOROUTINES_PEAK_OBSERVED=%d GOROUTINES_AFTER=%d", baselineGoroutines, peakGoroutines, afterGoroutines)
	t.Logf("FD_BASELINE=%d FD_AFTER=%d", baselineFDs, afterFDs)
	t.Logf("HEAP_ALLOC_BASELINE=%d HEAP_ALLOC_AFTER_GC=%d", baselineMem.HeapAlloc, afterMem.HeapAlloc)
	t.Log("RACE_CREATES_MORE_VALID_SESSION_LINEAGES_THAN_SERIAL_POLICY_ALLOWS=NO")
	t.Log("POST_ATTACK_SERVICE_RECOVERY=PASS")
}

func itoaModule1(value int) string {
	switch value {
	case 2:
		return "2"
	case 4:
		return "4"
	case 8:
		return "8"
	case 16:
		return "16"
	case 32:
		return "32"
	default:
		return "unknown"
	}
}


func module1FDCount(t *testing.T) int {
	t.Helper()
	entries, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		t.Logf("FD_COUNT=NOT_AVAILABLE err=%v", err)
		return -1
	}
	return len(entries)
}
