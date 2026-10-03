package httpapi

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func FuzzTransactionRouteUUID(f *testing.F) {
	for _, seed := range []string{
		"00000000-0000-4000-8000-000000000001",
		"550e8400-e29b-41d4-a716-446655440000",
		"",
		"not-a-uuid",
		"00000000-0000-0000-0000-000000000000",
		"../../etc/passwd",
		"%2f",
	} {
		f.Add(seed)
	}

	app := fiber.New()
	app.Get("/transactions/:transactionID", func(c fiber.Ctx) error {
		if _, err := transactionRouteUUID(c, "transactionID"); err != nil {
			return c.SendStatus(http.StatusBadRequest)
		}
		return c.SendStatus(http.StatusNoContent)
	})

	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > 4096 {
			return
		}
		request := httptest.NewRequest(
			http.MethodGet,
			"/transactions/"+url.PathEscape(input),
			nil,
		)
		response, err := app.Test(request)
		if err != nil {
			t.Fatalf("route processing failed: input=%q err=%v", input, err)
		}
		defer response.Body.Close()

		if response.StatusCode >= http.StatusInternalServerError {
			t.Fatalf("malformed UUID reached server error: input=%q status=%d", input, response.StatusCode)
		}
	})
}

func FuzzDecodeStrictJSON(f *testing.F) {
	type payload struct {
		Name   string `json:"name"`
		Amount string `json:"amount"`
		Active bool   `json:"active"`
	}

	for _, seed := range []string{
		`{"name":"test","amount":"1.00000000","active":true}`,
		`{"name":"test","unknown":1}`,
		`{"name":"test"} null`,
		`{}`,
		"",
		"{",
		"null",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > 256*1024 {
			return
		}

		var decoded payload
		err := decodeStrictJSON([]byte(input), &decoded)
		if err != nil {
			return
		}

		var second payload
		if err := decodeStrictJSON([]byte(input), &second); err != nil {
			t.Fatalf("strict JSON decode is non-deterministic: input=%q err=%v", input, err)
		}
		if decoded != second {
			t.Fatalf("strict JSON decode changed value across identical runs: first=%+v second=%+v", decoded, second)
		}

		withTrailing := append(append([]byte(nil), []byte(input)...), []byte(" null")...)
		var trailing payload
		if err := decodeStrictJSON(withTrailing, &trailing); err == nil {
			t.Fatalf("strict JSON accepted a trailing JSON value: input=%q", input)
		}
	})
}
