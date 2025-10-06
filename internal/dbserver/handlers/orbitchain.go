package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// ChainDeploymentHandler handles chain deployment operations
type ChainDeploymentHandler struct {
	orbitServiceURL string
	httpClient      *http.Client
	logger          logging.Logger
	repository      OrbitChainRepository
}

// OrbitChainRepository interface for orbit chain operations
type OrbitChainRepository interface {
	CreateOrbitChain(chain *types.CreateOrbitChainRequest) error
	GetOrbitChainsByUserAddress(userAddress string) ([]types.OrbitChainData, error)
	GetAllOrbitChains() ([]types.OrbitChainData, error)
	GetOrbitChainByID(chainID int64) (*types.OrbitChainData, error)
	UpdateOrbitChainStatus(chainID int64, status, orbitChainAddress string) error
	UpdateOrbitChainRPCUrl(chainID int64, rpcUrl string) error
}

// NewChainDeploymentHandler creates a new chain deployment handler
func NewChainDeploymentHandler(orbitServiceURL string, logger logging.Logger, repository OrbitChainRepository) *ChainDeploymentHandler {
	return &ChainDeploymentHandler{
		orbitServiceURL: orbitServiceURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger:     logger,
		repository: repository,
	}
}

// DeployChain handles the deployment of a new Orbit chain
func (h *ChainDeploymentHandler) DeployChain(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[%s] Starting chain deployment request", traceID)

	var req types.DeployChainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("[%s] Failed to bind JSON: %v", traceID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	// Validate required fields
	if req.ChainName == "" || req.ChainID == 0 || req.UserAddress == "" {
		h.logger.Errorf("[%s] Missing required fields in deployment request", traceID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return
	}

	h.logger.Infof("[%s] Deploying chain with ID: %d", traceID, req.ChainID)

	// Create database entry with pending status
	createReq := &types.CreateOrbitChainRequest{
		ChainID:     req.ChainID,
		ChainName:   req.ChainName,
		UserAddress: req.UserAddress,
	}

	trackDBOp := metrics.TrackDBOperation("create", "orbit_chain_data")
	if err := h.repository.CreateOrbitChain(createReq); err != nil {
		trackDBOp(err)
		h.logger.Errorf("[%s] Failed to create orbit chain record: %v", traceID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create orbit chain record"})
		return
	}
	trackDBOp(nil)

	// Call orbit service asynchronously
	go h.deployChainAsync(req.ChainID, &req, traceID)

	// Return immediate response
	response := types.ChainDeploymentResponse{
		Success:      true,
		DeploymentID: fmt.Sprintf("%d", req.ChainID),
		Status:       "pending",
		Message:      "Chain deployment initiated",
	}

	h.logger.Infof("[%s] Chain deployment initiated successfully with ID: %d", traceID, req.ChainID)
	c.JSON(http.StatusOK, response)
}

// deployChainAsync handles the asynchronous deployment process
func (h *ChainDeploymentHandler) deployChainAsync(chainID int64, req *types.DeployChainRequest, traceID string) {
	h.logger.Infof("[%s] Starting async deployment for chain ID: %d", traceID, chainID)

	// Update status to deploying_orbit
	trackDBOp := metrics.TrackDBOperation("update", "orbit_chain_data")
	if err := h.repository.UpdateOrbitChainStatus(chainID, "deploying_orbit", ""); err != nil {
		trackDBOp(err)
		h.logger.Errorf("[%s] Failed to update status to deploying_orbit: %v", traceID, err)
		return
	}
	trackDBOp(nil)

	// Prepare request for orbit service
	orbitReq := map[string]interface{}{
		"deployment_id":                  fmt.Sprintf("%d", chainID),
		"chain_name":                     req.ChainName,
		"chain_id":                       req.ChainID,
		"owner_address":                  req.UserAddress,
		"batch_poster":                   req.BatchPoster,
		"validator":                      req.Validator,
		"user_address":                   req.UserAddress,
		"native_token":                   req.NativeToken,
		"token_name":                     req.TokenName,
		"token_symbol":                   req.TokenSymbol,
		"token_decimals":                 req.TokenDecimals,
		"max_data_size":                  req.MaxDataSize,
		"max_fee_per_gas_for_retryables": req.MaxFeePerGasForRetryables,
	}

	// Call orbit service
	orbitResp, err := h.callOrbitService("/deploy-chain", orbitReq, traceID)
	if err != nil {
		h.logger.Errorf("[%s] Orbit service call failed: %v", traceID, err)
		trackDBOp = metrics.TrackDBOperation("update", "orbit_chain_data")
		if updateErr := h.repository.UpdateOrbitChainStatus(chainID, "failed", ""); updateErr != nil {
			h.logger.Errorf("[%s] Failed to update status to failed: %v", traceID, updateErr)
		}
		trackDBOp(nil)
		return
	}

	// Check if orbit deployment was successful
	if !orbitResp["success"].(bool) {
		h.logger.Errorf("[%s] Orbit deployment failed: %v", traceID, orbitResp["message"])
		trackDBOp = metrics.TrackDBOperation("update", "orbit_chain_data")
		if updateErr := h.repository.UpdateOrbitChainStatus(chainID, "failed", ""); updateErr != nil {
			h.logger.Errorf("[%s] Failed to update status to failed: %v", traceID, updateErr)
		}
		trackDBOp(nil)
		return
	}

	// Update status to orbit_deployed
	chainAddress := orbitResp["chain_address"].(string)
	trackDBOp = metrics.TrackDBOperation("update", "orbit_chain_data")
	if err := h.repository.UpdateOrbitChainStatus(chainID, "orbit_deployed", chainAddress); err != nil {
		trackDBOp(err)
		h.logger.Errorf("[%s] Failed to update status to orbit_deployed: %v", traceID, err)
		return
	}
	trackDBOp(nil)

	// Update status to deploying_contracts
	trackDBOp = metrics.TrackDBOperation("update", "orbit_chain_data")
	if err := h.repository.UpdateOrbitChainStatus(chainID, "deploying_contracts", chainAddress); err != nil {
		trackDBOp(err)
		h.logger.Errorf("[%s] Failed to update status to deploying_contracts: %v", traceID, err)
		return
	}
	trackDBOp(nil)

	// Deploy contracts
	contractResp, err := h.callOrbitService("/deploy-contracts", map[string]interface{}{
		"deployment_id": fmt.Sprintf("%d", chainID),
		"chain_address": chainAddress,
		"contracts": []map[string]interface{}{
			{
				"name":     "JobRegistry",
				"bytecode": "0x", // Placeholder - actual bytecode loaded from contract artifacts
			},
			{
				"name":     "TriggerGasRegistry",
				"bytecode": "0x", // Placeholder - actual bytecode loaded from contract artifacts
			},
			{
				"name":     "TaskExecutionSpoke",
				"bytecode": "0x", // Placeholder - actual bytecode loaded from contract artifacts
			},
		},
	}, traceID)

	if err != nil {
		h.logger.Errorf("[%s] Contract deployment failed: %v", traceID, err)
		trackDBOp = metrics.TrackDBOperation("update", "orbit_chain_data")
		if updateErr := h.repository.UpdateOrbitChainStatus(chainID, "failed", chainAddress); updateErr != nil {
			h.logger.Errorf("[%s] Failed to update status to failed: %v", traceID, updateErr)
		}
		trackDBOp(nil)
		return
	}

	// Check if contract deployment was successful
	if !contractResp["success"].(bool) {
		h.logger.Errorf("[%s] Contract deployment failed: %v", traceID, contractResp["message"])
		trackDBOp = metrics.TrackDBOperation("update", "orbit_chain_data")
		if updateErr := h.repository.UpdateOrbitChainStatus(chainID, "failed", chainAddress); updateErr != nil {
			h.logger.Errorf("[%s] Failed to update status to failed: %v", traceID, updateErr)
		}
		trackDBOp(nil)
		return
	}

	// Update status to completed
	trackDBOp = metrics.TrackDBOperation("update", "orbit_chain_data")
	if err := h.repository.UpdateOrbitChainStatus(chainID, "completed", chainAddress); err != nil {
		trackDBOp(err)
		h.logger.Errorf("[%s] Failed to update final status: %v", traceID, err)
		return
	}
	trackDBOp(nil)

	h.logger.Infof("[%s] Chain deployment completed successfully for chain ID: %d", traceID, chainID)
}

// callOrbitService makes HTTP calls to the orbit service
func (h *ChainDeploymentHandler) callOrbitService(endpoint string, payload map[string]interface{}, traceID string) (map[string]interface{}, error) {
	url := h.orbitServiceURL + endpoint
	h.logger.Infof("[%s] Calling orbit service: %s", traceID, url)

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Trace-ID", traceID)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			h.logger.Errorf("Failed to close response body: %v", closeErr)
		}
	}()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("orbit service returned status %d: %v", resp.StatusCode, result)
	}

	return result, nil
}

