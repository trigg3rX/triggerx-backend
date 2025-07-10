package ipfs

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func FetchIPFSContent(ipfsHost string, cid string) (types.IPFSData, error) {
	ipfsUrl := "https://" + ipfsHost + "/ipfs/" + cid
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

	return ipfsData, nil
}
