package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetSafeAddressesByUser handles GET /users/safe-addresses/:user_address
func (h *Handler) GetSafeAddressesByUser(c *gin.Context) {
	userAddress := c.Param("user_address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_address param required"})
		return
	}

	safeAddresses, err := h.safeAddressRepository.GetSafeAddressesByUser(userAddress)
	if err != nil {
		h.logger.Errorf("Error fetching safe addresses for user %s: %v", userAddress, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch safe addresses"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"safe_addresses": safeAddresses})
}

// GetJobsBySafeAddress handles GET /jobs/safe-address/:safe_address
func (h *Handler) GetJobsBySafeAddress(c *gin.Context) {
	safeAddress := c.Param("safe_address")
	if safeAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "safe_address param required"})
		return
	}

	jobs, err := h.jobRepository.GetJobsBySafeAddress(safeAddress)
	if err != nil {
		h.logger.Errorf("Error fetching jobs for safe_address %s: %v", safeAddress, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch jobs for safe address"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"jobs": jobs})
}
