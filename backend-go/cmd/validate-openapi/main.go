package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/importer"
	"gopkg.in/yaml.v3"
)

var httpMethods = []string{"get", "post", "put", "patch", "delete", "options", "head", "trace"}

var requiredOperations = map[string]string{
	"GET /api/v1/health":                                                   "getHealth",
	"GET /api/v1/ready":                                                    "getReadiness",
	"POST /api/v1/auth/register":                                           "register",
	"POST /api/v1/auth/login":                                              "login",
	"POST /api/v1/auth/refresh":                                            "refreshSession",
	"POST /api/v1/auth/logout":                                             "logout",
	"GET /api/v1/assets/search":                                            "searchAssets",
	"GET /api/v1/assets/{ticker}":                                          "getAsset",
	"GET /api/v1/portfolios":                                               "listPortfolios",
	"POST /api/v1/portfolios":                                              "createPortfolio",
	"GET /api/v1/portfolios/{portfolioId}":                                 "getPortfolio",
	"PATCH /api/v1/portfolios/{portfolioId}":                               "updatePortfolio",
	"DELETE /api/v1/portfolios/{portfolioId}":                              "deletePortfolio",
	"GET /api/v1/portfolios/{portfolioId}/summary":                         "getPortfolioSummary",
	"GET /api/v1/portfolios/{portfolioId}/snapshots":                       "listPortfolioSnapshots",
	"GET /api/v1/portfolios/{portfolioId}/transactions":                    "listTransactions",
	"POST /api/v1/portfolios/{portfolioId}/transactions":                   "createTransaction",
	"PATCH /api/v1/portfolios/{portfolioId}/transactions/{transactionId}":  "correctTransaction",
	"DELETE /api/v1/portfolios/{portfolioId}/transactions/{transactionId}": "reverseTransaction",
	"POST /api/v1/portfolios/{portfolioId}/imports/review":                 "reviewPortfolioImport",
	"POST /api/v1/portfolios/{portfolioId}/imports/append":                 "appendReviewedPortfolioImport",
	"GET /api/v1/dividends/calendar":                                       "getDividendCalendar",
	"POST /api/v1/dividends/calculate":                                     "calculateDividend",
	"GET /api/v1/dashboard":                                                "getDashboard",
}

var publicOperations = stringSet("getHealth", "getReadiness", "register", "login", "searchAssets", "getAsset", "calculateDividend")
var refreshOperations = stringSet("refreshSession", "logout")
var idempotentOperations = stringSet("createPortfolio", "deletePortfolio", "createTransaction", "correctTransaction", "reverseTransaction", "appendReviewedPortfolioImport", "calculateDividend")

var requiredSchemas = []string{
	"Money", "Decimal", "BusinessDate", "SystemTimestamp", "Asset", "AssetType", "Portfolio", "Transaction",
	"TransactionType", "PortfolioSummary", "PortfolioSnapshot", "ImportReviewResult", "ImportAppendResult", "DividendEvent", "DividendCalculation",
	"RealReturn", "PurchasingPower", "Pagination", "Error", "BaseResponse", "ErrorResponse",
}

var exampleSourceCodes = []string{"EXAMPLE_MARKET_DATA", "EXAMPLE_CORPORATE_ACTIONS", "EXAMPLE_PURCHASING_POWER"}

type validator struct {
	root           string
	documents      map[string]any
	visitedRefs    map[string]bool
	referenceCount int
	operationCount int
	errors         []string
	rootSecurity   any
}

func main() {
	root := filepath.Clean(filepath.Join("..", "openapi", "openapi.yaml"))
	v := &validator{root: root, documents: map[string]any{}, visitedRefs: map[string]bool{}}
	v.validate()
}

func (v *validator) validate() {
	rootDocument := v.loadDocument(v.root)
	v.validateStructure(asMap(rootDocument))
	v.validateFinancialGuardVectors()
	v.validateExampleSourceCodes()
	v.validateImportExamples()
	v.walk(rootDocument, v.root, "#")

	if len(v.errors) > 0 {
		for _, err := range v.errors {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		}
		os.Exit(1)
	}

	fmt.Printf("OpenAPI validation passed: %d operations, %d resolved references, %d documents\n", v.operationCount, v.referenceCount, len(v.documents))
}

