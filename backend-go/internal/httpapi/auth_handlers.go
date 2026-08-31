package httpapi

import (
	"bytes"
	"github.com/gofiber/fiber/v3"
	"github.com/openinvest/openinvest/backend-go/internal/auth"
	"net/http"
	"os"
	"strings"
	"time"
)

func (api *API) register(c fiber.Ctx) error {
	meta := requestMeta(c)
	if err := api.checkAuthRateLimit(c); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	if api.auth == nil {
		return writeErrorWithMeta(c, meta, http.StatusServiceUnavailable, "SERVICE_NOT_READY", "Authentication service is not ready")
	}
	var request registerRequestDTO
	if err := decodeStrictJSON(c.Request().Body(), &request); err != nil {
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
	}
	result, err := api.auth.Register(c.Context(), auth.RegistrationRequest{
		Email:    request.Email,
		Password: string(request.Password),
		Language: request.Language,
		Theme:    request.Theme,
		Timezone: request.Timezone,
	})
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	api.setRefreshCookie(c, result.RefreshToken)
	return writeStatusWithMeta(c, meta, http.StatusCreated, mapAuthData(result))
}

func (api *API) login(c fiber.Ctx) error {
	meta := requestMeta(c)
	if err := api.checkAuthRateLimit(c); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	if api.auth == nil {
		return writeErrorWithMeta(c, meta, http.StatusServiceUnavailable, "SERVICE_NOT_READY", "Authentication service is not ready")
	}
	var request loginRequestDTO
	if err := decodeStrictJSON(c.Request().Body(), &request); err != nil {
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
	}
	result, err := api.auth.Login(c.Context(), auth.LoginRequest{Email: request.Email, Password: string(request.Password)})
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	api.setRefreshCookie(c, result.RefreshToken)
	return writeStatusWithMeta(c, meta, http.StatusOK, mapAuthData(result))
}

func (api *API) refresh(c fiber.Ctx) error {
	meta := requestMeta(c)
	if err := api.checkAuthRateLimit(c); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	if api.auth == nil {
		return writeErrorWithMeta(c, meta, http.StatusServiceUnavailable, "SERVICE_NOT_READY", "Authentication service is not ready")
	}
	result, err := api.auth.Refresh(c.Context(), c.Cookies(auth.RefreshCookieName), c.Get("X-CSRF-Token"))
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	api.setRefreshCookie(c, result.RefreshToken)
	return writeStatusWithMeta(c, meta, http.StatusOK, mapAuthSession(result.Session))
}

func (api *API) logout(c fiber.Ctx) error {
	meta := requestMeta(c)
	if err := api.checkAuthRateLimit(c); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	if api.auth == nil {
		return writeErrorWithMeta(c, meta, http.StatusServiceUnavailable, "SERVICE_NOT_READY", "Authentication service is not ready")
	}
	var request logoutRequestDTO
	if len(bytes.TrimSpace(c.Request().Body())) == 0 {
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "allSessions is required")
	}
	if err := decodeStrictJSON(c.Request().Body(), &request); err != nil {
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
	}
	if request.AllSessions == nil {
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "allSessions is required")
	}
	revoked, err := api.auth.Logout(c.Context(), c.Cookies(auth.RefreshCookieName), c.Get("X-CSRF-Token"), *request.AllSessions)
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	api.clearRefreshCookie(c)
	return writeStatusWithMeta(c, meta, http.StatusOK, logoutResultDTO{Revoked: revoked})
}

func (api *API) subjectID(c fiber.Ctx) (string, error) {
	if api.auth != nil {
		if token := bearerToken(c.Get("Authorization")); token != "" {
			return api.auth.AuthenticateAccessToken(token)
		}
		if api.auth.AllowsDevelopmentBypass() {
			return developmentSubjectID(), nil
		}
		return "", auth.ErrInvalidSession
	}
	if api.allowDevelopmentSubject {
		return developmentSubjectID(), nil
	}
	return "", auth.ErrInvalidSession
}

func developmentSubjectID() string {
	if configured := strings.TrimSpace(os.Getenv("OPENINVEST_DEV_SUBJECT_ID")); configured != "" {
		return configured
	}
	return devSubjectID
}

func bearerToken(header string) string {
	prefix := "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}

func (api *API) setRefreshCookie(c fiber.Ctx, value string) {
	c.Cookie(&fiber.Cookie{
		Name:     auth.RefreshCookieName,
		Value:    value,
		Path:     "/api/v1/auth",
		MaxAge:   api.auth.RefreshCookieMaxAgeSeconds(),
		SameSite: "Strict",
		Secure:   api.auth.RefreshCookieSecure(),
		HTTPOnly: true,
	})
}

func (api *API) clearRefreshCookie(c fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     auth.RefreshCookieName,
		Value:    "",
		Path:     "/api/v1/auth",
		MaxAge:   -1,
		SameSite: "Strict",
		Secure:   api.auth.RefreshCookieSecure(),
		HTTPOnly: true,
	})
}

type registerRequestDTO struct {
	Email    string           `json:"email"`
	Password losslessPassword `json:"password"`
	Language string           `json:"language"`
	Theme    string           `json:"theme"`
	Timezone string           `json:"timezone"`
}

type loginRequestDTO struct {
	Email    string           `json:"email"`
	Password losslessPassword `json:"password"`
}

type logoutRequestDTO struct {
	AllSessions *bool `json:"allSessions"`
}

type userDTO struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	Language    string `json:"language"`
	Theme       string `json:"theme"`
	Timezone    string `json:"timezone"`
	PrivacyMode bool   `json:"privacyMode"`
	CreatedAt   string `json:"createdAt"`
}

type authSessionDTO struct {
	AccessToken          string `json:"accessToken"`
	AccessTokenExpiresAt string `json:"accessTokenExpiresAt"`
	CSRFToken            string `json:"csrfToken"`
}

type authDataDTO struct {
	User    userDTO        `json:"user"`
	Session authSessionDTO `json:"session"`
}

type logoutResultDTO struct {
	Revoked bool `json:"revoked"`
}

func mapAuthData(result auth.AuthResult) authDataDTO {
	return authDataDTO{User: mapUser(result.User), Session: mapAuthSession(result.Session)}
}

func mapUser(user auth.StoredUser) userDTO {
	return userDTO{
		ID:          user.ID,
		Email:       user.Email,
		Language:    user.Language,
		Theme:       user.Theme,
		Timezone:    user.Timezone,
		PrivacyMode: user.PrivacyMode,
		CreatedAt:   user.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func mapAuthSession(session auth.ClientSession) authSessionDTO {
	return authSessionDTO{
		AccessToken:          session.AccessToken,
		AccessTokenExpiresAt: session.AccessTokenExpiresAt.UTC().Format(time.RFC3339),
		CSRFToken:            session.CSRFToken,
	}
}
