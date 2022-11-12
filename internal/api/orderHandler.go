package api

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"practicum-gophermart/internal/app"
	"practicum-gophermart/internal/model"
)

var (
	ErrInvalidOrderNumber            = errors.New("invalid order number")
	ErrOrderWasUploadedByCurrentUser = errors.New("the order was uploaded by current user")
	ErrOrderWasUploadedByAnotherUser = errors.New("the order was uploaded by another user")
)

func (a *API) setOrderHandler(c *gin.Context) {
	log.Debug().Msg("api.setOrderHandler START")
	defer log.Debug().Msg("api.setOrderHandler END")

	userID, err := a.authMngr.getID(c)
	if err != nil {
		a.error(c, http.StatusUnauthorized, err)
		return
	}

	reqBody, err := io.ReadAll(c.Request.Body)
	if err != nil && err != io.EOF {
		a.error(c, http.StatusBadRequest, err)
		return
	}

	orderNumber := string(reqBody)
	if numberForCheckValid, err := strconv.Atoi(orderNumber); err != nil {
		a.error(c, http.StatusUnprocessableEntity, err)
		return
	} else if !isValidNumber(int64(numberForCheckValid)) {
		a.error(c, http.StatusUnprocessableEntity, ErrInvalidOrderNumber)
		return
	}

	order := model.Order{
		UserID:     userID,
		Number:     orderNumber,
		Status:     model.OrderStatusNew.String(),
		UploadedAt: time.Now(),
	}

	if err = a.app.AddOrder(c, &order); err != nil {
		if errors.Is(err, app.ErrOrderWasUploadedByAnotherUser) {
			a.error(c, http.StatusConflict, ErrOrderWasUploadedByAnotherUser)
		} else if errors.Is(err, app.ErrOrderWasUploadedByCurrentUser) {
			a.respond(c, http.StatusOK, ErrOrderWasUploadedByCurrentUser)
		} else {
			a.error(c, http.StatusInternalServerError, nil)
		}
		return
	}

	a.respond(c, http.StatusAccepted, nil)
}

func (a *API) ordersHandler(c *gin.Context) {
	log.Debug().Msg("api.ordersHandler START")
	defer log.Debug().Msg("api.ordersHandler END")

	userID, err := a.authMngr.getID(c)
	if err != nil {
		a.error(c, http.StatusUnauthorized, err)
		return
	}

	orders, err := a.app.GetOrdersByUser(c, userID)
	if err != nil {
		a.error(c, http.StatusInternalServerError, err)
		return
	}
	if len(orders) == 0 {
		a.respond(c, http.StatusNoContent, nil)
		return
	}

	a.respond(c, http.StatusOK, orders)
}

func isValidNumber(number int64) bool {
	return (number%10+checksum(number/10))%10 == 0
}

func checksum(number int64) int64 {
	var luhn int64

	for i := 0; number > 0; i++ {
		cur := number % 10

		if i%2 == 0 {
			cur = cur * 2
			if cur > 9 {
				cur = cur%10 + cur/10
			}
		}

		luhn += cur
		number = number / 10
	}
	return luhn % 10
}