func (v *validator) validateStructure(document map[string]any) {
	if document["openapi"] != "3.1.0" {
		v.errors = append(v.errors, "root openapi version must be 3.1.0")
	}

	v.rootSecurity = document["security"]
	if !jsonEqual(v.rootSecurity, []any{map[string]any{"bearerAuth": []any{}}}) {
		v.errors = append(v.errors, "root security must require bearerAuth")
	}

	paths := asMap(document["paths"])
	requiredPaths := make([]string, 0, len(requiredOperations))
	for key := range requiredOperations {
		requiredPaths = append(requiredPaths, strings.SplitN(key, " ", 2)[1])
	}
	requiredPaths = unique(requiredPaths)

	missingPaths := stringDifference(requiredPaths, mapKeys(paths))
	if len(missingPaths) > 0 {
		v.errors = append(v.errors, "missing required paths: "+strings.Join(missingPaths, ", "))
	}

	var unversioned []string
	for _, path := range mapKeys(paths) {
		if !strings.HasPrefix(path, "/api/v1/") {
			unversioned = append(unversioned, path)
		}
	}
	if len(unversioned) > 0 {
		v.errors = append(v.errors, "unversioned paths: "+strings.Join(unversioned, ", "))
	}

	schemas := asMap(asMap(document["components"])["schemas"])
	for _, schemaName := range requiredSchemas {
		if _, ok := schemas[schemaName]; !ok {
			v.errors = append(v.errors, "missing registered canonical schema "+schemaName)
		}
	}

	var operationIDs []string
	actualOperations := map[string]string{}
	for path, rawPathItem := range paths {
		pathItem := asMap(rawPathItem)
		for _, method := range httpMethods {
			rawOperation, ok := pathItem[method]
			if !ok {
				continue
			}
			operation := asMap(rawOperation)
			v.operationCount++
			operationID, _ := operation["operationId"].(string)
			if operationID == "" {
				v.errors = append(v.errors, fmt.Sprintf("%s %s has no operationId", strings.ToUpper(method), path))
			} else {
				operationIDs = append(operationIDs, operationID)
			}
			actualOperations[fmt.Sprintf("%s %s", strings.ToUpper(method), path)] = operationID
			v.validateOperation(path, method, operation, asSlice(pathItem["parameters"]))
		}
	}

	if duplicates := duplicates(operationIDs); len(duplicates) > 0 {
		v.errors = append(v.errors, "duplicate operationIds: "+strings.Join(duplicates, ", "))
	}

	missingOperations := stringDifference(mapKeys(requiredOperations), mapKeys(actualOperations))
	unexpectedOperations := stringDifference(mapKeys(actualOperations), mapKeys(requiredOperations))
	if len(missingOperations) > 0 {
		v.errors = append(v.errors, "missing required operations: "+strings.Join(missingOperations, ", "))
	}
	if len(unexpectedOperations) > 0 {
		v.errors = append(v.errors, "unexpected operations: "+strings.Join(unexpectedOperations, ", "))
	}
	for key, operationID := range requiredOperations {
		if actual := actualOperations[key]; actual != "" && actual != operationID {
			v.errors = append(v.errors, fmt.Sprintf("%s must use operationId %s, got %s", key, operationID, actual))
		}
	}
}

func (v *validator) validateOperation(path string, method string, operation map[string]any, pathParameters []any) {
	operationID, _ := operation["operationId"].(string)
	v.validateSecurity(path, method, operation, operationID)
	v.validateIdempotency(path, method, operation, operationID, pathParameters)

	responses := asMap(operation["responses"])
	var success any
	for status, response := range responses {
		if regexp.MustCompile(`^2[0-9][0-9]$`).MatchString(status) {
			success = response
			break
		}
	}

	location := fmt.Sprintf("%s %s", strings.ToUpper(method), path)
	if success == nil {
		v.errors = append(v.errors, location+" has no 2xx response")
		return
	}

	resolvedResponse, responsePath := v.dereferenceWithPath(success, v.root)
	examples := dig(resolvedResponse, "content", "application/json", "examples")
	if len(asMap(examples)) == 0 {
		v.errors = append(v.errors, location+" has no success example")
	}

	successSchema := dig(resolvedResponse, "content", "application/json", "schema")
	v.validateEnvelope(successSchema, responsePath, "BaseResponse", location+" success")
	v.validateExamples(examples, successSchema, responsePath, location+" success")

	for status, response := range responses {
		if regexp.MustCompile(`^2[0-9][0-9]$`).MatchString(status) {
			continue
		}
		resolvedError, errorPath := v.dereferenceWithPath(response, v.root)
		errorSchema := dig(resolvedError, "content", "application/json", "schema")
		if errorSchema == nil {
			continue
		}
		errorLocation := fmt.Sprintf("%s %s %s", strings.ToUpper(method), path, status)
		v.validateEnvelope(errorSchema, errorPath, "ErrorResponse", errorLocation)
		v.validateExamples(dig(resolvedError, "content", "application/json", "examples"), errorSchema, errorPath, errorLocation)
	}

	requestBody := operation["requestBody"]
	if requestBody == nil {
		return
	}

	resolvedBody, bodyPath := v.dereferenceWithPath(requestBody, v.root)
	bodyExamples := dig(resolvedBody, "content", "application/json", "examples")
	if len(asMap(bodyExamples)) == 0 {
		v.errors = append(v.errors, location+" request body has no example")
	}
	v.validateExamples(bodyExamples, dig(resolvedBody, "content", "application/json", "schema"), bodyPath, location+" request")
}

