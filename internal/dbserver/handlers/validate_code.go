package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type ValidateCodeRequest struct {
	Code           string `json:"code" binding:"required"`
	Language       string `json:"language" binding:"required"`
	SelectedSafe   string `json:"selected_safe" binding:"required_if=IsSafe true"`
	TargetFunction string `json:"target_function" binding:"required"`
	IsSafe         bool   `json:"is_safe"`
}

type ValidateCodeResponse struct {
	Executable bool   `json:"executable"`
	Output     string `json:"output"`
	Error      string `json:"error,omitempty"`
	SafeMatch  bool   `json:"safe_match"`
}

// ValidateCodeExecutable validates if provided code compiles/executes successfully in a sandbox
func (h *Handler) ValidateCodeExecutable(c *gin.Context) {
	var req ValidateCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	result, err := h.dockerExecutor.ExecuteSource(ctx, req.Code, req.Language)
	if err != nil {
		// If IsSafe is false, SafeMatch is always true
		safeMatch := !req.IsSafe
		c.JSON(http.StatusOK, ValidateCodeResponse{Executable: false, Output: "", Error: err.Error(), SafeMatch: safeMatch})
		return
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

	resp := ValidateCodeResponse{
		Executable: result.Success,
		Output:     result.Output,
		SafeMatch:  safeMatch,
	}
	if result.Error != nil {
		resp.Error = result.Error.Error()
	}
	c.JSON(http.StatusOK, resp)
}
