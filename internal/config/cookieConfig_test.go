package config

import (
	"net/http"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestNewCookieConfig_DevelopmentDefaultsToLax(t *testing.T) {
	configurator := viper.New()

	cookieConfig := NewCookieConfig(configurator)

	require.False(t, cookieConfig.Secure)
	require.True(t, cookieConfig.HttpOnly)
	require.Equal(t, http.SameSiteLaxMode, cookieConfig.SameSite)
}

func TestNewCookieConfig_UsesEnvSettings(t *testing.T) {
	configurator := viper.New()
	configurator.Set("COOKIE_SECURE", true)
	configurator.Set("COOKIE_SAMESITE", "none")

	cookieConfig := NewCookieConfig(configurator)

	require.True(t, cookieConfig.Secure)
	require.True(t, cookieConfig.HttpOnly)
	require.Equal(t, http.SameSiteNoneMode, cookieConfig.SameSite)
}
