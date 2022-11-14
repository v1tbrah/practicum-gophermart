package app

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"practicum-gophermart/internal/model"
	dberr "practicum-gophermart/internal/storage/errors"
)

var ErrInsufficientFunds = errors.New("there are not enough funds in the account")

func (a *app) GetBalance(c *gin.Context, userID int64) (float64, float64, error) {
	log.Debug().Msg("app.GetBalance START")
	var err error
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("app.GetBalance END")
		} else {
			log.Debug().Msg("app.GetBalance END")
		}
	}()

	balance, withdrawn, err := a.storage.GetBalance(c, userID)
	if err != nil {
		return -1, -1, err
	}

	return balance, withdrawn, nil
}

func (a *app) WithdrawFromBalance(c *gin.Context, userID int64, withdraw model.Withdraw) error {
	log.Debug().Msg("app.WithdrawFromBalance START")
	var err error
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("app.WithdrawFromBalance END")
		} else {
			log.Debug().Msg("app.WithdrawFromBalance END")
		}
	}()

	err = a.storage.AddWithdrawal(c, userID, withdraw)
	if err != nil {
		if errors.Is(err, dberr.ErrNegativeBalance) {
			return ErrInsufficientFunds
		}
		return err
	}

	return nil
}

func (a *app) GetWithdrawals(c *gin.Context, userID int64) ([]model.Withdraw, error) {
	log.Debug().Msg("app.GetWithdrawals START")
	var err error
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("app.GetWithdrawals END")
		} else {
			log.Debug().Msg("app.GetWithdrawals END")
		}
	}()

	withdrawals, err := a.storage.GetWithdrawals(c, userID)
	if err != nil {
		return nil, err
	}

	return withdrawals, nil
}
