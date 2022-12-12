package pg

import (
	"context"
	"database/sql"
	"fmt"
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

func (p *Pg) GetBalance(ctx context.Context, userID int64) (balance, withdrawn float64, err error) {
	err = p.balanceStmts.stmtGetBalance.QueryRowContext(ctx, userID).Scan(&balance, &withdrawn)
	if err != nil {
		return -1, -1, err
	}
	return balance, withdrawn, nil
}

func (b *balanceStmts) Close() (err error) {

	if err = b.stmtCreateStartingBalance.Close(); err != nil {
		return fmt.Errorf("closing stmt 'CreateStartingBalance' : %w", err)
	}

	if err = b.stmtGetBalance.Close(); err != nil {
		return fmt.Errorf("closing stmt 'GetBalance' : %w", err)
	}

	if err = b.stmtIncreaseBalance.Close(); err != nil {
		return fmt.Errorf("closing stmt 'IncreaseBalance' : %w", err)
	}

	if err = b.stmtReduceBalance.Close(); err != nil {
		return fmt.Errorf("closing stmt 'ReduceBalance' : %w", err)
	}

	return nil
}
