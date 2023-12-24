package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	_ "github.com/steadfastie/gokube/docs"
	"github.com/steadfastie/gokube/handlers"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os/signal"
	"strings"
	"syscall"
	"time"
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

	// Create gin router
	router := gin.Default()
	router.Use(metricsHandlerFunc)

	// Configure logs
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer func(logger *zap.Logger) {
		// Error is written if OS didn't take care of flushing buffers out
		if err := logger.Sync(); err != nil && !strings.Contains(err.Error(), "sync /dev/stderr: The handle is invalid.") {
			log.Fatalf("can't sync zap logger: %v", err)
		}
	}(logger)

	// Configure endpoints
	var api = router.Group("/api")
	{
		api.GET("/ping", handlers.EntryHandler)
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
			logger.Fatal("Could not serve server on listen: %s", zap.Error(err))
		}
	}()

	// Listen for the interrupt signal and notify user of shutdown
	<-ctx.Done()
	stop()
	logger.Info("Shutting down gracefully, press Ctrl+C again to force")

	// Use context to set 5 second timeout for server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown: ", zap.Error(err))
	}
	logger.Info("Server exiting")
}
