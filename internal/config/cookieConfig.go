package config

import (
	"net/http"
	"strings"

	"github.com/spf13/viper"
)

type Environment string

// FIXMES: environment declares in here?
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
	domain := Configurator.GetString("COOKIE_DOMAIN")
	sameSite := parseSameSite(Configurator.GetString("COOKIE_SAMESITE"))

	return CookieConfig{
		Secure:   Configurator.GetBool("COOKIE_SECURE"),
		HttpOnly: true,
		SameSite: sameSite,
		Domain:   domain,
	}
}

func parseSameSite(value string) http.SameSite {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "none":
		return http.SameSiteNoneMode
	case "strict":
		return http.SameSiteStrictMode
	case "lax", "":
		return http.SameSiteLaxMode
	default:
		return http.SameSiteLaxMode
	}
}
