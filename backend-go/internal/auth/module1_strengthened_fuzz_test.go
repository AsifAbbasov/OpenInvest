package auth

import (
	"strings"
	"testing"
)

func FuzzModule1EmailNormalizationProperties(f *testing.F) {
	for _, seed := range []string{
		" Investor@Example.COM ",
		"USER@example.test",
		"user+tag@example.test",
		"user.name@example.test",
		"\tMixed@Example.Test\n",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > 2048 {
			return
		}
		got := normalizeEmail(input)
		if got != normalizeEmail(got) {
			t.Fatalf("email normalization is not idempotent: input=%q normalized=%q renormalized=%q", input, got, normalizeEmail(got))
		}
		if got != strings.TrimSpace(got) {
			t.Fatalf("normalized email retains edge whitespace: input=%q normalized=%q", input, got)
		}
		if got != strings.ToLower(got) {
			t.Fatalf("normalized email retains uppercase characters: input=%q normalized=%q", input, got)
		}
		want := strings.ToLower(strings.TrimSpace(input))
		if got != want {
			t.Fatalf("normalization drift: input=%q got=%q want=%q", input, got, want)
		}
	})
}