func (v *validator) validateSecurity(path string, method string, operation map[string]any, operationID string) {
	security := v.rootSecurity
	if value, ok := operation["security"]; ok {
		security = value
	}

	var expected any
	switch {
	case publicOperations[operationID]:
		expected = []any{}
	case refreshOperations[operationID]:
		expected = []any{map[string]any{"refreshCookie": []any{}}}
	default:
		expected = []any{map[string]any{"bearerAuth": []any{}}}
	}
	if !jsonEqual(security, expected) {
		v.errors = append(v.errors, fmt.Sprintf("%s %s has incorrect security declaration", strings.ToUpper(method), path))
	}
}

func (v *validator) validateIdempotency(path string, method string, operation map[string]any, operationID string, pathParameters []any) {
	parameters := append([]any{}, pathParameters...)
	parameters = append(parameters, asSlice(operation["parameters"])...)

	hasKey := false
	for _, parameter := range parameters {
		ref, _ := asMap(parameter)["$ref"].(string)
		if strings.HasSuffix(ref, "/IdempotencyKey") {
			hasKey = true
			resolved, _ := v.dereferenceWithPath(parameter, v.root)
			if asMap(resolved)["required"] != true {
				v.errors = append(v.errors, fmt.Sprintf("%s %s Idempotency-Key must be required", strings.ToUpper(method), path))
			}
		}
	}

	required := idempotentOperations[operationID]
	if required && !hasKey {
		v.errors = append(v.errors, fmt.Sprintf("%s %s must require Idempotency-Key", strings.ToUpper(method), path))
	}
	if !required && hasKey {
		v.errors = append(v.errors, fmt.Sprintf("%s %s unexpectedly requires Idempotency-Key", strings.ToUpper(method), path))
	}
}

func (v *validator) validateEnvelope(schema any, schemaPath string, expected string, location string) {
	if schema == nil {
		v.errors = append(v.errors, location+" has no JSON schema")
		return
	}
	if !slices.Contains(v.referencedSchemaNames(schema, schemaPath, nil), expected) {
		v.errors = append(v.errors, location+" does not inherit "+expected)
	}
}

func (v *validator) referencedSchemaNames(node any, currentPath string, names []string) []string {
	switch value := node.(type) {
	case map[string]any:
		if ref, ok := value["$ref"].(string); ok {
			parts := strings.Split(ref, "/")
			names = append(names, parts[len(parts)-1])
			target, targetPath, _ := v.resolveReference(currentPath, ref)
			return v.referencedSchemaNames(target, targetPath, names)
		}
		for _, child := range value {
			names = v.referencedSchemaNames(child, currentPath, names)
		}
	case []any:
		for _, child := range value {
			names = v.referencedSchemaNames(child, currentPath, names)
		}
	}
	return names
}

func (v *validator) validateExamples(examples any, schema any, schemaPath string, location string) {
	if len(asMap(examples)) == 0 || schema == nil {
		return
	}
	for name, example := range asMap(examples) {
		resolvedExample, _ := v.dereferenceWithPath(example, schemaPath)
		value := resolvedExample
		if exampleMap := asMap(resolvedExample); len(exampleMap) > 0 {
			if inner, ok := exampleMap["value"]; ok {
				value = inner
			}
		}
		v.validateInstance(value, schema, schemaPath, location+" example "+name)
	}
}

