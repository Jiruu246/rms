package middlewares

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/Jiruu246/rms/pkg/utils"
	"github.com/gin-gonic/gin"
)

// CORSConfig holds the configuration for CORS middleware
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns a default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodHead,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Authorization",
			"Accept",
			"X-Requested-With",
			"Cache-Control",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"Content-Type",
		},
		AllowCredentials: false,
		MaxAge:           86400, // 24 hours
	}
}

// CORS returns a CORS middleware with default configuration
func CORS() gin.HandlerFunc {
	return CORSWithConfig(DefaultCORSConfig())
}

// CORSWithConfig returns a CORS middleware with custom configuration
func CORSWithConfig(config CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// Set allowed origins
		if len(config.AllowOrigins) > 0 {
			if slices.Contains(config.AllowOrigins, "*") {
				c.Header("Access-Control-Allow-Origin", "*")
			} else if slices.Contains(config.AllowOrigins, origin) {
				c.Header("Access-Control-Allow-Origin", origin)
			}
		}

		// Set allowed methods
		if len(config.AllowMethods) > 0 {
			c.Header("Access-Control-Allow-Methods", utils.JoinStrings(config.AllowMethods, ", "))
		}

		// Set allowed headers
		if len(config.AllowHeaders) > 0 {
			c.Header("Access-Control-Allow-Headers", utils.JoinStrings(config.AllowHeaders, ", "))
		}

		// Set exposed headers
		if len(config.ExposeHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", utils.JoinStrings(config.ExposeHeaders, ", "))
		}

		// Set allow credentials
		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// Set max age for preflight requests
		if config.MaxAge > 0 {
			c.Header("Access-Control-Max-Age", fmt.Sprintf("%d", config.MaxAge))
		}

		// Handle preflight OPTIONS request
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RestrictiveCORS returns a more restrictive CORS middleware for production
func RestrictiveCORS(allowedOrigins []string) gin.HandlerFunc {
	config := CORSConfig{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Authorization",
			"Accept",
			"X-Requested-With",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"Content-Type",
		},
		AllowCredentials: true,
		MaxAge:           3600, // 1 hour
	}
	return CORSWithConfig(config)
}
