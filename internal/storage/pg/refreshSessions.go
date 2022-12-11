package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"

	"practicum-gophermart/internal/model"
	dberr "practicum-gophermart/internal/storage/errors"
)

type refreshSessionStmts struct {
	stmtAddRefreshSession        *sql.Stmt
	stmtDeleteRefreshSession     *sql.Stmt
	stmtGetRefreshSessionByToken *sql.Stmt
}

func prepareRefreshSessionStmts(ctx context.Context, p *Pg) error {

	newRefreshSessionStmts := refreshSessionStmts{}

	var err error

	if newRefreshSessionStmts.stmtAddRefreshSession, err = p.db.PrepareContext(ctx, queryAddRefreshSession); err != nil {
		return err
	}

	if newRefreshSessionStmts.stmtDeleteRefreshSession, err = p.db.PrepareContext(ctx, queryDeleteRefreshSessions); err != nil {
		return err
	}

	if newRefreshSessionStmts.stmtGetRefreshSessionByToken, err = p.db.PrepareContext(ctx, queryGetRefreshSessionByToken); err != nil {
		return err
	}

	p.refreshSessionStmts = &newRefreshSessionStmts

	return nil
}

func (p *Pg) UpdateRefreshSession(ctx context.Context, newRefreshSession *model.RefreshSession) error {
	log.Debug().Str("UserID", fmt.Sprint(newRefreshSession.UserID)).Msg("Pg.UpdateRefreshSession START")
	var err error
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("Pg.UpdateRefreshSession END")
		} else {
			log.Debug().Msg("Pg.UpdateRefreshSession END")
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

	_, err = tx.StmtContext(ctx, p.refreshSessionStmts.stmtDeleteRefreshSession).ExecContext(ctx, newRefreshSession.UserID)
	if err != nil {
		return err
	}

	_, err = tx.StmtContext(ctx, p.refreshSessionStmts.stmtAddRefreshSession).ExecContext(ctx,
		newRefreshSession.UserID,
		newRefreshSession.Token,
		newRefreshSession.ExpiresIn)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (p *Pg) GetRefreshSessionByToken(c context.Context, token string) (*model.RefreshSession, error) {
	log.Debug().Msg("Pg.GetRefreshSessionByToken START")
	var err error
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("Pg.GetRefreshSessionByToken END")
		} else {
			log.Debug().Msg("Pg.GetRefreshSessionByToken END")
		}
	}()

	var refreshSession model.RefreshSession
	err = p.refreshSessionStmts.stmtGetRefreshSessionByToken.QueryRowContext(c, token).Scan(&refreshSession.UserID, &refreshSession.ExpiresIn)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf(`pg: %w: %s`, dberr.ErrRefreshSessionIsNotExists, err)
		}
		return nil, err
	}

	return &refreshSession, nil
}

func (r *refreshSessionStmts) Close() (err error) {

	if err = r.stmtAddRefreshSession.Close(); err != nil {
		return fmt.Errorf("closing stmt 'AddRefreshSession' : %w", err)
	}

	if err = r.stmtDeleteRefreshSession.Close(); err != nil {
		return fmt.Errorf("closing stmt 'DeleteRefreshSession' : %w", err)
	}

	if err = r.stmtGetRefreshSessionByToken.Close(); err != nil {
		return fmt.Errorf("closing stmt 'GetRefreshSessionByToken' : %w", err)
	}

	return nil
}
