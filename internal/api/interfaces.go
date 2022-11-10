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
	Config() *config.Config
	CloseStorage() error
}
