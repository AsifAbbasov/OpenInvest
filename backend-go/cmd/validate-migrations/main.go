package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	mode := flag.String("mode", "local", "local|repository|pr")
	base := flag.String("base-sha", "", "PR base commit SHA")
	flag.Parse()
	root, err := repoRoot()
	if err != nil {
		fatal(err)
	}
	v := &Validator{Root: root, Mode: *mode, BaseSHA: *base}
	if err := v.Validate(); err != nil {
		fatal(err)
	}
	fmt.Printf("Validated %d migration pair(s)\n", len(v.pairs))
	if *mode == "pr" {
		fmt.Printf("BASE_IMMUTABILITY=PROVEN base=%s\n", *base)
	} else {
		fmt.Println("BASE_IMMUTABILITY=NOT_APPLICABLE")
	}
}
func repoRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for d := cwd; ; d = filepath.Dir(d) {
		if _, err := os.Stat(filepath.Join(d, "infrastructure", "postgres", "migrations")); err == nil {
			return d, nil
		}
		p := filepath.Dir(d)
		if p == d {
			break
		}
	}
	return "", fmt.Errorf("repository root not found")
}
func fatal(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
