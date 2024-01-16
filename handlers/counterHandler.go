package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/steadfastie/gokube/data"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type CounterController struct {
	Repository data.CounterRepository
	Logger     *zap.Logger
}

// CreateHandler Creates a counter object
//
//	@Summary	creates a basic structure of the project
//	@Tags		counter
//	@Accept		json
//	@Produce	json
//	@Success	200	{string}	id	of	the	object
//	@Router		/counter [post]
func (controller *CounterController) CreateHandler(gc *gin.Context) {
	resultChan := make(chan primitive.ObjectID)
	errChan := make(chan error)

	defer close(resultChan)
	defer close(errChan)

	go controller.Repository.Create(gc, resultChan, errChan)

	select {
	case resultID := <-resultChan:
		gc.JSON(200, gin.H{"id": resultID.Hex()})
	case <-errChan:
		panic("Could not create a counter")
	}
}
