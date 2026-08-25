package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"gopkg.in/yaml.v3"
)

const stage336PublishedDecimalPattern = `^-?(0|[1-9][0-9]{0,19})(\.[0-9]{1,8})?$`

func TestStage336DecimalParserMatchesPublishedOpenAPIGrammar(t *testing.T) {
	body, err := os.ReadFile(filepath.Clean(filepath.Join("..", "..", "..", "openapi", "components", "schemas.yaml")))
	if err != nil {
		t.Fatalf("read schemas: %v", err)
	}
	var schemas struct {
		Decimal struct {
			Pattern string `yaml:"pattern"`
		} `yaml:"Decimal"`
	}
	if err := yaml.Unmarshal(body, &schemas); err != nil {
		t.Fatalf("parse schemas: %v", err)
	}
	if schemas.Decimal.Pattern != stage336PublishedDecimalPattern {
		t.Fatalf("published Decimal pattern changed: got %q want %q", schemas.Decimal.Pattern, stage336PublishedDecimalPattern)
	}

	for _, testCase := range []struct {
		name  string
		value string
		valid bool
	}{
		{name: "zero", value: "0", valid: true},
		{name: "negative zero", value: "-0", valid: true},
		{name: "integer", value: "1", valid: true},
		{name: "negative integer", value: "-1", valid: true},
		{name: "fraction", value: "0.5", valid: true},
		{name: "maximum precision", value: "99999999999999999999.99999999", valid: true},
		{name: "plus sign", value: "+1"},
		{name: "leading zero integer", value: "001"},
		{name: "leading zero fraction", value: "01.0"},
		{name: "surrounding whitespace", value: " 1.25 "},
		{name: "empty fraction with newline", value: "1.\n"},
		{name: "carriage return line ending", value: "1.0\r\n"},
		{name: "missing integer", value: ".1"},
		{name: "too many fraction digits", value: "1.000000000"},
		{name: "exponent", value: "1e2"},
		{name: "separator", value: "1,25"},
		{name: "unicode digit", value: "\u0661"},
		{name: "sign without digits", value: "-"},
		{name: "empty", value: ""},
		{name: "oversized leading zero lexical input", value: "0000000000000000000000000000000001"},
		{name: "too many integer digits", value: "100000000000000000000"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			_, parseErr := decimal.FromString(testCase.value)
			if got := parseErr == nil; got != testCase.valid {
				t.Fatalf("Decimal parser admission for %q: got %t want %t", testCase.value, got, testCase.valid)
			}
		})
	}
}
