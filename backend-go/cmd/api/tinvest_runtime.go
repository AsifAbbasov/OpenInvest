package main

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/openinvest/openinvest/backend-go/internal/provider/tinvest"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

const (
	tinvestCorporateActionsEnabledEnv = "OPENINVEST_TINVEST_CORPORATE_ACTIONS_ENABLED"
	tinvestReadOnlyTokenEnv           = "OPENINVEST_TINVEST_READONLY_TOKEN"
)

func configuredTInvestCorporateActionProvider() (verticalslice.CorporateActionProvider, error) {
	if !envBool(tinvestCorporateActionsEnabledEnv) {
		return nil, nil
	}

	token := os.Getenv(tinvestReadOnlyTokenEnv)
	if token == "" || strings.TrimSpace(token) != token {
		return nil, errors.New("T-Invest Corporate Actions is enabled but OPENINVEST_TINVEST_READONLY_TOKEN is not validly configured")
	}

	provider, err := tinvest.NewCorporateActionProvider(&http.Client{}, verticalslice.SystemClock{}, token)
	if err != nil {
		return nil, errors.New("T-Invest Corporate Actions provider configuration is invalid")
	}
	return provider, nil
}
