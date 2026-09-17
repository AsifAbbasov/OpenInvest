package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/postgres"
)

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "replay-admin:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.New("subcommand required: allocate|backfill|finalize|activate|invalidate|verify")
	}
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return errors.New("DATABASE_URL is required and must be the schema-owner/admin connection")
	}
	store, err := postgres.Open(databaseURL)
	if err != nil {
		return err
	}
	defer store.Close()
	now := time.Now().UTC()

	switch args[0] {
	case "allocate":
		if len(args) != 1 {
			return errors.New("allocate accepts no arguments")
		}
		generation, err := store.AllocateReplayGeneration(ctx, now)
		if err != nil {
			return err
		}
		fmt.Printf("POLICY_VERSION=%s\nACTIVATION_GENERATION=%d\nSTATE=PREACTIVE\n", generation.PolicyVersion, generation.ActivationGeneration)
		return nil
	case "backfill":
		generation, mutableRows, _, err := parseGenerationWorkFlags("backfill", args[1:], false)
		if err != nil {
			return err
		}
		result, err := store.ProvisionalReplayBackfill(ctx, generation, mutableRows, now)
		if err != nil {
			return err
		}
		printBackfillResult("PROVISIONAL", result)
		return nil
	case "finalize":
		generation, mutableRows, manifestPath, err := parseGenerationWorkFlags("finalize", args[1:], true)
		if err != nil {
			return err
		}
		result, manifest, err := store.FinalizeReplayGeneration(ctx, generation, mutableRows, now)
		if err != nil {
			return err
		}
		if err := writeExclusive(manifestPath, manifest.CanonicalBytes); err != nil {
			return err
		}
		printBackfillResult("FINALIZED", result)
		fmt.Printf("ACTIVATION_MANIFEST=%s\nACTIVATION_MANIFEST_SHA256=%s\nPORTFOLIO_COUNT=%d\n", manifestPath, manifest.SHA256, manifest.PortfolioCount)
		return nil
	case "activate":
		fs := flag.NewFlagSet("activate", flag.ContinueOnError)
		generation := fs.String("generation", "", "PREACTIVE activation generation")
		manifestSHA := fs.String("manifest-sha256", "", "expected finalization manifest SHA256")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 0 {
			return errors.New("activate has unexpected positional arguments")
		}
		value, err := requireGeneration(*generation)
		if err != nil {
			return err
		}
		if err := store.ActivateReplayGeneration(ctx, value, strings.TrimSpace(*manifestSHA), now); err != nil {
			return err
		}
		fmt.Printf("ACTIVATION_GENERATION=%d\nEVENT=ACTIVATED\n", value)
		return nil
	case "invalidate":
		fs := flag.NewFlagSet("invalidate", flag.ContinueOnError)
		generation := fs.String("generation", "", "active activation generation")
		reason := fs.String("reason", "", "operator reason_code")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 0 {
			return errors.New("invalidate has unexpected positional arguments")
		}
		value, err := requireGeneration(*generation)
		if err != nil {
			return err
		}
		if err := store.InvalidateReplayGeneration(ctx, value, *reason, now); err != nil {
			return err
		}
		fmt.Printf("ACTIVATION_GENERATION=%d\nEVENT=INVALIDATED\n", value)
		return nil
	case "verify":
		fs := flag.NewFlagSet("verify", flag.ContinueOnError)
		generation := fs.String("generation", "", "activation generation")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 0 {
			return errors.New("verify has unexpected positional arguments")
		}
		value, err := requireGeneration(*generation)
		if err != nil {
			return err
		}
		manifest, err := store.VerifyReplayGeneration(ctx, value)
		if err != nil {
			return err
		}
		fmt.Printf("ACTIVATION_GENERATION=%d\nVERIFIED=YES\nACTIVATION_MANIFEST_SHA256=%s\nPORTFOLIO_COUNT=%d\n", value, manifest.SHA256, manifest.PortfolioCount)
		return nil
	default:
		return fmt.Errorf("unknown subcommand %q", args[0])
	}
}

func parseGenerationWorkFlags(name string, args []string, requireManifest bool) (int64, int, string, error) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	generation := fs.String("generation", "", "PREACTIVE activation generation")
	mutableRows := fs.Int("mutable-rows", -1, "explicit target mutable raw-row suffix, 0..5000")
	manifest := fs.String("manifest", "", "exclusive output path for canonical activation manifest")
	if err := fs.Parse(args); err != nil {
		return 0, 0, "", err
	}
	if fs.NArg() != 0 {
		return 0, 0, "", fmt.Errorf("%s has unexpected positional arguments", name)
	}
	value, err := requireGeneration(*generation)
	if err != nil {
		return 0, 0, "", err
	}
	if *mutableRows < 0 || *mutableRows > 5000 {
		return 0, 0, "", errors.New("--mutable-rows must be supplied explicitly in range 0..5000")
	}
	if requireManifest && strings.TrimSpace(*manifest) == "" {
		return 0, 0, "", errors.New("--manifest output path is required for finalization")
	}
	return value, *mutableRows, strings.TrimSpace(*manifest), nil
}

func requireGeneration(value string) (int64, error) {
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil || parsed <= 0 {
		return 0, errors.New("--generation must be a positive integer")
	}
	return parsed, nil
}

func writeExclusive(path string, contents []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("create manifest: %w", err)
	}
	defer file.Close()
	if _, err := file.Write(contents); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}
	return file.Sync()
}

func printBackfillResult(state string, result postgres.ReplayBackfillResult) {
	fmt.Printf("POLICY_VERSION=%s\nACTIVATION_GENERATION=%d\nEPOCH_STATE=%s\nPORTFOLIOS=%d\nEPOCHS=%d\n",
		result.PolicyVersion, result.ActivationGeneration, state, len(result.PortfolioIDs), len(result.EpochIDs))
}
