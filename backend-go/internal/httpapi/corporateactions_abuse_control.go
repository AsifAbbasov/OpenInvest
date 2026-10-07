package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

const (
	// The public endpoint intentionally remains anonymous. The per-client budget is
	// materially below the provider's 60/minute process-local budget, while still
	// allowing normal UI refreshes and small shared-NAT bursts.
	defaultCorporateActionProjectionPerClientLimit = 12
	defaultCorporateActionProjectionGlobalLimit    = 48
	defaultCorporateActionProjectionMaxClientKeys  = 128
	defaultCorporateActionProjectionWindow         = time.Minute
	defaultCorporateActionProjectionMaxInflight    = 48

	corporateActionProjectionRateLimitRetryAfterSeconds = "60"
)

var errCorporateActionProjectionRateLimited = errors.New("corporate action projection rate limited")

type corporateActionProjectionCall struct {
	done      chan struct{}
	ctx       context.Context
	cancel    context.CancelFunc
	waiters   int
	completed bool
	result    corporateActionProjectionDTO
	err       error
}

type corporateActionProjectionCoalescer struct {
	mu       sync.Mutex
	maxCalls int
	calls    map[string]*corporateActionProjectionCall
}

func newCorporateActionProjectionRateLimiter() *authRateLimiter {
	return newBoundedAuthRateLimiter(
		defaultCorporateActionProjectionPerClientLimit,
		defaultCorporateActionProjectionGlobalLimit,
		defaultCorporateActionProjectionMaxClientKeys,
		defaultCorporateActionProjectionWindow,
	)
}

func newCorporateActionProjectionCoalescer() *corporateActionProjectionCoalescer {
	return &corporateActionProjectionCoalescer{
		maxCalls: defaultCorporateActionProjectionMaxInflight,
		calls:    make(map[string]*corporateActionProjectionCall),
	}
}

func (api *API) admitCorporateActionProjection(c fiber.Ctx) error {
	if api.corporateActionLimiter == nil {
		return nil
	}
	clientIP, err := normalizedClientIP(c)
	if err != nil {
		return err
	}
	key := c.Path() + "|" + clientIP
	if !api.corporateActionLimiter.allow(key, api.nowUTC()) {
		return errCorporateActionProjectionRateLimited
	}
	return nil
}

func corporateActionProjectionRequestKey(query verticalslice.CorporateActionQuery) string {
	material := strings.Join(query.InstrumentIDs, "\x00") + "\x00" + query.From + "\x00" + query.To
	sum := sha256.Sum256([]byte(material))
	return hex.EncodeToString(sum[:])
}

func (api *API) loadCorporateActionProjection(
	ctx context.Context,
	query verticalslice.CorporateActionQuery,
) (corporateActionProjectionDTO, error) {
	load := func(sharedCtx context.Context) (corporateActionProjectionDTO, error) {
		return fetchCorporateActionProjection(sharedCtx, api.corporateActionProvider, query)
	}
	if api.corporateActionCoalescer == nil {
		return load(ctx)
	}
	return api.corporateActionCoalescer.do(ctx, corporateActionProjectionRequestKey(query), load)
}

func fetchCorporateActionProjection(
	ctx context.Context,
	provider verticalslice.CorporateActionProvider,
	query verticalslice.CorporateActionQuery,
) (corporateActionProjectionDTO, error) {
	events, err := verticalslice.FetchCorporateActions(ctx, provider, query)
	if err != nil {
		return corporateActionProjectionDTO{}, err
	}
	calendar, err := verticalslice.ProjectCorporateActionCalendar(events)
	if err != nil {
		return corporateActionProjectionDTO{}, err
	}
	heatmap, err := verticalslice.ProjectCorporateActionHeatmap(events)
	if err != nil {
		return corporateActionProjectionDTO{}, err
	}
	calendar = filterCorporateActionCalendarWindow(calendar, query.From, query.To)
	heatmap = filterCorporateActionHeatmapWindow(heatmap, query.From, query.To)

	return corporateActionProjectionDTO{
		Calendar: mapCorporateActionCalendar(calendar),
		Heatmap:  mapCorporateActionHeatmap(heatmap),
		Coverage: corporateActionCoverageDTO{
			InputMode:     "PROVIDER",
			InstrumentIDs: append([]string(nil), query.InstrumentIDs...),
			From:          query.From,
			To:            query.To,
		},
	}, nil
}

func (coalescer *corporateActionProjectionCoalescer) do(
	ctx context.Context,
	key string,
	load func(context.Context) (corporateActionProjectionDTO, error),
) (corporateActionProjectionDTO, error) {
	if coalescer == nil {
		return load(ctx)
	}

	coalescer.mu.Lock()
	if call, ok := coalescer.calls[key]; ok {
		call.waiters++
		coalescer.mu.Unlock()
		return coalescer.wait(ctx, key, call)
	}
	if coalescer.maxCalls <= 0 || len(coalescer.calls) >= coalescer.maxCalls {
		coalescer.mu.Unlock()
		return corporateActionProjectionDTO{}, errCorporateActionProjectionRateLimited
	}

	sharedCtx, cancel := context.WithCancel(context.Background())
	call := &corporateActionProjectionCall{
		done:    make(chan struct{}),
		ctx:     sharedCtx,
		cancel:  cancel,
		waiters: 1,
	}
	coalescer.calls[key] = call
	coalescer.mu.Unlock()

	go func() {
		result, err := load(sharedCtx)

		coalescer.mu.Lock()
		call.result = result
		call.err = err
		call.completed = true
		if current, ok := coalescer.calls[key]; ok && current == call {
			delete(coalescer.calls, key)
		}
		close(call.done)
		coalescer.mu.Unlock()
		cancel()
	}()

	return coalescer.wait(ctx, key, call)
}

func (coalescer *corporateActionProjectionCoalescer) wait(
	ctx context.Context,
	key string,
	call *corporateActionProjectionCall,
) (corporateActionProjectionDTO, error) {
	select {
	case <-call.done:
		return call.result, call.err
	case <-ctx.Done():
		var cancel context.CancelFunc
		coalescer.mu.Lock()
		if !call.completed {
			if current, ok := coalescer.calls[key]; ok && current == call {
				if call.waiters > 0 {
					call.waiters--
				}
				if call.waiters == 0 {
					cancel = call.cancel
				}
			}
		}
		coalescer.mu.Unlock()
		if cancel != nil {
			cancel()
		}
		return corporateActionProjectionDTO{}, ctx.Err()
	}
}

func (coalescer *corporateActionProjectionCoalescer) activeCalls() int {
	if coalescer == nil {
		return 0
	}
	coalescer.mu.Lock()
	defer coalescer.mu.Unlock()
	return len(coalescer.calls)
}
