package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"

	"practicum-gophermart/internal/model"
	dberr "practicum-gophermart/internal/storage/errors"
)

var (
	ErrOrderWasUploadedByCurrentUser = errors.New("the order was uploaded by current user")
	ErrOrderWasUploadedByAnotherUser = errors.New("the order was uploaded by another user")
)

func (a *App) AddOrder(c context.Context, order *model.Order) (err error) {
	log.Debug().Str("order_number", order.Number).Msg("app.AddOrder START")
	defer func() {
		logMethodEnd("app.AddOrder", err)
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

func (a *App) GetOrdersByUser(c context.Context, userID int64) (orders []model.Order, err error) {
	log.Debug().Str("userID", fmt.Sprint(userID)).Msg("app.GetOrdersByUser START")
	defer func() {
		logMethodEnd("app.GetOrdersByUser", err)
	}()

	orders, err = a.storage.GetOrdersByUser(c, userID)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (a *App) GetOrdersByStatuses(statuses []string) (orders []model.Order, err error) {
	log.Debug().Msg("app.GetOrdersByStatuses START")
	defer func() {
		logMethodEnd("app.GetOrdersByStatuses", err)
	}()

	orders, err = a.storage.GetOrdersByStatuses(statuses)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (a *App) UpdateOrders(newOrderStatuses []model.Order) (err error) {
	log.Debug().Msg("app.UpdateOrderStatuses START")
	defer func() {
		logMethodEnd("app.UpdateOrderStatuses", err)
	}()

	if err = a.storage.UpdateOrderStatuses(newOrderStatuses); err != nil {
		return err
	}

	return nil
}
