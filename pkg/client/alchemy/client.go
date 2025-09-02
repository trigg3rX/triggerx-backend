package alchemy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	httppkg "github.com/trigg3rX/triggerx-backend/pkg/http"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type Client struct {
	apiKey     string
	httpClient *httppkg.HTTPClient
	config     *Config
	logger     logging.Logger
}

func NewAlchemyClient(apiKey string, logger logging.Logger) (*Client, error) {
	config, err := LoadConfig(apiKey)
	if err != nil {
		return nil, err
	}
	httpClient, err := httppkg.NewHTTPClient(httppkg.DefaultHTTPRetryConfig(), logger)
	if err != nil {
		return nil, err
	}

	return &Client{
		apiKey:     config.APIKey,
		httpClient: httpClient,
		config:     config,
		logger:     logger,
	}, nil
}

// makeRequest makes a JSON-RPC request to the Alchemy API
func (c *Client) makeRequest(method string, params interface{}, chainID string) (*Response, error) {
	payload := Payload{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	endpoint := c.config.GetEndpoint(chainID)
	resp, err := c.httpClient.Post(endpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("RPC error: %s (code: %d)", response.Error.Message, response.Error.Code)
	}

	return &response, nil
}
