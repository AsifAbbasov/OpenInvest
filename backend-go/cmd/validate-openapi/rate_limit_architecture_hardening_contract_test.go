package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestArchitectureHardeningExpensiveReadResponsesArePublishedOnAllEndpoints(t *testing.T) {
	openAPIPath := filepath.Clean(filepath.Join("..", "..", "..", "openapi", "openapi.yaml"))
	body, err := os.ReadFile(openAPIPath)
	if err != nil {
		t.Fatalf("read OpenAPI: %v", err)
	}
	var document map[string]any
	if err := yaml.Unmarshal(body, &document); err != nil {
		t.Fatalf("parse OpenAPI: %v", err)
	}

	paths := asMap(document["paths"])
	for _, path := range []string{
		"/api/v1/portfolios/{portfolioId}/summary",
		"/api/v1/portfolios/{portfolioId}/positions",
		"/api/v1/portfolios/{portfolioId}/returns",
		"/api/v1/portfolios/{portfolioId}/cash-flow",
	} {
		operation := asMap(asMap(paths[path])["get"])
		responses := asMap(operation["responses"])
		if got := asMap(responses["429"])["$ref"]; got != "./components/responses.yaml#/responses/ExpensiveReadSubjectLimited" {
			t.Fatalf("%s 429 response ref=%v", path, got)
		}
		if got := asMap(responses["503"])["$ref"]; got != "./components/responses.yaml#/responses/ExpensiveReadCapacityUnavailable" {
			t.Fatalf("%s 503 response ref=%v", path, got)
		}
	}
}

func TestArchitectureHardeningExpensiveReadResponseComponentsFreezeRetryAfterAndEnvelope(t *testing.T) {
	responsePath := filepath.Clean(filepath.Join("..", "..", "..", "openapi", "components", "responses.yaml"))
	body, err := os.ReadFile(responsePath)
	if err != nil {
		t.Fatalf("read response components: %v", err)
	}
	var document map[string]any
	if err := yaml.Unmarshal(body, &document); err != nil {
		t.Fatalf("parse response components: %v", err)
	}
	responses := asMap(document["responses"])

	tests := []struct {
		name              string
		wantDescription   string
		wantExampleRef    string
	}{
		{
			name:            "ExpensiveReadSubjectLimited",
			wantDescription: "authenticated subject",
			wantExampleRef:  "../examples/errors.json#/expensiveReadSubjectLimited",
		},
		{
			name:            "ExpensiveReadCapacityUnavailable",
			wantDescription: "Process-local shared capacity",
			wantExampleRef:  "../examples/errors.json#/expensiveReadCapacityUnavailable",
		},
	}
	for _, test := range tests {
		response := asMap(responses[test.name])
		description, _ := response["description"].(string)
		if !strings.Contains(description, test.wantDescription) {
			t.Fatalf("%s description=%q", test.name, description)
		}
		headers := asMap(response["headers"])
		if got := asMap(headers["Retry-After"])["$ref"]; got != "#/headers/RetryAfter" {
			t.Fatalf("%s Retry-After ref=%v", test.name, got)
		}
		media := asMap(asMap(response["content"])["application/json"])
		if got := asMap(media["schema"])["$ref"]; got != "./schemas.yaml#/ErrorResponse" {
			t.Fatalf("%s schema=%v", test.name, got)
		}
		examples := asMap(media["examples"])
		if got := asMap(examples["default"])["$ref"]; got != test.wantExampleRef {
			t.Fatalf("%s example ref=%v", test.name, got)
		}
	}
}

func TestArchitectureHardeningExpensiveReadExamplesMatchRuntimeCodes(t *testing.T) {
	examplePath := filepath.Clean(filepath.Join("..", "..", "..", "openapi", "examples", "errors.json"))
	body, err := os.ReadFile(examplePath)
	if err != nil {
		t.Fatalf("read error examples: %v", err)
	}
	var examples map[string]any
	if err := json.Unmarshal(body, &examples); err != nil {
		t.Fatalf("parse error examples: %v", err)
	}

	subject := asMap(dig(examples, "expensiveReadSubjectLimited", "value", "error"))
	if subject["code"] != "RATE_LIMITED" || subject["message"] != "Too many concurrent portfolio projection requests" {
		t.Fatalf("subject-limit example drifted: %+v", subject)
	}
	capacity := asMap(dig(examples, "expensiveReadCapacityUnavailable", "value", "error"))
	if capacity["code"] != "READ_CAPACITY_EXHAUSTED" || capacity["message"] != "Portfolio projection capacity is temporarily exhausted" {
		t.Fatalf("shared-capacity example drifted: %+v", capacity)
	}
}
