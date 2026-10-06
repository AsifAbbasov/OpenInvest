package main

import "testing"

func TestDeploymentAbuseControlFailsClosedOutsideDevelopment(t *testing.T) {
	for _, env := range []string{"", "production", "staging"} {
		t.Run(env, func(t *testing.T) {
			t.Setenv("OPENINVEST_ENV", env)
			t.Setenv(deploymentAbuseControlEnv, "")
			if _, err := configuredDeploymentAbuseControl(); err == nil {
				t.Fatalf("missing %s must fail outside development/local", deploymentAbuseControlEnv)
			}
			t.Setenv(deploymentAbuseControlEnv, "edge-v1")
			if got, err := configuredDeploymentAbuseControl(); err != nil || got != "edge-v1" {
				t.Fatalf("edge-v1 ownership got=%q err=%v", got, err)
			}
		})
	}
}

func TestDeploymentAbuseControlDevelopmentDefaultIsExplicitlyProcessLocal(t *testing.T) {
	t.Setenv("OPENINVEST_ENV", "development")
	t.Setenv(deploymentAbuseControlEnv, "")
	got, err := configuredDeploymentAbuseControl()
	if err != nil {
		t.Fatal(err)
	}
	if got != "process-local-development" {
		t.Fatalf("development owner=%q", got)
	}
}

func TestDeploymentAbuseControlRejectsUnknownOwnership(t *testing.T) {
	t.Setenv("OPENINVEST_ENV", "production")
	t.Setenv(deploymentAbuseControlEnv, "redis")
	if _, err := configuredDeploymentAbuseControl(); err == nil {
		t.Fatal("unknown deployment-global ownership was accepted")
	}
}