// GetUserChains retrieves all chains for a specific user
func (h *ChainDeploymentHandler) GetUserChains(c *gin.Context) {
	traceID := h.getTraceID(c)
	userAddress := c.Param("user_address")

	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User address is required"})
		return
	}

	h.logger.Infof("[%s] Getting chains for user: %s", traceID, userAddress)

	trackDBOp := metrics.TrackDBOperation("read", "orbit_chain_data")
	chains, err := h.repository.GetOrbitChainsByUserAddress(userAddress)
	trackDBOp(err)

	if err != nil {
		h.logger.Errorf("[%s] Error retrieving chains for user %s: %v", traceID, userAddress, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve chains"})
		return
	}

	// Convert to ChainDeploymentStatus format for response
	var chainStatuses []types.ChainDeploymentStatus
	for _, chain := range chains {
		chainStatus := types.ChainDeploymentStatus{
			ChainID:           chain.ChainID,
			ChainName:         chain.ChainName,
			UserAddress:       chain.UserAddress,
			DeploymentStatus:  chain.DeploymentStatus,
			OrbitChainAddress: chain.OrbitChainAddress,
			CreatedAt:         chain.CreatedAt,
			UpdatedAt:         chain.UpdatedAt,
		}
		chainStatuses = append(chainStatuses, chainStatus)
	}

	h.logger.Infof("[%s] Retrieved %d chains for user %s", traceID, len(chainStatuses), userAddress)
	c.JSON(http.StatusOK, gin.H{"chains": chainStatuses})
}

