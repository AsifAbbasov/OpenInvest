package httpapi

import (
	"github.com/gofiber/fiber/v3"
	"net/http"
)

func (api *API) health(c fiber.Ctx) error {
	return writeOK(c, map[string]string{"status": "ok"})
}

func (api *API) ready(c fiber.Ctx) error {
	if err := api.service.Ready(c.Context()); err != nil {
		return writeError(c, http.StatusServiceUnavailable, "SERVICE_NOT_READY", "Service is not ready")
	}
	return writeOK(c, map[string]string{"status": "ready"})
}
