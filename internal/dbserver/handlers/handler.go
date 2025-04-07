package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// Handler struct update
type Handler struct {
	db     *database.Connection
	logger logging.Logger
	config NotificationConfig
}

// NewHandler function update
func NewHandler(db *database.Connection, logger logging.Logger, config NotificationConfig) *Handler {
	return &Handler{
		db:     db,
		logger: logger,
		config: config,
	}
}

func (h *Handler) SendDataToManager(route string, data interface{}) (bool, error) {
	apiURL := fmt.Sprintf("%s%s", config.ManagerIPAddress, route)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return false, fmt.Errorf("error marshaling data: %v", err)
	}

	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Errorf("error sending data to manager: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("manager service error (status=%d): %s", resp.StatusCode, string(body))
	}

	return true, nil
}
