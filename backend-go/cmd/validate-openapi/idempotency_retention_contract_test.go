package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type stage338IdempotencySchema struct {
	MinLength int    `yaml:"minLength"`
	MaxLength int    `yaml:"maxLength"`
	Pattern   string `yaml:"pattern"`
}

type stage338IdempotencyParameter struct {
	Name        string                    `yaml:"name"`
	In          string                    `yaml:"in"`
	Required    bool                      `yaml:"required"`
	Description string                    `yaml:"description"`
	Schema      stage338IdempotencySchema `yaml:"schema"`
}

type stage338OpenAPI struct {
	Components struct {
		Parameters struct {
			IdempotencyKey stage338IdempotencyParameter `yaml:"IdempotencyKey"`
		} `yaml:"parameters"`
	} `yaml:"components"`
}

func TestStage338IdempotencyRetentionContract(t *testing.T) {
	path := filepath.Clean(filepath.Join("..", "..", "..", "openapi", "openapi.yaml"))
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read OpenAPI: %v", err)
	}

	var spec stage338OpenAPI
	if err := yaml.Unmarshal(body, &spec); err != nil {
		t.Fatalf("parse OpenAPI: %v", err)
	}
	parameter := spec.Components.Parameters.IdempotencyKey
	if parameter.Name != "Idempotency-Key" || parameter.In != "header" || !parameter.Required {
		t.Fatalf("Idempotency-Key transport contract changed: %+v", parameter)
	}
	if parameter.Schema.MinLength != 16 || parameter.Schema.MaxLength != 128 ||
		parameter.Schema.Pattern != "^[A-Za-z0-9._:-]+$" {
		t.Fatalf("Idempotency-Key lexical contract changed: %+v", parameter.Schema)
	}

	lower := strings.ToLower(parameter.Description)
	for _, needle := range []string{
		"24 hours",
		"server command admission",
		"before expiry",
		"original result",
		"different payload returns 409",
		"at or after expiry",
		"new command",
	} {
		if !strings.Contains(lower, needle) {
			t.Fatalf("Idempotency-Key description missing %q: %q", needle, parameter.Description)
		}
	}
}
