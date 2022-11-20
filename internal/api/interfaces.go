package api

import (
	"context"

	"practicum-gophermart/internal/config"
	"practicum-gophermart/internal/model"
)

type Application interface {
	CreateUser(c context.Context, user *model.User) (int64, error)
	GetUser(c context.Context, login, pwd string) (*model.User, error)
	NewRefreshSession(c context.Context, newRefreshSession *model.RefreshSession) error
	GetRefreshSessionByToken(c context.Context, refreshToken string) (*model.RefreshSession, error)
	AddOrder(c context.Context, order *model.Order) error
	GetOrdersByUser(c context.Context, userID int64) ([]model.Order, error)
	GetOrdersByStatuses(statuses []string) ([]model.Order, error)
	UpdateOrderStatuses(newOrderStatuses []model.Order) error
	GetBalance(c context.Context, userID int64) (balance float64, withdrawn float64, err error)
	WithdrawFromBalance(c context.Context, userID int64, withdraw model.Withdraw) error
	GetWithdrawals(c context.Context, userID int64) ([]model.Withdraw, error)
	Config() *config.Config
	CloseStorage() error
}
