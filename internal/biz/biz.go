package biz

import (
	"demo/internal/conf"

	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(NewGreeterUsecase, NewAuthUseCase, ProvideJWTSecret)

// ProvideJWTSecret provides the JWT secret from configuration
func ProvideJWTSecret(c *conf.Auth) string {
	return c.JWTSecret
}
