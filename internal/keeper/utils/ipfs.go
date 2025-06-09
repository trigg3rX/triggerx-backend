package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/trigg3rX/triggerx-backend/internal/keeper/config"
	"github.com/trigg3rX/triggerx-backend/internal/keeper/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func UploadToIPFS(filename string, data []byte) (string, error) {
	metrics.IPFSUploadSizeBytes.Add(float64(len(data)))

	url := "https://uploads.pinata.cloud/v3/files"

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

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
	defer resp.Body.Close()

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
	ipfsUrl := "https://" + config.GetIpfsHost() + "/ipfs/" + cid
	resp, err := http.Get(ipfsUrl)
	if err != nil {
		return types.IPFSData{}, fmt.Errorf("failed to fetch IPFS content: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return types.IPFSData{}, fmt.Errorf("failed to fetch IPFS content: status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return types.IPFSData{}, fmt.Errorf("failed to read response body: %v", err)
	}

	var ipfsData types.IPFSData
	if err := json.Unmarshal(body, &ipfsData); err != nil {
		return types.IPFSData{}, fmt.Errorf("failed to unmarshal IPFS data: %v", err)
	}

	metrics.IPFSDownloadSizeBytes.Add(float64(len(body)))

	return ipfsData, nil
}
