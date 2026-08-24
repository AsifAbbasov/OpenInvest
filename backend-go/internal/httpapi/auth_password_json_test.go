package httpapi

import (
	"encoding/json"
	"testing"
)

func TestLosslessPasswordRejectsMalformedUnicodeBeforeJSONReplacement(t *testing.T) {
	tests := []struct {
		name string
		raw  []byte
	}{
		{name: "raw invalid utf8", raw: append([]byte{'"'}, append([]byte{0xff}, '"')...)},
		{name: "unpaired high surrogate", raw: []byte(`"\uD800"`)},
		{name: "unpaired low surrogate", raw: []byte(`"\uDC00"`)},
		{name: "invalid surrogate pair", raw: []byte(`"\uD83D\u0041"`)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var password losslessPassword
			if err := json.Unmarshal(test.raw, &password); err == nil {
				t.Fatalf("expected malformed password transport to be rejected")
			}
		})
	}
}

func TestLosslessPasswordPreservesValidUnicodeAndEscapes(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "surrogate pair", raw: `"\uD83D\uDE00"`, want: "😀"},
		{name: "legitimate replacement character", raw: `"\uFFFD"`, want: "�"},
		{name: "escaped backslash is not a surrogate", raw: `"\\uD800"`, want: `\uD800`},
		{name: "direct utf8", raw: `"пароль"`, want: "пароль"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var password losslessPassword
			if err := json.Unmarshal([]byte(test.raw), &password); err != nil {
				t.Fatalf("decode valid password: %v", err)
			}
			if got := string(password); got != test.want {
				t.Fatalf("expected %q, got %q", test.want, got)
			}
		})
	}
}
