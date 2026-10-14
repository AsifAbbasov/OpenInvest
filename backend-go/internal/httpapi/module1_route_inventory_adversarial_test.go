package httpapi

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func TestModule1MachineDerivedProtectedRouteInventoryRejectsAnonymous(t *testing.T) {
	requireModule1HTTPAdversarial(t)

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve test source path")
	}
	routesPath := filepath.Join(filepath.Dir(thisFile), "routes.go")
	raw, err := os.ReadFile(routesPath)
	if err != nil {
		t.Fatalf("read routes.go: %v", err)
	}

	re := regexp.MustCompile(`app\.(Get|Post|Put|Patch|Delete)\("([^"]+)", api\.([A-Za-z0-9_]+)\)`)
	matches := re.FindAllStringSubmatch(string(raw), -1)
	if len(matches) == 0 {
		t.Fatal("route parser discovered zero routes")
	}

	authExemptHandlers := map[string]bool{
		"health": true,
		"ready": true,
		"register": true,
		"verifyEmail": true,
		"login": true,
		"refresh": true,
		"logout": true,
		"searchAssets": true,
		"getAsset": true,
		"getCorporateActionProjection": true,
		"calculateDividend": true,
	}

	methodMap := map[string]string{
		"Get": http.MethodGet,
		"Post": http.MethodPost,
		"Put": http.MethodPut,
		"Patch": http.MethodPatch,
		"Delete": http.MethodDelete,
	}

	concretize := func(path string) string {
		path = strings.ReplaceAll(path, ":portfolioId", "11111111-1111-4111-8111-111111111111")
		path = strings.ReplaceAll(path, ":transactionId", "22222222-2222-4222-8222-222222222222")
		path = strings.ReplaceAll(path, ":ticker", "SBER")
		return path
	}

	app := newHTTPAuthApp(t, newHTTPAuthService(t, &httpAuthTestStore{}))

	discovered := 0
	tested := 0
	for _, m := range matches {
		if len(m) != 4 {
			t.Fatalf("unexpected route match: %q", m)
		}
		methodName, path, handler := m[1], m[2], m[3]
		if authExemptHandlers[handler] {
			continue
		}
		method, exists := methodMap[methodName]
		if !exists {
			t.Fatalf("unsupported route method in inventory: %s", methodName)
		}
		discovered++

		name := method + "_" + strings.ReplaceAll(strings.TrimPrefix(path, "/"), "/", "_")
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(method, concretize(path), nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("anonymous request: %v", err)
			}
			defer resp.Body.Close()
			t.Logf("MACHINE_PROTECTED_ROUTE=%s %s HANDLER=%s STATUS=%d", method, path, handler, resp.StatusCode)
			if resp.StatusCode != http.StatusUnauthorized {
				t.Fatalf("protected route did not fail at auth boundary: method=%s path=%s handler=%s status=%d", method, path, handler, resp.StatusCode)
			}
			tested++
		})
	}

	if discovered == 0 {
		t.Fatal("no protected routes discovered")
	}
	if tested != discovered {
		t.Fatalf("protected route inventory incomplete: discovered=%d tested=%d", discovered, tested)
	}
	t.Logf("PROTECTED_ROUTES_DISCOVERED=%d PROTECTED_ROUTES_TESTED=%d PROTECTED_ROUTES_WITHOUT_AUTH_GUARD=0", discovered, tested)
}
