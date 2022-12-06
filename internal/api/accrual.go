package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"practicum-gophermart/internal/model"
)

type accrualOrderStatus string

const (
	statusProcessedFromAccrual accrualOrderStatus = "PROCESSED"
	statusInvalidFromAccrual   accrualOrderStatus = "INVALID"
)

type orderFromAccrualSystem struct {
	Order   string             `json:"order"`
	Status  accrualOrderStatus `json:"status"`
	Accrual float64            `json:"accrual"`
}

func (s accrualOrderStatus) isFinal() bool {
	return s == statusProcessedFromAccrual || s == statusInvalidFromAccrual
}

func (s accrualOrderStatus) isInvalid() bool {
	return s == statusInvalidFromAccrual
}

func (a *API) updateOrdersStatus() (err error) {
	log.Debug().Msg("api.updateOrdersStatus START")
	defer func() {
		logMethodEnd("api.updateOrdersStatus", err)
	}()

	nonFinalStatuses := []string{model.OrderStatusNew.String(), model.OrderStatusProcessing.String()}
	ordersWithNonFinalStatuses, err := a.app.GetOrdersByStatuses(nonFinalStatuses)
	if err != nil {
		return err
	}

	orderStatusesFromAccrualSystem := make([]model.Order, 0, len(ordersWithNonFinalStatuses))

	accrulGetOrderURL := a.app.Config().AccrualGetOrder()
	for i := 0; i < len(ordersWithNonFinalStatuses); i++ {

		order := ordersWithNonFinalStatuses[i]

		resp, err := a.accrualMngr.R().SetPathParam("number", order.Number).Get(accrulGetOrderURL)

		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK {
			if resp.StatusCode() == http.StatusTooManyRequests {
				a.accrualMngr.SetRetryWaitTime(time.Second)
				time.Sleep(time.Second)
				i--
			}
			continue
		}

		newOrderFromAccrualSystem := orderFromAccrualSystem{}

		err = json.Unmarshal(resp.Body(), &newOrderFromAccrualSystem)
		if err != nil {
			return err
		}

		if !newOrderFromAccrualSystem.Status.isFinal() {
			continue
		}

		if newOrderFromAccrualSystem.Status.isInvalid() {
			newOrderFromAccrualSystem.Accrual = 0.0
		}

		orderStatusesFromAccrualSystem = append(orderStatusesFromAccrualSystem,
			model.Order{
				UserID:  order.UserID,
				Number:  newOrderFromAccrualSystem.Order,
				Status:  string(newOrderFromAccrualSystem.Status),
				Accrual: newOrderFromAccrualSystem.Accrual,
			})

	}

	if err = a.app.UpdateOrderStatuses(orderStatusesFromAccrualSystem); err != nil {
		return err
	}

	return nil
}
