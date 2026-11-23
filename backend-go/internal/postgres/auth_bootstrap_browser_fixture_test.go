package postgres_test

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/openinvest/openinvest/backend-go/internal/auth"
)

// This writes ONLY disposable browser-test account credentials, NEVER tokens.
// The runner creates a private temporary directory and deletes it on exit.
func TestStageA2BrowserCreateDisposableAccounts(t *testing.T) {
	target := os.Getenv("OPENINVEST_A2_BROWSER_ACCOUNT_FILE")
	if target == "" {
		t.Skip("browser fixture file is not configured")
	}
	if os.Getenv("OPENINVEST_A2_DISPOSABLE_DB_CONFIRMED") != "1" {
		t.Fatal("browser fixture generation requires disposable database confirmation")
	}
	service, _ := newBootstrapIntegrationFixture(t)
	accounts := make([]map[string]string, 0, 2)
	for i := 0; i < 2; i++ {
		raw := make([]byte, 27)
		if _, err := rand.Read(raw); err != nil {
			t.Fatalf("generate test password: %v", err)
		}
		password := "A2!" + base64.RawURLEncoding.EncodeToString(raw)
		email := "a2-browser-" + uuid.NewString() + "@example.com"
		result, err := service.Register(context.Background(), auth.RegistrationRequest{
			Email: email, Password: password, Language: auth.LanguageEN,
			Theme: auth.ThemeSystem, Timezone: "UTC",
		})
		if err != nil {
			t.Fatalf("register browser fixture: %v", err)
		}
		accounts = append(accounts, map[string]string{"email": result.User.Email, "password": password})
	}
	data, err := json.Marshal(accounts)
	if err != nil {
		t.Fatalf("encode fixtures: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
		t.Fatalf("private fixture directory: %v", err)
	}
	if err := os.WriteFile(target, data, 0600); err != nil {
		t.Fatalf("save disposable fixtures: %v", err)
	}
	t.Log("Created two disposable browser accounts; credentials intentionally omitted from logs")
}
