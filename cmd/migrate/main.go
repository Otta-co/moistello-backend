package main

import (
	"github.com/moistello/backend/config"
	"github.com/moistello/backend/pkg/logger"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg, err := config.Load(".")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}
	logger.Init(cfg.Logging.Level, cfg.Logging.Format)
	log.Info().Msg("migrations applied")
}
