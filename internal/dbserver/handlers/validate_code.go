package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

type ValidateCodeRequest struct {
	Code             string `json:"code" binding:"required"`
	Language         string `json:"language" binding:"required"`
	SelectedSafe     string `json:"selected_safe" binding:"required_if=IsSafe true"`
	TargetFunction   string `json:"target_function" binding:"required_unless=TaskDefinitionID 7"`
	TaskDefinitionID int    `json:"task_definition_id" binding:"required"`
	IsSafe           bool   `json:"is_safe"`
}

type ValidateCodeResponse struct { 
	Executable bool   `json:"executable"`
	Output     string `json:"output"`
	Error      string `json:"error,omitempty"`
	SafeMatch  bool   `json:"safe_match"`
}

// generateCacheKey creates a cache key from validation parameters
func (h *Handler) generateCacheKey(ipfsUrl string, req ValidateCodeRequest) string {
	// Create a hash of the validation parameters to use as cache key
	keyData := fmt.Sprintf("%s:%s:%s:%s:%t", ipfsUrl, req.Language, req.SelectedSafe, req.TargetFunction, req.IsSafe)
	hash := sha256.Sum256([]byte(keyData))
	hashStr := hex.EncodeToString(hash[:])
	return fmt.Sprintf("ipfs:validation:%s", hashStr)
}

// ValidateCodeInternal does the actual validation and can be used by other logic (not just HTTP handler)
func (h *Handler) ValidateCodeInternal(ctx context.Context, req ValidateCodeRequest, ipfsUrl string, alchemyAPIKey string) (ValidateCodeResponse, error) {
	// Check cache if Redis client is available and IPFS URL is provided
	// Use validation parameters in the cache key so different req contexts revalidate
	if h.redisClient != nil && ipfsUrl != "" {
		cacheKey := h.generateCacheKey(ipfsUrl, req)
		cachedResult, err := h.redisClient.Get(ctx, cacheKey)
		if err == nil && cachedResult != "" {
			h.logger.Info(ctx, "[ValidateCodeInternal] Cache hit for IPFS URL", observability.String("ipfs_url", ipfsUrl))
			var cachedResp ValidateCodeResponse
			if err := json.Unmarshal([]byte(cachedResult), &cachedResp); err == nil {
				// Return the cached response directly
				return cachedResp, nil
			}
			// If unmarshal fails, continue with validation
			h.logger.Warn(ctx, "[ValidateCodeInternal] Failed to unmarshal cached result, proceeding with validation")
		} else if err != nil {
			h.logger.Warn(ctx, "[ValidateCodeInternal] Error checking cache", observability.Error(err))
		}
	}

	result, err := h.dockerExecutor.ExecuteSource(ctx, req.Code, req.Language, alchemyAPIKey)
	if err != nil {
		// If IsSafe is false, SafeMatch is always true
		safeMatch := !req.IsSafe
		resp := ValidateCodeResponse{Executable: false, Output: "", Error: err.Error(), SafeMatch: safeMatch}
		// Log validation result (failure)
		h.logger.Info(ctx, "[ValidateCodeInternal] Validation result",
			observability.String("language", req.Language),
			observability.String("target_function", req.TargetFunction),
			observability.Bool("is_safe", req.IsSafe),
			observability.String("selected_safe", req.SelectedSafe),
			observability.Bool("executable", resp.Executable),
			observability.Bool("safe_match", resp.SafeMatch),
			observability.String("error", err.Error()))

		// Cache the failure response if Redis client is available and IPFS URL is provided
		if h.redisClient != nil && ipfsUrl != "" {
			cacheKey := h.generateCacheKey(ipfsUrl, req)
			respJSON, err := json.Marshal(resp)
			if err == nil {
				// Cache for 24 hours
				if err := h.redisClient.Set(ctx, cacheKey, string(respJSON), 24*time.Hour); err != nil {
					h.logger.Warn(ctx, "[ValidateCodeInternal] Failed to cache validation result", observability.Error(err))
				} else {
					h.logger.Info(ctx, "[ValidateCodeInternal] Cached validation result for IPFS URL", observability.String("ipfs_url", ipfsUrl))
				}
			}
		}

		return resp, nil
	}

	// Parse the output to get the first element (address) from the printed JSON array
	var firstField string
	if result.Output != "" {
		var arr []interface{}
		if err := json.Unmarshal([]byte(result.Output), &arr); err == nil {
			if len(arr) > 0 {
				if s, ok := arr[0].(string); ok {
					firstField = s
				}
			}
		} else {
			// Fallback: try to parse first non-empty line and take first token
			outputLines := strings.Split(strings.TrimSpace(result.Output), "\n")
			for _, line := range outputLines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				fields := strings.Fields(line)
				if len(fields) > 0 {
					candidate := fields[0]
					candidate = strings.Trim(candidate, ",[]\" ")
					firstField = candidate
					break
				}
			}
		}
	}

	// Check if first field matches selected_safe
	// If IsSafe is false, safeMatch is always true (no validation required)
	var safeMatch bool
	if req.IsSafe {
		safeMatch = strings.EqualFold(firstField, req.SelectedSafe)
	} else {
		safeMatch = true
	}

	// Emit explicit warning when safe address does not match to aid debugging/observability
	if req.IsSafe && !safeMatch {
		h.logger.Warn(ctx, "[ValidateCodeInternal] Safe address mismatch",
			observability.String("language", req.Language),
			observability.String("target_function", req.TargetFunction),
			observability.String("selected_safe", req.SelectedSafe),
			observability.String("first_field", firstField))
	}

	resp := ValidateCodeResponse{
		Executable: result.Success,
		Output:     result.Output,
		SafeMatch:  safeMatch,
	}
	if result.Error != nil {
		resp.Error = result.Error.Error()
	}

	// Cache the full response if Redis client is available and IPFS URL is provided
	// Cache both success and failure cases with validation-parameter key
	if h.redisClient != nil && ipfsUrl != "" {
		cacheKey := h.generateCacheKey(ipfsUrl, req)
		respJSON, err := json.Marshal(resp)
		if err == nil {
			// Cache for 24 hours
			if err := h.redisClient.Set(ctx, cacheKey, string(respJSON), 24*time.Hour); err != nil {
				h.logger.Warn(ctx, "[ValidateCodeInternal] Failed to cache validation result", observability.Error(err))
			} else {
				h.logger.Info(ctx, "[ValidateCodeInternal] Cached validation result for IPFS URL", observability.String("ipfs_url", ipfsUrl))
			}
		}
	}

	return resp, nil
}

// ValidateCodeExecutable validates if provided code compiles/executes successfully in a sandbox (HTTP handler)
func (h *Handler) ValidateCodeExecutable(c *gin.Context) {
	var req ValidateCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// For HTTP endpoint, no IPFS URL is provided, so pass empty string
	resp, _ := h.ValidateCodeInternal(c.Request.Context(), req, "", config.GetAlchemyAPIKey())
	h.logger.Debug(c.Request.Context(), "[ValidateCodeExecutable] Validation result", observability.Bool("executable", resp.Executable), observability.Bool("safe_match", resp.SafeMatch), observability.String("error", resp.Error))
	c.JSON(http.StatusOK, resp)
}
