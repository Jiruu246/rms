package config

import (
	"net/http"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestNewCookieConfig_DevelopmentDefaultsToLax(t *testing.T) {
	configurator := viper.New()
	configurator.Set("COOKIE_DOMAIN", "localhost")

	cookieConfig := NewCookieConfig(configurator)

	require.False(t, cookieConfig.Secure)
	require.True(t, cookieConfig.HttpOnly)
	require.Equal(t, http.SameSiteLaxMode, cookieConfig.SameSite)
	require.Equal(t, "localhost", cookieConfig.Domain)
}

func TestNewCookieConfig_UsesEnvSettings(t *testing.T) {
	configurator := viper.New()
	configurator.Set("COOKIE_SECURE", true)
	configurator.Set("COOKIE_SAMESITE", "none")
	configurator.Set("COOKIE_DOMAIN", ".example.com")

	cookieConfig := NewCookieConfig(configurator)

	require.True(t, cookieConfig.Secure)
	require.True(t, cookieConfig.HttpOnly)
	require.Equal(t, http.SameSiteNoneMode, cookieConfig.SameSite)
	require.Equal(t, ".example.com", cookieConfig.Domain)
}
