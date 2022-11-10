package api

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"practicum-gophermart/internal/app"
	"practicum-gophermart/internal/model"
)

var ErrUserAlreadyExists = errors.New("user already exists")

var (
	ErrEmptyAuthHeader   = errors.New("empty auth header")
	ErrInvalidAuthHeader = errors.New("invalid auth header")
)

func (a *API) signUpHandler(c *gin.Context) {
	log.Debug().Msg("api.signUp START")
	defer log.Debug().Msg("api.signUp START")

	var requestUser model.User

	if err := c.BindJSON(&requestUser); err != nil {
		a.error(c, http.StatusBadRequest, err)
		return
	}

	id, err := a.app.CreateUser(c, &requestUser)
	if err != nil {
		if errors.Is(err, app.ErrUserAlreadyExists) {
			a.error(c, http.StatusConflict, ErrUserAlreadyExists)
		} else {
			a.error(c, http.StatusBadRequest, err)
		}
		return
	}

	accessToken, refreshToken, refreshExpiresIn, err := a.authMngr.newAccessAndRefreshTokens(id)
	if err != nil {
		a.error(c, http.StatusInternalServerError, err)
		return
	}

	newRefreshSession := model.RefreshSession{UserID: id, Token: refreshToken, ExpiresIn: refreshExpiresIn}
	err = a.app.NewRefreshSession(c, &newRefreshSession)
	if err != nil {
		a.error(c, http.StatusInternalServerError, err)
		return
	}

	c.Header("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	c.SetCookie("refreshToken", newRefreshSession.Token, int(newRefreshSession.ExpiresIn), "/api", "", true, true)

	a.respond(c, http.StatusCreated, map[string]string{"accessToken": accessToken, "refreshToken": newRefreshSession.Token})
}

func (a *API) checkAuthMiddleware(c *gin.Context) {

	id, err := a.authMngr.getIDFromAuthHeader(c)
	if err != nil {
		if errors.Is(err, ErrAccessTokenIsExpired) {

			refreshToken, err := c.Cookie("refreshToken")
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]interface{}{"access token error": ErrAccessTokenIsExpired.Error()})
				return
			}

			refreshSession, err := a.app.GetRefreshSessionByToken(c, refreshToken)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]interface{}{"access token error": ErrAccessTokenIsExpired.Error()})
				return
			}

			if refreshTokenIsExpired := refreshSession.ExpiresIn < time.Now().Unix(); refreshTokenIsExpired {
				c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]interface{}{
					"access token error":  ErrAccessTokenIsExpired.Error(),
					"refresh token error": ErrRefreshTokenIsExpired.Error(),
				})
				return
			}

			newAccessToken, newRefreshToken, newRefreshExpiresIn, err := a.authMngr.newAccessAndRefreshTokens(refreshSession.UserID)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]interface{}{"access token error": ErrAccessTokenIsExpired.Error()})
				return
			}

			newRefreshSession := model.RefreshSession{UserID: refreshSession.UserID, Token: newRefreshToken, ExpiresIn: newRefreshExpiresIn}
			err = a.app.NewRefreshSession(c, &newRefreshSession)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]interface{}{"access token error": ErrAccessTokenIsExpired.Error()})
				return
			}

			c.Header("Authorization", fmt.Sprintf("Bearer %s", newAccessToken))
			c.SetCookie("refreshToken", newRefreshToken, int(newRefreshExpiresIn), "/api", "", true, true)

		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, err.Error())
			return
		}
	}

	a.authMngr.setID(c, id)
}

func (a *API) signInHandler(c *gin.Context) {
	log.Debug().Msg("api.signIn started")
	defer log.Debug().Msg("api.signIn ended")

	var requestUser model.User
	if err := c.BindJSON(&requestUser); err != nil {
		a.error(c, http.StatusBadRequest, err)
		return
	}

	user, err := a.app.GetUser(c, requestUser.Login, requestUser.Password)
	if err != nil {
		if errors.Is(err, app.ErrInvalidLoginOrPassword) {
			a.error(c, http.StatusUnauthorized, err)
		} else {
			a.error(c, http.StatusBadRequest, err)
		}
		return
	}

	accessToken, refreshToken, refreshExpiresIn, err := a.authMngr.newAccessAndRefreshTokens(user.ID)
	if err != nil {
		a.error(c, http.StatusInternalServerError, err)
		return
	}

	newRefreshSession := model.RefreshSession{UserID: user.ID, Token: refreshToken, ExpiresIn: refreshExpiresIn}
	err = a.app.NewRefreshSession(c, &newRefreshSession)
	if err != nil {
		a.error(c, http.StatusInternalServerError, err)
		return
	}

	c.Header("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	c.SetCookie("refreshToken", newRefreshSession.Token, int(newRefreshSession.ExpiresIn), "/api", "", true, true)

	a.respond(c, http.StatusOK, map[string]string{"accessToken": accessToken, "refreshToken": newRefreshSession.Token})
}
