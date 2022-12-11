package app

import (
	"errors"

	"golang.org/x/crypto/bcrypt"

	"practicum-gophermart/internal/config"
	"practicum-gophermart/internal/storage"

	"github.com/rs/zerolog/log"
)

var (
	ErrEmptyConfig  = errors.New("empty config")
	ErrEmptyStorage = errors.New("empty storage")
)

type App struct {
	storage storage.Storage
	cfg     *config.Config
	pwdMngr *pwdMngr
}

// New returns new App.
func New(thisStorage storage.Storage, cfg *config.Config) (newApp *App, err error) {
	log.Debug().Str("cfg", cfg.String()).Msg("app.New started")
	defer func() {
		logMethodEnd("app.New", err)
	}()

	if cfg == nil {
		return nil, ErrEmptyConfig
	}
	if thisStorage == nil {
		return nil, ErrEmptyStorage
	}

	newApp = &App{
		storage: thisStorage,
		cfg:     cfg,
		pwdMngr: newPwdMngr(bcrypt.DefaultCost),
	}

	return newApp, nil
}

func (a *App) Config() *config.Config {
	return a.cfg
}

func (a *App) CloseStorage() error {
	return a.storage.Close()
}

func logMethodEnd(method string, err error) {
	msg := method + " END"
	if err != nil {
		log.Error().Err(err).Msg(msg)
	} else {
		log.Debug().Msg(msg)
	}
}
