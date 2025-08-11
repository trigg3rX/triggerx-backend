package taskmanager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/trigg3rX/triggerx-backend/internal/registrar/config"
	httppkg "github.com/trigg3rX/triggerx-backend/pkg/http"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type Client struct {
	retryClient *httppkg.HTTPClient
	logger      logging.Logger
}

func NewClient(logger logging.Logger) (*Client, error) {
	retryClient, err := httppkg.NewHTTPClient(httppkg.DefaultHTTPRetryConfig(), logger)
	if err != nil {
		return nil, err
	}
	return &Client{retryClient: retryClient, logger: logger}, nil
}

func (c *Client) InformTaskManager(taskID int64, isAccepted bool) error {
	payload := map[string]interface{}{
		"task_id":     taskID,
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
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			c.logger.Errorf("error closing response body: %v", cerr)
		}
	}()

	return nil
}