// GetChainDeploymentStatus retrieves the status of a specific chain deployment
func (h *ChainDeploymentHandler) GetChainDeploymentStatus(c *gin.Context) {
	traceID := h.getTraceID(c)
	chainIDStr := c.Param("chain_id")

	if chainIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Chain ID is required"})
		return
	}

	chainID, err := strconv.ParseInt(chainIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chain ID format"})
		return
	}

	h.logger.Infof("[%s] Getting deployment status for chain ID: %d", traceID, chainID)

	trackDBOp := metrics.TrackDBOperation("read", "orbit_chain_data")
	chain, err := h.repository.GetOrbitChainByID(chainID)
	trackDBOp(err)

	if err != nil {
		if err == gocql.ErrNotFound {
			h.logger.Errorf("[%s] Chain deployment not found for chain ID: %d", traceID, chainID)
			c.JSON(http.StatusNotFound, gin.H{"error": "Chain deployment not found"})
			return
		}
		h.logger.Errorf("[%s] Error retrieving chain deployment status: %v", traceID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve chain deployment status"})
		return
	}

	// Convert to ChainDeploymentStatus format for response
	chainStatus := types.ChainDeploymentStatus{
		ChainID:           chain.ChainID,
		ChainName:         chain.ChainName,
		UserAddress:       chain.UserAddress,
		DeploymentStatus:  chain.DeploymentStatus,
		OrbitChainAddress: chain.OrbitChainAddress,
		CreatedAt:         chain.CreatedAt,
		UpdatedAt:         chain.UpdatedAt,
	}

	h.logger.Infof("[%s] Retrieved deployment status for chain ID %d: %s", traceID, chainID, chainStatus.DeploymentStatus)
	c.JSON(http.StatusOK, chainStatus)
}

