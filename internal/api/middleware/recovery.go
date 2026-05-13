package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Error().Interface("panic", err).Str("path", c.Request.URL.Path).Msg("panic recovered")
				c.AbortWithStatusJSON(500, gin.H{"success": false, "error": "internal server error"})
			}
		}()
		c.Next()
	}
}
