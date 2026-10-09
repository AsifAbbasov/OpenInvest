package main

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/openinvest/openinvest/backend-go/internal/provider/tinvest"
	"github.com/openinvest/openinvest/backend-go/internal/sharedbudget"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

const (
	tinvestCorporateActionsEnabledEnv     = "OPENINVEST_TINVEST_CORPORATE_ACTIONS_ENABLED"
	tinvestReadOnlyTokenEnv                = "OPENINVEST_TINVEST_READONLY_TOKEN"
	tinvestGlobalBudgetOwnerEnv            = "OPENINVEST_TINVEST_GLOBAL_BUDGET_OWNER"
	verifiedTInvestGlobalBudgetOwner       = "verified-shared-provider-budget-v1"
)

func configuredTInvestCorporateActionProvider(budget sharedbudget.Authority) (verticalslice.CorporateActionProvider, error) {
	if !envBool(tinvestCorporateActionsEnabledEnv) {
		return nil, nil
	}

	if strings.TrimSpace(os.Getenv(tinvestGlobalBudgetOwnerEnv)) != verifiedTInvestGlobalBudgetOwner {
		return nil, errors.New("T-Invest Corporate Actions is enabled but deployment-global provider budget ownership is not verified")
	}

	token := os.Getenv(tinvestReadOnlyTokenEnv)
	if token == "" || strings.TrimSpace(token) != token {
		return nil, errors.New("T-Invest Corporate Actions is enabled but OPENINVEST_TINVEST_READONLY_TOKEN is not validly configured")
	}
	if budget == nil && !isExplicitDevelopmentEnvironment() {
		return nil, errors.New("T-Invest Corporate Actions is enabled but shared Redis budget enforcement is not configured")
	}

	var provider verticalslice.CorporateActionProvider
	var err error
	if budget != nil {
		provider, err = tinvest.NewCorporateActionProviderWithSharedBudget(&http.Client{}, verticalslice.SystemClock{}, token, budget)
	} else {
		provider, err = tinvest.NewCorporateActionProvider(&http.Client{}, verticalslice.SystemClock{}, token)
	}
	if err != nil {
		return nil, errors.New("T-Invest Corporate Actions provider configuration is invalid")
	}
	return provider, nil
}
