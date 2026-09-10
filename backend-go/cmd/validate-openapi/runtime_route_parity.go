package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
)

const runtimeStatusExtension = "x-openinvest-runtime-status"
const runtimeStatusPlanned = "planned"

var allowedPlannedRuntimeOperations = map[string]bool{
	"PATCH /api/v1/portfolios/{portfolioId}":         true,
	"DELETE /api/v1/portfolios/{portfolioId}":        true,
	"GET /api/v1/portfolios/{portfolioId}/snapshots": true,
	"GET /api/v1/dividends/calendar":                 true,
	"GET /api/v1/dashboard":                          true,
}

var fiberRouteMethods = map[string]string{
	"Get":     "GET",
	"Post":    "POST",
	"Put":     "PUT",
	"Patch":   "PATCH",
	"Delete":  "DELETE",
	"Options": "OPTIONS",
	"Head":    "HEAD",
	"Trace":   "TRACE",
}

func (v *validator) validateRuntimeRouteParity(document map[string]any) {
	v.validateRuntimeRouteParityAgainstFile(document, filepath.Join("internal", "httpapi", "replay_app.go"))
}

func (v *validator) validateRuntimeRouteParityAgainstFile(document map[string]any, filename string) {
	implemented, planned, classificationErrors := openAPIRuntimeRouteKeys(document)
	v.errors = append(v.errors, classificationErrors...)

	actual, err := replayRouteKeys(filename)
	if err != nil {
		v.errors = append(v.errors, "inspect production replay routes: "+err.Error())
		return
	}

	if missing := stringDifference(implemented, actual); len(missing) > 0 {
		v.errors = append(v.errors, "production replay router is missing implemented OpenAPI operations: "+strings.Join(missing, ", "))
	}
	if shippedPlanned := stringIntersection(planned, actual); len(shippedPlanned) > 0 {
		v.errors = append(v.errors, "production replay router ships OpenAPI operations still marked planned: "+strings.Join(shippedPlanned, ", "))
	}
	known := append(append([]string{}, implemented...), planned...)
	if unexpected := stringDifference(actual, known); len(unexpected) > 0 {
		v.errors = append(v.errors, "production replay router exposes operations absent from OpenAPI contract: "+strings.Join(unexpected, ", "))
	}
}

func openAPIRuntimeRouteKeys(document map[string]any) (implemented []string, planned []string, problems []string) {
	paths := asMap(document["paths"])
	for path, rawPathItem := range paths {
		pathItem := asMap(rawPathItem)
		for _, method := range httpMethods {
			rawOperation, ok := pathItem[method]
			if !ok {
				continue
			}
			operation := asMap(rawOperation)
			routeKey := fmt.Sprintf("%s %s", strings.ToUpper(method), path)
			rawStatus, hasStatus := operation[runtimeStatusExtension]
			if !hasStatus {
				implemented = append(implemented, routeKey)
				continue
			}
			status, ok := rawStatus.(string)
			if !ok || status != runtimeStatusPlanned {
				problems = append(problems, fmt.Sprintf(
					"%s must omit %s or set it exactly to %q",
					routeKey,
					runtimeStatusExtension,
					runtimeStatusPlanned,
				))
				continue
			}
			if !allowedPlannedRuntimeOperations[routeKey] {
				problems = append(problems, fmt.Sprintf(
					"%s is not an approved frozen runtime reservation and cannot be marked %s=%q",
					routeKey,
					runtimeStatusExtension,
					runtimeStatusPlanned,
				))
				continue
			}
			planned = append(planned, routeKey)
		}
	}
	return unique(implemented), unique(planned), problems
}

