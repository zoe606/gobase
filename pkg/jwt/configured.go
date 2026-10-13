package jwt

import (
	"fmt"

	"go-boilerplate/config"
)

func NewFromConfig(cfg config.JWT) (Service, error) {
	switch cfg.Algorithm {
	case "rs256":
		return NewRS256(cfg.PrivateKeyPath, cfg.PublicKeyPath, cfg.AccessExpiry, cfg.RefreshExpiry)
	case "es256":
		return NewES256(cfg.PrivateKeyPath, cfg.PublicKeyPath, cfg.AccessExpiry, cfg.RefreshExpiry)
	case "hs256", "":
		return New(cfg.SecretKey, cfg.AccessExpiry, cfg.RefreshExpiry), nil
	default:
		return nil, fmt.Errorf("unsupported JWT algorithm %q", cfg.Algorithm)
	}
}
