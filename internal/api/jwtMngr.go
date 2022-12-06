package api

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

var (
	errAccessTokenIsExpired  = errors.New("access token is expired")
	errRefreshTokenIsExpired = errors.New("refresh token is expired")
)

type jwtMngr struct {
	accessTokenTTL  time.Duration
	signingKey      string
	refreshTokenTTL time.Duration
}

func newJwtMngr(signingKey string, accessTokenTTL, refreshTokenTTL time.Duration) *jwtMngr {
	log.Debug().Msg("api.newJwtMngr START")
	defer log.Debug().Msg("api.newJwtMngr END")

	if signingKey == "" {
		signingKey = "6Q7TibVvx32RBzMU4j3I5hIKMY2A2azi"
	}
	if accessTokenTTL == time.Second*0 {
		accessTokenTTL = time.Minute * 30
	}
	if refreshTokenTTL == time.Second*0 {
		refreshTokenTTL = time.Hour * 24 * 30
	}
	return &jwtMngr{signingKey: signingKey, accessTokenTTL: accessTokenTTL, refreshTokenTTL: refreshTokenTTL}
}

type tokenClaims struct {
	jwt.RegisteredClaims
	ID int64 `json:"id"`
}

func (j *jwtMngr) newAccessToken(id int64) (accessToken string, err error) {
	log.Debug().Msg("jwtMngr.newAccessToken START")
	defer func() {
		logMethodEnd("jwtMngr.newAccessToken", err)
	}()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		tokenClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.accessTokenTTL)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
			ID: id,
		},
	)

	accessToken, err = token.SignedString([]byte(j.signingKey))
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func (j *jwtMngr) newRefreshToken() string {
	log.Debug().Msg("jwtMngr.NewRefreshToken START")
	defer log.Debug().Msg("jwtMngr.NewRefreshToken END")

	return uuid.New().String()
}

func (j *jwtMngr) getID(accessToken string) (userID int64, err error) {
	log.Debug().Msg("jwtMngr.getID START")
	defer func() {
		logMethodEnd("jwtMngr.getID", err)
	}()

	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.signingKey), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return 0, errAccessTokenIsExpired
		}
		return 0, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, fmt.Errorf("error get user claims from token")
	}

	return int64(claims["id"].(float64)), nil
}
