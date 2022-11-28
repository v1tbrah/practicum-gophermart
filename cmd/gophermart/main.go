package main

import (
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
	log.Info().Msg("gophermart START")

	setGlobalLogLevel(zerolog.DebugLevel)

	newCfg, err := config.New(config.WithFlag, config.WithEnv)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to create new config")
	}

	newStorage, err := storage.New(newCfg)
	if err != nil {
		log.Fatal().Err(err).Str("config", newCfg.String()).Msg("unable to create new storage")
	}

	newApp, err := app.New(newStorage, newCfg)
	if err != nil {
		log.Fatal().Err(err).Str("config", newCfg.String()).Msg("unable to create new app")
	}

	newAPI, err := api.New(newApp)
	if err != nil {
		log.Fatal().Err(err).Str("config", newCfg.String()).Msg("unable to create new api")
	}

	if err = newAPI.Run(); err != nil {
		log.Fatal().Err(err)
	}

	log.Info().Msg("gophermart END")
}
