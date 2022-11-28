package api

//go:generate mockery --all

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

var errInvalidIntervalUpdateOrderStatus = errors.New("invalid order status update interval")

type API struct {
	authMngr    *authMngr
	app         Application
	serv        *http.Server
	accrualMngr *resty.Client
}

// New returns new API.
func New(application Application) (*API, error) {
	log.Debug().Msg("api.New started")
	var err error
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("api.New ended")
		} else {
			log.Debug().Msg("api.New ended")
		}
	}()

	newAPI := &API{}

	newAPI.authMngr = newAuthMngr()

	newAPI.app = application

	newAPI.serv = &http.Server{
		Addr:    application.Config().ServAPIAddr(),
		Handler: newAPI.newRouter(),
	}

	newAPI.accrualMngr = resty.New()

	return newAPI, nil
}

func (a *API) newRouter() *gin.Engine {

	r := gin.Default()

	user := r.Group("/api/user")
	{
		auth := user.Group("/")
		{
			auth.POST("register", a.signUpHandler)
			auth.POST("login", a.signInHandler)
		}

		orders := user.Group("/").Use(a.checkAuthMiddleware)
		{
			orders.POST("orders", a.setOrderHandler)
			orders.GET("orders", a.ordersHandler)
		}

		balance := user.Group("/balance").Use(a.checkAuthMiddleware)
		{
			balance.GET("/", a.balanceHandler)
			balance.POST("/withdraw", a.withdrawPointsHandler)
		}

		withdraw := user.Group("/").Use(a.checkAuthMiddleware)
		{
			withdraw.GET("/withdrawals", a.withdrawnPointsHandler)
		}

	}

	return r
}

// Run API starts the API.
func (a *API) Run() error {
	log.Debug().Msg("api.Run started")
	defer log.Debug().Msg("api.Run ended")

	errG, ctx := errgroup.WithContext(context.Background())

	errG.Go(func() error {
		return a.startListener(ctx)
	})

	errG.Go(func() error {
		return a.startUpdatingOrdersStatus(ctx)
	})

	if err := errG.Wait(); err != nil {
		if errCloser := a.app.CloseStorage(); errCloser != nil {
			return errCloser
		}
		return err
	}

	if err := a.app.CloseStorage(); err != nil {
		return err
	}

	return nil

}

func (a *API) startListener(ctx context.Context) (err error) {
	log.Debug().Msg("api.startListener started")
	defer log.Debug().Msg("api.startListener ended")

	c := make(chan struct{})

	go func() {
		select {
		case <-ctx.Done():
			return
		default:
			defer a.serv.Close()
			log.Info().Str("addr", a.serv.Addr).Msg("starting http server")
			err = a.serv.ListenAndServe()
			c <- struct{}{}
		}
	}()

	select {
	case <-ctx.Done():
		return nil
	case <-c:
		return err
	}

}

func (a *API) startUpdatingOrdersStatus(ctx context.Context) (err error) {
	log.Debug().Msg("api.startUpdatingOrdersStatus started")
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("api.startUpdatingOrdersStatus ended")
		} else {
			log.Debug().Msg("api.startUpdatingOrdersStatus ended")
		}
	}()

	interval := a.app.Config().OrderStatusUpdateInterval()
	if interval == 0 {
		return errInvalidIntervalUpdateOrderStatus
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			err = a.updateOrdersStatus()
			if err != nil {
				return err
			}
		}
	}

}
