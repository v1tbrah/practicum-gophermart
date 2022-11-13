package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"

	"practicum-gophermart/internal/model"
)

type accrualMngr struct {
	client *resty.Client
}

func newAccrualMngr() *accrualMngr {
	return &accrualMngr{client: resty.New()}
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
	numbersOfOrdersWithNonFinalStatuses, err := a.app.GetOrderNumbersByStatuses(nonFinalStatuses)
	if err != nil {
		return err
	}

	type orderFromAccrualSystem struct {
		Order   string  `json:"order"`
		Status  string  `json:"status"`
		Accrual float64 `json:"accrual"`
	}

	orderStatusesFromAccrualSystem := make([]model.Order, 0, len(numbersOfOrdersWithNonFinalStatuses))

	for _, orderNumber := range numbersOfOrdersWithNonFinalStatuses {
		resp, err := a.accrualMngr.client.R().SetPathParam("number", orderNumber).Get(a.app.Config().AccrualGetOrder())
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

		orderStatusesFromAccrualSystem = append(orderStatusesFromAccrualSystem,
			model.Order{
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
