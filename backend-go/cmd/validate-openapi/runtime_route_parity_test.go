package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestOpenAPIPathFromFiberNormalizesNamedSegments(t *testing.T) {
	got := openAPIPathFromFiber("/api/v1/portfolios/:portfolioId/transactions/:transactionId")
	want := "/api/v1/portfolios/{portfolioId}/transactions/{transactionId}"
	if got != want {
		t.Fatalf("normalized path=%q, want %q", got, want)
	}
}

func TestOpenAPIRuntimeRouteKeysSeparatesImplementedAndPlanned(t *testing.T) {
	document := map[string]any{
		"paths": map[string]any{
			"/api/v1/health": map[string]any{
				"get": map[string]any{"operationId": "getHealth"},
			},
			"/api/v1/dashboard": map[string]any{
				"get": map[string]any{
					"operationId":          "getDashboard",
					runtimeStatusExtension: runtimeStatusPlanned,
				},
			},
		},
	}

	implemented, planned, problems := openAPIRuntimeRouteKeys(document)
	if len(problems) != 0 {
		t.Fatalf("unexpected classification problems: %v", problems)
	}
	if !reflect.DeepEqual(implemented, []string{"GET /api/v1/health"}) {
		t.Fatalf("implemented=%v", implemented)
	}
	if !reflect.DeepEqual(planned, []string{"GET /api/v1/dashboard"}) {
		t.Fatalf("planned=%v", planned)
	}
}

func TestOpenAPIRuntimeRouteKeysRejectsUnknownLifecycleStatus(t *testing.T) {
	document := map[string]any{
		"paths": map[string]any{
			"/api/v1/dashboard": map[string]any{
				"get": map[string]any{
					"operationId":          "getDashboard",
					runtimeStatusExtension: "later",
				},
			},
		},
	}

	implemented, planned, problems := openAPIRuntimeRouteKeys(document)
	if len(implemented) != 0 || len(planned) != 0 || len(problems) != 1 {
		t.Fatalf("implemented=%v planned=%v problems=%v", implemented, planned, problems)
	}
}

func TestOpenAPIRuntimeRouteKeysRejectsUnapprovedPlannedOperation(t *testing.T) {
	document := map[string]any{
		"paths": map[string]any{
			"/api/v1/health": map[string]any{
				"get": map[string]any{
					"operationId":          "getHealth",
					runtimeStatusExtension: runtimeStatusPlanned,
				},
			},
		},
	}

	implemented, planned, problems := openAPIRuntimeRouteKeys(document)
	if len(implemented) != 0 || len(planned) != 0 || len(problems) != 1 {
		t.Fatalf("implemented=%v planned=%v problems=%v", implemented, planned, problems)
	}
	if !strings.Contains(problems[0], "not an approved frozen runtime reservation") {
		t.Fatalf("unexpected problem: %v", problems)
	}
}

func TestReplayRouteKeysDetectsRouteOnlyInsideNewReplayApp(t *testing.T) {
	path := writeReplayFixture(t, `package fixture
func newReplayApp(api *API) *fiber.App {
	app := fiber.New()
	app.Post("/api/v1/dividends/calculate", handler)
	return app
}`)

	got, err := replayRouteKeys(path)
	if err != nil {
		t.Fatalf("replayRouteKeys: %v", err)
	}
	want := []string{"POST /api/v1/dividends/calculate"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("routes=%v, want %v", got, want)
	}
}

