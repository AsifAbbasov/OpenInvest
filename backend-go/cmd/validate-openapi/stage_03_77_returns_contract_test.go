package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestStage377ReturnsRequiresExplicitAsOfDateAndCanonicalResponse(t *testing.T) {
	path := filepath.Clean(filepath.Join("..", "..", "..", "openapi", "openapi.yaml"))
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read OpenAPI: %v", err)
	}
	var document map[string]any
	if err := yaml.Unmarshal(body, &document); err != nil {
		t.Fatalf("parse OpenAPI: %v", err)
	}
	paths := asMap(document["paths"])
	returnsPath := asMap(paths["/api/v1/portfolios/{portfolioId}/returns"])
	operation := asMap(returnsPath["get"])
	if operation["operationId"] != "getPortfolioReturns" {
		t.Fatalf("unexpected Stage 3.77 operationId: %v", operation["operationId"])
	}
	parameters := asSlice(operation["parameters"])
	if len(parameters) != 1 {
		t.Fatalf("Stage 3.77 GET must define exactly one operation parameter, got %d", len(parameters))
	}
	asOfDate := asMap(parameters[0])
	if asOfDate["name"] != "asOfDate" || asOfDate["in"] != "query" || asOfDate["required"] != true {
		t.Fatalf("Stage 3.77 asOfDate must be an explicit required query parameter: %+v", asOfDate)
	}
	if ref, _ := asOfDate["$ref"].(string); ref == "#/components/parameters/AsOfDate" {
		t.Fatal("Stage 3.77 must not reuse a parameter with default-date semantics")
	}
	responses := asMap(operation["responses"])
	for _, status := range []string{"200", "400", "401", "404"} {
		if responses[status] == nil {
			t.Fatalf("Stage 3.77 response %s is missing", status)
		}
	}
	content := asMap(asMap(responses["200"])["content"])
	jsonMedia := asMap(content["application/json"])
	if asMap(jsonMedia["schema"])["$ref"] != "./components/stage_03_77.yaml#/PortfolioReturnResponse" {
		t.Fatalf("Stage 3.77 200 schema drifted: %+v", jsonMedia["schema"])
	}
}

func TestStage377SchemaFreezesAvailabilityReasonsAndStringDecimal(t *testing.T) {
	path := filepath.Clean(filepath.Join("..", "..", "..", "openapi", "components", "stage_03_77.yaml"))
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read Stage 3.77 component: %v", err)
	}
	var schemas map[string]any
	if err := yaml.Unmarshal(body, &schemas); err != nil {
		t.Fatalf("parse Stage 3.77 component: %v", err)
	}
	available := asMap(schemas["PortfolioReturnAvailable"])
	availableProperties := asMap(available["properties"])
	if asMap(availableProperties["status"])["const"] != "AVAILABLE" || asMap(availableProperties["reason"])["type"] != "null" {
		t.Fatalf("AVAILABLE discriminator drifted: %+v", availableProperties)
	}
	xirr := asMap(availableProperties["xirr"])
	if xirr["type"] == "number" || xirr["format"] == "double" {
		t.Fatalf("XIRR must never be a floating JSON number: %+v", xirr)
	}
	unavailable := asMap(schemas["PortfolioReturnUnavailable"])
	unavailableProperties := asMap(unavailable["properties"])
	if asMap(unavailableProperties["xirr"])["type"] != "null" {
		t.Fatalf("UNAVAILABLE xirr must be schema-enforced null: %+v", unavailableProperties["xirr"])
	}
	reasons := asSlice(asMap(schemas["PortfolioReturnUnavailableReason"])["enum"])
	want := []string{
		"INCOMPLETE_VALUATION",
		"NO_EXTERNAL_CONTRIBUTIONS",
		"INSUFFICIENT_DATE_SPAN",
		"NO_SIGN_CHANGE",
		"NON_POSITIVE_TERMINAL_VALUE",
		"AMBIGUOUS_MULTIPLE_ROOTS",
		"NUMERICAL_SOLUTION_FAILED",
	}
	if len(reasons) != len(want) {
		t.Fatalf("Stage 3.77 unavailable reason count drifted: %+v", reasons)
	}
	for index, expected := range want {
		if reasons[index] != expected {
			t.Fatalf("Stage 3.77 reason[%d]=%v want %s", index, reasons[index], expected)
		}
	}
}

func TestStage377PublishedExamplesKeepUnavailableDistinctFromZero(t *testing.T) {
	path := filepath.Clean(filepath.Join("..", "..", "..", "openapi", "examples", "portfolio-returns.json"))
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read Stage 3.77 examples: %v", err)
	}
	var examples map[string]any
	if err := json.Unmarshal(body, &examples); err != nil {
		t.Fatalf("parse Stage 3.77 examples: %v", err)
	}
	available := asMap(dig(examples, "availableResponse", "value", "data"))
	if available["status"] != "AVAILABLE" || available["xirr"] != "0.12048717" || available["reason"] != nil {
		t.Fatalf("available example drifted: %+v", available)
	}
	unavailable := asMap(dig(examples, "unavailableResponse", "value", "data"))
	if unavailable["status"] != "UNAVAILABLE" || unavailable["xirr"] != nil || unavailable["reason"] != "INCOMPLETE_VALUATION" {
		t.Fatalf("unavailable example must remain null, not zero: %+v", unavailable)
	}
}
