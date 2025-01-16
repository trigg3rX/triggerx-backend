package main

import (
	// "context"
	"fmt"
	// "os"
	// "os/signal"
	// "syscall"

	"github.com/trigg3rX/triggerx-backend/execute/keeper"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

func main() {
	if err := logging.InitLogger(logging.Development, "keeper"); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	logger := logging.GetLogger()

	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	keeperNode, err := keeper.NewKeeperFromConfigFile("config-files/triggerx_operator.yaml")
	if err != nil {
		logger.Fatalf("Failed to create keeper node: %v", err)
	}

	logger.Infof("Keeper node: %+v", keeperNode)

	// sigChan := make(chan os.Signal, 1)
	// signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// errChan := make(chan error, 1)
	// go func() {
	// 	if err := keeperNode.Start(ctx); err != nil {
	// 		errChan <- err
	// 	}
	// }()

	// select {
	// case <-sigChan:
	// 	logger.Info("Received shutdown signal")
	// 	cancel()
	// case err := <-errChan:
	// 	logger.Errorf("Keeper node error: %v", err)
	// 	cancel()
	// }

	logger.Info("Shutting down keeper node...")
}
