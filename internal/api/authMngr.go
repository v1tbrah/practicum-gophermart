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
	errUserIDNotFound       = errors.New("user id not found")
	errUnexpectedUserIDType = errors.New("unexpected user id type")
)

var (
	errEmptyAuthHeader   = errors.New("empty auth header")
	errInvalidAuthHeader = errors.New("invalid auth header")
)

type authMngr struct {
	jwtMngr *jwtMngr
}

func newAuthMngr() *authMngr {
	return &authMngr{jwtMngr: newJwtMngr("", time.Second*0, time.Second*0)}
}

func (a *authMngr) newAccessAndRefreshTokens(id int64) (accessToken, refreshToken string, refreshExpiresIn time.Time, err error) {
	log.Debug().Str("id", fmt.Sprint(id)).Msg("authMngr.newAccessAndRefreshTokens START")
	defer func() {
		logMethodEnd("authMngr.newAccessAndRefreshTokens", err)
	}()

	accessToken, err = a.jwtMngr.newAccessToken(id)
	if err != nil {
		return "", "", time.Time{}, err
	}
	refreshToken = a.jwtMngr.newRefreshToken()

	refreshExpiresIn = time.Now().Add(a.jwtMngr.refreshTokenTTL)

	return accessToken, refreshToken, refreshExpiresIn, nil
}

func (a *authMngr) getIDFromAuthHeader(c *gin.Context) (id int64, err error) {
	log.Debug().Msg("authMngr.getIDFromAuthHeader START")
	defer func() {
		logMethodEnd("authMngr.getIDFromAuthHeader", err)
	}()

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return 0, errEmptyAuthHeader
	}

	headerParts := strings.Split(authHeader, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return 0, errInvalidAuthHeader
	}

	accessToken := headerParts[1]
	id, err = a.jwtMngr.getID(accessToken)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (a *authMngr) getID(c *gin.Context) (userID int64, err error) {
	log.Debug().Msg("authMngr.getID START")
	defer func() {
		logMethodEnd("authMngr.getID", err)
	}()

	id, ok := c.Get("id")
	if !ok {
		return 0, errUserIDNotFound
	}

	if userID, ok = id.(int64); !ok {
		return 0, errUnexpectedUserIDType
	}
	return userID, nil
}

func (a *authMngr) setID(c *gin.Context, id int64) {
	log.Debug().Msg("authMngr.setID START")
	defer log.Debug().Msg("authMngr.setID END")

	c.Set("id", id)
}
