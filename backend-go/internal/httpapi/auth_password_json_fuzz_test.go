package httpapi

import (
	"encoding/json"
	"testing"
	"unicode/utf8"
)

func FuzzLosslessPasswordJSON(f *testing.F) {
	for _, seed := range [][]byte{
		[]byte(`"password"`),
		[]byte(`"пароль"`),
		[]byte(`"\uD83D\uDE00"`),
		[]byte(`"\uD800"`),
		[]byte(`"\\uD800"`),
		[]byte(`null`),
		[]byte{'"', 0xff, '"'},
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, raw []byte) {
		var password losslessPassword
		if err := json.Unmarshal(raw, &password); err != nil {
			return
		}
		if !utf8.ValidString(string(password)) {
			t.Fatalf("decoder accepted invalid UTF-8: %x", raw)
		}

		encoded, err := json.Marshal(string(password))
		if err != nil {
			t.Fatalf("marshal accepted password: %v", err)
		}
		var roundTrip losslessPassword
		if err := json.Unmarshal(encoded, &roundTrip); err != nil {
			t.Fatalf("canonical password JSON failed to decode: %v", err)
		}
		if roundTrip != password {
			t.Fatalf("password round trip changed value")
		}
	})
}
