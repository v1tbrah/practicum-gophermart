package storage

import (
	"errors"

	"practicum-gophermart/internal/config"
	"practicum-gophermart/internal/model"
	"practicum-gophermart/internal/storage/pg"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

var ErrEmptyConfig = errors.New("empty config")

type Storage interface {
	AddUser(c *gin.Context, user *model.User) (int64, error)
	GetUser(c *gin.Context, login string, password string) (*model.User, error)
	GetUserPassword(c *gin.Context, login string) (string, error)
	UpdateRefreshSession(c *gin.Context, newRefreshSession *model.RefreshSession) error
	GetRefreshSessionByToken(c *gin.Context, refreshToken string) (*model.RefreshSession, error)
	AddOrder(c *gin.Context, order *model.Order) error
	GetOrdersByUser(c *gin.Context, userID int64) ([]model.Order, error)
	Close() error
}

func New(cfg *config.Config) (Storage, error) {
	log.Debug().Str("config", cfg.String()).Msg("storage.New started")
	var err error
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("storage.New ended")
		} else {
			log.Debug().Msg("storage.New ended")
		}
	}()

	if cfg == nil {
		return nil, ErrEmptyConfig
	}

	return pg.New(cfg.PgConnString())
}
