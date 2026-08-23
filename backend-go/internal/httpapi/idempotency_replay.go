package httpapi

import (
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v3"

	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
)

func buildCommandReplayArtifact(meta metaDTO, status int, data any) (verticalslice.CommandReplayArtifact, error) {
	body, err := json.Marshal(baseResponse{Data: data, Meta: meta})
	if err != nil {
		return verticalslice.CommandReplayArtifact{}, err
	}
	return verticalslice.CommandReplayArtifact{
		StatusCode: status,
		Body:       body,
		RequestID:  meta.RequestID,
		TraceID:    meta.TraceID,
	}, nil
}

func writeCommandReplayArtifact(c fiber.Ctx, artifact verticalslice.CommandReplayArtifact) error {
	if artifact.StatusCode < 100 || artifact.StatusCode > 599 || len(artifact.Body) == 0 {
		return fmt.Errorf("invalid command replay artifact")
	}
	c.Set("X-Request-ID", artifact.RequestID)
	c.Set("X-Trace-ID", artifact.TraceID)
	c.Set("Content-Type", "application/json")
	return c.Status(artifact.StatusCode).Send(artifact.Body)
}

func writeReplayAwareError(c fiber.Ctx, meta metaDTO, err error) error {
	switch err {
	case verticalslice.ErrReplayUnavailable:
		return writeErrorWithMeta(c, meta, 503, "SERVICE_NOT_READY", "Idempotency replay storage is not available")
	case postgres.ErrUnsupportedDuplicate:
		return writeErrorWithMeta(c, meta, 409, "IDEMPOTENCY_REPLAY_UNAVAILABLE", "Completed idempotency record predates exact replay support")
	default:
		return writeMappedErrorWithMeta(c, meta, err)
	}
}
