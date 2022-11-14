package api

import (
	"github.com/gin-gonic/gin"

	"practicum-gophermart/internal/config"
	"practicum-gophermart/internal/model"
)

type Application interface {
	CreateUser(c *gin.Context, user *model.User) (int64, error)
	GetUser(c *gin.Context, login, pwd string) (*model.User, error)
	NewRefreshSession(c *gin.Context, newRefreshSession *model.RefreshSession) error
	GetRefreshSessionByToken(c *gin.Context, refreshToken string) (*model.RefreshSession, error)
	AddOrder(c *gin.Context, order *model.Order) error
	GetOrdersByUser(c *gin.Context, userID int64) ([]model.Order, error)
	GetOrderNumbersByStatuses(statuses []string) ([]string, error)
	UpdateOrderStatuses(newOrderStatuses []model.Order) error
	GetBalance(c *gin.Context, userID int64) (balance float64, withdrawn float64, err error)
	WithdrawFromBalance(c *gin.Context, userID int64, withdraw model.Withdraw) error
	GetWithdrawals(c *gin.Context, userID int64) ([]model.Withdraw, error)
	Config() *config.Config
	CloseStorage() error
}
