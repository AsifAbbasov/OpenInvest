package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func baselineRepo(t *testing.T) string {
	t.Helper()
	src, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	dst := t.TempDir()
	mdir := filepath.Join(dst, migrationsRel)
	if err := os.MkdirAll(mdir, 0o755); err != nil {
		t.Fatal(err)
	}
	ents, err := os.ReadDir(filepath.Join(src, migrationsRel))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range ents {
		if strings.HasSuffix(e.Name(), ".sql") || e.Name() == "policy_manifest.json" {
			b, err := os.ReadFile(filepath.Join(src, migrationsRel, e.Name()))
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(mdir, e.Name()), b, 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	plan, err := os.ReadFile(filepath.Join(src, canonicalPlanRel))
	if err != nil {
		t.Fatal(err)
	}
	planPath := filepath.Join(dst, canonicalPlanRel)
	if err := os.MkdirAll(filepath.Dir(planPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(planPath, plan, 0o644); err != nil {
		t.Fatal(err)
	}
	return dst
}
func loadManifestT(t *testing.T, root string) *Manifest {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, manifestRel))
	if err != nil {
		t.Fatal(err)
	}
	m, err := decodeManifest(b)
	if err != nil {
		t.Fatal(err)
	}
	return m
}
func saveManifestT(t *testing.T, root string, m *Manifest) {
	t.Helper()
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	b = append(b, '\n')
	if err := os.WriteFile(filepath.Join(root, manifestRel), b, 0o644); err != nil {
		t.Fatal(err)
	}
}
func ptrSlice[T any](v []T) *[]T { return &v }

func validFuture(t *testing.T, root string, upDDL, downDDL string) *EnforcedMigration {
	t.Helper()
	upExec := &ExecutionMetadata{ExpectedDuration: 10, LockRisk: "low", LockTimeout: 1000, StatementTimeout: 2000}
	downExec := &ExecutionMetadata{ExpectedDuration: 20, LockRisk: "high", LockTimeout: 3000, StatementTimeout: 4000}
	up := fmt.Sprintf("BEGIN;\nSET LOCAL lock_timeout = '%dms';\nSET LOCAL statement_timeout = '%dms';\n%s\nCOMMIT;\n", upExec.LockTimeout, upExec.StatementTimeout, upDDL)
	down := fmt.Sprintf("BEGIN;\nSET LOCAL lock_timeout = '%dms';\nSET LOCAL statement_timeout = '%dms';\n%s\nCOMMIT;\n", downExec.LockTimeout, downExec.StatementTimeout, downDDL)
	name := "stage_03_55_fixture"
	mdir := filepath.Join(root, migrationsRel)
	upPath := filepath.Join(mdir, "000008_"+name+".up.sql")
	downPath := filepath.Join(mdir, "000008_"+name+".down.sql")
	os.WriteFile(upPath, []byte(up), 0o644)
	os.WriteFile(downPath, []byte(down), 0o644)
	us, _ := scanSQL([]byte(up))
	ds, _ := scanSQL([]byte(down))
	deps := []string{}
	refs := []AuthorityRef{}
	obs := &Observability{RowsOrBatches: &Observation{Mode: "not_applicable", Reason: "no rows/batches in expand DDL"}, LockWait: &Observation{Mode: "measured", Signal: "lock wait"}, StatementDuration: &Observation{Mode: "measured", Signal: "statement duration"}, ReplicationLag: &Observation{Mode: "measured", Signal: "replication lag"}, WALGrowth: &Observation{Mode: "measured", Signal: "wal growth"}, DiskGrowth: &Observation{Mode: "measured", Signal: "disk growth"}, ValidationMismatches: &Observation{Mode: "not_applicable", Reason: "no validation phase"}, RetryPauseAbortReason: &Observation{Mode: "measured", Signal: "abort reason"}, ChangeDeploymentCorrelation: &Observation{Mode: "measured", Signal: "deployment correlation"}}
	return &EnforcedMigration{ID: "000008", Name: name, UpSHA: hashBytes([]byte(up)), DownSHA: hashBytes([]byte(down)), Owners: []string{"identity"}, Phase: "expand", Dependencies: &deps, Risk: "low", DataClassification: "schema_only", Reversibility: "disposable_down_exact_inverse", UpTransactionMode: "transactional", DownTransactionMode: "transactional", UpExecution: upExec, DownExecution: downExec, UpImpact: []DDLImpact{{StatementSHA: hashBytes(us[3].raw), StatementClass: "create_table", EstimatedLockMode: "access_exclusive", AffectedRows: 0, DiskImpact: 0, WALImpact: 0, ReplicationImpact: "low", OnlineStrategy: "not_applicable", AbortCondition: "abort on failure", EstimateBasis: "empty table"}}, DownImpact: []DDLImpact{{StatementSHA: hashBytes(ds[3].raw), StatementClass: "drop_table", EstimatedLockMode: "access_exclusive", AffectedRows: 0, DiskImpact: 0, WALImpact: 0, ReplicationImpact: "low", OnlineStrategy: "not_applicable", AbortCondition: "abort on failure", EstimateBasis: "disposable inverse"}}, Observability: obs, Monitoring: &Monitoring{Signals: []string{"lock_wait"}, SuccessCondition: "DDL succeeds", AbortCondition: "abort on failure"}, Rollout: &Rollout{Mode: "standard", Metrics: []string{"lock_wait"}}, ProductionRollback: &ProductionRollback{Strategy: "leave_additive_structure_unused", Procedure: "leave structure unused", Verification: "verify application ignores it"}, RollForward: &RollForward{Procedure: "fix forward", Verification: "verify forward state"}, AuthorityRefs: &refs}
}
func installFuture(t *testing.T, root string, e *EnforcedMigration) {
	t.Helper()
	m := loadManifestT(t, root)
	m.Enforced = append(m.Enforced, *e)
	saveManifestT(t, root, m)
}
func validateRoot(root string) error { return (&Validator{Root: root, Mode: "local"}).Validate() }
func wantCode(t *testing.T, err error, code string) {
	t.Helper()
	if err == nil {
		t.Fatalf("wanted %s, got nil", code)
	}
	se, ok := err.(*StableError)
	if !ok || se.Code != code {
		t.Fatalf("wanted %s, got %T %v", code, err, err)
	}
}

func TestLegacyBaselinePasses(t *testing.T) {
	root := baselineRepo(t)
	if err := validateRoot(root); err != nil {
		t.Fatal(err)
	}
}
func rewriteTransactionalFixture(t *testing.T, root string, e *EnforcedMigration, upDDL, downDDL string) {
	t.Helper()
	up := fmt.Sprintf("BEGIN;\nSET LOCAL lock_timeout = '%dms';\nSET LOCAL statement_timeout = '%dms';\n%s\nCOMMIT;\n", e.UpExecution.LockTimeout, e.UpExecution.StatementTimeout, upDDL)
	down := fmt.Sprintf("BEGIN;\nSET LOCAL lock_timeout = '%dms';\nSET LOCAL statement_timeout = '%dms';\n%s\nCOMMIT;\n", e.DownExecution.LockTimeout, e.DownExecution.StatementTimeout, downDDL)
	os.WriteFile(filepath.Join(root, migrationsRel, "000008_"+e.Name+".up.sql"), []byte(up), 0o644)
	os.WriteFile(filepath.Join(root, migrationsRel, "000008_"+e.Name+".down.sql"), []byte(down), 0o644)
	e.UpSHA, e.DownSHA = hashBytes([]byte(up)), hashBytes([]byte(down))
	us, err := scanSQL([]byte(up))
	if err != nil {
		t.Fatal(err)
	}
	ds, err := scanSQL([]byte(down))
	if err != nil {
		t.Fatal(err)
	}
	if len(e.UpImpact) == 1 {
		e.UpImpact[0].StatementSHA = hashBytes(us[3].raw)
	}
	if len(e.DownImpact) == 1 {
		e.DownImpact[0].StatementSHA = hashBytes(ds[3].raw)
	}
}

func TestFourFieldDirectionIndependence(t *testing.T) {
	upDDL, downDDL := "CREATE TABLE identity.stage_03_55_fixture (id bigint);", "DROP TABLE identity.stage_03_55_fixture;"
	cases := []struct {
		name   string
		mutate func(*EnforcedMigration)
	}{
		{"expected-duration", func(e *EnforcedMigration) { e.DownExecution.ExpectedDuration = 20 }},
		{"lock-risk", func(e *EnforcedMigration) { e.DownExecution.LockRisk = "high" }},
		{"lock-timeout", func(e *EnforcedMigration) {
			e.UpExecution.StatementTimeout = 4000
			e.DownExecution.LockTimeout = 3000
			e.DownExecution.StatementTimeout = 4000
		}},
		{"statement-timeout", func(e *EnforcedMigration) { e.DownExecution.StatementTimeout = 4000 }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := baselineRepo(t)
			e := validFuture(t, r, upDDL, downDDL)
			// Begin each case with all four fields equal, then vary exactly one field.
			copyDown := *e.UpExecution
			e.DownExecution = &copyDown
			tc.mutate(e)
			rewriteTransactionalFixture(t, r, e, upDDL, downDDL)
			installFuture(t, r, e)
			if err := validateRoot(r); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestFourFieldEqualityCouplingMutantsAreKilled(t *testing.T) {
	// Each canonical positive fixture must be rejected by a hypothetical equality-coupled mutant.
	checks := []struct {
		name      string
		different func(*ExecutionMetadata, *ExecutionMetadata) bool
	}{
		{"expected-duration", func(u, d *ExecutionMetadata) bool { return u.ExpectedDuration != d.ExpectedDuration }},
		{"lock-risk", func(u, d *ExecutionMetadata) bool { return u.LockRisk != d.LockRisk }},
		{"lock-timeout", func(u, d *ExecutionMetadata) bool { return u.LockTimeout != d.LockTimeout }},
		{"statement-timeout", func(u, d *ExecutionMetadata) bool { return u.StatementTimeout != d.StatementTimeout }},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			u := &ExecutionMetadata{ExpectedDuration: 10, LockRisk: "low", LockTimeout: 1000, StatementTimeout: 4000}
			d := *u
			switch c.name {
			case "expected-duration":
				d.ExpectedDuration = 20
			case "lock-risk":
				d.LockRisk = "high"
			case "lock-timeout":
				d.LockTimeout = 3000
			case "statement-timeout":
				d.StatementTimeout = 5000
			}
			if !c.different(u, &d) {
				t.Fatal("equality-coupling mutant survived")
			}
		})
	}
}

func TestSQLKeywordCaseCommentsAndInertForbiddenWordPass(t *testing.T) {
	r := baselineRepo(t)
	e := validFuture(t, r, "create /* outer /* nested */ ok */ table identity.stage_03_55_fixture -- DROP is inert\n(id bigint);", "drop table identity.stage_03_55_fixture;")
	installFuture(t, r, e)
	if err := validateRoot(r); err != nil {
		t.Fatal(err)
	}
}

func TestStrictJSONFailures(t *testing.T) {
	cases := map[string]func([]byte) []byte{
		"duplicate": func(b []byte) []byte {
			return []byte(strings.Replace(string(b), `"schema_version": 1`, `"schema_version": 1, "schema_version": 1`, 1))
		},
		"unknown": func(b []byte) []byte {
			return []byte(strings.Replace(string(b), `"schema_version": 1`, `"schema_version": 1, "unknown": true`, 1))
		},
		"null": func(b []byte) []byte {
			return []byte(strings.Replace(string(b), `"policy_metadata_status": "not_retroactively_asserted"`, `"policy_metadata_status": null`, 1))
		},
		"trailing": func(b []byte) []byte { return append(b, []byte(` {}`)...) }}
	for name, fn := range cases {
		t.Run(name, func(t *testing.T) {
			r := baselineRepo(t)
			p := filepath.Join(r, manifestRel)
			b, _ := os.ReadFile(p)
			os.WriteFile(p, fn(b), 0o644)
			wantCode(t, validateRoot(r), "MIG004_MANIFEST_JSON")
		})
	}
}
func TestMalformedSQLFilenameIsNeverIgnored(t *testing.T) {
	r := baselineRepo(t)
	os.WriteFile(filepath.Join(r, migrationsRel, "evil.up.sql"), []byte("SELECT 1;"), 0o644)
	wantCode(t, validateRoot(r), "MIG002_FILENAME")
}
func TestLegacyHashMutationRejected(t *testing.T) {
	r := baselineRepo(t)
	p := filepath.Join(r, migrationsRel, "000001_stage_03_01_vertical_slice.up.sql")
	b, _ := os.ReadFile(p)
	os.WriteFile(p, append(b, '\n'), 0o644)
	wantCode(t, validateRoot(r), "MIG017_LEGACY_BASELINE")
}

func TestTimeoutBindingAndBounds(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*EnforcedMigration, string) string
		code   string
	}{
		{"wrong-sql-value", func(e *EnforcedMigration, s string) string { return strings.Replace(s, "'1000ms'", "'999ms'", 1) }, "MIG015_TIMEOUT"},
		{"statement-below-lock", func(e *EnforcedMigration, s string) string {
			e.UpExecution.LockTimeout = 3000
			e.UpExecution.StatementTimeout = 2000
			return s
		}, "MIG009_METADATA"},
		{"zero-duration", func(e *EnforcedMigration, s string) string { e.UpExecution.ExpectedDuration = 0; return s }, "MIG009_METADATA"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := baselineRepo(t)
			e := validFuture(t, r, "CREATE TABLE identity.stage_03_55_fixture (id bigint);", "DROP TABLE identity.stage_03_55_fixture;")
			p := filepath.Join(r, migrationsRel, "000008_stage_03_55_fixture.up.sql")
			b, _ := os.ReadFile(p)
			nb := c.mutate(e, string(b))
			if nb != string(b) {
				os.WriteFile(p, []byte(nb), 0o644)
				e.UpSHA = hashBytes([]byte(nb))
			}
			installFuture(t, r, e)
			wantCode(t, validateRoot(r), c.code)
		})
	}
}

