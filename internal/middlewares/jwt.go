package middlewares

import (
	"net/http"

	"github.com/Jiruu246/rms/pkg/utils"
	"github.com/gin-gonic/gin"
)

// JWTMiddleware is a middleware for validating JWT tokens
func JWTMiddleware(secretKey []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		// Validate the token
		claims, err := utils.ValidateJWT(secretKey, tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Set claims to context
		c.Set("claims", claims)
		c.Next()
	}
}
