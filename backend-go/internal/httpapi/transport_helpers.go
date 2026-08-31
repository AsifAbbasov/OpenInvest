package httpapi

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/openinvest/openinvest/backend-go/internal/auth"
	"github.com/openinvest/openinvest/backend-go/internal/importer"
	"github.com/openinvest/openinvest/backend-go/internal/importflow"
	"github.com/openinvest/openinvest/backend-go/internal/postgres"
	"github.com/openinvest/openinvest/backend-go/internal/verticalslice"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func queryLimit(c fiber.Ctx, fallback int) int {
	value, err := strconv.Atoi(c.Query("limit", strconv.Itoa(fallback)))
	if err != nil || value <= 0 {
		return fallback
	}
	if value > 100 {
		return 100
	}
	return value
}

func queryLimitStrict(c fiber.Ctx, fallback int) (int, error) {
	raw, present := queryValue(c, "limit")
	if !present {
		return fallback, nil
	}
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0, fmt.Errorf("%w: limit must be an integer between 1 and 100", verticalslice.ErrInvalidInput)
	}
	value, err := strconv.Atoi(trimmed)
	if err != nil || value < 1 || value > 100 {
		return 0, fmt.Errorf("%w: limit must be an integer between 1 and 100", verticalslice.ErrInvalidInput)
	}
	return value, nil
}

func optionalQueryValue(c fiber.Ctx, name string) (string, error) {
	raw, present := queryValue(c, name)
	if !present {
		return "", nil
	}
	if strings.TrimSpace(raw) == "" {
		return "", fmt.Errorf("%w: %s must not be empty when supplied", verticalslice.ErrInvalidInput, name)
	}
	return raw, nil
}

func queryValue(c fiber.Ctx, name string) (string, bool) {
	args := c.Request().URI().QueryArgs()
	if !args.Has(name) {
		return "", false
	}
	return string(args.Peek(name)), true
}

func writeOK(c fiber.Ctx, data any) error {
	return writeStatus(c, http.StatusOK, data)
}

func writeStatus(c fiber.Ctx, status int, data any) error {
	return writeStatusWithMeta(c, requestMeta(c), status, data)
}

func writeStatusWithMeta(c fiber.Ctx, meta metaDTO, status int, data any) error {
	c.Set("X-Request-ID", meta.RequestID)
	c.Set("X-Trace-ID", meta.TraceID)
	return c.Status(status).JSON(baseResponse{Data: data, Meta: meta})
}

func writeMappedError(c fiber.Ctx, err error) error {
	return writeMappedErrorWithMeta(c, requestMeta(c), err)
}

func writeMappedErrorWithMeta(c fiber.Ctx, meta metaDTO, err error) error {
	switch {
	case errors.Is(err, errAuthRateLimited):
		c.Set("Retry-After", authRateLimitRetryAfterSeconds)
		return writeErrorWithMeta(c, meta, http.StatusTooManyRequests, "RATE_LIMITED", "Too many authentication attempts")
	case errors.Is(err, auth.ErrInvalidInput):
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, auth.ErrEmailAlreadyExists):
		return writeErrorWithMeta(c, meta, http.StatusConflict, "CONFLICT", "Email is already registered")
	case errors.Is(err, auth.ErrInvalidCredentials):
		return writeErrorWithMeta(c, meta, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid email or password")
	case errors.Is(err, auth.ErrInvalidSession):
		return writeErrorWithMeta(c, meta, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired session")
	case errors.Is(err, auth.ErrInvalidCSRF):
		return writeErrorWithMeta(c, meta, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid CSRF token")
	case errors.Is(err, verticalslice.ErrInvalidInput):
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, verticalslice.ErrMissingIdempotency):
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "Idempotency-Key header is required")
	case errors.Is(err, verticalslice.ErrNotFound):
		return writeErrorWithMeta(c, meta, http.StatusNotFound, "NOT_FOUND", "Resource not found")
	case errors.Is(err, importer.ErrInvalidImport):
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, importer.ErrUnsafeAppend):
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, importflow.ErrInvalidFlowInput):
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, importflow.ErrNoApprovedRows):
		return writeErrorWithMeta(c, meta, http.StatusBadRequest, "VALIDATION_ERROR", "at least one appendable import row must be approved")
	case errors.Is(err, postgres.ErrNotFound):
		return writeErrorWithMeta(c, meta, http.StatusNotFound, "NOT_FOUND", "Resource not found")
	case errors.Is(err, postgres.ErrIdempotencyConflict):
		return writeErrorWithMeta(c, meta, http.StatusConflict, "IDEMPOTENCY_CONFLICT", "Idempotency-Key is already bound to another request")
	case errors.Is(err, postgres.ErrIdempotencyInFlight):
		return writeErrorWithMeta(c, meta, http.StatusConflict, "IDEMPOTENCY_IN_FLIGHT", "Idempotency-Key is currently being processed")
	default:
		return writeErrorWithMeta(c, meta, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
	}
}

