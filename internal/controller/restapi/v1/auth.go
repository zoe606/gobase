package v1

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go-boilerplate/internal/controller/restapi/middleware"
	"go-boilerplate/internal/controller/restapi/v1/request"
	authdto "go-boilerplate/internal/dto/auth"
	"go-boilerplate/internal/usecase/auth"
	"go-boilerplate/pkg/jwt"
	"go-boilerplate/pkg/logger"
	"go-boilerplate/pkg/response"
)

// AuthController handles authentication endpoints.
type AuthController struct {
	authUC     *auth.UseCase
	jwtService jwt.Service
	l          logger.Interface
	v          *validator.Validate
}

// NewAuthController creates a new auth controller.
func NewAuthController(authUC *auth.UseCase, jwtService jwt.Service, l logger.Interface) *AuthController {
	return &AuthController{
		authUC:     authUC,
		jwtService: jwtService,
		l:          l,
		v:          validator.New(validator.WithRequiredStructEnabled()),
	}
}

// NewAuthRoutes sets up auth routes.
func NewAuthRoutes(router fiber.Router, authUC *auth.UseCase, jwtService jwt.Service, l logger.Interface) {
	c := NewAuthController(authUC, jwtService, l)

	authGroup := router.Group("/auth")
	{
		// Public routes
		authGroup.Post("/register", c.Register)
		authGroup.Post("/login", c.Login)
		authGroup.Post("/refresh", c.Refresh)

		// Protected routes
		authGroup.Post("/logout", middleware.JWTAuth(jwtService, l), c.Logout)
		authGroup.Get("/me", middleware.JWTAuth(jwtService, l), c.Me)
	}
}

// Register godoc
// @Summary     Register a new user
// @Description Register a new user account
// @ID          auth-register
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       request body request.Register true "Registration details"
// @Success     201 {object} response.Response[auth.LoginResponse]
// @Failure     400 {object} response.ErrorResponse
// @Failure     409 {object} response.ErrorResponse
// @Router      /auth/register [post]
func (c *AuthController) Register(ctx *fiber.Ctx) error {
	var req request.Register
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, "INVALID_JSON", "Invalid request body")
	}

	if err := c.v.Struct(req); err != nil {
		return response.ValidationError(ctx, parseValidationErrors(err))
	}

	result, err := c.authUC.Register(ctx.UserContext(), authdto.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	})
	if err != nil {
		if errors.Is(err, auth.ErrEmailExists) {
			return response.Conflict(ctx, "Email already exists")
		}
		c.l.Error(err, "restapi - v1 - auth - Register")
		return response.InternalError(ctx)
	}

	return response.Created(ctx, result.ToResponse())
}

// Login godoc
// @Summary     Login
// @Description Authenticate user and get tokens
// @ID          auth-login
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       request body request.Login true "Login credentials"
// @Success     200 {object} response.Response[auth.LoginResponse]
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Router      /auth/login [post]
func (c *AuthController) Login(ctx *fiber.Ctx) error {
	var req request.Login
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, "INVALID_JSON", "Invalid request body")
	}

	if err := c.v.Struct(req); err != nil {
		return response.ValidationError(ctx, parseValidationErrors(err))
	}

	result, err := c.authUC.Login(ctx.UserContext(), authdto.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return response.Unauthorized(ctx, "Invalid email or password")
		}
		if errors.Is(err, auth.ErrUserNotActive) {
			return response.Forbidden(ctx, "Account is not active")
		}
		c.l.Error(err, "restapi - v1 - auth - Login")
		return response.InternalError(ctx)
	}

	return response.OK(ctx, result.ToResponse())
}

// Refresh godoc
// @Summary     Refresh tokens
// @Description Refresh access token using refresh token
// @ID          auth-refresh
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       request body request.RefreshToken true "Refresh token"
// @Success     200 {object} response.Response[auth.TokenResponse]
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Router      /auth/refresh [post]
func (c *AuthController) Refresh(ctx *fiber.Ctx) error {
	var req request.RefreshToken
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, "INVALID_JSON", "Invalid request body")
	}

	if err := c.v.Struct(req); err != nil {
		return response.ValidationError(ctx, parseValidationErrors(err))
	}

	result, err := c.authUC.Refresh(ctx.UserContext(), authdto.RefreshInput{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		if errors.Is(err, auth.ErrInvalidToken) {
			return response.Unauthorized(ctx, "Invalid or expired refresh token")
		}
		c.l.Error(err, "restapi - v1 - auth - Refresh")
		return response.InternalError(ctx)
	}

	return response.OK(ctx, result.ToResponse())
}

// Logout godoc
// @Summary     Logout
// @Description Invalidate refresh token
// @ID          auth-logout
// @Tags        auth
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       request body request.RefreshToken true "Refresh token to invalidate"
// @Success     204
// @Failure     400 {object} response.ErrorResponse
// @Failure     401 {object} response.ErrorResponse
// @Router      /auth/logout [post]
func (c *AuthController) Logout(ctx *fiber.Ctx) error {
	var req request.RefreshToken
	if err := ctx.BodyParser(&req); err != nil {
		return response.BadRequest(ctx, "INVALID_JSON", "Invalid request body")
	}

	if err := c.authUC.Logout(ctx.UserContext(), req.RefreshToken); err != nil {
		c.l.Error(err, "restapi - v1 - auth - Logout")
		return response.InternalError(ctx)
	}

	return response.NoContent(ctx)
}

// Me godoc
// @Summary     Get current user
// @Description Get currently authenticated user info
// @ID          auth-me
// @Tags        auth
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} response.Response[auth.UserResponse]
// @Failure     401 {object} response.ErrorResponse
// @Router      /auth/me [get]
func (c *AuthController) Me(ctx *fiber.Ctx) error {
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		return response.Unauthorized(ctx, "User not found in context")
	}

	result, err := c.authUC.GetCurrentUser(ctx.UserContext(), userID)
	if err != nil {
		c.l.Error(err, "restapi - v1 - auth - Me")
		return response.InternalError(ctx)
	}

	if result == nil {
		return response.NotFound(ctx, "User not found")
	}

	return response.OK(ctx, result.ToResponse())
}
