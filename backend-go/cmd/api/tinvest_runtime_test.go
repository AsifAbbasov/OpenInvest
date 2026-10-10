package main

import "testing"

func TestConfiguredTInvestCorporateActionProviderIsDisabledByDefault(t *testing.T) {
	t.Setenv(tinvestCorporateActionsEnabledEnv, "")
	t.Setenv(tinvestReadOnlyTokenEnv, "token-must-not-activate-by-itself")
	provider, err := configuredTInvestCorporateActionProvider(nil)
	if err != nil {
		t.Fatalf("configuration: %v", err)
	}
	if provider != nil {
		t.Fatal("provider activated without explicit enable flag")
	}
}

func TestConfiguredTInvestCorporateActionProviderRequiresTokenWhenEnabled(t *testing.T) {
	t.Setenv(tinvestCorporateActionsEnabledEnv, "true")
	t.Setenv(tinvestGlobalBudgetOwnerEnv, verifiedTInvestGlobalBudgetOwner)
	for _, token := range []string{"", " ", " token", "token "} {
		t.Run(token, func(t *testing.T) {
			t.Setenv(tinvestReadOnlyTokenEnv, token)
			provider, err := configuredTInvestCorporateActionProvider(nil)
			if err == nil {
				t.Fatal("expected fail-closed configuration error")
			}
			if provider != nil {
				t.Fatal("provider must remain nil when token configuration is invalid")
			}
		})
	}
}

func TestConfiguredTInvestCorporateActionProviderRequiresBothActivationInputs(t *testing.T) {
	t.Setenv("OPENINVEST_ENV", "development")
	t.Setenv(tinvestCorporateActionsEnabledEnv, "true")
	t.Setenv(tinvestGlobalBudgetOwnerEnv, verifiedTInvestGlobalBudgetOwner)
	t.Setenv(tinvestReadOnlyTokenEnv, "test-readonly-token")
	provider, err := configuredTInvestCorporateActionProvider(nil)
	if err != nil {
		t.Fatalf("configuration: %v", err)
	}
	if provider == nil {
		t.Fatal("expected explicitly enabled provider")
	}
}


func TestArchitectureHardeningTInvestRequiresVerifiedSharedBudgetOwner(t *testing.T) {
	t.Setenv(tinvestCorporateActionsEnabledEnv, "true")
	t.Setenv(tinvestReadOnlyTokenEnv, "test-readonly-token")
	for _, owner := range []string{"", "local", "process-local"} {
		t.Run(owner, func(t *testing.T) {
			t.Setenv(tinvestGlobalBudgetOwnerEnv, owner)
			provider, err := configuredTInvestCorporateActionProvider(nil)
			if err == nil {
				t.Fatal("expected provider activation to fail without verified shared budget ownership")
			}
			if provider != nil {
				t.Fatal("provider must remain disabled without verified shared budget ownership")
			}
		})
	}
}

func TestConfiguredTInvestCorporateActionProviderRequiresSharedBudgetOutsideDevelopment(t *testing.T) {
	t.Setenv("OPENINVEST_ENV", "production")
	t.Setenv(tinvestCorporateActionsEnabledEnv, "true")
	t.Setenv(tinvestGlobalBudgetOwnerEnv, verifiedTInvestGlobalBudgetOwner)
	t.Setenv(tinvestReadOnlyTokenEnv, "test-readonly-token")
	provider, err := configuredTInvestCorporateActionProvider(nil)
	if err == nil {
		t.Fatal("expected production provider activation to fail without shared Redis budget")
	}
	if provider != nil {
		t.Fatal("provider must remain nil without shared Redis budget")
	}
}
