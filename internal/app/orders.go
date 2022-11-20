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

func (a *App) AddOrder(c context.Context, order *model.Order) error {
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

func (a *App) GetOrdersByUser(c context.Context, userID int64) ([]model.Order, error) {
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

func (a *App) GetOrdersByStatuses(statuses []string) ([]model.Order, error) {
	log.Debug().Msg("app.GetOrdersByStatuses START")
	var err error
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("app.GetOrdersByStatuses END")
		} else {
			log.Debug().Msg("app.GetOrdersByStatuses END")
		}
	}()

	orders, err := a.storage.GetOrdersByStatuses(statuses)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (a *App) UpdateOrderStatuses(newOrderStatuses []model.Order) error {
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
