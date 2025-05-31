package ipfs

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

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
