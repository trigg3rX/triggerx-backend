package ipfs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	httppkg "github.com/trigg3rX/triggerx-backend/pkg/http"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// Client interface defines the methods for IPFS operations
type IPFSClient interface {
	// Upload uploads data to IPFS and returns the CID
	Upload(filename string, data []byte) (string, error)

	// Fetch retrieves content from IPFS by CID
	Fetch(cid string) (types.IPFSData, error)

	// Delete deletes a file from IPFS by CID
	Delete(cid string) error

	// ListFiles lists all files from Pinata v3 API
	ListFiles() ([]PinataFile, error)

	// Close closes the client and cleans up resources
	Close() error
}

// Pinata v3 API structures
type PinataFile struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	CID       string            `json:"cid"`
	Size      int64             `json:"size"`
	MimeType  string            `json:"mime_type"`
	GroupID   string            `json:"group_id,omitempty"`
	Keyvalues map[string]string `json:"keyvalues,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

type PinataListResponse struct {
	Data struct {
		Files         []PinataFile `json:"files"`
		NextPageToken string       `json:"next_page_token,omitempty"`
	} `json:"data"`
}

type PinataDeleteResponse struct {
	Data interface{} `json:"data"`
}

// client implements the Client interface
type ipfsClient struct {
	config     *Config
	logger     logging.Logger
	httpClient httppkg.HTTPClientInterface
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

// Delete file by ID using Pinata v3 API
func (c *ipfsClient) Delete(cid string) error {
	if cid == "" {
		return fmt.Errorf("CID cannot be empty")
	}

	// Find file ID by CID
	network := c.config.PinataHost // Use configurable network from config
	url := fmt.Sprintf("https://api.pinata.cloud/v3/files/%s?cid=%s&limit=1", network, cid)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("failed to search for CID %s: %w", cid, err)
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			if err := resp.Body.Close(); err != nil {
				c.logger.Debugf("failed to close response body: %v", err)
			}
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to search for CID %s: status %d, body: %s",
			cid, resp.StatusCode, string(body))
	}

	var listResp PinataListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return fmt.Errorf("failed to decode search response for CID %s: %w", cid, err)
	}

	if len(listResp.Data.Files) == 0 {
		return fmt.Errorf("no file found with CID %s", cid)
	}

	url = fmt.Sprintf("https://api.pinata.cloud/v3/files/%s/%s", network, listResp.Data.Files[0].ID)

	resp, err = c.httpClient.Delete(url)
	if err != nil {
		return fmt.Errorf("failed to delete file %s: %w", listResp.Data.Files[0].ID, err)
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			if err := resp.Body.Close(); err != nil {
				c.logger.Debugf("failed to close response body: %v", err)
			}
		}
	}()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("failed to delete file %s: status %d, body: %s",
			listResp.Data.Files[0].ID, resp.StatusCode, string(body))
	}

	var deleteResp PinataDeleteResponse
	if err := json.Unmarshal(body, &deleteResp); err != nil {
		// If we can't parse the response but got a success status, that's still OK
		c.logger.Debugf("Could not parse delete response for file %s, but status was %d", listResp.Data.Files[0].ID, resp.StatusCode)
	}

	c.logger.Infof("Successfully deleted file %s from Pinata", listResp.Data.Files[0].ID)
	return nil
}

// List all files from Pinata v3 API
func (c *ipfsClient) ListFiles() ([]PinataFile, error) {
	network := c.config.PinataHost // Use configurable network from config
	url := fmt.Sprintf("https://api.pinata.cloud/v3/files/%s?limit=1000", network)

	var allFiles []PinataFile
	nextPageToken := ""

	for {
		requestURL := url
		if nextPageToken != "" {
			requestURL = fmt.Sprintf("%s&pageToken=%s", url, nextPageToken)
		}

		resp, err := c.httpClient.Get(requestURL)
		if err != nil {
			return nil, fmt.Errorf("failed to list files: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			if err := resp.Body.Close(); err != nil {
				c.logger.Debugf("failed to close response body: %v", err)
			}
			return nil, fmt.Errorf("failed to list files: status %d, body: %s",
				resp.StatusCode, string(body))
		}

		var listResp PinataListResponse
		if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
			if err := resp.Body.Close(); err != nil {
				c.logger.Debugf("failed to close response body: %v", err)
			}
			return nil, fmt.Errorf("failed to decode list response: %w", err)
		}
		if err := resp.Body.Close(); err != nil {
			c.logger.Debugf("failed to close response body: %v", err)
		}

		allFiles = append(allFiles, listResp.Data.Files...)

		// Check if there are more pages
		if listResp.Data.NextPageToken == "" {
			break
		}
		nextPageToken = listResp.Data.NextPageToken
	}

	return allFiles, nil
}

// Close closes the client and cleans up resources
func (c *ipfsClient) Close() error {
	// Close the HTTP client if needed
	c.httpClient.Close()
	return nil
}
