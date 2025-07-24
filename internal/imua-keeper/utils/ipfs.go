package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/imua-keeper/config"
	"github.com/trigg3rX/triggerx-backend/internal/imua-keeper/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func UploadToIPFS(filename string, data []byte) (string, error) {
	metrics.IPFSUploadSizeBytes.Add(float64(len(data)))

	url := "https://uploads.pinata.cloud/v3/files"

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add the "network" field and set it to "public"
	if err := writer.WriteField("network", "public"); err != nil {
		return "", fmt.Errorf("failed to write network field: %v", err)
	}

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

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+config.GetPinataJWT())
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Warning: failed to close response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to upload to IPFS: status code %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	var ipfsResponse struct {
		Data struct {
			CID string `json:"cid"`
		} `json:"data"`
	}

	if err := json.Unmarshal([]byte(respBody), &ipfsResponse); err != nil {
		return string(respBody), fmt.Errorf("failed to unmarshal IPFS response: %v", err)
	}

	cid := ipfsResponse.Data.CID

	return cid, nil
}

func FetchIPFSContent(cid string) (types.IPFSData, error) {
	const maxRetries = 5
	ipfsURL := "https://" + config.GetIpfsHost() + "/ipfs/" + cid

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		resp, err := http.Get(ipfsURL)
		if err != nil {
			lastErr = fmt.Errorf("failed to fetch IPFS content (attempt %d): %v", attempt, err)
			time.Sleep(300 * time.Millisecond)
			continue
		}

		func() {
			defer func() {
				_ = resp.Body.Close()
			}()

			if resp.StatusCode != http.StatusOK {
				lastErr = fmt.Errorf("failed to fetch IPFS content: status code %d", resp.StatusCode)
				time.Sleep(300 * time.Millisecond)
				return
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				lastErr = fmt.Errorf("failed to read response body: %v", err)
				time.Sleep(300 * time.Millisecond)
				return
			}

			var ipfsData types.IPFSData
			if err := json.Unmarshal(body, &ipfsData); err != nil {
				lastErr = fmt.Errorf("failed to unmarshal IPFS data: %v", err)
				time.Sleep(300 * time.Millisecond)
				return
			}

			metrics.IPFSDownloadSizeBytes.Add(float64(len(body)))
			lastErr = nil
			// Return from the outer function with the result
			ipfsDataResult = ipfsData
		}()

		if lastErr == nil {
			return ipfsDataResult, nil
		}
	}

	return types.IPFSData{}, lastErr
}

var ipfsDataResult types.IPFSData
