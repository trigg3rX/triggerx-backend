package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (h *Handler) HandleStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}