package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) CreateKeeperData(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[CreateKeeperData] trace_id=%s - Creating keeper data", traceID)
	var keeperData types.CreateKeeperData
	if err := c.ShouldBindJSON(&keeperData); err != nil {
		h.logger.Errorf("[CreateKeeperData] Error decoding request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	trackDBOp := metrics.TrackDBOperation("read", "keeper_data")
	existingKeeperID, err := h.keeperRepository.CheckKeeperExists(keeperData.KeeperAddress)
	trackDBOp(err)
	if err != nil {
		h.logger.Errorf("[CreateKeeperData] Database error while checking keeper existence: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database error while checking keeper status",
			"code":  "DB_ERROR",
		})
		return
	}

	if existingKeeperID != -1 {
		h.logger.Infof("[CreateKeeperData] Keeper already exists with ID: %d", existingKeeperID)
		c.JSON(http.StatusOK, gin.H{
			"message":   "Keeper already exists",
			"keeper_id": existingKeeperID,
			"status":    "existing",
		})
		return
	}

	trackDBOp = metrics.TrackDBOperation("create", "keeper_data")
	currentKeeperID, err := h.keeperRepository.CreateKeeper(keeperData)
	trackDBOp(err)
	if err != nil {
		h.logger.Errorf("[CreateKeeperData] Error creating keeper data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create keeper",
			"code":  "KEEPER_CREATION_ERROR",
		})
		return
	}

	// TODO: Send email to keeper
	// subject := "Welcome to TriggerX: Operator Whitelisting Confirmed"
	// emailBody := fmt.Sprintf(`
	// 	Hey %s,
	// 	<br><br>
	// 	Thanks for filling out the TriggerX whitelisting form — you're all set!
	// 	<br><br>
	// 	To stay updated on what's coming next (and not miss anything important), hop into our Telegram group:
	// 	<br><br>
	// 	<a href="https://t.me/+1I4euCfrchMxZDhl">https://t.me/+1I4euCfrchMxZDhl</a>
	// 	<br><br>
	// 	This is where we'll be sharing everything you need to know as an operator — from technical updates to rewards info and more.
	// 	<br><br>
	// 	See you there!
	// 	<br><br>
	// 	– Team TriggerX
	// `, keeperData.KeeperName)

	// if err := h.sendEmailNotification(keeperData.EmailID, subject, emailBody); err != nil {
	// 	h.logger.Errorf(" Error sending welcome email to keeper %s: %v", keeperData.KeeperName, err)
	// 	// Note: We don't return here as the keeper creation was successful
	// } else {
	// 	h.logger.Infof(" Welcome email sent successfully to keeper %s at %s", keeperData.KeeperName, keeperData.EmailID)
	// }

	h.logger.Infof("[CreateKeeperData] Successfully created keeper with ID: %d", currentKeeperID)
	c.JSON(http.StatusCreated, gin.H{"keeper_id": currentKeeperID})
}

func (h *Handler) CreateKeeperDataGoogleForm(c *gin.Context) {
	traceID := h.getTraceID(c)
	h.logger.Infof("[CreateKeeperDataGoogleForm] trace_id=%s - Creating keeper data from Google Form", traceID)
	var keeperData types.GoogleFormCreateKeeperData
	if err := c.ShouldBindJSON(&keeperData); err != nil {
		h.logger.Errorf("[CreateKeeperDataGoogleForm] Error decoding request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "code": "INVALID_REQUEST"})
		return
	}

	keeperData.KeeperAddress = strings.ToLower(keeperData.KeeperAddress)

	trackDBOp := metrics.TrackDBOperation("read", "keeper_data")
	existingKeeperID, err := h.keeperRepository.CheckKeeperExistsByAddress(keeperData.KeeperAddress)
	trackDBOp(err)
	if err != nil {
		h.logger.Errorf("[CreateKeeperDataGoogleForm] Database error while checking keeper existence: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error while checking keeper status", "code": "DB_ERROR"})
		return
	}

	var status string
	if existingKeeperID != 0 {
		status = "existing"
	} else {
		status = "created"
	}

	trackDBOp = metrics.TrackDBOperation("create", "keeper_data")
	keeperID, err := h.keeperRepository.CreateOrUpdateKeeperFromGoogleForm(keeperData)
	trackDBOp(err)
	if err != nil {
		h.logger.Errorf("[CreateKeeperDataGoogleForm] Error creating/updating keeper data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create/update keeper", "code": "KEEPER_CREATION_ERROR"})
		return
	}

	h.logger.Infof("[CreateKeeperDataGoogleForm] Successfully processed keeper with ID: %d", keeperID)
	c.JSON(http.StatusCreated, gin.H{"keeper_id": keeperID, "status": status})
}
