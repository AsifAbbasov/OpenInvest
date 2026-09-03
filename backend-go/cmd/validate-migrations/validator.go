package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type pair struct {
	id, name, upPath, downPath string
	up, down                   []byte
}
type Validator struct {
	Root, Mode, BaseSHA string
	pairs               []pair
	manifest            *Manifest
	baseManifest        *Manifest
	basePrepared        bool
}

func (v *Validator) Validate() error {
	if err := contractInventory(v.Root); err != nil {
		return err
	}
	if v.Root == "" {
		return verr("MIG001_DISCOVERY", "R001", "repository root missing")
	}
	pairs, err := discoverPairs(v.Root)
	if err != nil {
		return err
	}
	v.pairs = pairs
	mb, fi, err := readRegular(filepath.Join(v.Root, manifestRel))
	if err != nil {
		if fi != nil {
			return verr("MIG001_DISCOVERY", "R001", "policy manifest must be regular non-symlink")
		}
		return verr("MIG004_MANIFEST_JSON", "R003", err.Error())
	}
	m, err := decodeManifest(mb)
	if err != nil {
		return err
	}
	v.manifest = m
	if v.Mode == "pr" {
		if err := v.prepareBaseContext(); err != nil {
			return err
		}
	} else if b, err := v.git("show", "HEAD:"+manifestRel); err == nil {
		if v.baseManifest, err = decodeManifest(b); err != nil {
			return err
		}
	}
	if err := v.validateManifest(); err != nil {
		return err
	}
	switch v.Mode {
	case "local", "repository":
		return nil
	case "pr":
		return v.validateBase()
	default:
		return verr("MIG007_BASE_CONTEXT", "R006", "mode must be local, repository or pr")
	}
}
func discoverPairs(root string) ([]pair, error) {
	dir := filepath.Join(root, migrationsRel)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, verr("MIG001_DISCOVERY", "R001", "missing migrations directory")
	}
	byKey := map[string]*pair{}
	ids, names, fold := map[string]string{}, map[string]string{}, map[string]string{}
	sqlCount := 0
	upCount := 0
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}
		sqlCount++
		if e.Type()&os.ModeSymlink != 0 {
			return nil, verr("MIG001_DISCOVERY", "R001", "migration SQL must be regular non-symlink")
		}
		info, err := e.Info()
		if err != nil || !info.Mode().IsRegular() {
			return nil, verr("MIG001_DISCOVERY", "R001", "migration SQL must be regular non-symlink")
		}
		m := fileRE.FindStringSubmatch(name)
		if m == nil || m[1] == "000000" {
			return nil, verr("MIG002_FILENAME", "R002", "noncanonical migration SQL filename: "+name)
		}
		id, nm, dirn := m[1], m[2], m[3]
		f := strings.ToLower(name)
		if prev, ok := fold[f]; ok && prev != name {
			return nil, verr("MIG002_FILENAME", "R002", "case-insensitive filename collision")
		}
		fold[f] = name
		if prev, ok := ids[id]; ok && prev != nm {
			return nil, verr("MIG002_FILENAME", "R002", "duplicate numeric ID under different names")
		}
		ids[id] = nm
		if prev, ok := names[nm]; ok && prev != id {
			return nil, verr("MIG002_FILENAME", "R002", "duplicate name under different IDs")
		}
		names[nm] = id
		k := id + "_" + nm
		p := byKey[k]
		if p == nil {
			p = &pair{id: id, name: nm}
			byKey[k] = p
		}
		full := filepath.Join(dir, name)
		b, _, err := readRegular(full)
		if err != nil {
			return nil, verr("MIG001_DISCOVERY", "R001", err.Error())
		}
		if dirn == "up" {
			if p.upPath != "" {
				return nil, verr("MIG003_PAIRING", "R002", "duplicate up")
			}
			p.upPath = full
			p.up = b
			upCount++
		} else {
			if p.downPath != "" {
				return nil, verr("MIG003_PAIRING", "R002", "duplicate down")
			}
			p.downPath = full
			p.down = b
		}
	}
	if sqlCount == 0 || upCount == 0 {
		return nil, verr("MIG001_DISCOVERY", "R001", "no up migrations found")
	}
	out := make([]pair, 0, len(byKey))
	for _, p := range byKey {
		if p.upPath == "" || p.downPath == "" {
			return nil, verr("MIG003_PAIRING", "R002", "orphan migration pair")
		}
		out = append(out, *p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].id < out[j].id })
	return out, nil
}
func readRegular(path string) ([]byte, os.FileInfo, error) {
	fi, err := os.Lstat(path)
	if err != nil {
		return nil, nil, err
	}
	if !fi.Mode().IsRegular() || fi.Mode()&os.ModeSymlink != 0 {
		return nil, fi, fmt.Errorf("not a regular non-symlink file: %s", path)
	}
	b, err := os.ReadFile(path)
	return b, fi, err
}
func (v *Validator) validateManifest() error {
	m := v.manifest
	if int64(m.SchemaVersion) != 1 || m.Legacy == nil || m.Legacy.MaxID != "000007" || m.Legacy.PolicyMetadataStatus != "not_retroactively_asserted" {
		return verr("MIG017_LEGACY_BASELINE", "R004", "legacy header mismatch")
	}
	if len(m.Legacy.Entries) != 7 {
		return verr("MIG017_LEGACY_BASELINE", "R004", "legacy baseline must contain exactly seven entries")
	}
	pm := map[string]pair{}
	for _, p := range v.pairs {
		pm[p.id] = p
	}
	seen := map[string]bool{}
	last := ""
	for i, e := range m.Legacy.Entries {
		want := legacyExpected[i]
		p, ok := pm[want.ID]
		if !ok || e != want || e.Name != p.name || e.UpSHA != hashBytes(p.up) || e.DownSHA != hashBytes(p.down) || seen[e.ID] || e.ID <= last {
			return verr("MIG017_LEGACY_BASELINE", "R004", "legacy entry identity/hash/order mismatch")
		}
		seen[e.ID] = true
		last = e.ID
	}
	entries := map[string]*EnforcedMigration{}
	last = "000007"
	for i := range m.Enforced {
		e := &m.Enforced[i]
		if !idRE.MatchString(e.ID) || e.ID == "000000" || e.ID <= last {
			return verr("MIG005_MANIFEST_BIJECTION", "R005", "enforced IDs must be unique and strictly increasing above baseline")
		}
		p, ok := pm[e.ID]
		if !ok || p.name != e.Name || e.UpSHA != hashBytes(p.up) || e.DownSHA != hashBytes(p.down) || !validSHA(e.UpSHA) || !validSHA(e.DownSHA) {
			return verr("MIG005_MANIFEST_BIJECTION", "R005", "enforced manifest/file identity/hash mismatch")
		}
		entries[e.ID] = e
		last = e.ID
	}
	for _, p := range v.pairs {
		if p.id > "000007" {
			if entries[p.id] == nil {
				return verr("MIG005_MANIFEST_BIJECTION", "R005", "future migration missing enforced manifest entry")
			}
		}
	}
	allIDs := map[string]bool{}
	for _, p := range v.pairs {
		allIDs[p.id] = true
	}
	for i := range m.Enforced {
		if err := v.validateEnforced(&m.Enforced[i], pm[m.Enforced[i].ID], allIDs); err != nil {
			return err
		}
	}
	return nil
}
func (v *Validator) validateEnforced(e *EnforcedMigration, p pair, allIDs map[string]bool) (err error) {
	defer func() {
		if se, ok := err.(*StableError); ok && se.MigrationID == "" {
			se.MigrationID = e.ID
		}
	}()
	if e.Phase != "expand" {
		if !in(finite.phase, e.Phase) {
			return verr("MIG009_METADATA", "R007", "unknown phase")
		}
		return verr("MIG013_SQL_SAFETY", "R007", "paired-SQL v1 supports only expand")
	}
	if !in(finite.risk, e.Risk) || e.Risk == "destructive" {
		return verr("MIG009_METADATA", "R016", "invalid/unsupported risk")
	}
	if !in(finite.classification, e.DataClassification) {
		return verr("MIG024_OWNER_CLASSIFICATION", "R017", "invalid data classification")
	}
	if !in(finite.reversibility, e.Reversibility) {
		return verr("MIG022_DOWN_INVERSE", "R009", "invalid reversibility")
	}
	if err := validateExecution(e.UpExecution); err != nil {
		return err
	}
	if err := validateExecution(e.DownExecution); err != nil {
		return err
	}
	deps := map[string]bool{}
	if e.Dependencies == nil {
		return verr("MIG004_MANIFEST_JSON", "R003", "dependencies field missing")
	}
	for _, d := range *e.Dependencies {
		if !idRE.MatchString(d) || !allIDs[d] || d >= e.ID || d == e.ID || deps[d] {
			return verr("MIG010_DEPENDENCY", "R018", "invalid dependency graph")
		}
		deps[d] = true
	}
	upSt := &upState{created: map[string]map[string]sqlType{}, added: map[string]sqlType{}}
	ups, err := directionEffects(p.up, e.UpTransactionMode, e.UpExecution, true, upSt)
	if err != nil {
		return contextual(err, "up", relPath(v.Root, p.upPath))
	}
	downs, err := directionEffects(p.down, e.DownTransactionMode, e.DownExecution, false, upSt)
	if err != nil {
		return contextual(err, "down", relPath(v.Root, p.downPath))
	}
	if len(ups) != len(downs) {
		return verr("MIG022_DOWN_INVERSE", "R028", "UP/DOWN effect count mismatch")
	}
	for i := range ups {
		if !inverseMatches(ups[i], downs[len(downs)-1-i]) {
			return verr("MIG022_DOWN_INVERSE", "R028", "DOWN is not exact reverse-order inverse")
		}
	}
	if err := validateImpact(e.UpImpact, ups); err != nil {
		return contextual(err, "up", relPath(v.Root, p.upPath))
	}
	if err := validateImpact(e.DownImpact, downs); err != nil {
		return contextual(err, "down", relPath(v.Root, p.downPath))
	}
	min := 1
	owners := map[string]struct{}{}
	for _, x := range ups {
		if x.minRisk > min {
			min = x.minRisk
		}
		owners[x.schema] = struct{}{}
	}
	if riskRank(e.Risk) < min {
		return verr("MIG009_METADATA", "R034", "declared risk below operation-derived minimum")
	}
	if !sameStringSet(e.Owners, owners, finite.owner) {
		return verr("MIG024_OWNER_CLASSIFICATION", "R019", "owners must equal touched schema set")
	}
	measured, err := validateObservation(e.Observability)
	if err != nil {
		return err
	}
	if e.Monitoring == nil || !validRefList(e.Monitoring.Signals, measured) || !nonEmptyOpen(e.Monitoring.SuccessCondition) || !nonEmptyOpen(e.Monitoring.AbortCondition) {
		return verr("MIG019_OBSERVABILITY", "R044", "invalid monitoring declaration")
	}
	if e.Rollout == nil || !validRefList(e.Rollout.Metrics, measured) {
		return verr("MIG023_ROLLOUT", "R025", "invalid rollout metrics")
	}
	if e.AuthorityRefs == nil {
		return verr("MIG004_MANIFEST_JSON", "R003", "authority_refs field missing")
	}
	verifyAuthorityTargets := true
	if v.baseManifest != nil {
		for i := range v.baseManifest.Enforced {
			if v.baseManifest.Enforced[i].ID == e.ID {
				verifyAuthorityTargets = false
				break
			}
		}
	}
	refs, err := v.validateAuthorityRefs(*e.AuthorityRefs, verifyAuthorityTargets)
	if err != nil {
		return err
	}
	staged := refs["staged_rollout"]
	if e.Rollout == nil || !in(finite.rollout, e.Rollout.Mode) {
		return verr("MIG023_ROLLOUT", "R025", "invalid rollout mode")
	}
	if e.Risk == "low" {
		if e.Rollout.Mode != "standard" || staged != 0 {
			return verr("MIG023_ROLLOUT", "R025", "low risk requires standard rollout and zero staged refs")
		}
	} else if e.Rollout.Mode != "staged" || staged != 1 {
		return verr("MIG023_ROLLOUT", "R025", "medium/high requires staged rollout and exactly one staged ref")
	}
	if e.Risk == "high" {
		for _, k := range []string{"adr", "security_privacy_review", "golden_vectors", "restore_rehearsal"} {
			if refs[k] < 1 {
				return verr("MIG016_AUTHORITY_REFERENCE", "R016", "high risk authority reference missing")
			}
		}
	}
	if e.DataClassification == "identity_personal" || e.DataClassification == "sensitive" || e.DataClassification == "mixed" {
		if refs["security_privacy_review"] < 1 {
			return verr("MIG016_AUTHORITY_REFERENCE", "R017", "security/privacy reference required")
		}
	}
	if e.ProductionRollback == nil || !in(finite.rollback, e.ProductionRollback.Strategy) || !nonEmptyOpen(e.ProductionRollback.Procedure) || !nonEmptyOpen(e.ProductionRollback.Verification) || e.RollForward == nil || !nonEmptyOpen(e.RollForward.Procedure) || !nonEmptyOpen(e.RollForward.Verification) {
		return verr("MIG009_METADATA", "R026", "invalid rollback/roll-forward declaration")
	}
	return nil
}
func inverseMatches(up, down effect) bool {
	if up.key != down.key {
		return false
	}
	switch up.class {
	case "create_table":
		return down.class == "drop_table"
	case "add_column":
		return down.class == "drop_column"
	case "add_check_constraint", "add_foreign_key":
		return down.class == "drop_constraint"
	case "create_index", "create_unique_index":
		return down.class == "drop_index" && !down.concurrent
	case "create_index_concurrently", "create_unique_index_concurrently":
		return down.class == "drop_index_concurrently" && down.concurrent
	}
	return false
}
func directionEffects(b []byte, mode string, execMeta *ExecutionMetadata, up bool, st *upState) ([]effect, error) {
	if !in(finite.transactionMode, mode) {
		return nil, verr("MIG014_TRANSACTION", "R011", "invalid transaction mode")
	}
	ss, err := scanSQL(b)
	if err != nil {
		return nil, err
	}
	if len(ss) < 3 {
		return nil, verr("MIG014_TRANSACTION", "R011", "insufficient framed statements")
	}
	start, end := 0, len(ss)
	if mode == "transactional" {
		if !isSimple(ss[0], "BEGIN") || !isSimple(ss[len(ss)-1], "COMMIT") {
			return nil, verr("MIG014_TRANSACTION", "R011", "transactional direction requires BEGIN/COMMIT")
		}
		start = 1
		end--
	}
	if end-start < 3 {
		return nil, verr("MIG015_TIMEOUT", "R012", "timeouts must precede DDL")
	}
	if !isTimeout(ss[start], mode == "transactional", "lock_timeout", int64(execMeta.LockTimeout)) || !isTimeout(ss[start+1], mode == "transactional", "statement_timeout", int64(execMeta.StatementTimeout)) {
		return nil, verr("MIG015_TIMEOUT", "R012", "exact timeout controls missing/mismatched")
	}
	start += 2
	var out []effect
	seenEffects := map[string]bool{}
	for i := start; i < end; i++ {
		var e effect
		if up {
			e, err = parseUpDDL(ss[i], st)
		} else {
			e, err = parseDownDDL(ss[i])
		}
		if err != nil {
			if se, ok := err.(*StableError); ok && se.StatementIndex == 0 {
				se.StatementIndex = i - start + 1
			}
			return nil, err
		}
		if mode == "transactional" && e.concurrent || mode == "non_transactional" && !e.concurrent {
			return nil, verr("MIG014_TRANSACTION", "R032", "mixed/unsupported transaction class")
		}
		if seenEffects[e.key] {
			return nil, verr("MIG022_DOWN_INVERSE", "R028", "duplicate derived effect identity")
		}
		seenEffects[e.key] = true
		out = append(out, e)
	}
	if len(out) == 0 {
		return nil, verr("MIG012_STATEMENT_CLASS", "R008", "no governed DDL")
	}
	return out, nil
}
func isSimple(s statement, k string) bool { p := &parser{t: s.tokens}; return p.take(k) && p.end() }
func isTimeout(s statement, local bool, guc string, n int64) bool {
	p := &parser{t: s.tokens}
	if !p.take("SET") || local && !p.take("LOCAL") || !local && p.peek("LOCAL") || !p.take(guc) || !p.exact("=") || p.i >= len(p.t) {
		return false
	}
	want := fmt.Sprintf("'%dms'", n)
	return p.t[p.i].text == want && func() bool { p.i++; return p.end() }()
}
func validateImpact(got []DDLImpact, eff []effect) error {
	if got == nil || len(got) != len(eff) {
		return verr("MIG018_DDL_IMPACT", "R013", "DDL impact must biject statements")
	}
	bySHA := map[string]effect{}
	for _, e := range eff {
		bySHA[e.stmtSHA] = e
	}
	seen := map[string]bool{}
	for _, x := range got {
		e, ok := bySHA[x.StatementSHA]
		if !ok || seen[x.StatementSHA] || !validSHA(x.StatementSHA) || x.StatementClass != e.class || !in(finite.class, x.StatementClass) || !in(finite.lockMode, x.EstimatedLockMode) || !in(finite.repl, x.ReplicationImpact) || !inBound("affected_rows_estimate", int64(x.AffectedRows)) || !inBound("disk_impact_bytes_estimate", int64(x.DiskImpact)) || !inBound("wal_impact_bytes_estimate", int64(x.WALImpact)) || !nonEmptyOpen(x.AbortCondition) || !nonEmptyOpen(x.EstimateBasis) {
			return verr("MIG018_DDL_IMPACT", "R013", "invalid statement-bound impact")
		}
		seen[x.StatementSHA] = true
		want := "not_applicable"
		if strings.Contains(e.class, "index") {
			if e.concurrent {
				want = "concurrent"
			} else {
				want = "same_migration_new_table"
			}
		}
		if x.OnlineStrategy != want {
			return verr("MIG018_DDL_IMPACT", "R029", "online_strategy mismatch")
		}
	}
	return nil
}
func validRefList(xs []string, measured map[string]bool) bool {
	if len(xs) == 0 {
		return false
	}
	seen := map[string]bool{}
	for _, x := range xs {
		if seen[x] || !measured[x] {
			return false
		}
		seen[x] = true
	}
	return true
}
func sameStringSet(xs []string, want, mapAllowed map[string]struct{}) bool {
	if len(xs) == 0 {
		return false
	}
	got := map[string]struct{}{}
	for _, x := range xs {
		if !in(mapAllowed, x) {
			return false
		}
		if _, dup := got[x]; dup {
			return false
		}
		got[x] = struct{}{}
	}
	if len(got) != len(want) {
		return false
	}
	for k := range want {
		if _, ok := got[k]; !ok {
			return false
		}
	}
	return true
}
func (v *Validator) validateAuthorityRefs(xs []AuthorityRef, verifyTargets bool) (map[string]int, error) {
	counts := map[string]int{}
	seen := map[string]bool{}
	for _, r := range xs {
		if !in(finite.authKind, r.Kind) || !validateRefPath(r.Path) || !validSHA(r.ContentSHA) {
			return nil, verr("MIG016_AUTHORITY_REFERENCE", "R046", "invalid authority reference")
		}
		key := r.Kind + "\x00" + r.Path
		if seen[key] {
			return nil, verr("MIG016_AUTHORITY_REFERENCE", "R046", "duplicate (kind,path) authority reference")
		}
		seen[key] = true
		if verifyTargets {
			b, err := readAuthorityRegular(v.Root, r.Path)
			if err != nil || hashBytes(b) != r.ContentSHA {
				return nil, verr("MIG016_AUTHORITY_REFERENCE", "R015", "authority target/hash mismatch")
			}
		}
		counts[r.Kind]++
	}
	return counts, nil
}
func readAuthorityRegular(root, rel string) ([]byte, error) {
	cur := root
	parts := strings.Split(rel, "/")
	for i, part := range parts {
		cur = filepath.Join(cur, part)
		fi, err := os.Lstat(cur)
		if err != nil || fi.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("authority path component is missing or symlink")
		}
		if i < len(parts)-1 && !fi.IsDir() {
			return nil, fmt.Errorf("authority parent is not a directory")
		}
		if i == len(parts)-1 && !fi.Mode().IsRegular() {
			return nil, fmt.Errorf("authority target is not regular")
		}
	}
	return os.ReadFile(cur)
}
func (v *Validator) git(args ...string) ([]byte, error) {
	c := exec.Command("git", args...)
	c.Dir = v.Root
	return c.Output()
}
func (v *Validator) prepareBaseContext() error {
	if v.basePrepared {
		return nil
	}
	if !regexp40(v.BaseSHA) {
		return verr("MIG007_BASE_CONTEXT", "R006", "PR mode requires full 40-hex base SHA")
	}
	if _, err := v.git("cat-file", "-e", v.BaseSHA+"^{commit}"); err != nil {
		return verr("MIG007_BASE_CONTEXT", "R006", "base commit unavailable")
	}
	if err := exec.Command("git", "-C", v.Root, "merge-base", "--is-ancestor", v.BaseSHA, "HEAD").Run(); err != nil {
		return verr("MIG007_BASE_CONTEXT", "R006", "base is not ancestor of candidate")
	}
	if b, err := v.git("show", v.BaseSHA+":"+manifestRel); err == nil {
		m, decErr := decodeManifest(b)
		if decErr != nil {
			return decErr
		}
		v.baseManifest = m
	}
	v.basePrepared = true
	return nil
}
func (v *Validator) validateBase() error {
	if err := v.prepareBaseContext(); err != nil {
		return err
	}
	basePaths, err := v.git("ls-tree", "-r", "--name-only", v.BaseSHA, "--", migrationsRel)
	if err != nil {
		return verr("MIG007_BASE_CONTEXT", "R006", "cannot enumerate base migration tree")
	}
	candidate := map[string][]byte{}
	for _, p := range v.pairs {
		for _, f := range []struct {
			path string
			b    []byte
		}{{p.upPath, p.up}, {p.downPath, p.down}} {
			rel, _ := filepath.Rel(v.Root, f.path)
			candidate[filepath.ToSlash(rel)] = f.b
		}
	}
	for _, line := range bytes.Split(bytes.TrimSpace(basePaths), []byte("\n")) {
		rel := string(line)
		if filepath.Dir(rel) != migrationsRel || !strings.HasSuffix(rel, ".sql") {
			continue
		}
		cur, ok := candidate[rel]
		old, showErr := v.git("show", v.BaseSHA+":"+rel)
		if !ok || showErr != nil || !bytes.Equal(old, cur) {
			return verr("MIG008_BASE_IMMUTABILITY", "R006", "base-existing migration changed/deleted/renamed")
		}
	}
	if v.baseManifest != nil && !appendOnlyManifest(v.baseManifest, v.manifest) {
		return verr("MIG008_BASE_IMMUTABILITY", "R006", "manifest existing entries are not append-only immutable")
	}
	return nil
}
func appendOnlyManifest(base, cand *Manifest) bool {
	if base == nil || cand == nil || len(base.Enforced) > len(cand.Enforced) {
		return false
	}
	a, _ := json.Marshal(base.Legacy)
	b, _ := json.Marshal(cand.Legacy)
	if !bytes.Equal(a, b) {
		return false
	}
	for i := range base.Enforced {
		x, _ := json.Marshal(base.Enforced[i])
		y, _ := json.Marshal(cand.Enforced[i])
		if !bytes.Equal(x, y) {
			return false
		}
	}
	return true
}
func contextual(err error, direction, path string) error {
	if se, ok := err.(*StableError); ok {
		if se.Direction == "" {
			se.Direction = direction
		}
		if se.Path == "" {
			se.Path = path
		}
	}
	return err
}
func relPath(root, path string) string { r, _ := filepath.Rel(root, path); return filepath.ToSlash(r) }
func regexp40(s string) bool {
	if len(s) != 40 {
		return false
	}
	for _, r := range s {
		if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'f') {
			return false
		}
	}
	return true
}
