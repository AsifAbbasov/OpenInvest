package httpapi

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestObservabilityErrorResponsesDoNotLeakInternalFailureDetails(t *testing.T) {
	app := fiber.New()
	app.Get("/boom", func(c fiber.Ctx) error {
		return writeMappedError(c, errors.New("postgres password=super-secret sql=SELECT * FROM identity.users"))
	})

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/boom", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	for _, forbidden := range []string{"super-secret", "SELECT *", "identity.users", "password="} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("internal detail leaked through observable error envelope: %q", forbidden)
		}
	}
	if resp.Header.Get("X-Request-ID") == "" || resp.Header.Get("X-Trace-ID") == "" {
		t.Fatal("error response omitted correlation identifiers")
	}
}

func TestObservabilityHealthAndReadinessRemainSemanticallyDistinct(t *testing.T) {
	store := &oiNew03ReadyStore{pingErr: errors.New("database unavailable")}
	app := oiNew03App(store)

	health, err := app.Test(httptest.NewRequest(http.MethodGet, "/api/v1/health", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer health.Body.Close()
	if health.StatusCode != http.StatusOK {
		t.Fatalf("liveness must remain up while dependency readiness is down: %d", health.StatusCode)
	}

	ready, err := app.Test(httptest.NewRequest(http.MethodGet, "/api/v1/ready", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer ready.Body.Close()
	if ready.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("readiness must expose dependency failure: %d", ready.StatusCode)
	}
}
