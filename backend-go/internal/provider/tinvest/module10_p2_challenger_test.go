package tinvest

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/openinvest/openinvest/backend-go/internal/httpapi"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

type module10ChallengerClock struct{ now time.Time }

func (c module10ChallengerClock) Now() time.Time { return c.now }

func module10ChallengerQuery() verticalslice.CorporateActionQuery {
	return verticalslice.CorporateActionQuery{
		InstrumentIDs: []string{"SBER"},
		From:          "2026-01-01",
		To:            "2026-12-31",
	}
}

func module10ChallengerValidDividend(units int, recordDate string) string {
	return fmt.Sprintf(
		`{"dividendNet":{"currency":"rub","units":"%d","nano":0},"paymentDate":"%s","recordDate":"%s","dividendType":"Regular Cash"}`,
		units,
		recordDate,
		recordDate,
	)
}

func module10ChallengerProvider(t *testing.T, h http.Handler) (*Provider, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(h)
	provider, err := newCorporateActionProvider(
		server.Client(),
		module10ChallengerClock{now: time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC)},
		"challenger-readonly-token",
		server.URL+"/rest",
	)
	if err != nil {
		server.Close()
		t.Fatal(err)
	}
	return provider, server
}

func module10ChallengerApp(t *testing.T, provider *Provider) *fiber.App {
	t.Helper()
	return httpapi.NewDevelopmentReplayWithCorporateActionProvider(nil, provider)
}

func module10ChallengerRequest(t *testing.T, app *fiber.App) (int, string) {
	t.Helper()
	target := "/api/v1/corporate-actions/projection?instrumentId=SBER&from=2026-01-01&to=2026-12-31"
	response, err := app.Test(httptest.NewRequest(http.MethodGet, target, nil))
	if err != nil {
		t.Fatalf("app.Test(): %v", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	return response.StatusCode, string(body)
}

func TestModule10ChallengerRoutedBudgetExhaustion(t *testing.T) {
	const successBody = `{"dividends":[{"dividendNet":{"currency":"rub","units":"1","nano":0},"paymentDate":"2026-06-01T00:00:00Z","recordDate":"2026-06-01T00:00:00Z","dividendType":"Regular Cash"}]}`

	var upstream atomic.Int64
	provider, server := module10ChallengerProvider(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstream.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, successBody)
	}))
	defer func() {
		server.CloseClientConnections()
		server.Close()
	}()

	now := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	provider.requestGate = newRequestBudget(maxRequestsPerMinute, time.Minute, func() time.Time { return now })
	app := module10ChallengerApp(t, provider)

	statusCounts := map[int]int{}
	for i := 0; i < maxRequestsPerMinute; i++ {
		status, _ := module10ChallengerRequest(t, app)
		statusCounts[status]++
		if status != http.StatusOK {
			t.Fatalf("anonymous routed request %d status=%d, want 200", i+1, status)
		}
	}
	if got := upstream.Load(); got != int64(maxRequestsPerMinute) {
		t.Fatalf("provider calls after attacker budget fill=%d want=%d", got, maxRequestsPerMinute)
	}

	beforeLegitimate := upstream.Load()
	legitimateStatus, _ := module10ChallengerRequest(t, app)
	legitimateProviderCall := upstream.Load() != beforeLegitimate
	if legitimateStatus != http.StatusServiceUnavailable {
		t.Fatalf("legitimate request after exhaustion status=%d want=503", legitimateStatus)
	}
	if legitimateProviderCall {
		t.Fatal("provider call occurred after local provider budget was exhausted")
	}

	now = now.Add(time.Minute)
	recoveryStatus, _ := module10ChallengerRequest(t, app)
	if recoveryStatus != http.StatusOK {
		t.Fatalf("request did not recover after budget window status=%d", recoveryStatus)
	}
	if upstream.Load() != int64(maxRequestsPerMinute+1) {
		t.Fatalf("recovery provider calls=%d want=%d", upstream.Load(), maxRequestsPerMinute+1)
	}

	t.Logf(
		"MODULE10_CHALLENGER_ROUTED sequential_anonymous_success=%d provider_calls_before_exhaustion=%d legitimate_after_exhaustion_status=%d legitimate_provider_call=%t recovery_status=%d recovery_window_seconds=60 recovery_requires_restart=false auth_required=false csrf_required=false outer_rate_limiter=false provider_rate_limit=%d provider_concurrency_limit=%d",
		statusCounts[http.StatusOK],
		maxRequestsPerMinute,
		legitimateStatus,
		legitimateProviderCall,
		recoveryStatus,
		maxRequestsPerMinute,
		maxConcurrency,
	)

	t.Run("concurrent_valid_requests", func(t *testing.T) {
		var concurrentUpstream atomic.Int64
		release := make(chan struct{})
		concurrentProvider, concurrentServer := module10ChallengerProvider(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			concurrentUpstream.Add(1)
			select {
			case <-release:
			case <-r.Context().Done():
				return
			}
			_, _ = io.WriteString(w, successBody)
		}))
		defer func() {
			concurrentServer.CloseClientConnections()
			concurrentServer.Close()
		}()
		concurrentProvider.requestGate = newRequestBudget(maxRequestsPerMinute, time.Minute, func() time.Time { return now })
		concurrentApp := module10ChallengerApp(t, concurrentProvider)

		const clients = 10
		statuses := make(chan int, clients)
		var wg sync.WaitGroup
		wg.Add(clients)
		for i := 0; i < clients; i++ {
			go func() {
				defer wg.Done()
				target := "/api/v1/corporate-actions/projection?instrumentId=SBER&from=2026-01-01&to=2026-12-31"
				response, err := concurrentApp.Test(httptest.NewRequest(http.MethodGet, target, nil))
				if err != nil {
					statuses <- 0
					return
				}
				_, _ = io.Copy(io.Discard, response.Body)
				_ = response.Body.Close()
				statuses <- response.StatusCode
			}()
		}

		deadline := time.Now().Add(2 * time.Second)
		for concurrentUpstream.Load() < int64(maxConcurrency) && time.Now().Before(deadline) {
			time.Sleep(5 * time.Millisecond)
		}
		if concurrentUpstream.Load() != int64(maxConcurrency) {
			close(release)
			wg.Wait()
			t.Fatalf("concurrent upstream reached=%d want=%d", concurrentUpstream.Load(), maxConcurrency)
		}
		close(release)
		wg.Wait()
		close(statuses)

		okCount := 0
		unavailableCount := 0
		for status := range statuses {
			switch status {
			case http.StatusOK:
				okCount++
			case http.StatusServiceUnavailable:
				unavailableCount++
			default:
				t.Fatalf("unexpected concurrent status=%d", status)
			}
		}
		if okCount != maxConcurrency || unavailableCount != clients-maxConcurrency {
			t.Fatalf("concurrent statuses 200=%d 503=%d", okCount, unavailableCount)
		}
		t.Logf("MODULE10_CHALLENGER_CONCURRENT clients=%d status_200=%d status_503=%d upstream=%d provider_concurrency_limit=%d", clients, okCount, unavailableCount, concurrentUpstream.Load(), maxConcurrency)
	})
}

