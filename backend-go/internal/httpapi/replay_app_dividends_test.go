package httpapi

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestReplayProductionConstructorRegistersDividendCalculatorRoute(t *testing.T) {
	store := &dividendHTTPTestStore{}
	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	app, err := NewReplay(service, nil, []byte("openinvest-test-import-review-token-secret-32bytes"))
	if err != nil {
		t.Fatalf("NewReplay: %v", err)
	}

	body := []byte(`{"ticker":"SBER","quantity":"1000.00000000","dividendPerUnit":{"amount":"34.84000000","currency":"RUB"},"positionCost":{"amount":"280000.00000000","currency":"RUB"}}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/dividends/calculate", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "audit-remediation-007-dividend-route-000001")

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("status=%d, want %d; a 404 means the shipped replay route regressed", response.StatusCode, http.StatusOK)
	}
	if store.writeCalls != 1 {
		t.Fatalf("calculator replay write calls=%d, want 1", store.writeCalls)
	}
}

func TestReplayConstructorsEachPreserveDividendCalculatorLimiter(t *testing.T) {
	node, err := parser.ParseFile(token.NewFileSet(), "replay_app.go", nil, 0)
	if err != nil {
		t.Fatalf("parse replay_app.go: %v", err)
	}

	for _, functionName := range []string{
		"NewReplayWithCorporateActionProvider",
		"NewDevelopmentReplayWithCorporateActionProvider",
	} {
		t.Run(functionName, func(t *testing.T) {
			function := findReplayConstructor(t, node, functionName)
			if got := countReplayConstructorCalls(function.Body, "newDividendCalculatorRateLimiter"); got != 1 {
				t.Fatalf("%s newDividendCalculatorRateLimiter calls=%d, want exactly 1", functionName, got)
			}
		})
	}
}

func findReplayConstructor(t *testing.T, node *ast.File, name string) *ast.FuncDecl {
	t.Helper()
	var found []*ast.FuncDecl
	for _, declaration := range node.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Name == nil || function.Name.Name != name {
			continue
		}
		found = append(found, function)
	}
	if len(found) != 1 {
		t.Fatalf("constructor %s declarations=%d, want exactly 1", name, len(found))
	}
	return found[0]
}

func countReplayConstructorCalls(node ast.Node, name string) int {
	count := 0
	ast.Inspect(node, func(current ast.Node) bool {
		call, ok := current.(*ast.CallExpr)
		if !ok {
			return true
		}
		identifier, ok := call.Fun.(*ast.Ident)
		if ok && identifier.Name == name {
			count++
		}
		return true
	})
	return count
}
