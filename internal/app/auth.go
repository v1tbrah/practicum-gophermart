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

func (a *App) CreateUser(c context.Context, user *model.User) (int64, error) {
	log.Debug().Msg("app.CreateUser START")
	var err error
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("app.CreateUser END")
		} else {
			log.Debug().Msg("app.CreateUser END")
		}
	}()

	hash, err := a.pwdMngr.hash([]byte(user.Password))
	if err != nil {
		return 0, err
	}
	user.Password = string(hash)

	id, err := a.storage.AddUser(c, user)
	if err != nil {
		if errors.Is(err, dberr.ErrLoginAlreadyExists) {
			return 0, fmt.Errorf(`app: %w: %s`, ErrUserAlreadyExists, err)
		}
		return 0, err
	}

	return id, nil
}

func (a *App) GetUser(c context.Context, login, pwd string) (*model.User, error) {
	log.Debug().Msg("app.GetUser START")
	var err error
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("app.GetUser END")
		} else {
			log.Debug().Msg("app.GetUser END")
		}
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

	user, err := a.storage.GetUser(c, login, currHashedPwd)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (a *App) NewRefreshSession(c context.Context, newRefreshSession *model.RefreshSession) error {
	log.Debug().Str("userID", fmt.Sprint(newRefreshSession.UserID)).Msg("app.NewRefreshSession START")
	var err error
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("app.NewRefreshSession END")
		} else {
			log.Debug().Msg("app.NewRefreshSession END")
		}
	}()

	err = a.storage.UpdateRefreshSession(c, newRefreshSession)
	if err != nil {
		return err
	}

	return nil
}

func (a *App) GetRefreshSessionByToken(c context.Context, refreshToken string) (*model.RefreshSession, error) {
	refreshSession, err := a.storage.GetRefreshSessionByToken(c, refreshToken)
	if err != nil {
		if errors.Is(err, dberr.ErrRefreshSessionIsNotExists) {
			return nil, ErrRefreshSessionIsNotExist
		}
		return nil, err
	}

	return refreshSession, nil
}
