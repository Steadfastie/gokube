package handlers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/steadfastie/gokube/infrastucture/errors"
)

type PanicType struct {
	Type string `uri:"type" binding:"required"`
}

// PanicHandler PanicExample godoc
//
//	@Summary	throws a panic
//	@Tags			panic
//	@Accept		json
//	@Produce	json
//	@Param		type	path		string	true	"Type of panic"
//	@Failure	409		{object}	errors.HTTPError
//	@Failure	422		{object}	errors.HTTPError
//	@Failure	500		{object}	errors.HTTPError
//	@Router		/panic/{type} [get]
func PanicHandler(c *gin.Context) {
	var panicType PanicType
	c.ShouldBindUri(&panicType)

	switch panicType.Type {
	case "business":
		panic(errors.NewBusinessRuleError("We do not welcome such behavior here"))
	case "optimistic":
		panic(errors.NewOptimisticLockError("We will recover!"))
	default:
		panic(fmt.Sprintf("Panic type(%s) is not recognized", panicType.Type))
	}
}
