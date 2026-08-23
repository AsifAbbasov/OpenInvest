package postgres

import (
	"bytes"
	"context"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestStage0332ReplaySurvivesStoreRestart(t *testing.T) {
	databaseURL := os.Getenv("OPENINVEST_DATABASE_TEST_URL")
	if databaseURL == "" {
		t.Skip("OPENINVEST_DATABASE_TEST_URL is not set")
	}
	ctx := context.Background()
	subjectID := uuid.NewString()
	key := "stage-03-32-restart-key-000001"
	path := "/api/v1/portfolios"
	request := verticalslice.CreatePortfolioRequest{Name: "Restart replay", BaseCurrency: verticalslice.RUB}

	cleanupStore := openStage0332Store(t)
	cleanupStage0332Subject(t, cleanupStore, subjectID)

	firstStore, err := Open(databaseURL)
	if err != nil {
		t.Fatalf("open first store: %v", err)
	}
	firstService := verticalslice.NewService(firstStore, verticalslice.SystemClock{})
	firstContext := verticalslice.RequestContext{
		RequestID: uuid.NewString(),
		TraceID:   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	_, firstArtifact, err := firstService.CreatePortfolioWithReplay(
		ctx,
		firstContext,
		subjectID,
		key,
		path,
		request,
		stage0332ArtifactBuilder(firstContext, []byte(`{"restart":"original"}`)),
	)
	if err != nil {
		_ = firstStore.Close()
		t.Fatalf("first create: %v", err)
	}
	if err := firstStore.Close(); err != nil {
		t.Fatalf("close first store: %v", err)
	}

	secondStore, err := Open(databaseURL)
	if err != nil {
		t.Fatalf("open second store: %v", err)
	}
	defer secondStore.Close()
	secondService := verticalslice.NewService(secondStore, verticalslice.SystemClock{})
	builderCalls := 0
	_, replayArtifact, err := secondService.CreatePortfolioWithReplay(
		ctx,
		verticalslice.RequestContext{RequestID: uuid.NewString(), TraceID: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"},
		subjectID,
		key,
		path,
		request,
		func(verticalslice.Portfolio) (verticalslice.CommandReplayArtifact, error) {
			builderCalls++
			return verticalslice.CommandReplayArtifact{}, errors.New("restart replay must not rebuild response")
		},
	)
	if err != nil {
		t.Fatalf("replay after store restart: %v", err)
	}
	if builderCalls != 0 {
		t.Fatalf("replay after restart rebuilt response %d times", builderCalls)
	}
	if replayArtifact.StatusCode != firstArtifact.StatusCode ||
		replayArtifact.RequestID != firstArtifact.RequestID ||
		replayArtifact.TraceID != firstArtifact.TraceID ||
		!bytes.Equal(replayArtifact.Body, firstArtifact.Body) {
		t.Fatalf("restart replay artifact changed: first=%+v replay=%+v", firstArtifact, replayArtifact)
	}

	assertStage0332Count(t, cleanupStore, `SELECT count(*) FROM investment.portfolios WHERE subject_id = $1`, 1, subjectID)
	assertStage0332Count(t, cleanupStore, `SELECT count(*) FROM investment.command_deduplication WHERE principal_id = $1 AND idempotency_key = $2`, 1, subjectID, key)
}
