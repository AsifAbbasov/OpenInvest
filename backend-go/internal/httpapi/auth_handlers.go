package httpapi

import (
	"bytes"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/openinvest/openinvest/backend-go/internal/auth"
)

func (api *API) register(c fiber.Ctx) error {
	meta := requestMeta(c)
	if err := api.checkAuthLoginIPRateLimit(c); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	if api.auth == nil {
		return writeErrorWithMeta(c, meta, http.StatusServiceUnavailable, "SERVICE_NOT_READY", "Authentication service is not ready")
	}
	var request struct {
		Email string `json:"email"`
	}
	if err := decodeStrictJSON(c.Request().Body(), &request); err != nil {
		return writeErrorWithMeta(c, meta, 400, "VALIDATION_ERROR", "Invalid JSON request body")
	}
	if err := api.auth.BeginEmailRegistration(c.Context(), request.Email); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	return writeStatusWithMeta(c, meta, http.StatusAccepted, map[string]string{"status": "REQUEST_ACCEPTED", "message": "If this address is eligible, instructions will be sent."})
}

func (api *API) verifyEmail(c fiber.Ctx) error {
	meta := requestMeta(c)
	if err := api.checkAuthRateLimit(c); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	if api.auth == nil {
		return writeErrorWithMeta(c, meta, 503, "SERVICE_NOT_READY", "Authentication service is not ready")
	}
	var request struct {
		Token    string           `json:"token"`
		Password losslessPassword `json:"password"`
		Language string           `json:"language"`
		Theme    string           `json:"theme"`
		Timezone string           `json:"timezone"`
	}
	if err := decodeStrictJSON(c.Request().Body(), &request); err != nil {
		return writeErrorWithMeta(c, meta, 400, "VALIDATION_ERROR", "Invalid JSON request body")
	}
	if err := api.auth.CompleteEmailRegistration(c.Context(), request.Token, auth.RegistrationRequest{Password: string(request.Password), Language: request.Language, Theme: request.Theme, Timezone: request.Timezone}); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	return writeStatusWithMeta(c, meta, 200, map[string]string{"status": "ACCOUNT_READY"})
}

func (api *API) login(c fiber.Ctx) error {
	meta := requestMeta(c)
	if err := api.checkAuthLoginIPRateLimit(c); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	if api.auth == nil {
		return writeErrorWithMeta(c, meta, http.StatusServiceUnavailable, "SERVICE_NOT_READY", "Authentication service is not ready")
	}
	var request loginRequestDTO
	if err := decodeStrictJSON(c.Request().Body(), &request); err != nil {
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
	}
	if err := api.checkAuthCredentialRateLimit(c.Path(), request.Email); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	result, err := api.auth.Login(c.Context(), auth.LoginRequest{Email: request.Email, Password: string(request.Password)})
	if err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}
	api.setRefreshCookie(c, result.RefreshToken)
	return writeStatusWithMeta(c, meta, http.StatusOK, mapAuthData(result))
}

// bootstrap restores authentication using the existing refresh cookie.
// It does not rotate the refresh token or issue a replacement cookie.
func (api *API) bootstrap(c fiber.Ctx) error {
	meta := requestMeta(c)

	if err := validateBootstrapOrigin(c); err != nil {
		return writeErrorWithMeta(
			c,
			meta,
			http.StatusForbidden,
			"BOOTSTRAP_ORIGIN_DENIED",
			"Bootstrap request origin is not permitted",
		)
	}

	if !validBootstrapRequest(c) {
		return writeErrorWithMeta(
			c,
			meta,
			http.StatusBadRequest,
			"BOOTSTRAP_INVALID_REQUEST",
			"Invalid bootstrap request",
		)
	}

	if err := api.checkAuthRateLimit(c); err != nil {
		return writeMappedErrorWithMeta(c, meta, err)
	}

	if api.auth == nil {
		return writeErrorWithMeta(
			c,
			meta,
			http.StatusServiceUnavailable,
			"SERVICE_NOT_READY",
			"Authentication service is not ready",
		)
	}

	refreshToken := c.Cookies(auth.RefreshCookieName)
	if strings.TrimSpace(refreshToken) == "" {
		return writeErrorWithMeta(
			c,
			meta,
			http.StatusUnauthorized,
			"SESSION_INVALID",
			"Authentication session is invalid",
		)
	}

	result, err := api.auth.Bootstrap(c.Context(), refreshToken)
	if err != nil {
		if err == auth.ErrInvalidSession {
			return writeErrorWithMeta(
				c,
				meta,
				http.StatusUnauthorized,
				"SESSION_INVALID",
				"Authentication session is invalid",
			)
		}
		return writeMappedErrorWithMeta(c, meta, err)
	}

	// Bootstrap must never invoke setRefreshCookie.
	// The original HttpOnly refresh cookie remains unchanged.
	return writeStatusWithMeta(c, meta, http.StatusOK, mapAuthData(result))
}

func validateBootstrapOrigin(c fiber.Ctx) error {
	origins := c.Request().Header.PeekAll("Origin")
	if len(origins) != 1 {
		return auth.ErrInvalidInput
	}

	origin := string(origins[0])
	if origin == "" || origin != strings.TrimSpace(origin) {
		return auth.ErrInvalidInput
	}

	if origin == "null" || strings.ContainsAny(origin, ",\r\n\t ") {
		return auth.ErrInvalidInput
	}

	parsed, err := url.Parse(origin)
	if err != nil {
		return auth.ErrInvalidInput
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return auth.ErrInvalidInput
	}

	if parsed.Host == "" ||
		parsed.User != nil ||
		parsed.Path != "" ||
		parsed.RawQuery != "" ||
		parsed.Fragment != "" ||
		parsed.Opaque != "" {
		return auth.ErrInvalidInput
	}

	if parsed.String() != origin {
		return auth.ErrInvalidInput
	}

	if !allowedWebOrigin(origin) {
		return auth.ErrInvalidInput
	}

	fetchSite := c.Request().Header.PeekAll("Sec-Fetch-Site")
	if len(fetchSite) > 1 {
		return auth.ErrInvalidInput
	}
	if len(fetchSite) == 1 {
		switch string(fetchSite[0]) {
		case "same-origin", "same-site":
		default:
			return auth.ErrInvalidInput
		}
	}

	fetchMode := c.Request().Header.PeekAll("Sec-Fetch-Mode")
	if len(fetchMode) > 1 {
		return auth.ErrInvalidInput
	}
	if len(fetchMode) == 1 && string(fetchMode[0]) != "cors" {
		return auth.ErrInvalidInput
	}

	fetchDest := c.Request().Header.PeekAll("Sec-Fetch-Dest")
	if len(fetchDest) > 1 {
		return auth.ErrInvalidInput
	}
	if len(fetchDest) == 1 && string(fetchDest[0]) != "empty" {
		return auth.ErrInvalidInput
	}

	return nil
}

func validBootstrapRequest(c fiber.Ctx) bool {
	if c.Method() != http.MethodPost {
		return false
	}

	if len(c.Request().Body()) != 0 {
		return false
	}

	headers := c.Request().Header.PeekAll("X-OpenInvest-Bootstrap")
	if len(headers) != 1 || string(headers[0]) != "1" {
		return false
	}

	acceptHeaders := c.Request().Header.PeekAll("Accept")
	if len(acceptHeaders) != 1 {
		return false
	}

	accept := strings.TrimSpace(string(acceptHeaders[0]))
	return accept == "application/json"
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
