package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	tmdb "github.com/trigg3rX/triggerx-backend/internal/taskmonitor/clients/database"
	"github.com/trigg3rX/triggerx-backend/internal/taskmonitor/clients/notify"
	tmconfig "github.com/trigg3rX/triggerx-backend/internal/taskmonitor/config"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

func main() {
	var taskID int64
	var status string
	var isAccepted bool
	var errMsg string

	flag.Int64Var(&taskID, "task_id", 0, "Task ID to notify for")
	flag.StringVar(&status, "status", "completed", "Task status: completed|failed")
	flag.BoolVar(&isAccepted, "accepted", true, "IsAccepted flag for task")
	flag.StringVar(&errMsg, "error", "", "Error message if failed")
	flag.Parse()

	if taskID == 0 {
		panic("task_id is required")
	}

	if err := tmconfig.Init(); err != nil {
		panic(fmt.Errorf("failed to init config: %w", err))
	}

	logger, err := logging.NewTestLogger()
	if err != nil {
		panic(fmt.Errorf("failed to init logger: %w", err))
	}

	dbCfg := &database.Config{
		Hosts:       []string{tmconfig.GetDatabaseHostAddress() + ":" + tmconfig.GetDatabaseHostPort()},
		Keyspace:    "triggerx",
		Consistency: gocql.Quorum,
		Timeout:     10 * time.Second,
		Retries:     3,
		ConnectWait: 5 * time.Second,
		RetryConfig: retry.DefaultRetryConfig(),
	}
	conn, err := database.NewConnection(dbCfg, logger)
	if err != nil {
		panic(fmt.Errorf("failed to connect db: %w", err))
	}
	defer conn.Close()

	dbClient := tmdb.NewDatabaseClient(logger, conn)
	// Send via both webhook and SMTP; logs will indicate which succeeds/fails
	notifier := notify.NewCompositeNotifier(logger, notify.NewWebhookNotifier(logger), notify.NewSMTPNotifier(logger))

	email, err := dbClient.GetUserEmailByTaskID(taskID)
	if err != nil {
		panic(fmt.Errorf("failed to lookup email for task %d: %w", taskID, err))
	}

	payload := notify.TaskStatusPayload{
		TaskID:     taskID,
		JobID:      0,
		Status:     status,
		IsAccepted: isAccepted,
		Error:      errMsg,
		OccurredAt: time.Now(),
	}

	if err := notifier.NotifyTaskStatus(context.Background(), email, payload); err != nil {
		panic(fmt.Errorf("notify failed: %w", err))
	}

	logger.Infof("Notification sent to %s for task %d with status %s", email, taskID, status)
}
