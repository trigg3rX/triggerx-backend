package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) CreateKeeperData(c *gin.Context) {
	logger := h.getLogger(c)
	logger.Debugf("POST [CreateKeeperData] Creating keeper data")
	var keeperData types.CreateKeeperData
	if err := c.ShouldBindJSON(&keeperData); err != nil {
		logger.Errorf("Error decoding request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	trackDBOp := metrics.TrackDBOperation("read", "keeper_data")
	existingKeeper, err := h.keeperRepository.GetByNonID(c.Request.Context(), "keeper_address", strings.ToLower(keeperData.KeeperAddress))
	trackDBOp(err)
	if err != nil {
		logger.Errorf("Database error while checking keeper existence: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database error while checking keeper status",
			"code":  "DB_ERROR",
		})
		return
	}

	if existingKeeper != nil && existingKeeper.OperatorID != 0 {
		logger.Debugf("Keeper already exists with operator ID: %d", existingKeeper.OperatorID)
		c.JSON(http.StatusOK, gin.H{
			"message":     "Keeper already exists",
			"operator_id": existingKeeper.OperatorID,
			"status":      "existing",
		})
		return
	}

	// Create new keeper entity
	newKeeper := &types.KeeperDataEntity{
		KeeperName:      keeperData.KeeperName,
		KeeperAddress:   strings.ToLower(keeperData.KeeperAddress),
		Whitelisted:     true,
		Registered:      true,
		Online:          false,
		OnImua:          false,
		EmailID:         keeperData.EmailID,
		RewardsBooster:  "1",
		NoExecutedTasks: 0,
		NoAttestedTasks: 0,
		Uptime:          0,
		KeeperPoints:    "0",
		LastCheckedIn:   time.Now().UTC(),
	}

	trackDBOp = metrics.TrackDBOperation("create", "keeper_data")
	err = h.keeperRepository.Create(c.Request.Context(), newKeeper)
	trackDBOp(err)
	if err != nil {
		logger.Errorf("Error creating keeper data: %v", err)
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
	// 	logger.Errorf(" Error sending welcome email to keeper %s: %v", keeperData.KeeperName, err)
	// 	// Note: We don't return here as the keeper creation was successful
	// } else {
	// 	logger.Infof(" Welcome email sent successfully to keeper %s at %s", keeperData.KeeperName, keeperData.EmailID)
	// }

	logger.Debugf("Successfully created keeper with operator ID: %d", newKeeper.OperatorID)
	c.JSON(http.StatusCreated, gin.H{"operator_id": newKeeper.OperatorID})
}

func (h *Handler) CreateKeeperDataGoogleForm(c *gin.Context) {
	logger := h.getLogger(c)
	logger.Debugf("POST [CreateKeeperDataGoogleForm] Creating keeper data from Google Form")
	var keeperData types.GoogleFormCreateKeeperData
	if err := c.ShouldBindJSON(&keeperData); err != nil {
		logger.Errorf("Error decoding request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "code": "INVALID_REQUEST"})
		return
	}

	keeperData.KeeperAddress = strings.ToLower(keeperData.KeeperAddress)

	trackDBOp := metrics.TrackDBOperation("read", "keeper_data")
	existingKeeper, err := h.keeperRepository.GetByNonID(c.Request.Context(), "keeper_address", keeperData.KeeperAddress)
	trackDBOp(err)
	if err != nil {
		logger.Errorf("Database error while checking keeper existence: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error while checking keeper status", "code": "DB_ERROR"})
		return
	}

	var status string
	var operatorID int64

	if existingKeeper != nil {
		// Update existing keeper
		status = "existing"
		existingKeeper.KeeperName = keeperData.KeeperName
		existingKeeper.EmailID = keeperData.EmailID
		existingKeeper.LastCheckedIn = time.Now().UTC()

		trackDBOp = metrics.TrackDBOperation("update", "keeper_data")
		err = h.keeperRepository.Update(c.Request.Context(), existingKeeper)
		trackDBOp(err)
		if err != nil {
			logger.Errorf("Error updating keeper data: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update keeper", "code": "KEEPER_UPDATE_ERROR"})
			return
		}

		if existingKeeper.OperatorID != 0 {
			operatorID = existingKeeper.OperatorID
		}
	} else {
		// Create new keeper
		status = "created"
		newKeeper := &types.KeeperDataEntity{
			KeeperName:       keeperData.KeeperName,
			KeeperAddress:    keeperData.KeeperAddress,
			RewardsAddress:   keeperData.KeeperAddress, // Use keeper address as default
			ConsensusAddress: "",
			RegisteredTx:     "",
			OperatorID:       0, // Will be assigned later
			VotingPower:      "0",
			Whitelisted:      false,
			Registered:       false,
			Online:           false,
			Version:          "",
			OnImua:           false,
			ChatID:           0,
			EmailID:          keeperData.EmailID,
			RewardsBooster:   "1",
			NoExecutedTasks:  0,
			NoAttestedTasks:  0,
			Uptime:           0,
			KeeperPoints:     "0",
			LastCheckedIn:    time.Now().UTC(),
		}

		trackDBOp = metrics.TrackDBOperation("create", "keeper_data")
		err = h.keeperRepository.Create(c.Request.Context(), newKeeper)
		trackDBOp(err)
		if err != nil {
			logger.Errorf("Error creating keeper data: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create keeper", "code": "KEEPER_CREATION_ERROR"})
			return
		}
		operatorID = 0 // New keeper doesn't have operator ID yet
	}

	logger.Debugf("Successfully processed keeper with address: %s", keeperData.KeeperAddress)
	c.JSON(http.StatusCreated, gin.H{"keeper_address": keeperData.KeeperAddress, "operator_id": operatorID, "status": status})
}
