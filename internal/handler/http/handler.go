package http

import (
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/jason2071/go-starter-kit-lite/internal/usecase"
)

const userContextKey = "auth_user"

type AuthUser struct {
	ID    uuid.UUID
	Roles []string
}

type Dependencies struct {
	Ready          func() error
	Auth           *usecase.AuthService
	Users          *usecase.UserService
	Tokens         usecase.TokenManager
	Logger         *slog.Logger
	AllowedOrigins string
}

type errorResponse struct {
	Success bool     `json:"success"`
	Error   apiError `json:"error"`
}
type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}
type dataResponse struct {
	Success bool `json:"success"`
	Data    any  `json:"data,omitempty"`
}

// Handler adapts HTTP requests to use cases. It does not register routes.
type Handler struct {
	ready     func() error
	auth      *usecase.AuthService
	users     *usecase.UserService
	validator *validator.Validate
}

func NewHandler(dep Dependencies) *Handler {
	return &Handler{ready: dep.Ready, auth: dep.Auth, users: dep.Users, validator: validator.New()}
}

func (h *Handler) Register(c *fiber.Ctx) error {
	var req usecase.RegisterRequest
	if err := parseAndValidate(c, h.validator, &req); err != nil {
		return err
	}
	out, err := h.auth.Register(c.UserContext(), req)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(dataResponse{true, out})
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var req usecase.LoginRequest
	if err := parseAndValidate(c, h.validator, &req); err != nil {
		return err
	}
	out, err := h.auth.Login(c.UserContext(), req)
	if err != nil {
		return err
	}
	return c.JSON(dataResponse{true, out})
}

func (h *Handler) Refresh(c *fiber.Ctx) error {
	var req usecase.RefreshRequest
	if err := parseAndValidate(c, h.validator, &req); err != nil {
		return err
	}
	out, err := h.auth.Refresh(c.UserContext(), req.RefreshToken)
	if err != nil {
		return err
	}
	return c.JSON(dataResponse{true, out})
}

func (h *Handler) Logout(c *fiber.Ctx) error {
	var req usecase.RefreshRequest
	if err := parseAndValidate(c, h.validator, &req); err != nil {
		return err
	}
	if err := h.auth.Logout(c.UserContext(), req.RefreshToken); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) Me(c *fiber.Ctx) error {
	out, err := h.auth.Me(c.UserContext(), currentUser(c).ID)
	if err != nil {
		return err
	}
	return c.JSON(dataResponse{true, out})
}

func (h *Handler) ListUsers(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	size, _ := strconv.Atoi(c.Query("page_size", "20"))
	var active *bool
	if raw := c.Query("is_active"); raw != "" {
		value, err := strconv.ParseBool(raw)
		if err != nil {
			return usecase.NewError(usecase.ErrValidation, "INVALID_FILTER", "is_active must be boolean")
		}
		active = &value
	}
	out, err := h.users.List(c.UserContext(), usecase.UserListRequest{Page: page, PageSize: size, Search: c.Query("search"), IsActive: active, Sort: c.Query("sort", "created_at"), Order: c.Query("order", "desc")})
	if err != nil {
		return err
	}
	return c.JSON(dataResponse{true, out})
}

func (h *Handler) Health(c *fiber.Ctx) error { return c.JSON(fiber.Map{"status": "ok"}) }

func (h *Handler) Ready(c *fiber.Ctx) error {
	if h.ready != nil && h.ready() != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"status": "not_ready"})
	}
	return c.JSON(fiber.Map{"status": "ready"})
}

func authMiddleware(tokens TokenParser) fiber.Handler {
	return func(c *fiber.Ctx) error {
		raw := strings.TrimSpace(c.Get(fiber.HeaderAuthorization))
		if !strings.HasPrefix(raw, "Bearer ") {
			return usecase.NewError(usecase.ErrUnauthorized, "UNAUTHORIZED", "missing bearer token")
		}
		claims, err := tokens.ParseAccessToken(strings.TrimSpace(strings.TrimPrefix(raw, "Bearer ")))
		if err != nil {
			return usecase.NewError(usecase.ErrUnauthorized, "UNAUTHORIZED", "invalid access token")
		}
		id, err := uuid.Parse(claims.Subject)
		if err != nil {
			return usecase.NewError(usecase.ErrUnauthorized, "UNAUTHORIZED", "invalid access token")
		}
		c.Locals(userContextKey, AuthUser{ID: id, Roles: claims.Roles})
		return c.Next()
	}
}
func requireRole(role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		for _, r := range currentUser(c).Roles {
			if r == role {
				return c.Next()
			}
		}
		return usecase.NewError(usecase.ErrForbidden, "FORBIDDEN", "insufficient permissions")
	}
}
func currentUser(c *fiber.Ctx) AuthUser { u, _ := c.Locals(userContextKey).(AuthUser); return u }
func requestLogger(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		logger.Info("http_request", "request_id", c.GetRespHeader(fiber.HeaderXRequestID), "method", c.Method(), "path", c.Path(), "status", c.Response().StatusCode(), "duration_ms", time.Since(start).Milliseconds())
		return err
	}
}
func parseAndValidate(c *fiber.Ctx, v *validator.Validate, dst any) error {
	if err := c.BodyParser(dst); err != nil {
		return usecase.NewError(usecase.ErrValidation, "INVALID_REQUEST", "invalid request body")
	}
	if err := v.Struct(dst); err != nil {
		return &usecase.AppError{Kind: usecase.ErrValidation, Code: "VALIDATION_ERROR", Message: "validation failed", Details: validationDetails(err)}
	}
	return nil
}
func validationDetails(err error) map[string]string {
	out := map[string]string{}
	var errs validator.ValidationErrors
	if errors.As(err, &errs) {
		for _, e := range errs {
			out[strings.ToLower(e.Field())] = "invalid value"
		}
	}
	return out
}
func errorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		var appErr *usecase.AppError
		if errors.As(err, &appErr) {
			status := map[usecase.ErrorKind]int{usecase.ErrValidation: 400, usecase.ErrUnauthorized: 401, usecase.ErrForbidden: 403, usecase.ErrNotFound: 404, usecase.ErrConflict: 409, usecase.ErrInternal: 500}[appErr.Kind]
			if status == 0 {
				status = 500
			}
			if status >= 500 {
				logger.Error("request_failed", "error", appErr.Err, "code", appErr.Code)
			}
			return c.Status(status).JSON(errorResponse{false, apiError{appErr.Code, appErr.Message, appErr.Details}})
		}
		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			return c.Status(fiberErr.Code).JSON(errorResponse{false, apiError{"HTTP_ERROR", fiberErr.Message, nil}})
		}
		logger.Error("unhandled_error", "error", err)
		return c.Status(500).JSON(errorResponse{false, apiError{"INTERNAL_ERROR", "internal server error", nil}})
	}
}

const swaggerHTML = `<!doctype html><html><head><meta charset="utf-8"><title>API Docs</title><link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"></head><body><div id="swagger-ui"></div><script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script><script>SwaggerUIBundle({url:'/openapi.yaml',dom_id:'#swagger-ui'});</script></body></html>`
