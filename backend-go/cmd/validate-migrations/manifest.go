package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type strictInt int64

func (v *strictInt) UnmarshalJSON(b []byte) error {
	s := string(b)
	if s == "null" || !regexp.MustCompile(`^-?(0|[1-9][0-9]*)$`).MatchString(s) {
		return fmt.Errorf("expected canonical JSON integer")
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	*v = strictInt(n)
	return nil
}

type Manifest struct {
	SchemaVersion strictInt           `json:"schema_version"`
	Legacy        *LegacyBaseline     `json:"legacy_baseline"`
	Enforced      []EnforcedMigration `json:"enforced_migrations"`
}
type LegacyBaseline struct {
	MaxID                string        `json:"max_id"`
	PolicyMetadataStatus string        `json:"policy_metadata_status"`
	Entries              []LegacyEntry `json:"entries"`
}
type LegacyEntry struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	UpSHA   string `json:"up_sha256"`
	DownSHA string `json:"down_sha256"`
}
type EnforcedMigration struct {
	ID                  string              `json:"id"`
	Name                string              `json:"name"`
	UpSHA               string              `json:"up_sha256"`
	DownSHA             string              `json:"down_sha256"`
	Owners              []string            `json:"owners"`
	Phase               string              `json:"phase"`
	Dependencies        *[]string           `json:"dependencies"`
	Risk                string              `json:"risk"`
	DataClassification  string              `json:"data_classification"`
	Reversibility       string              `json:"reversibility"`
	UpTransactionMode   string              `json:"up_transaction_mode"`
	DownTransactionMode string              `json:"down_transaction_mode"`
	UpExecution         *ExecutionMetadata  `json:"up_execution"`
	DownExecution       *ExecutionMetadata  `json:"down_execution"`
	UpImpact            []DDLImpact         `json:"up_ddl_impact"`
	DownImpact          []DDLImpact         `json:"down_ddl_impact"`
	Observability       *Observability      `json:"observability"`
	Monitoring          *Monitoring         `json:"monitoring"`
	Rollout             *Rollout            `json:"rollout"`
	ProductionRollback  *ProductionRollback `json:"production_rollback"`
	RollForward         *RollForward        `json:"roll_forward"`
	AuthorityRefs       *[]AuthorityRef     `json:"authority_refs"`
}
type ExecutionMetadata struct {
	ExpectedDuration strictInt `json:"expected_duration_seconds"`
	LockRisk         string    `json:"lock_risk"`
	LockTimeout      strictInt `json:"lock_timeout_ms"`
	StatementTimeout strictInt `json:"statement_timeout_ms"`
}
type DDLImpact struct {
	StatementSHA      string    `json:"statement_sha256"`
	StatementClass    string    `json:"statement_class"`
	EstimatedLockMode string    `json:"estimated_lock_mode"`
	AffectedRows      strictInt `json:"affected_rows_estimate"`
	DiskImpact        strictInt `json:"disk_impact_bytes_estimate"`
	WALImpact         strictInt `json:"wal_impact_bytes_estimate"`
	ReplicationImpact string    `json:"replication_impact"`
	OnlineStrategy    string    `json:"online_strategy"`
	AbortCondition    string    `json:"abort_condition"`
	EstimateBasis     string    `json:"estimate_basis"`
}
type Observation struct {
	Mode   string `json:"mode"`
	Signal string `json:"signal_or_method,omitempty"`
	Reason string `json:"reason,omitempty"`
}
type Observability struct {
	RowsOrBatches               *Observation `json:"rows_or_batches"`
	LockWait                    *Observation `json:"lock_wait"`
	StatementDuration           *Observation `json:"statement_duration"`
	ReplicationLag              *Observation `json:"replication_lag"`
	WALGrowth                   *Observation `json:"wal_growth"`
	DiskGrowth                  *Observation `json:"disk_growth"`
	ValidationMismatches        *Observation `json:"validation_mismatches"`
	RetryPauseAbortReason       *Observation `json:"retry_pause_abort_reason"`
	ChangeDeploymentCorrelation *Observation `json:"change_deployment_correlation"`
}
type Monitoring struct {
	Signals          []string `json:"signals"`
	SuccessCondition string   `json:"success_condition"`
	AbortCondition   string   `json:"abort_condition"`
}
type Rollout struct {
	Mode    string   `json:"mode"`
	Metrics []string `json:"metrics"`
}
type ProductionRollback struct {
	Strategy     string `json:"strategy"`
	Procedure    string `json:"procedure"`
	Verification string `json:"verification"`
}
type RollForward struct {
	Procedure    string `json:"procedure"`
	Verification string `json:"verification"`
}
type AuthorityRef struct {
	Kind       string `json:"kind"`
	Path       string `json:"path"`
	ContentSHA string `json:"content_sha256"`
}