func (v *validator) validateFinancialGuardVectors() {
	schemasPath := filepath.Join(filepath.Dir(v.root), "components", "schemas.yaml")
	schemas := asMap(v.loadDocument(schemasPath))
	transaction := asMap(asMap(schemas["Transaction"])["properties"])

	v.rejectVector("Transaction.quantity negative", "-1.00000000", transaction["quantity"], schemasPath)
	v.rejectVector("Transaction.grossAmount negative", map[string]any{"amount": "-1.00000000", "currency": "RUB"}, transaction["grossAmount"], schemasPath)

	transactionSchema := schemas["Transaction"]
	transactionExample := dig(v.loadDocument(filepath.Join(filepath.Dir(v.root), "examples", "transactions.json")), "createResponse", "value", "data")
	v.rejectVector("Transaction.id invalid UUID", withValue(transactionExample, "id", "not-a-uuid"), transactionSchema, schemasPath)
	v.rejectVector("Transaction.tradeDate invalid BusinessDate", withValue(transactionExample, "tradeDate", "2026-02-30"), transactionSchema, schemasPath)
	v.rejectVector("Transaction BUY without quantity", withValue(transactionExample, "quantity", nil), transactionSchema, schemasPath)
	v.rejectVector("Transaction DEPOSIT with asset fields", withValue(transactionExample, "transactionType", "DEPOSIT"), transactionSchema, schemasPath)

	incomeRequest := map[string]any{
		"transactionType": "DIVIDEND",
		"ticker":          "SBER",
		"quantity":        "10.00000000",
		"unitPrice":       map[string]any{"amount": "1.00000000", "currency": "RUB"},
		"grossAmount":     map[string]any{"amount": "100.00000000", "currency": "RUB"},
		"commission":      map[string]any{"amount": "0.00000000", "currency": "RUB"},
		"tax":             map[string]any{"amount": "0.00000000", "currency": "RUB"},
		"tradeDate":       "2026-06-19",
		"settlementDate":  nil,
	}
	v.rejectVector("CreateTransactionRequest DIVIDEND with unit price", incomeRequest, schemas["CreateTransactionRequest"], schemasPath)
	v.rejectVector("UpdateTransactionRequest DIVIDEND with unit price", map[string]any{"expectedRevision": 1, "reason": "correct income", "corrected": incomeRequest}, schemas["UpdateTransactionRequest"], schemasPath)

	reverseRequest := asMap(dig(v.loadDocument(filepath.Join(filepath.Dir(v.root), "examples", "transactions.json")), "deleteRequest", "value"))
	v.rejectVector("ReverseTransactionRequest without effectiveDate", withoutKey(reverseRequest, "effectiveDate"), schemas["ReverseTransactionRequest"], schemasPath)
	v.rejectVector("ReverseTransactionRequest invalid effectiveDate", withValue(reverseRequest, "effectiveDate", "2026-02-30"), schemas["ReverseTransactionRequest"], schemasPath)

	negativeDecimal := "-1.00000000"
	negativeMoney := map[string]any{"amount": negativeDecimal, "currency": "RUB"}
	v.rejectNegativeFields("AssetSummary", schemas["AssetSummary"], []string{"lotSize"}, negativeDecimal, schemasPath)
	v.rejectNegativeFields("AssetSummary", schemas["AssetSummary"], []string{"lastPrice"}, negativeMoney, schemasPath)
	v.rejectNegativeFields("AssetBase", schemas["AssetBase"], []string{"lotSize"}, negativeDecimal, schemasPath)
	v.rejectNegativeFields("AssetBase", schemas["AssetBase"], []string{"lastPrice"}, negativeMoney, schemasPath)
	v.rejectVector("StockAsset unknown property through unevaluatedProperties", withValue(dig(v.loadDocument(filepath.Join(filepath.Dir(v.root), "examples", "assets.json")), "assetResponse", "value", "data"), "unexpected", true), schemas["StockAsset"], schemasPath)
	v.rejectNegativeFields("BondDetails", schemas["BondDetails"], []string{"faceValue"}, negativeMoney, schemasPath)
	v.rejectNegativeFields("BondDetails", schemas["BondDetails"], []string{"couponRate"}, negativeDecimal, schemasPath)
	v.rejectNegativeFields("PurchasingPowerEquivalent", schemas["PurchasingPowerEquivalent"], []string{"unitPrice"}, negativeMoney, schemasPath)
	v.rejectNegativeFields("PurchasingPowerEquivalent", schemas["PurchasingPowerEquivalent"], []string{"quantity"}, negativeDecimal, schemasPath)
	v.rejectNegativeFields("PortfolioPosition", schemas["PortfolioPosition"], []string{"quantity", "weight"}, negativeDecimal, schemasPath)
	v.rejectNegativeFields("PortfolioPosition", schemas["PortfolioPosition"], []string{"weightedAverageCost", "marketPrice", "marketValue"}, negativeMoney, schemasPath)
	v.rejectNegativeFields("DividendEvent", schemas["DividendEvent"], []string{"amountPerUnit"}, negativeMoney, schemasPath)
	v.rejectNegativeFields("PortfolioSummary", schemas["PortfolioSummary"], []string{"stockValue", "bondValue", "investedCapital", "dividendsReceived", "couponsReceived"}, negativeMoney, schemasPath)
	v.rejectNegativeFields("PortfolioSnapshot", schemas["PortfolioSnapshot"], []string{"stockValue", "bondValue", "investedCapital"}, negativeMoney, schemasPath)
	v.rejectNegativeFields("Dashboard", schemas["Dashboard"], []string{"dividendsReceived", "expectedDividends"}, negativeMoney, schemasPath)

	calculator := asMap(asMap(schemas["DividendCalculationRequest"])["properties"])
	v.rejectVector("DividendCalculationRequest.quantity zero", "0.00000000", calculator["quantity"], schemasPath)
	v.rejectVector("DividendCalculationRequest.quantity negative", "-1.00000000", calculator["quantity"], schemasPath)
	v.rejectVector("DividendCalculationRequest.dividendPerUnit negative", negativeMoney, calculator["dividendPerUnit"], schemasPath)
	v.rejectVector("DividendCalculationRequest.positionCost zero", map[string]any{"amount": "0.00000000", "currency": "RUB"}, calculator["positionCost"], schemasPath)

	resultSchema := schemas["DividendCalculation"]
	resultExample := dig(v.loadDocument(filepath.Join(filepath.Dir(v.root), "examples", "dividends.json")), "calculateResponse", "value", "data")
	v.rejectVector("DividendCalculation yield without position cost", withValue(resultExample, "positionCost", nil), resultSchema, schemasPath)
	v.rejectVector("DividendCalculation null yield with position cost", withValue(resultExample, "grossYield", nil), resultSchema, schemasPath)

	traceparent := dig(v.loadDocument(v.root), "components", "parameters", "Traceparent", "schema")
	v.rejectVector("traceparent version ff", "ff-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01", traceparent, v.root)
	v.rejectVector("traceparent all-zero trace id", "00-00000000000000000000000000000000-00f067aa0ba902b7-01", traceparent, v.root)
	v.rejectVector("traceparent all-zero parent id", "00-4bf92f3577b34da6a3ce929d0e0e4736-0000000000000000-01", traceparent, v.root)
}

