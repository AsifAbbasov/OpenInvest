package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
)

func testApp(t *testing.T) *fiber.App {
	t.Helper()

	runtime, err := newRuntime()
	if err != nil {
		t.Fatalf("initialize test runtime: %v", err)
	}
	t.Cleanup(func() {
		if err := runtime.Close(); err != nil {
			t.Errorf("close test runtime: %v", err)
		}
	})
	return runtime.app
}

func waitForSignal(t *testing.T, signal <-chan struct{}, timeout time.Duration, name string) {
	t.Helper()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-signal:
	case <-timer.C:
		t.Fatalf("timed out waiting for %s", name)
	}
}

func waitForError(t *testing.T, result <-chan error, timeout time.Duration, name string) error {
	t.Helper()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case err := <-result:
		return err
	case <-timer.C:
		t.Fatalf("timed out waiting for %s", name)
		return nil
	}
}

func TestOINew09GracefulShutdownDrainsInFlightRequest(t *testing.T) {
	app := fiber.New()

	requestStarted := make(chan struct{})
	releaseRequest := make(chan struct{})
	shutdownStarted := make(chan struct{})

	app.Hooks().OnPreShutdown(func() error {
		close(shutdownStarted)
		return nil
	})

	app.Get("/slow", func(c fiber.Ctx) error {
		close(requestStarted)
		<-releaseRequest
		return c.SendStatus(http.StatusNoContent)
	})

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("create listener: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serveDone := make(chan error, 1)
	go func() {
		serveDone <- serveHTTP(ctx, app, listener, time.Second)
	}()

	clientDone := make(chan error, 1)
	go func() {
		client := &http.Client{Timeout: 2 * time.Second}
		response, err := client.Get("http://" + listener.Addr().String() + "/slow")
		if err != nil {
			clientDone <- err
			return
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusNoContent {
			clientDone <- fmt.Errorf("unexpected response status %d", response.StatusCode)
			return
		}
		clientDone <- nil
	}()

	waitForSignal(t, requestStarted, 2*time.Second, "in-flight request")
	cancel()
	waitForSignal(t, shutdownStarted, 2*time.Second, "Fiber shutdown start")

	select {
	case err := <-serveDone:
		t.Fatalf("server returned before the in-flight request drained: %v", err)
	default:
	}

	close(releaseRequest)

	if err := waitForError(t, clientDone, 2*time.Second, "client response"); err != nil {
		t.Fatalf("client request failed during graceful drain: %v", err)
	}
	if err := waitForError(t, serveDone, 2*time.Second, "server shutdown"); err != nil {
		t.Fatalf("graceful shutdown failed: %v", err)
	}
}

type oiNew09DeadlineShutdownServer struct {
	sawDeadline bool
	remaining   time.Duration
}

func (server *oiNew09DeadlineShutdownServer) ShutdownWithContext(ctx context.Context) error {
	deadline, ok := ctx.Deadline()
	if !ok {
		return errors.New("shutdown context has no deadline")
	}

	server.sawDeadline = true
	server.remaining = time.Until(deadline)

	<-ctx.Done()
	return ctx.Err()
}

func TestOINew09ShutdownIsBoundedByDeadline(t *testing.T) {
	server := &oiNew09DeadlineShutdownServer{}

	started := time.Now()
	err := shutdownHTTP(server, 20*time.Millisecond)
	elapsed := time.Since(started)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
	if !server.sawDeadline {
		t.Fatal("shutdown did not receive a deadline")
	}
	if server.remaining <= 0 || server.remaining > 100*time.Millisecond {
		t.Fatalf("unexpected shutdown deadline window: %s", server.remaining)
	}
	if elapsed > time.Second {
		t.Fatalf("bounded shutdown took unexpectedly long: %s", elapsed)
	}
}

type oiNew09ImmediateShutdownErrorServer struct {
	err error
}

func (server oiNew09ImmediateShutdownErrorServer) ShutdownWithContext(context.Context) error {
	return server.err
}

func TestOINew09ShutdownErrorPropagates(t *testing.T) {
	shutdownErr := errors.New("shutdown failure")

	err := shutdownHTTP(
		oiNew09ImmediateShutdownErrorServer{err: shutdownErr},
		time.Second,
	)

	if !errors.Is(err, shutdownErr) {
		t.Fatalf("expected shutdown error to propagate, got %v", err)
	}
}

func TestOINew09RuntimeCleanupRunsWhenServerStartupFails(t *testing.T) {
	startupErr := errors.New("listen failure")
	closeCalls := 0

	runtime := &applicationRuntime{
		app: fiber.New(),
		close: func() error {
			closeCalls++
			return nil
		},
	}

	err := runApplication(
		context.Background(),
		func() (*applicationRuntime, error) {
			return runtime, nil
		},
		func(context.Context, *fiber.App) error {
			return startupErr
		},
	)

	if !errors.Is(err, startupErr) {
		t.Fatalf("expected startup failure to propagate, got %v", err)
	}
	if closeCalls != 1 {
		t.Fatalf("expected runtime resources to close exactly once, got %d", closeCalls)
	}
}

func TestOINew09InitializationFailureDoesNotPretendSuccess(t *testing.T) {
	initializationErr := errors.New("initialization failure")
	serveCalled := false

	err := runApplication(
		context.Background(),
		func() (*applicationRuntime, error) {
			return nil, initializationErr
		},
		func(context.Context, *fiber.App) error {
			serveCalled = true
			return nil
		},
	)

	if !errors.Is(err, initializationErr) {
		t.Fatalf("expected initialization failure to propagate, got %v", err)
	}
	if serveCalled {
		t.Fatal("server must not start after initialization failure")
	}
}

func TestOINew09RuntimeCloseErrorPropagates(t *testing.T) {
	closeErr := errors.New("close failure")

	err := runApplication(
		context.Background(),
		func() (*applicationRuntime, error) {
			return &applicationRuntime{
				app: fiber.New(),
				close: func() error {
					return closeErr
				},
			}, nil
		},
		func(context.Context, *fiber.App) error {
			return nil
		},
	)

	if !errors.Is(err, closeErr) {
		t.Fatalf("expected runtime close error to propagate, got %v", err)
	}
}
