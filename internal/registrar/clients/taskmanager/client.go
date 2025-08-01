package taskmanager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/trigg3rX/triggerx-backend/internal/registrar/config"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

type Client struct {
	retryClient *retry.HTTPClient
}

func NewClient(logger logging.Logger) (*Client, error) {
	retryClient, err := retry.NewHTTPClient(retry.DefaultHTTPRetryConfig(), logger)
	if err != nil {
		return nil, err
	}
	return &Client{retryClient: retryClient}, nil
}

func (c *Client) InformTaskManager(taskID int64, isAccepted bool) error {
	payload := map[string]interface{}{
		"task_id": taskID,
		"is_accepted": isAccepted,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/task/submit", config.GetTaskManagerURL()),
		bytes.NewBuffer(jsonPayload),
	)
	if err != nil {
		return err
	}

	resp, err := c.retryClient.DoWithRetry(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
