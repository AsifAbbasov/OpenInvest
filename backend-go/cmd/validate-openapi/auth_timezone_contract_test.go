package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type stage337TimezoneSchema struct {
	MinLength   int    `yaml:"minLength"`
	MaxLength   int    `yaml:"maxLength"`
	Pattern     string `yaml:"pattern"`
	Description string `yaml:"description"`
}

type stage337TimezoneProperties struct {
	Timezone stage337TimezoneSchema `yaml:"timezone"`
}

type stage337TimezoneObjectSchema struct {
	Properties stage337TimezoneProperties `yaml:"properties"`
}

type stage337AuthSchemas struct {
	User            stage337TimezoneObjectSchema `yaml:"User"`
	RegisterRequest stage337TimezoneObjectSchema `yaml:"RegisterRequest"`
}

func TestStage337AuthTimezoneContract(t *testing.T) {
	path := filepath.Clean(filepath.Join("..", "..", "..", "openapi", "components", "schemas.yaml"))
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read schemas: %v", err)
	}
	var schemas stage337AuthSchemas
	if err := yaml.Unmarshal(body, &schemas); err != nil {
		t.Fatalf("parse schemas: %v", err)
	}

	registerTimezone := schemas.RegisterRequest.Properties.Timezone
	if registerTimezone.MinLength != 1 || registerTimezone.MaxLength != 64 {
		t.Fatalf("RegisterRequest timezone bounds changed: min=%d max=%d", registerTimezone.MinLength, registerTimezone.MaxLength)
	}
	assertStage337TimezoneDescription(t, "RegisterRequest", registerTimezone)
	assertStage337TimezoneDescription(t, "User", schemas.User.Properties.Timezone)
}

func assertStage337TimezoneDescription(t *testing.T, schemaName string, timezone stage337TimezoneSchema) {
	t.Helper()
	if timezone.Pattern != "" {
		t.Fatalf("%s timezone must not use a handcrafted IANA pattern: %q", schemaName, timezone.Pattern)
	}
	lower := strings.ToLower(timezone.Description)
	for _, needle := range []string{
		"iana timezone database identifier",
		"server timezone database resolution",
		"utc is supported",
		"local",
		"+04:00",
		"utc+04:00",
		"etc/gmt+4",
	} {
		if !strings.Contains(lower, needle) {
			t.Fatalf("%s timezone description missing %q: %q", schemaName, needle, timezone.Description)
		}
	}
}
