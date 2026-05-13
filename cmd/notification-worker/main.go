package main

import (
	"github.com/moistello/backend/config"
	"github.com/moistello/backend/pkg/logger"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg, _ := config.Load(".")
	logger.Init(cfg.Logging.Level, cfg.Logging.Format)
	log.Info().Msg("notification worker starting...")
	select {}
}
