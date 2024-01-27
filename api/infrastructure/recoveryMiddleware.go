package infrastructure

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golobby/container/v3"
	domainErros "github.com/steadfastie/gokube/data/errors"
	"go.uber.org/zap"
)

func RecoveryMiddleware(c *gin.Context, e any) {
	container.Call(func(logger *zap.Logger) {
		GlobalPanicRecovery(c, e, logger)
	})
}

func GlobalPanicRecovery(c *gin.Context, err any, logger *zap.Logger) {
	switch data := err.(type) {
	case *domainErros.OptimisticLockError:
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": data.Message, "details": data.Error.Error()})
	case *domainErros.BusinessRuleError:
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"error": data.Message, "details": data.Error.Error()})
	case *domainErros.NotFoundError:
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": data.Message, "details": data.Error.Error()})
	default:
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
	}

	logger.Panic("Something went wrong", zap.Any("error", err))
}
