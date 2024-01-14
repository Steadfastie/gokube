package infrastucture

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golobby/container/v3"
	"github.com/steadfastie/gokube/infrastucture/errors"
	"go.uber.org/zap"
)

type caughtError struct {
	data interface{}
}

func RecoveryMiddleware(c *gin.Context, e any) {
	container.Call(func(logger *zap.Logger) {
		GlobalPanicRecovery(c, e, logger)
	})
}

func GlobalPanicRecovery(c *gin.Context, err any, logger *zap.Logger) {
	switch data := err.(type) {
	case *errors.OptimisticLockError:
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": data.Message, "details": data.Error.Error()})
	case *errors.BusinessRuleError:
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"error": data.Message, "details": data.Error.Error()})
	default:
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
	}

	logger.Panic("Something went wrong", zap.Any("error", err))
}
