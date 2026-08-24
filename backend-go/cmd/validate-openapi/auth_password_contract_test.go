package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type stage335PasswordSchema struct {
	MinLength   int    `yaml:"minLength"`
	MaxLength   int    `yaml:"maxLength"`
	Description string `yaml:"description"`
}

type stage335RequestSchema struct {
	Properties struct {
		Password stage335PasswordSchema `yaml:"password"`
	} `yaml:"properties"`
}

type stage335AuthSchemas struct {
	RegisterRequest stage335RequestSchema `yaml:"RegisterRequest"`
	LoginRequest    stage335RequestSchema `yaml:"LoginRequest"`
}

func TestStage335AuthPasswordContract(t *testing.T) {
	path := filepath.Clean(filepath.Join("..", "..", "..", "openapi", "components", "schemas.yaml"))
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read schemas: %v", err)
	}
	var schemas stage335AuthSchemas
	if err := yaml.Unmarshal(body, &schemas); err != nil {
		t.Fatalf("parse schemas: %v", err)
	}

	assertPasswordBounds(t, "RegisterRequest", schemas.RegisterRequest.Properties.Password, 12, 256, "creation")
	assertPasswordBounds(t, "LoginRequest", schemas.LoginRequest.Properties.Password, 1, 256, "historical")
}

func assertPasswordBounds(t *testing.T, schemaName string, password stage335PasswordSchema, minLength int, maxLength int, descriptionNeedle string) {
	t.Helper()
	if password.MinLength != minLength {
		t.Fatalf("%s minLength: expected %d, got %d", schemaName, minLength, password.MinLength)
	}
	if password.MaxLength != maxLength {
		t.Fatalf("%s maxLength: expected %d, got %d", schemaName, maxLength, password.MaxLength)
	}
	lower := strings.ToLower(password.Description)
	for _, needle := range []string{"unicode code points", "exact utf-8 bytes", "normalize", descriptionNeedle} {
		if !strings.Contains(lower, strings.ToLower(needle)) {
			t.Fatalf("%s password description missing %q: %q", schemaName, needle, password.Description)
		}
	}
}
