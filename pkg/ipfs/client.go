package ipfs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	httppkg "github.com/trigg3rX/triggerx-backend/pkg/http"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// Client interface defines the methods for IPFS operations
type IPFSClient interface {
	// Upload uploads data to IPFS and returns the CID
	Upload(ctx context.Context, filename string, data []byte) (string, error)

	// Fetch retrieves content from IPFS by CID
	Fetch(ctx context.Context, cid string) (types.IPFSData, error)

	// Delete deletes a file from IPFS by CID
	Delete(ctx context.Context, cid string) error

	// ListFiles lists all files from Pinata v3 API
	ListFiles(ctx context.Context) ([]PinataFile, error)

	// Close closes the client and cleans up resources
	Close()
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
	httpClient httppkg.HTTPClientInterface
}

// NewClient creates a new IPFS client with the given configuration
func NewClient(config *Config) (IPFSClient, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	httpClient, err := httppkg.NewHTTPClient(httppkg.DefaultHTTPRetryConfig())
	if err != nil {
		return nil, err
	}

	return &ipfsClient{
		config:     config,
		httpClient: httpClient,
	}, nil
}

// Upload uploads data to IPFS using Pinata and returns the CID
func (c *ipfsClient) Upload(ctx context.Context, filename string, data []byte) (string, error) {
	if filename == "" {
		return "", fmt.Errorf("error uploading to IPFS: filename cannot be empty")
	}

	if len(data) == 0 {
		return "", fmt.Errorf("error uploading to IPFS: data cannot be empty")
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add the "network" field and set it to "public" to ensure content is publicly accessible
	if err := writer.WriteField("network", "public"); err != nil {
		return "", fmt.Errorf("error uploading to IPFS: failed to write network field: %v", err)
	}

	// Add the file
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", fmt.Errorf("error uploading to IPFS: failed to create form file: %v", err)
	}

	if _, err := part.Write(data); err != nil {
		return "", fmt.Errorf("error uploading to IPFS: failed to write data to form: %v", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("error uploading to IPFS: failed to close writer: %v", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", c.config.PinataBaseURL, body)
	if err != nil {
		return "", fmt.Errorf("error uploading to IPFS: failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.config.PinataJWT)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	resp, err := c.httpClient.DoWithRetry(ctx, req)
	if err != nil {
		return "", fmt.Errorf("error uploading to IPFS: failed to send request: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			return
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error uploading to IPFS: http error: status code %d", resp.StatusCode)
	}

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error uploading to IPFS: failed to read response body: %v", err)
	}

	// Parse response
	var ipfsResponse struct {
		Data struct {
			CID string `json:"cid"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &ipfsResponse); err != nil {
		return "", fmt.Errorf("error uploading to IPFS: failed to unmarshal IPFS response: %v", err)
	}

	cid := ipfsResponse.Data.CID
	if cid == "" {
		return "", fmt.Errorf("error uploading to IPFS: received empty CID from IPFS")
	}

	return cid, nil
}

// Fetch retrieves content from IPFS by CID
func (c *ipfsClient) Fetch(ctx context.Context, cid string) (types.IPFSData, error) {
	if cid == "" {
		return types.IPFSData{}, fmt.Errorf("error fetching IPFS content: CID cannot be empty")
	}

	tryHosts := []string{
		c.config.PinataHost,
		"ipfs.io",
	}

	var lastErr error
	for _, host := range tryHosts {
		ipfsURL := "https://" + host + "/ipfs/" + cid
		resp, err := c.httpClient.Get(ctx, ipfsURL)
		if err != nil {
			lastErr = fmt.Errorf("error fetching IPFS content from host %s: %v", host, err)
			continue
		}

		// Check status code before reading body
		if resp.StatusCode != http.StatusOK {
			_ = resp.Body.Close()
			lastErr = fmt.Errorf("error fetching IPFS content from host %s: http error: status code %d", host, resp.StatusCode)
			continue
		}

		// Read response body
		body, err := io.ReadAll(resp.Body)
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log close error but don't fail if read was successful
			if err == nil {
				lastErr = fmt.Errorf("error closing response body from host %s: %v", host, closeErr)
				continue
			}
		}
		if err != nil {
			lastErr = fmt.Errorf("error fetching IPFS content from host %s: failed to read response body: %v", host, err)
			continue
		}

		// Parse JSON response
		var ipfsData types.IPFSData
		if err := json.Unmarshal(body, &ipfsData); err != nil {
			lastErr = fmt.Errorf("error fetching IPFS content from host %s: failed to unmarshal IPFS data: %v", host, err)
			continue
		}

		// Success - return immediately
		return ipfsData, nil
	}

	// All hosts failed
	if lastErr != nil {
		return types.IPFSData{}, lastErr
	}
	return types.IPFSData{}, fmt.Errorf("error fetching IPFS content: unknown error")
}

// Delete file by ID using Pinata v3 API
func (c *ipfsClient) Delete(ctx context.Context, cid string) error {
	if cid == "" {
		return fmt.Errorf("error deleting IPFS content: CID cannot be empty")
	}

	// Find file ID by CID
	// For Pinata v3 API, network should be "public" or "private"
	// PinataHost is used for gateway access, not for API endpoints
	// Default to "public" for searching files
	network := "public"
	url := fmt.Sprintf("https://api.pinata.cloud/v3/files/%s?cid=%s&limit=1", network, cid)

	// Create request with Authorization header
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("error deleting IPFS content: CID %s: failed to create request: %w", cid, err)
	}
	req.Header.Set("Authorization", "Bearer "+c.config.PinataJWT)

	resp, err := c.httpClient.DoWithRetry(ctx, req)
	if err != nil {
		return fmt.Errorf("error deleting IPFS content: CID %s: %w", cid, err)
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			if err := resp.Body.Close(); err != nil {
				return
			}
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error deleting IPFS content: CID %s: status %d, body: %s",
			cid, resp.StatusCode, string(body))
	}

	var listResp PinataListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return fmt.Errorf("error deleting IPFS content: failed to decode search response for CID %s: %w", cid, err)
	}

	if len(listResp.Data.Files) == 0 {
		return fmt.Errorf("error deleting IPFS content: no file found with CID %s", cid)
	}

	url = fmt.Sprintf("https://api.pinata.cloud/v3/files/%s/%s", network, listResp.Data.Files[0].ID)

	// Create DELETE request with Authorization header
	deleteReq, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("error deleting IPFS content: file %s: failed to create request: %w", listResp.Data.Files[0].ID, err)
	}
	deleteReq.Header.Set("Authorization", "Bearer "+c.config.PinataJWT)

	resp, err = c.httpClient.DoWithRetry(ctx, deleteReq)
	if err != nil {
		return fmt.Errorf("error deleting IPFS content: file %s: %w", listResp.Data.Files[0].ID, err)
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			if err := resp.Body.Close(); err != nil {
				return
			}
		}
	}()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("error deleting IPFS content: file %s: status %d, body: %s",
			listResp.Data.Files[0].ID, resp.StatusCode, string(body))
	}

	var deleteResp PinataDeleteResponse
	if err := json.Unmarshal(body, &deleteResp); err != nil {
		// If we can't parse the response but got a success status, that's still OK
		return fmt.Errorf("error deleting IPFS content: failed to parse response %s: %v", listResp.Data.Files[0].ID, err)
	}

	return nil
}

// List all files from Pinata v3 API
func (c *ipfsClient) ListFiles(ctx context.Context) ([]PinataFile, error) {
	// For Pinata v3 API, network should be "public" or "private"
	// PinataHost is used for gateway access, not for API endpoints
	// Default to "public" for listing files
	network := "public"
	url := fmt.Sprintf("https://api.pinata.cloud/v3/files/%s?limit=500", network)

	var allFiles []PinataFile
	nextPageToken := ""

	for {
		requestURL := url
		if nextPageToken != "" {
			requestURL = fmt.Sprintf("%s&pageToken=%s", url, nextPageToken)
		}

		// Create request with Authorization header
		req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
		if err != nil {
			return nil, fmt.Errorf("error listing IPFS files: failed to create request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.config.PinataJWT)

		resp, err := c.httpClient.DoWithRetry(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("error listing IPFS files: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			if err := resp.Body.Close(); err != nil {
				return nil, fmt.Errorf("error listing IPFS files: failed to close response body: %v", err)
			}
			return nil, fmt.Errorf("error listing IPFS files: status %d, body: %s",
				resp.StatusCode, string(body))
		}

		var listResp PinataListResponse
		if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
			if err := resp.Body.Close(); err != nil {
				return nil, fmt.Errorf("error listing IPFS files: failed to close response body: %v", err)
			}
			return nil, fmt.Errorf("error listing IPFS files: failed to decode response: %w", err)
		}
		if err := resp.Body.Close(); err != nil {
			return nil, fmt.Errorf("error listing IPFS files: failed to close response body: %v", err)
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
func (c *ipfsClient) Close() {
	c.httpClient.Close()
}
