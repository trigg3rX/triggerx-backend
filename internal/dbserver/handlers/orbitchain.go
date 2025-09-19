package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
)

func (h *Handler) CreateOrbitChain(c *gin.Context) {
	var req types.CreateOrbitChainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}
	err := h.orbitChainRepository.CreateOrbitChain(&req)
	if err != nil {
		if err.Error() == "chain_id already exists" || err.Error() == "chain_name already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create orbit chain", "details": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "orbit chain created successfully"})
}

func (h *Handler) GetOrbitChainsByUserAddress(c *gin.Context) {
	userAddress := c.Param("user_address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_address is required"})
		return
	}
	chains, err := h.orbitChainRepository.GetOrbitChainsByUserAddress(userAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch orbit chains", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, chains)
}

func (h *Handler) GetAllOrbitChains(c *gin.Context) {
	chains, err := h.orbitChainRepository.GetAllOrbitChains()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch orbit chains", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, chains)
}
