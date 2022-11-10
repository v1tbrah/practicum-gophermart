package api

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func (a *API) respond(c *gin.Context, code int, data interface{}) {
	c.JSON(code, data)
}

func (a *API) error(c *gin.Context, code int, err error) {
	if err != nil {
		log.Error().Err(err).Msg(c.Request.RequestURI)
		a.respond(c, code, gin.H{"error": err.Error()})
	} else {
		a.respond(c, code, nil)
	}
}
