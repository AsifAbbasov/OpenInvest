package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

var requiredUpFragments = []string{
	"CREATE SCHEMA IF NOT EXISTS identity",
	"CREATE SCHEMA IF NOT EXISTS investment",
	"CREATE SCHEMA IF NOT EXISTS analytics",
	"CREATE SCHEMA IF NOT EXISTS audit",
	"CREATE TABLE identity.users",
	"CREATE TABLE identity.user_investment_links",
	"CREATE TABLE investment.subjects",
	"CREATE TABLE investment.assets",
	"CREATE TABLE investment.portfolios",
	"CREATE TABLE investment.transaction_entries",
	"CREATE TABLE investment.command_deduplication",
	"CREATE TABLE investment.outbox_events",
	"CREATE TABLE analytics.portfolio_snapshots",
	"CREATE TABLE analytics.snapshot_positions",
	"CREATE TABLE analytics.calculation_runs",
	"CREATE TABLE analytics.inbox_messages",
	"CREATE TABLE audit.actors",
	"CREATE TABLE audit.events",
}

type forbiddenPattern struct {
	pattern *regexp.Regexp
	message string
}

var forbiddenUpPatterns = []forbiddenPattern{
	{regexp.MustCompile(`(?i)\bDROP\b`), "DROP is forbidden in up migrations"},
	{regexp.MustCompile(`(?i)\bTRUNCATE\b`), "TRUNCATE is forbidden in up migrations"},
	{regexp.MustCompile(`(?i)\bDELETE\s+FROM\b`), "DELETE FROM is forbidden in up migrations"},
	{regexp.MustCompile(`(?i)\bUPDATE\s+[a-zA-Z0-9_."]+\s+SET`), "UPDATE is forbidden in up migrations"},
	{regexp.MustCompile(`(?is)\bALTER\s+TABLE\b.*\bDROP\b`), "ALTER TABLE ... DROP is forbidden in up migrations"},
}

func main() {
	root := filepath.Clean(filepath.Join(".."))
	migrationsDir := filepath.Join(root, "infrastructure", "postgres", "migrations")
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		fail("Missing migrations directory: %s", migrationsDir)
	}

	var upFiles []string
	var downFiles []string
	for _, entry := range entries {
		name := entry.Name()
		switch {
		case strings.HasSuffix(name, ".up.sql"):
			upFiles = append(upFiles, filepath.Join(migrationsDir, name))
		case strings.HasSuffix(name, ".down.sql"):
			downFiles = append(downFiles, filepath.Join(migrationsDir, name))
		}
	}
	slices.Sort(upFiles)
	slices.Sort(downFiles)

	if len(upFiles) == 0 {
		fail("No up migrations found")
	}

	upPrefixes := prefixes(upFiles, ".up.sql")
	downPrefixes := prefixes(downFiles, ".down.sql")
	missingDown := difference(upPrefixes, downPrefixes)
	missingUp := difference(downPrefixes, upPrefixes)
	if len(missingDown) > 0 || len(missingUp) > 0 {
		fmt.Fprintln(os.Stderr, "Migration up/down mismatch")
		if len(missingDown) > 0 {
			fmt.Fprintf(os.Stderr, "Missing down migrations for: %s\n", strings.Join(missingDown, ", "))
		}
		if len(missingUp) > 0 {
			fmt.Fprintf(os.Stderr, "Missing up migrations for: %s\n", strings.Join(missingUp, ", "))
		}
		os.Exit(1)
	}
	if !slices.IsSorted(upPrefixes) {
		fail("Migration files must be lexicographically ordered")
	}

	for _, path := range upFiles {
		content, err := os.ReadFile(path)
		if err != nil {
			fail("cannot read %s: %v", path, err)
		}
		sql := string(content)
		for _, forbidden := range forbiddenUpPatterns {
			if forbidden.pattern.MatchString(sql) {
				fail("%s: %s", relative(root, path), forbidden.message)
			}
		}
		if strings.HasPrefix(filepath.Base(path), "000001_stage_03_01_vertical_slice") {
			for _, fragment := range requiredUpFragments {
				if !strings.Contains(sql, fragment) {
					fail("%s: missing required fragment: %s", relative(root, path), fragment)
				}
			}
			if regexp.MustCompile(`(?i)\bFLOAT\b|\bDOUBLE\b|\bREAL\b`).MatchString(sql) {
				fail("%s: binary floating-point types are forbidden for financial schema", relative(root, path))
			}
			if !strings.Contains(sql, "NUMERIC(28, 8)") {
				fail("%s: expected NUMERIC(28, 8) decimal financial columns", relative(root, path))
			}
		}
	}

	fmt.Printf("Validated %d migration pair(s)\n", len(upFiles))
}

func prefixes(paths []string, suffix string) []string {
	result := make([]string, 0, len(paths))
	for _, path := range paths {
		result = append(result, strings.TrimSuffix(filepath.Base(path), suffix))
	}
	return result
}

func difference(left []string, right []string) []string {
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
	return result
}

func relative(root string, path string) string {
	value, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return value
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
