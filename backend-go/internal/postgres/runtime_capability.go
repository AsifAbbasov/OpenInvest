package postgres

import (
	"fmt"
	"strings"
)

type RuntimeCapabilityProfile string

const (
	RuntimeCapabilityR0 RuntimeCapabilityProfile = "R0"
	RuntimeCapabilityR1 RuntimeCapabilityProfile = "R1"
	RuntimeCapabilityR2 RuntimeCapabilityProfile = "R2"
)

func ParseRuntimeCapabilityProfile(raw string) (RuntimeCapabilityProfile, error) {
	profile := RuntimeCapabilityProfile(strings.ToUpper(strings.TrimSpace(raw)))
	if profile == "" {
		return RuntimeCapabilityR0, nil
	}
	switch profile {
	case RuntimeCapabilityR0, RuntimeCapabilityR1, RuntimeCapabilityR2:
		return profile, nil
	default:
		return "", fmt.Errorf("%w: unknown runtime capability profile %q", ErrUnsafeRuntimeDatabaseRole, raw)
	}
}

func (profile RuntimeCapabilityProfile) valid() bool {
	switch profile {
	case RuntimeCapabilityR0, RuntimeCapabilityR1, RuntimeCapabilityR2:
		return true
	default:
		return false
	}
}

func runtimeRelationCapabilitiesForProfile(profile RuntimeCapabilityProfile) []runtimeRelationCapability {
	result := append([]runtimeRelationCapability(nil), runtimeRelationCapabilities...)
	if profile == RuntimeCapabilityR0 {
		return result
	}
	result = append(result,
		runtimeRelationCapability{Schema: "analytics", Name: "replay_policy_generations", Select: true},
		runtimeRelationCapability{Schema: "analytics", Name: "replay_policy_events", Select: true},
		runtimeRelationCapability{Schema: "analytics", Name: "portfolio_replay_epochs", Select: true, Insert: profile == RuntimeCapabilityR2},
		runtimeRelationCapability{Schema: "analytics", Name: "portfolio_replay_positions", Select: true, Insert: profile == RuntimeCapabilityR2},
		runtimeRelationCapability{Schema: "analytics", Name: "portfolio_replay_financial_state", Select: true, Insert: profile == RuntimeCapabilityR2},
	)
	return result
}

// OpenOwnerWithApplicationCapability is for the explicitly selected local/development schema-owner path.
// It does not replace OpenRuntimeWithCapability in staging/production and performs no least-privilege waiver.
func OpenOwnerWithApplicationCapability(databaseURL string, profile RuntimeCapabilityProfile) (*Store, error) {
	if !profile.valid() {
		return nil, fmt.Errorf("%w: unknown application capability profile %q", ErrUnsafeRuntimeDatabaseRole, profile)
	}
	store, err := Open(databaseURL)
	if err != nil {
		return nil, err
	}
	store.runtimeCapabilityProfile = profile
	return store, nil
}
