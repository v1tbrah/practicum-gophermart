package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"

	"practicum-gophermart/internal/model"
)

type accrualSystemOrderStatus string

const (
	accrualSystemStatusProcessed accrualSystemOrderStatus = "PROCESSED"
	accrualSystemStatusInvalid   accrualSystemOrderStatus = "INVALID"

	accrualSystemMethodGetOrderParamNumber = "number"

	accrualRetryAfterHeader = "Retry - After"
)

type accrualSystemOrder struct {
	Order   string                   `json:"order"`
	Status  accrualSystemOrderStatus `json:"status"`
	Accrual float64                  `json:"accrual"`
}

func (s accrualSystemOrderStatus) isFinal() bool {
	return s == accrualSystemStatusProcessed || s == accrualSystemStatusInvalid
}

func (s accrualSystemOrderStatus) isInvalid() bool {
	return s == accrualSystemStatusInvalid
}

func (a *API) updateOrdersStatus() (err error) {
	log.Debug().Msg("api.updateOrdersStatus START")
	defer func() {
		logMethodEnd("api.updateOrdersStatus", err)
	}()

	nonFinalStatuses := []string{model.OrderStatusNew.String(), model.OrderStatusProcessing.String()}
	ordersWithNonFinalStatuses, err := a.app.GetOrdersByStatuses(nonFinalStatuses)
	if err != nil {
		return fmt.Errorf("getting orders with non final statuses : %w", err)
	}

	ordersFromAccrualSystem := make([]model.Order, 0, len(ordersWithNonFinalStatuses))
	accrualSystemGetOrderURL := a.app.Config().AccrualGetOrder()
	for i := 0; i < len(ordersWithNonFinalStatuses); i++ {

		order := ordersWithNonFinalStatuses[i]

		resp, err := a.accrualMngr.R().SetPathParam(accrualSystemMethodGetOrderParamNumber, order.Number).Get(accrualSystemGetOrderURL)
		if err != nil {
			return fmt.Errorf("getting order with number %s from accrual system: %w", order.Number, err)
		}

		if resp.StatusCode() != http.StatusOK {
			if needRetry := resp.StatusCode() == http.StatusTooManyRequests; needRetry {
				retryTime := time.Second * 60

				retryTimeFromAccrualStr := resp.Header().Get(accrualRetryAfterHeader)
				retryTimeFromAccrualInt, errParsing := strconv.Atoi(retryTimeFromAccrualStr)
				if errParsing == nil {
					retryTime = time.Second * time.Duration(retryTimeFromAccrualInt)
				} else {
					log.Error().Err(errParsing).Msg("parsing retry time from accrual system")
				}

				time.Sleep(retryTime)
				i--
			}

			continue
		}

		newOrderFromAccrualSystem := accrualSystemOrder{}

		err = json.Unmarshal(resp.Body(), &newOrderFromAccrualSystem)
		if err != nil {
			return fmt.Errorf("unmarshalling response body from accrual system: %w", err)
		}

		if !newOrderFromAccrualSystem.Status.isFinal() {
			continue
		}

		if newOrderFromAccrualSystem.Status.isInvalid() {
			newOrderFromAccrualSystem.Accrual = 0.0
		}

		ordersFromAccrualSystem = append(ordersFromAccrualSystem,
			model.Order{
				UserID:  order.UserID,
				Number:  newOrderFromAccrualSystem.Order,
				Status:  string(newOrderFromAccrualSystem.Status),
				Accrual: newOrderFromAccrualSystem.Accrual,
			})

	}

	if err = a.app.UpdateOrders(ordersFromAccrualSystem); err != nil {
		return fmt.Errorf("updating orders: %w", err)
	}

	return nil
}
