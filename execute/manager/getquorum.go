package manager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type QuorumResponse struct {
	FreeQuorumIDs []int64 `json:"free_quorum_ids"`
}

type QuorumData struct {
	QuorumID int64 `json:"quorum_id"`
	Keepers  []string `json:"keepers"`
}


func GetQuorum() (int64, error) {
	// Make GET request to API endpoint
	resp, err := http.Get("http://localhost:8080/api/quorums/free")
	if err != nil {
		return 0, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response: %v", err)
	}

	// Parse JSON response
	var quorumResp QuorumResponse
	if err := json.Unmarshal(body, &quorumResp); err != nil {
		return 0, fmt.Errorf("failed to parse response: %v", err)
	}

	// Check if there are any free quorums
	if len(quorumResp.FreeQuorumIDs) == 0 {
		return 0, fmt.Errorf("no free quorums available")
	}

	// Select random quorum ID from the available ones
	randomIndex := rand.Intn(len(quorumResp.FreeQuorumIDs))
	return quorumResp.FreeQuorumIDs[randomIndex], nil
}

func AssignQuorumHead(quorumID int64) (string, error) {
	resp, err := http.Get(fmt.Sprintf("http://localhost:8080/api/quorums/%d", quorumID))
	if err != nil {
		return "", fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	var quorumData types.QuorumData
	if err := json.Unmarshal(body, &quorumData); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if len(quorumData.Keepers) == 0 {
		return "", fmt.Errorf("no keepers found in quorum")
	}

	// Select random keeper from the available ones
	randomIndex := rand.Intn(len(quorumData.Keepers))
	return quorumData.Keepers[randomIndex], nil
}