func TestModule10ChallengerDistributedBudgetThroughRoutedApps(t *testing.T) {
	for _, instances := range []int{1, 2, 4} {
		t.Run(fmt.Sprintf("instances_%d", instances), func(t *testing.T) {
			var upstream atomic.Int64
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				upstream.Add(1)
				_, _ = io.WriteString(w, `{"dividends":[]}`)
			}))
			defer func() {
				server.CloseClientConnections()
				server.Close()
			}()

			now := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
			apps := make([]*fiber.App, 0, instances)
			for i := 0; i < instances; i++ {
				provider, err := newCorporateActionProvider(
					server.Client(),
					module10ChallengerClock{now: now},
					"challenger-readonly-token",
					server.URL+"/rest",
				)
				if err != nil {
					t.Fatal(err)
				}
				provider.requestGate = newRequestBudget(maxRequestsPerMinute, time.Minute, func() time.Time { return now })
				apps = append(apps, module10ChallengerApp(t, provider))
			}

			for _, app := range apps {
				for n := 0; n < maxRequestsPerMinute+5; n++ {
					status, _ := module10ChallengerRequest(t, app)
					if n < maxRequestsPerMinute && status != http.StatusOK {
						t.Fatalf("instance admitted request %d status=%d", n+1, status)
					}
					if n >= maxRequestsPerMinute && status != http.StatusServiceUnavailable {
						t.Fatalf("instance over-budget request %d status=%d", n+1, status)
					}
				}
			}

			want := int64(instances * maxRequestsPerMinute)
			if got := upstream.Load(); got != want {
				t.Fatalf("aggregate routed upstream calls=%d want=%d", got, want)
			}
			t.Logf("MODULE10_CHALLENGER_DISTRIBUTED instances=%d observed_upstream=%d nominal_per_process=%d", instances, upstream.Load(), maxRequestsPerMinute)
		})
	}
}

