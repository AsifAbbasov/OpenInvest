package main

import "testing"

func TestM12ProductionOwnershipAcknowledgementsAreNotExternalProof(t *testing.T) {
	t.Setenv("OPENINVEST_ENV", "production")
	t.Setenv(deploymentGlobalAbuseControlEnv, verifiedDeploymentGlobalAbuseControl)
	t.Setenv("OPENINVEST_DEV_AUTH_BYPASS", "false")
	t.Setenv("OPENINVEST_REFRESH_COOKIE_INSECURE", "false")
	t.Setenv("OPENINVEST_ALLOW_EPHEMERAL_ACCESS_TOKEN_SECRET", "false")
	if err := validateRuntimeSafety("postgres://runtime-user:placeholder@db.invalid/openinvest?sslmode=require"); err != nil {
		t.Fatalf("production safety acknowledgement unexpectedly rejected: %v", err)
	}
	t.Log("M12_DEPLOYMENT_GLOBAL_ABUSE_CONTROL_EXTERNAL_SHARED_STATE_PROOF=NO startup_gate=ACKNOWLEDGEMENT_ONLY")

	t.Setenv(tinvestCorporateActionsEnabledEnv, "true")
	t.Setenv(tinvestGlobalBudgetOwnerEnv, verifiedTInvestGlobalBudgetOwner)
	t.Setenv(tinvestReadOnlyTokenEnv, "module12-test-readonly-token")
	provider, err := configuredTInvestCorporateActionProvider()
	if err != nil || provider == nil {
		t.Fatalf("provider activation acknowledgement: provider=%v err=%v", provider, err)
	}
	t.Log("M12_PROVIDER_GLOBAL_BUDGET_EXTERNAL_ACCOUNT_LEVEL_PROOF=NO startup_gate=ACKNOWLEDGEMENT_ONLY provider_budget_state=PROCESS_LOCAL")

	t.Setenv(apiListenAddressEnv, "0.0.0.0:8080")
	address, err := configuredAPIListenAddress()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("M12_PRODUCTION_WILDCARD_LISTEN_EXPLICITLY_ALLOWED=%t address=%s", address == "0.0.0.0:8080", address)
}
