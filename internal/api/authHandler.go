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

var errUserAlreadyExists = errors.New("user already exists")

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
			a.error(c, http.StatusConflict, errUserAlreadyExists)
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
	c.SetCookie("refreshToken", newRefreshSession.Token, int(newRefreshSession.ExpiresIn.Unix()), "/api", "", true, true)

	a.respond(c, http.StatusOK, map[string]string{"accessToken": accessToken, "refreshToken": newRefreshSession.Token})
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
	c.SetCookie("refreshToken", newRefreshSession.Token, int(newRefreshSession.ExpiresIn.Unix()), "/api", "", true, true)

	a.respond(c, http.StatusOK, map[string]string{"accessToken": accessToken, "refreshToken": newRefreshSession.Token})
}

func (a *API) checkAuthMiddleware(c *gin.Context) {
	log.Debug().Msg("api.checkAuthMiddleware started")
	defer log.Debug().Msg("api.checkAuthMiddleware ended")

	id, err := a.authMngr.getIDFromAuthHeader(c)

	if err == nil {
		a.authMngr.setID(c, id)
		return
	}

	accessIsAccessTokenIsExpired := errors.Is(err, errAccessTokenIsExpired)
	if !accessIsAccessTokenIsExpired {
		c.AbortWithStatusJSON(http.StatusUnauthorized, err.Error())
		return
	}

	refreshToken, errGettingRefreshToken := c.Cookie("refreshToken")
	if errGettingRefreshToken != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]interface{}{
			"access token error":          errAccessTokenIsExpired.Error(),
			"getting refresh token error": errGettingRefreshToken.Error(),
		})
		return
	}

	refreshSession, errGettingRefreshSessionByToken := a.app.GetRefreshSessionByToken(c, refreshToken)
	if errGettingRefreshSessionByToken != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]interface{}{
			"access token error": errAccessTokenIsExpired.Error(),
		})
		log.Error().Err(errGettingRefreshSessionByToken).Msg("getting refresh session by token")
		return
	}

	if refreshTokenIsExpired := refreshSession.ExpiresIn.Before(time.Now()); refreshTokenIsExpired {
		c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]interface{}{
			"access token error":  errAccessTokenIsExpired.Error(),
			"refresh token error": errRefreshTokenIsExpired.Error(),
		})
		return
	}

	newAccessToken, newRefreshToken, newRefreshExpiresIn, errCreatingNewTokens := a.authMngr.newAccessAndRefreshTokens(refreshSession.UserID)
	if errCreatingNewTokens != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]interface{}{
			"access token error": errAccessTokenIsExpired.Error(),
		})
		return
	}

	newRefreshSession := model.RefreshSession{UserID: refreshSession.UserID, Token: newRefreshToken, ExpiresIn: newRefreshExpiresIn}
	errSavingRefreshSession := a.app.NewRefreshSession(c, &newRefreshSession)
	if errSavingRefreshSession != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]interface{}{
			"access token error": errAccessTokenIsExpired.Error(),
		})
		log.Error().Err(errSavingRefreshSession).Msg("getting refresh session by token")
		return
	}

	c.Header("Authorization", fmt.Sprintf("Bearer %s", newAccessToken))
	c.SetCookie("refreshToken", newRefreshToken, int(newRefreshExpiresIn.Unix()), "/api", "", true, true)

	a.authMngr.setID(c, id)
}
