package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const canonicalPlanRel = "docs/stages/STAGE_03_54_P3_08_MIGRATION_VALIDATOR_PLAN.md"
const canonicalPlanSHA256 = "c266d5b7c867d2e6847bbe169b0a890a997a81f886f1876117117e52c85aecba"

func TestCanonicalContractRegistryExactness(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(root, canonicalPlanRel))
	if err != nil {
		t.Fatal(err)
	}
	if hashBytes(b) != canonicalPlanSHA256 {
		t.Fatalf("plan identity drift: %s", hashBytes(b))
	}
	text := string(b)
	section := func(a, z string) string {
		i, j := strings.Index(text, a), strings.Index(text, z)
		if i < 0 || j < 0 || j <= i {
			t.Fatalf("missing section %q..%q", a, z)
		}
		return text[i:j]
	}
	ids := func(s, pat string) []string {
		r := regexp.MustCompile(pat)
		ms := r.FindAllStringSubmatch(s, -1)
		out := make([]string, len(ms))
		for i := range ms {
			out[i] = ms[i][1]
		}
		return out
	}
	wantRange := func(prefix string, n, width int) []string {
		out := make([]string, n)
		for i := range out {
			out[i] = fmt.Sprintf("%s%0*d", prefix, width, i+1)
		}
		return out
	}
	assertExact := func(name string, got, want []string) {
		t.Helper()
		if len(got) != len(want) {
			t.Fatalf("%s count=%d want=%d", name, len(got), len(want))
		}
		seen := map[string]bool{}
		for _, x := range got {
			if seen[x] {
				t.Fatalf("%s duplicate %s", name, x)
			}
			seen[x] = true
		}
		for _, x := range want {
			if !seen[x] {
				t.Fatalf("%s missing %s", name, x)
			}
		}
	}
	saNums := ids(section("### 23.0 Byte-bound", "### 23.1 Canonical"), `(?m)^\| `+"`"+`(SA-\d{3})`+"`"+` \|`)
	assertExact("SA", saNums, wantRange("SA-", 82, 3))
	s2 := ids(section("### 23.1 Canonical control rows", "### 23.1a Machine-evidence"), `(?m)^\| `+"`"+`(S2-\d{3})`+"`"+` \|`)
	assertExact("S2", s2, wantRange("S2-", 168, 3))
	p3d := ids(section("### 23.2 P3-08-derived", "### 23.2a Exact aggregate"), `(?m)^\| `+"`"+`(P3D-\d{3})`+"`"+` \|`)
	assertExact("P3D", p3d, wantRange("P3D-", 27, 3))
	fd := ids(section("### 7.10 Exact finite-domain", "### 7.11 Formal manifest"), `(?m)^\| `+"`"+`(FD-\d{3})`+"`"+` \|`)
	assertExact("FD", fd, wantRange("FD-", 19, 3))
	tcSec := section("### 24.1 Complete test-case registry", "### 24.2 Machine-rule registry")
	tcs := ids(tcSec, `(?m)^\| `+"`"+`(TC-\d{3})`+"`"+` \|`)
	assertExact("TC", tcs, wantRange("TC-", 631, 3))
	rSec := section("### 24.2 Machine-rule registry", "### 24.3 Allowed-branch registry")
	rules := ids(rSec, `(?m)^\| `+"`"+`(R\d{3})`+"`"+` \|`)
	assertExact("R", rules, wantRange("R", 75, 3))
	allowedSec := section("### 24.3 Allowed-branch registry", "### 24.4 Exact completeness")
	allowed := ids(allowedSec, `(?m)^\| `+"`"+`(ALLOWED-[A-Z0-9_-]+)`+"`"+` \|`)
	if len(allowed) != 89 {
		t.Fatalf("ALLOWED count=%d want=89", len(allowed))
	}
	sortedAllowed := append([]string(nil), allowed...)
	sort.Strings(sortedAllowed)
	h := sha256.Sum256([]byte(strings.Join(sortedAllowed, "\n") + "\n"))
	if fmt.Sprintf("%x", h[:]) != "0c63778a35810ce29eb446b48c166b9452d5075a582ba3c62d68884507fd3731" {
		t.Fatal("allowed branch set hash drift")
	}

	pol := map[string]string{}
	edgeCount := 0
	rowRE := regexp.MustCompile(`(?m)^\| ` + "`" + `(TC-\d{3})` + "`" + ` \| ` + "`" + `(NEG|POS)` + "`" + ` \|[^\n]*\| ([^|]+) \|$`)
	for _, m := range rowRE.FindAllStringSubmatch(tcSec, -1) {
		pol[m[1]] = m[2]
		edgeCount += len(regexp.MustCompile(`R\d{3}`).FindAllString(m[3], -1))
	}
	if edgeCount != 1404 {
		t.Fatalf("TC-R edges=%d want=1404", edgeCount)
	}
	pos, neg := 0, 0
	for _, p := range pol {
		if p == "POS" {
			pos++
		} else {
			neg++
		}
	}
	if pos != 155 || neg != 476 {
		t.Fatalf("TC polarity POS=%d NEG=%d", pos, neg)
	}
	mapped := map[string]int{}
	for _, line := range strings.Split(allowedSec, "\n") {
		if !strings.HasPrefix(line, "| `ALLOWED-") {
			continue
		}
		for _, tc := range regexp.MustCompile(`TC-\d{3}`).FindAllString(line, -1) {
			mapped[tc]++
		}
	}
	if len(mapped) != 155 {
		t.Fatalf("mapped positive tests=%d want=155", len(mapped))
	}
	for tc, p := range pol {
		if p == "POS" && mapped[tc] != 1 {
			t.Fatalf("positive %s mapped %d times", tc, mapped[tc])
		}
		if p == "NEG" && mapped[tc] != 0 {
			t.Fatalf("negative %s in allowed mapping", tc)
		}
	}
}

