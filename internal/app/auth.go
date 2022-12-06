package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"

	"practicum-gophermart/internal/model"
	dberr "practicum-gophermart/internal/storage/errors"
)

var (
	ErrUserAlreadyExists        = errors.New("user already exists")
	ErrInvalidLoginOrPassword   = errors.New("invalid login or password")
	ErrRefreshSessionIsNotExist = errors.New("refresh session is not exists")
)

func (a *App) CreateUser(c context.Context, user *model.User) (id int64, err error) {
	log.Debug().Msg("app.CreateUser START")
	defer func() {
		logMethodEnd("app.CreateUser", err)
	}()

	hash, err := a.pwdMngr.hash([]byte(user.Password))
	if err != nil {
		return 0, err
	}
	user.Password = string(hash)

	id, err = a.storage.AddUser(c, user)
	if err != nil {
		if errors.Is(err, dberr.ErrLoginAlreadyExists) {
			return 0, fmt.Errorf(`app: %w: %s`, ErrUserAlreadyExists, err)
		}
		return 0, err
	}

	return id, nil
}

func (a *App) GetUser(c context.Context, login, pwd string) (user *model.User, err error) {
	log.Debug().Msg("app.GetUser START")
	defer func() {
		logMethodEnd("app.GetUser", err)
	}()

	currHashedPwd, err := a.storage.GetUserPassword(c, login)
	if err != nil {
		if errors.Is(err, dberr.ErrInvalidLoginOrPassword) {
			return nil, fmt.Errorf(`app: %w: %s`, ErrInvalidLoginOrPassword, err)
		}
		return nil, err
	}

	err = a.pwdMngr.compare([]byte(currHashedPwd), []byte(pwd))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return nil, fmt.Errorf(`app: %w: %s`, ErrInvalidLoginOrPassword, err)
		}
		return nil, err
	}

	user, err = a.storage.GetUser(c, login, currHashedPwd)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (a *App) NewRefreshSession(c context.Context, newRefreshSession *model.RefreshSession) (err error) {
	log.Debug().Str("userID", fmt.Sprint(newRefreshSession.UserID)).Msg("app.NewRefreshSession START")
	defer func() {
		logMethodEnd("app.NewRefreshSession", err)
	}()

	err = a.storage.UpdateRefreshSession(c, newRefreshSession)
	if err != nil {
		return err
	}

	return nil
}

func (a *App) GetRefreshSessionByToken(c context.Context, refreshToken string) (refreshSession *model.RefreshSession, err error) {
	log.Debug().Msg("app.GetRefreshSessionByToken START")
	defer func() {
		logMethodEnd("app.GetRefreshSessionByToken", err)
	}()

	refreshSession, err = a.storage.GetRefreshSessionByToken(c, refreshToken)
	if err != nil {
		if errors.Is(err, dberr.ErrRefreshSessionIsNotExists) {
			return nil, ErrRefreshSessionIsNotExist
		}
		return nil, err
	}

	return refreshSession, nil
}
