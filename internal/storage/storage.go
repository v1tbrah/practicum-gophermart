package storage

import (
	"context"
	"errors"

	"practicum-gophermart/internal/config"
	"practicum-gophermart/internal/model"
	"practicum-gophermart/internal/storage/pg"

	"github.com/rs/zerolog/log"
)

var ErrEmptyConfig = errors.New("empty config")

type Storage interface {
	AddUser(c context.Context, user *model.User) (int64, error)
	GetUser(c context.Context, login string, password string) (*model.User, error)
	GetUserPassword(c context.Context, login string) (string, error)
	UpdateRefreshSession(c context.Context, newRefreshSession *model.RefreshSession) error
	GetRefreshSessionByToken(c context.Context, refreshToken string) (*model.RefreshSession, error)
	AddOrder(c context.Context, order *model.Order) error
	GetOrdersByUser(c context.Context, userID int64) ([]model.Order, error)
	GetOrdersByStatuses(statuses []string) ([]model.Order, error)
	UpdateOrderStatuses(newOrderStatuses []model.Order) error
	GetBalance(c context.Context, userID int64) (balance float64, withdrawn float64, err error)
	AddWithdrawal(c context.Context, userID int64, withdraw model.Withdraw) error
	GetWithdrawals(c context.Context, userID int64) ([]model.Withdraw, error)
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
