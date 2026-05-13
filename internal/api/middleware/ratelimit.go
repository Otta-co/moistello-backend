package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/moistello/backend/config"
)

func RateLimitMiddleware(redisClient *redis.Client, cfg config.RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "ratelimit:" + c.ClientIP()
		limit := cfg.Global
		if _, exists := c.Get("userID"); exists {
			key = "ratelimit:user:" + GetUserID(c)
			limit = cfg.Authenticated
		}
		reqCtx := c.Request.Context()
		current, err := redisClient.Get(reqCtx, key).Int()
		if err != nil && err != redis.Nil {
			c.Next()
			return
		}
		if current >= limit {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"success": false, "error": "rate limit exceeded"})
			return
		}
		pipe := redisClient.Pipeline()
		pipe.Incr(reqCtx, key)
		pipe.Expire(reqCtx, key, 1*time.Minute)
		if _, err := pipe.Exec(reqCtx); err != nil {
			log.Warn().Err(err).Msg("rate limit pipeline failed")
		}
		c.Next()
	}
}

func AuthRateLimitMiddleware(redisClient *redis.Client, cfg config.RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "ratelimit:auth:" + c.ClientIP()
		current, err := redisClient.Get(c.Request.Context(), key).Int()
		if err != nil && err != redis.Nil {
			c.Next()
			return
		}
		if current >= cfg.Auth {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"success": false, "error": "too many auth attempts"})
			return
		}
		pipe := redisClient.Pipeline()
		pipe.Incr(c.Request.Context(), key)
		pipe.Expire(c.Request.Context(), key, 1*time.Minute)
		if _, err := pipe.Exec(c.Request.Context()); err != nil {
			log.Warn().Err(err).Msg("rate limit pipeline failed")
		}
		c.Next()
	}
}