func (v *validator) rejectNegativeFields(schemaName string, schema any, fields []string, value any, schemaPath string) {
	properties := asMap(asMap(schema)["properties"])
	for _, field := range fields {
		v.rejectVector(schemaName+"."+field+" negative", value, properties[field], schemaPath)
	}
}

func (v *validator) rejectVector(name string, value any, schema any, schemaPath string) {
	if v.schemaMatches(value, schema, schemaPath) {
		v.errors = append(v.errors, "invalid financial vector accepted: "+name)
	}
}

func (v *validator) validateExampleSourceCodes() {
	var observed []string
	paths, _ := filepath.Glob(filepath.Join(filepath.Dir(v.root), "examples", "*.json"))
	slices.Sort(paths)
	for _, path := range paths {
		collectSourceCodes(v.loadDocument(path), &observed)
	}

	unexpected := stringDifference(unique(observed), exampleSourceCodes)
	missing := stringDifference(exampleSourceCodes, unique(observed))
	if len(unexpected) > 0 {
		v.errors = append(v.errors, "examples use non-reserved source codes: "+strings.Join(unexpected, ", "))
	}
	if len(missing) > 0 {
		v.errors = append(v.errors, "reserved example source codes are unused: "+strings.Join(missing, ", "))
	}
}

func (v *validator) validateImportExamples() {
	examplesPath := filepath.Join(filepath.Dir(v.root), "examples", "imports.json")
	examples := v.loadDocument(examplesPath)
	reviewRequest := asMap(dig(examples, "reviewRequest", "value"))
	appendRequest := asMap(dig(examples, "appendRequest", "value"))
	reviewResponse := asMap(dig(examples, "reviewResponse", "value", "data"))
	appendResponse := asMap(dig(examples, "appendResponse", "value", "data"))

	payload, _ := reviewRequest["csvPayload"].(string)
	appendPayload, _ := appendRequest["csvPayload"].(string)
	portfolioID, _ := reviewResponse["portfolioId"].(string)
	sourceAccountLabel, _ := reviewRequest["sourceAccountLabel"].(string)
	if payload == "" || appendPayload == "" || portfolioID == "" {
		v.errors = append(v.errors, "import examples must include csvPayload and portfolioId")
		return
	}
	if appendPayload != payload {
		v.errors = append(v.errors, "import append example csvPayload must match review csvPayload for stateless review-to-append flow")
	}

	hash := sha256.Sum256([]byte(payload))
	sourceFileHash := hex.EncodeToString(hash[:])
	if got, _ := reviewResponse["sourceFileHash"].(string); got != sourceFileHash {
		v.errors = append(v.errors, "import review example sourceFileHash does not match csvPayload")
	}
	if got, _ := appendResponse["sourceFileHash"].(string); got != sourceFileHash {
		v.errors = append(v.errors, "import append example sourceFileHash does not match csvPayload")
	}

	review, err := importer.ReviewCSV(importer.ReviewRequest{
		SubjectID:          "00000000-0000-4000-8000-000000000001",
		PortfolioID:        portfolioID,
		SourceKind:         importer.SourceKindUserUploadedFile,
		SourceAccountLabel: sourceAccountLabel,
		FileHash:           sourceFileHash,
		Reader:             strings.NewReader(payload),
	})
	if err != nil {
		v.errors = append(v.errors, "import review example cannot be parsed by importer: "+err.Error())
		return
	}

	exampleRows := asSlice(reviewResponse["rows"])
	if len(exampleRows) != len(review.Rows) {
		v.errors = append(v.errors, fmt.Sprintf("import review example row count mismatch: got %d want %d", len(exampleRows), len(review.Rows)))
		return
	}
	for index, row := range review.Rows {
		exampleRow := asMap(exampleRows[index])
		if got, _ := asInt(exampleRow["rowNumber"]); got != row.RowNumber {
			v.errors = append(v.errors, fmt.Sprintf("import example row %d rowNumber mismatch", index+1))
		}
		if got, _ := exampleRow["rowHash"].(string); got != row.RowHash {
			v.errors = append(v.errors, fmt.Sprintf("import example row %d rowHash mismatch", index+1))
		}
		if got, _ := exampleRow["fingerprint"].(string); got != row.Fingerprint {
			v.errors = append(v.errors, fmt.Sprintf("import example row %d fingerprint mismatch", index+1))
		}
		if got, _ := exampleRow["status"].(string); got != row.Status {
			v.errors = append(v.errors, fmt.Sprintf("import example row %d status mismatch", index+1))
		}
	}
}

