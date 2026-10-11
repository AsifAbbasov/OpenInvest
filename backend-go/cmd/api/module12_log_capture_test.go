package main

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// Invoke the real main/log.Fatal path in a child process. Missing database
// configuration fails before listener, Redis or provider initialization.
func TestModule12ApplicationDiagnosticCapture(t *testing.T) {
	if os.Getenv("M12_LOG_CAPTURE_CHILD") == "1" {
		main()
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestModule12ApplicationDiagnosticCapture$")
	// An explicit, minimal environment prevents inherited runtime configuration
	// and production credentials from reaching this fixture.
	command.Env = []string{"M12_LOG_CAPTURE_CHILD=1", "OPENINVEST_ENV=production", "DATABASE_URL="}
	var sentinels []string
	for _, category := range []string{"ACCESS_TOKEN", "REFRESH_COOKIE", "AUTHORIZATION_BEARER", "PROVIDER_TOKEN", "DATABASE_PASSWORD", "REDIS_PASSWORD", "IMPORT_REVIEW_TOKEN", "CSRF_TOKEN", "FINANCIAL_PAYLOAD_SENTINEL"} {
		value := "M12_DIAGNOSTIC_" + category + "_43851f"
		sentinels = append(sentinels, value)
		command.Env = append(command.Env, category+"="+value)
	}
	output, err := command.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("diagnostic process exceeded bounded lifetime: %v", ctx.Err())
	}
	exit, ok := err.(*exec.ExitError)
	if !ok || exit.ExitCode() != 1 {
		t.Fatalf("expected real main fatal exit 1, got %v", err)
	}
	if len(output) == 0 || !strings.Contains(string(output), "validate runtime safety: DATABASE_URL is required") {
		t.Error("application-owned startup diagnostic was not captured")
	}
	for _, value := range sentinels {
		if strings.Contains(string(output), value) {
			t.Error("synthetic secret appeared in startup diagnostic")
		}
	}
	t.Logf("M12_SAFE_LOG_CAPTURE_MECHANISM_OBSERVED_NONZERO_OUTPUT=YES CAPTURED_APPLICATION_DIAGNOSTIC_BYTES=%d PATH=REAL_MAIN_LOG_FATAL_MISSING_DATABASE_CONFIG", len(output))
}
