package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"

	"practicum-gophermart/internal/model"
)

type accrualMngr struct {
	client *resty.Client
}

func newAccrualMngr() *accrualMngr {
	client := resty.New()
	client.AddRetryCondition(
		func(r *resty.Response, err error) bool {
			return r.StatusCode() == http.StatusTooManyRequests
		},
	).SetRetryWaitTime(time.Second * 60)
	return &accrualMngr{client: client}
}

func (a *API) updateOrdersStatus() error {
	log.Debug().Msg("api.updateOrdersStatus START")
	var err error
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("api.updateOrdersStatus END")
		} else {
			log.Debug().Msg("api.updateOrdersStatus END")
		}
	}()

	nonFinalStatuses := []string{model.OrderStatusNew.String(), model.OrderStatusProcessing.String()}
	ordersWithNonFinalStatuses, err := a.app.GetOrdersByStatuses(nonFinalStatuses)
	if err != nil {
		return err
	}

	type orderFromAccrualSystem struct {
		Order   string  `json:"order"`
		Status  string  `json:"status"`
		Accrual float64 `json:"accrual"`
	}

	orderStatusesFromAccrualSystem := make([]model.Order, 0, len(ordersWithNonFinalStatuses))

	for _, order := range ordersWithNonFinalStatuses {
		resp, err := a.accrualMngr.client.R().SetPathParam("number", order.Number).Get(a.app.Config().AccrualGetOrder())
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK {
			continue
		}

		newOrderFromAccrualSystem := orderFromAccrualSystem{}

		err = json.Unmarshal(resp.Body(), &newOrderFromAccrualSystem)
		if err != nil {
			return err
		}

		if newOrderFromAccrualSystem.Status == "REGISTERED" || newOrderFromAccrualSystem.Status == "PROCESSING" {
			continue
		}

		if newOrderFromAccrualSystem.Status == "INVALID" {
			newOrderFromAccrualSystem.Accrual = 0.0
		}

		orderStatusesFromAccrualSystem = append(orderStatusesFromAccrualSystem,
			model.Order{
				UserID:  order.UserID,
				Number:  newOrderFromAccrualSystem.Order,
				Status:  newOrderFromAccrualSystem.Status,
				Accrual: newOrderFromAccrualSystem.Accrual,
			})

	}

	if err = a.app.UpdateOrderStatuses(orderStatusesFromAccrualSystem); err != nil {
		return err
	}

	return nil
}