func collectSourceCodes(node any, codes *[]string) {
	switch value := node.(type) {
	case map[string]any:
		if source := asMap(value["source"]); len(source) > 0 {
			if code, ok := source["code"].(string); ok {
				*codes = append(*codes, code)
			}
		}
		for _, child := range value {
			collectSourceCodes(child, codes)
		}
	case []any:
		for _, child := range value {
			collectSourceCodes(child, codes)
		}
	}
}

func (v *validator) validateInstance(value any, schema any, schemaPath string, location string) {
	if schema == nil {
		return
	}
	resolved, resolvedPath := v.dereferenceWithPath(schema, schemaPath)
	resolvedMap := asMap(resolved)

	for _, part := range asSlice(resolvedMap["allOf"]) {
		v.validateInstance(value, part, resolvedPath, location)
	}
	if condition := resolvedMap["if"]; condition != nil {
		branch := resolvedMap["else"]
		if v.schemaMatches(value, condition, resolvedPath) {
			branch = resolvedMap["then"]
		}
		if branch != nil {
			v.validateInstance(value, branch, resolvedPath, location)
		}
	}
	if oneOf := asSlice(resolvedMap["oneOf"]); len(oneOf) > 0 {
		matches := 0
		for _, part := range oneOf {
			if v.schemaMatches(value, part, resolvedPath) {
				matches++
			}
		}
		if matches != 1 {
			v.errors = append(v.errors, fmt.Sprintf("%s must match exactly one oneOf branch (matched %d)", location, matches))
		}
		return
	}

	if types := asStringSlice(resolvedMap["type"]); len(types) > 0 && !slices.ContainsFunc(types, func(kind string) bool { return typeMatches(value, kind) }) {
		v.errors = append(v.errors, fmt.Sprintf("%s has type %T, expected %s", location, value, strings.Join(types, " or ")))
		return
	}
	if enum := asSlice(resolvedMap["enum"]); len(enum) > 0 && !containsValue(enum, value) {
		v.errors = append(v.errors, location+" is not in enum")
	}
	if constant, ok := resolvedMap["const"]; ok && !jsonEqual(constant, value) {
		v.errors = append(v.errors, location+" does not equal const")
	}

	switch typed := value.(type) {
	case map[string]any:
		properties := asMap(resolvedMap["properties"])
		if missing := stringDifference(asStringSlice(resolvedMap["required"]), mapKeys(typed)); len(missing) > 0 {
			v.errors = append(v.errors, location+" misses required properties: "+strings.Join(missing, ", "))
		}
		if resolvedMap["additionalProperties"] == false {
			if extras := stringDifference(mapKeys(typed), mapKeys(properties)); len(extras) > 0 {
				v.errors = append(v.errors, location+" has unknown properties: "+strings.Join(extras, ", "))
			}
		}
		if resolvedMap["unevaluatedProperties"] == false {
			if extras := stringDifference(mapKeys(typed), v.evaluatedProperties(resolvedMap, resolvedPath)); len(extras) > 0 {
				v.errors = append(v.errors, location+" has unevaluated properties: "+strings.Join(extras, ", "))
			}
		}
		for key, child := range typed {
			if property := properties[key]; property != nil {
				v.validateInstance(child, property, resolvedPath, location+"."+key)
			}
		}
		if minProperties, ok := asInt(resolvedMap["minProperties"]); ok && len(typed) < minProperties {
			v.errors = append(v.errors, fmt.Sprintf("%s has fewer than %d properties", location, minProperties))
		}
	case []any:
		if items := resolvedMap["items"]; items != nil {
			for index, child := range typed {
				v.validateInstance(child, items, resolvedPath, fmt.Sprintf("%s[%d]", location, index))
			}
		}
	case string:
		if minLength, ok := asInt(resolvedMap["minLength"]); ok && len(typed) < minLength {
			v.errors = append(v.errors, location+" is shorter than minLength")
		}
		if maxLength, ok := asInt(resolvedMap["maxLength"]); ok && len(typed) > maxLength {
			v.errors = append(v.errors, location+" is longer than maxLength")
		}
		if pattern, ok := resolvedMap["pattern"].(string); ok && !patternMatches(pattern, typed) {
			v.errors = append(v.errors, location+" does not match pattern")
		}
		if format, ok := resolvedMap["format"].(string); ok {
			v.validateStringFormat(typed, format, location)
		}
	case int, int64, float64:
		number, _ := asNumber(typed)
		if minimum, ok := asNumber(resolvedMap["minimum"]); ok && number < minimum {
			v.errors = append(v.errors, location+" is below minimum")
		}
		if maximum, ok := asNumber(resolvedMap["maximum"]); ok && number > maximum {
			v.errors = append(v.errors, location+" is above maximum")
		}
	}
	if notSchema := resolvedMap["not"]; notSchema != nil && v.schemaMatches(value, notSchema, resolvedPath) {
		v.errors = append(v.errors, location+" matches a forbidden schema")
	}
}

