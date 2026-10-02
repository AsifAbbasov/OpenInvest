package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestValuationXIRRP3ContractKeepsLegacyROIUnavailableAndBasesExplicit(t *testing.T) {
	schemaPath := filepath.Clean(filepath.Join("..", "..", "..", "openapi", "components", "schemas.yaml"))
	body, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("read schemas: %v", err)
	}
	var schemas map[string]any
	if err := yaml.Unmarshal(body, &schemas); err != nil {
		t.Fatalf("parse schemas: %v", err)
	}

	summaryProperties := asMap(asMap(schemas["PortfolioSummary"])["properties"])
	for field, needles := range map[string][]string{
		"totalValue":      {"canonical cash", "remaining local acquisition-cost basis", "not current market value", "not terminal market value", "not a performance metric"},
		"stockValue":      {"acquisition-cost basis", "not market valuation"},
		"bondValue":       {"acquisition-cost basis", "not market valuation"},
		"investedCapital": {"buy acquisition outflow basis", "not external investor contribution capital"},
	} {
		description, _ := asMap(summaryProperties[field])["description"].(string)
		for _, needle := range needles {
			if !strings.Contains(strings.ToLower(description), needle) {
				t.Fatalf("PortfolioSummary.%s description missing %q: %q", field, needle, description)
			}
		}
	}

	snapshotProperties := asMap(asMap(schemas["PortfolioSnapshot"])["properties"])
	legacy := asMap(snapshotProperties["legacyReturnStatus"])
	if enums := asSlice(legacy["enum"]); len(enums) != 1 || enums[0] != "UNAVAILABLE" {
		t.Fatalf("legacy snapshot return status must be exactly UNAVAILABLE: %+v", legacy)
	}
	for _, field := range []string{"nominalReturnRate", "realReturnRate"} {
		branches := asSlice(asMap(snapshotProperties[field])["oneOf"])
		if len(branches) != 2 || asMap(branches[1])["type"] != "null" {
			t.Fatalf("%s must remain Decimal-or-null unavailable state: %+v", field, snapshotProperties[field])
		}
	}

	returnsPath := filepath.Clean(filepath.Join("..", "..", "..", "openapi", "components", "stage_03_77.yaml"))
	returnsBody, err := os.ReadFile(returnsPath)
	if err != nil {
		t.Fatalf("read Stage 3.77 schema: %v", err)
	}
	var returns map[string]any
	if err := yaml.Unmarshal(returnsBody, &returns); err != nil {
		t.Fatalf("parse Stage 3.77 schema: %v", err)
	}
	terminal := asMap(asMap(asMap(returns["PortfolioReturnAvailable"])["properties"])["terminalPortfolioValue"])
	description, _ := terminal["description"].(string)
	for _, needle := range []string{"terminal market value", "xirr", "user_supplied", "complete", "intentionally distinct", "portfoliosummary.totalvalue"} {
		if !strings.Contains(strings.ToLower(description), needle) {
			t.Fatalf("terminalPortfolioValue description missing %q: %q", needle, description)
		}
	}
}

func TestValuationXIRRP3PublishedSnapshotExampleDoesNotInventROI(t *testing.T) {
	path := filepath.Clean(filepath.Join("..", "..", "..", "openapi", "examples", "portfolios.json"))
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read portfolios example: %v", err)
	}
	var example map[string]any
	if err := json.Unmarshal(body, &example); err != nil {
		t.Fatalf("parse portfolios example: %v", err)
	}
	item := asMap(asSlice(dig(example, "snapshotsResponse", "value", "data", "items"))[0])
	if item["nominalReturnRate"] != nil || item["realReturnRate"] != nil || item["legacyReturnStatus"] != "UNAVAILABLE" {
		t.Fatalf("published snapshot must mark legacy ROI unavailable: %+v", item)
	}
}
