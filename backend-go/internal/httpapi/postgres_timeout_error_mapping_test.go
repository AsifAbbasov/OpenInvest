package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestPostgresTimeoutErrorsUseSanitizedFailClosedResponse(t *testing.T) {
	for _, sqlState := range []string{"57014", "55P03", "25P03"} {
		t.Run(sqlState, func(t *testing.T) {
			app := fiber.New()
			app.Get("/", func(c fiber.Ctx) error {
				return writeMappedError(c, &pgconn.PgError{Code: sqlState, Message: "database internals must not leak"})
			})

			response, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
			if err != nil {
				t.Fatalf("request: %v", err)
			}
			defer response.Body.Close()
			if response.StatusCode != http.StatusInternalServerError {
				t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusInternalServerError)
			}
			var payload struct {
				Error struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				} `json:"error"`
			}
			if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if payload.Error.Code != "INTERNAL_ERROR" || payload.Error.Message != "Internal server error" {
				t.Fatalf("timeout response = %+v, want sanitized internal error", payload.Error)
			}
		})
	}
}
