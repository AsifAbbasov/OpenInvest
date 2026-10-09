package httpapi

import (
	"encoding/json"
	"testing"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
)

func FuzzDecodeStrictJSON(f *testing.F) {
	for _, seed := range []string{
		`{"name":"alice","count":1}`,
		`{"name":"alice","unknown":true}`,
		`{"name":"alice"} trailing`,
		`{}`,
		`null`,
		"",
	} {
		f.Add([]byte(seed))
	}

	type payload struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	f.Fuzz(func(t *testing.T, body []byte) {
		if len(body) > 1<<20 {
			t.Skip()
		}
		var target payload
		err := decodeStrictJSON(body, &target)
		if err != nil {
			return
		}
		if !json.Valid(body) {
			t.Fatalf("decodeStrictJSON accepted invalid JSON: %q", string(body))
		}
	})
}

func FuzzParseMoney(f *testing.F) {
	for _, seed := range [][2]string{
		{"0", "RUB"},
		{"1.00000000", "RUB"},
		{"99999999999999999999.99999999", "RUB"},
		{"01", "RUB"},
		{"1e2", "USD"},
		{"", ""},
	} {
		f.Add(seed[0], seed[1])
	}

	f.Fuzz(func(t *testing.T, amount, currency string) {
		value, err := parseMoney(moneyDTO{Amount: amount, Currency: currency})
		if err != nil {
			return
		}
		if !value.Amount.FitsStorage() {
			t.Fatalf("parseMoney accepted non-storable amount: %q", amount)
		}
		reparsed, err := decimal.FromString(value.Amount.String())
		if err != nil || !value.Amount.Equal(reparsed) {
			t.Fatalf("parseMoney produced non-canonical amount: input=%q canonical=%q err=%v", amount, value.Amount.String(), err)
		}
		if value.Currency != currency {
			t.Fatalf("parseMoney changed currency: got=%q want=%q", value.Currency, currency)
		}
	})
}
