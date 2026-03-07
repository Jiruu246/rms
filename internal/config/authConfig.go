package config

import (
	"time"

	"github.com/spf13/viper"
)

type AuthConfig struct {
	JwtSecret              []byte
	AccessTokenExpiration  time.Duration
	RefreshTokenExpiration time.Duration
}

func NewAuthConfig(configurator *viper.Viper) AuthConfig {
	jwtSecret := configurator.GetString("JWT_SECRET")
	accessExp := time.Duration(configurator.GetInt("ACCESS_TOKEN_EXPIRATION")) * time.Second
	refreshExp := time.Duration(configurator.GetInt("REFRESH_TOKEN_EXPIRATION")) * time.Second

	return AuthConfig{
		JwtSecret:              []byte(jwtSecret),
		AccessTokenExpiration:  accessExp,
		RefreshTokenExpiration: refreshExp,
	}
}