func replayRouteKeys(filename string) ([]string, error) {
	node, err := parser.ParseFile(token.NewFileSet(), filename, nil, 0)
	if err != nil {
		return nil, err
	}

	candidates := make([]*ast.FuncDecl, 0, 1)
	for _, declaration := range node.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Name == nil || function.Name.Name != "newReplayApp" {
			continue
		}
		candidates = append(candidates, function)
	}
	if len(candidates) != 1 {
		return nil, fmt.Errorf("expected exactly one newReplayApp function, found %d", len(candidates))
	}
	function := candidates[0]
	if err := validateNewReplayAppSignature(function); err != nil {
		return nil, err
	}

	routes := []string{}
	var inspectErr error
	ast.Inspect(function.Body, func(n ast.Node) bool {
		if inspectErr != nil {
			return false
		}
		if _, nested := n.(*ast.FuncLit); nested {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		receiver, ok := selector.X.(*ast.Ident)
		if !ok || receiver.Name != "app" {
			return true
		}
		method, ok := fiberRouteMethods[selector.Sel.Name]
		if !ok {
			return true
		}
		if len(call.Args) < 2 {
			inspectErr = fmt.Errorf("%s route registration in newReplayApp has fewer than two arguments", selector.Sel.Name)
			return false
		}
		literal, ok := call.Args[0].(*ast.BasicLit)
		if !ok || literal.Kind != token.STRING {
			inspectErr = fmt.Errorf("%s route registration in newReplayApp must use a string-literal path", selector.Sel.Name)
			return false
		}
		path, err := strconv.Unquote(literal.Value)
		if err != nil {
			inspectErr = fmt.Errorf("decode %s route path in newReplayApp: %w", selector.Sel.Name, err)
			return false
		}
		routes = append(routes, method+" "+openAPIPathFromFiber(path))
		return true
	})
	if inspectErr != nil {
		return nil, inspectErr
	}
	if duplicate := duplicates(routes); len(duplicate) > 0 {
		return nil, fmt.Errorf("duplicate production replay route registration: %s", strings.Join(duplicate, ", "))
	}
	return unique(routes), nil
}

func validateNewReplayAppSignature(function *ast.FuncDecl) error {
	if function == nil || function.Recv != nil || function.Type == nil || function.Body == nil {
		return fmt.Errorf("newReplayApp has unsupported declaration structure")
	}
	if function.Type.Params == nil || len(function.Type.Params.List) != 1 {
		return fmt.Errorf("newReplayApp must have exactly one parameter")
	}
	parameter := function.Type.Params.List[0]
	if len(parameter.Names) != 1 || parameter.Names[0].Name != "api" || !isPointerToIdent(parameter.Type, "API") {
		return fmt.Errorf("newReplayApp must declare exactly parameter api *API")
	}
	if function.Type.Results == nil || len(function.Type.Results.List) != 1 {
		return fmt.Errorf("newReplayApp must return exactly one result")
	}
	result := function.Type.Results.List[0]
	if len(result.Names) != 0 || !isPointerToSelector(result.Type, "fiber", "App") {
		return fmt.Errorf("newReplayApp must return *fiber.App")
	}
	return nil
}

func isPointerToIdent(expression ast.Expr, name string) bool {
	pointer, ok := expression.(*ast.StarExpr)
	if !ok {
		return false
	}
	identifier, ok := pointer.X.(*ast.Ident)
	return ok && identifier.Name == name
}

func isPointerToSelector(expression ast.Expr, packageName string, selectorName string) bool {
	pointer, ok := expression.(*ast.StarExpr)
	if !ok {
		return false
	}
	selector, ok := pointer.X.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	packageIdentifier, ok := selector.X.(*ast.Ident)
	return ok && packageIdentifier.Name == packageName && selector.Sel.Name == selectorName
}
func openAPIPathFromFiber(path string) string {
	segments := strings.Split(path, "/")
	for index, segment := range segments {
		if strings.HasPrefix(segment, ":") && len(segment) > 1 {
			segments[index] = "{" + strings.TrimPrefix(segment, ":") + "}"
		}
	}
	return strings.Join(segments, "/")
}

func stringIntersection(left []string, right []string) []string {
	seen := map[string]bool{}
	for _, item := range right {
		seen[item] = true
	}
	intersection := make([]string, 0)
	for _, item := range left {
		if seen[item] {
			intersection = append(intersection, item)
		}
	}
	return unique(intersection)
}
