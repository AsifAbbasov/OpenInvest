package main

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"gopkg.in/yaml.v3"
)

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
	pattern, err := regexp.Compile(schemas.Decimal.Pattern)
	if err != nil {
		t.Fatalf("compile Decimal pattern %q: %v", schemas.Decimal.Pattern, err)
	}

	for _, value := range []string{
		"0", "-0", "1", "-1", "0.5", "99999999999999999999.99999999",
		"+1", "001", "01.0", " 1.25 ", "1.\n", ".1", "1.000000000", "1e2", "1,25", "\u0661", "-", "",
		"0000000000000000000000000000000001", "100000000000000000000",
	} {
		_, parseErr := decimal.FromString(value)
		if got, want := parseErr == nil, pattern.MatchString(value); got != want {
			t.Fatalf("Decimal parser/schema disagreement for %q: parser=%t schema=%t", value, got, want)
		}
	}
}
