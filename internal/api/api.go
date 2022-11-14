package api

//go:generate mockery --all

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

type API struct {
	authMngr    *authMngr
	app         Application
	serv        *http.Server
	accrualMngr *accrualMngr
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
		Addr:    application.Config().ServAddr(),
		Handler: newAPI.newRouter(),
	}

	newAPI.accrualMngr = newAccrualMngr()

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

	log.Info().Msg("api started")

	errG, _ := errgroup.WithContext(context.Background())

	errG.Go(a.startListener)

	errG.Go(a.startUpdatingOrdersStatus)

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

func (a *API) startListener() error {
	log.Debug().Msg("api.startListener started")
	defer log.Debug().Msg("api.startListener ended")

	defer a.serv.Close()
	return a.serv.ListenAndServe()
}

func (a *API) startUpdatingOrdersStatus() error {
	interval := a.app.Config().OrderStatusUpdateInterval()
	if interval == time.Second*0 {
		return nil
	}

	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			err := a.updateOrdersStatus()
			if err != nil {
				return err
			}
		}
	}

}
