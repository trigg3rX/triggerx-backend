package ipfs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	httppkg "github.com/trigg3rX/triggerx-backend/pkg/http"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// HTTPClientInterface defines the interface for HTTP operations
type HTTPClientInterface interface {
	DoWithRetry(req *http.Request) (*http.Response, error)
	Get(url string) (*http.Response, error)
	Post(url, contentType string, body io.Reader) (*http.Response, error)
	Put(url, contentType string, body io.Reader) (*http.Response, error)
	Delete(url string) (*http.Response, error)
	Close()
}

// Client interface defines the methods for IPFS operations
type IPFSClient interface {
	// Upload uploads data to IPFS and returns the CID
	Upload(filename string, data []byte) (string, error)

	// Fetch retrieves content from IPFS by CID
	Fetch(cid string) (types.IPFSData, error)

	// Close closes the client and cleans up resources
	Close() error
}

// client implements the Client interface
type ipfsClient struct {
	config     *Config
	logger     logging.Logger
	httpClient HTTPClientInterface
}

// NewClient creates a new IPFS client with the given configuration
func NewClient(config *Config, logger logging.Logger) (IPFSClient, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	httpClient, err := httppkg.NewHTTPClient(httppkg.DefaultHTTPRetryConfig(), logger)
	if err != nil {
		return nil, err
	}

	return &ipfsClient{
		config:     config,
		logger:     logger,
		httpClient: httpClient,
	}, nil
}

// Upload uploads data to IPFS using Pinata and returns the CID
func (c *ipfsClient) Upload(filename string, data []byte) (string, error) {
	if filename == "" {
		return "", fmt.Errorf("filename cannot be empty")
	}

	if len(data) == 0 {
		return "", fmt.Errorf("data cannot be empty")
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add the file
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %v", err)
	}

	if _, err := part.Write(data); err != nil {
		return "", fmt.Errorf("failed to write data to form: %v", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %v", err)
	}

	// Create request
	req, err := http.NewRequest("POST", c.config.PinataBaseURL, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.config.PinataJWT)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	resp, err := c.httpClient.DoWithRetry(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Warn("Failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http error: status code %d", resp.StatusCode)
	}

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	// Parse response
	var ipfsResponse struct {
		Data struct {
			CID string `json:"cid"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &ipfsResponse); err != nil {
		return "", fmt.Errorf("failed to unmarshal IPFS response: %v", err)
	}

	cid := ipfsResponse.Data.CID
	if cid == "" {
		return "", fmt.Errorf("received empty CID from IPFS")
	}

	c.logger.Info("Successfully uploaded to IPFS", "filename", filename, "cid", cid)
	return cid, nil
}

// Fetch retrieves content from IPFS by CID
func (c *ipfsClient) Fetch(cid string) (types.IPFSData, error) {
	if cid == "" {
		return types.IPFSData{}, fmt.Errorf("CID cannot be empty")
	}

	ipfsURL := "https://" + c.config.PinataHost + "/ipfs/" + cid

	resp, err := c.httpClient.Get(ipfsURL)
	if err != nil {
		return types.IPFSData{}, fmt.Errorf("failed to fetch IPFS content: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Warn("Failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return types.IPFSData{}, fmt.Errorf("http error: status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return types.IPFSData{}, fmt.Errorf("failed to read response body: %v", err)
	}

	var ipfsData types.IPFSData
	if err := json.Unmarshal(body, &ipfsData); err != nil {
		return types.IPFSData{}, fmt.Errorf("failed to unmarshal IPFS data: %v", err)
	}

	return ipfsData, nil
}

// Close closes the client and cleans up resources
func (c *ipfsClient) Close() error {
	// Close the HTTP client if needed
	c.httpClient.Close()
	return nil
}
