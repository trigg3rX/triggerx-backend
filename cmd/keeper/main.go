package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/attester"
	"github.com/trigg3rX/triggerx-backend/internal/performer"
	"github.com/trigg3rX/triggerx-backend/internal/performer/services"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

func main() {
	logger := logging.GetLogger(logging.Development, logging.KeeperProcess)

	services.Init()

	// Set up performer server using Gin for better performance and middleware support
	router := gin.Default()
	router.POST("/task/execute", performer.ExecuteTask)

	// Set up attester server using standard net/http
	mux := http.NewServeMux()
	mux.HandleFunc("/task/validate", attester.ValidateTask)

	performerSrv := &http.Server{
		Addr:    ":4003",
		Handler: router,
	}

	attesterSrv := &http.Server{
		Addr:    ":4002",
		Handler: mux,
	}

	// Channel to collect server errors from both goroutines
	serverErrors := make(chan error, 1)

	// Start both servers concurrently
	go func() {
		logger.Info("Execution Service starting", "address", performerSrv.Addr)
		serverErrors <- performerSrv.ListenAndServe()
	}()

	go func() {
		logger.Info("Validation Server starting", "address", attesterSrv.Addr)
		serverErrors <- attesterSrv.ListenAndServe()
	}()

	// Handle graceful shutdown on interrupt/termination signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		logger.Fatal("error starting server", "error", err)

	case sig := <-shutdown:
		logger.Info("starting shutdown", "signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := performerSrv.Shutdown(ctx); err != nil {
			logger.Error("graceful shutdown performer server failed", 
				"timeout", 5*time.Second,
				"error", err)
			
			if err := performerSrv.Close(); err != nil {
				logger.Fatal("could not stop performer server gracefully", "error", err)
			}
		}

		if err := attesterSrv.Shutdown(ctx); err != nil {
			logger.Error("graceful shutdown attester server failed",
				"timeout", 5*time.Second,
				"error", err)
			
			if err := attesterSrv.Close(); err != nil {
				logger.Fatal("could not stop attester server gracefully", "error", err)
			}
		}
	}
}
