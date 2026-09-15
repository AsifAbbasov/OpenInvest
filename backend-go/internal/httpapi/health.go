package httpapi

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
	"net/http"
)

const readinessTimeout = 2 * time.Second

func (api *API) health(c fiber.Ctx) error {
	return writeOK(c, map[string]string{"status": "ok"})
}

func (api *API) ready(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), readinessTimeout)
	defer cancel()
	if err := api.service.Ready(ctx); err != nil {
		return writeError(c, http.StatusServiceUnavailable, "SERVICE_NOT_READY", "Service is not ready")
	}
	return writeOK(c, map[string]string{"status": "ready"})
}
