package httpapi

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

func FuzzTransactionRouteUUID(f *testing.F) {
	for _, seed := range []string{
		"00000000-0000-4000-8000-000000000001",
		"550e8400-e29b-41d4-a716-446655440000",
		"",
		"not-a-uuid",
		"../../etc/passwd",
		"00000000000000000000000000000000",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, raw string) {
		if len(raw) > 4096 {
			t.Skip()
		}

		app := fiber.New()
		app.Get("/fuzz/:id", func(c fiber.Ctx) error {
			value, err := transactionRouteUUID(c, "id")
			if err != nil {
				return c.SendStatus(http.StatusBadRequest)
			}
			if _, err := uuid.Parse(value); err != nil {
				return c.SendStatus(http.StatusInternalServerError)
			}
			return c.SendStatus(http.StatusNoContent)
		})

		request := httptest.NewRequest(http.MethodGet, "/fuzz/"+url.PathEscape(raw), nil)
		response, err := app.Test(request)
		if err != nil {
			return
		}
		defer response.Body.Close()

		if response.StatusCode >= 500 {
			t.Fatalf("route UUID parser produced server error for input %q: status=%d", raw, response.StatusCode)
		}
	})
}

func FuzzDecodeStrictJSON(f *testing.F) {
	type fuzzDTO struct {
		Name   string `json:"name"`
		Count  int    `json:"count"`
		Active bool   `json:"active"`
	}

	for _, seed := range []string{
		`{"name":"alice","count":1,"active":true}`,
		`{"name":"alice","unknown":1}`,
		`{"name":"alice"}{"name":"bob"}`,
		`null`,
		`{`,
		"",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, body string) {
		if len(body) > 256*1024 {
			t.Skip()
		}

		var first fuzzDTO
		firstErr := decodeStrictJSON([]byte(body), &first)

		var second fuzzDTO
		secondErr := decodeStrictJSON([]byte(body), &second)

		if (firstErr == nil) != (secondErr == nil) {
			t.Fatalf("strict JSON decoding is non-deterministic: first=%v second=%v body=%q", firstErr, secondErr, body)
		}
		if firstErr == nil && !reflect.DeepEqual(first, second) {
			t.Fatalf("strict JSON decoding changed value between identical runs: first=%+v second=%+v", first, second)
		}
	})
}