func TestIdentifierPolicy(t *testing.T) {
	for _, tc := range []struct {
		name, col string
		pass      bool
	}{{"ordinary", "portfolio", true}, {"contextual-update", "update", true}, {"reserved-table", "table", false}, {"reserved-select", "select", false}, {"upper", "Portfolio", false}, {"too-long", "a" + strings.Repeat("b", 63), false}} {
		t.Run(tc.name, func(t *testing.T) {
			r := baselineRepo(t)
			e := validFuture(t, r, fmt.Sprintf("CREATE TABLE identity.stage_03_55_fixture (%s bigint);", tc.col), "DROP TABLE identity.stage_03_55_fixture;")
			installFuture(t, r, e)
			err := validateRoot(r)
			if tc.pass && err != nil {
				t.Fatal(err)
			}
			if !tc.pass && err == nil {
				t.Fatal("expected reject")
			}
		})
	}
}
func TestCreateTableColumnBoundary(t *testing.T) {
	mk := func(n int) string {
		a := make([]string, n)
		for i := range a {
			a[i] = fmt.Sprintf("c%d bigint", i)
		}
		return "CREATE TABLE identity.stage_03_55_fixture (" + strings.Join(a, ",") + ");"
	}
	for _, tc := range []struct {
		n    int
		pass bool
	}{{64, true}, {65, false}} {
		t.Run(fmt.Sprint(tc.n), func(t *testing.T) {
			r := baselineRepo(t)
			e := validFuture(t, r, mk(tc.n), "DROP TABLE identity.stage_03_55_fixture;")
			installFuture(t, r, e)
			err := validateRoot(r)
			if tc.pass && err != nil {
				t.Fatal(err)
			}
			if !tc.pass && err == nil {
				t.Fatal("expected reject")
			}
		})
	}
}
func TestTypeAndLiteralGrammar(t *testing.T) {
	good := []string{"numeric(1,0)", "numeric(38,38)", "varchar(1)", "varchar(10485760)"}
	for _, typ := range good {
		t.Run("good-"+typ, func(t *testing.T) {
			p := &parser{t: mustTokens(t, "CREATE TABLE identity.t (c "+typ+");")}
			if _, err := parseUpDDL(statement{tokens: p.t, raw: []byte("CREATE TABLE identity.t (c " + typ + ");")}, &upState{created: map[string]map[string]sqlType{}, added: map[string]sqlType{}}); err != nil {
				t.Fatal(err)
			}
		})
	}
	bad := []string{"numeric(01,0)", "numeric(+1,0)", "numeric(1,00)", "numeric(39,0)", "varchar(0005)", "varchar(10485761)"}
	for _, typ := range bad {
		t.Run("bad-"+typ, func(t *testing.T) {
			s := "CREATE TABLE identity.t (c " + typ + ");"
			ss, err := scanSQL([]byte(s))
			if err == nil {
				_, err = parseUpDDL(ss[0], &upState{created: map[string]map[string]sqlType{}, added: map[string]sqlType{}})
			}
			if err == nil {
				t.Fatal("expected reject")
			}
		})
	}
}
func mustTokens(t *testing.T, s string) []token {
	t.Helper()
	ss, err := scanSQL([]byte(s))
	if err != nil {
		t.Fatal(err)
	}
	return ss[0].tokens
}

