package securityci

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func TestCIUsesLeastPrivilegeTokenAndImmutableActionPins(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(repoRoot(t), ".github", "workflows", "ci.yml"))
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(data)

	if !strings.Contains(workflow, "permissions:\n  contents: read") {
		t.Fatal("CI must explicitly default GITHUB_TOKEN to contents: read")
	}
	for _, forbidden := range []string{
		"permissions: write-all",
		"contents: write",
		"pull-requests: write",
		"actions: write",
	} {
		if strings.Contains(workflow, forbidden) {
			t.Fatalf("CI contains write-capable default permission %q", forbidden)
		}
	}

	usesLine := regexp.MustCompile("(?m)^\\s*-?\\s*uses:\\s*([^\\s#]+)")
	shaPin := regexp.MustCompile("@[0-9a-fA-F]{40}$")
	for _, match := range usesLine.FindAllStringSubmatch(workflow, -1) {
		ref := match[1]
		if strings.HasPrefix(ref, "./") {
			continue
		}
		if !shaPin.MatchString(ref) {
			t.Fatalf("third-party GitHub Action is not pinned to an immutable 40-char SHA: %s", ref)
		}
	}
}

func TestRepositoryConfigDoesNotContainObviousProductionSecrets(t *testing.T) {
	root := repoRoot(t)
	candidates := []string{
		".env.example",
		"docker-compose.yml",
		".github/workflows/ci.yml",
	}
	highRisk := regexp.MustCompile("(?i)(BEGIN (RSA|EC|OPENSSH) PRIVATE KEY|gh[pousr]_[A-Za-z0-9_]{20,}|sk-[A-Za-z0-9]{20,})")

	for _, relative := range candidates {
		data, err := os.ReadFile(filepath.Join(root, relative))
		if err != nil {
			t.Fatal(err)
		}
		if highRisk.Match(data) {
			t.Fatalf("high-risk secret-like value found in tracked configuration %s", relative)
		}
	}

	envData, err := os.ReadFile(filepath.Join(root, ".env.example"))
	if err != nil {
		t.Fatal(err)
	}
	env := string(envData)
	if !strings.Contains(env, "replace-with-32-plus-byte-secret") ||
		!strings.Contains(env, "replace-with-a-different-32-plus-byte-secret") {
		t.Fatal(".env.example must contain explicit non-production placeholder secrets")
	}
}