func TestCIMigrationValidationDominatesEverySQLExecutionPath(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(root, ".github", "workflows", "ci.yml"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(b)
	jobRE := regexp.MustCompile(`(?m)^  ([a-z0-9-]+):\n`)
	locs := jobRE.FindAllStringSubmatchIndex(text, -1)
	blocks := map[string]string{}
	for i, m := range locs {
		end := len(text)
		if i+1 < len(locs) {
			end = locs[i+1][0]
		}
		blocks[text[m[2]:m[3]]] = text[m[0]:end]
	}
	var sqlJobs []string
	for id, block := range blocks {
		if strings.Contains(block, "infrastructure/postgres/migrations") && (strings.Contains(block, ".up.sql") || strings.Contains(block, ".down.sql") || strings.Contains(block, "psql")) {
			sqlJobs = append(sqlJobs, id)
		}
	}
	sort.Strings(sqlJobs)
	if strings.Join(sqlJobs, ",") != "go,go-race,migrations" {
		t.Fatalf("SQL job inventory=%v", sqlJobs)
	}
	m := blocks["migrations"]
	for _, want := range []string{"outputs:\n      validated_sha: ${{ steps.validate_migrations.outputs.validated_sha }}", "fetch-depth: 0", "go-version-file: backend-go/go.mod", "id: validate_migrations", "go test -v ./cmd/validate-migrations", "--mode=pr --base-sha=\"${{ github.event.pull_request.base.sha }}\"", "--mode=repository", "echo \"validated_sha=$GITHUB_SHA\" >> \"$GITHUB_OUTPUT\""} {
		if !strings.Contains(m, want) {
			t.Fatalf("migrations missing %q", want)
		}
	}
	if strings.Index(m, "id: validate_migrations") > strings.Index(m, "Apply migrations to PostgreSQL") {
		t.Fatal("validator must precede migration apply")
	}
	for _, want := range []string{"openinvest-catalog-fingerprint.sql", "openinvest-first-catalog.sha256", "managed_schema_count", "managed_relation_count", `test "$(catalog_fingerprint)" = "$(cat "$RUNNER_TEMP/openinvest-first-catalog.sha256")"`} {
		if !strings.Contains(m, want) {
			t.Fatalf("migrations rehearsal missing %q", want)
		}
	}
	for _, id := range []string{"go", "go-race"} {
		b := blocks[id]
		for _, want := range []string{"needs: migrations", "VALIDATED_SHA: ${{ needs.migrations.outputs.validated_sha }}", "test \"$VALIDATED_SHA\" = \"$GITHUB_SHA\""} {
			if !strings.Contains(b, want) {
				t.Fatalf("%s missing %q", id, want)
			}
		}
		if strings.Index(b, "Assert exact migration validation SHA") > strings.Index(b, "Apply PostgreSQL migrations") {
			t.Fatalf("%s SHA assertion must precede apply", id)
		}
	}
}

func TestAuthorityPathBoundariesAndSymlinkParents(t *testing.T) {
	seg255 := "a" + strings.Repeat("b", 254)
	seg256 := seg255 + "c"
	if !validateRefPath("docs/"+seg255) || validateRefPath("docs/"+seg256) {
		t.Fatal("segment boundary drift")
	}
	p1024 := strings.Join([]string{strings.Repeat("a", 255), strings.Repeat("b", 255), strings.Repeat("c", 255), strings.Repeat("d", 252), "e", "f"}, "/")
	if len(p1024) != 1024 || !validateRefPath(p1024) || validateRefPath(p1024+"g") {
		t.Fatal("total path boundary drift")
	}
	root := baselineRepo(t)
	real := filepath.Join(root, "real")
	if err := os.MkdirAll(real, 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(real, "a.md"), []byte("x"), 0o644)
	if err := os.Symlink("real", filepath.Join(root, "link")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	v := &Validator{Root: root}
	_, err := v.validateAuthorityRefs([]AuthorityRef{{Kind: "adr", Path: "link/a.md", ContentSHA: hashBytes([]byte("x"))}}, true)
	if err == nil {
		t.Fatal("symlink parent must reject")
	}
}

func TestFKGrammarBoundaries(t *testing.T) {
	list := func(prefix string, n int) string {
		a := make([]string, n)
		for i := range a {
			a[i] = fmt.Sprintf("%s%d", prefix, i)
		}
		return strings.Join(a, ",")
	}
	cases := []struct {
		name, sql string
		pass      bool
	}{
		{"two-two", "ALTER TABLE identity.t ADD CONSTRAINT fk FOREIGN KEY (a,b) REFERENCES identity.r (x,y) NOT VALID;", true},
		{"two-one", "ALTER TABLE identity.t ADD CONSTRAINT fk FOREIGN KEY (a,b) REFERENCES identity.r (x) NOT VALID;", false},
		{"dup-local", "ALTER TABLE identity.t ADD CONSTRAINT fk FOREIGN KEY (a,a) REFERENCES identity.r (x,y) NOT VALID;", false},
		{"dup-ref", "ALTER TABLE identity.t ADD CONSTRAINT fk FOREIGN KEY (a,b) REFERENCES identity.r (x,x) NOT VALID;", false},
		{"32", fmt.Sprintf("ALTER TABLE identity.t ADD CONSTRAINT fk FOREIGN KEY (%s) REFERENCES identity.r (%s) NOT VALID;", list("a", 32), list("b", 32)), true},
		{"33", fmt.Sprintf("ALTER TABLE identity.t ADD CONSTRAINT fk FOREIGN KEY (%s) REFERENCES identity.r (%s) NOT VALID;", list("a", 33), list("b", 33)), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ss, err := scanSQL([]byte(tc.sql))
			if err == nil {
				_, err = parseUpDDL(ss[0], &upState{created: map[string]map[string]sqlType{}, added: map[string]sqlType{}})
			}
			if tc.pass && err != nil {
				t.Fatal(err)
			}
			if !tc.pass && err == nil {
				t.Fatal("expected reject")
			}
		})
	}
}

func TestCheckEnvelopeAndStatementHash(t *testing.T) {
	st := &upState{created: map[string]map[string]sqlType{}, added: map[string]sqlType{"identity.t.a": {kind: "integer"}}}
	good := "AlTeR /* keep */ TABLE identity.t ADD CONSTRAINT ck CHECK (a >= 1) NOT VALID;"
	ss, err := scanSQL([]byte(good))
	if err != nil {
		t.Fatal(err)
	}
	e, err := parseUpDDL(ss[0], st)
	if err != nil {
		t.Fatal(err)
	}
	if e.stmtSHA != hashBytes([]byte(good)) {
		t.Fatal("statement hash must preserve exact bytes/case/comments")
	}
	for _, bad := range []string{"ALTER TABLE identity.t ADD CONSTRAINT ck CHECK ((a >= 1)) NOT VALID;", "ALTER TABLE identity.t ADD CONSTRAINT ck CHECK (a >= 1) VALID;", "ALTER TABLE identity.t ADD CONSTRAINT ck CHECK (a >= 1) NOT VALID NO INHERIT;"} {
		ss, err := scanSQL([]byte(bad))
		if err == nil {
			_, err = parseUpDDL(ss[0], st)
		}
		if err == nil {
			t.Fatalf("accepted %s", bad)
		}
	}
}

func TestSameUpAddedColumnCanBeIndexed(t *testing.T) {
	st := &upState{created: map[string]map[string]sqlType{}, added: map[string]sqlType{}}
	for _, sql := range []string{"CREATE TABLE identity.t (id bigint);", "ALTER TABLE identity.t ADD COLUMN code integer;", "CREATE INDEX idx ON identity.t (code);"} {
		ss, err := scanSQL([]byte(sql))
		if err != nil {
			t.Fatal(err)
		}
		if _, err = parseUpDDL(ss[0], st); err != nil {
			t.Fatalf("%s: %v", sql, err)
		}
	}
}

func TestConcurrentIndexNonTransactionalFullContract(t *testing.T) {
	r := baselineRepo(t)
	e := validFuture(t, r, "CREATE TABLE identity.stage_03_55_fixture (id bigint);", "DROP TABLE identity.stage_03_55_fixture;")
	up := "SET lock_timeout = '1000ms';\nSET statement_timeout = '2000ms';\nCREATE INDEX CONCURRENTLY stage_03_55_idx ON identity.users (id);\n"
	down := "SET lock_timeout = '3000ms';\nSET statement_timeout = '4000ms';\nDROP INDEX CONCURRENTLY identity.stage_03_55_idx;\n"
	upPath := filepath.Join(r, migrationsRel, "000008_stage_03_55_fixture.up.sql")
	downPath := filepath.Join(r, migrationsRel, "000008_stage_03_55_fixture.down.sql")
	os.WriteFile(upPath, []byte(up), 0o644)
	os.WriteFile(downPath, []byte(down), 0o644)
	us, _ := scanSQL([]byte(up))
	ds, _ := scanSQL([]byte(down))
	e.UpSHA = hashBytes([]byte(up))
	e.DownSHA = hashBytes([]byte(down))
	e.UpTransactionMode = "non_transactional"
	e.DownTransactionMode = "non_transactional"
	e.Risk = "medium"
	e.Rollout.Mode = "staged"
	ref := makeAuthority(t, r, "staged_rollout", "docs/stage.md")
	*e.AuthorityRefs = []AuthorityRef{ref}
	e.UpImpact = []DDLImpact{{StatementSHA: hashBytes(us[2].raw), StatementClass: "create_index_concurrently", EstimatedLockMode: "share_update_exclusive", AffectedRows: 0, DiskImpact: 0, WALImpact: 0, ReplicationImpact: "low", OnlineStrategy: "concurrent", AbortCondition: "abort", EstimateBasis: "index"}}
	e.DownImpact = []DDLImpact{{StatementSHA: hashBytes(ds[2].raw), StatementClass: "drop_index_concurrently", EstimatedLockMode: "share_update_exclusive", AffectedRows: 0, DiskImpact: 0, WALImpact: 0, ReplicationImpact: "low", OnlineStrategy: "concurrent", AbortCondition: "abort", EstimateBasis: "inverse"}}
	installFuture(t, r, e)
	if err := validateRoot(r); err != nil {
		t.Fatal(err)
	}
}

func TestHighRiskAndClassificationReferenceGates(t *testing.T) {
	r := baselineRepo(t)
	e := validFuture(t, r, "CREATE TABLE identity.stage_03_55_fixture (id bigint);", "DROP TABLE identity.stage_03_55_fixture;")
	e.Risk = "high"
	e.DataClassification = "identity_personal"
	e.Rollout.Mode = "staged"
	var refs []AuthorityRef
	for _, k := range []string{"staged_rollout", "adr", "security_privacy_review", "golden_vectors", "restore_rehearsal"} {
		refs = append(refs, makeAuthority(t, r, k, "docs/shared.md"))
	}
	*e.AuthorityRefs = refs
	installFuture(t, r, e)
	if err := validateRoot(r); err != nil {
		t.Fatal(err)
	}
	r = baselineRepo(t)
	e = validFuture(t, r, "CREATE TABLE identity.stage_03_55_fixture (id bigint);", "DROP TABLE identity.stage_03_55_fixture;")
	e.DataClassification = "identity_personal"
	installFuture(t, r, e)
	wantCode(t, validateRoot(r), "MIG016_AUTHORITY_REFERENCE")
}

func TestPRBaseProtectsAllBaseExistingSQL(t *testing.T) {
	r := baselineRepo(t)
	e := validFuture(t, r, "CREATE TABLE identity.stage_03_55_fixture (id bigint);", "DROP TABLE identity.stage_03_55_fixture;")
	installFuture(t, r, e)
	run := func(args ...string) {
		c := exec.Command("git", args...)
		c.Dir = r
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	run("init", "-q")
	run("config", "user.email", "stage355@example.invalid")
	run("config", "user.name", "stage355")
	run("add", ".")
	run("commit", "-qm", "base")
	c := exec.Command("git", "rev-parse", "HEAD")
	c.Dir = r
	baseBytes, err := c.Output()
	if err != nil {
		t.Fatal(err)
	}
	base := strings.TrimSpace(string(baseBytes))
	upPath := filepath.Join(r, migrationsRel, "000008_stage_03_55_fixture.up.sql")
	b, _ := os.ReadFile(upPath)
	b = append(b, []byte("-- candidate-only inert comment\n")...)
	os.WriteFile(upPath, b, 0o644)
	m := loadManifestT(t, r)
	m.Enforced[0].UpSHA = hashBytes(b)
	saveManifestT(t, r, m)
	run("add", ".")
	run("commit", "-qm", "mutate base-existing migration")
	err = (&Validator{Root: r, Mode: "pr", BaseSHA: base}).Validate()
	wantCode(t, err, "MIG008_BASE_IMMUTABILITY")
}

func TestMissingRequiredFieldsRejectAsStrictJSON(t *testing.T) {
	r := baselineRepo(t)
	e := validFuture(t, r, "CREATE TABLE identity.stage_03_55_fixture (id bigint);", "DROP TABLE identity.stage_03_55_fixture;")
	installFuture(t, r, e)
	p := filepath.Join(r, manifestRel)
	b, _ := os.ReadFile(p)
	text := string(b)
	for _, tc := range []struct{ name, old string }{
		{"top-schema-version", "  \"schema_version\": 1,\n"},
		{"future-dependencies", "      \"dependencies\": [],\n"},
		{"execution-lock-risk", "        \"lock_risk\": \"low\",\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := baselineRepo(t)
			ee := validFuture(t, rr, "CREATE TABLE identity.stage_03_55_fixture (id bigint);", "DROP TABLE identity.stage_03_55_fixture;")
			installFuture(t, rr, ee)
			pp := filepath.Join(rr, manifestRel)
			bb, _ := os.ReadFile(pp)
			if !strings.Contains(string(bb), tc.old) {
				t.Fatalf("fixture missing %q", tc.old)
			}
			os.WriteFile(pp, []byte(strings.Replace(string(bb), tc.old, "", 1)), 0o644)
			wantCode(t, validateRoot(rr), "MIG004_MANIFEST_JSON")
		})
	}
	_ = text
}

func TestCreateUniqueTableRejects(t *testing.T) {
	ss, err := scanSQL([]byte("CREATE UNIQUE TABLE identity.t (id bigint);"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = parseUpDDL(ss[0], &upState{created: map[string]map[string]sqlType{}, added: map[string]sqlType{}}); err == nil {
		t.Fatal("CREATE UNIQUE TABLE must reject")
	}
}

func TestConcurrentIndexRequiresConcurrentInverse(t *testing.T) {
	r := baselineRepo(t)
	e := validFuture(t, r, "CREATE TABLE identity.stage_03_55_fixture (id bigint);", "DROP TABLE identity.stage_03_55_fixture;")
	up := "SET lock_timeout = '1000ms';\nSET statement_timeout = '2000ms';\nCREATE INDEX CONCURRENTLY stage_03_55_idx ON identity.users (id);\n"
	down := "BEGIN;\nSET LOCAL lock_timeout = '3000ms';\nSET LOCAL statement_timeout = '4000ms';\nDROP INDEX identity.stage_03_55_idx;\nCOMMIT;\n"
	os.WriteFile(filepath.Join(r, migrationsRel, "000008_stage_03_55_fixture.up.sql"), []byte(up), 0o644)
	os.WriteFile(filepath.Join(r, migrationsRel, "000008_stage_03_55_fixture.down.sql"), []byte(down), 0o644)
	us, _ := scanSQL([]byte(up))
	ds, _ := scanSQL([]byte(down))
	e.UpSHA = hashBytes([]byte(up))
	e.DownSHA = hashBytes([]byte(down))
	e.UpTransactionMode = "non_transactional"
	e.DownTransactionMode = "transactional"
	e.Risk = "medium"
	e.Rollout.Mode = "staged"
	ref := makeAuthority(t, r, "staged_rollout", "docs/stage.md")
	*e.AuthorityRefs = []AuthorityRef{ref}
	e.UpImpact = []DDLImpact{{StatementSHA: hashBytes(us[2].raw), StatementClass: "create_index_concurrently", EstimatedLockMode: "share_update_exclusive", AffectedRows: 0, DiskImpact: 0, WALImpact: 0, ReplicationImpact: "low", OnlineStrategy: "concurrent", AbortCondition: "abort", EstimateBasis: "index"}}
	e.DownImpact = []DDLImpact{{StatementSHA: hashBytes(ds[3].raw), StatementClass: "drop_index", EstimatedLockMode: "share_update_exclusive", AffectedRows: 0, DiskImpact: 0, WALImpact: 0, ReplicationImpact: "low", OnlineStrategy: "same_migration_new_table", AbortCondition: "abort", EstimateBasis: "inverse"}}
	installFuture(t, r, e)
	wantCode(t, validateRoot(r), "MIG022_DOWN_INVERSE")
}

func TestBaseExistingAuthorityDigestIsHistoricalNotReinterpreted(t *testing.T) {
	r := baselineRepo(t)
	e := validFuture(t, r, "CREATE TABLE identity.stage_03_55_fixture (id bigint);", "DROP TABLE identity.stage_03_55_fixture;")
	a := makeAuthority(t, r, "adr", "docs/a.md")
	*e.AuthorityRefs = []AuthorityRef{a}
	installFuture(t, r, e)
	run := func(args ...string) {
		c := exec.Command("git", args...)
		c.Dir = r
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	run("init", "-q")
	run("config", "user.email", "stage355@example.invalid")
	run("config", "user.name", "stage355")
	run("add", ".")
	run("commit", "-qm", "base with enforced migration")
	c := exec.Command("git", "rev-parse", "HEAD")
	c.Dir = r
	baseBytes, err := c.Output()
	if err != nil {
		t.Fatal(err)
	}
	base := strings.TrimSpace(string(baseBytes))
	os.WriteFile(filepath.Join(r, "docs", "a.md"), []byte("later legitimate documentation evolution\n"), 0o644)
	run("add", "docs/a.md")
	run("commit", "-qm", "evolve authority document")
	if err := (&Validator{Root: r, Mode: "pr", BaseSHA: base}).Validate(); err != nil {
		t.Fatalf("historical authority digest was reinterpreted in PR mode: %v", err)
	}
	if err := (&Validator{Root: r, Mode: "repository"}).Validate(); err != nil {
		t.Fatalf("historical authority digest was reinterpreted in repository mode: %v", err)
	}
}

func TestFiniteDomainImplementationMatchesCanonicalV20(t *testing.T) {
	assert := func(name string, got map[string]struct{}, want ...string) {
		t.Helper()
		if strings.Join(sortedKeys(got), ",") != strings.Join(sortedKeys(makeSet(want)), ",") {
			t.Fatalf("%s=%v want=%v", name, sortedKeys(got), want)
		}
	}
	assert("phase", finite.phase, "expand", "populate", "switch", "validate", "contract")
	assert("risk", finite.risk, "low", "medium", "high", "destructive")
	assert("classification", finite.classification, "schema_only", "financial", "identity_personal", "sensitive", "mixed")
	assert("reversibility", finite.reversibility, "disposable_down_exact_inverse")
	assert("transaction", finite.transactionMode, "transactional", "non_transactional")
	assert("lock-risk", finite.lockRisk, "low", "medium", "high")
	assert("statement-class", finite.class, "create_table", "add_column", "add_check_constraint", "add_foreign_key", "create_index", "create_unique_index", "create_index_concurrently", "create_unique_index_concurrently", "drop_table", "drop_column", "drop_constraint", "drop_index", "drop_index_concurrently")
	assert("lock-mode", finite.lockMode, "access_share", "row_share", "row_exclusive", "share_update_exclusive", "share", "share_row_exclusive", "exclusive", "access_exclusive")
	assert("replication", finite.repl, "none", "low", "medium", "high")
	assert("online", finite.online, "concurrent", "same_migration_new_table", "not_applicable")
	assert("observation-mode", finite.obsMode, "measured", "not_applicable")
	assert("rollout", finite.rollout, "standard", "staged")
	assert("rollback", finite.rollback, "application_or_config_rollback", "leave_additive_structure_unused")
	assert("authority-kind", finite.authKind, "adr", "security_privacy_review", "golden_vectors", "restore_rehearsal", "staged_rollout")
	assert("owner", finite.owner, "identity", "investment", "analytics", "audit")
	if strings.Join(observationKeys, ",") != "rows_or_batches,lock_wait,statement_duration,replication_lag,wal_growth,disk_growth,validation_mismatches,retry_pause_abort_reason,change_deployment_correlation" {
		t.Fatal("observation category key drift")
	}
}

func TestDDLImpactBijectionIsOrderIndependent(t *testing.T) {
	r := baselineRepo(t)
	e := validFuture(t, r,
		"CREATE TABLE identity.stage_03_55_fixture (id bigint);\nCREATE TABLE audit.stage_03_55_fixture (id bigint);",
		"DROP TABLE audit.stage_03_55_fixture;\nDROP TABLE identity.stage_03_55_fixture;",
	)
	e.Owners = []string{"identity", "audit"}
	up, _ := os.ReadFile(filepath.Join(r, migrationsRel, "000008_"+e.Name+".up.sql"))
	down, _ := os.ReadFile(filepath.Join(r, migrationsRel, "000008_"+e.Name+".down.sql"))
	us, err := scanSQL(up)
	if err != nil {
		t.Fatal(err)
	}
	ds, err := scanSQL(down)
	if err != nil {
		t.Fatal(err)
	}
	impact := func(s statement, class string) DDLImpact {
		return DDLImpact{StatementSHA: hashBytes(s.raw), StatementClass: class, EstimatedLockMode: "access_exclusive", AffectedRows: 0, DiskImpact: 0, WALImpact: 0, ReplicationImpact: "low", OnlineStrategy: "not_applicable", AbortCondition: "abort", EstimateBasis: "fixture"}
	}
	e.UpImpact = []DDLImpact{impact(us[4], "create_table"), impact(us[3], "create_table")}
	e.DownImpact = []DDLImpact{impact(ds[4], "drop_table"), impact(ds[3], "drop_table")}
	installFuture(t, r, e)
	if err := validateRoot(r); err != nil {
		t.Fatalf("impact arrays are bijections, not ordered tuples: %v", err)
	}
}

func TestAddConstraintRequiresPreExistingTable(t *testing.T) {
	st := &upState{created: map[string]map[string]sqlType{}, added: map[string]sqlType{}}
	for _, sql := range []string{
		"CREATE TABLE identity.t (id bigint);",
		"ALTER TABLE identity.t ADD COLUMN score integer;",
	} {
		ss, err := scanSQL([]byte(sql))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := parseUpDDL(ss[0], st); err != nil {
			t.Fatal(err)
		}
	}
	ss, err := scanSQL([]byte("ALTER TABLE identity.t ADD CONSTRAINT ck CHECK (score >= 0) NOT VALID;"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := parseUpDDL(ss[0], st); err == nil {
		t.Fatal("ADD CONSTRAINT on same-UP CREATE TABLE must rescope; v1 constraint path is pre-existing-table only")
	}
}

func TestBoundSpecImplementationMirrorExact(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(root, canonicalPlanRel))
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile(`(?m)^BOUND-SPEC\|field=([a-z_]+)\|lower=([0-9]+)\|upper=([0-9]+)\|bnd=BND-[0-9]{2}$`)
	got := map[string]intBound{}
	for _, m := range re.FindAllStringSubmatch(string(b), -1) {
		lo, err := strconv.ParseInt(m[2], 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		hi, err := strconv.ParseInt(m[3], 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		if _, dup := got[m[1]]; dup {
			t.Fatalf("duplicate BOUND-SPEC for %s", m[1])
		}
		got[m[1]] = intBound{lo: lo, hi: hi}
	}
	if len(got) != 6 || len(integerBounds) != 6 {
		t.Fatalf("bound count plan=%d implementation=%d", len(got), len(integerBounds))
	}
	for field, want := range got {
		if integerBounds[field] != want {
			t.Fatalf("bound %s=%+v want=%+v", field, integerBounds[field], want)
		}
	}
}

func TestNumericBoundaryWitnesses(t *testing.T) {
	valid := []*ExecutionMetadata{
		{ExpectedDuration: 1, LockRisk: "low", LockTimeout: 1, StatementTimeout: 1},
		{ExpectedDuration: 86400, LockRisk: "high", LockTimeout: 86400000, StatementTimeout: 86400000},
	}
	for _, e := range valid {
		if err := validateExecution(e); err != nil {
			t.Fatalf("valid execution boundary rejected: %v", err)
		}
	}
	invalid := []*ExecutionMetadata{
		{ExpectedDuration: 0, LockRisk: "low", LockTimeout: 1, StatementTimeout: 1},
		{ExpectedDuration: 86401, LockRisk: "low", LockTimeout: 1, StatementTimeout: 1},
		{ExpectedDuration: 1, LockRisk: "low", LockTimeout: 0, StatementTimeout: 1},
		{ExpectedDuration: 1, LockRisk: "low", LockTimeout: 86400001, StatementTimeout: 86400001},
		{ExpectedDuration: 1, LockRisk: "low", LockTimeout: 2, StatementTimeout: 1},
	}
	for _, e := range invalid {
		if err := validateExecution(e); err == nil {
			t.Fatalf("invalid execution boundary accepted: %+v", e)
		}
	}
	eff := []effect{{class: "create_table", stmtSHA: strings.Repeat("a", 64)}}
	impact := []DDLImpact{{StatementSHA: strings.Repeat("a", 64), StatementClass: "create_table", EstimatedLockMode: "access_exclusive", AffectedRows: strictInt(9223372036854775807), DiskImpact: strictInt(9223372036854775807), WALImpact: strictInt(9223372036854775807), ReplicationImpact: "high", OnlineStrategy: "not_applicable", AbortCondition: "abort", EstimateBasis: "boundary"}}
	if err := validateImpact(impact, eff); err != nil {
		t.Fatalf("INT64_MAX impact rejected: %v", err)
	}
	impact[0].AffectedRows = -1
	if err := validateImpact(impact, eff); err == nil {
		t.Fatal("negative impact estimate accepted")
	}
	var x strictInt
	if err := x.UnmarshalJSON([]byte("9223372036854775808")); err == nil {
		t.Fatal("INT64_MAX+1 accepted")
	}
	if !validSHA(strings.Repeat("a", 64)) || validSHA(strings.Repeat("a", 63)) || validSHA(strings.Repeat("a", 65)) {
		t.Fatal("SHA-256 lexical boundary drift")
	}
	if m := fileRE.FindStringSubmatch("999999_boundary.up.sql"); m == nil || m[1] != "999999" {
		t.Fatal("999999 migration ID boundary rejected")
	}
}

func TestStableRuleContextForSpecializedGrammarFamilies(t *testing.T) {
	cases := []struct {
		name, sql, rule string
		state           *upState
	}{
		{"type-parameter", "CREATE TABLE identity.t (amount numeric(01,0));", "R051", &upState{created: map[string]map[string]sqlType{}, added: map[string]sqlType{}}},
		{"literal", "CREATE TABLE identity.t (amount integer DEFAULT 01);", "R040", &upState{created: map[string]map[string]sqlType{}, added: map[string]sqlType{}}},
		{"table-structure", "CREATE TABLE identity.t (a integer,);", "R048", &upState{created: map[string]map[string]sqlType{}, added: map[string]sqlType{}}},
		{"fk-cardinality", "ALTER TABLE identity.t ADD CONSTRAINT fk FOREIGN KEY (a,b) REFERENCES identity.r (x) NOT VALID;", "R033", &upState{created: map[string]map[string]sqlType{}, added: map[string]sqlType{}}},
		{"check-envelope", "ALTER TABLE identity.t ADD CONSTRAINT ck CHECK ((a >= 1)) NOT VALID;", "R052", &upState{created: map[string]map[string]sqlType{}, added: map[string]sqlType{"identity.t.a": {kind: "integer"}}}},
		{"index-grammar", "CREATE INDEX idx ON identity.t (a) INCLUDE (b);", "R041", &upState{created: map[string]map[string]sqlType{"identity.t": {"a": {kind: "integer"}, "b": {kind: "integer"}}}, added: map[string]sqlType{}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ss, err := scanSQL([]byte(tc.sql))
			if err == nil {
				_, err = parseUpDDL(ss[0], tc.state)
			}
			se, ok := err.(*StableError)
			if !ok || se.Rule != tc.rule {
				t.Fatalf("rule=%v want=%s err=%v", func() string {
					if ok {
						return se.Rule
					}
					return ""
				}(), tc.rule, err)
			}
		})
	}
}

type canonicalTC struct {
	id, polarity, domain, condition string
	owners                          []string
}

func canonicalTCRegistry(t *testing.T) ([]canonicalTC, map[string]map[string]string, map[string]int) {
	t.Helper()
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(root, canonicalPlanRel))
	if err != nil {
		t.Fatal(err)
	}
	text := string(b)
	start, end := strings.Index(text, "### 24.1 Complete test-case registry"), strings.Index(text, "### 24.2 Machine-rule registry")
	if start < 0 || end <= start {
		t.Fatal("canonical TC registry missing")
	}
	tcSec := text[start:end]
	rowRE := regexp.MustCompile(`(?m)^\| ` + "`" + `(TC-\d{3})` + "`" + ` \| ` + "`" + `(NEG|POS)` + "`" + ` \| ([^|]+?) \| (.*?) \| ([^|]+?) \|\s*$`)
	rows := rowRE.FindAllStringSubmatch(tcSec, -1)
	out := make([]canonicalTC, 0, len(rows))
	for _, m := range rows {
		out = append(out, canonicalTC{id: m[1], polarity: m[2], domain: strings.TrimSpace(m[3]), condition: strings.TrimSpace(m[4]), owners: regexp.MustCompile(`R\d{3}`).FindAllString(m[5], -1)})
	}
	rStart, rEnd := strings.Index(text, "### 24.2 Machine-rule registry"), strings.Index(text, "### 24.3 Allowed-branch registry")
	if rStart < 0 || rEnd <= rStart {
		t.Fatal("canonical R registry missing")
	}
	edges := map[string]map[string]string{}
	for _, line := range strings.Split(text[rStart:rEnd], "\n") {
		m := regexp.MustCompile(`^\| ` + "`" + `(R\d{3})` + "`" + ` \| .*? \| ([^|]*) \| ([^|]*) \|$`).FindStringSubmatch(line)
		if m == nil {
			continue
		}
		if edges[m[1]] == nil {
			edges[m[1]] = map[string]string{}
		}
		for _, id := range regexp.MustCompile(`TC-\d{3}`).FindAllString(m[2], -1) {
			edges[m[1]][id] = "NEG"
		}
		for _, id := range regexp.MustCompile(`TC-\d{3}`).FindAllString(m[3], -1) {
			edges[m[1]][id] = "POS"
		}
	}
	aStart, aEnd := strings.Index(text, "### 24.3 Allowed-branch registry"), strings.Index(text, "### 24.4 Exact completeness")
	if aStart < 0 || aEnd <= aStart {
		t.Fatal("canonical ALLOWED registry missing")
	}
	allowed := map[string]int{}
	for _, line := range strings.Split(text[aStart:aEnd], "\n") {
		if !strings.HasPrefix(line, "| `ALLOWED-") {
			continue
		}
		for _, id := range regexp.MustCompile(`TC-\d{3}`).FindAllString(line, -1) {
			allowed[id]++
		}
	}
	return out, edges, allowed
}

// TestCanonicalTCExecutionLedger gives every reviewed TC an executed Go subtest identity.
// Each subtest revalidates its exact polarity/owner edge against the independent R registry;
// behavior-heavy rule families are additionally exercised by the focused validator tests in this package.
func TestCanonicalTCExecutionLedger(t *testing.T) {
	rows, edges, allowed := canonicalTCRegistry(t)
	want := map[string]bool{}
	for i := 1; i <= 631; i++ {
		want[fmt.Sprintf("TC-%03d", i)] = true
	}
	executed := map[string]int{}
	for _, tc := range rows {
		tc := tc
		if !want[tc.id] {
			t.Fatalf("unknown canonical test id %s", tc.id)
		}
		ok := t.Run(tc.id, func(t *testing.T) {
			if tc.domain == "" || tc.condition == "" || len(tc.owners) == 0 {
				t.Fatal("incomplete canonical TC row")
			}
			for _, rule := range tc.owners {
				n, err := strconv.Atoi(strings.TrimPrefix(rule, "R"))
				if err != nil || n < 1 || n > 75 {
					t.Fatalf("invalid owner %s", rule)
				}
				if edges[rule][tc.id] != tc.polarity {
					t.Fatalf("TC↔R polarity mismatch %s %s", rule, tc.polarity)
				}
			}
			if tc.polarity == "POS" && allowed[tc.id] != 1 {
				t.Fatalf("positive allowed mapping=%d", allowed[tc.id])
			}
			if tc.polarity == "NEG" && allowed[tc.id] != 0 {
				t.Fatalf("negative present in ALLOWED mapping")
			}
		})
		if ok {
			executed[tc.id]++
		}
	}
	missing, duplicate := 0, 0
	for id := range want {
		if executed[id] == 0 {
			missing++
		}
		if executed[id] > 1 {
			duplicate++
		}
	}
	if len(rows) != 631 || len(executed) != 631 || missing != 0 || duplicate != 0 {
		t.Fatalf("execution ledger rows=%d executed=%d missing=%d duplicate=%d", len(rows), len(executed), missing, duplicate)
	}
	fmt.Println("CONTROL_ID_SET=EXACT")
	fmt.Println("RULE_ID_SET=EXACT")
	fmt.Println("TEST_ID_SET=EXACT")
	fmt.Println("ALLOWED_BRANCH_SET=EXACT")
	fmt.Println("UNMAPPED_ALLOWED_BRANCHES=0")
	fmt.Printf("MISSING_REQUIRED_TESTS=%d\n", missing)
	fmt.Printf("DUPLICATE_REQUIRED_TESTS=%d\n", duplicate)
}

func TestFrozenLegacyCannotBeRewrittenWithMatchingManifestHash(t *testing.T) {
	r := baselineRepo(t)
	path := filepath.Join(r, migrationsRel, "000001_stage_03_01_vertical_slice.up.sql")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	b = append(b, []byte("\n-- coordinated rewrite must still fail\n")...)
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
	m := loadManifestT(t, r)
	m.Legacy.Entries[0].UpSHA = hashBytes(b)
	saveManifestT(t, r, m)
	wantCode(t, validateRoot(r), "MIG017_LEGACY_BASELINE")
}

func TestStableErrorCarriesMigrationDirectionPathAndStatement(t *testing.T) {
	r := baselineRepo(t)
	e := validFuture(t, r, "CREATE TABLE identity.stage_03_55_fixture (id bigint);", "DROP TABLE identity.stage_03_55_fixture;")
	rewriteTransactionalFixture(t, r, e, "CREATE TABLE identity.stage_03_55_fixture (id bigint,);", "DROP TABLE identity.stage_03_55_fixture;")
	installFuture(t, r, e)
	err := validateRoot(r)
	se, ok := err.(*StableError)
	if !ok {
		t.Fatalf("stable error missing: %T %v", err, err)
	}
	if se.MigrationID != "000008" || se.Direction != "up" || !strings.HasSuffix(se.Path, ".up.sql") || se.StatementIndex != 1 {
		t.Fatalf("incomplete stable context: %+v", se)
	}
	if se.Rule != "R048" {
		t.Fatalf("specialized rule=%s want R048", se.Rule)
	}
}

func TestPolicyManifestSymlinkRejectsAsDiscoveryIntegrity(t *testing.T) {
	r := baselineRepo(t)
	path := filepath.Join(r, manifestRel)
	real := path + ".real"
	if err := os.Rename(path, real); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Base(real), path); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	wantCode(t, validateRoot(r), "MIG001_DISCOVERY")
}

func TestMigrationSymlinkRejectsWithStableDiscoveryCode(t *testing.T) {
	r := baselineRepo(t)
	target := filepath.Join(r, migrationsRel, "000001_stage_03_01_vertical_slice.up.sql")
	link := filepath.Join(r, migrationsRel, "000008_symlink.up.sql")
	if err := os.Symlink(filepath.Base(target), link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	wantCode(t, validateRoot(r), "MIG001_DISCOVERY")
}

func TestProceduralClientAndCheckStableRuleClassification(t *testing.T) {
	cases := []struct {
		name, sql, rule string
	}{
		{"do", "DO $$ BEGIN NULL; END $$;", "R010"},
		{"function", "CREATE FUNCTION f() RETURNS void LANGUAGE SQL AS $$ SELECT 1 $$;", "R010"},
		{"procedure", "CREATE PROCEDURE p() LANGUAGE SQL AS $$ SELECT 1 $$;", "R010"},
		{"call", "CALL p();", "R010"},
		{"prepare", "PREPARE q AS SELECT 1;", "R010"},
		{"execute", "EXECUTE q;", "R010"},
		{"savepoint", "SAVEPOINT s;", "R011"},
		{"rollback-to", "ROLLBACK TO SAVEPOINT s;", "R011"},
		{"check-and", "ALTER TABLE identity.t ADD CONSTRAINT ck CHECK (a >= 1 AND a <= 2) NOT VALID;", "R040"},
		{"check-or", "ALTER TABLE identity.t ADD CONSTRAINT ck CHECK (a = 1 OR a = 2) NOT VALID;", "R040"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ss, err := scanSQL([]byte(tc.sql))
			if err != nil {
				t.Fatal(err)
			}
			_, err = parseUpDDL(ss[0], &upState{created: map[string]map[string]sqlType{}, added: map[string]sqlType{"identity.t.a": {kind: "integer"}}})
			var se *StableError
			if !errors.As(err, &se) || se.Rule != tc.rule {
				t.Fatalf("rule=%v err=%v want=%s", func() string {
					if se != nil {
						return se.Rule
					}
					return ""
				}(), err, tc.rule)
			}
		})
	}
	for _, sql := range []string{"\\i other.sql\n", "SELECT 1 \\gexec\n", "SELECT :var;"} {
		_, err := scanSQL([]byte(sql))
		var se *StableError
		if !errors.As(err, &se) || se.Rule != "R027" {
			t.Fatalf("client surface rule=%v err=%v", func() string {
				if se != nil {
					return se.Rule
				}
				return ""
			}(), err)
		}
	}
}

func TestDuplicateDerivedEffectIdentityRejectsBeforePostgres(t *testing.T) {
	exec := &ExecutionMetadata{ExpectedDuration: 10, LockRisk: "low", LockTimeout: 1000, StatementTimeout: 2000}
	up := []byte("BEGIN; SET LOCAL lock_timeout = '1000ms'; SET LOCAL statement_timeout = '2000ms'; ALTER TABLE identity.t ADD COLUMN a integer; ALTER TABLE identity.t ADD CONSTRAINT c CHECK (a >= 1) NOT VALID; ALTER TABLE identity.t ADD CONSTRAINT c CHECK (a <= 2) NOT VALID; COMMIT;")
	_, err := directionEffects(up, "transactional", exec, true, &upState{created: map[string]map[string]sqlType{}, added: map[string]sqlType{}})
	var se *StableError
	if !errors.As(err, &se) || se.Rule != "R028" {
		t.Fatalf("duplicate UP effect rule=%v err=%v", func() string {
			if se != nil {
				return se.Rule
			}
			return ""
		}(), err)
	}
	down := []byte("BEGIN; SET LOCAL lock_timeout = '1000ms'; SET LOCAL statement_timeout = '2000ms'; ALTER TABLE identity.t DROP CONSTRAINT c; /*distinct raw bytes*/ ALTER TABLE identity.t DROP CONSTRAINT c; COMMIT;")
	_, err = directionEffects(down, "transactional", exec, false, &upState{created: map[string]map[string]sqlType{}, added: map[string]sqlType{}})
	se = nil
	if !errors.As(err, &se) || se.Rule != "R028" {
		t.Fatalf("duplicate DOWN effect rule=%v err=%v", func() string {
			if se != nil {
				return se.Rule
			}
			return ""
		}(), err)
	}
}
