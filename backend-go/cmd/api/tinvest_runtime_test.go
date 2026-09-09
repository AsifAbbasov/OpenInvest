package main

import "testing"

func TestConfiguredTInvestCorporateActionProviderIsDisabledByDefault(t *testing.T) {
	t.Setenv(tinvestCorporateActionsEnabledEnv, "")
	t.Setenv(tinvestReadOnlyTokenEnv, "token-must-not-activate-by-itself")
	provider, err := configuredTInvestCorporateActionProvider()
	if err != nil {
		t.Fatalf("configuration: %v", err)
	}
	if provider != nil {
		t.Fatal("provider activated without explicit enable flag")
	}
}

func TestConfiguredTInvestCorporateActionProviderRequiresTokenWhenEnabled(t *testing.T) {
	t.Setenv(tinvestCorporateActionsEnabledEnv, "true")
	for _, token := range []string{"", " ", " token", "token "} {
		t.Run(token, func(t *testing.T) {
			t.Setenv(tinvestReadOnlyTokenEnv, token)
			provider, err := configuredTInvestCorporateActionProvider()
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
	t.Setenv(tinvestCorporateActionsEnabledEnv, "true")
	t.Setenv(tinvestReadOnlyTokenEnv, "test-readonly-token")
	provider, err := configuredTInvestCorporateActionProvider()
	if err != nil {
		t.Fatalf("configuration: %v", err)
	}
	if provider == nil {
		t.Fatal("expected explicitly enabled provider")
	}
}
