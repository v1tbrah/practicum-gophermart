package pg

import (
	"context"
	"database/sql"

	"github.com/gin-gonic/gin"
)

type balanceStmts struct {
	stmtCreateStartingBalance *sql.Stmt
	stmtGetBalance            *sql.Stmt
	stmtIncreaseBalance       *sql.Stmt
	stmtReduceBalance         *sql.Stmt
}

func prepareBalanceStmts(ctx context.Context, p *Pg) error {

	newBalanceStmts := balanceStmts{}

	if stmtCreateStartingBalance, err := p.db.PrepareContext(ctx, queryCreateStartingBalance); err != nil {
		return err
	} else {
		newBalanceStmts.stmtCreateStartingBalance = stmtCreateStartingBalance
	}

	if stmtGetBalance, err := p.db.PrepareContext(ctx, queryGetBalance); err != nil {
		return err
	} else {
		newBalanceStmts.stmtGetBalance = stmtGetBalance
	}

	if stmtIncreaseBalance, err := p.db.PrepareContext(ctx, queryIncreaseBalance); err != nil {
		return err
	} else {
		newBalanceStmts.stmtIncreaseBalance = stmtIncreaseBalance
	}

	if stmtReduceBalance, err := p.db.PrepareContext(ctx, queryReduceBalance); err != nil {
		return err
	} else {
		newBalanceStmts.stmtReduceBalance = stmtReduceBalance
	}

	p.balanceStmts = &newBalanceStmts

	return nil
}

func (p *Pg) GetBalance(c *gin.Context, userID int64) (float64, float64, error) {
	var balance, withdrawn float64
	err := p.balanceStmts.stmtGetBalance.QueryRowContext(c, userID).Scan(&balance, &withdrawn)
	if err != nil {
		return -1, -1, err
	}
	return balance, withdrawn, nil
}
