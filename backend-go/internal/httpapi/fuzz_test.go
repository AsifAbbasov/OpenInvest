package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func FuzzTransactionRouteUUID(f *testing.F) {
	for _, seed := range []string{
		"00000000-0000-4000-8000-000000000002",
		"550e8400-e29b-41d4-a716-446655440000",
		"",
		"not-a-uuid",
		"../../etc/passwd",
		"%00",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, value string) {
		app := fiber.New()
		app.Get("/fuzz/:id", func(c fiber.Ctx) error {
			got, err := transactionRouteUUID(c, "id")
			if err != nil {
				return c.SendStatus(http.StatusBadRequest)
			}
			if got == "" {
				t.Fatalf("UUID route helper accepted an empty value")
			}
			return c.SendStatus(http.StatusNoContent)
		})

		target := "/fuzz/" + url.PathEscape(value)
		response, err := app.Test(httptest.NewRequest(http.MethodGet, target, nil))
		if err != nil {
			return
		}
		if response.StatusCode >= http.StatusInternalServerError {
			t.Fatalf("route UUID input caused server error: value=%q status=%d", value, response.StatusCode)
		}
	})
}

func FuzzDecodeStrictJSON(f *testing.F) {
	for _, seed := range [][]byte{
		[]byte("{\"Name\":\"alice\",\"Count\":1}"),
		[]byte("{\"Name\":\"alice\",\"Count\":1,\"unknown\":true}"),
		[]byte("{\"Name\":\"alice\"} {\"Count\":1}"),
		[]byte("null"),
		[]byte(""),
		[]byte("{\"Name\":"),
	} {
		f.Add(seed)
	}

	type fuzzDTO struct {
		Name  string
		Count int
	}

	f.Fuzz(func(t *testing.T, body []byte) {
		var target fuzzDTO
		err := decodeStrictJSON(body, &target)
		if err != nil {
			return
		}
		if !json.Valid(body) {
			t.Fatalf("decodeStrictJSON accepted syntactically invalid JSON: %q", body)
		}
	})
}
