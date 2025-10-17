package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ValidateCodeRequest struct {
	Code     string `json:"code" binding:"required"`
	Language string `json:"language" binding:"required"`
}

type ValidateCodeResponse struct {
	Executable bool   `json:"executable"`
	Output     string `json:"output"`
	Error      string `json:"error,omitempty"`
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
		c.JSON(http.StatusOK, ValidateCodeResponse{Executable: false, Output: "", Error: err.Error()})
		return
	}

	resp := ValidateCodeResponse{
		Executable: result.Success,
		Output:     result.Output,
	}
	if result.Error != nil {
		resp.Error = result.Error.Error()
	}
	c.JSON(http.StatusOK, resp)
}
