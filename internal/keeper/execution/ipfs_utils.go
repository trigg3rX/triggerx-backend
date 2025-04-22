package execution

import (
	
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

)

func (e *JobExecutor) fetchFromIPFS(url string) (string, error) {
	// Convert IPFS URL to gateway URL if needed
	gatewayURL := url
	if strings.HasPrefix(url, "ipfs://") {
		cid := strings.TrimPrefix(url, "ipfs://")
		gatewayURL = fmt.Sprintf("https://ipfs.io/ipfs/%s", cid)
	}

	// Fetch the content
	resp, err := http.Get(gatewayURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch from IPFS: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("IPFS fetch failed with status code: %d", resp.StatusCode)
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read IPFS response: %v", err)
	}

	return string(content), nil
}

// func (e *JobExecutor) downloadAndExecuteIPFSScript(scriptUrl string) (*resources.ContainerStats, error) {
// 	codePath, err := resources.DownloadIPFSFile(scriptUrl)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to download script: %v", err)
// 	}
// 	defer os.RemoveAll(filepath.Dir(codePath))

// 	stats, err := e.executeDockerContainer(context.Background(), codePath)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return stats, nil
// }
