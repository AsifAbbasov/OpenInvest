package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/openinvest/openinvest/backend-go/internal/sharedbudget"
)

const (
	sharedBudgetRedisURLEnv = "OPENINVEST_SHARED_BUDGET_REDIS_URL"
	sharedBudgetPingTimeout = 2 * time.Second
)

func configuredSharedBudgetAuthority() (sharedbudget.Authority, error) {
	rawURL := strings.TrimSpace(os.Getenv(sharedBudgetRedisURLEnv))
	if rawURL == "" {
		return nil, nil
	}
	namespace := "openinvest:" + sharedBudgetEnvironmentNamespace() + ":security-budget:v1"
	authority, err := sharedbudget.NewRedisAuthority(rawURL, namespace)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), sharedBudgetPingTimeout)
	defer cancel()
	if err := authority.Ping(ctx); err != nil {
		_ = authority.Close()
		return nil, fmt.Errorf("verify shared budget Redis backend: %w", err)
	}
	return authority, nil
}

func sharedBudgetEnvironmentNamespace() string {
	environment := strings.ToLower(strings.TrimSpace(os.Getenv("OPENINVEST_ENV")))
	if environment == "" {
		return "default"
	}
	var builder strings.Builder
	for _, r := range environment {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '_':
			builder.WriteRune(r)
		default:
			builder.WriteByte('-')
		}
	}
	if builder.Len() == 0 {
		return "default"
	}
	return builder.String()
}

func closeSharedBudget(authority sharedbudget.Authority) error {
	if authority == nil {
		return nil
	}
	return authority.Close()
}

func closeRuntimeDependencies(storeClose func() error, authority sharedbudget.Authority) error {
	var errs []error
	if storeClose != nil {
		if err := storeClose(); err != nil {
			errs = append(errs, err)
		}
	}
	if err := closeSharedBudget(authority); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}