func TestReplayRouteKeysIgnoresDeadHelperRoutes(t *testing.T) {
	path := writeReplayFixture(t, `package fixture
func newReplayApp(api *API) *fiber.App {
	app := fiber.New()
	return app
}
func unusedHelper() {
	app.Post("/api/v1/dividends/calculate", handler)
}`)

	got, err := replayRouteKeys(path)
	if err != nil {
		t.Fatalf("replayRouteKeys: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("dead/helper route leaked into production route set: %v", got)
	}
}

func TestRuntimeParityStillReportsMissingRouteWhenOnlyDeadHelperRegistersIt(t *testing.T) {
	path := writeReplayFixture(t, `package fixture
func newReplayApp(api *API) *fiber.App {
	app := fiber.New()
	return app
}
func unusedHelper() {
	app.Post("/api/v1/dividends/calculate", handler)
}`)
	document := map[string]any{
		"paths": map[string]any{
			"/api/v1/dividends/calculate": map[string]any{
				"post": map[string]any{"operationId": "calculateDividend"},
			},
		},
	}

	v := &validator{}
	v.validateRuntimeRouteParityAgainstFile(document, path)
	if len(v.errors) != 1 || !strings.Contains(v.errors[0], "missing implemented") {
		t.Fatalf("errors=%v", v.errors)
	}
}

func TestReplayRouteKeysFailsClosedWithoutNewReplayApp(t *testing.T) {
	path := writeReplayFixture(t, `package fixture
func unusedHelper() {
	app.Post("/api/v1/dividends/calculate", handler)
}`)

	if _, err := replayRouteKeys(path); err == nil || !strings.Contains(err.Error(), "expected exactly one newReplayApp") {
		t.Fatalf("expected missing newReplayApp failure, got %v", err)
	}
}

func TestReplayRouteKeysFailsClosedWithDuplicateNewReplayApp(t *testing.T) {
	path := writeReplayFixture(t, `package fixture
func newReplayApp(api *API) *fiber.App { return nil }
func newReplayApp(api *API) *fiber.App { return nil }
`)

	if _, err := replayRouteKeys(path); err == nil || !strings.Contains(err.Error(), "expected exactly one newReplayApp") {
		t.Fatalf("expected duplicate newReplayApp failure, got %v", err)
	}
}

func TestReplayRouteKeysFailsClosedOnUnsupportedNewReplayAppSignature(t *testing.T) {
	path := writeReplayFixture(t, `package fixture
func newReplayApp(other *API) *fiber.App { return nil }
`)

	if _, err := replayRouteKeys(path); err == nil || !strings.Contains(err.Error(), "parameter api *API") {
		t.Fatalf("expected unsupported signature failure, got %v", err)
	}
}

func TestReplayRouteKeysRejectsDuplicateProductionRoutes(t *testing.T) {
	path := writeReplayFixture(t, `package fixture
func newReplayApp(api *API) *fiber.App {
	app := fiber.New()
	app.Get("/api/v1/portfolios/:portfolioId", handler)
	app.Get("/api/v1/portfolios/:portfolioId", handler)
	return app
}`)

	if _, err := replayRouteKeys(path); err == nil || !strings.Contains(err.Error(), "duplicate production replay route") {
		t.Fatalf("expected duplicate route failure, got %v", err)
	}
}

func TestReplayRouteKeysIgnoresNestedFunctionLiteralRoutes(t *testing.T) {
	path := writeReplayFixture(t, `package fixture
func newReplayApp(api *API) *fiber.App {
	app := fiber.New()
	unused := func() {
		app.Post("/api/v1/dividends/calculate", handler)
	}
	_ = unused
	return app
}`)

	got, err := replayRouteKeys(path)
	if err != nil {
		t.Fatalf("replayRouteKeys: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("nested function literal route leaked into production route set: %v", got)
	}
}

func TestRuntimeParityRejectsShippedPlannedOperation(t *testing.T) {
	path := writeReplayFixture(t, `package fixture
func newReplayApp(api *API) *fiber.App {
	app := fiber.New()
	app.Get("/api/v1/dashboard", handler)
	return app
}`)
	document := map[string]any{
		"paths": map[string]any{
			"/api/v1/dashboard": map[string]any{
				"get": map[string]any{
					"operationId":          "getDashboard",
					runtimeStatusExtension: runtimeStatusPlanned,
				},
			},
		},
	}

	v := &validator{}
	v.validateRuntimeRouteParityAgainstFile(document, path)
	if len(v.errors) != 1 || !strings.Contains(v.errors[0], "still marked planned") {
		t.Fatalf("errors=%v", v.errors)
	}
}

func TestRuntimeParityRejectsMissingImplementedOperation(t *testing.T) {
	path := writeReplayFixture(t, `package fixture
func newReplayApp(api *API) *fiber.App {
	app := fiber.New()
	return app
}`)
	document := map[string]any{
		"paths": map[string]any{
			"/api/v1/health": map[string]any{
				"get": map[string]any{"operationId": "getHealth"},
			},
		},
	}

	v := &validator{}
	v.validateRuntimeRouteParityAgainstFile(document, path)
	if len(v.errors) != 1 || !strings.Contains(v.errors[0], "missing implemented") {
		t.Fatalf("errors=%v", v.errors)
	}
}

func TestRuntimeParityRejectsFrozenReservationWhenPlannedMarkerIsRemoved(t *testing.T) {
	path := writeReplayFixture(t, `package fixture
func newReplayApp(api *API) *fiber.App {
	app := fiber.New()
	return app
}`)
	document := map[string]any{
		"paths": map[string]any{
			"/api/v1/dashboard": map[string]any{
				"get": map[string]any{"operationId": "getDashboard"},
			},
		},
	}

	v := &validator{}
	v.validateRuntimeRouteParityAgainstFile(document, path)
	if len(v.errors) != 1 || !strings.Contains(v.errors[0], "missing implemented") {
		t.Fatalf("errors=%v", v.errors)
	}
}

func writeReplayFixture(t *testing.T, source string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "replay_app.go")
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return path
}
