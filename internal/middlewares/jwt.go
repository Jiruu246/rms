package middlewares

import (
	"strings"

	"github.com/Jiruu246/rms/pkg/utils"
	"github.com/gin-gonic/gin"
)

// JWTMiddleware is a middleware for validating JWT tokens
func JWTMiddleware(secretKey []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.WriteUnauthorized(c.Writer, "Authorization header is required")
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := utils.ValidateJWT(secretKey, tokenString)
		if err != nil {
			utils.WriteUnauthorized(c.Writer, "Invalid token")
			c.Abort()
			return
		}

		c.Set("claims", claims)
		c.Next()
	}
}
