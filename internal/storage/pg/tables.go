package pg

import (
	"context"

	"github.com/rs/zerolog/log"
)

func initTables(ctx context.Context, p *Pg) error {
	log.Debug().Msg("pg.initTables START")
	var err error
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("pg.initTables END")
		} else {
			log.Debug().Msg("pg.initTables END")
		}
	}()

	if p.db == nil {
		err = ErrDBIsNilPointer
		return err
	}

	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if errTxRollback := tx.Rollback(); errTxRollback != nil {
			log.Error().Err(errTxRollback).Msg("tx rollback")
		}
	}()

	_, err = tx.ExecContext(ctx, queryCreateTypeOrderStatus)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, queryCreateTableUsers)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, queryCreateTableRefreshSessions)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, queryCreateTableOrders)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, queryCreateTableBalance)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, queryCreateTableWithdrawals)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
