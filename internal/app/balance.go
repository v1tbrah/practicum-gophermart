package app

import (
	"context"
	"errors"

	"github.com/rs/zerolog/log"

	"practicum-gophermart/internal/model"
	dberr "practicum-gophermart/internal/storage/errors"
)

var ErrInsufficientFunds = errors.New("there are not enough funds in the account")

func (a *App) GetBalance(c context.Context, userID int64) (balance, withdrawn float64, err error) {
	log.Debug().Msg("app.GetBalance START")
	defer func() {
		logMethodEnd("app.GetBalance", err)
	}()

	balance, withdrawn, err = a.storage.GetBalance(c, userID)
	if err != nil {
		return -1, -1, err
	}

	return balance, withdrawn, nil
}

func (a *App) WithdrawFromBalance(c context.Context, userID int64, withdraw model.Withdraw) (err error) {
	log.Debug().Msg("app.WithdrawFromBalance START")
	defer func() {
		logMethodEnd("app.WithdrawFromBalance", err)
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

func (a *App) GetWithdrawals(c context.Context, userID int64) (withdrawals []model.Withdraw, err error) {
	log.Debug().Msg("app.GetWithdrawals START")
	defer func() {
		logMethodEnd("app.WithdrawFromBalance", err)
	}()

	withdrawals, err = a.storage.GetWithdrawals(c, userID)
	if err != nil {
		return nil, err
	}

	return withdrawals, nil
}
