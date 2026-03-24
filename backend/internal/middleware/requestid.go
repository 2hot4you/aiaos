package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-ID")
		if rid == "" {
			b := make([]byte, 16)
			_, _ = rand.Read(b)
			rid = hex.EncodeToString(b)
		}
		c.Set("requestID", rid)
		c.Header("X-Request-ID", rid)
		c.Next()
	}
}
