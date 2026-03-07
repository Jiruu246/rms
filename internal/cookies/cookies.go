// internal/cookies/cookies.go
package cookies

import (
	"net/http"
	"time"

	"github.com/Jiruu246/rms/internal/config"
)

const (
	RefreshTokenPath = "/api/auth"
)

type Factory struct {
	cfg config.CookieConfig
}

func NewFactory(cfg config.CookieConfig) *Factory {
	return &Factory{cfg: cfg}
}

// NewRefreshToken returns a pre-configured refresh token cookie
func (f *Factory) NewRefreshToken(token string, ttl time.Duration) *http.Cookie {
	return &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     RefreshTokenPath,
		Domain:   f.cfg.Domain,
		MaxAge:   int(ttl.Seconds()),
		Secure:   f.cfg.Secure,
		HttpOnly: f.cfg.HttpOnly,
		SameSite: f.cfg.SameSite,
	}
}

// ExpireRefreshToken returns a zeroed-out cookie to delete it
func (f *Factory) ExpireRefreshToken() *http.Cookie {
	return &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     RefreshTokenPath,
		Domain:   f.cfg.Domain,
		MaxAge:   -1,
		Secure:   f.cfg.Secure,
		HttpOnly: f.cfg.HttpOnly,
		SameSite: f.cfg.SameSite,
	}
}
