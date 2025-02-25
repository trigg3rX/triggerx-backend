package validation

import (
	"fmt"
	"io"
	"net/http"

	// "github.com/trigg3rX/triggerx-backend/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

var logger = logging.GetLogger(logging.Development, logging.KeeperProcess)

type ProofResponse struct {
	ProofHash string `json:"proofHash"`
	CID       string `json:"cid"`
}

type ValidationResult struct {
	IsValid bool   `json:"isValid"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

type TaskValidationRequest struct {
	ProofOfTask      string `json:"proofOfTask"`
	Data             string `json:"data"`
	TaskDefinitionID uint16 `json:"taskDefinitionId"`
	Performer        string `json:"performer"`
}

type ValidationResponse struct {
	Data    bool   `json:"data"`
	Error   bool   `json:"error"`
	Message string `json:"message,omitempty"`
}

func ValidateTask(c *gin.Context) {
	var taskRequest TaskValidationRequest
	if err := c.ShouldBindJSON(&taskRequest); err != nil {
		c.JSON(http.StatusBadRequest, ValidationResponse{
			Data:    false,
			Error:   true,
			Message: fmt.Sprintf("Failed to parse request body: %v", err),
		})
		return
	}

	logger.Info("Received Task Validation Request:")
	logger.Infof("Proof of Task: %s", taskRequest.ProofOfTask)
	logger.Infof("Data: %s", taskRequest.Data)
	logger.Infof("Task Definition ID: %d", taskRequest.TaskDefinitionID)
	logger.Infof("Performer Address: %s", taskRequest.Performer)

	ipfsData, err := fetchIPFSContent(taskRequest.ProofOfTask)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ValidationResponse{
			Data:    false,
			Error:   true,
			Message: fmt.Sprintf("Failed to fetch IPFS content: %v", err),
		})
		return
	}

	logger.Info("IPFS Data Fetched")
	logger.Infof("IPFS Data: %s", ipfsData)

	// var ipfsResp types.IPFSResponse
	// if err := json.Unmarshal([]byte(ipfsData), &ipfsResp); err != nil {
	//     c.JSON(http.StatusInternalServerError, ValidationResponse{
	//         Data:    false,
	//         Error:   true,
	//         Message: fmt.Sprintf("Failed to parse IPFS content: %v", err),
	//     })
	//     return
	// }

	c.JSON(http.StatusOK, ValidationResponse{
		Data:    true,
		Error:   false,
		Message: "",
	})
}

func fetchIPFSContent(cid string) (string, error) {
	ipfsGateway := "https://aquamarine-urgent-limpet-846.mypinata.cloud/ipfs/"
	resp, err := http.Get(ipfsGateway + cid)
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