package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type stage339StringProperty struct {
	MinLength    int    `yaml:"minLength"`
	MaxLength    int    `yaml:"maxLength"`
	Description  string `yaml:"description"`
	MaxUTF8Bytes int    `yaml:"x-openinvest-max-utf8-bytes"`
}

type stage339ObjectSchema struct {
	Properties map[string]stage339StringProperty `yaml:"properties"`
}

type stage339Schemas struct {
	Portfolio                 stage339ObjectSchema `yaml:"Portfolio"`
	CreatePortfolioRequest    stage339ObjectSchema `yaml:"CreatePortfolioRequest"`
	Transaction               stage339ObjectSchema `yaml:"Transaction"`
	TradeTransactionRequest   stage339ObjectSchema `yaml:"TradeTransactionRequest"`
	IncomeTransactionRequest  stage339ObjectSchema `yaml:"IncomeTransactionRequest"`
	ExpenseTransactionRequest stage339ObjectSchema `yaml:"ExpenseTransactionRequest"`
	CashTransactionRequest    stage339ObjectSchema `yaml:"CashTransactionRequest"`
	ImportReviewRequest       stage339ObjectSchema `yaml:"ImportReviewRequest"`
	ImportAppendRequest       stage339ObjectSchema `yaml:"ImportAppendRequest"`
}

func TestStage339UnicodeAndCSVByteContract(t *testing.T) {
	schemasPath := filepath.Clean(filepath.Join("..", "..", "..", "openapi", "components", "schemas.yaml"))
	body, err := os.ReadFile(schemasPath)
	if err != nil {
		t.Fatalf("read schemas: %v", err)
	}
	var schemas stage339Schemas
	if err := yaml.Unmarshal(body, &schemas); err != nil {
		t.Fatalf("parse schemas: %v", err)
	}

	assertStage339CodePointField(t, "Portfolio.name", schemas.Portfolio.Properties["name"], 100)
	assertStage339CodePointField(t, "CreatePortfolioRequest.name", schemas.CreatePortfolioRequest.Properties["name"], 100)
	for name, schema := range map[string]stage339ObjectSchema{
		"Transaction":               schemas.Transaction,
		"TradeTransactionRequest":   schemas.TradeTransactionRequest,
		"IncomeTransactionRequest":  schemas.IncomeTransactionRequest,
		"ExpenseTransactionRequest": schemas.ExpenseTransactionRequest,
		"CashTransactionRequest":    schemas.CashTransactionRequest,
	} {
		assertStage339CodePointField(t, name+".note", schema.Properties["note"], 500)
	}
	assertStage339CodePointField(t, "ImportReviewRequest.sourceAccountLabel", schemas.ImportReviewRequest.Properties["sourceAccountLabel"], 120)
	assertStage339CodePointField(t, "ImportAppendRequest.sourceAccountLabel", schemas.ImportAppendRequest.Properties["sourceAccountLabel"], 120)

	for name, csv := range map[string]stage339StringProperty{
		"ImportReviewRequest.csvPayload": schemas.ImportReviewRequest.Properties["csvPayload"],
		"ImportAppendRequest.csvPayload": schemas.ImportAppendRequest.Properties["csvPayload"],
	} {
		if csv.MaxLength != 0 {
			t.Fatalf("%s must not publish JSON Schema maxLength for a byte budget; got %d", name, csv.MaxLength)
		}
		if csv.MaxUTF8Bytes != 2097152 {
			t.Fatalf("%s byte extension: got %d want 2097152", name, csv.MaxUTF8Bytes)
		}
		lower := strings.ToLower(csv.Description)
		if !strings.Contains(lower, "utf-8 bytes") || !strings.Contains(lower, "not a json schema character limit") {
			t.Fatalf("%s must document the UTF-8 byte resource budget: %q", name, csv.Description)
		}
	}
	if strings.Contains(string(body), "maxLength: 2097152") {
		t.Fatal("CSV byte budget leaked back into JSON Schema maxLength")
	}

	openAPIPath := filepath.Clean(filepath.Join("..", "..", "..", "openapi", "openapi.yaml"))
	openAPIBody, err := os.ReadFile(openAPIPath)
	if err != nil {
		t.Fatalf("read openapi: %v", err)
	}
	if !strings.Contains(string(openAPIBody), "Unicode code points on the raw submitted query before trimming") {
		t.Fatal("asset query contract must publish raw code-point admission semantics")
	}
}

func assertStage339CodePointField(t *testing.T, name string, field stage339StringProperty, max int) {
	t.Helper()
	if field.MaxLength != max {
		t.Fatalf("%s maxLength: got %d want %d", name, field.MaxLength, max)
	}
	if !strings.Contains(strings.ToLower(field.Description), "unicode code points") {
		t.Fatalf("%s must state Unicode code-point semantics: %q", name, field.Description)
	}
}
