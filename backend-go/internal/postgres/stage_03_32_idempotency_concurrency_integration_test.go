package postgres

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"

	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func TestStage0332ConcurrentIdenticalCreateProducesOneEffectAndOneArtifact(t *testing.T) {
	store := openStage0332Store(t)
	ctx := context.Background()
	subjectID := uuid.NewString()
	cleanupStage0332Subject(t, store, subjectID)
	service := verticalslice.NewService(store, verticalslice.SystemClock{})
	request := verticalslice.CreatePortfolioRequest{Name: "Concurrent replay", BaseCurrency: verticalslice.RUB}
	key := "stage-03-32-concurrent-key-0001"
	path := "/api/v1/portfolios"

	start := make(chan struct{})
	type outcome struct {
		artifact verticalslice.CommandReplayArtifact
		err      error
	}
	outcomes := make(chan outcome, 2)
	var builderCalls atomic.Int32
	var wait sync.WaitGroup
	wait.Add(2)

	for index := 0; index < 2; index++ {
		index := index
		go func() {
			defer wait.Done()
			<-start
			requestID := uuid.NewString()
			traceID := fmt.Sprintf("%032x", index+1)
			_, artifact, err := service.CreatePortfolioWithReplay(
				ctx,
				verticalslice.RequestContext{RequestID: requestID, TraceID: traceID},
				subjectID,
				key,
				path,
				request,
				func(verticalslice.Portfolio) (verticalslice.CommandReplayArtifact, error) {
					builderCalls.Add(1)
					return verticalslice.CommandReplayArtifact{
						StatusCode: 201,
						Body:       []byte(`{"data":{"created":true},"meta":{"requestId":"` + requestID + `"}}`),
						RequestID:  requestID,
						TraceID:    traceID,
					}, nil
				},
			)
			outcomes <- outcome{artifact: artifact, err: err}
		}()
	}

	close(start)
	wait.Wait()
	close(outcomes)

	results := make([]outcome, 0, 2)
	for result := range outcomes {
		if result.err != nil {
			t.Fatalf("concurrent create failed: %v", result.err)
		}
		results = append(results, result)
	}
	if len(results) != 2 {
		t.Fatalf("expected two results, got %d", len(results))
	}
	if builderCalls.Load() != 1 {
		t.Fatalf("expected exactly one response builder execution, got %d", builderCalls.Load())
	}
	if results[0].artifact.StatusCode != results[1].artifact.StatusCode ||
		results[0].artifact.RequestID != results[1].artifact.RequestID ||
		results[0].artifact.TraceID != results[1].artifact.TraceID ||
		!bytes.Equal(results[0].artifact.Body, results[1].artifact.Body) {
		t.Fatalf("concurrent callers did not receive the same artifact: first=%+v second=%+v", results[0].artifact, results[1].artifact)
	}

	assertStage0332CountArgs(t, store, `SELECT count(*) FROM investment.portfolios WHERE subject_id = $1`, 1, subjectID)
	assertStage0332CountArgs(t, store, `SELECT count(*) FROM investment.command_deduplication WHERE principal_id = $1 AND idempotency_key = $2`, 1, subjectID, key)
}
