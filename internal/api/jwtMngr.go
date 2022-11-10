package api

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"time"
)

var (
	ErrAccessTokenIsExpired  = errors.New("access token is expired")
	ErrRefreshTokenIsExpired = errors.New("refresh token is expired")
)

type jwtMngr struct {
	accessTokenTTL  time.Duration
	signingKey      string
	refreshTokenTTL time.Duration
}

func newJwtMngr(signingKey string, accessTokenTTL, refreshTokenTTL time.Duration) *jwtMngr {
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

func (j *jwtMngr) newAccessToken(id int64) (string, error) {
	log.Debug().Msg("jwtMngr.newAccessToken START")
	var err error
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("jwtMngr.newAccessToken END")
		} else {
			log.Debug().Msg("jwtMngr.newAccessToken END")
		}
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

	signedToken, err := token.SignedString([]byte(j.signingKey))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func (j *jwtMngr) newRefreshToken() string {
	log.Debug().Msg("jwtMngr.NewRefreshToken START")
	defer log.Debug().Msg("jwtMngr.NewRefreshToken END")

	return uuid.New().String()
}

func (j *jwtMngr) getID(accessToken string) (int64, error) {

	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.signingKey), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return 0, ErrAccessTokenIsExpired
		}
		return 0, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, fmt.Errorf("error get user claims from token")
	}

	return int64(claims["id"].(float64)), nil
}