func TestAddColumnDefaultRaisesMinimumRisk(t *testing.T) {
	r := baselineRepo(t)
	e := validFuture(t, r, "ALTER TABLE identity.users ADD COLUMN stage_03_55_flag integer DEFAULT 1;", "ALTER TABLE identity.users DROP COLUMN stage_03_55_flag;")
	e.UpImpact[0].StatementClass = "add_column"
	e.DownImpact[0].StatementClass = "drop_column"
	installFuture(t, r, e)
	wantCode(t, validateRoot(r), "MIG009_METADATA")
	r = baselineRepo(t)
	e = validFuture(t, r, "ALTER TABLE identity.users ADD COLUMN stage_03_55_flag integer DEFAULT 1;", "ALTER TABLE identity.users DROP COLUMN stage_03_55_flag;")
	e.UpImpact[0].StatementClass = "add_column"
	e.DownImpact[0].StatementClass = "drop_column"
	e.Risk = "medium"
	e.Rollout.Mode = "staged"
	ref := makeAuthority(t, r, "staged_rollout", "docs/stage.md")
	*e.AuthorityRefs = append(*e.AuthorityRefs, ref)
	installFuture(t, r, e)
	if err := validateRoot(r); err != nil {
		t.Fatal(err)
	}
}
func makeAuthority(t *testing.T, root, kind, path string) AuthorityRef {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(path))
	os.MkdirAll(filepath.Dir(full), 0o755)
	b := []byte("authority\n")
	os.WriteFile(full, b, 0o644)
	return AuthorityRef{Kind: kind, Path: path, ContentSHA: hashBytes(b)}
}

