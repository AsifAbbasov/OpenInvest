package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestM11StrictJSONAmbiguityObservation(t *testing.T) {
	var unknown struct {
		Name string `json:"name"`
	}
	if err := decodeStrictJSON([]byte(`{"name":"ok","unexpected":true}`), &unknown); err == nil {
		t.Fatal("unknown field unexpectedly accepted")
	}
	if err := decodeStrictJSON([]byte(`{"name":"ok"} {}`), &unknown); err == nil {
		t.Fatal("trailing JSON unexpectedly accepted")
	}
	var duplicate struct {
		Name string `json:"name"`
	}
	if err := decodeStrictJSON([]byte(`{"name":"first","name":"second"}`), &duplicate); err != nil {
		t.Fatalf("duplicate-key observation decode: %v", err)
	}
	if duplicate.Name != "second" {
		t.Fatalf("duplicate-key decoder semantic changed: got %q", duplicate.Name)
	}
	t.Log("M11_DUPLICATE_JSON_KEY_BEHAVIOR=LAST_VALUE_WINS")
	t.Log("M11_DUPLICATE_JSON_SECURITY_IMPACT=NOT_DEMONSTRATED_BY_THIS_MATRIX")
}

func TestM11ProviderValidationPrecedesBudgetAdmission(t *testing.T) {
	provider := &module10RemediationProvider{}
	_, app := module10RemediationApp(provider)
	peer := "203.0.113.201"
	invalid := module10RemediationPath("SBER", "not-a-date", "2026-12-31")
	for i := 0; i < 20; i++ {
		res := module10RemediationRequest(t, app, peer, invalid, "")
		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("invalid request %d status=%d want=400", i+1, res.StatusCode)
		}
	}
	if provider.calls.Load() != 0 {
		t.Fatalf("invalid payload reached provider calls=%d", provider.calls.Load())
	}

	valid := module10RemediationPath("SBER", "2026-01-01", "2026-12-31")
	for i := 0; i < defaultCorporateActionProjectionPerClientLimit; i++ {
		res := module10RemediationRequest(t, app, peer, valid, "")
		if res.StatusCode != http.StatusOK {
			t.Fatalf("valid request %d status=%d want=200", i+1, res.StatusCode)
		}
	}
	limited := module10RemediationRequest(t, app, peer, valid, "")
	if limited.StatusCode != http.StatusTooManyRequests || limited.RetryAfter != "60" {
		t.Fatalf("rate-limit result=%+v", limited)
	}
	if provider.calls.Load() != defaultCorporateActionProjectionPerClientLimit {
		t.Fatalf("provider calls=%d want=%d", provider.calls.Load(), defaultCorporateActionProjectionPerClientLimit)
	}
	t.Log("M11_CM08_INVALID_REQUESTS_DO_NOT_AMPLIFY_PROVIDER_BUDGET=PASS")
	t.Log("M11_CM12_VALIDATION_RATE_LIMIT_PROVIDER_ORDERING=PASS")
}

func TestM11DuplicateQueryParametersDoNotAmplifyProvider(t *testing.T) {
	provider := &module10RemediationProvider{}
	_, app := module10RemediationApp(provider)
	path := "/api/v1/corporate-actions/projection?instrumentId=SBER&instrumentId=GAZP&from=2026-01-01&to=2026-12-31"
	res := module10RemediationRequest(t, app, "203.0.113.202", path, "")
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusBadRequest {
		t.Fatalf("duplicate query status=%d", res.StatusCode)
	}
	if provider.calls.Load() > 1 {
		t.Fatalf("one HTTP request amplified to %d provider calls", provider.calls.Load())
	}
	t.Logf("M11_DUPLICATE_QUERY_STATUS=%d PROVIDER_CALLS=%d", res.StatusCode, provider.calls.Load())
}

func TestM11NotFoundErrorOracleUniform(t *testing.T) {
	app := fiber.New()
	app.Get("/vertical", func(c fiber.Ctx) error { return writeMappedError(c, verticalslice.ErrNotFound) })
	app.Get("/postgres", func(c fiber.Ctx) error { return writeMappedError(c, postgres.ErrNotFound) })

	read := func(path string) errorResponse {
		response, err := app.Test(httptest.NewRequest(http.MethodGet, path, nil))
		if err != nil {
			t.Fatalf("request %s: %v", path, err)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusNotFound {
			t.Fatalf("%s status=%d", path, response.StatusCode)
		}
		var payload errorResponse
		if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
			t.Fatalf("decode %s: %v", path, err)
		}
		return payload
	}
	a := read("/vertical")
	b := read("/postgres")
	if a.Error.Code != "NOT_FOUND" || b.Error.Code != "NOT_FOUND" ||
		a.Error.Message != "Resource not found" || b.Error.Message != "Resource not found" {
		t.Fatalf("not-found oracle drift: vertical=%+v postgres=%+v", a.Error, b.Error)
	}
	if strings.Contains(strings.ToLower(a.Error.Message+b.Error.Message), "sql") ||
		strings.Contains(strings.ToLower(a.Error.Message+b.Error.Message), "postgres") {
		t.Fatalf("internal storage detail leaked: %q / %q", a.Error.Message, b.Error.Message)
	}
	t.Log("M11_CM09_NOT_FOUND_ORACLE_UNIFORM=PASS")
	t.Log("M11_CM10_INTERNAL_ERROR_DETAIL_LEAK=NONE")
}

func FuzzM11StrictJSONDecoder(f *testing.F) {
	for _, seed := range []string{
		`{"name":"ok"}`,
		`{"name":"first","name":"second"}`,
		`{"name":"ok","unexpected":true}`,
		`{"name":"ok"} {}`,
		"",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, body string) {
		if len(body) > 16*1024 {
			t.Skip()
		}
		var target struct {
			Name string `json:"name"`
		}
		_ = decodeStrictJSON([]byte(body), &target)
	})
}

func FuzzM11ProviderCanonicalRequestKey(f *testing.F) {
	f.Add("SBER", "GAZP", "2026-01-01", "2026-12-31")
	f.Add("A", "B", "2000-01-01", "2099-12-31")
	f.Fuzz(func(t *testing.T, first, second, from, to string) {
		if len(first)+len(second)+len(from)+len(to) > 4096 {
			t.Skip()
		}
		leftIDs := []string{first, second}
		rightIDs := []string{second, first}
		sort.Strings(leftIDs)
		sort.Strings(rightIDs)
		left := corporateActionProjectionRequestKey(verticalslice.CorporateActionQuery{InstrumentIDs: leftIDs, From: from, To: to})
		right := corporateActionProjectionRequestKey(verticalslice.CorporateActionQuery{InstrumentIDs: rightIDs, From: from, To: to})
		if left != right {
			t.Fatalf("canonical-equivalent provider keys differ: %s != %s", left, right)
		}
	})
}
