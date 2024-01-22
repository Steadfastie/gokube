package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/steadfastie/gokube/data"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type CounterController struct {
	Repository data.CounterRepository `container:"type"`
	Logger     *zap.Logger            `container:"type"`
}

// GetByIdHandler Gets a counter by Id
//
//	@Summary	retrieves a counter by id from database
//	@Tags		counter
//	@Accept		json
//	@Produce	json
//	@Security	OAuth2AccessCode
//	@Param		id	path		string					true	"Counter ID"
//	@Success	200	{object}	data.CounterResponse	"Requested counter"
//	@Failure	400	{object}	errors.HTTPError
//	@Failure	404	{object}	errors.HTTPError	"Counter not found"
//	@Router		/counter/{id} [get]
func (controller *CounterController) GetByIdHandler(gc *gin.Context) {
	resultChan := make(chan *data.CounterDocument)
	errChan := make(chan error)

	defer close(resultChan)
	defer close(errChan)

	go controller.Repository.GetById(gc, gc.Param("id"), resultChan, errChan)

	select {
	case foundCounter := <-resultChan:
		gc.JSON(200, foundCounter.MapToResponseModel())
	case err := <-errChan:
		gc.JSON(400, err)
	}
}

// CreateHandler Creates a counter object
//
//	@Summary	creates a basic structure of the project
//	@Tags		counter
//	@Accept		json
//	@Produce	json
//	@Security	OAuth2AccessCode
//	@Success	200	{string}	id	"ID of the created counter object"
//	@Failure	400	{object}	errors.HTTPError
//	@Failure	404	{object}	errors.HTTPError	"Counter not found"
//	@Router		/counter [post]
func (controller *CounterController) CreateHandler(gc *gin.Context) {
	resultChan := make(chan primitive.ObjectID)
	errChan := make(chan error)

	defer close(resultChan)
	defer close(errChan)

	go controller.Repository.Create(gc, resultChan, errChan)

	select {
	case resultID := <-resultChan:
		gc.JSON(200, resultID.Hex())
	case err := <-errChan:
		gc.JSON(400, err)
	}
}

// PatchHandler Updates a counter
//
//	@Summary	changes counter value
//	@Tags		counter
//	@Accept		json
//	@Produce	json
//	@Security	OAuth2AccessCode
//	@Param		id		path		string						true	"Counter ID"
//	@Param		patch	body		data.PatchModel				true	"Describe your desires"
//	@Success	200		{object}	data.PatchCounterResponse	"ID of the created counter object"
//	@Failure	400		{object}	errors.HTTPError
//	@Failure	404		{object}	errors.HTTPError	"Counter not found"
//	@Router		/counter/{id} [patch]
func (controller *CounterController) PatchHandler(gc *gin.Context) {
	var patchModel data.PatchModel

	if err := gc.ShouldBindJSON(&patchModel); err != nil {
		gc.JSON(400, err)
	}

	resultChan := make(chan *data.PatchCounterResponse)
	errChan := make(chan error)

	defer close(resultChan)
	defer close(errChan)

	go controller.Repository.Patch(gc, gc.Param("id"), &patchModel, resultChan, errChan)

	select {
	case patchResult := <-resultChan:
		gc.JSON(200, patchResult)
	case err := <-errChan:
		gc.JSON(400, err)
	}
}
