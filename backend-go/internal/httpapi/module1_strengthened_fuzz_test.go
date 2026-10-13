package httpapi

import (
	"strings"
	"testing"
)

func FuzzModule1BearerParserProperties(f *testing.F) {
	for _, seed := range []string{
		"",
		"Bearer ",
		"Bearer synthetic",
		"bearer synthetic",
		"Basic synthetic",
		"Bearer  synthetic ",
		"Bearer\tsynthetic",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, header string) {
		if len(header) > 8192 {
			return
		}
		got := bearerToken(header)
		if got == "" {
			return
		}
		if !strings.HasPrefix(header, "Bearer ") {
			t.Fatalf("parser returned token without exact Bearer prefix: header=%q token=%q", header, got)
		}
		if got != strings.TrimSpace(strings.TrimPrefix(header, "Bearer ")) {
			t.Fatalf("bearer parser result drift: header=%q got=%q", header, got)
		}
	})
}

func FuzzModule1StrictAuthJSONDecoderNeverPanics(f *testing.F) {
	for _, seed := range []string{
		"",
		"{}",
		"[]",
		"{\"email\":\"a@example.test\"}",
		"{\"email\":\"first@example.test\",\"email\":\"second@example.test\"}",
		"{\"unknown\":true}",
		"{\"email\":\"a@example.test\"} {}",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, raw string) {
		if len(raw) > 16384 {
			return
		}
		var request registerRequestDTO
		_ = decodeStrictJSON([]byte(raw), &request)
	})
}
