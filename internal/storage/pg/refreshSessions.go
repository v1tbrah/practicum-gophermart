package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
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

	if stmtAddRefreshSession, err := p.db.PrepareContext(ctx, queryAddRefreshSession); err != nil {
		return err
	} else {
		newRefreshSessionStmts.stmtAddRefreshSession = stmtAddRefreshSession
	}

	if stmtDeleteRefreshSession, err := p.db.PrepareContext(ctx, queryDeleteRefreshSessions); err != nil {
		return err
	} else {
		newRefreshSessionStmts.stmtDeleteRefreshSession = stmtDeleteRefreshSession
	}

	if stmtGetRefreshSessionByToken, err := p.db.PrepareContext(ctx, queryGetRefreshSessionByToken); err != nil {
		return err
	} else {
		newRefreshSessionStmts.stmtGetRefreshSessionByToken = stmtGetRefreshSessionByToken
	}

	p.refreshSessionStmts = &newRefreshSessionStmts

	return nil
}

func (p *Pg) UpdateRefreshSession(c *gin.Context, newRefreshSession *model.RefreshSession) error {
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
	defer tx.Rollback()

	_, err = tx.StmtContext(c, p.refreshSessionStmts.stmtDeleteRefreshSession).ExecContext(c, newRefreshSession.UserID)
	if err != nil {
		return err
	}

	_, err = tx.StmtContext(c, p.refreshSessionStmts.stmtAddRefreshSession).ExecContext(c,
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

func (p *Pg) GetRefreshSessionByToken(c *gin.Context, token string) (*model.RefreshSession, error) {
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
