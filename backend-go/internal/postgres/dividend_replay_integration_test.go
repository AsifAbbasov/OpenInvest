package postgres

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/openinvest/openinvest/backend-go/internal/decimal"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestDividendReplayLookupPrecedesFreshAdmissionAndBusinessMutation(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}

	store, err := Open(databaseURL)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	key := "stage-03-68-" + uuid.NewString()
	path := "/api/v1/dividends/calculate"
	subjectID := uuid.NewString()
	cleanupDividendReplay(t, store.db, subjectID, path, key)
	defer cleanupDividendReplay(t, store.db, subjectID, path, key)

	cost := verticalslice.Money{Amount: decimal.Must("280000.00000000"), Currency: verticalslice.RUB}
	request := verticalslice.DividendCalculationRequest{
		Ticker:          "SBER",
		Quantity:        decimal.Must("1000.00000000"),
		DividendPerUnit: verticalslice.Money{Amount: decimal.Must("34.84000000"), Currency: verticalslice.RUB},
		PositionCost:    &cost,
	}
	firstRequestContext := verticalslice.RequestContext{
		RequestID: uuid.NewString(),
		TraceID:   "0123456789abcdef0123456789abcdef",
	}
	artifact := verticalslice.CommandReplayArtifact{
		StatusCode: 200,
		Body:       []byte(`{"data":{"ticker":"SBER"}}`),
		RequestID:  firstRequestContext.RequestID,
		TraceID:    firstRequestContext.TraceID,
	}

	admissionCalls := 0
	admit := func() error {
		admissionCalls++
		return nil
	}
	_, first, err := service.CalculateDividendWithReplay(
		ctx,
		firstRequestContext,
		subjectID,
		key,
		path,
		request,
		admit,
		func(verticalslice.DividendCalculation) (verticalslice.CommandReplayArtifact, error) {
			return artifact, nil
		},
	)
	if err != nil {
		t.Fatalf("first CalculateDividendWithReplay: %v", err)
	}
	if string(first.Body) != string(artifact.Body) {
		t.Fatalf("first artifact body = %q", first.Body)
	}
	if admissionCalls != 1 {
		t.Fatalf("fresh command admission calls = %d, want 1", admissionCalls)
	}

	builderCalled := false
	_, replay, err := service.CalculateDividendWithReplay(
		ctx,
		verticalslice.RequestContext{RequestID: uuid.NewString(), TraceID: "11111111111111111111111111111111"},
		subjectID,
		key,
		path,
		request,
		admit,
		func(verticalslice.DividendCalculation) (verticalslice.CommandReplayArtifact, error) {
			builderCalled = true
			return verticalslice.CommandReplayArtifact{}, nil
		},
	)
	if err != nil {
		t.Fatalf("replay CalculateDividendWithReplay: %v", err)
	}
	if builderCalled {
		t.Fatal("replay must not rebuild an already-completed response")
	}
	if admissionCalls != 1 {
		t.Fatalf("completed replay must bypass fresh-command admission: calls = %d, want 1", admissionCalls)
	}
	if replay.StatusCode != artifact.StatusCode || string(replay.Body) != string(artifact.Body) || replay.RequestID != artifact.RequestID || replay.TraceID != artifact.TraceID {
		t.Fatalf("replay artifact differs from original: %#v", replay)
	}

	assertDividendReplayCount(t, store.db, `
		SELECT count(*)
		FROM investment.command_deduplication
		WHERE principal_id = $1 AND method = 'POST' AND canonical_path = $2 AND idempotency_key = $3
	`, 1, subjectID, path, key)
	assertDividendReplayCount(t, store.db, `SELECT count(*) FROM investment.subjects WHERE id = $1`, 0, subjectID)

	conflicting := request
	conflicting.DividendPerUnit = verticalslice.Money{Amount: decimal.Must("35.00000000"), Currency: verticalslice.RUB}
	conflictBuilderCalled := false
	_, _, err = service.CalculateDividendWithReplay(
		ctx,
		verticalslice.RequestContext{RequestID: uuid.NewString(), TraceID: "22222222222222222222222222222222"},
		subjectID,
		key,
		path,
		conflicting,
		admit,
		func(verticalslice.DividendCalculation) (verticalslice.CommandReplayArtifact, error) {
			conflictBuilderCalled = true
			return verticalslice.CommandReplayArtifact{}, nil
		},
	)
	if !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("same key with different request hash error = %v, want ErrIdempotencyConflict", err)
	}
	if conflictBuilderCalled {
		t.Fatal("idempotency conflict must be detected before rebuilding a response")
	}
	if admissionCalls != 1 {
		t.Fatalf("idempotency conflict must be decided before fresh-command admission: calls = %d, want 1", admissionCalls)
	}

	deniedSubjectID := uuid.NewString()
	deniedKey := "stage-03-68-denied-" + uuid.NewString()
	cleanupDividendReplay(t, store.db, deniedSubjectID, path, deniedKey)
	defer cleanupDividendReplay(t, store.db, deniedSubjectID, path, deniedKey)
	denied := errors.New("test dividend admission denied")
	deniedBuilderCalled := false
	_, _, err = service.CalculateDividendWithReplay(
		ctx,
		verticalslice.RequestContext{RequestID: uuid.NewString(), TraceID: "33333333333333333333333333333333"},
		deniedSubjectID,
		deniedKey,
		path,
		request,
		func() error { return denied },
		func(verticalslice.DividendCalculation) (verticalslice.CommandReplayArtifact, error) {
			deniedBuilderCalled = true
			return verticalslice.CommandReplayArtifact{}, nil
		},
	)
	if !errors.Is(err, denied) {
		t.Fatalf("denied fresh command error = %v, want admission denial", err)
	}
	if deniedBuilderCalled {
		t.Fatal("denied fresh command must not build a response")
	}
	assertDividendReplayCount(t, store.db, `
		SELECT count(*)
		FROM investment.command_deduplication
		WHERE principal_id = $1 AND method = 'POST' AND canonical_path = $2 AND idempotency_key = $3
	`, 0, deniedSubjectID, path, deniedKey)
}

func cleanupDividendReplay(t *testing.T, db *sql.DB, subjectID string, path string, key string) {
	t.Helper()
	if _, err := db.Exec(`
		DELETE FROM investment.command_deduplication
		WHERE principal_id = $1 AND method = 'POST' AND canonical_path = $2 AND idempotency_key = $3
	`, subjectID, path, key); err != nil {
		t.Fatalf("cleanup dividend replay: %v", err)
	}
}

func assertDividendReplayCount(t *testing.T, db *sql.DB, query string, want int, args ...any) {
	t.Helper()
	var got int
	if err := db.QueryRow(query, args...).Scan(&got); err != nil {
		t.Fatalf("count query: %v", err)
	}
	if got != want {
		t.Fatalf("count = %d, want %d", got, want)
	}
}
