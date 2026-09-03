package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

const (
	manifestRel   = "infrastructure/postgres/migrations/policy_manifest.json"
	migrationsRel = "infrastructure/postgres/migrations"
)

var (
	fileRE          = regexp.MustCompile(`^([0-9]{6})_([a-z0-9]+(?:_[a-z0-9]+)*)\.(up|down)\.sql$`)
	idRE            = regexp.MustCompile(`^[0-9]{6}$`)
	shaRE           = regexp.MustCompile(`^[0-9a-f]{64}$`)
	identRE         = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)
	canonUnsignedRE = regexp.MustCompile(`^(0|[1-9][0-9]*)$`)
	canonPositiveRE = regexp.MustCompile(`^[1-9][0-9]*$`)
	canonIntegerRE  = regexp.MustCompile(`^(0|-?[1-9][0-9]*)$`)
)
var colIDDisallowed = makeSet(strings.Fields(`
all analyse analyze and any array as asc asymmetric authorization binary both case cast check collate collation column concurrently constraint create cross current_catalog current_date current_role current_schema current_time current_timestamp current_user default deferrable desc distinct do else end except false fetch for foreign freeze from full grant group having ilike in initially inner intersect into is isnull join lateral leading left like limit localtime localtimestamp natural not notnull null offset on only or order outer overlaps placing primary references returning right select session_user similar some symmetric system_user table tablesample then to trailing true union unique user using variadic verbose when where window with
`))
var finite = struct {
	phase, risk, classification, reversibility, transactionMode, class, lockMode, repl, online, obsMode, rollout, rollback, authKind, owner, lockRisk map[string]struct{}
}{
	phase:           makeSet([]string{"expand", "populate", "switch", "validate", "contract"}),
	risk:            makeSet([]string{"low", "medium", "high", "destructive"}),
	classification:  makeSet([]string{"schema_only", "financial", "identity_personal", "sensitive", "mixed"}),
	reversibility:   makeSet([]string{"disposable_down_exact_inverse"}),
	transactionMode: makeSet([]string{"transactional", "non_transactional"}),
	class:           makeSet([]string{"create_table", "add_column", "add_check_constraint", "add_foreign_key", "create_index", "create_unique_index", "create_index_concurrently", "create_unique_index_concurrently", "drop_table", "drop_column", "drop_constraint", "drop_index", "drop_index_concurrently"}),
	lockMode:        makeSet([]string{"access_share", "row_share", "row_exclusive", "share_update_exclusive", "share", "share_row_exclusive", "exclusive", "access_exclusive"}),
	repl:            makeSet([]string{"none", "low", "medium", "high"}),
	online:          makeSet([]string{"concurrent", "same_migration_new_table", "not_applicable"}),
	obsMode:         makeSet([]string{"measured", "not_applicable"}),
	rollout:         makeSet([]string{"standard", "staged"}),
	rollback:        makeSet([]string{"application_or_config_rollback", "leave_additive_structure_unused"}),
	authKind:        makeSet([]string{"adr", "security_privacy_review", "golden_vectors", "restore_rehearsal", "staged_rollout"}),
	owner:           makeSet([]string{"identity", "investment", "analytics", "audit"}),
	lockRisk:        makeSet([]string{"low", "medium", "high"}),
}

type intBound struct{ lo, hi int64 }

var integerBounds = map[string]intBound{
	"expected_duration_seconds":  {1, 86400},
	"lock_timeout_ms":            {1, 86400000},
	"statement_timeout_ms":       {1, 86400000},
	"affected_rows_estimate":     {0, 9223372036854775807},
	"disk_impact_bytes_estimate": {0, 9223372036854775807},
	"wal_impact_bytes_estimate":  {0, 9223372036854775807},
}

func inBound(name string, v int64) bool { b := integerBounds[name]; return v >= b.lo && v <= b.hi }

