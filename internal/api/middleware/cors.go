package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/moistello/backend/config"
)

func CORSMiddleware(cfg config.CORSConfig) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     cfg.AllowedMethods,
		AllowHeaders:     cfg.AllowedHeaders,
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID", "X-API-Version"},
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           12 * time.Hour,
	})
}