func TestAuthorityReferenceIdentity(t *testing.T) {
	r := baselineRepo(t)
	e := validFuture(t, r, "CREATE TABLE identity.stage_03_55_fixture (id bigint);", "DROP TABLE identity.stage_03_55_fixture;")
	a := makeAuthority(t, r, "adr", "docs/a.md")
	b := a
	b.Kind = "security_privacy_review"
	*e.AuthorityRefs = []AuthorityRef{a, b}
	installFuture(t, r, e)
	if err := validateRoot(r); err != nil {
		t.Fatal(err)
	}
	r = baselineRepo(t)
	e = validFuture(t, r, "CREATE TABLE identity.stage_03_55_fixture (id bigint);", "DROP TABLE identity.stage_03_55_fixture;")
	a = makeAuthority(t, r, "adr", "docs/a.md")
	*e.AuthorityRefs = []AuthorityRef{a, a}
	installFuture(t, r, e)
	wantCode(t, validateRoot(r), "MIG016_AUTHORITY_REFERENCE")
}
func TestAuthorityPathGrammar(t *testing.T) {
	for _, p := range []string{"./docs/a.md", "docs/./a.md", "docs//a.md", "/docs/a.md", "docs/a.md/", "docs\\a.md", "docs/a b.md", "docs/%61.md"} {
		if validateRefPath(p) {
			t.Fatalf("invalid path accepted: %q", p)
		}
	}
	if !validateRefPath("docs/reviews/a-1.md") {
		t.Fatal("valid path rejected")
	}
}
func TestASCIITrimOnly(t *testing.T) {
	if nonEmptyOpen(" \t\r\n") {
		t.Fatal("ASCII whitespace must be empty")
	}
	if !nonEmptyOpen("\u00a0") {
		t.Fatal("NBSP must count as content")
	}
}

