package httpapi

// Restores #225–228 transport/DTO assertions; exact source heads are retained in #280.
import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

func FuzzTransactionRouteUUID(f *testing.F) {
	for _, seed := range []string{
		"00000000-0000-4000-8000-000000000001", "550e8400-e29b-41d4-a716-446655440000",
		"", "not-a-uuid", "../../etc/passwd", "00000000000000000000000000000000",
		"00000000-0000-4000-8000-000000000002", "%00", "00000000000040008000000000000002",
		"00000000-0000-0000-0000-000000000000", "%2f", "\x00",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, raw string) {
		if len(raw) > 4096 {
			t.Skip("bounded parser domain; not historical unbounded-domain equivalence")
		}
		// Escaped target is at most 3*4096 bytes: size this synthetic transport to
		// admit the entire parser domain instead of confusing header-size rejection with parsing.
		app := fiber.New(fiber.Config{ReadBufferSize: 32 * 1024})
		handled := false
		app.Get("/fuzz/:id", func(c fiber.Ctx) error {
			handled = true
			value, err := transactionRouteUUID(c, "id")
			if err != nil {
				return c.SendStatus(http.StatusBadRequest)
			}
			if value == "" {
				t.Error("UUID helper accepted empty value")
			}
			if _, err := uuid.Parse(value); err != nil {
				t.Error("UUID helper accepted malformed value")
			}
			return c.SendStatus(http.StatusNoContent)
		})
		response, err := app.Test(httptest.NewRequest(http.MethodGet, "/fuzz/"+url.PathEscape(raw), nil))
		if err != nil {
			t.Fatalf("transport-valid bounded request failed: %v", err)
		}
		defer response.Body.Close()
		switch response.StatusCode {
		case 204, 400, 404:
		default:
			t.Fatalf("unexpected UUID status %d", response.StatusCode)
		}
		_, parseErr := uuid.Parse(raw)
		if parseErr == nil {
			if !handled || response.StatusCode != 204 {
				t.Fatalf("valid UUID positive control not accepted: status=%d handled=%v", response.StatusCode, handled)
			}
		} else if response.StatusCode == 204 {
			t.Fatal("invalid UUID accepted")
		}
	})
}

type nativeJSONCountActive struct {
	Name   string `json:"name"`
	Count  int    `json:"count"`
	Active bool   `json:"active"`
}
type nativeJSONUntagged struct {
	Name  string
	Count int
}
type nativeJSONCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}
type nativeJSONAmount struct {
	Name   string `json:"name"`
	Amount string `json:"amount"`
	Active bool   `json:"active"`
}

func FuzzDecodeStrictJSON(f *testing.F) {
	for _, seed := range []string{
		`{"name":"alice","count":1,"active":true}`, `{"name":"alice","unknown":1}`,
		`{"name":"alice"}{"name":"bob"}`, `null`, `{`, "",
		`{"Name":"alice","Count":1}`, `{"Name":"alice","Count":1,"unknown":true}`,
		`{"Name":"alice"} {"Count":1}`, `{"Name":`,
		`{"name":"portfolio","count":1}`, `{}`, `{"unknown":true}`, `{"name":"a"} {"name":"b"}`, `[]`,
		`{"name":"test","amount":"1.00000000","active":true}`, `{"name":"test","unknown":1}`, `{"name":"test"} null`,
	} {
		f.Add(seed)
	}
	profiles := []struct {
		newTarget func() any
		positive  string
		fields    []string
	}{
		{func() any { return &nativeJSONCountActive{} }, `{"name":"alice","count":1,"active":true}`, []string{"name", "count", "active"}},
		{func() any { return &nativeJSONUntagged{} }, `{"Name":"alice","Count":1}`, []string{"Name", "Count"}},
		{func() any { return &nativeJSONCount{} }, `{"name":"portfolio","count":1}`, []string{"name", "count"}},
		{func() any { return &nativeJSONAmount{} }, `{"name":"test","amount":"1.00000000","active":true}`, []string{"name", "amount", "active"}},
	}
	for _, p := range profiles {
		if err := decodeStrictJSON([]byte(p.positive), p.newTarget()); err != nil {
			f.Fatalf("DTO positive control failed: %v", err)
		}
		// null is accepted as a zero-value DTO by the current helper: do not invent null rejection.
		if err := decodeStrictJSON([]byte("null"), p.newTarget()); err != nil {
			f.Fatalf("null contract changed: %v", err)
		}
		for _, negative := range []string{"", " ", `{"name":`, `{"__openinvest_unknown":1}`, p.positive + " null", p.positive + " {}"} {
			if err := decodeStrictJSON([]byte(negative), p.newTarget()); err == nil {
				f.Fatalf("DTO negative control accepted %q", negative)
			}
		}
	}
	f.Fuzz(func(t *testing.T, body string) {
		if len(body) > 256*1024 {
			t.Skip("bounded JSON domain")
		}
		for _, p := range profiles {
			first, second := p.newTarget(), p.newTarget()
			e1, e2 := decodeStrictJSON([]byte(body), first), decodeStrictJSON([]byte(body), second)
			if (e1 == nil) != (e2 == nil) {
				t.Fatal("JSON success/error nondeterminism")
			}
			if e1 != nil {
				if e1.Error() != e2.Error() {
					t.Fatal("JSON error nondeterminism")
				}
				continue
			}
			if !json.Valid([]byte(body)) || !reflect.DeepEqual(first, second) {
				t.Fatal("JSON syntax/value invariant failed")
			}
			var raw map[string]json.RawMessage
			if err := json.Unmarshal([]byte(body), &raw); err != nil {
				t.Fatalf("struct DTO accepted nonobject: %v", err)
			}
			for key := range raw {
				known := false
				// encoding/json recognizes case-insensitive field matches. The old #227
				// literal-case whitelist was stricter than that helper contract; preserve
				// unknown-field rejection without introducing a new casing policy.
				for _, field := range p.fields {
					if strings.EqualFold(field, key) {
						known = true
						break
					}
				}
				if !known {
					t.Fatalf("unknown DTO field accepted: %q", key)
				}
			}
			if err := decodeStrictJSON(append([]byte(body), []byte(" null")...), p.newTarget()); err == nil {
				t.Fatal("trailing JSON value accepted")
			}
		}
	})
}
