package middleware

import (
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"

	"go-boilerplate/pkg/jwt"
)

// Global idempotency runs before route authentication. Validate the token before replaying a private response.
func idempotencyScope(userID uint, authorization string, body []byte, services []jwt.Service) (string, bool) {
	if authorization != "" && len(services) > 0 {
		parts := strings.Fields(authorization)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return "", false
		}
		claims, err := services[0].ValidateToken(parts[1])
		if err != nil {
			return "", false
		}
		userID = claims.UserID
	}
	if userID > 0 {
		return strconv.FormatUint(uint64(userID), 10), true
	}
	return fmt.Sprintf("anonymous:%x", sha256.Sum256(body)), true
}
