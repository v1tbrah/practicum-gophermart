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
	AddUser(ctx context.Context, user *model.User) (int64, error)
	GetUser(ctx context.Context, login string, password string) (*model.User, error)
	GetUserPassword(ctx context.Context, login string) (string, error)
	UpdateRefreshSession(ctx context.Context, newRefreshSession *model.RefreshSession) error
	GetRefreshSessionByToken(ctx context.Context, refreshToken string) (*model.RefreshSession, error)
	AddOrder(ctx context.Context, order *model.Order) error
	GetOrdersByUser(ctx context.Context, userID int64) ([]model.Order, error)
	GetOrdersByStatuses(statuses []string) ([]model.Order, error)
	UpdateOrderStatuses(newOrderStatuses []model.Order) error
	GetBalance(ctx context.Context, userID int64) (balance float64, withdrawn float64, err error)
	AddWithdrawal(ctx context.Context, userID int64, withdraw model.Withdraw) error
	GetWithdrawals(ctx context.Context, userID int64) ([]model.Withdraw, error)
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
