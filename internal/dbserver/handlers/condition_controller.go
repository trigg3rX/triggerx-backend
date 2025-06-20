package handlers

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateCondition(c *gin.Context) {
	
	// Return a fixed value for condition dollar value
	c.JSON(200, gin.H{
		"condition_dollar_value": 42.0, // You can change 42.0 to any fixed value you want to return
	})
}