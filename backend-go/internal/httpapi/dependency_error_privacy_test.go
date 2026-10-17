package httpapi

// Retains dependency-disclosure assertions from PR #246 at
// 6efe990164e89c212844353867d2f50beb4f7841 without obsolete registration
// expectations or process-global stdout replacement.
import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestDependencyResponsesExcludeSyntheticSecrets(t *testing.T) {
	categories := []string{"ACCESS_TOKEN", "REFRESH_COOKIE", "AUTHORIZATION_BEARER", "PROVIDER_TOKEN", "DATABASE_PASSWORD", "REDIS_PASSWORD", "IMPORT_REVIEW_TOKEN", "CSRF_TOKEN", "FINANCIAL_PAYLOAD_SENTINEL", "USER_EMAIL_SENTINEL"}
	var secrets []string
	for _, category := range categories {
		secrets = append(secrets, "PR246_PRIVACY_"+category+"_8e275b")
	}
	detail := "postgres://fixture:password@db.internal:5432/db redis://redis.internal:6379 /workspace/private/store.go SELECT confidential " + strings.Join(secrets, " ")
	t.Run("database_login_error", func(t *testing.T) {
		store := &httpAuthTestStore{findErr: errors.New(detail)}
		app := newHTTPAuthApp(t, newHTTPAuthService(t, store))
		response := authRequest(t, app, http.MethodPost, "/api/v1/auth/login", `{"email":"privacy@example.test","password":"synthetic password material"}`, "", "")
		assertDependencyPrivacy(t, response, http.StatusInternalServerError, secrets)
	})
	t.Run("provider_error", func(t *testing.T) {
		provider := &corporateActionFixtureProvider{err: fmt.Errorf("provider dependency: %s", detail)}
		response := performCorporateActionRequest(t, provider, "/api/v1/corporate-actions/projection?instrumentId=SBER&from=2026-01-01&to=2026-12-31")
		assertDependencyPrivacy(t, response, http.StatusServiceUnavailable, secrets)
	})
}

func assertDependencyPrivacy(t *testing.T, response *http.Response, status int, secrets []string) {
	t.Helper()
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != status {
		t.Errorf("status=%d want=%d", response.StatusCode, status)
	}
	if len(body) == 0 {
		t.Error("missing positive response-body control")
	}
	for _, needle := range append(append([]string{}, secrets...), "panic:", "goroutine ", "/workspace/", "SELECT ", "postgres://", "redis://", "db.internal", "redis.internal") {
		if strings.Contains(string(body), needle) || strings.Contains(fmt.Sprint(response.Header), needle) {
			t.Error("dependency response disclosed synthetic sensitive detail")
		}
	}
}
