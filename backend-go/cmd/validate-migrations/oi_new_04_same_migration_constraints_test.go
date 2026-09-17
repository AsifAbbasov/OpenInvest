package main

import "testing"

func TestOINew04ValidatorAllowsSafeConstraintsOnSameMigrationTables(t *testing.T) {
	up := []byte(`
CREATE TABLE analytics.oi_new_04_parent (
    id UUID NOT NULL,
    kind TEXT NOT NULL,
    generation BIGINT NOT NULL,
    digest CHAR(64) NOT NULL
);
CREATE UNIQUE INDEX oi_new_04_parent_uidx ON analytics.oi_new_04_parent (id);
ALTER TABLE analytics.oi_new_04_parent ADD CONSTRAINT oi_new_04_parent_generation_check CHECK (generation > 0);
ALTER TABLE analytics.oi_new_04_parent ADD CONSTRAINT oi_new_04_parent_kind_check CHECK (kind IN ('A', 'B'));
CREATE TABLE analytics.oi_new_04_child (
    id UUID NOT NULL,
    parent_id UUID NOT NULL
);
ALTER TABLE analytics.oi_new_04_child ADD CONSTRAINT oi_new_04_child_fk FOREIGN KEY (parent_id) REFERENCES analytics.oi_new_04_parent (id) ON DELETE CASCADE;
CREATE INDEX oi_new_04_child_idx ON analytics.oi_new_04_child (parent_id, id DESC);
`)
	statements, err := scanSQL(up)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	state := &upState{created: map[string]map[string]sqlType{}, added: map[string]sqlType{}}
	for index, statement := range statements {
		if _, err := parseUpDDL(statement, state); err != nil {
			t.Fatalf("statement %d rejected: %v", index+1, err)
		}
	}
}

func TestOINew04ValidatorStillRequiresNotValidForPreExistingFK(t *testing.T) {
	statements, err := scanSQL([]byte(`ALTER TABLE investment.portfolios ADD CONSTRAINT forbidden_fk FOREIGN KEY (subject_id) REFERENCES investment.subjects (id);`))
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	state := &upState{created: map[string]map[string]sqlType{}, added: map[string]sqlType{}}
	if _, err := parseUpDDL(statements[0], state); err == nil {
		t.Fatal("expected validated FK on pre-existing table to remain rejected")
	}
}
