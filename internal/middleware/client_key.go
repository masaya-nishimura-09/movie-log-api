package middleware

import (
	"crypto/subtle"
	"net"

	"github.com/gin-gonic/gin"
)

func ClientKey(internalSecret string) func(c *gin.Context) string {
	return func(c *gin.Context) string {
		if internalSecret == "" {
			return c.ClientIP()
		}
		got := c.GetHeader("X-Internal-Secret")
		if subtle.ConstantTimeCompare([]byte(got), []byte(internalSecret)) != 1 {
			return c.ClientIP()
		}
		ip := net.ParseIP(c.GetHeader("X-Real-IP"))
		if ip == nil {
			return c.ClientIP()
		}
		return ip.String()
	}
}
