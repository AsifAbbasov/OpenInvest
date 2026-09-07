package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"gopkg.in/yaml.v3"
)

func TestStage372PositionsUsesEndpointSpecificAsOfDateSemantics(t *testing.T) {
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
	positionPath := asMap(paths["/api/v1/portfolios/{portfolioId}/positions"])
	operation := asMap(positionPath["get"])
	if operation["operationId"] != "getPortfolioPositions" {
		t.Fatalf("unexpected Stage 3.72 operationId: %v", operation["operationId"])
	}
	parameters := asSlice(operation["parameters"])
	if len(parameters) != 1 {
		t.Fatalf("Stage 3.72 GET must define exactly one operation parameter, got %d", len(parameters))
	}
	parameter := asMap(parameters[0])
	if parameter["name"] != "asOfDate" || parameter["in"] != "query" {
		t.Fatalf("Stage 3.72 asOfDate must be endpoint-local query parameter: %+v", parameter)
	}
	description := strings.ToLower(parameter["description"].(string))
	for _, required := range []string{"replay all accepted", "fails closed", "does not mean", "latest completed moex"} {
		if !strings.Contains(description, required) {
			t.Fatalf("Stage 3.72 asOfDate description missing %q: %s", required, description)
		}
	}
	if strings.Contains(description, "accepted supported") {
		t.Fatal("Stage 3.72 asOfDate wording must not imply unsupported accepted rows are silently skipped")
	}
	if ref, _ := parameter["$ref"].(string); ref == "#/components/parameters/AsOfDate" {
		t.Fatal("Stage 3.72 must not reuse the summary/dashboard AsOfDate default semantics")
	}
}

func TestStage372MarketValuationSchemaCanOnlyPublishUnavailableNullState(t *testing.T) {
	path := filepath.Clean(filepath.Join("..", "..", "..", "openapi", "components", "schemas.yaml"))
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read schemas: %v", err)
	}
	var schemas map[string]any
	if err := yaml.Unmarshal(body, &schemas); err != nil {
		t.Fatalf("parse schemas: %v", err)
	}
	market := asMap(schemas["MarketValuationUnavailable"])
	properties := asMap(market["properties"])
	if asMap(properties["status"])["const"] != "UNAVAILABLE" {
		t.Fatalf("market status must be frozen to UNAVAILABLE: %+v", properties["status"])
	}
	if asMap(properties["reason"])["const"] != "NO_APPROVED_MARKET_PRICE_SOURCE" {
		t.Fatalf("market reason drifted: %+v", properties["reason"])
	}
	for _, field := range []string{"marketPrice", "marketValue", "unrealizedGain", "marketWeight", "provider", "asOf"} {
		if asMap(properties[field])["type"] != "null" {
			t.Fatalf("Stage 3.72 field %s must be schema-enforced null: %+v", field, properties[field])
		}
	}
}

func TestStage372AcquisitionBasisWeightSchemaIsBoundedToUnitInterval(t *testing.T) {
	path := filepath.Clean(filepath.Join("..", "..", "..", "openapi", "components", "schemas.yaml"))
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read schemas: %v", err)
	}
	var schemas map[string]any
	if err := yaml.Unmarshal(body, &schemas); err != nil {
		t.Fatalf("parse schemas: %v", err)
	}
	properties := asMap(asMap(schemas["PortfolioPositionProjection"])["properties"])
	weight := asMap(properties["acquisitionBasisWeight"])
	branches := asSlice(weight["oneOf"])
	if len(branches) != 2 {
		t.Fatalf("acquisitionBasisWeight must be Decimal-or-null, got %+v", weight)
	}
	decimalBranch := asMap(branches[0])
	allOf := asSlice(decimalBranch["allOf"])
	if len(allOf) != 2 {
		t.Fatalf("weight Decimal branch must compose NonNegativeDecimal with unit-interval pattern: %+v", decimalBranch)
	}
	pattern := asMap(allOf[1])["pattern"]
	if pattern != `^(0(\.[0-9]{1,8})?|1(\.0{1,8})?)$` {
		t.Fatalf("weight unit-interval pattern drifted: %v", pattern)
	}
	if asMap(branches[1])["type"] != "null" {
		t.Fatalf("weight zero-denominator branch must remain null: %+v", branches[1])
	}
}

func TestStage372PublishedExampleContainsNoFabricatedMarketValue(t *testing.T) {
	path := filepath.Clean(filepath.Join("..", "..", "..", "openapi", "examples", "positions.json"))
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read positions example: %v", err)
	}
	var examples map[string]any
	if err := json.Unmarshal(body, &examples); err != nil {
		t.Fatalf("parse positions example: %v", err)
	}
	items := asSlice(dig(examples, "positionsResponse", "value", "data", "items"))
	if len(items) != 2 {
		t.Fatalf("published Stage 3.72 example must expose both denominator components, got %d items", len(items))
	}
	var sber map[string]any
	totalFromItems := decimal.Zero()
	for _, raw := range items {
		item := asMap(raw)
		market := asMap(item["marketValuation"])
		for _, field := range []string{"marketPrice", "marketValue", "unrealizedGain", "marketWeight", "provider", "asOf"} {
			if market[field] != nil {
				t.Fatalf("published market field %s must remain null, got %#v", field, market[field])
			}
		}
		basis := asMap(item["acquisitionBasis"])["amount"].(string)
		totalFromItems = totalFromItems.Add(decimal.Must(basis))
		if item["ticker"] == "SBER" {
			sber = item
		}
	}
	publishedTotal := dig(examples, "positionsResponse", "value", "data", "totalAcquisitionBasis", "amount").(string)
	if totalFromItems.String() != publishedTotal {
		t.Fatalf("published totalAcquisitionBasis must equal the sum of item acquisition bases: items=%s total=%s", totalFromItems.String(), publishedTotal)
	}
	if sber == nil || sber["acquisitionBasisWeight"] != "0.34210000" {
		t.Fatalf("published SBER acquisitionBasisWeight drifted: %+v", sber)
	}
}
