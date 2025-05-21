package ipfs

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func UploadToIPFS(gateway string, data []byte) (string, error) {
	resp, err := http.Post(gateway, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return "", fmt.Errorf("failed to upload to IPFS: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to upload to IPFS: status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	return string(body), nil
}

func FetchIPFSContent(gateway string, cid string) (string, error) {
	if strings.HasPrefix(cid, "https://") {
		resp, err := http.Get(cid)
		if err != nil {
			return "", fmt.Errorf("failed to fetch IPFS content from URL: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("failed to fetch IPFS content from URL: status code %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read response body: %v", err)
		}

		return string(body), nil
	}

	ipfsGateway := "https://" + gateway + "/ipfs/" + cid
	resp, err := http.Get(ipfsGateway)
	if err != nil {
		return "", fmt.Errorf("failed to fetch IPFS content: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch IPFS content: status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	return string(body), nil
}
