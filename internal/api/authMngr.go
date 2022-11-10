package api

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

var (
	ErrUserIDNotFound       = errors.New("user id not found")
	ErrUnexpectedUserIDType = errors.New("unexpected user id type")
)

type authMngr struct {
	jwtMngr *jwtMngr
}

func newAuthMngr() *authMngr {
	return &authMngr{jwtMngr: newJwtMngr("", time.Second*0, time.Second*0)}
}

func (a *authMngr) newAccessAndRefreshTokens(id int64) (accessToken string, refreshToken string, refreshExpiresIn int64, err error) {
	log.Debug().Str("id", fmt.Sprint(id)).Msg("authMngr.newAccessAndRefreshTokens START")
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("authMngr.newAccessAndRefreshTokens END")
		} else {
			log.Debug().Msg("authMngr.newAccessAndRefreshTokens END")
		}
	}()

	accessToken, err = a.jwtMngr.newAccessToken(id)
	if err != nil {
		return "", "", 0, err
	}
	refreshToken = a.jwtMngr.newRefreshToken()

	refreshExpiresIn = time.Now().Add(a.jwtMngr.refreshTokenTTL).Unix()

	return accessToken, refreshToken, refreshExpiresIn, nil
}

func (a *authMngr) getIDFromAuthHeader(c *gin.Context) (int64, error) {
	authHeader := c.GetHeader("Authorization")
	if len(authHeader) == 0 {
		return 0, ErrEmptyAuthHeader
	}

	headerParts := strings.Split(authHeader, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return 0, ErrInvalidAuthHeader
	}

	accessToken := headerParts[1]
	id, err := a.jwtMngr.getID(accessToken)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (a *authMngr) getID(c *gin.Context) (int64, error) {
	id, ok := c.Get("id")
	if !ok {
		return 0, ErrUserIDNotFound
	}

	var userID int64
	if userID, ok = id.(int64); !ok {
		return 0, ErrUnexpectedUserIDType
	}
	return userID, nil
}

func (a *authMngr) setID(c *gin.Context, id int64) {
	c.Set("id", id)
}
