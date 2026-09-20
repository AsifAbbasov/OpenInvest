package httpapi

import (
	"context"
	"errors"
	"sync"

	"github.com/gofiber/fiber/v3"
)

// RequestLifecycle separates graceful HTTP draining from forced request
// cancellation. Requests admitted before ForceCancel keep running during the
// graceful phase. ForceCancel seals the lifecycle so no later request reaches
// application code and cancels every already-admitted request context.
type RequestLifecycle struct {
	forceCtx    context.Context
	forceCancel context.CancelFunc

	mu     sync.Mutex
	active int
	idle   chan struct{}
	forced bool
}

func NewRequestLifecycle() *RequestLifecycle {
	forceCtx, forceCancel := context.WithCancel(context.Background())
	idle := make(chan struct{})
	close(idle)

	return &RequestLifecycle{
		forceCtx:    forceCtx,
		forceCancel: forceCancel,
		idle:        idle,
	}
}

func (lifecycle *RequestLifecycle) begin() (func(), bool) {
	if lifecycle == nil {
		return func() {}, true
	}

	lifecycle.mu.Lock()
	defer lifecycle.mu.Unlock()

	if lifecycle.forced {
		return nil, false
	}

	if lifecycle.active == 0 {
		lifecycle.idle = make(chan struct{})
	}
	lifecycle.active++

	return func() {
		lifecycle.mu.Lock()
		defer lifecycle.mu.Unlock()

		lifecycle.active--
		if lifecycle.active == 0 {
			close(lifecycle.idle)
		}
	}, true
}

// Middleware installs a stable stdlib context into Fiber before application
// handlers run. Existing user context values/deadlines are preserved because
// the lifecycle context derives from c.Context().
func (lifecycle *RequestLifecycle) Middleware(c fiber.Ctx) error {
	if lifecycle == nil {
		return c.Next()
	}

	finish, admitted := lifecycle.begin()
	if !admitted {
		c.SetContext(lifecycle.forceCtx)
		return context.Canceled
	}
	defer finish()

	requestCtx, cancel := context.WithCancel(c.Context())
	stopForceCallback := context.AfterFunc(lifecycle.forceCtx, cancel)
	if lifecycle.forceCtx.Err() != nil {
		cancel()
	}

	c.SetContext(requestCtx)

	defer func() {
		stopForceCallback()
		cancel()
	}()

	return c.Next()
}

// ForceCancel starts the forced phase. It is idempotent.
//
// The mutex ordering is important: either a request is admitted before forced
// state is set and therefore becomes tracked, or it observes forced state and
// never enters downstream application handlers.
func (lifecycle *RequestLifecycle) ForceCancel() {
	if lifecycle == nil {
		return
	}

	lifecycle.mu.Lock()
	if lifecycle.forced {
		lifecycle.mu.Unlock()
		return
	}
	lifecycle.forced = true
	lifecycle.mu.Unlock()

	lifecycle.forceCancel()
}

// Wait blocks until every request admitted before ForceCancel has left the
// application handler chain. After ForceCancel no new downstream request can
// increase active again.
// Idle reports whether no request currently owns application runtime work.
func (lifecycle *RequestLifecycle) Idle() bool {
	if lifecycle == nil {
		return true
	}

	lifecycle.mu.Lock()
	defer lifecycle.mu.Unlock()

	return lifecycle.active == 0
}

func (lifecycle *RequestLifecycle) WaitContext(ctx context.Context) error {
	if lifecycle == nil {
		return nil
	}
	if ctx == nil {
		return errors.New("request lifecycle wait context is required")
	}

	for {
		lifecycle.mu.Lock()
		if lifecycle.active == 0 {
			lifecycle.mu.Unlock()
			return nil
		}
		idle := lifecycle.idle
		lifecycle.mu.Unlock()

		select {
		case <-idle:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (lifecycle *RequestLifecycle) Wait() {
	_ = lifecycle.WaitContext(context.Background())
}
