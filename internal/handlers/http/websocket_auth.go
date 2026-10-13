package httphandler

import (
	"encoding/base64"
	"strings"

	"go-boilerplate/pkg/jwt"
)

func websocketHeadersValid(version, key string) bool {
	decoded, err := base64.StdEncoding.DecodeString(key)
	return version == "13" && err == nil && len(decoded) == 16
}

func websocketClaims(service jwt.Service, header string) (*jwt.Claims, error) {
	parts := strings.Split(header, " ")
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return nil, jwt.ErrInvalidToken
	}
	claims, err := service.ValidateToken(parts[1])
	if err != nil {
		return nil, err
	}
	if claims.UserID == 0 || claims.ExpiresAt == nil {
		return nil, jwt.ErrInvalidToken
	}
	return claims, nil
}
