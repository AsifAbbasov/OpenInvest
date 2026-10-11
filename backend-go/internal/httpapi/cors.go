package httpapi

import (
	"github.com/gofiber/fiber/v3"
	"net/http"
	"os"
	"strings"
)

func localDevelopmentCORS(c fiber.Ctx) error {
	// Origin is not a list-valued header. Inspect every field before using
	// its value so duplicates cannot be hidden by a single-value lookup.
	origins := c.Request().Header.PeekAll("Origin")
	if len(origins) > 1 {
		return c.SendStatus(http.StatusBadRequest)
	}
	origin := ""
	if len(origins) == 1 {
		origin = strings.TrimSpace(string(origins[0]))
		// Intermediaries may combine repeated fields into one comma list.
		if strings.Contains(origin, ",") {
			return c.SendStatus(http.StatusBadRequest)
		}
	}
	if origin != "" && allowedWebOrigin(origin) {
		c.Set("Access-Control-Allow-Origin", origin)
		c.Set("Access-Control-Allow-Credentials", "true")
		c.Set("Vary", "Origin")
		c.Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, Idempotency-Key, X-CSRF-Token, X-Request-ID, traceparent")
		c.Set("Access-Control-Expose-Headers", "X-Request-ID, X-Trace-ID")
	}
	if c.Method() == http.MethodOptions {
		return c.SendStatus(http.StatusNoContent)
	}
	return c.Next()
}

func allowedWebOrigin(origin string) bool {
	configured := strings.TrimSpace(os.Getenv("OPENINVEST_ALLOWED_WEB_ORIGINS"))
	if configured == "" {
		configured = "http://localhost:3000,http://127.0.0.1:3000"
	}
	for _, allowed := range strings.Split(configured, ",") {
		if strings.TrimSpace(allowed) == origin {
			return true
		}
	}
	return false
}
