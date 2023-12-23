package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// EntryHandler PingExample godoc
//
//	@Summary	ping entry
//	@Tags		ping
//	@Accept		json
//	@Produce	json
//	@Success	200	{string}	Hello	world
//	@Router		/ping [get]
func EntryHandler(g *gin.Context) {
	g.JSON(http.StatusOK, "Hello world!")
}
