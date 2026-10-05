package httpapi

import (
	"strings"

	"github.com/gofiber/fiber/v3"
)

const sensitiveResponseCacheControl = "no-store"

// sensitiveResponseCachePolicy applies a server-owned storage prohibition to routes that
// carry session material or authenticated portfolio data. Public routes intentionally keep
// their route-specific cache policy so they can be classified separately.
func sensitiveResponseCachePolicy(c fiber.Ctx) error {
	if isSensitiveResponsePath(c.Path()) {
		c.Set("Cache-Control", sensitiveResponseCacheControl)
	}
	return c.Next()
}

func isSensitiveResponsePath(path string) bool {
	return hasAPIPathPrefix(path, "/api/v1/auth") ||
		hasAPIPathPrefix(path, "/api/v1/portfolios")
}

func hasAPIPathPrefix(path string, prefix string) bool {
	return path == prefix || strings.HasPrefix(path, prefix+"/")
}
