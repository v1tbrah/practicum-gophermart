package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
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

	if stmtAddUser, err := p.db.PrepareContext(ctx, queryAddUser); err != nil {
		return err
	} else {
		newUsersStmts.stmtAddUser = stmtAddUser
	}

	if stmtGetUser, err := p.db.PrepareContext(ctx, queryGetUser); err != nil {
		return err
	} else {
		newUsersStmts.stmtGetUser = stmtGetUser
	}

	if stmtGetUserPassword, err := p.db.PrepareContext(ctx, queryGetUserPassword); err != nil {
		return err
	} else {
		newUsersStmts.stmtGetUserPassword = stmtGetUserPassword
	}

	p.usersStmts = &newUsersStmts

	return nil
}

func (p *Pg) AddUser(c *gin.Context, user *model.User) (int64, error) {
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
	err = tx.StmtContext(c, p.usersStmts.stmtAddUser).QueryRowContext(c, user.Login, user.Password).Scan(&id)
	if err != nil {
		if pgError, ok := err.(*pgconn.PgError); ok &&
			pgError.Code == pgerrcode.UniqueViolation &&
			pgError.ConstraintName == "users_login_key" {
			return 0, fmt.Errorf(`pg: %w: %s`, dberr.ErrLoginAlreadyExists, err)
		}
		return 0, err
	}

	_, err = tx.StmtContext(c, p.balanceStmts.stmtCreateStartingBalance).ExecContext(c, id)
	if err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return id, nil
}

func (p *Pg) GetUser(c *gin.Context, login, password string) (*model.User, error) {
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

func (p *Pg) GetUserPassword(c *gin.Context, login string) (string, error) {
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
