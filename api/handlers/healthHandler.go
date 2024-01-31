package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golobby/container/v3"
	"github.com/steadfastie/gokube/data/services"
)

func HealthHandler(gc *gin.Context) {
	var mongodb *services.MongoDB
	container.Resolve(&mongodb)

	mongoConHealthy := mongodb.CheckConnection(gc)
	if mongoConHealthy {
		gc.Status(http.StatusOK)
	} else {
		gc.Status(http.StatusInternalServerError)
	}
}
