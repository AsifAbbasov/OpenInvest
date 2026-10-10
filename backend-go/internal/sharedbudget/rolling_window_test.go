package sharedbudget

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
)

func testRedisAuthority(t *testing.T) *RedisAuthority {
	t.Helper()
	url := os.Getenv("OPENINVEST_SHARED_BUDGET_TEST_REDIS_URL")
	if url == "" {
		t.Skip("real Redis test backend required")
	}
	authority, err := NewRedisAuthority(url, fmt.Sprintf("openinvest:test:rolling:%d:%s", time.Now().UnixNano(), t.Name()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = authority.Close() })
	if err := authority.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
	return authority
}

// A first entry expires while later admissions are still inside the same
// rolling interval. Only that first entry may be replaced; no fresh full
// allowance is issued at the oldest entry's expiry.
func TestRollingWindowRetainsLaterAdmissions(t *testing.T) {
	const window = 2 * time.Second
	for _, property := range []string{"ENDPOINT_GLOBAL", "PER_CLIENT", "PROVIDER"} {
		t.Run(property, func(t *testing.T) {
			authority := testRedisAuthority(t)
			limit := map[string]int{"ENDPOINT_GLOBAL": 48, "PER_CLIENT": 12, "PROVIDER": 60}[property]
			admit := func(i int) bool {
				t.Helper()
				var allowed bool
				var err error
				switch property {
				case "ENDPOINT_GLOBAL":
					allowed, err = authority.AdmitEndpoint(context.Background(), fmt.Sprintf("client-%d", i), 100, limit, window)
				case "PER_CLIENT":
					allowed, err = authority.AdmitEndpoint(context.Background(), "same-client", limit, 100, window)
				default:
					allowed, err = authority.AdmitProvider(context.Background(), limit, window)
				}
				if err != nil {
					t.Fatal(err)
				}
				return allowed
			}
			if !admit(0) {
				t.Fatal("first admission rejected")
			}
			start := time.Now()
			time.Sleep(window / 2)
			for i := 1; i < limit; i++ {
				if !admit(i) {
					t.Fatal("initial allowance rejected")
				}
			}
			later := time.Now()
			// Use a generous half-window gap to tolerate ordinary CI scheduling jitter.
			time.Sleep(time.Until(start.Add(window + 100*time.Millisecond)))
			if time.Since(later) >= window {
				t.Fatal("test scheduling exceeded rolling fixture interval")
			}
			accepted := 0
			for i := 0; i < limit; i++ {
				if admit(100 + i) {
					accepted++
				}
			}
			if accepted != 1 {
				t.Fatalf("replacement allowance=%d want=1", accepted)
			}
			t.Logf("BOUNDARY_BURST_%s=BLOCKED_TO_ROLLING_LIMIT ROLLING_INTERVAL_ACCEPTED=%d LIMIT=%d", property, limit-1+accepted, limit)
		})
	}
}

func TestRollingEndpointAdmissionIsAtomic(t *testing.T) {
	authority := testRedisAuthority(t)
	ctx := context.Background()
	window := time.Second
	must := func(identity string, client, global int, want bool) {
		t.Helper()
		got, err := authority.AdmitEndpoint(ctx, identity, client, global, window)
		if err != nil || got != want {
			t.Fatalf("identity=%s admitted=%v want=%v err=%v", identity, got, want, err)
		}
	}
	must("first", 1, 2, true)
	must("first", 1, 2, false) // Does not consume a global admission.
	must("second", 1, 2, true)
	must("third", 1, 2, false) // Does not create a client admission.
	if got := authority.client.ZCard(ctx, authority.key("corporate-actions:endpoint:global")).Val(); got != 2 {
		t.Fatalf("global=%d", got)
	}
}

func TestRollingConcurrentAdmissionAndTTL(t *testing.T) {
	authority := testRedisAuthority(t)
	ctx := context.Background()
	const limit = 12
	const window = time.Second
	var wg sync.WaitGroup
	accepted := make(chan bool, 40)
	errors := make(chan error, 40)
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			allowed, err := authority.AdmitProvider(ctx, limit, window)
			accepted <- allowed
			errors <- err
		}()
	}
	wg.Wait()
	close(accepted)
	close(errors)
	count := 0
	for err := range errors {
		if err != nil {
			t.Fatal(err)
		}
	}
	for allowed := range accepted {
		if allowed {
			count++
		}
	}
	if count != limit {
		t.Fatalf("concurrent accepted=%d want=%d", count, limit)
	}
	key := authority.key("tinvest:provider:used")
	if n := authority.client.ZCard(ctx, key).Val(); n != limit {
		t.Fatalf("bounded log size=%d", n)
	}
	if ttl := authority.client.PTTL(ctx, key).Val(); ttl <= 0 || ttl > window {
		t.Fatalf("TTL=%v", ttl)
	}
	time.Sleep(window + 100*time.Millisecond)
	if n := authority.client.Exists(ctx, key).Val(); n != 0 {
		t.Fatal("rolling key did not expire")
	}
}

func TestLegacyFixedWindowStateReservesFullRollingWindow(t *testing.T) {
	authority := testRedisAuthority(t)
	ctx := context.Background()
	key := authority.key("tinvest:provider:used")
	if err := authority.client.Set(ctx, key, 1, 200*time.Millisecond).Err(); err != nil {
		t.Fatal(err)
	}
	allowed, err := authority.AdmitProvider(ctx, 4, time.Second)
	if err != nil || allowed {
		t.Fatalf("legacy state allowed=%v err=%v", allowed, err)
	}
	time.Sleep(250 * time.Millisecond)
	allowed, err = authority.AdmitProvider(ctx, 4, time.Second)
	if err != nil || allowed {
		t.Fatalf("old fixed expiry reset rolling reservation: admitted=%v err=%v", allowed, err)
	}
	time.Sleep(850 * time.Millisecond)
	allowed, err = authority.AdmitProvider(ctx, 4, time.Second)
	if err != nil || !allowed {
		t.Fatalf("expired legacy state allowed=%v err=%v", allowed, err)
	}
}