func module10ChallengerEventsFromBody(t *testing.T, body string) ([]verticalslice.CorporateActionEvent, int, string) {
	t.Helper()
	provider, server := module10ChallengerProvider(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, body)
	}))
	defer func() {
		server.CloseClientConnections()
		server.Close()
	}()

	events, providerErr := provider.CorporateActions(context.Background(), module10ChallengerQuery())
	app := module10ChallengerApp(t, provider)
	status, projectionBody := module10ChallengerRequest(t, app)
	if providerErr != nil {
		if status == http.StatusOK {
			t.Fatalf("provider rejected duplicate body but routed projection accepted it: %v", providerErr)
		}
		return nil, status, projectionBody
	}
	if status != http.StatusOK {
		t.Fatalf("provider accepted body but routed projection status=%d body=%s", status, projectionBody)
	}
	return events, status, projectionBody
}

func module10ChallengerArray(items ...string) string {
	return "[" + strings.Join(items, ",") + "]"
}

func TestModule10ChallengerConflictingDuplicateKeys(t *testing.T) {
	validA := module10ChallengerValidDividend(1, "2026-06-01T00:00:00Z")
	validB := module10ChallengerValidDividend(2, "2026-07-01T00:00:00Z")
	invalid := `{"dividendNet":{"currency":"rub","units":"-1","nano":0},"paymentDate":"2026-08-01T00:00:00Z","recordDate":"2026-08-01T00:00:00Z","dividendType":"Regular Cash"}`

	bodyAB := fmt.Sprintf(`{"dividends":%s,"dividends":%s}`, module10ChallengerArray(validA), module10ChallengerArray(validB))
	eventsAB, statusAB, projectionAB := module10ChallengerEventsFromBody(t, bodyAB)
	if statusAB != http.StatusOK || len(eventsAB) != 1 || eventsAB[0].AmountPerUnit == nil || eventsAB[0].AmountPerUnit.Amount.String() != "2.00000000" {
		t.Fatalf("A->B duplicate semantics unexpected status=%d events=%v", statusAB, eventsAB)
	}

	bodyBA := fmt.Sprintf(`{"dividends":%s,"dividends":%s}`, module10ChallengerArray(validB), module10ChallengerArray(validA))
	eventsBA, statusBA, projectionBA := module10ChallengerEventsFromBody(t, bodyBA)
	if statusBA != http.StatusOK || len(eventsBA) != 1 || eventsBA[0].AmountPerUnit == nil || eventsBA[0].AmountPerUnit.Amount.String() != "1.00000000" {
		t.Fatalf("B->A duplicate semantics unexpected status=%d events=%v", statusBA, eventsBA)
	}
	if projectionAB == projectionBA {
		t.Fatal("duplicate-key order did not alter routed projection output")
	}

	validThenInvalid := fmt.Sprintf(`{"dividends":%s,"dividends":%s}`, module10ChallengerArray(validA), module10ChallengerArray(invalid))
	_, statusValidInvalid, _ := module10ChallengerEventsFromBody(t, validThenInvalid)
	if statusValidInvalid != http.StatusBadGateway {
		t.Fatalf("valid->invalid duplicate status=%d want=502", statusValidInvalid)
	}

	invalidThenValid := fmt.Sprintf(`{"dividends":%s,"dividends":%s}`, module10ChallengerArray(invalid), module10ChallengerArray(validA))
	eventsInvalidValid, statusInvalidValid, _ := module10ChallengerEventsFromBody(t, invalidThenValid)
	if statusInvalidValid != http.StatusOK || len(eventsInvalidValid) != 1 {
		t.Fatalf("invalid->valid duplicate status=%d events=%d", statusInvalidValid, len(eventsInvalidValid))
	}

	largeItems := make([]string, 0, 100)
	for i := 1; i <= 100; i++ {
		largeItems = append(largeItems, module10ChallengerValidDividend(i, "2026-09-01T00:00:00Z"))
	}
	large := module10ChallengerArray(largeItems...)
	small := module10ChallengerArray(validA)

	largeThenSmall := fmt.Sprintf(`{"dividends":%s,"dividends":%s}`, large, small)
	eventsLargeSmall, statusLargeSmall, _ := module10ChallengerEventsFromBody(t, largeThenSmall)
	if statusLargeSmall != http.StatusOK || len(eventsLargeSmall) != 1 {
		t.Fatalf("large->small status=%d events=%d", statusLargeSmall, len(eventsLargeSmall))
	}

	smallThenLarge := fmt.Sprintf(`{"dividends":%s,"dividends":%s}`, small, large)
	eventsSmallLarge, statusSmallLarge, _ := module10ChallengerEventsFromBody(t, smallThenLarge)
	if statusSmallLarge != http.StatusOK || len(eventsSmallLarge) != 100 {
		t.Fatalf("small->large status=%d events=%d", statusSmallLarge, len(eventsSmallLarge))
	}

	t.Logf(
		"MODULE10_CHALLENGER_DUPLICATES accepted=true semantics=LAST_VALUE_WINS order_changes_accepted_result=true order_changes_projection=true valid_then_invalid_status=%d invalid_then_valid_status=%d large_then_small_events=%d small_then_large_events=%d persistent_financial_write=false",
		statusValidInvalid,
		statusInvalidValid,
		len(eventsLargeSmall),
		len(eventsSmallLarge),
	)
}

