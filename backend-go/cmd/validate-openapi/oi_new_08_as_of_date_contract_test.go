package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func oiNew08OpenAPIDocument(t *testing.T) map[string]any {
	t.Helper()
	path := filepath.Clean(filepath.Join("..", "..", "..", "openapi", "openapi.yaml"))
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read OpenAPI: %v", err)
	}
	var document map[string]any
	if err := yaml.Unmarshal(body, &document); err != nil {
		t.Fatalf("parse OpenAPI: %v", err)
	}
	return document
}

func TestOINew08PortfolioSummaryOwnsRuntimeTruthAsOfDateContract(t *testing.T) {
	document := oiNew08OpenAPIDocument(t)
	paths := asMap(document["paths"])
	pathItem := asMap(paths["/api/v1/portfolios/{portfolioId}/summary"])
	operation := asMap(pathItem["get"])
	if operation["operationId"] != "getPortfolioSummary" {
		t.Fatalf("unexpected summary operationId: %v", operation["operationId"])
	}

	parameters := asSlice(operation["parameters"])
	if len(parameters) != 1 {
		t.Fatalf("summary GET must define exactly one operation parameter, got %d", len(parameters))
	}
	parameter := asMap(parameters[0])
	if ref, _ := parameter["$ref"].(string); ref != "" {
		t.Fatalf("summary must not reuse shared AsOfDate parameter, got ref=%q", ref)
	}
	if parameter["name"] != "asOfDate" || parameter["in"] != "query" {
		t.Fatalf("summary asOfDate identity drifted: %+v", parameter)
	}
	required, ok := parameter["required"].(bool)
	if !ok || required {
		t.Fatalf("summary asOfDate must remain optional: %+v", parameter["required"])
	}
	schema := asMap(parameter["schema"])
	if schema["$ref"] != "./components/schemas.yaml#/BusinessDate" {
		t.Fatalf("summary asOfDate must use BusinessDate: %+v", schema)
	}

	description, _ := parameter["description"].(string)
	description = strings.ToLower(description)
	if strings.Contains(description, "defaults to the latest completed moex business date") {
		t.Fatalf("summary asOfDate still claims obsolete MOEX-calendar default: %q", description)
	}
	for _, requiredText := range []string{
		"when supplied",
		"latest available calculated portfolio snapshot",
		"less than or equal",
		"actual selected snapshot date",
		"when omitted",
		"latest available calculated snapshot",
		"wall-clock",
		"moex-calendar",
		"future-dated snapshot",
	} {
		if !strings.Contains(description, requiredText) {
			t.Fatalf("summary asOfDate description missing %q: %q", requiredText, description)
		}
	}
}

func TestOINew08DashboardKeepsSharedAsOfDateContractUntouched(t *testing.T) {
	document := oiNew08OpenAPIDocument(t)
	paths := asMap(document["paths"])
	dashboard := asMap(asMap(paths["/api/v1/dashboard"])["get"])
	if dashboard["operationId"] != "getDashboard" {
		t.Fatalf("unexpected dashboard operationId: %v", dashboard["operationId"])
	}
	foundShared := false
	for _, raw := range asSlice(dashboard["parameters"]) {
		if asMap(raw)["$ref"] == "#/components/parameters/AsOfDate" {
			foundShared = true
		}
	}
	if !foundShared {
		t.Fatal("planned Dashboard must keep shared #/components/parameters/AsOfDate")
	}

	components := asMap(document["components"])
	parameters := asMap(components["parameters"])
	shared := asMap(parameters["AsOfDate"])
	if shared["name"] != "asOfDate" || shared["in"] != "query" || shared["required"] != false {
		t.Fatalf("shared Dashboard AsOfDate identity drifted: %+v", shared)
	}
	if shared["description"] != "Defaults to the latest completed MOEX business date." {
		t.Fatalf("shared Dashboard AsOfDate description changed outside OI-NEW-08 scope: %q", shared["description"])
	}
	if asMap(shared["schema"])["$ref"] != "./components/schemas.yaml#/BusinessDate" {
		t.Fatalf("shared Dashboard AsOfDate schema drifted: %+v", shared["schema"])
	}
}
