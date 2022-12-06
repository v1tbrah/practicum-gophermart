package api

//go:generate mockery --all

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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
func New(application Application) (newAPI *API, err error) {
	log.Debug().Msg("api.New started")
	defer func() {
		logMethodEnd("api.New", err)
	}()

	newAPI = &API{}

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
func (a *API) Run() {
	log.Debug().Msg("api.Run started")
	defer log.Debug().Msg("api.Run ended")

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	errG, ctx := errgroup.WithContext(context.Background())

	errG.Go(func() error {
		return a.startListener(ctx, shutdown)
	})

	errG.Go(func() error {
		return a.startUpdatingOrdersStatus(ctx, shutdown)
	})

	if err := errG.Wait(); err != nil {
		log.Error().Err(err).Msg(err.Error())
		_, ok := <-shutdown
		if ok {
			close(shutdown)
		}
	}

	<-shutdown
	if err := a.app.CloseStorage(); err != nil {
		log.Err(err).Msg("storage closing")
	} else {
		log.Info().Msg("storage closed")
	}
	if err := a.serv.Shutdown(context.Background()); err != nil {
		log.Err(err).Msg("HTTP server shutdown")
	} else {
		log.Info().Msg("HTTP server gracefully shutdown")
	}

}

func (a *API) startListener(ctx context.Context, shutdown chan os.Signal) (err error) {
	log.Debug().Msg("api.startListener started")
	defer func() {
		logMethodEnd("api.startListener", err)
	}()

	ended := make(chan struct{})

	go func() {
		select {
		case <-ctx.Done():
		case _, ok := <-shutdown:
			if ok {
				close(shutdown)
			}
		default:
			defer a.serv.Close()
			log.Info().Str("addr", a.serv.Addr).Msg("starting http server")
			err = a.serv.ListenAndServe()
			ended <- struct{}{}
		}
	}()

	select {
	case <-ctx.Done():
	case <-ended:
	case _, ok := <-shutdown:
		if ok {
			close(shutdown)
		}
	}

	return err

}

func (a *API) startUpdatingOrdersStatus(ctx context.Context, shutdown chan os.Signal) (err error) {
	log.Debug().Msg("api.startUpdatingOrdersStatus started")
	defer func() {
		logMethodEnd("api.startUpdatingOrdersStatus", err)
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
		case _, ok := <-shutdown:
			if ok {
				close(shutdown)
			}
			return nil
		}
	}

}

func logMethodEnd(method string, err error) {
	msg := method + " END"
	if err != nil {
		log.Error().Err(err).Msg(msg)
	} else {
		log.Debug().Msg(msg)
	}
}
