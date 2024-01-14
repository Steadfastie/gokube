package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	_ "github.com/steadfastie/gokube/docs"
	handlers "github.com/steadfastie/gokube/handlers"
	infra "github.com/steadfastie/gokube/infrastucture"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

var (
	httpReqs = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "How many HTTP requests processed, partitioned by status code and HTTP method.",
		},
		[]string{"code", "method"},
	)
)

func init() {
	prometheus.MustRegister(httpReqs)
	zap.ReplaceGlobals(zap.Must(zap.NewProduction()))
}

func metricsHandlerFunc(c *gin.Context) {
	httpReqs.With(prometheus.Labels{
		"code":   fmt.Sprintf("%d", c.Writer.Status()),
		"method": c.Request.Method}).Inc()
	c.Next()
}

//	@title			Swagger for steadfastie/gokube project
//	@version		1.0
//	@contact.email	alexander.divovich@gmail.com

//	@host		localhost:8080
//	@BasePath	/api

//	@securityDefinitions.basic	BasicAuth

// @externalDocs.description	GitHub repository
// @externalDocs.url			https://github.com/Steadfastie/gokube
func main() {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	infra.InitializeServices(ctx, zap.L())
	defer infra.DisconnectServices(ctx)

	// Create gin router
	router := gin.Default()
	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router.Use(metricsHandlerFunc)
	router.Use(gin.RecoveryWithWriter(gin.DefaultErrorWriter, infra.RecoveryMiddleware))

	// Configure endpoints
	var api = router.Group("/api")
	{
		api.GET("/ping", handlers.EntryHandler)
		api.GET("/panic/:type", handlers.PanicHandler)
	}
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// Run
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			zap.L().Fatal("Could not serve server on listen: %s", zap.Error(err))
		}
	}()

	// Listen for the interrupt signal and notify user of shutdown
	<-ctx.Done()
	stop()
	zap.L().Info("Shutting down gracefully, press Ctrl+C again to force")

	// Use context to set 5 second timeout for server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		zap.L().Fatal("Server forced to shutdown: ", zap.Error(err))
	}
	zap.L().Info("Server exiting")
}