func TestModule10ChallengerSlowProviderRawTCP(t *testing.T) {
	cases := []struct {
		name          string
		partialBody   bool
		manualCancel  bool
		expectedError error
	}{
		{name: "never_headers", expectedError: context.DeadlineExceeded},
		{name: "partial_body", partialBody: true, expectedError: context.DeadlineExceeded},
		{name: "caller_cancel", manualCancel: true, expectedError: context.Canceled},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			defer listener.Close()

			requestSeen := make(chan struct{})
			connectionClosed := make(chan error, 1)
			recoveryServed := make(chan error, 1)

			go func() {
				conn, acceptErr := listener.Accept()
				if acceptErr != nil {
					connectionClosed <- acceptErr
					return
				}
				reader := bufio.NewReader(conn)
				request, readErr := http.ReadRequest(reader)
				if readErr != nil {
					_ = conn.Close()
					connectionClosed <- readErr
					return
				}
				if request.Body != nil {
					_, _ = io.Copy(io.Discard, request.Body)
					_ = request.Body.Close()
				}
				close(requestSeen)
				if tc.partialBody {
					_, _ = io.WriteString(conn, "HTTP/1.1 200 OK\r\nContent-Length: 1000\r\n\r\n{\"dividends\":[")
				}
				_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
				_, readErr = reader.ReadByte()
				_ = conn.Close()
				connectionClosed <- readErr

				recoveryConn, acceptErr := listener.Accept()
				if acceptErr != nil {
					recoveryServed <- acceptErr
					return
				}
				defer recoveryConn.Close()
				recoveryReader := bufio.NewReader(recoveryConn)
				recoveryRequest, readErr := http.ReadRequest(recoveryReader)
				if readErr != nil {
					recoveryServed <- readErr
					return
				}
				if recoveryRequest.Body != nil {
					_, _ = io.Copy(io.Discard, recoveryRequest.Body)
					_ = recoveryRequest.Body.Close()
				}
				body := `{"dividends":[]}`
				_, writeErr := fmt.Fprintf(recoveryConn, "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: %d\r\nConnection: close\r\n\r\n%s", len(body), body)
				recoveryServed <- writeErr
			}()

			provider, err := newCorporateActionProvider(
				&http.Client{},
				module10ChallengerClock{now: time.Now().UTC()},
				"challenger-readonly-token",
				"http://"+listener.Addr().String()+"/rest",
			)
			if err != nil {
				t.Fatal(err)
			}

			var ctx context.Context
			var cancel context.CancelFunc
			if tc.manualCancel {
				ctx, cancel = context.WithCancel(context.Background())
			} else {
				ctx, cancel = context.WithTimeout(context.Background(), 120*time.Millisecond)
			}
			defer cancel()

			callResult := make(chan error, 1)
			go func() {
				_, callErr := provider.CorporateActions(ctx, module10ChallengerQuery())
				callResult <- callErr
			}()

			select {
			case <-requestSeen:
			case <-time.After(time.Second):
				t.Fatal("upstream did not observe request")
			}
			if tc.manualCancel {
				cancel()
			}

			var callErr error
			select {
			case callErr = <-callResult:
			case <-time.After(2 * time.Second):
				t.Fatal("provider call did not return")
			}
			if !errors.Is(callErr, tc.expectedError) {
				t.Fatalf("call error=%v want=%v", callErr, tc.expectedError)
			}
			if len(provider.semaphore) != 0 {
				t.Fatal("provider concurrency permit leaked")
			}

			select {
			case readErr := <-connectionClosed:
				if readErr == nil {
					t.Fatal("upstream connection unexpectedly remained readable")
				}
				if networkErr, ok := readErr.(net.Error); ok && networkErr.Timeout() {
					t.Fatalf("upstream connection stayed open after cancellation: %v", readErr)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("raw TCP close witness timed out")
			}

			if _, err := provider.CorporateActions(context.Background(), module10ChallengerQuery()); err != nil {
				t.Fatalf("provider did not recover after cancellation: %v", err)
			}
			select {
			case serveErr := <-recoveryServed:
				if serveErr != nil {
					t.Fatalf("recovery server error: %v", serveErr)
				}
			case <-time.After(time.Second):
				t.Fatal("recovery request was not served")
			}

			t.Logf("MODULE10_CHALLENGER_SLOW case=%s caller_result=%v upstream_connection_closed=true provider_slot_released=true recovery=true", tc.name, tc.expectedError)
		})
	}
}
