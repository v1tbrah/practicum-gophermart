package api

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"practicum-gophermart/internal/app"
	"practicum-gophermart/internal/model"
)

func (a *API) balanceHandler(c *gin.Context) {
	log.Debug().Msg("api.balanceHandler START")
	defer log.Debug().Msg("api.balanceHandler END")

	userID, err := a.authMngr.getID(c)
	if err != nil {
		a.error(c, http.StatusUnauthorized, err)
		return
	}

	balance, withdrawn, err := a.app.GetBalance(c, userID)
	if err != nil {
		a.error(c, http.StatusInternalServerError, err)
		return
	}

	resp := struct {
		Current   float64 `json:"current"`
		Withdrawn float64 `json:"withdrawn"`
	}{
		Current:   balance,
		Withdrawn: withdrawn,
	}

	a.respond(c, http.StatusOK, resp)
}

func (a *API) withdrawPointsHandler(c *gin.Context) {
	log.Debug().Msg("api.withdrawPointsHandler START")
	defer log.Debug().Msg("api.withdrawPointsHandler END")

	userID, err := a.authMngr.getID(c)
	if err != nil {
		a.error(c, http.StatusUnauthorized, err)
		return
	}

	reqWithdraw := model.Withdraw{ProcessedAt: time.Now()}
	if err = c.BindJSON(&reqWithdraw); err != nil {
		a.error(c, http.StatusBadRequest, err)
		return
	}

	orderNumber := reqWithdraw.Order
	if numberForCheckValid, err := strconv.Atoi(orderNumber); err != nil {
		a.error(c, http.StatusUnprocessableEntity, err)
		return
	} else if !isValidNumber(int64(numberForCheckValid)) {
		a.error(c, http.StatusUnprocessableEntity, errInvalidOrderNumber)
		return
	}

	err = a.app.WithdrawFromBalance(c, userID, reqWithdraw)
	if err != nil {
		if errors.Is(err, app.ErrInsufficientFunds) {
			a.error(c, http.StatusPaymentRequired, app.ErrInsufficientFunds)
		} else {
			a.error(c, http.StatusInternalServerError, err)
		}
		return
	}

	a.respond(c, http.StatusOK, nil)
}

func (a *API) withdrawnPointsHandler(c *gin.Context) {
	log.Debug().Msg("api.withdrawnPointsHandler START")
	defer log.Debug().Msg("api.withdrawnPointsHandler END")

	userID, err := a.authMngr.getID(c)
	if err != nil {
		a.error(c, http.StatusUnauthorized, err)
		return
	}

	withdrawals, err := a.app.GetWithdrawals(c, userID)
	if err != nil {
		a.error(c, http.StatusInternalServerError, err)
		return
	}
	if len(withdrawals) == 0 {
		a.respond(c, http.StatusNoContent, nil)
		return
	}

	a.respond(c, http.StatusOK, withdrawals)
}