var legacyExpected = []LegacyEntry{
	{"000001", "stage_03_01_vertical_slice", "5ad39fb45af4f50f6707de0ba8b59d470cf7e4869fea8e7e63273710c8ed1741", "b136333141e173ea4743e6f66bf512ffa165615e90988a8ab2b5a9a989dbaa9e"},
	{"000002", "stage_03_11_auth_privacy", "ee6aa96a65c88883197770d1d82f75da8a78dc356c7bb5ce2763354ac7513b6b", "da068b69ece81e704415b07a99a26597c4c2567767d1cfdceb37180ae7aa39e8"},
	{"000003", "stage_03_16_transaction_source_provenance", "805b352eb29832de7cd2af8596d91a6fcc3a19455cda8af98564bd2c955a8aae", "cf04624f8c7071e95325b7c4ace6e2f8e0f6e733c4d1e3991047907084d4bced"},
	{"000004", "stage_03_27_import_financial_identity", "bf78ef0c6fd4c9c04a3f30b6ee3b65b6c099093d106d639b0e2d48ca98887cb2", "1866ad553c04e418344ceca8110593089356460bdb81f5f8ba7a457679328c5a"},
	{"000005", "stage_03_28_auth_security", "7615b12d8cffc2803fde393b8b77fd911bb0d751cc42f237590ffb9b7500b1e5", "f9d3a744a998af8ebd60359b0c895e8c1dab5bae07428df80aa4bc5763083533"},
	{"000006", "stage_03_32_idempotency_replay", "ad4e01abd25c65b1bd4b9d0a39ad2720102ffc6452d63e996729ba73967b80c6", "a85619479ff26d451bc222043da9994a329ce8bcbb64cb35fe9e72b797ffea15"},
	{"000007", "stage_03_38_operational_retention", "d145e2f10657d3b499200aa55c960dc6986ac14fe1f101c2538cc9e1d880f83f", "51e9994af57c6fc9db7a1c23f241a9f7e49de9c0ea106d943f1ef953844ddb6b"},
}
var observationKeys = []string{"rows_or_batches", "lock_wait", "statement_duration", "replication_lag", "wal_growth", "disk_growth", "validation_mismatches", "retry_pause_abort_reason", "change_deployment_correlation"}
var measuredObservationKeys = makeSet([]string{"lock_wait", "statement_duration", "replication_lag", "wal_growth", "disk_growth", "retry_pause_abort_reason", "change_deployment_correlation"})

// StableError intentionally carries only non-sensitive structural context.
type StableError struct {
	Code, Rule, MigrationID, Direction, Path, EffectID, ControlID, Detail string
	StatementIndex                                                        int
}

func (e *StableError) Error() string {
	parts := []string{e.Code, e.Rule}
	if e.MigrationID != "" {
		parts = append(parts, "migration="+e.MigrationID)
	}
	if e.Direction != "" {
		parts = append(parts, "direction="+e.Direction)
	}
	if e.Path != "" {
		parts = append(parts, "path="+e.Path)
	}
	if e.EffectID != "" {
		parts = append(parts, "effect="+e.EffectID)
	}
	if e.ControlID != "" {
		parts = append(parts, "control="+e.ControlID)
	}
	if e.StatementIndex > 0 {
		parts = append(parts, fmt.Sprintf("statement=%d", e.StatementIndex))
	}
	if e.Detail != "" {
		parts = append(parts, e.Detail)
	}
	return strings.Join(parts, ": ")
}
func verr(code, rule, detail string) error {
	return &StableError{Code: code, Rule: rule, Detail: detail}
}
func makeSet[T ~string](xs []T) map[string]struct{} {
	m := map[string]struct{}{}
	for _, x := range xs {
		m[string(x)] = struct{}{}
	}
	return m
}
func in(set map[string]struct{}, v string) bool { _, ok := set[v]; return ok }
func sortedKeys(set map[string]struct{}) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
func hashBytes(b []byte) string     { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func validSHA(s string) bool        { return shaRE.MatchString(s) }
func asciiTrim(s string) string     { return strings.Trim(s, "\t\n\v\f\r ") }
func nonEmptyOpen(s string) bool    { return asciiTrim(s) != "" }
func validIdentifier(s string) bool { return identRE.MatchString(s) && !in(colIDDisallowed, s) }
func validUTF8NoBOMNUL(b []byte) bool {
	return utf8.Valid(b) && !strings.HasPrefix(string(b), "\ufeff") && !strings.ContainsRune(string(b), '\x00')
}
func riskRank(v string) int {
	switch v {
	case "low":
		return 1
	case "medium":
		return 2
	case "high":
		return 3
	case "destructive":
		return 4
	}
	return 0
}
func contractInventory(root string) error {
	plan, err := os.ReadFile(filepath.Join(root, "docs", "stages", "STAGE_03_54_P3_08_MIGRATION_VALIDATOR_PLAN.md"))
	if err != nil || hashBytes(plan) != "c266d5b7c867d2e6847bbe169b0a890a997a81f886f1876117117e52c85aecba" {
		return verr("MIG025_TEST_CONTRACT", "R022", "canonical Stage 3.54 contract identity drift")
	}
	if len(observationKeys) != 9 || len(finite.authKind) != 5 || len(finite.owner) != 4 || len(finite.class) != 13 || len(finite.phase) != 5 || len(finite.risk) != 4 || len(finite.classification) != 5 || len(finite.reversibility) != 1 || len(finite.transactionMode) != 2 || len(colIDDisallowed) != 101 {
		return verr("MIG025_TEST_CONTRACT", "R022", "finite contract inventory drift")
	}
	if hashBytes([]byte(strings.Join(sortedKeys(colIDDisallowed), "\n")+"\n")) != "3a9027604ec759856e3f9fdbaadaccc4588c00b213328ab5ca0018231448e0d6" {
		return verr("MIG026_SEMANTIC_FREEZE", "R058", "PostgreSQL ColId projection drift")
	}
	return nil
}
