package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/metrics"
	"github.com/trigg3rX/triggerx-backend/pkg/errors"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func (h *Handler) CreateKeeperDataFromGoogleForm(c *gin.Context) {
	logger := h.getLogger(c)
	var keeperData types.CreateKeeperDataFromGoogleFormRequest
	if err := c.ShouldBindJSON(&keeperData); err != nil {
		logger.Errorf("%s: %v", errors.ErrInvalidRequestBody, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.ErrInvalidRequestBody})
		return
	}
	keeperData.KeeperAddress = strings.ToLower(keeperData.KeeperAddress)
	keeperData.RewardsAddress = strings.ToLower(keeperData.RewardsAddress)
	logger.Debugf("POST [CreateKeeperDataGoogleForm] Address: %s", keeperData.KeeperAddress)

	trackDBOp := metrics.TrackDBOperation("read", "keeper_data")
	existingKeeper, err := h.keeperRepository.GetByID(c.Request.Context(), keeperData.KeeperAddress)
	trackDBOp(err)
	if err != nil {
		logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrDBOperationFailed})
		return
	}

	// If existing keeper, return error
	if existingKeeper != nil {
		logger.Errorf("%s: %s", errors.ErrDBDuplicateRecord, keeperData.KeeperAddress)
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.ErrDBDuplicateRecord})
		return
	}
	
	// Create new keeper
	keeper := &types.KeeperDataEntity{
		KeeperName:       keeperData.KeeperName,
		KeeperAddress:    keeperData.KeeperAddress,
		RewardsAddress:   keeperData.RewardsAddress,
		EmailID:          keeperData.EmailID,
		RewardsBooster:   "1",
		KeeperPoints:     "0",
	}

	trackDBOp = metrics.TrackDBOperation("create", "keeper_data")
	err = h.keeperRepository.Create(c.Request.Context(), keeper)
	trackDBOp(err)
	if err != nil {
		logger.Errorf("%s: %v", errors.ErrDBOperationFailed, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrDBOperationFailed})
		return
	}

	logger.Infof("POST [CreateKeeperDataGoogleForm] Successful, keeper address: %s", keeperData.KeeperAddress)
	c.JSON(http.StatusCreated, gin.H{"keeper_address": keeperData.KeeperAddress})
}
