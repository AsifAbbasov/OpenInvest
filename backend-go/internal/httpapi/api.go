package httpapi

import (
	"github.com/gofiber/fiber/v3"
	"github.com/openinvest/openinvest/backend-go/internal/auth"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
	"time"
)

type API struct {
	service                 *verticalslice.Service
	auth                    *auth.Service
	corporateActionProvider verticalslice.CorporateActionProvider
	allowDevelopmentSubject bool
	authLimiter             *authRateLimiter
	importReviewSecret      []byte
	paginationCursorSecret  []byte
	now                     func() time.Time
}

func New(service *verticalslice.Service, authService *auth.Service, importReviewTokenSecret []byte) (*fiber.App, error) {
	secret, err := normalizedImportReviewSecret(importReviewTokenSecret)
	if err != nil {
		return nil, err
	}
	return newApp(&API{
		service:                service,
		auth:                   authService,
		authLimiter:            newAuthRateLimiter(20, time.Minute),
		importReviewSecret:     secret,
		paginationCursorSecret: derivePaginationCursorSecret(secret),
	}), nil
}

func NewDevelopment(service *verticalslice.Service) *fiber.App {
	secret, err := normalizedImportReviewSecret([]byte("openinvest-development-import-review-token-secret"))
	if err != nil {
		panic(err)
	}
	return newApp(&API{
		service:                 service,
		allowDevelopmentSubject: true,
		authLimiter:             newAuthRateLimiter(20, time.Minute),
		importReviewSecret:      secret,
		paginationCursorSecret:  derivePaginationCursorSecret(secret),
	})
}

func (api *API) nowUTC() time.Time {
	if api.now != nil {
		return api.now().UTC()
	}
	return time.Now().UTC()
}

func newApp(api *API) *fiber.App {
	app := fiber.New(fiber.Config{AppName: "OpenInvest API"})

	app.Use(localDevelopmentCORS)
	registerRoutes(app, api)

	return app
}
