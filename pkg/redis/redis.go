package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/moistello/backend/config"
)

func New(cfg config.RedisConfig) (*redis.Client, error) {
	opts, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parsing redis URL: %w", err)
	}

	if cfg.Password != "" {
		opts.Password = cfg.Password
	}
	opts.DB = cfg.DB
	opts.PoolSize = cfg.PoolSize

	client := redis.NewClient(opts)

	// context.Background() is correct here: this is a startup-time health check
	// not tied to any request scope.
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("connecting to redis: %w", err)
	}

	log.Info().Msg("connected to Redis")
	return client, nil
}

func Exists(ctx context.Context, client *redis.Client, key string) (bool, error) {
	n, err := client.Exists(ctx, key).Result()
	return n > 0, err
}
