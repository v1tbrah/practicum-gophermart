package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib" // for use pgx + stdlib
	"github.com/rs/zerolog/log"
)

var ErrDBIsNilPointer = errors.New("database is nil pointer")

const (
	maxOpenConns = 20
	maxIdleConns = 20
	maxIdleTime  = time.Second * 30
	maxLifeTime  = time.Minute * 2
)

type Pg struct {
	db                  *sql.DB
	usersStmts          *usersStmts
	ordersStmts         *ordersStmts
	refreshSessionStmts *refreshSessionStmts
	balanceStmts        *balanceStmts
	withdrawalsStmts    *withdrawalsStmts
}

func New(pgConn string) (*Pg, error) {
	log.Debug().Str("pgConn", pgConn).Msg("Pg.New START")
	var err error
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("Pg.New END")
		} else {
			log.Debug().Msg("Pg.New END")
		}
	}()

	newPg := Pg{}

	db, err := sql.Open("pgx", pgConn)
	if err != nil {
		return nil, err
	}
	newPg.db = db

	newPg.db.SetMaxOpenConns(maxOpenConns)
	newPg.db.SetMaxIdleConns(maxIdleConns)
	newPg.db.SetConnMaxIdleTime(maxIdleTime)
	newPg.db.SetConnMaxLifetime(maxLifeTime)

	ctx := context.Background()

	if err = initTables(ctx, &newPg); err != nil {
		return nil, err
	}

	if err = prepareUserStmts(ctx, &newPg); err != nil {
		return nil, err
	}

	if err = prepareRefreshSessionStmts(ctx, &newPg); err != nil {
		return nil, err
	}

	if err = prepareOrdersStmts(ctx, &newPg); err != nil {
		return nil, err
	}

	if err = prepareBalanceStmts(ctx, &newPg); err != nil {
		return nil, err
	}

	if err = prepareWithdrawalsStmts(ctx, &newPg); err != nil {
		return nil, err
	}

	return &newPg, nil
}

func (p *Pg) Close() error {
	log.Debug().Msg("Pg.CloseConnection START")
	var err error
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("Pg.CloseConnection END")
		} else {
			log.Debug().Msg("Pg.CloseConnection END")
		}
	}()

	if err = p.usersStmts.Close(); err != nil {
		return fmt.Errorf("closing user stmts: %w", err)
	}

	if err = p.ordersStmts.Close(); err != nil {
		return fmt.Errorf("closing order stmts: %w", err)
	}

	if err = p.refreshSessionStmts.Close(); err != nil {
		return fmt.Errorf("closing refresh session stmts: %w", err)
	}

	if err = p.balanceStmts.Close(); err != nil {
		return fmt.Errorf("closing balance stmts: %w", err)
	}

	if err = p.withdrawalsStmts.Close(); err != nil {
		return fmt.Errorf("closing withdrawals stmts: %w", err)
	}

	err = p.db.Close()
	if err != nil {
		return fmt.Errorf("closing db connection: %w", err)
	}

	return nil
}
