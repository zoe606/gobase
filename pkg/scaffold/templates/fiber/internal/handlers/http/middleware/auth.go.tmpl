package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"go-boilerplate/pkg/jwt"
	"go-boilerplate/pkg/logger"
	"go-boilerplate/pkg/response"
)

// Context keys for storing user information.
const (
	UserIDKey      = "user_id"
	EmailKey       = "email"
	RoleKey        = "role"
	PermissionsKey = "permissions"
)

// JWTAuth validates JWT tokens and extracts claims to context.
func JWTAuth(jwtService jwt.Service, l logger.Interface) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return response.Unauthorized(c, "Missing authorization header")
		}

		// Extract Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return response.Unauthorized(c, "Invalid authorization format. Use: Bearer <token>")
		}

		tokenString := parts[1]

		// Validate token
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			l.Debug("JWT validation failed: %v", err)
			if errors.Is(err, jwt.ErrExpiredToken) {
				return response.Unauthorized(c, "Token has expired")
			}
			return response.Unauthorized(c, "Invalid token")
		}

		// Store claims in context
		c.Locals(UserIDKey, claims.UserID)
		c.Locals(EmailKey, claims.Email)
		c.Locals(RoleKey, claims.Role)
		c.Locals(PermissionsKey, claims.Permissions)

		return c.Next()
	}
}

// RequireRole checks if the authenticated user has one of the required roles.
func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole, ok := c.Locals(RoleKey).(string)
		if !ok || userRole == "" {
			return response.Forbidden(c, "No role found in token")
		}

		for _, role := range roles {
			if strings.EqualFold(userRole, role) {
				return c.Next()
			}
		}

		return response.Forbidden(c, "Insufficient role permissions")
	}
}

// RequirePermission checks if the authenticated user has the required permission.
func RequirePermission(permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		permissions, ok := c.Locals(PermissionsKey).([]string)
		if !ok {
			return response.Forbidden(c, "No permissions found in token")
		}

		for _, p := range permissions {
			if p == permission {
				return c.Next()
			}
		}

		return response.Forbidden(c, "Missing required permission: "+permission)
	}
}

// RequireAnyPermission checks if the authenticated user has any of the required permissions.
func RequireAnyPermission(permissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userPermissions, ok := c.Locals(PermissionsKey).([]string)
		if !ok {
			return response.Forbidden(c, "No permissions found in token")
		}

		permSet := make(map[string]struct{}, len(userPermissions))
		for _, p := range userPermissions {
			permSet[p] = struct{}{}
		}

		for _, required := range permissions {
			if _, ok := permSet[required]; ok {
				return c.Next()
			}
		}

		return response.Forbidden(c, "Missing required permissions")
	}
}

// GetUserID extracts the user ID from context.
func GetUserID(c *fiber.Ctx) uint {
	if id, ok := c.Locals(UserIDKey).(uint); ok {
		return id
	}
	return 0
}

// GetEmail extracts the email from context.
func GetEmail(c *fiber.Ctx) string {
	if email, ok := c.Locals(EmailKey).(string); ok {
		return email
	}
	return ""
}

// GetRole extracts the role from context.
func GetRole(c *fiber.Ctx) string {
	if role, ok := c.Locals(RoleKey).(string); ok {
		return role
	}
	return ""
}

// GetPermissions extracts the permissions from context.
func GetPermissions(c *fiber.Ctx) []string {
	if perms, ok := c.Locals(PermissionsKey).([]string); ok {
		return perms
	}
	return nil
}
