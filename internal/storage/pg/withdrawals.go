package pg

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/rs/zerolog/log"

	"practicum-gophermart/internal/model"
	dberr "practicum-gophermart/internal/storage/errors"
)

type withdrawalsStmts struct {
	stmtAddWithdrawal  *sql.Stmt
	stmtGetWithdrawals *sql.Stmt
}

func prepareWithdrawalsStmts(ctx context.Context, p *Pg) error {

	newWithdrawalsStmts := withdrawalsStmts{}

	var err error

	if newWithdrawalsStmts.stmtAddWithdrawal, err = p.db.PrepareContext(ctx, queryAddWithdrawal); err != nil {
		return err
	}

	if newWithdrawalsStmts.stmtGetWithdrawals, err = p.db.PrepareContext(ctx, queryGetWithdrawals); err != nil {
		return err
	}

	p.withdrawalsStmts = &newWithdrawalsStmts

	return nil
}

func (p *Pg) AddWithdrawal(ctx context.Context, userID int64, withdraw model.Withdraw) error {
	log.Debug().Msg("Pg.AddWithdrawal START")
	var err error
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("Pg.AddWithdrawal END")
		} else {
			log.Debug().Msg("Pg.AddWithdrawal END")
		}
	}()

	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if errTxRollback := tx.Rollback(); errTxRollback != nil {
			log.Error().Err(errTxRollback).Msg("tx rollback")
		}
	}()

	_, err = tx.StmtContext(ctx, p.withdrawalsStmts.stmtAddWithdrawal).ExecContext(ctx, userID, withdraw.Order, withdraw.Sum, withdraw.ProcessedAt)
	if err != nil {
		return err
	}

	var newBalance float64
	err = tx.StmtContext(ctx, p.balanceStmts.stmtReduceBalance).QueryRowContext(ctx, userID, withdraw.Sum).Scan(&newBalance)
	if err != nil {
		return err
	}
	if newBalance < 0 {
		if errTxRollback := tx.Rollback(); errTxRollback != nil {
			log.Error().Err(errTxRollback).Msg("tx rollback")
		}
		return dberr.ErrNegativeBalance
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (p *Pg) GetWithdrawals(ctx context.Context, userID int64) (withdrawals []model.Withdraw, err error) {
	log.Debug().Msg("Pg.GetWithdrawals START")
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("Pg.GetWithdrawals END")
		} else {
			log.Debug().Msg("Pg.GetWithdrawals END")
		}
	}()

	rows, err := p.withdrawalsStmts.stmtGetWithdrawals.QueryContext(ctx, userID)
	if err != nil {
		return nil, err
	}
	defer func() {
		if errRowsClose := rows.Close(); errRowsClose != nil {
			log.Error().Err(err).Msg("closing sql rows")
		}
	}()

	for rows.Next() {
		currWithdraw := model.Withdraw{}
		if err = rows.Scan(&currWithdraw.Order, &currWithdraw.Sum, &currWithdraw.ProcessedAt); err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, currWithdraw)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return withdrawals, nil
}

func (w *withdrawalsStmts) Close() (err error) {

	if err = w.stmtAddWithdrawal.Close(); err != nil {
		return fmt.Errorf("closing stmt 'AddWithdrawal' : %w", err)
	}

	if err = w.stmtGetWithdrawals.Close(); err != nil {
		return fmt.Errorf("closing stmt 'GetWithdrawals' : %w", err)
	}

	return nil
}
