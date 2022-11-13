package app

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"practicum-gophermart/internal/model"
	dberr "practicum-gophermart/internal/storage/errors"
)

var (
	ErrOrderWasUploadedByCurrentUser = errors.New("the order was uploaded by current user")
	ErrOrderWasUploadedByAnotherUser = errors.New("the order was uploaded by another user")
)

func (a *app) AddOrder(c *gin.Context, order *model.Order) error {
	log.Debug().Str("order_number", order.Number).Msg("app.AddOrder START")
	var err error
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("app.AddOrder END")
		} else {
			log.Debug().Msg("app.AddOrder END")
		}
	}()

	err = a.storage.AddOrder(c, order)
	if err != nil {
		if errors.Is(err, dberr.ErrOrderWasUploadedByCurrentUser) {
			return ErrOrderWasUploadedByCurrentUser
		} else if errors.Is(err, dberr.ErrOrderWasUploadedByAnotherUser) {
			return ErrOrderWasUploadedByAnotherUser
		}
	}

	return nil
}

func (a *app) GetOrdersByUser(c *gin.Context, userID int64) ([]model.Order, error) {
	log.Debug().Str("userID", fmt.Sprint(userID)).Msg("app.GetOrdersByUser START")
	var err error
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("app.GetOrdersByUser END")
		} else {
			log.Debug().Msg("app.GetOrdersByUser END")
		}
	}()

	orders, err := a.storage.GetOrdersByUser(c, userID)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (a *app) GetOrderNumbersByStatuses(statuses []string) ([]string, error) {
	log.Debug().Msg("app.GetOrderNumbersByStatuses START")
	var err error
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("app.GetOrderNumbersByStatuses END")
		} else {
			log.Debug().Msg("app.GetOrderNumbersByStatuses END")
		}
	}()

	orderNumbers, err := a.storage.GetOrderNumbersByStatuses(statuses)
	if err != nil {
		return nil, err
	}

	return orderNumbers, nil
}

func (a *app) UpdateOrderStatuses(newOrderStatuses []model.Order) error {
	log.Debug().Msg("app.UpdateOrderStatuses START")
	var err error
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("app.UpdateOrderStatuses END")
		} else {
			log.Debug().Msg("app.UpdateOrderStatuses END")
		}
	}()
	if err = a.storage.UpdateOrderStatuses(newOrderStatuses); err != nil {
		return err
	}

	return nil
}
