package jwt_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"go-boilerplate/config"
	"go-boilerplate/pkg/jwt"
)

func TestNewFromConfig(t *testing.T) {
	for _, algorithm := range []string{"", "hs256", "rs256", "es256"} {
		t.Run(algorithm, func(t *testing.T) {
			cfg := config.JWT{Algorithm: algorithm, SecretKey: "configured-test", AccessExpiry: time.Minute, RefreshExpiry: time.Hour}
			if algorithm == "rs256" {
				cfg.PrivateKeyPath, cfg.PublicKeyPath = generateRSAKeys(t)
			}
			if algorithm == "es256" {
				cfg.PrivateKeyPath, cfg.PublicKeyPath = generateES256Keys(t)
			}
			svc, err := jwt.NewFromConfig(cfg)
			require.NoError(t, err)
			token, _, err := svc.GenerateAccessToken(7, "test@example.com", "user", nil)
			require.NoError(t, err)
			claims, err := svc.ValidateToken(token)
			require.NoError(t, err)
			require.Equal(t, uint(7), claims.UserID)
		})
	}
	for _, algorithm := range []string{"unknown", "rs256", "es256"} {
		_, err := jwt.NewFromConfig(config.JWT{Algorithm: algorithm})
		require.Error(t, err)
	}
}