// GetAllChains retrieves all orbit chains
func (h *ChainDeploymentHandler) GetAllChains(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[%s] Getting all chains", traceID)

	trackDBOp := metrics.TrackDBOperation("read", "orbit_chain_data")
	chains, err := h.repository.GetAllOrbitChains()
	trackDBOp(err)

	if err != nil {
		h.logger.Errorf("[%s] Error retrieving all chains: %v", traceID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve chains"})
		return
	}

	// Convert to ChainDeploymentStatus format for response
	var chainStatuses []types.ChainDeploymentStatus
	for _, chain := range chains {
		chainStatus := types.ChainDeploymentStatus{
			ChainID:           chain.ChainID,
			ChainName:         chain.ChainName,
			UserAddress:       chain.UserAddress,
			DeploymentStatus:  chain.DeploymentStatus,
			OrbitChainAddress: chain.OrbitChainAddress,
			CreatedAt:         chain.CreatedAt,
			UpdatedAt:         chain.UpdatedAt,
		}
		chainStatuses = append(chainStatuses, chainStatus)
	}

	h.logger.Infof("[%s] Retrieved %d total chains", traceID, len(chainStatuses))
	c.JSON(http.StatusOK, gin.H{"chains": chainStatuses})
}

// UpdateChainDeploymentStatus updates the deployment status of a chain
func (h *ChainDeploymentHandler) UpdateChainDeploymentStatus(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[%s] Updating chain deployment status", traceID)

	var req struct {
		DeploymentID      string `json:"deployment_id" binding:"required"`
		Status            string `json:"status" binding:"required"`
		OrbitChainAddress string `json:"orbit_chain_address,omitempty"`
		ErrorMessage      string `json:"error_message,omitempty"`
		DeploymentLogs    string `json:"deployment_logs,omitempty"`
		Contracts         []struct {
			Name    string `json:"name"`
			Address string `json:"address"`
		} `json:"contracts,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("[%s] Failed to bind JSON: %v", traceID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	// Parse deployment ID to chain ID
	chainID, err := strconv.ParseInt(req.DeploymentID, 10, 64)
	if err != nil {
		h.logger.Errorf("[%s] Invalid deployment ID format: %v", traceID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deployment ID format"})
		return
	}

	h.logger.Infof("[%s] Updating status for chain ID %d to %s", traceID, chainID, req.Status)

	// Update the chain status in the database
	trackDBOp := metrics.TrackDBOperation("update", "orbit_chain_data")
	if err := h.repository.UpdateOrbitChainStatus(chainID, req.Status, req.OrbitChainAddress); err != nil {
		trackDBOp(err)
		h.logger.Errorf("[%s] Failed to update chain status: %v", traceID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update chain status"})
		return
	}
	trackDBOp(nil)

	h.logger.Infof("[%s] Chain deployment status updated successfully", traceID)
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"message":  "Chain deployment status updated successfully",
		"chain_id": chainID,
		"status":   req.Status,
	})
}

// getTraceID extracts trace ID from gin context
func (h *ChainDeploymentHandler) getTraceID(c *gin.Context) string {
	traceID, exists := c.Get("trace_id")
	if !exists {
		return ""
	}
	return traceID.(string)
}
