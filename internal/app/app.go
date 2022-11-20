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
func New(storage storage.Storage, cfg *config.Config) (*App, error) {
	log.Debug().Str("cfg", cfg.String()).Msg("app.New started")
	var err error
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("app.New ended")
		} else {
			log.Debug().Msg("app.New ended")
		}
	}()

	if cfg == nil {
		err = ErrEmptyConfig
		return nil, err
	}
	if storage == nil {
		err = ErrEmptyStorage
		return nil, err
	}

	newApp := &App{
		storage: storage,
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
