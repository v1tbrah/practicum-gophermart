package pg

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/rs/zerolog/log"
)

var ErrDBIsNilPointer = errors.New("database is nil pointer")

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

	newPg.db.SetMaxOpenConns(20)
	newPg.db.SetMaxIdleConns(20)
	newPg.db.SetConnMaxIdleTime(time.Second * 30)
	newPg.db.SetConnMaxLifetime(time.Minute * 2)

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

	err = p.db.Close()
	if err != nil {
		return err
	}

	return err
}
