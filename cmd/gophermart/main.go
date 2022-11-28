package main

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"practicum-gophermart/internal/api"
	"practicum-gophermart/internal/app"
	"practicum-gophermart/internal/config"
	"practicum-gophermart/internal/storage"
)

func setGlobalLogLevel(lvl zerolog.Level) {
	zerolog.SetGlobalLevel(lvl)
}

func main() {
	gin.SetMode(gin.ReleaseMode)

	cfgOptions := []string{config.WithFlag, config.WithEnv}
	newCfg, err := config.New(cfgOptions...)
	if err != nil {
		log.Fatal().Err(err).Str("cfg options", cfgOptions[0]+", "+cfgOptions[1]).Msg("creating new config")
	}

	logLevel, err := zerolog.ParseLevel(newCfg.LogLevel())
	if err != nil {
		log.Fatal().Err(err).Strs("cfg options", cfgOptions).Msg("parsing log level")
	}
	zerolog.SetGlobalLevel(logLevel)

	newStorage, err := storage.New(newCfg)
	if err != nil {
		log.Fatal().Err(err).Str("config", newCfg.String()).Msg("creating new storage")
	}
	log.Info().Msg("storage created")

	newApp, err := app.New(newStorage, newCfg)
	if err != nil {
		log.Fatal().Err(err).Str("config", newCfg.String()).Msg("creating new app")
	}
	log.Info().Msg("app created")

	newAPI, err := api.New(newApp)
	if err != nil {
		log.Fatal().Err(err).Str("config", newCfg.String()).Msg("creating new API")
	}
	log.Info().Msg("API created")

	if err = newAPI.Run(); err != nil {
		log.Fatal().Err(err).Msg("running api")
	}

}
