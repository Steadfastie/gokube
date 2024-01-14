package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// EntryHandler PingExample godoc
//
//	@Summary	ping entry
//	@Tags			ping
//	@Accept		json
//	@Produce	json
//	@Success	200	{string}	Hello	world
//	@Router		/ping [get]
func EntryHandler(gc *gin.Context) {
	gc.JSON(http.StatusOK, "Hello world!")
}
