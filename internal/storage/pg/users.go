package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/rs/zerolog/log"

	"practicum-gophermart/internal/model"
	dberr "practicum-gophermart/internal/storage/errors"
)

type usersStmts struct {
	stmtAddUser         *sql.Stmt
	stmtGetUser         *sql.Stmt
	stmtGetUserPassword *sql.Stmt
}

func prepareUserStmts(ctx context.Context, p *Pg) error {

	newUsersStmts := usersStmts{}

	var err error

	if newUsersStmts.stmtAddUser, err = p.db.PrepareContext(ctx, queryAddUser); err != nil {
		return err
	}

	if newUsersStmts.stmtGetUser, err = p.db.PrepareContext(ctx, queryGetUser); err != nil {
		return err
	}

	if newUsersStmts.stmtGetUserPassword, err = p.db.PrepareContext(ctx, queryGetUserPassword); err != nil {
		return err
	}

	p.usersStmts = &newUsersStmts

	return nil
}

func (p *Pg) AddUser(ctx context.Context, user *model.User) (int64, error) {
	log.Debug().Str("user", user.String()).Msg("Pg.AddUser START")
	var err error
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("Pg.AddUser END")
		} else {
			log.Debug().Msg("Pg.AddUser END")
		}
	}()

	tx, err := p.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var id int64
	err = tx.StmtContext(ctx, p.usersStmts.stmtAddUser).QueryRowContext(ctx, user.Login, user.Password).Scan(&id)
	if err != nil {
		if pgError, ok := err.(*pgconn.PgError); ok &&
			pgError.Code == pgerrcode.UniqueViolation &&
			pgError.ConstraintName == "users_login_key" {
			return 0, fmt.Errorf(`pg: %w: %s`, dberr.ErrLoginAlreadyExists, err)
		}
		return 0, err
	}

	_, err = tx.StmtContext(ctx, p.balanceStmts.stmtCreateStartingBalance).ExecContext(ctx, id)
	if err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return id, nil
}

func (p *Pg) GetUser(c context.Context, login, password string) (*model.User, error) {
	log.Debug().Str("login", login).Str("password", password).Msg("Pg.GetUser START")
	var err error
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("Pg.GetUser END")
		} else {
			log.Debug().Msg("Pg.GetUser END")
		}
	}()

	var user model.User
	err = p.usersStmts.stmtGetUser.QueryRowContext(c, login, password).Scan(&user.ID, &user.Login, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf(`pg: %w: %s`, dberr.ErrInvalidLoginOrPassword, err)
		}
		return nil, err
	}

	return &user, nil
}

func (p *Pg) GetUserPassword(c context.Context, login string) (string, error) {
	log.Debug().Str("login", login).Msg("Pg.GetUserPassword START")
	var err error
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("Pg.GetUserPassword END")
		} else {
			log.Debug().Msg("Pg.GetUserPassword END")
		}
	}()

	var password string
	err = p.usersStmts.stmtGetUserPassword.QueryRowContext(c, login).Scan(&password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", dberr.ErrInvalidLoginOrPassword
		}
		return "", err
	}
	return password, nil
}

func (u *usersStmts) Close() (err error) {

	if err = u.stmtAddUser.Close(); err != nil {
		return fmt.Errorf("closing stmt 'AddUser' : %w", err)
	}

	if err = u.stmtGetUser.Close(); err != nil {
		return fmt.Errorf("closing stmt 'GetUser' : %w", err)
	}

	if err = u.stmtGetUserPassword.Close(); err != nil {
		return fmt.Errorf("closing stmt 'GetUserPassword' : %w", err)
	}

	return nil
}
