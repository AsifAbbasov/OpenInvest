package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
)

type requestLifecycleContextKey string

func TestRequestLifecyclePreservesExistingContextValueAndDeadline(t *testing.T) {
	lifecycle := NewRequestLifecycle()
	app := fiber.New()

	parentDeadline := time.Now().Add(30 * time.Second)
	parentCtx, cancelParent := context.WithDeadline(
		context.WithValue(
			context.Background(),
			requestLifecycleContextKey("existing"),
			"preserved",
		),
		parentDeadline,
	)
	defer cancelParent()

	app.Use(func(c fiber.Ctx) error {
		c.SetContext(parentCtx)
		return c.Next()
	})

	app.Use(lifecycle.Middleware)

	app.Get("/", func(c fiber.Ctx) error {
		if got := c.Context().Value(requestLifecycleContextKey("existing")); got != "preserved" {
			return fiber.ErrInternalServerError
		}

		gotDeadline, ok := c.Context().Deadline()
		if !ok {
			return fiber.ErrInternalServerError
		}

		if !gotDeadline.Equal(parentDeadline) {
			return fiber.ErrInternalServerError
		}

		if c.Context().Err() != nil {
			return fiber.ErrInternalServerError
		}

		return c.SendStatus(http.StatusNoContent)
	})

	response, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusNoContent {
		t.Fatalf("unexpected status %d", response.StatusCode)
	}
}

func TestRequestLifecycleForceCancelCancelsActiveRequest(t *testing.T) {
	lifecycle := NewRequestLifecycle()
	app := fiber.New()
	app.Use(lifecycle.Middleware)

	requestStarted := make(chan struct{})
	requestCanceled := make(chan struct{})

	app.Get("/", func(c fiber.Ctx) error {
		close(requestStarted)

		select {
		case <-c.Context().Done():
			close(requestCanceled)
			return c.Context().Err()
		case <-time.After(2 * time.Second):
			return fiber.ErrRequestTimeout
		}
	})

	requestDone := make(chan struct{})
	go func() {
		response, _ := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
		if response != nil {
			_ = response.Body.Close()
		}
		close(requestDone)
	}()

	select {
	case <-requestStarted:
	case <-time.After(time.Second):
		t.Fatal("request did not start")
	}

	select {
	case <-requestCanceled:
		t.Fatal("request was canceled before ForceCancel")
	default:
	}

	lifecycle.ForceCancel()

	select {
	case <-requestCanceled:
	case <-time.After(time.Second):
		t.Fatal("active request did not receive force cancellation")
	}

	lifecycle.Wait()

	select {
	case <-requestDone:
	case <-time.After(time.Second):
		t.Fatal("request handler did not exit after force cancellation")
	}
}
