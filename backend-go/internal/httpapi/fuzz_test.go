package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func FuzzDecodeStrictJSON(f *testing.F) {
	for _, seed := range [][]byte{
		[]byte(`{"name":"portfolio","count":1}`),
		[]byte(`{}`),
		[]byte(`{"unknown":true}`),
		[]byte(`{"name":"a"} {"name":"b"}`),
		[]byte(`null`),
		[]byte(`[]`),
		{},
	} {
		f.Add(seed)
	}

	type targetDTO struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	f.Fuzz(func(t *testing.T, body []byte) {
		if len(body) > 256*1024 {
			t.Skip()
		}

		var target targetDTO
		err := decodeStrictJSON(body, &target)
		if err != nil {
			return
		}

		trimmed := bytes.TrimSpace(body)
		if !json.Valid(trimmed) {
			t.Fatalf("decodeStrictJSON accepted invalid JSON: %q", body)
		}

		// A successful decode must not have admitted an unknown top-level object key.
		// decodeStrictJSON is the production helper that enables DisallowUnknownFields.
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(trimmed, &raw); err == nil && raw != nil {
			for key := range raw {
				if key != "name" && key != "count" {
					t.Fatalf("decodeStrictJSON admitted unknown field %q in %q", key, body)
				}
			}
		}

		var repeated targetDTO
		if err2 := decodeStrictJSON(body, &repeated); err2 != nil {
			t.Fatalf("decodeStrictJSON is non-deterministic: first accepted, second rejected: %v", err2)
		}
		if target != repeated {
			t.Fatalf("decodeStrictJSON changed decoded value across identical runs: first=%+v second=%+v", target, repeated)
		}
	})
}

func FuzzTransactionRouteUUID(f *testing.F) {
	for _, seed := range []string{
		"00000000-0000-4000-8000-000000000002",
		"not-a-uuid",
		"",
		"../../etc/passwd",
		"%00",
		"00000000000040008000000000000002",
	} {
		f.Add(seed)
	}

	app := fiber.New()
	app.Get("/fuzz/:id", func(c fiber.Ctx) error {
		if _, err := transactionRouteUUID(c, "id"); err != nil {
			return c.SendStatus(http.StatusBadRequest)
		}
		return c.SendStatus(http.StatusNoContent)
	})

	f.Fuzz(func(t *testing.T, raw string) {
		if len(raw) > 4096 {
			t.Skip()
		}

		path := "/fuzz/" + url.PathEscape(raw)
		request := httptest.NewRequest(http.MethodGet, path, nil)
		response, err := app.Test(request)
		if err != nil {
			t.Fatalf("Fiber route test failed for %q: %v", raw, err)
		}
		defer response.Body.Close()

		if response.StatusCode >= 500 {
			t.Fatalf("malformed route UUID reached a server error: input=%q status=%d", raw, response.StatusCode)
		}
		if response.StatusCode != http.StatusNoContent &&
			response.StatusCode != http.StatusBadRequest &&
			response.StatusCode != http.StatusNotFound {
			t.Fatalf("unexpected route UUID status: input=%q status=%d", raw, response.StatusCode)
		}
	})
}