func (v *validator) validateStringFormat(value string, format string, location string) {
	switch format {
	case "uuid":
		if !regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`).MatchString(value) {
			v.errors = append(v.errors, location+" is not a valid uuid")
		}
	case "date":
		if !validFullDate(value) {
			v.errors = append(v.errors, location+" is not a valid RFC 3339 full-date")
		}
	case "date-time":
		if !strings.HasSuffix(value, "Z") {
			v.errors = append(v.errors, location+" is not a valid UTC date-time")
			return
		}
		if _, err := time.Parse(time.RFC3339, value); err != nil {
			v.errors = append(v.errors, location+" is not a valid date-time")
		}
	}
}

func (v *validator) evaluatedProperties(schema map[string]any, schemaPath string) []string {
	resolved, resolvedPath := v.dereferenceWithPath(schema, schemaPath)
	resolvedMap := asMap(resolved)
	names := mapKeys(asMap(resolvedMap["properties"]))
	for _, part := range asSlice(resolvedMap["allOf"]) {
		names = append(names, v.evaluatedProperties(asMap(part), resolvedPath)...)
	}
	return unique(names)
}

func (v *validator) schemaMatches(value any, schema any, schemaPath string) bool {
	before := len(v.errors)
	v.validateInstance(value, schema, schemaPath, "candidate")
	matched := len(v.errors) == before
	if !matched {
		v.errors = v.errors[:before]
	}
	return matched
}

func (v *validator) dereferenceWithPath(object any, currentPath string) (any, string) {
	if ref, ok := asMap(object)["$ref"].(string); ok {
		target, targetPath, _ := v.resolveReference(currentPath, ref)
		return v.dereferenceWithPath(target, targetPath)
	}
	return object, currentPath
}

func (v *validator) walk(node any, currentPath string, location string) {
	switch value := node.(type) {
	case map[string]any:
		if ref, ok := value["$ref"].(string); ok {
			target, targetPath, key := v.resolveReference(currentPath, ref)
			if !v.visitedRefs[key] {
				v.visitedRefs[key] = true
				v.walk(target, targetPath, key)
			}
		}
		for key, child := range value {
			if key != "$ref" {
				v.walk(child, currentPath, location+"/"+key)
			}
		}
	case []any:
		for index, child := range value {
			v.walk(child, currentPath, fmt.Sprintf("%s/%d", location, index))
		}
	}
}

func (v *validator) resolveReference(currentPath string, reference string) (any, string, string) {
	v.referenceCount++
	parts := strings.SplitN(reference, "#", 2)
	filePart := parts[0]
	fragment := ""
	if len(parts) == 2 {
		fragment = parts[1]
	}
	targetPath := currentPath
	if filePart != "" {
		targetPath = filepath.Clean(filepath.Join(filepath.Dir(currentPath), filePart))
	}
	document := v.loadDocument(targetPath)
	pointer := []string{}
	if fragment != "" {
		pointer = strings.Split(strings.TrimPrefix(fragment, "/"), "/")
	}
	target := document
	for _, rawToken := range pointer {
		token := strings.ReplaceAll(strings.ReplaceAll(rawToken, "~1", "/"), "~0", "~")
		targetMap := asMap(target)
		next, ok := targetMap[token]
		if !ok {
			v.errors = append(v.errors, fmt.Sprintf("unresolved reference %s from %s", reference, currentPath))
			target = map[string]any{}
			break
		}
		target = next
	}
	absolute, _ := filepath.Abs(targetPath)
	return target, absolute, absolute + "#" + fragment
}

func (v *validator) loadDocument(path string) any {
	normalized, _ := filepath.Abs(filepath.Clean(path))
	if document, ok := v.documents[normalized]; ok {
		return document
	}
	content, err := os.ReadFile(normalized)
	if err != nil {
		v.errors = append(v.errors, "missing referenced file "+normalized)
		return map[string]any{}
	}
	var document any
	if filepath.Ext(normalized) == ".json" {
		err = json.Unmarshal(content, &document)
	} else {
		err = yaml.Unmarshal(content, &document)
	}
	if err != nil {
		v.errors = append(v.errors, fmt.Sprintf("cannot parse %s: %v", normalized, err))
		return map[string]any{}
	}
	document = normalize(document)
	v.documents[normalized] = document
	return document
}

func normalize(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		result := map[string]any{}
		for key, child := range typed {
			result[key] = normalize(child)
		}
		return result
	case []any:
		for index, child := range typed {
			typed[index] = normalize(child)
		}
	}
	return value
}

func asMap(value any) map[string]any {
	if typed, ok := value.(map[string]any); ok {
		return typed
	}
	return map[string]any{}
}

func asSlice(value any) []any {
	if typed, ok := value.([]any); ok {
		return typed
	}
	return nil
}

func asStringSlice(value any) []string {
	switch typed := value.(type) {
	case string:
		return []string{typed}
	case []any:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			if text, ok := item.(string); ok {
				result = append(result, text)
			}
		}
		return result
	}
	return nil
}

func asInt(value any) (int, bool) {
	switch typed := value.(type) {
	case int:
		return typed, true
	case int64:
		return int(typed), true
	case float64:
		return int(typed), true
	}
	return 0, false
}

func asNumber(value any) (float64, bool) {
	switch typed := value.(type) {
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case float64:
		return typed, true
	}
	return 0, false
}

func mapKeys[T any](source map[string]T) []string {
	result := make([]string, 0, len(source))
	for key := range source {
		result = append(result, key)
	}
	slices.Sort(result)
	return result
}

func unique(source []string) []string {
	seen := map[string]bool{}
	var result []string
	for _, item := range source {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	slices.Sort(result)
	return result
}

func stringDifference(left []string, right []string) []string {
	seen := map[string]bool{}
	for _, item := range right {
		seen[item] = true
	}
	var result []string
	for _, item := range left {
		if !seen[item] {
			result = append(result, item)
		}
	}
	slices.Sort(result)
	return result
}

func duplicates(source []string) []string {
	counts := map[string]int{}
	for _, item := range source {
		counts[item]++
	}
	var result []string
	for item, count := range counts {
		if count > 1 {
			result = append(result, item)
		}
	}
	slices.Sort(result)
	return result
}

func stringSet(values ...string) map[string]bool {
	result := map[string]bool{}
	for _, value := range values {
		result[value] = true
	}
	return result
}

func dig(value any, path ...string) any {
	current := value
	for _, key := range path {
		current = asMap(current)[key]
		if current == nil {
			return nil
		}
	}
	return current
}

func withValue(value any, key string, replacement any) map[string]any {
	copy := cloneMap(value)
	copy[key] = replacement
	return copy
}

func withoutKey(value map[string]any, key string) map[string]any {
	copy := cloneMap(value)
	delete(copy, key)
	return copy
}

func cloneMap(value any) map[string]any {
	content, _ := json.Marshal(value)
	var result map[string]any
	_ = json.Unmarshal(content, &result)
	return result
}

func typeMatches(value any, kind string) bool {
	switch kind {
	case "object":
		_, ok := value.(map[string]any)
		return ok
	case "array":
		_, ok := value.([]any)
		return ok
	case "string":
		_, ok := value.(string)
		return ok
	case "integer":
		switch typed := value.(type) {
		case int, int64:
			return true
		case float64:
			return typed == float64(int64(typed))
		}
	case "number":
		_, ok := asNumber(value)
		return ok
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "null":
		return value == nil
	}
	return false
}

func containsValue(values []any, expected any) bool {
	for _, value := range values {
		if jsonEqual(value, expected) {
			return true
		}
	}
	return false
}

func jsonEqual(left any, right any) bool {
	leftJSON, _ := json.Marshal(normalize(left))
	rightJSON, _ := json.Marshal(normalize(right))
	return string(leftJSON) == string(rightJSON)
}

func validFullDate(value string) bool {
	if !regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`).MatchString(value) {
		return false
	}
	parsed, err := time.Parse("2006-01-02", value)
	return err == nil && parsed.Format("2006-01-02") == value
}

func patternMatches(pattern string, value string) bool {
	if pattern == `^(?!ff)[0-9a-f]{2}-(?!0{32})[0-9a-f]{32}-(?!0{16})[0-9a-f]{16}-[0-9a-f]{2}$` {
		if !regexp.MustCompile(`^[0-9a-f]{2}-[0-9a-f]{32}-[0-9a-f]{16}-[0-9a-f]{2}$`).MatchString(value) {
			return false
		}
		parts := strings.Split(value, "-")
		return parts[0] != "ff" && parts[1] != strings.Repeat("0", 32) && parts[2] != strings.Repeat("0", 16)
	}
	compiled, err := regexp.Compile(pattern)
	return err == nil && compiled.MatchString(value)
}
