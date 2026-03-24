package middleware

import (
	"strconv"
	"strings"

	"github.com/2hot4you/aiaos/backend/internal/service"
	"github.com/2hot4you/aiaos/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

func Auth(authSvc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, 40102, "missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, 40102, "invalid authorization format")
			c.Abort()
			return
		}

		claims, err := authSvc.ValidateToken(parts[1])
		if err != nil {
			response.Unauthorized(c, 40102, "invalid or expired token")
			c.Abort()
			return
		}

		sub, _ := claims["sub"].(string)
		userID, _ := strconv.ParseInt(sub, 10, 64)
		role, _ := claims["role"].(string)
		username, _ := claims["username"].(string)

		c.Set("userID", userID)
		c.Set("role", role)
		c.Set("username", username)
		c.Next()
	}
}

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != "admin" {
			response.Forbidden(c, "admin role required")
			c.Abort()
			return
		}
		c.Next()
	}
}

func GetUserID(c *gin.Context) int64 {
	id, _ := c.Get("userID")
	if v, ok := id.(int64); ok {
		return v
	}
	return 0
}

func GetUserRole(c *gin.Context) string {
	role, _ := c.Get("role")
	if v, ok := role.(string); ok {
		return v
	}
	return ""
}
