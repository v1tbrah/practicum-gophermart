package pg

import (
	"context"
	"database/sql"
)

type balanceStmts struct {
	stmtCreateStartingBalance *sql.Stmt
	stmtGetBalance            *sql.Stmt
	stmtIncreaseBalance       *sql.Stmt
	stmtReduceBalance         *sql.Stmt
}

func prepareBalanceStmts(ctx context.Context, p *Pg) error {

	newBalanceStmts := balanceStmts{}

	var err error

	if newBalanceStmts.stmtCreateStartingBalance, err = p.db.PrepareContext(ctx, queryCreateStartingBalance); err != nil {
		return err
	}

	if newBalanceStmts.stmtGetBalance, err = p.db.PrepareContext(ctx, queryGetBalance); err != nil {
		return err
	}

	if newBalanceStmts.stmtIncreaseBalance, err = p.db.PrepareContext(ctx, queryIncreaseBalance); err != nil {
		return err
	}

	if newBalanceStmts.stmtReduceBalance, err = p.db.PrepareContext(ctx, queryReduceBalance); err != nil {
		return err
	}

	p.balanceStmts = &newBalanceStmts

	return nil
}

func (p *Pg) GetBalance(ctx context.Context, userID int64) (float64, float64, error) {
	var balance, withdrawn float64
	err := p.balanceStmts.stmtGetBalance.QueryRowContext(ctx, userID).Scan(&balance, &withdrawn)
	if err != nil {
		return -1, -1, err
	}
	return balance, withdrawn, nil
}