func writeError(c fiber.Ctx, status int, code string, message string) error {
	return writeErrorWithMeta(c, requestMeta(c), status, code, message)
}

func writeErrorWithMeta(c fiber.Ctx, meta metaDTO, status int, code string, message string) error {
	c.Set("X-Request-ID", meta.RequestID)
	c.Set("X-Trace-ID", meta.TraceID)
	return c.Status(status).JSON(errorResponse{
		Error: errorBody{Code: code, Message: message, Details: []errorDetailDTO{}},
		Meta:  meta,
	})
}

func requestMeta(c fiber.Ctx) metaDTO {
	requestID := strings.TrimSpace(c.Get("X-Request-ID"))
	if _, err := uuid.Parse(requestID); err != nil {
		requestID = uuid.NewString()
	}
	traceIDValue := traceID()
	if traceparent := strings.TrimSpace(c.Get("traceparent")); validTraceparent(traceparent) {
		traceIDValue = strings.Split(traceparent, "-")[1]
	}
	return metaDTO{
		RequestID:   requestID,
		TraceID:     traceIDValue,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func (meta metaDTO) toApp() verticalslice.RequestContext {
	return verticalslice.RequestContext{RequestID: meta.RequestID, TraceID: meta.TraceID}
}

func traceID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "00000000000000000000000000000001"
	}
	return hex.EncodeToString(bytes)
}

func validTraceparent(value string) bool {
	parts := strings.Split(value, "-")
	if len(parts) != 4 {
		return false
	}
	version, traceIDPart, parentIDPart, flags := parts[0], parts[1], parts[2], parts[3]
	return len(version) == 2 &&
		len(traceIDPart) == 32 &&
		len(parentIDPart) == 16 &&
		len(flags) == 2 &&
		version != "ff" &&
		traceIDPart != strings.Repeat("0", 32) &&
		parentIDPart != strings.Repeat("0", 16) &&
		isLowerHex(version) &&
		isLowerHex(traceIDPart) &&
		isLowerHex(parentIDPart) &&
		isLowerHex(flags)
}

func isLowerHex(value string) bool {
	for _, char := range value {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
			return false
		}
	}
	return true
}

func jsonFieldPresent(body []byte, field string) bool {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return false
	}
	_, ok := raw[field]
	return ok
}

func decodeStrictJSON(body []byte, target any) error {
	if len(bytes.TrimSpace(body)) == 0 {
		return io.EOF
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("unexpected trailing JSON")
	}
	return nil
}

type baseResponse struct {
	Data any     `json:"data"`
	Meta metaDTO `json:"meta"`
}

type errorResponse struct {
	Error errorBody `json:"error"`
	Meta  metaDTO   `json:"meta"`
}

type errorBody struct {
	Code    string           `json:"code"`
	Message string           `json:"message"`
	Details []errorDetailDTO `json:"details"`
}

type errorDetailDTO struct {
	Field   *string `json:"field"`
	Code    string  `json:"code"`
	Message string  `json:"message"`
}

type metaDTO struct {
	RequestID   string `json:"requestId"`
	TraceID     string `json:"traceId"`
	GeneratedAt string `json:"generatedAt"`
}

type paginationDTO struct {
	NextCursor *string `json:"nextCursor"`
	HasMore    bool    `json:"hasMore"`
	Limit      int     `json:"limit"`
}

type listData[T any] struct {
	Items      []T           `json:"items"`
	Pagination paginationDTO `json:"pagination"`
}
