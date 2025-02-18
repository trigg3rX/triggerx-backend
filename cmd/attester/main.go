package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/attester"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

func main() {
	if err := logging.InitLogger(logging.Development, logging.AttesterProcess); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	logger := logging.GetLogger(logging.Development, logging.AttesterProcess)
	logger.Info("Starting attester node...")

	// Set up attester server using Gin
	attesterRouter := gin.New()
	attesterRouter.Use(gin.Recovery())
	attesterRouter.Use(gin.Logger())
	attesterRouter.POST("/task/validate", attester.ValidateTask)

	// Custom middleware for error handling
	errorHandler := func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			logger.Error("request failed", "errors", c.Errors)
			if !c.Writer.Written() {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal Server Error",
				})
			}
		}
	}

	attesterRouter.Use(errorHandler)
	attesterSrv := &http.Server{
		Addr:    ":4002",
		Handler: attesterRouter,
	}

	// Channel to collect server errors
	serverErrors := make(chan error, 1)

	go func() {
		for {
			logger.Info("Validation Server starting", "address", attesterSrv.Addr)
			if err := attesterSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				serverErrors <- err
				logger.Error("attester server failed, restarting...", "error", err)
				time.Sleep(time.Second)
				continue
			}
			break
		}
	}()

	// Handle graceful shutdown on interrupt/termination signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		logger.Error("server error received", "error", err)
	
	case sig := <-shutdown:
		logger.Info("starting shutdown", "signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := attesterSrv.Shutdown(ctx); err != nil {
			logger.Error("graceful shutdown attester server failed",
				"timeout", 2*time.Second,
				"error", err)
			
			if err := attesterSrv.Close(); err != nil {
				logger.Fatal("could not stop attester server gracefully", "error", err)
			}
		}
	}
	logger.Info("shutdown complete")
}