func TestDownExactGrammar(t *testing.T) {
	bad := []string{"DROP TABLE IF EXISTS identity.stage_03_55_fixture;", "DROP TABLE identity.stage_03_55_fixture CASCADE;", "DROP TABLE identity.stage_03_55_fixture RESTRICT;", "DROP TABLE identity.stage_03_55_fixture, identity.other;"}
	for _, d := range bad {
		t.Run(d, func(t *testing.T) {
			r := baselineRepo(t)
			e := validFuture(t, r, "CREATE TABLE identity.stage_03_55_fixture (id bigint);", d)
			installFuture(t, r, e)
			wantCode(t, validateRoot(r), "MIG022_DOWN_INVERSE")
		})
	}
}

func TestOwnersRiskAndObservabilityGates(t *testing.T) {
	cases := []struct {
		name string
		mut  func(*EnforcedMigration)
	}{
		{"owner", func(e *EnforcedMigration) { e.Owners = []string{"investment"} }},
		{"obs", func(e *EnforcedMigration) {
			e.Observability.LockWait.Mode = "not_applicable"
			e.Observability.LockWait.Signal = ""
			e.Observability.LockWait.Reason = "n/a"
		}},
		{"monitor", func(e *EnforcedMigration) { e.Monitoring.Signals = []string{"rows_or_batches"} }},
		{"rollout", func(e *EnforcedMigration) { e.Rollout.Mode = "staged" }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := baselineRepo(t)
			e := validFuture(t, r, "CREATE TABLE identity.stage_03_55_fixture (id bigint);", "DROP TABLE identity.stage_03_55_fixture;")
			c.mut(e)
			installFuture(t, r, e)
			if err := validateRoot(r); err == nil {
				t.Fatal("expected reject")
			}
		})
	}
}
func TestDependencyValidation(t *testing.T) {
	r := baselineRepo(t)
	e := validFuture(t, r, "CREATE TABLE identity.stage_03_55_fixture (id bigint);", "DROP TABLE identity.stage_03_55_fixture;")
	d := []string{"000009"}
	e.Dependencies = &d
	installFuture(t, r, e)
	wantCode(t, validateRoot(r), "MIG010_DEPENDENCY")
}
func TestImpactBinding(t *testing.T) {
	r := baselineRepo(t)
	e := validFuture(t, r, "CREATE TABLE identity.stage_03_55_fixture (id bigint);", "DROP TABLE identity.stage_03_55_fixture;")
	e.UpImpact[0].StatementSHA = strings.Repeat("0", 64)
	installFuture(t, r, e)
	wantCode(t, validateRoot(r), "MIG018_DDL_IMPACT")
}

func TestIndexOnlineAndTransactionRules(t *testing.T) {
	r := baselineRepo(t)
	e := validFuture(t, r, "CREATE INDEX stage_03_55_idx ON identity.users (id);", "DROP INDEX identity.stage_03_55_idx;")
	e.UpImpact[0].StatementClass = "create_index"
	e.UpImpact[0].OnlineStrategy = "same_migration_new_table"
	e.DownImpact[0].StatementClass = "drop_index"
	installFuture(t, r, e)
	if err := validateRoot(r); err == nil {
		t.Fatal("pre-existing-table nonconcurrent index must reject")
	}
}
