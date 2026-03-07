package config

import (
	"net/http"

	"github.com/spf13/viper"
)

type Environment string

const (
	EnvDevelopment Environment = "development"
	EnvProduction  Environment = "production"
)

type CookieConfig struct {
	Secure   bool
	HttpOnly bool
	SameSite http.SameSite
	Domain   string
}

func NewCookieConfig(Configurator *viper.Viper) CookieConfig {
	env := Environment(Configurator.GetString("ENV"))
	domain := Configurator.GetString("COOKIE_DOMAIN")

	return CookieConfig{
		Secure:   env == EnvProduction,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Domain:   domain,
	}
}