func decodeManifest(b []byte) (*Manifest, error) {
	if !validUTF8NoBOMNUL(b) {
		return nil, verr("MIG004_MANIFEST_JSON", "R003", "manifest must be UTF-8 without BOM/NUL")
	}
	if err := rejectDuplicateJSONKeys(b); err != nil {
		return nil, verr("MIG004_MANIFEST_JSON", "R003", err.Error())
	}
	if err := requireJSONFields(b, reflect.TypeOf(Manifest{})); err != nil {
		return nil, verr("MIG004_MANIFEST_JSON", "R003", err.Error())
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	var m Manifest
	if err := dec.Decode(&m); err != nil {
		return nil, verr("MIG004_MANIFEST_JSON", "R003", err.Error())
	}
	if err := ensureJSONEOF(dec); err != nil {
		return nil, verr("MIG004_MANIFEST_JSON", "R003", err.Error())
	}
	if m.Legacy == nil || m.Enforced == nil {
		return nil, verr("MIG004_MANIFEST_JSON", "R003", "required top-level field missing or null")
	}
	return &m, nil
}
func requireJSONFields(b []byte, typ reflect.Type) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	var raw any
	if err := dec.Decode(&raw); err != nil {
		return err
	}
	return requireValueFields(raw, typ, "$")
}
func requireValueFields(raw any, typ reflect.Type, path string) error {
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	switch typ.Kind() {
	case reflect.Struct:
		if typ == reflect.TypeOf(strictInt(0)) {
			return nil
		}
		obj, ok := raw.(map[string]any)
		if !ok {
			return nil
		} // type errors are owned by strict decode
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			tag := f.Tag.Get("json")
			if tag == "-" {
				continue
			}
			parts := strings.Split(tag, ",")
			key := parts[0]
			if key == "" {
				key = f.Name
			}
			optional := false
			for _, x := range parts[1:] {
				if x == "omitempty" {
					optional = true
				}
			}
			v, exists := obj[key]
			if !exists {
				if optional {
					continue
				}
				return fmt.Errorf("missing required field %s.%s", path, key)
			}
			if err := requireValueFields(v, f.Type, path+"."+key); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		a, ok := raw.([]any)
		if !ok {
			return nil
		}
		for i, v := range a {
			if err := requireValueFields(v, typ.Elem(), fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return err
			}
		}
	}
	return nil
}
func ensureJSONEOF(dec *json.Decoder) error {
	var x any
	if err := dec.Decode(&x); err == io.EOF {
		return nil
	}
	return fmt.Errorf("trailing JSON token/document")
}
func rejectDuplicateJSONKeys(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	if err := walkJSON(dec); err != nil {
		return err
	}
	if err := ensureJSONEOF(dec); err != nil {
		return err
	}
	return nil
}
func walkJSON(dec *json.Decoder) error {
	t, err := dec.Token()
	if err != nil {
		return err
	}
	switch d := t.(type) {
	case json.Delim:
		switch d {
		case '{':
			seen := map[string]struct{}{}
			for dec.More() {
				kt, err := dec.Token()
				if err != nil {
					return err
				}
				k, ok := kt.(string)
				if !ok {
					return fmt.Errorf("object key is not string")
				}
				if _, dup := seen[k]; dup {
					return fmt.Errorf("duplicate JSON key %q", k)
				}
				seen[k] = struct{}{}
				if err := walkJSON(dec); err != nil {
					return err
				}
			}
			_, err = dec.Token()
			return err
		case '[':
			for dec.More() {
				if err := walkJSON(dec); err != nil {
					return err
				}
			}
			_, err = dec.Token()
			return err
		default:
			return fmt.Errorf("unexpected delimiter")
		}
	default:
		if t == nil {
			return fmt.Errorf("JSON null is forbidden")
		}
		return nil
	}
}
func validateExecution(e *ExecutionMetadata) error {
	if e == nil {
		return verr("MIG009_METADATA", "R012", "execution metadata missing")
	}
	if !inBound("expected_duration_seconds", int64(e.ExpectedDuration)) {
		return verr("MIG009_METADATA", "R036", "expected_duration_seconds out of bounds")
	}
	if !in(finite.lockRisk, e.LockRisk) {
		return verr("MIG009_METADATA", "R012", "invalid lock_risk")
	}
	if !inBound("lock_timeout_ms", int64(e.LockTimeout)) || !inBound("statement_timeout_ms", int64(e.StatementTimeout)) || e.StatementTimeout < e.LockTimeout {
		return verr("MIG009_METADATA", "R012", "invalid timeout bounds/order")
	}
	return nil
}
func validateObservation(o *Observability) (map[string]bool, error) {
	if o == nil {
		return nil, verr("MIG019_OBSERVABILITY", "R014", "observability missing")
	}
	vals := map[string]*Observation{"rows_or_batches": o.RowsOrBatches, "lock_wait": o.LockWait, "statement_duration": o.StatementDuration, "replication_lag": o.ReplicationLag, "wal_growth": o.WALGrowth, "disk_growth": o.DiskGrowth, "validation_mismatches": o.ValidationMismatches, "retry_pause_abort_reason": o.RetryPauseAbortReason, "change_deployment_correlation": o.ChangeDeploymentCorrelation}
	measured := map[string]bool{}
	for _, k := range observationKeys {
		v := vals[k]
		if v == nil {
			return nil, verr("MIG019_OBSERVABILITY", "R014", "missing observability category "+k)
		}
		wantMeasured := in(measuredObservationKeys, k)
		if wantMeasured {
			if v.Mode != "measured" || !nonEmptyOpen(v.Signal) || v.Reason != "" {
				return nil, verr("MIG019_OBSERVABILITY", "R014", "invalid measured category "+k)
			}
			measured[k] = true
		} else if v.Mode != "not_applicable" || v.Signal != "" || !nonEmptyOpen(v.Reason) {
			return nil, verr("MIG019_OBSERVABILITY", "R014", "invalid not_applicable category "+k)
		}
	}
	return measured, nil
}
func validateRefPath(s string) bool {
	if len(s) < 1 || len(s) > 1024 || strings.HasPrefix(s, "/") || strings.HasSuffix(s, "/") || strings.Contains(s, "//") {
		return false
	}
	for _, seg := range strings.Split(s, "/") {
		if seg == "." || seg == ".." || len(seg) < 1 || len(seg) > 255 {
			return false
		}
		for _, r := range seg {
			if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '.' || r == '_' || r == '-') {
				return false
			}
		}
	}
	return true
}
